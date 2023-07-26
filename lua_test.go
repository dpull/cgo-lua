package cgo_lua

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateVM(t *testing.T) {
	vm, err := Open()
	require.Nil(t, err)
	require.NotNil(t, vm)
	defer vm.Close()
}

func TestDoStringAndGetGlobal(t *testing.T) {
	vm, err := Open()
	require.Nil(t, err)
	require.NotNil(t, vm)
	defer vm.Close()

	ret, err := vm.DoString(`
		_G.cgo_test = {
			[1] = "test1",
			["test2"] = 2.0,
			["test3"] = {
				[1] = 5,
				[2] = go_call,
				[3] = 6
			}
		}
		return 1, "a", 2.0
	`)
	require.Nil(t, err)
	require.Equal(t, LuaInt(1), ret[0])
	require.Equal(t, LuaString("a"), ret[1])
	require.Equal(t, LuaDouble(2.0), ret[2])

	v := vm.GetGlobal("cgo_test")
	tab := v.(LuaTable)
	require.Equal(t, LuaString("test1"), tab[LuaInt(1)])
	require.Equal(t, LuaDouble(2.0), tab[LuaString("test2")])

	tab2 := tab[LuaString("test3")].(LuaTable)
	require.Equal(t, LuaInt(5), tab2[LuaInt(1)])
	require.Equal(t, nil, tab2[LuaInt(2)])
	require.Equal(t, LuaInt(6), tab2[LuaInt(3)])
}

func TestImport(t *testing.T) {
	vm, err := Open()
	require.Nil(t, err)
	require.NotNil(t, vm)
	defer vm.Close()

	err = vm.Import("XXXXXX.XXXXXXX")
	require.NotNil(t, err)
}
