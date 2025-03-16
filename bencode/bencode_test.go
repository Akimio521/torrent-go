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

// ---------------------- 字符串测试 ----------------------
type StringEncodeCase struct {
	Raw     string
	Encoded string
	ByteLen int
	Err     error
}

type StringDecodeCase struct {
	Encoded     string
	ExpectedRaw string
	Err         error
}

var stringEncodeTests = []StringEncodeCase{
	{Raw: "abc", Encoded: "3:abc", ByteLen: 5},
	{Raw: "", Encoded: "0:", ByteLen: 2},
	{Raw: "invalid:test", Encoded: "12:invalid:test", ByteLen: 15},
}

var stringDecodeTests = []StringDecodeCase{
	{Encoded: "3:abc", ExpectedRaw: "abc"},
	{Encoded: "0:", ExpectedRaw: ""},
	{Encoded: "5:12345", ExpectedRaw: "12345"},
	{Encoded: "12:invalid:test", ExpectedRaw: "invalid:test"},
	{Encoded: "20:FGHIJKLMNOPQRSTUVWXY", ExpectedRaw: "FGHIJKLMNOPQRSTUVWXY"},
	// 错误用例
	{Encoded: "3ab", Err: bencode.ErrInvalidStringFormat},  // 缺少冒号
	{Encoded: "-5:abc", Err: bencode.ErrStringLength},      // 负长度
	{Encoded: "2abc", Err: bencode.ErrInvalidStringFormat}, // 无效长度前缀
}

func TestEncodeString(t *testing.T) {
	for _, tc := range stringEncodeTests {
		t.Run(fmt.Sprintf("Encode/%s", tc.Raw), func(t *testing.T) {
			buf := new(bytes.Buffer)
			encodedLen, err := bencode.EncodeString(buf, tc.Raw)

			if tc.Err != nil {
				require.ErrorIs(t, err, tc.Err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.ByteLen, encodedLen)
			require.Equal(t, tc.Encoded, buf.String())
		})
	}
}

func TestDecodeString(t *testing.T) {
	for _, tc := range stringDecodeTests {
		t.Run(fmt.Sprintf("Decode/%s", tc.Encoded), func(t *testing.T) {
			r := bytes.NewReader([]byte(tc.Encoded))
			decoded, err := bencode.DecodeString(r)

			if tc.Err != nil {
				require.ErrorIs(t, err, tc.Err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.ExpectedRaw, decoded)
		})
	}
}

// ---------------------- 整数测试 ----------------------
type IntEncodeCase struct {
	Raw     int
	Encoded string
	ByteLen int
}

type IntDecodeCase struct {
	Encoded     string
	ExpectedRaw int
	Err         error
}

var intEncodeTests = []IntEncodeCase{
	{Raw: 999, Encoded: "i999e", ByteLen: 5},
	{Raw: 0, Encoded: "i0e", ByteLen: 3},
	{Raw: -42, Encoded: "i-42e", ByteLen: 5},
	{Raw: math.MaxInt64, Encoded: fmt.Sprintf("i%de", math.MaxInt64), ByteLen: len(fmt.Sprintf("i%de", math.MaxInt64))},
}

var intDecodeTests = []IntDecodeCase{
	{Encoded: "i999e", ExpectedRaw: 999},
	{Encoded: "i0e", ExpectedRaw: 0},
	{Encoded: "i-42e", ExpectedRaw: -42},
	{Encoded: fmt.Sprintf("i%de", math.MaxInt64), ExpectedRaw: math.MaxInt64},
	// 错误用例
	{Encoded: "i1a0e", Err: bencode.ErrInvalidIntFormat}, // 非法字符
	{Encoded: "ie", Err: bencode.ErrInvalidIntFormat},    // 空数字
	{Encoded: "i123", Err: io.ErrUnexpectedEOF},          // 缺少终止符
	{Encoded: "123e", Err: bencode.ErrInvalidIntFormat},  // 缺少前缀i
}

func TestEncodeInt(t *testing.T) {
	for _, tc := range intEncodeTests {
		t.Run(fmt.Sprintf("Encode/%d", tc.Raw), func(t *testing.T) {
			buf := new(bytes.Buffer)
			encodedLen, err := bencode.EncodeInt(buf, tc.Raw)

			require.NoError(t, err)
			require.Equal(t, tc.ByteLen, encodedLen)
			require.Equal(t, tc.Encoded, buf.String())
		})
	}
}

func TestDecodeInt(t *testing.T) {
	for _, tc := range intDecodeTests {
		t.Run(fmt.Sprintf("Decode/%s", tc.Encoded), func(t *testing.T) {
			r := bytes.NewReader([]byte(tc.Encoded))
			decoded, err := bencode.DecodeInt(r)

			if tc.Err != nil {
				require.ErrorIs(t, err, tc.Err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.ExpectedRaw, decoded)
		})
	}
}
