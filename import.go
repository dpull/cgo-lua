package cgo_lua

/*
#include <lua.h>
#include <lauxlib.h>
#include <assert.h>
#include <stdlib.h>

static const char* cgo_lua_call_entry(lua_State* L, int ref, int arg_count, int result_count) {
    lua_rawgeti(L, LUA_REGISTRYINDEX, ref);
    assert(lua_isfunction(L, -1));
    lua_insert(L, -1 - arg_count);

    int ret = lua_pcall(L, arg_count, result_count, 0);
    if (ret != LUA_OK) {
        return lua_tostring(L, -1);
    }
    return NULL;
}
*/
import "C"
import (
	"unsafe"
)

var Preload = `
local loadfile = loadfile
local setmetatable = setmetatable
local xpcall = xpcall
local debug_traceback = debug.traceback

_G.__IMPORT_FILES = _G.__IMPORT_FILES or {}

local try_load = function(node)
    local path = node.fullpath
    local fn = assert(loadfile(path, "t", node.env))
    assert(xpcall(fn, debug_traceback))
end

function _G.import(filename, force)
    local fullpath = filename
    local node = _G.__IMPORT_FILES[fullpath]
    if node and not force then
        return _G.__IMPORT_FILES[fullpath].env
    end

    local env = {}
    setmetatable(env, {__index = _G})
    node = {env=env, fullpath=fullpath, filename=filename}
    _G.__IMPORT_FILES[fullpath] = node
    try_load(node)
    return node.env
end

local function entry(file, method, ...)
	print("entry", file, method, ...)
	module = _G
	if file ~= nil then
		module = import(file)
	end 

	if method == nil then
		return module
	end 	
	
	local fn = module[method]
	if type(fn) ~= "function" then
		return fn
	end
	return fn(...)
end

return entry
`

func moudleInit(L *C.lua_State) (C.int, error) {
	cname := C.CString("Preload")
	defer C.free(unsafe.Pointer(cname))

	top := gettop(L)
	defer top.settop(L)

	buff, sz := quickCStr(Preload)
	err := C.luaL_loadbufferx(L, buff, sz, cname, nil)
	if err != C.LUA_OK {
		str := C.GoString(C.lua_tolstring(L, -1, nil))
		return 0, Errorf("luaL_loadstring failed, %s", str)
	}

	err = C.lua_pcallk(L, 0, 1, 0, 0, nil)
	if err != C.LUA_OK {
		str := C.GoString(C.lua_tolstring(L, -1, nil))
		return 0, Errorf("lua_pcall failed, %s", str)
	}

	ref := C.luaL_ref(L, C.LUA_REGISTRYINDEX)
	if ref == C.LUA_REFNIL || ref == C.LUA_NOREF {
		return 0, Errorf("luaL_ref failed")
	}

	return ref, nil
}

func (vm *LuaVM) Import(file string, ret *LuaTable) error {
	top := gettop(vm.L)
	defer top.settop(vm.L)

	C.lua_pushnil(vm.L)
	pushString(vm.L, "import")
	pushString(vm.L, file)

	err := C.cgo_lua_call_entry(vm.L, vm.entry, 3, 1)
	if err != nil {
		str := C.GoString(err)
		return Errorf(str)
	}

	if ret != nil {
		v := toGoValue(vm.L, C.LUA_TTABLE, -1)
		if vt, ok := v.(LuaTable); ok {
			*ret = vt
		}
	}
	return nil
}

func (vm *LuaVM) GetMember(file, name string) (interface{}, error) {
	top := gettop(vm.L)
	defer top.settop(vm.L)

	pushString(vm.L, file)
	pushString(vm.L, name)

	err := C.cgo_lua_call_entry(vm.L, vm.entry, 2, 1)
	if err != nil {
		str := C.GoString(err)
		return nil, Errorf(str)
	}
	return toGoValue(vm.L, C.LUA_NUMTAGS, -1), nil
}

func (vm *LuaVM) Call(file, name string, resultCount int, args ...interface{}) ([]LuaValue, error) {
	top := gettop(vm.L)
	defer top.settop(vm.L)

	pushString(vm.L, file)
	pushString(vm.L, name)
	cnt := pushGoValue(vm.L, args)

	err := C.cgo_lua_call_entry(vm.L, vm.entry, cnt+2, C.int(resultCount))
	if err != nil {
		str := C.GoString(err)
		return nil, Errorf(str)
	}

	ret := stackToGoValue(vm.L, C.int(resultCount))
	return ret, nil
}
