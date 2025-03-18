package torrent

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io"
	"net/http"
	"time"

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
	Announce     string          `bencode:"announce"`      // 首选 tracker 地址 必选
	Info         RawInfo         `bencode:"info"`          // 文件信息 必选
	AnnounceList []string        `bencode:"announce-list"` // 备选 tracker 列表 可选
	Comment      string          `bencode:"comment"`       // 备注 可选
	CreatBy      string          `bencode:"created by"`    // 创建者信息 可选
	infoSHA1     [sha1.Size]byte `bencode:"-"`             // 用于存储 Info 的哈希（种子的唯一标识）
}

// 获取种子的唯一标识（string）
func (tf *TorrentFile) GetInfoSHA1() [sha1.Size]byte {
	return tf.infoSHA1
}

// 获取所有的 SHA1 值
func (tf *TorrentFile) GetAllPieceSHA() [][sha1.Size]byte {
	pieces := []byte(tf.Info.Pieces)
	num := tf.GetPiecesNum()
	hashes := make([][sha1.Size]byte, num)
	for i := 0; i < num; i++ {
		copy(hashes[i][:], pieces[i*sha1.Size:(i+1)*sha1.Size])
	}
	return hashes
}

// 计算 Pieces 的数量
func (tf *TorrentFile) GetPiecesNum() int {
	return len(tf.Info.Pieces) / sha1.Size
}

// 向 TorrentFile 中的 Tracker 发送请求获取 Peer 列表
func (tf *TorrentFile) FindPeers(peerID [PEER_ID_LEN]byte, port uint16) ([]PeerInfo, error) {
	url, err := buildUrl(tf, peerID, port)
	if err != nil {
		return nil, fmt.Errorf("build tracker URL error: %s", err.Error())
	}
	// fmt.Println(url)

	client := http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fail to connect to tracker: %s", err.Error())
	}
	defer resp.Body.Close()

	trackerResp := new(TrackerResponse)
	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read tracker response error: %s", err.Error())
	}
	if err = bencode.Unmarshal(bytes.NewBuffer(bs), trackerResp); err != nil {
		return nil, fmt.Errorf("unmarshal tracker response error: %s", err.Error())
	}

	return trackerResp.ParsePeerInfos()
}

// 获取种子文件转的任务
func (tf *TorrentFile) GetTask(peerID [PEER_ID_LEN]byte, port uint16) (*TorrentTask, error) {
	peers, err := tf.FindPeers(peerID, port)
	if err != nil {
		return nil, fmt.Errorf("find peers faild: %s", err.Error())
	}
	if len(peers) == 0 {
		return nil, fmt.Errorf("can not find peers")
	}
	return &TorrentTask{
		PeerId:   peerID,
		PeerList: peers,
		InfoSHA:  tf.GetInfoSHA1(),
		FileName: tf.Info.Name,
		FileLen:  tf.Info.Length,
		PieceLen: tf.Info.PiceLength,
		PieceSHA: tf.GetAllPieceSHA(),
	}, nil
}

// 解析种子文件
func ParseFile(r io.Reader) (*TorrentFile, error) {
	tf := new(TorrentFile)
	bObj, err := bencode.Parse(r)
	if err != nil {
		return nil, err
	}
	if err = bencode.UnmarshalBObject(bObj, tf); err != nil {
		return nil, err
	}
	info, err := bObj.GetDictKeyDay("info")
	if err != nil {
		return nil, err
	}
	tf.infoSHA1 = sha1.Sum(info)
	return tf, err
}
