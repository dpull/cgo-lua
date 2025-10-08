package cgo_lua

import "reflect"

func Convert(src Value, dst interface{}) error {
	dv := reflect.ValueOf(dst)
	if dv.Kind() != reflect.Pointer {
		return Errorf("dst must be a pointer")
	}
	dv = dv.Elem()
	return convert(src, dv)
}

func convert(src Value, dst reflect.Value) error {
	switch dst.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if v, ok := src.(int64); ok {
			dst.SetInt(v)
			return nil
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if v, ok := src.(int64); ok {
			dst.SetUint(uint64(v))
			return nil
		}
	case reflect.Float32, reflect.Float64:
		if v, ok := src.(float64); ok {
			dst.SetFloat(v)
			return nil
		}
		if v, ok := src.(int64); ok {
			dst.SetFloat(float64(v))
			return nil
		}
	case reflect.String:
		if v, ok := src.(string); ok {
			dst.SetString(v)
			return nil
		}
	case reflect.Bool:
		if v, ok := src.(bool); ok {
			dst.SetBool(v)
			return nil
		}
	case reflect.Map:
		if v, ok := src.(Table); ok {
			return convertMap(v, dst)
		}
	case reflect.Interface:
		dst.Set(reflect.ValueOf(src))
		return nil
	}
	return Errorf("type cannot be converted, %v -> %v", reflect.TypeOf(src).Kind(), dst.Kind())
}

func convertMap(src Table, dst reflect.Value) error {
	t := dst.Type()
	tk := t.Key()
	te := t.Elem()

	if dst.IsNil() {
		dst.Set(reflect.MakeMapWithSize(t, len(src)))
	}

	for k, v := range src {
		vk := reflect.New(tk).Elem()
		if err := convert(k, vk); err != nil {
			return err
		}

		ve := reflect.New(te).Elem()
		if err := convert(v, ve); err != nil {
			return err
		}

		dst.SetMapIndex(vk, ve)
	}
	return nil
}
