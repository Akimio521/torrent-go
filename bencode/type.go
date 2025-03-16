package bencode

import "errors"

type BType uint8
type allowedTypes interface {
	string | int | []*BObject | map[string]*BObject
}

const (
	BSTR BType = iota
	BINT
	BLIST
	BDICT
)

var (
	ErrBType                  = errors.New("error BObject type")               // 非法 BObject 类型
	ErrInvalidBObject         = errors.New("invalid BObject encoding")         // 非法 BObject 编码
	ErrInvalidStringFormat    = errors.New("invalid string format")            // 非法字符串格式
	ErrInvalidIntFormat       = errors.New("invalid int format")               // 非法整数格式
	ErrMissingDigits          = errors.New("missing digits after sign")        // 正负号后缺少数字
	ErrStringLength           = errors.New("invalid string length prefix")     // 字符串长度非法
	ErrNoPtr                  = errors.New("dest must be a pointer")           // 传递参数非指针类型
	ErrSrcMustBeStructOrSlice = errors.New("src code must be struct or slice") // 被绑定的必须是结构体或者切片
	ErrDestMustBeSlice        = errors.New("dest must be pointer to slice")    // 目标必须是切片指针
	ErrType                   = errors.New("error type")                       // 类型错误
)
