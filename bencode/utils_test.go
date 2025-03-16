package bencode_test

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/Akimio521/torrent-go/bencode"
)

var (
	testCases = []struct {
		name  string
		input string
	}{
		{"Short", "12345"},
		{"Long", "12345678901234567890"},
		{"WithSign", "-9876543210"},
	}
)

// goos: darwin
// goarch: arm64
// pkg: github.com/Akimio521/torrent-go/bencode
// cpu: Apple M1
// BenchmarkReadInt
// BenchmarkReadInt/Short
// BenchmarkReadInt/Short-8         	 2597202	       478.6 ns/op	    4248 B/op	       4 allocs/op
// BenchmarkReadInt/Long
// BenchmarkReadInt/Long-8          	 2622433	       455.4 ns/op	    4264 B/op	       4 allocs/op
// BenchmarkReadInt/WithSign
// BenchmarkReadInt/WithSign-8      	 2528713	       467.2 ns/op	    4256 B/op	       4 allocs/op
// PASS
// ok  	github.com/Akimio521/torrent-go/bencode	5.898s
func BenchmarkReadInt(b *testing.B) {

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				r := bufio.NewReader(bytes.NewReader([]byte(tc.input)))
				_, _ = bencode.ExportReadInt(r)
			}
		})
	}
}

// goos: darwin
// goarch: arm64
// pkg: github.com/Akimio521/torrent-go/bencode
// cpu: Apple M1
// BenchmarkReadIntOld
// BenchmarkReadIntOld/Short
// BenchmarkReadIntOld/Short-8         	 2652002	       469.9 ns/op	    4248 B/op	       4 allocs/op
// BenchmarkReadIntOld/Long
// BenchmarkReadIntOld/Long-8          	 2508354	       483.6 ns/op	    4264 B/op	       4 allocs/op
// BenchmarkReadIntOld/WithSign
// BenchmarkReadIntOld/WithSign-8      	 2460427	       460.5 ns/op	    4256 B/op	       4 allocs/op
// PASS
// ok  	github.com/Akimio521/torrent-go/bencode	5.663s
// func BenchmarkReadIntOld(b *testing.B) {

// 	for _, tc := range testCases {
// 		b.Run(tc.name, func(b *testing.B) {
// 			for i := 0; i < b.N; i++ {
// 				r := bufio.NewReader(bytes.NewReader([]byte(tc.input)))
// 				_, _, _ = bencode.ExportReadIntOld(r)
// 			}
// 		})
// 	}
// }
