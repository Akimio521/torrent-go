package bencode_test

import (
	"bytes"
	"testing"

	"github.com/Akimio521/torrent-go/bencode"
	"github.com/stretchr/testify/require"
)

type User struct {
	Name string `bencode:"name"`
	Age  int    `bencode:"age"`
}

type Role struct {
	Id   int
	User `bencode:"user"`
}

type Score struct {
	User  `bencode:"user"`
	Value []int `bencode:"value"`
}

type Team struct {
	Name   string `bencode:"name"`
	Size   int    `bencode:"size"`
	Member []User `bencode:"member"`
}

func TestMarshalBasic(t *testing.T) {
	buf := new(bytes.Buffer)
	str := "abc"
	len, err := bencode.Marshal(buf, str)
	require.NoError(t, err)
	require.Equal(t, 5, len)
	require.Equal(t, "3:abc", buf.String())

	buf.Reset()
	val := 199
	len, err = bencode.Marshal(buf, val)
	require.NoError(t, err)
	require.Equal(t, 5, len)
	require.Equal(t, "i199e", buf.String())
}

func TestUnmarshalList(t *testing.T) {
	str := "li85ei90ei95ee"
	l := &[]int{}
	err := bencode.Unmarshal(bytes.NewBufferString(str), l)
	require.NoError(t, err)
	require.Equal(t, []int{85, 90, 95}, *l)

	buf := new(bytes.Buffer)
	length, err := bencode.Marshal(buf, l)
	require.NoError(t, err)
	require.Equal(t, len(str), length)
	require.Equal(t, str, buf.String())
}

func TestUnmarshalUser(t *testing.T) {
	str := "d4:name6:archer3:agei29ee"
	u := &User{}
	bencode.Unmarshal(bytes.NewBufferString(str), u)
	require.Equal(t, "archer", u.Name)
	require.Equal(t, 29, u.Age)

	buf := new(bytes.Buffer)
	length, err := bencode.Marshal(buf, u)
	require.NoError(t, err)
	require.Equal(t, len(str), length)
	require.Equal(t, str, buf.String())
}

func TestUnmarshalRole(t *testing.T) {
	str := "d2:idi1e4:userd4:name6:archer3:agei29eee"
	r := &Role{}
	err := bencode.Unmarshal(bytes.NewBufferString(str), r)
	require.NoError(t, err)
	require.Equal(t, 1, r.Id)
	require.Equal(t, "archer", r.Name)
	require.Equal(t, 29, r.Age)

	buf := new(bytes.Buffer)
	length, err := bencode.Marshal(buf, r)
	require.NoError(t, err)
	require.Equal(t, len(str), length)
	require.Equal(t, str, buf.String())
}

func TestUnmarshalScore(t *testing.T) {
	str := "d4:userd4:name6:archer3:agei29ee5:valueli80ei85ei90eee"
	s := &Score{}
	err := bencode.Unmarshal(bytes.NewBufferString(str), s)
	require.NoError(t, err)
	require.Equal(t, "archer", s.Name)
	require.Equal(t, 29, s.Age)
	require.Equal(t, []int{80, 85, 90}, s.Value)

	buf := new(bytes.Buffer)
	length, err := bencode.Marshal(buf, s)
	require.NoError(t, err)
	require.Equal(t, len(str), length)
	require.Equal(t, str, buf.String())
}

func TestUnmarshalTeam(t *testing.T) {
	str := "d4:name3:ace4:sizei2e6:memberld4:name6:archer3:agei29eed4:name5:nancy3:agei31eeee"
	team := &Team{}
	err := bencode.Unmarshal(bytes.NewBufferString(str), team)
	require.NoError(t, err)
	require.Equal(t, "ace", team.Name)
	require.Equal(t, 2, team.Size)

	buf := new(bytes.Buffer)
	length, err := bencode.Marshal(buf, team)
	require.NoError(t, err)
	require.Equal(t, len(str), length)
	require.Equal(t, str, buf.String())
}
