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
	ErrBType               = errors.New("error BOject type")            // 非法 BOject 类型
	ErrInvalidBObject      = errors.New("invalid BObject encoding")     // 非法 BObject 编码
	ErrInvalidStringFormat = errors.New("invalid string format")        // 非法字符串格式
	ErrInvalidIntFormat    = errors.New("invalid int format")           // 非法整数格式
	ErrMissingDigits       = errors.New("missing digits after sign")    // 正负号后缺少数字
	ErrStringLength        = errors.New("invalid string length prefix") // 字符串长度非法
)
