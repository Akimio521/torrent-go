package bencode

import (
	"bufio"
	"io"
	"strconv"
)

// 将整数写入 Wirter 中，并返回写入的长度
func writeInt(w *bufio.Writer, val int) (int, error) {
	n, err := w.WriteString(strconv.Itoa(val))
	if err != nil {
		return 0, err
	}
	return n, nil
}

// 高效内联数字检查
func isDigit(b byte) bool {
	return b >= '0' && b <= '9'
}

// 从 Reader 中读取一个整数，返回整数值
func readInt(r *bufio.Reader) (int, error) {
	var (
		sign    = 1    // 符号位
		val     = 0    // 累计值
		length  = 0    // 总读取长度
		initial = true // 初始字符阶段
	)

	for {
		b, err := r.ReadByte()
		if err != nil {
			if err == io.EOF {
				switch {
				case length == 0:
					return 0, io.ErrUnexpectedEOF // 完全无输入
				case initial && sign == -1:
					return 0, ErrMissingDigits // 只有负号
				default:
					return sign * val, nil // 正常结束
				}
			}
			return 0, err
		}
		length++

		switch {
		case initial && b == '-':
			sign = -1
			initial = false // 符号处理完成
			continue

		case initial && b == '+':
			initial = false // 显式正号
			continue

		case isDigit(b):
			initial = false
			val = val*10 + int(b-'0')

			// 批量预读优化
			if peekBytes, _ := r.Peek(r.Buffered()); len(peekBytes) > 0 {
				i := 0
				for ; i < len(peekBytes); i++ {
					if !isDigit(peekBytes[i]) {
						break
					}
					val = val*10 + int(peekBytes[i]-'0')
				}
				r.Discard(i)
				length += i
			}
			return sign * val, nil

		case initial: // 非数字初始字符
			r.UnreadByte()
			length--
			return 0, ErrInvalidIntFormat

		default: // 后续非数字字符
			r.UnreadByte()
			length--
			if length == 0 || (length == 1 && sign == -1) {
				return 0, ErrMissingDigits
			}
			return sign * val, nil // 正常结束
		}
	}
}
