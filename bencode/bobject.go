package bencode

import (
	"bufio"
	"io"
)

type BObject struct {
	t BType // 类型
	v any   // 值
}

// Get 根据泛型类型 T 返回对应的值，并进行类型检查
// Get 泛型函数（包级函数）
func Get[T allowedTypes](o *BObject) (T, error) {
	expectedType := getBType[T]()
	if o.t != expectedType {
		var zeroVal T
		return zeroVal, ErrBType
	}
	val, ok := o.v.(T)
	if !ok {
		var zeroVal T
		return zeroVal, ErrBType
	}
	return val, nil
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

func (o *BObject) GetBType() BType {
	return o.t
}

// 将 BObject 编码写入 Writer 中，返回写入的字节长度
func (o *BObject) Bencode(w io.Writer) (int, error) {
	bw := bufio.NewWriter(w)

	wLen := 0
	switch o.t {
	case BSTR:
		str, err := Get[string](o)
		if err != nil {
			return 0, err
		}
		n, err := EncodeString(bw, str)
		if err != nil {
			return 0, err
		}
		wLen += n
	case BINT:
		str, err := Get[int](o)
		if err != nil {
			return 0, err
		}
		n, err := EncodeInt(bw, str)
		if err != nil {
			return 0, err
		}
		wLen += n
	case BLIST:
		bw.WriteByte('l')
		list, err := Get[[]*BObject](o)
		if err != nil {
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
		dict, err := Get[map[string]*BObject](o)
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
