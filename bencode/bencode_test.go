package bencode_test

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"strings"
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
	{Encoded: "i123", Err: io.EOF},                       // 缺少终止符
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

// ---------------------- BObject 编码测试 ----------------------
func TestBObject_Bencode(t *testing.T) {
	// 测试用例结构
	testCases := []struct {
		name     string
		obj      *bencode.BObject
		expected string
		wantLen  int
		wantErr  error
	}{
		// -------------------- 字符串类型测试 --------------------
		{
			name:     "BSTR normal",
			obj:      bencode.GetBOject("hello"),
			expected: "5:hello",
			wantLen:  7, // len("5:hello") = 6 + 1 (bufio flush?)
		},
		{
			name:     "BSTR empty",
			obj:      bencode.GetBOject(""),
			expected: "0:",
			wantLen:  2,
		},

		// -------------------- 整数类型测试 --------------------
		{
			name:     "BINT positive",
			obj:      bencode.GetBOject(42),
			expected: "i42e",
			wantLen:  4,
		},
		{
			name:     "BINT zero",
			obj:      bencode.GetBOject(0),
			expected: "i0e",
			wantLen:  3,
		},
		{
			name:     "BINT negative",
			obj:      bencode.GetBOject(-123),
			expected: "i-123e",
			wantLen:  6,
		},

		// -------------------- 列表类型测试 --------------------
		{
			name:     "BLIST empty",
			obj:      bencode.GetBOject([]*bencode.BObject{}),
			expected: "le",
			wantLen:  2,
		},
		{
			name: "BLIST mixed types",
			obj: bencode.GetBOject([]*bencode.BObject{
				bencode.GetBOject("abc"),
				bencode.GetBOject(123),
			}),
			expected: "l3:abci123ee",
			wantLen:  len("l3:abci123ee"),
		},

		// -------------------- 字典类型测试 --------------------
		{
			name:     "BDICT empty",
			obj:      bencode.GetBOject(map[string]*bencode.BObject{}),
			expected: "de",
			wantLen:  2,
		},
		{
			name: "BDICT sorted keys",
			obj: bencode.GetBOject(map[string]*bencode.BObject{
				"a": bencode.GetBOject("test"),
				"b": bencode.GetBOject(1),
			}),
			expected: "d1:a4:test1:zi1ee", // 按键排序后 a -> z
			wantLen:  len("d1:a4:test1:zi1ee"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			gotLen, err := tc.obj.Bencode(buf)

			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.wantLen, gotLen)
			require.Equal(t, tc.expected, buf.String())
		})
	}
}

// 辅助函数创建复杂对象
func createNestedList() *bencode.BObject {
	return bencode.GetBOject(
		[]*bencode.BObject{
			bencode.GetBOject(
				map[string]*bencode.BObject{
					"id": bencode.GetBOject(123),
				},
			),
		},
	)
}

func TestNestedStructures(t *testing.T) {
	// 测试嵌套结构
	testCases := []struct {
		name     string
		obj      *bencode.BObject
		expected string
	}{
		{
			name:     "Nested list-dict",
			obj:      createNestedList(),
			expected: "ld2:idi123eee",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			_, err := tc.obj.Bencode(buf)
			require.NoError(t, err)
			require.Equal(t, tc.expected, buf.String())
		})
	}
}

