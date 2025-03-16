package bencode

import (
	"errors"
	"io"
	"reflect"
	"strings"
)

// 从 io.Reader 读数据绑定在结构体上
func Unmarshal(r io.Reader, s any) error {
	p := reflect.ValueOf(s)
	if p.Kind() != reflect.Ptr {
		return ErrNoPtr
	}

	o, err := Parse(r)
	if err != nil {
		return err
	}

	switch o.GetBType() {
	case BLIST:
		list, err := Get[[]*BObject](o)
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
		dict, err := Get[map[string]*BObject](o)
		if err != nil {
			return err
		}
		err = unmarshalDict(p, dict)
		if err != nil {
			return err
		}
	default:
		return ErrSrcMustBeStructOrSlice
	}
	return nil
}

// p.Kind must be Ptr && p.Elem().Type().Kind() must be Slice
func unmarshalList(p reflect.Value, list []*BObject) error {
	if p.Kind() != reflect.Ptr || p.Elem().Type().Kind() != reflect.Slice {
		return ErrDestMustBeSlice
	}
	v := p.Elem()
	if len(list) == 0 {
		return nil
	}
	switch list[0].GetBType() {
	case BSTR:
		for i, o := range list {
			val, err := Get[string](o)
			if err != nil {
				return err
			}
			v.Index(i).SetString(val)
		}
	case BINT:
		for i, o := range list {
			val, err := Get[int](o)
			if err != nil {
				return err
			}
			v.Index(i).SetInt(int64(val))
		}
	case BLIST:
		for i, o := range list {
			val, err := Get[[]*BObject](o)
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
			val, err := Get[map[string]*BObject](o)
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

// p.Kind() must be Ptr && p.Elem().Type().Kind() must be Struct
func unmarshalDict(p reflect.Value, dict map[string]*BObject) error {
	if p.Kind() != reflect.Ptr || p.Elem().Type().Kind() != reflect.Struct {
		return errors.New("dest must be pointer")
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
		switch fo.GetBType() {
		case BSTR:
			if ft.Type.Kind() != reflect.String {
				break
			}
			val, err := Get[string](fo)
			if err != nil {
				return err
			}
			fv.SetString(val)
		case BINT:
			if ft.Type.Kind() != reflect.Int {
				break
			}
			val, err := Get[int](fo)
			if err != nil {
				return err
			}
			fv.SetInt(int64(val))
		case BLIST:
			if ft.Type.Kind() != reflect.Slice {
				break
			}
			list, err := Get[[]*BObject](fo)
			if err != nil {
				break
			}
			lp := reflect.New(ft.Type)
			ls := reflect.MakeSlice(ft.Type, len(list), len(list))
			lp.Elem().Set(ls)
			err = unmarshalList(lp, list)
			if err != nil {
				break
			}
			fv.Set(lp.Elem())
		case BDICT:
			if ft.Type.Kind() != reflect.Struct {
				break
			}
			dp := reflect.New(ft.Type)
			dict, err := Get[map[string]*BObject](fo)
			if err != nil {
				break
			}
			err = unmarshalDict(dp, dict)
			if err != nil {
				break
			}
			fv.Set(dp.Elem())
		}
	}
	return nil
}

func marshalValue(w io.Writer, v reflect.Value) (int, error) {
	len := 0
	switch v.Kind() {
	case reflect.String:
		if n, err := EncodeString(w, v.String()); err != nil {
			return 0, err
		} else {
			len += n
		}
	case reflect.Int:
		if n, err := EncodeInt(w, int(v.Int())); err != nil {
			return 0, err
		} else {
			len += n
		}
	case reflect.Slice:
		n, err := marshalList(w, v)
		if err != nil {
			return 0, err
		}
		len += n
	case reflect.Struct:
		n, err := marshalDict(w, v)
		if err != nil {
			return 0, err
		}
		len += n
	}
	return len, nil
}

func marshalList(w io.Writer, vl reflect.Value) (int, error) {
	len := 2
	w.Write([]byte{'l'})
	for i := 0; i < vl.Len(); i++ {
		ev := vl.Index(i)
		n, err := marshalValue(w, ev)
		if err != nil {
			return 0, err
		}
		len += n
	}
	w.Write([]byte{'e'})
	return len, nil
}

func marshalDict(w io.Writer, vd reflect.Value) (int, error) {
	len := 2
	w.Write([]byte{'d'})
	for i := 0; i < vd.NumField(); i++ {
		fv := vd.Field(i)
		ft := vd.Type().Field(i)
		key := ft.Tag.Get("bencode")
		if key == "" {
			key = strings.ToLower(ft.Name)
		}
		n, err := EncodeString(w, key)
		if err != nil {
			return 0, err
		}
		len += n
		n, err = marshalValue(w, fv)
		if err != nil {
			return 0, err
		}
		len += n
	}
	w.Write([]byte{'e'})
	return len, nil
}

func Marshal(w io.Writer, s any) (int, error) {
	v := reflect.ValueOf(s)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	return marshalValue(w, v)
}
