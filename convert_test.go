package cgo_lua

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func getTestData() []Value {
	tab := make(Table)
	tab[int64(1)] = "a"
	tab[float64(2)] = "b"

	return []Value{int64(300), 5.1, "str", tab}
}

func test(t *testing.T, want []bool, src interface{}, cmp func(v Value)) {
	data := getTestData()
	require.Equal(t, len(data), len(want))

	for i, v := range data {
		err := Convert(v, src)
		if want[i] {
			require.NoError(t, err)
			cmp(v)
		} else {
			require.NotNil(t, err)
		}
	}
}

func TestConvertInt(t *testing.T) {
	want := []bool{true, false, false, false}

	var dst1 int8
	test(t, want, &dst1, func(v Value) {
		require.EqualValues(t, v, dst1)
	})

	var dst2 int16
	test(t, want, &dst2, func(v Value) {
		require.EqualValues(t, v, dst2)
	})
}

func TestConvertUint(t *testing.T) {
	want := []bool{true, false, false, false}

	var dst1 uint8
	test(t, want, &dst1, func(v Value) {
		require.EqualValues(t, v, dst1)
	})

	var dst2 uint16
	test(t, want, &dst2, func(v Value) {
		require.EqualValues(t, v, dst2)
	})
}

func TestConvertFloat(t *testing.T) {
	want := []bool{true, true, false, false}

	var dst float32
	test(t, want, &dst, func(v Value) {
		require.EqualValues(t, v, dst)
	})
}

func TestConvertString(t *testing.T) {
	want := []bool{false, false, true, false}

	var dst string
	test(t, want, &dst, func(v Value) {
		require.Equal(t, v, dst)
	})
}

func TestConvertTable1(t *testing.T) {
	want := []bool{false, false, false, true}

	var dst map[float32]string
	test(t, want, &dst, func(v Value) {
		for k, vv := range v.(Table) {
			switch kk := k.(type) {
			case float64:
				require.Equal(t, vv, dst[float32(kk)])
			case int64:
				require.Equal(t, vv, dst[float32(kk)])
			default:
				panic("????")
			}
		}
	})
}

func TestConvertTable2(t *testing.T) {
	want := []bool{false, false, false, false}

	var dst map[int]string
	test(t, want, &dst, func(v Value) {
	})
}
