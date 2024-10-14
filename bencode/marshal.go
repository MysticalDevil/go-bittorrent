package bencode

import (
	"io"
	"reflect"
	"strings"
)

func Marshal(w io.Writer, s any) (int, error) {
	v := reflect.ValueOf(s)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	return marshalValue(w, v)
}

func marshalValue(w io.Writer, v reflect.Value) (len int, err error) {
	var tmpLen int
	switch v.Kind() {
	case reflect.String:
		tmpLen, err = EncodeString(w, v.String())
		if err != nil {
			return 0, err
		}
		len += tmpLen
	case reflect.Int:
		tmpLen, err = EncodeInt(w, int(v.Int()))
		if err != nil {
			return 0, err
		}
		len += tmpLen
	case reflect.Slice:
		tmpLen, err = marshalList(w, v)
		if err != nil {
			return 0, err
		}
		len += tmpLen
	case reflect.Struct:
		tmpLen, err = marshalDict(w, v)
		if err != nil {
			return 0, err
		}
		len += tmpLen
	default:
		panic("unhandled default case")
	}
	return
}

func marshalList(w io.Writer, v reflect.Value) (len int, err error) {
	len = 2
	var tmpLen int
	_, err = w.Write([]byte("l"))
	if err != nil {
		return 0, err
	}
	for i := 0; i < v.Len(); i++ {
		ev := v.Index(i)
		tmpLen, err = marshalValue(w, ev)
		if err != nil {
			return 0, err
		}
		len += tmpLen
	}
	_, err = w.Write([]byte{'e'})
	if err != nil {
		return 0, err
	}
	return
}

func marshalDict(w io.Writer, v reflect.Value) (len int, err error) {
	len = 2
	var tmpLen int
	_, err = w.Write([]byte("d"))
	if err != nil {
		return 0, err
	}
	for i := 0; i < v.NumField(); i++ {
		fv := v.Field(i)
		ft := v.Type().Field(i)
		key := ft.Tag.Get("bencode")
		if key == "" {
			key = strings.ToLower(ft.Name)
		}

		tmpLen, err = EncodeString(w, key)
		if err != nil {
			return 0, err
		}
		len += tmpLen

		tmpLen, err = marshalValue(w, fv)
		if err != nil {
			return 0, err
		}
		len += tmpLen
	}
	_, err = w.Write([]byte("e"))
	if err != nil {
		return 0, err
	}
	return
}
