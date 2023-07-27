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
				[2] = print,
				[3] = 6
			}
		}
		return 1, "a", 2.0
	`)
	require.Nil(t, err)
	require.Equal(t, int64(1), ret[0])
	require.Equal(t, "a", ret[1])
	require.Equal(t, float64(2.0), ret[2])

	v := vm.GetGlobal("cgo_test")
	tab := v.(LuaTable)
	require.Equal(t, "test1", tab[int64(1)])
	require.Equal(t, float64(2.0), tab["test2"])

	tab2 := tab["test3"].(LuaTable)
	require.Equal(t, int64(5), tab2[int64(1)])
	require.Equal(t, nil, tab2[int64(2)])
	require.Equal(t, int64(6), tab2[int64(3)])
}

func TestDoStringCircularRef(t *testing.T) {
	vm, err := Open()
	require.Nil(t, err)
	require.NotNil(t, vm)
	defer vm.Close()

	ret, err := vm.DoString(`
		local tb = {
			a = 1, b = 2
		}
		tb.c = tb
		return tb
	`)
	require.Nil(t, err)
	retTab := ret[0].(LuaTable)
	require.Equal(t, int64(1), retTab["a"])
	require.Equal(t, int64(2), retTab["b"])
	require.Equal(t, retTab, retTab["c"].(LuaTable))
}

func TestImport(t *testing.T) {
	vm, err := Open()
	require.Nil(t, err)
	require.NotNil(t, vm)
	defer vm.Close()

	err = vm.Import("XXXXXX.XXXXXXX", nil)
	require.NotNil(t, err)
}
