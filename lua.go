package cgo_lua

/*
#cgo linux LDFLAGS: -llua -ldl -lm

#include <lua.h>
#include <lualib.h>
#include <lauxlib.h>
#include <stdlib.h>

static int cgo_lua_next(lua_State* L, int top, int idx, int* tkey, int* tvalue) {
	lua_settop(L, top);
	int ret = lua_next(L, idx);
	if (ret) {
		*tkey = lua_type(L, -2);
		*tvalue = lua_type(L, -1);
	}
	return ret;
}

static const void* cgo_lua_tpointer(lua_State* L, int idx) {
	return lua_istable(L, idx) ? lua_topointer(L, idx) : NULL;
}
*/
import "C"
import (
	"fmt"
	"reflect"
	"unsafe"
)

var Errorf = fmt.Errorf

type Value interface{} // string, int64, float64, Table
type Table map[Value]Value
type VM struct {
	L     *C.lua_State
	entry C.int
}
type tableCache map[unsafe.Pointer]Table

var DefaultMapSize = 256

func Open() (*VM, error) {
	L := C.luaL_newstate()
	if L == nil {
		return nil, Errorf("luaL_newstate failed")
	}

	C.luaL_openlibs(L)

	ref, err := moudleInit(L)
	if err != nil {
		C.lua_close(L)
		return nil, err
	}

	return &VM{L: L, entry: ref}, nil
}

func (vm *VM) Close() {
	C.lua_close(vm.L)
}

func (vm *VM) Version() float64 {
	v := C.lua_version(vm.L)
	return float64(*v)
}

func (vm *VM) DoString(str string) ([]Value, error) {
	cname := C.CString("DoString")
	defer C.free(unsafe.Pointer(cname))

	top := getTop(vm.L)
	defer top.setTop(vm.L)

	buff, sz := quickCStr(str)
	err := C.luaL_loadbufferx(vm.L, buff, sz, cname, nil)
	if err != C.LUA_OK {
		str := C.GoString(C.lua_tolstring(vm.L, -1, nil))
		return nil, Errorf("luaL_loadbufferx failed, %s", str)
	}

	err = C.lua_pcallk(vm.L, 0, C.LUA_MULTRET, 0, 0, nil)
	if err != C.LUA_OK {
		str := C.GoString(C.lua_tolstring(vm.L, -1, nil))
		return nil, Errorf("lua_pcallk failed, %s", str)
	}

	resultCount := C.lua_gettop(vm.L) - C.int(top)
	ret := stackToGoValue(vm.L, resultCount)
	return ret, nil
}

func (vm *VM) GetGlobal(name string) Value {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	top := getTop(vm.L)
	defer top.setTop(vm.L)

	tv := C.lua_getglobal(vm.L, cname)
	return toGoValue(vm.L, tv, -1)
}

func quickCStr(str string) (*C.char, C.size_t) {
	gostr := (*reflect.StringHeader)(unsafe.Pointer(&str))
	return (*C.char)(unsafe.Pointer(gostr.Data)), C.size_t(gostr.Len)
}

func pushString(L *C.lua_State, str string) {
	/*
		cstr := C.CString(str)
		defer C.free(unsafe.Pointer(cstr))

		C.lua_pushlstring(L, cstr, C.ulong(len(str)))
	*/
	cstr, sz := quickCStr(str)
	C.lua_pushlstring(L, cstr, sz)
}

func stackToGoValue(L *C.lua_State, resultCount C.int) []Value {
	if resultCount == 0 {
		return nil
	}
	ret := make([]Value, resultCount)
	for i := C.int(0); i < resultCount; i++ {
		ret[i] = toGoValue(L, C.LUA_NUMTAGS, i-resultCount)
	}
	return ret
}

func toGoValue(L *C.lua_State, t C.int, idx C.int) Value {
	tc := make(tableCache)
	return toGoValueSafe(L, t, idx, tc)
}

func toGoValueSafe(L *C.lua_State, t C.int, idx C.int, tc tableCache) Value {
	if t == C.LUA_NUMTAGS {
		t = C.lua_type(L, idx)
	}
	switch t {
	case C.LUA_TNUMBER:
		if C.lua_isinteger(L, idx) != 0 {
			return int64(C.lua_tointegerx(L, idx, nil))
		}
		return float64(C.lua_tonumberx(L, idx, nil))
	case C.LUA_TSTRING:
		return C.GoString(C.lua_tolstring(L, idx, nil))
	case C.LUA_TTABLE:
		return table2Map(L, idx, tc)
	default:
		return nil
	}
}

func table2Map(L *C.lua_State, idx C.int, tc tableCache) Table {
	top := getTop(L)
	defer top.setTop(L)

	if idx < 0 {
		idx = C.int(top) + idx + 1
	}

	ptr := C.cgo_lua_tpointer(L, idx)
	if ptr == nil {
		return nil
	}
	m, ok := tc[ptr]
	if ok {
		return m
	}

	m = make(Table, DefaultMapSize)
	tc[ptr] = m

	C.lua_pushnil(L)
	for {
		var tkey, tvalue C.int
		if C.cgo_lua_next(L, C.int(top)+1, idx, &tkey, &tvalue) == 0 {
			break
		}

		value := toGoValueSafe(L, tvalue, -1, tc)
		if value == nil {
			continue
		}

		key := toGoValueSafe(L, tkey, -2, tc)
		if key == nil {
			continue
		}

		switch key.(type) {
		case Table:
			continue
		}

		m[key] = value
	}
	return m
}

func pushGoValue(L *C.lua_State, args ...interface{}) C.int {
	for _, arg := range args {
		switch argv := arg.(type) {
		case string:
			pushString(L, argv)
		case Table:
			C.lua_pushnil(L) // TODO: lua_createtable
		default:
			v := reflect.ValueOf(arg)
			if v.CanInt() {
				C.lua_pushinteger(L, C.longlong(v.Int()))
			} else if v.CanUint() {
				C.lua_pushinteger(L, C.longlong(v.Uint()))
			} else if v.CanFloat() {
				C.lua_pushnumber(L, C.double(v.Float()))
			} else {
				C.lua_pushnil(L)
			}
		}
	}
	return C.int(len(args))
}

type topStack C.int

func getTop(L *C.lua_State) topStack {
	return topStack(C.lua_gettop(L))
}

func (s topStack) setTop(L *C.lua_State) {
	C.lua_settop(L, C.int(s))
}
