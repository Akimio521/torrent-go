package torrent

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"os"
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

func (t *TorrentTask) peerRoutine(peer PeerInfo, taskQueue chan *PieceTask, resultQueue chan *PieceResult) {
	// set up conn with peer
	conn, err := peer.NewConn(t.InfoSHA, t.PeerId)
	if err != nil {
		fmt.Printf("connect peer %s failed: %s\n", peer.IP.String(), err.Error())
		return
	}
	defer conn.Close()

	fmt.Println("complete handshake with peer : " + peer.IP.String())
	conn.WriteMsg(&PeerMsg{MsgInterested, nil})
	// get piece task & download
	for task := range taskQueue {
		if !conn.Field.HasPiece(task.Index) {
			taskQueue <- task
			continue
		}
		fmt.Printf("get task, index: %v, peer : %v\n", task.Index, peer.IP.String())
		res, err := conn.DownloadPiece(task)
		if err != nil {
			taskQueue <- task
			fmt.Println("fail to download piece" + err.Error())
			return
		}
		if !task.CheckPiece(res) {
			taskQueue <- task
			continue
		}
		resultQueue <- res
	}
}

// 获取 Piece 的起始和结束位置
func (t *TorrentTask) getPieceBounds(index int) (bengin, end int) {
	bengin = index * t.PieceLen
	end = bengin + t.PieceLen
	if end > t.FileLen {
		end = t.FileLen
	}
	return
}

// 下载种子任务
func (task *TorrentTask) Download() error {
	fmt.Println("start downloading " + task.FileName)
	// split pieceTasks and init task&result channel
	taskChan := make(chan *PieceTask, len(task.PieceSHA))
	resultChan := make(chan *PieceResult)
	for index, sha := range task.PieceSHA {
		begin, end := task.getPieceBounds(index)
		taskChan <- &PieceTask{index, sha, (end - begin)}
	}
	// init goroutines for each peer
	for _, peer := range task.PeerList {
		go task.peerRoutine(peer, taskChan, resultChan)
	}
	// collect piece result
	buf := make([]byte, task.FileLen)
	count := 0
	for count < len(task.PieceSHA) {
		res := <-resultChan
		begin, end := task.getPieceBounds(res.Index)
		copy(buf[begin:end], res.Data)
		count++
		// print progress
		percent := float64(count) / float64(len(task.PieceSHA)) * 100
		fmt.Printf("downloading, progress : (%0.2f%%)\n", percent)
	}
	close(taskChan)
	close(resultChan)
	// create file & copy data
	file, err := os.Create(task.FileName)
	if err != nil {
		return fmt.Errorf("create %s faild: %s", task.FileName, err.Error())
	}
	defer file.Close()

	if _, err = file.Write(buf); err != nil {
		return fmt.Errorf("write data faild: %s", err.Error())
	}
	return nil
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
