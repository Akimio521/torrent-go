package torrent

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"sync/atomic"
)

type TorrentTask struct { // 种子任务
	FileName string            // 文件名
	FileLen  int               // 文件长度
	InfoSHA  [sha1.Size]byte   // 种子的 Info 的 SHA-1 哈希
	PeerList []PeerInfo        // Peer 列表
	PeerId   [20]byte          // 本地 Peer ID
	PieceLen int               // 每一块 Piece 的长度
	PieceSHA [][sha1.Size]byte // 所有 Piece 的 SHA-1 哈希值
}

func (t *TorrentTask) peerRoutine(peer PeerInfo, taskChan chan *PieceTask, ctx *Context) {
	// set up conn with peer
	conn, err := peer.NewConn(t.InfoSHA, t.PeerId)
	if err != nil {
		ctx.errChan <- fmt.Errorf("connect peer %s failed: %s", peer.IP.String(), err.Error())
		return
	}
	defer conn.Close()

	ctx.rwm.Lock()
	ctx.peerInfos = append(ctx.peerInfos, peer)
	ctx.rwm.Unlock()

	conn.WriteMsg(&PeerMsg{MsgInterested, nil})
	// get piece task & download
	for task := range taskChan {
		if !conn.Field.HasPiece(task.Index) {
			taskChan <- task
			continue
		}
		res, err := conn.DownloadPiece(task)
		if err != nil {
			taskChan <- task
			ctx.errChan <- fmt.Errorf("fail to download piece" + err.Error())
			return
		}
		if !task.CheckPiece(res) {
			taskChan <- task
			ctx.errChan <- fmt.Errorf("check piece failed")
			continue
		}
		ctx.resultChan <- res
		atomic.AddUint64(&ctx.currentBytes, uint64(len(res.Data)))
		atomic.AddUint64(&ctx.currentPieces, 1)
		if atomic.LoadUint64(&ctx.currentPieces) == uint64(len(t.PieceSHA)) {
			ctx.Finish()
		}
	}
}

// 获取 Piece 的起始和结束位置
func (t *TorrentTask) GetPieceBounds(index int) (bengin int, end int) {
	bengin = index * t.PieceLen
	end = bengin + t.PieceLen
	if end > t.FileLen {
		end = t.FileLen
	}
	return
}

// 下载种子任务
func (task *TorrentTask) Download() *Context {
	ctx := newContext()
	taskChan := make(chan *PieceTask, len(task.PieceSHA))

	for index, sha := range task.PieceSHA {
		begin, end := task.GetPieceBounds(index)
		taskChan <- &PieceTask{index, sha, (end - begin)}
	}
	// init goroutines for each peer
	for _, peer := range task.PeerList {
		go task.peerRoutine(peer, taskChan, ctx)
	}
	return ctx
}

type PieceTask struct { // Piece 任务
	Index  int             // 任务索引
	SHA1   [sha1.Size]byte // 该 Piece 的 SHA-1 哈希值
	Length int             // 该 Piece 长度（一般是默认，最后一个 Piece 可能较短）
}

// 检查下载的 PieceResult SHA-1 哈希值是否匹配
func (pt *PieceTask) CheckPiece(res *PieceResult) bool {
	sha := sha1.Sum(res.Data)
	return bytes.Equal(pt.SHA1[:], sha[:])
}

type TaskState struct { // 每一片 Piece 的任务状态
	Index      int       // Piece 的索引
	Conn       *PeerConn // 与对端 Peer 的连接
	Requested  int       // 请求量
	Downloaded int       // 下载量
	Backlog    int       // 并发度
	Data       []byte    // 数据
}

// 处理消息
func (ts *TaskState) HandleMsg(msg *PeerMsg) error {
	if msg == nil { // 保活消息
		return nil
	}
	switch msg.Id {
	case MsgChoke:
		ts.Conn.Choked = true
	case MsgUnchoke:
		ts.Conn.Choked = false
	case MsgHave:
		index, err := msg.GetHaveIndex()
		if err != nil {
			return err
		}
		ts.Conn.Field.SetPiece(index)
	case MsgPiece:
		n, err := msg.CopyPieceData(ts.Index, ts.Data)
		if err != nil {
			return err
		}
		ts.Downloaded += n
		ts.Backlog--
	}
	return nil
}

type PieceResult struct { // Piece 结果
	Index int    // Piece 的索引
	Data  []byte // 数据
}
