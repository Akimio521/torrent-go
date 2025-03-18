package bencode

import (
	"bufio"
	"bytes"
	"errors"
	"io"
)

type BObject struct {
	t BType // 类型
	v any   // 值
}

func (o *BObject) GetBType() BType {
	return o.t
}

// Get 根据泛型类型 T 返回对应的值，
func GetValue[T allowedTypes](o *BObject, dest *T) error {
	expectedType := getBType[T]()
	if o.t != expectedType {
		return ErrBType
	}
	var ok = false

	if *dest, ok = o.v.(T); !ok {
		return ErrBType
	}
	return nil
}

// 类型映射函数
func getBType[T allowedTypes]() BType {
	var t T
	switch any(t).(type) {
	case string:
		return BSTR
	case int:
		return BINT
	case []*BObject:
		return BLIST
	case map[string]*BObject:
		return BDICT
	default:
		panic("unsupported type")
	}
}

// 将 BObject 编码写入 Writer 中，返回写入的字节长度
func (o *BObject) Bencode(w io.Writer) (int, error) {
	bw := bufio.NewWriter(w)

	wLen := 0
	switch o.t {
	case BSTR:
		var str string
		if err := GetValue(o, &str); err != nil {
			return 0, err
		}
		n, err := EncodeString(bw, str)
		if err != nil {
			return 0, err
		}
		wLen += n
	case BINT:
		var num int
		if err := GetValue(o, &num); err != nil {
			return 0, err
		}
		n, err := EncodeInt(bw, num)
		if err != nil {
			return 0, err
		}
		wLen += n
	case BLIST:
		bw.WriteByte('l')
		var list []*BObject
		if err := GetValue(o, &list); err != nil {
			return 0, err
		}
		for _, elem := range list {
			n, err := elem.Bencode(bw)
			if err != nil {
				return 0, nil
			}
			wLen += n
		}
		bw.WriteByte('e')
		wLen += 2 // 一个是"l"一个是"e"
	case BDICT:
		bw.WriteByte('d')
		var dict map[string]*BObject
		err := GetValue(o, &dict)
		if err != nil {
			return 0, nil
		}
		for k, v := range dict {
			n, err := EncodeString(bw, k)
			if err != nil {
				return 0, nil
			}
			wLen += n
			n, err = v.Bencode(bw)
			if err != nil {
				return 0, nil
			}
			wLen += n
		}
		bw.WriteByte('e')
		wLen += 2 // 一个是"d"一个是"e"
	}
	bw.Flush()
	return wLen, nil
}

func (bObj *BObject) GetDictKeyDay(keyName string) ([]byte, error) {
	if bObj.GetBType() != BDICT {
		return nil, errors.New("not a dict")
	}
	dict := make(map[string]*BObject)
	GetValue(bObj, &dict)
	if v, ok := dict[keyName]; ok {
		buf := new(bytes.Buffer)
		if _, err := v.Bencode(buf); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	}
	return nil, errors.New("key not found")
}

func GetBObject[T allowedTypes](v T) *BObject {
	return &BObject{
		t: getBType[T](),
		v: v,
	}
}

func Parse(r io.Reader) (*BObject, error) {
	br := bufio.NewReader(r)

	b, err := br.Peek(1)
	if err != nil {
		return nil, err
	}
	var ret *BObject
	switch {
	case b[0] >= '0' && b[0] <= '9': // 字符串类型
		val, err := DecodeString(br)
		if err != nil {
			return nil, err
		}
		ret = GetBObject(val)
	case b[0] == 'i': // 整数类型
		val, err := DecodeInt(br)
		if err != nil {
			return nil, err
		}
		ret = GetBObject(val)
	case b[0] == 'l': // 列表类型
		br.ReadByte() // 读取 "l"
		var val []*BObject
		for {
			if p, _ := br.Peek(1); p[0] == 'e' {
				br.ReadByte() // 读取 "e"
				break
			}
			elem, err := Parse(br)
			if err != nil {
				return nil, err
			}
			val = append(val, elem)
		}
		ret = GetBObject(val)
	case b[0] == 'd': // 字典类型
		br.ReadByte() // 读取 "d"
		dict := make(map[string]*BObject)
		for {
			if p, _ := br.Peek(1); p[0] == 'e' {
				br.ReadByte() // 读取 "e"
				break
			}
			key, err := DecodeString(br)
			if err != nil {
				return nil, err
			}
			val, err := Parse(br)
			if err != nil {
				return nil, err
			}
			dict[key] = val
		}
		ret = GetBObject(dict)
	default:
		return nil, ErrInvalidBObject
	}
	return ret, nil
}
