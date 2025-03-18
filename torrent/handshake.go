package torrent

import (
	"crypto/sha1"
	"io"
)

type HandshakeMsg struct {
	PreStr  string            // 协议
	InfoSHA [sha1.Size]byte   // 种子的 info 的 SHA-1 哈希
	PeerId  [PEER_ID_LEN]byte // 本地的 peerId
}

func NewHandShakeMsg(infoSHA, peerId [PEER_ID_LEN]byte) *HandshakeMsg {
	return &HandshakeMsg{
		PreStr:  "BitTorrent protocol",
		InfoSHA: infoSHA,
		PeerId:  peerId,
	}
}

func (msg *HandshakeMsg) WriteHandShakeMsg(w io.Writer) (int, error) {
	buf := make([]byte, len(msg.PreStr)+HS_MSG_LEN+1) // 1 byte for prelen
	buf[0] = byte(len(msg.PreStr))
	curr := 1
	curr += copy(buf[curr:], []byte(msg.PreStr))
	curr += copy(buf[curr:], make([]byte, RESERVED_LEN))
	curr += copy(buf[curr:], msg.InfoSHA[:])
	curr += copy(buf[curr:], msg.PeerId[:])
	return w.Write(buf)
}

func ReadHandshake(r io.Reader) (*HandshakeMsg, error) {
	lenBuf := make([]byte, 1)
	_, err := io.ReadFull(r, lenBuf)
	if err != nil {
		return nil, err
	}
	prelen := int(lenBuf[0])

	if prelen == 0 {
		return nil, ErrZeroPrelen
	}

	msgBuf := make([]byte, HS_MSG_LEN+prelen)
	if _, err = io.ReadFull(r, msgBuf); err != nil {
		return nil, err
	}

	var peerId [PEER_ID_LEN]byte
	var infoSHA [sha1.Size]byte

	copy(infoSHA[:], msgBuf[prelen+RESERVED_LEN:prelen+RESERVED_LEN+sha1.Size])
	copy(peerId[:], msgBuf[prelen+RESERVED_LEN+sha1.Size:])

	return &HandshakeMsg{
		PreStr:  string(msgBuf[0:prelen]),
		InfoSHA: infoSHA,
		PeerId:  peerId,
	}, nil
}
