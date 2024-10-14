package bencode

import (
	"errors"
	"io"
	"reflect"
	"strings"
)

func Unmarshal(r io.Reader, src any) error {
	o, err := Parse(r)
	if err != nil {
		return err
	}
	p := reflect.ValueOf(src)
	if p.Kind() != reflect.Ptr {
		return errors.New("dest must be a pointer")
	}
	switch o.type_ {
	case BLIST:
		list, err := o.List()
		if err != nil {
			return err
		}
		l := reflect.MakeSlice(p.Elem().Type(), len(list), len(list))
		p.Elem().Set(l)
		err = unmarshalList(p, list)
		if err != nil {
			return err
		}
	case BDICT:
		dict, err := o.Dict()
		if err != nil {
			return err
		}
		err = unmarshalDict(p, dict)
		if err != nil {
			return err
		}
	default:
		return errors.New("src code must be struct or slice")
	}
	return nil
}

func unmarshalList(p reflect.Value, list []*BObject) error {
	if p.Kind() != reflect.Ptr || p.Elem().Type().Kind() != reflect.Slice {
		return errors.New("dest must be a pointer to a slice")
	}
	v := p.Elem()
	if len(list) == 0 {
		return nil
	}
	switch list[0].type_ {
	case BSTR:
		for i, o := range list {
			val, err := o.Str()
			if err != nil {
				return err
			}
			v.Index(i).SetString(val)
		}
	case BINT:
		for i, o := range list {
			val, err := o.Int()
			if err != nil {
				return err
			}
			v.Index(i).SetInt(int64(val))
		}
	case BLIST:
		for i, o := range list {
			val, err := o.List()
			if err != nil {
				return err
			}
			if v.Type().Elem().Kind() != reflect.Slice {
				return ErrType
			}
			lp := reflect.New(v.Type().Elem())
			ls := reflect.MakeSlice(v.Type().Elem(), len(val), len(val))
			lp.Elem().Set(ls)
			err = unmarshalList(lp, val)
			if err != nil {
				return err
			}
			v.Index(i).Set(lp.Elem())
		}
	case BDICT:
		for i, o := range list {
			val, err := o.Dict()
			if err != nil {
				return err
			}
			if v.Type().Elem().Kind() != reflect.Struct {
				return ErrType
			}
			dp := reflect.New(v.Type().Elem())
			err = unmarshalDict(dp, val)
			if err != nil {
				return err
			}
			v.Index(i).Set(dp.Elem())
		}
	}
	return nil
}

func unmarshalDict(p reflect.Value, dict map[string]*BObject) error {
	if p.Kind() != reflect.Ptr || p.Elem().Type().Kind() != reflect.Struct {
		return errors.New("dest must be a pointer to a struct")
	}
	v := p.Elem()
	for i, n := 0, v.NumField(); i < n; i++ {
		fv := v.Field(i)
		if !fv.CanSet() {
			continue
		}
		ft := v.Type().Field(i)
		key := ft.Tag.Get("bencode")
		if key == "" {
			key = strings.ToLower(ft.Name)
		}
		fo := dict[key]
		if fo == nil {
			continue
		}
		switch fo.type_ {
		case BSTR:
			if ft.Type.Kind() != reflect.String {
				break
			}
			val, err := fo.Str()
			if err != nil {
				return err
			}
			fv.SetString(val)
		case BINT:
			if ft.Type.Kind() != reflect.Int {
				break
			}
			val, err := fo.Int()
			if err != nil {
				return err
			}
			fv.SetInt(int64(val))
		case BLIST:
			if ft.Type.Kind() != reflect.Slice {
				break
			}
			list, err := fo.List()
			if err != nil {
				return err
			}
			lp := reflect.New(ft.Type)
			ls := reflect.MakeSlice(ft.Type, len(list), len(list))
			lp.Elem().Set(ls)
			err = unmarshalList(lp, list)
			if err != nil {
				return err
			}
			fv.Set(lp.Elem())
		case BDICT:
			if ft.Type.Kind() != reflect.Struct {
				break
			}
			dp := reflect.New(ft.Type)
			dict, err := fo.Dict()
			if err != nil {
				return err
			}
			err = unmarshalDict(dp, dict)
			if err != nil {
				return err
			}
			fv.Set(dp.Elem())
		}
	}
	return nil
}
