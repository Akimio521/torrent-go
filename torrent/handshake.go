package torrent

import (
	"bufio"
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

func (msg *HandshakeMsg) WriteHandShakeMsg(w io.Writer) error {
	bw := bufio.NewWriterSize(w, 1+len(msg.PreStr)+HS_MSG_LEN) // 1 byte for prelen
	defer bw.Flush()
	if err := bw.WriteByte(byte(len(msg.PreStr))); err != nil { // prelen
		return err
	}
	if _, err := bw.WriteString(msg.PreStr); err != nil { // 协议 prestr
		return err
	}
	if _, err := bw.Write(make([]byte, RESERVED_LEN)); err != nil { // 保留字段
		return err
	}
	if _, err := bw.Write(msg.InfoSHA[:]); err != nil { // infoSHA
		return err
	}
	if _, err := bw.Write(msg.PeerId[:]); err != nil { // peerId
		return err
	}
	return nil
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
