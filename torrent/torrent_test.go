package torrent_test

import (
	"bufio"
	"os"
	"testing"

	"github.com/Akimio521/torrent-go/torrent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseFile(t *testing.T) {
	file, err := os.Open("./../test_files/debian-iso.torrent")
	assert.Equal(t, nil, err)
	defer file.Close()
	tf, err := torrent.ParseFile(bufio.NewReader(file))
	require.NoError(t, err)
	require.Equal(t, nil, err)
	require.Equal(t, "http://bttracker.debian.org:6969/announce", tf.Announce)
	require.Equal(t, "debian-11.2.0-amd64-netinst.iso", tf.Info.Name)
	require.Equal(t, 396361728, tf.Info.Length)
	require.Equal(t, 262144, tf.Info.PiceLength)
	require.Equal(t, 1512, len(tf.Info.Pieces)/20)
	var expectHASH = [20]byte{0x28, 0xc5, 0x51, 0x96, 0xf5, 0x77, 0x53, 0xc4, 0xa,
		0xce, 0xb6, 0xfb, 0x58, 0x61, 0x7e, 0x69, 0x95, 0xa7, 0xed, 0xdb}
	require.Equal(t, expectHASH, tf.GetInfoSHA1())
}

func BenchmarkParseFile(b *testing.B) {
	file, err := os.Open("./../test_files/debian-iso.torrent")
	assert.Equal(b, nil, err)
	defer file.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		file.Seek(0, 0)
		torrent.ParseFile(file)
	}
}
