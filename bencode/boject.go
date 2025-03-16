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

func GetBOject[T allowedTypes](v T) *BObject {
	return &BObject{
		t: getBType[T](),
		v: v,
	}
}
