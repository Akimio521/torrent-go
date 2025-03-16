package bencode

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

func GetBOject[T allowedTypes](v T) *BObject {
	return &BObject{
		t: getBType[T](),
		v: v,
	}
}
