package torrent

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strconv"
	"time"
)

type PeerConn struct {
	net.Conn                   // 连接通道
	Choked   bool              // 对端上传是否被阻塞
	Field    Bitfield          // 对端的 Bitfield
	peer     PeerInfo          // 对端信息
	peerId   [PEER_ID_LEN]byte // 本地的 peerId
	infoSHA  [sha1.Size]byte   // 请求种子的 info 的 SHA-1 哈希
}

func handshake(conn net.Conn, infoSHA [sha1.Size]byte, peerId [PEER_ID_LEN]byte) error {
	conn.SetDeadline(time.Now().Add(3 * time.Second))
	defer conn.SetDeadline(time.Time{})
	// send HandshakeMsg
	reqMsg := NewHandShakeMsg(infoSHA, peerId)

	if err := reqMsg.WriteHandShakeMsg(conn); err != nil {
		return fmt.Errorf("send handshake failed: %s", err.Error())
	}
	// read HandshakeMsg
	respMsg, err := ReadHandshake(conn)
	if err != nil {
		return fmt.Errorf("read handshake failed: %s", err.Error())
	}
	// check HandshakeMsg
	if !bytes.Equal(respMsg.InfoSHA[:], infoSHA[:]) {
		return fmt.Errorf("check handshake hash failed: %s", string(respMsg.InfoSHA[:]))
	}
	return nil
}
func (peer PeerInfo) NewConn(infoSHA [sha1.Size]byte, peerId [PEER_ID_LEN]byte) (*PeerConn, error) {
	// setup tcp conn
	conn, err := net.DialTimeout("tcp", peer.GetConnAddr(), 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("set tcp conn to %s failed: %s", peer.GetConnAddr(), err.Error())
	}
	// torrent p2p handshake
	if err = handshake(conn, infoSHA, peerId); err != nil {
		conn.Close()
		return nil, fmt.Errorf("handshake failed: %s", err.Error())
	}
	c := &PeerConn{
		Conn:    conn,
		Choked:  true,
		peer:    peer,
		peerId:  peerId,
		infoSHA: infoSHA,
	}

	if err = c.GetBitfield(); err != nil {
		return nil, fmt.Errorf("fill bitfield failed: %s" + err.Error())
	}
	return c, nil
}

// 获取对端 Bitfield
func (c *PeerConn) GetBitfield() error {
	c.SetDeadline(time.Now().Add(5 * time.Second))
	defer c.SetDeadline(time.Time{})

	msg, err := c.ReadMsg()
	if err != nil {
		return fmt.Errorf("read PeerMsg faild: %s", err.Error())
	}
	if msg == nil {
		return fmt.Errorf("expected bitfield")
	}
	if msg.Id != MsgBitfield {
		return fmt.Errorf("expected bitfield, get " + strconv.Itoa(int(msg.Id)))
	}
	c.Field = msg.Payload
	return nil
}

// 读取消息
func (c *PeerConn) ReadMsg() (*PeerMsg, error) {
	// read msg length
	lenBuf := make([]byte, 4)
	_, err := io.ReadFull(c, lenBuf)
	if err != nil {
		return nil, err
	}
	length := binary.BigEndian.Uint32(lenBuf)
	// keep alive msg
	if length == 0 {
		return nil, nil
	}
	// read msg body
	msgBuf := make([]byte, length)
	_, err = io.ReadFull(c, msgBuf)
	if err != nil {
		return nil, err
	}
	return &PeerMsg{
		Id:      MsgId(msgBuf[0]),
		Payload: msgBuf[1:],
	}, nil
}

// 写入消息
func (c *PeerConn) WriteMsg(m *PeerMsg) (int, error) {
	length := uint32(len(m.Payload) + 1) // +1 for id
	buf := make([]byte, PEER_MSG_HEAD_LEN+length)
	binary.BigEndian.PutUint32(buf[0:PEER_MSG_HEAD_LEN], length)
	buf[PEER_MSG_HEAD_LEN] = byte(m.Id)
	copy(buf[PEER_MSG_HEAD_LEN+1:], m.Payload)
	return c.Write(buf)
}

// 将下载任务分配给该连接
func (conn *PeerConn) DownloadPiece(task *PieceTask) (*PieceResult, error) {
	state := &TaskState{
		Index: task.Index,
		Conn:  conn,
		Data:  make([]byte, task.Length),
	}
	conn.SetDeadline(time.Now().Add(15 * time.Second))
	defer conn.SetDeadline(time.Time{})

	for state.Downloaded < task.Length {
		if !conn.Choked {
			for state.Backlog < MAX_BACKLOG && state.Requested < task.Length { // 并发度未达到最大值，且请求量未达到任务长度
				length := BLOCK_SIZE
				if task.Length-state.Requested < length { //最后一片的长度可能小与 Block Size
					length = task.Length - state.Requested
				}
				msg := NewRequestMsg(state.Index, state.Requested, length)
				if _, err := state.Conn.WriteMsg(msg); err != nil {
					return nil, err
				}
				state.Backlog++
				state.Requested += length
			}
		}
		msg, err := state.Conn.ReadMsg()
		if err != nil {
			return nil, fmt.Errorf("read msg failed: %s", err.Error())
		}
		if err := state.HandleMsg(msg); err != nil {
			return nil, fmt.Errorf("handle msg failed: %s", err.Error())
		}
	}
	return &PieceResult{state.Index, state.Data}, nil
}
