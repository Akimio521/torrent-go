package bencode

import (
	"bufio"
	"io"
)

// 编码写入一个字符串
// 写入"Hello World" -> "11:Hello World"
func EncodeString(w io.Writer, val string) (int, error) {
	bw := bufio.NewWriter(w)

	strLen := len(val)
	wLen, err := writeInt(bw, strLen) // 先写入字符串的长度
	if err != nil {
		return 0, err
	}
	err = bw.WriteByte(':') // 再写入分隔符
	if err != nil {
		return 0, err
	}
	wLen++
	bw.WriteString(val) // 写入字符串
	wLen += strLen
	err = bw.Flush() // 将缓冲区内容写入
	if err != nil {
		return 0, err
	}
	return wLen, nil
}

// 解码读出一个字符串
func DecodeString(r io.Reader) (string, error) {
	br := bufio.NewReader(r)

	num, err := readInt(br)
	if err != nil {
		return "", err
	}
	if num < 0 {
		return "", ErrStringLength
	}

	if b, err := br.ReadByte(); err != nil {
		return "", err
	} else if b != ':' { // 分隔符错误
		return "", ErrInvalidStringFormat
	}

	buf := make([]byte, num)

	if n, err := io.ReadAtLeast(br, buf, num); err != nil {
		return "", err
	} else if n != num { // 读取长度不匹配
		return "", ErrStringLength
	}

	return string(buf), nil
}

// 编码写入一个整数
// 写入123 -> i123e
func EncodeInt(w io.Writer, val int) (int, error) {
	bw := bufio.NewWriter(w)

	err := bw.WriteByte('i')
	if err != nil {
		return 0, err
	}
	wLen := 1

	nLen, err := writeInt(bw, val)
	if err != nil {
		return 0, err
	}
	wLen += nLen

	if err = bw.WriteByte('e'); err != nil {
		return 0, err
	}
	wLen++

	if err = bw.Flush(); err != nil {
		return 0, err
	}
	return wLen, nil
}

// 解码读出一个整数
func DecodeInt(r io.Reader) (int, error) {
	br := bufio.NewReader(r)

	if b, err := br.ReadByte(); err != nil {
		return 0, err
	} else if b != 'i' {
		return 0, ErrInvalidIntFormat
	}

	val, err := readInt(br)
	if err != nil {
		return 0, err
	}

	if b, err := br.ReadByte(); err != nil {
		return 0, err
	} else if b != 'e' {
		return 0, ErrInvalidIntFormat
	}

	return val, nil
}
