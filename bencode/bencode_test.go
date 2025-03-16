package bencode_test

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"testing"

	"github.com/Akimio521/torrent-go/bencode"
	"github.com/stretchr/testify/require"
)

type StringTestCase struct {
	Raw     string
	Encoded string
	ByteLen int
	Err     error
}

var stringTestCases = []StringTestCase{
	{Raw: "abc", Encoded: "3:abc", ByteLen: 5},
	{Raw: "", Encoded: "0:", ByteLen: 2},
	{Raw: "12345", Encoded: "5:12345", ByteLen: 7},
	{Raw: "invalid:test", Encoded: "12:invalid:test", ByteLen: 15},
	{Raw: "FGHIJKLMNOPQRSTUVWXY", Encoded: "20:FGHIJKLMNOPQRSTUVWXY", ByteLen: 23},
	{Raw: "Z", Encoded: "1:Z", ByteLen: 3},
}

func TestString(t *testing.T) {

	for _, tc := range stringTestCases {
		t.Run(tc.Raw, func(t *testing.T) {
			buf := new(bytes.Buffer)
			encodedLenth, err := bencode.EncodeString(buf, tc.Raw)
			if tc.Err != nil {
				require.Equal(t, tc.Err, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.ByteLen, encodedLenth)

			require.Equal(t, tc.Encoded, buf.String())

			decodedStr, err := bencode.DecodeString(buf)
			require.NoError(t, err)
			require.Equal(t, tc.Raw, decodedStr)
		})
	}
}

type IntTestCase struct {
	Raw     int    // 原始整数值
	Encoded string // 编码后字符串
	ByteLen int    // 编码字节长度
	Err     error  // 预期错误
}

var intTestCases = []IntTestCase{
	// 正常用例
	{Raw: 999, Encoded: "i999e", ByteLen: 5},
	{Raw: 0, Encoded: "i0e", ByteLen: 3},
	{Raw: -42, Encoded: "i-42e", ByteLen: 5},
	{Raw: 123456789, Encoded: "i123456789e", ByteLen: 11},

	// 边界值测试
	{Raw: math.MaxInt64, Encoded: fmt.Sprintf("i%de", math.MaxInt64), ByteLen: len(fmt.Sprintf("i%de", math.MaxInt64))},
	{Raw: math.MinInt64, Encoded: fmt.Sprintf("i%de", math.MinInt64), ByteLen: len(fmt.Sprintf("i%de", math.MinInt64))},

	// 错误用例
	{Raw: 100, Encoded: "i1a0e", Err: bencode.ErrInvalidIntFormat}, // 非法字符
	{Raw: 0, Encoded: "ie", Err: bencode.ErrInvalidIntFormat},      // 空数字
	{Raw: 0, Encoded: "i123", Err: io.EOF},                         // 缺少终止符
	{Raw: 0, Encoded: "123e", Err: bencode.ErrInvalidIntFormat},    // 缺少i前缀
}

func TestInt(t *testing.T) {
	for _, tc := range intTestCases {
		t.Run(tc.Encoded, func(t *testing.T) {
			// 编码测试
			if tc.Err == nil {
				buf := new(bytes.Buffer)
				encodedLength, err := bencode.EncodeInt(buf, tc.Raw)

				require.NoError(t, err)
				require.Equal(t, tc.ByteLen, encodedLength)
				require.Equal(t, tc.Encoded, buf.String())
			}

			// 解码测试（包含错误用例）
			if tc.Encoded != "" {
				reader := bytes.NewReader([]byte(tc.Encoded))
				decodedVal, err := bencode.DecodeInt(reader)

				if tc.Err != nil {
					require.ErrorIs(t, err, tc.Err)
					return
				}

				require.NoError(t, err)
				require.Equal(t, tc.Raw, decodedVal)
			}
		})
	}
}
