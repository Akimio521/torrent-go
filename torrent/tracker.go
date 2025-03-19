package torrent

import (
	"encoding/binary"
	"fmt"
	"net"
	"net/url"
	"strconv"
)

type PeerInfo struct { // Peer 对端信息
	IP   net.IP // IP 地址
	Port uint16 // 端口号
}

// 获取连接地址
func (pi *PeerInfo) GetConnAddr() string {
	return net.JoinHostPort(pi.IP.String(), strconv.Itoa(int(pi.Port)))
}

type TrackerResponse struct { // Tracker 响应
	Interval int    `bencode:"interval"` // 间隔时间
	Peers    string `bencode:"peers"`    // Peer 列表
	Peers6   string `bencode:"peers6"`   // Peer 列表（IPv6）
}

// 解析 Tracker 响应中的 Peers 信息，返回 PeerInfo 列表
func (tr *TrackerResponse) ParsePeerInfos() ([]PeerInfo, error) {
	peers := []byte(tr.Peers)
	peers6 := []byte(tr.Peers6)

	// 校验数据完整性
	if len(peers)%PEER_V4_LEN != 0 {
		return nil, fmt.Errorf("%w: IPv4 peers length %d not divisible by %d",
			ErrMalformedPeersFormat, len(peers), PEER_V4_LEN)
	}
	if len(peers6)%PEER_V6_LEN != 0 {
		return nil, fmt.Errorf("%w: IPv6 peers length %d not divisible by %d",
			ErrMalformedPeersFormat, len(peers6), PEER_V6_LEN)
	}

	v4num := len(peers) / PEER_V4_LEN
	v6num := len(peers6) / PEER_V6_LEN

	// 正确初始化切片（容量预分配）
	peerInfos := make([]PeerInfo, 0, v4num+v6num)

	// 处理 IPv4 Peers
	for i := 0; i < v4num; i++ {
		offset := i * PEER_V4_LEN
		peerInfos = append(peerInfos, PeerInfo{
			IP:   net.IP(peers[offset : offset+net.IPv4len]),
			Port: binary.BigEndian.Uint16(peers[offset+net.IPv4len : offset+PEER_V4_LEN]),
		})
	}

	// 处理 IPv6 Peers
	for i := 0; i < v6num; i++ {
		offset := i * PEER_V6_LEN
		peerInfos = append(peerInfos, PeerInfo{
			IP:   net.IP(peers6[offset : offset+net.IPv6len]),
			Port: binary.BigEndian.Uint16(peers6[offset+net.IPv6len : offset+PEER_V6_LEN]),
		})
	}
	return peerInfos, nil
}

// 根据 TorrentFile 构建 Tracker 请求 URL
func buildUrl(tf *TorrentFile, peerID [PEER_ID_LEN]byte, port uint16) (string, error) {
	base, err := url.Parse(tf.Announce)
	if err != nil {
		return "", fmt.Errorf("parse URL %s error: %s", tf.Announce, err.Error())
	}
	infoHash := tf.GetInfoSHA1()
	params := url.Values{
		"info_hash":  []string{string(infoHash[:])},     // 文件唯一表示
		"peer_id":    []string{string(peerID[:])},       // 客户端唯一标识
		"port":       []string{strconv.Itoa(int(port))}, // 本机开放端口
		"uploaded":   []string{"0"},                     // 截至目前上传的总数，以十进制 ASCII 编码
		"downloaded": []string{"0"},                     // 截至目前下载的总数，以十进制 ASCII 编码
		"compact":    []string{"1"},
		"left":       []string{strconv.Itoa(tf.Info.Length)}, // 还需下载的字节数
	}

	base.RawQuery = params.Encode()
	return base.String(), nil
}
