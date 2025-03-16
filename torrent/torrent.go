package torrent

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io"

	"github.com/Akimio521/torrent-go/bencode"
)

type Files struct { // 文件信息
	Path   []string `bencode:"path"`   // 文件路径
	Length int      `bencode:"length"` // 文件大小
}

type RawInfo struct {
	Name       string  `bencode:"name"`         // 文件/目录名
	Length     int     `bencode:"length"`       // 文件/目录大小
	PiceLength int     `bencode:"piece length"` // 每个 piece 的大小
	Pieces     string  `bencode:"pieces"`       // 所有 piece 的 hash 值
	Files      []Files `bencode:"files"`        // 文件列表（当种子是目录时不为空）
}

type TorrentFile struct {
	Announce     string       `bencode:"announce"`      // 首选 tracker 地址 必选
	Info         RawInfo      `bencode:"info"`          // 文件信息 必选
	AnnounceList []string     `bencode:"announce-list"` // 备选 tracker 列表 可选
	Comment      string       `bencode:"comment"`       // 备注 可选
	CreatBy      string       `bencode:"created by"`    // 创建者信息 可选
	infoSHA1     [SHALEN]byte `bencode:"-"`             // 用于存储 Info 的哈希
	pieces       []byte       `bencode:"-"`             // Info.Pieces 的 []byte
}

var zeroSHA1 [20]byte // 零值
// 获取种子的唯一标识
func (tf *TorrentFile) GetInfoSHA1() ([SHALEN]byte, error) {
	if tf.infoSHA1 == zeroSHA1 {
		buf := new(bytes.Buffer)
		_, err := bencode.Marshal(buf, tf.Info)
		if err != nil {
			fmt.Println("Fail to marshal raw file info")
			return zeroSHA1, err
		}
		tf.infoSHA1 = sha1.Sum(buf.Bytes())
	}
	return tf.infoSHA1, nil
}

// 获取指定区间 的 SHA1 值
func (tf *TorrentFile) GetPieceSHA(num uint) [SHALEN]byte {
	if tf.pieces == nil {
		tf.pieces = []byte(tf.Info.Pieces)
	}
	var hash [SHALEN]byte
	copy(hash[:], tf.pieces[num*SHALEN:(num+1)*SHALEN])
	return hash
}

// 计算 PieceSHA 的数量
func (tf *TorrentFile) GetPieceLen() uint {
	if tf.pieces == nil {
		tf.pieces = []byte(tf.Info.Pieces)
	}
	return uint(len(tf.pieces) / SHALEN)
}

// 解析种子文件
func ParseFile(r io.Reader) (*TorrentFile, error) {
	tf := new(TorrentFile)
	err := bencode.Unmarshal(r, tf)
	return tf, err
}