func TestParse(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected func() *bencode.BObject
		wantErr  error
	}{
		// -------------------- 基础类型测试 --------------------
		{
			name:  "BSTR basic",
			input: "3:foo",
			expected: func() *bencode.BObject {
				return bencode.GetBOject("foo")
			},
		},
		{
			name:  "BINT positive",
			input: "i42e",
			expected: func() *bencode.BObject {
				return bencode.GetBOject(42)
			},
		},
		{
			name:  "BINT negative",
			input: "i-123e",
			expected: func() *bencode.BObject {
				return bencode.GetBOject(-123)
			},
		},
		{
			name:  "BINT zero",
			input: "i0e",
			expected: func() *bencode.BObject {
				return bencode.GetBOject(0)
			},
		},

		// -------------------- 复合类型测试 --------------------
		{
			name:  "BLIST empty",
			input: "le",
			expected: func() *bencode.BObject {
				return bencode.GetBOject([]*bencode.BObject{})
			},
		},
		{
			name:  "BLIST nested",
			input: "li42el3:abcee",
			expected: func() *bencode.BObject {
				return bencode.GetBOject([]*bencode.BObject{
					bencode.GetBOject(42),
					bencode.GetBOject([]*bencode.BObject{
						bencode.GetBOject("abc"),
					}),
				})
			},
		},
		{
			name:  "BDICT ordered keys",
			input: "d1:ai1e1:zi2ee",
			expected: func() *bencode.BObject {
				return bencode.GetBOject(map[string]*bencode.BObject{
					"a": bencode.GetBOject(1),
					"z": bencode.GetBOject(2),
				})
			},
		},
		{
			name:  "BDICT nested",
			input: "d4:infod6:lengthi1024e5:filesld4:pathl3:dir8:file.txteeeee",
			expected: func() *bencode.BObject {
				return bencode.GetBOject(map[string]*bencode.BObject{
					"info": bencode.GetBOject(map[string]*bencode.BObject{
						"length": bencode.GetBOject(1024),
						"files": bencode.GetBOject([]*bencode.BObject{
							bencode.GetBOject(map[string]*bencode.BObject{
								"path": bencode.GetBOject([]*bencode.BObject{
									bencode.GetBOject("dir"),
									bencode.GetBOject("file.txt"),
								}),
							}),
						}),
					}),
				})
			},
		},

		// -------------------- 错误场景测试 --------------------
		{
			name:    "Invalid prefix",
			input:   "x123",
			wantErr: bencode.ErrInvalidBObject,
		},
		{
			name:    "Truncated integer",
			input:   "i123",
			wantErr: io.ErrUnexpectedEOF,
		},
		{
			name:    "Invalid string length",
			input:   "5:abc",
			wantErr: io.ErrUnexpectedEOF,
		},
		{
			name:  "Unordered dict keys",
			input: "d1:zi2e1:ai1ee",
			expected: func() *bencode.BObject {
				return bencode.GetBOject(map[string]*bencode.BObject{
					"a": bencode.GetBOject(1),
					"z": bencode.GetBOject(2),
				})
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 执行解析
			r := bytes.NewReader([]byte(tc.input))
			obj, err := bencode.Parse(r)

			// 错误处理
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)

			// 验证结果
			expectedObj := tc.expected()
			validateBObject(t, expectedObj, obj)
		})
	}
}

// ---------------------- 验证工具函数 ----------------------
func validateBObject(t *testing.T, expected, actual *bencode.BObject) {
	require.Equal(t, expected.GetBType(), actual.GetBType())

	switch expected.GetBType() {
	case bencode.BSTR:
		ev, _ := bencode.Get[string](expected)
		av, _ := bencode.Get[string](actual)
		require.Equal(t, ev, av)

	case bencode.BINT:
		ev, _ := bencode.Get[int](expected)
		av, _ := bencode.Get[int](actual)
		require.Equal(t, ev, av)

	case bencode.BLIST:
		elist, _ := bencode.Get[[]*bencode.BObject](expected)
		alist, _ := bencode.Get[[]*bencode.BObject](actual)
		require.Equal(t, len(elist), len(alist))
		for i := range elist {
			validateBObject(t, elist[i], alist[i])
		}

	case bencode.BDICT:
		edict, _ := bencode.Get[map[string]*bencode.BObject](expected)
		adict, _ := bencode.Get[map[string]*bencode.BObject](actual)
		require.Equal(t, len(edict), len(adict))
		for k, ev := range edict {
			av, exists := adict[k]
			require.True(t, exists)
			validateBObject(t, ev, av)
		}
	}
}

// ---------------------- 边界条件测试 ----------------------
func TestEdgeCases(t *testing.T) {
	t.Run("Empty input", func(t *testing.T) {
		_, err := bencode.Parse(bytes.NewReader(nil))
		require.ErrorIs(t, err, io.EOF)
	})

	t.Run("Huge integer", func(t *testing.T) {
		input := "i9223372036854775807e" // math.MaxInt64
		obj, err := bencode.Parse(strings.NewReader(input))
		require.NoError(t, err)
		val, err := bencode.Get[int](obj)
		require.NoError(t, err)
		require.Equal(t, 9223372036854775807, val)
	})

	t.Run("Nested depth limit", func(t *testing.T) {
		deepNested := strings.Repeat("l", 1000) + strings.Repeat("e", 1000)
		_, err := bencode.Parse(strings.NewReader(deepNested))
		require.NoError(t, err) // 验证无栈溢出
	})
}

// ---------------------- 性能测试 ----------------------
// goos: darwin
// goarch: arm64
// pkg: github.com/Akimio521/torrent-go/bencode
// cpu: Apple M1
// BenchmarkParse
// BenchmarkParse/simple
// BenchmarkParse/simple-8         	 2541420	       493.4 ns/op	    4248 B/op	       4 allocs/op
// BenchmarkParse/string
// BenchmarkParse/string-8         	 2355640	       502.6 ns/op	    4296 B/op	       7 allocs/op
// BenchmarkParse/complex
// BenchmarkParse/complex-8        	  975999	      1227 ns/op	    5376 B/op	      37 allocs/op
// PASS
func BenchmarkParse(b *testing.B) {
	testData := map[string]string{
		"simple":  "i42e",
		"string":  "11:hello_world",
		"complex": "d4:infod6:lengthi1024e5:filesld4:pathl3:dir8:file.txteeeee",
	}

	for name, data := range testData {
		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				r := strings.NewReader(data)
				_, _ = bencode.Parse(r)
			}
		})
	}
}
