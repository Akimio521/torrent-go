package torrent

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

type Context struct {
	doneChan          chan struct{}     // 任务完成的通知
	resultChan        chan *PieceResult // 下载 Piece 结果通道
	errChan           chan error        // 错误通知通道
	rwm               sync.RWMutex      // peerInfos 读写锁
	peerInfos         []PeerInfo        // 正在下载的 Peer 列表
	successedBtye     uint64            // 已下载大小
	successedPieceNum uint64            // 已下载 Piece 数量
}

// 生成一个新的 Context
func newContext() *Context {
	return &Context{
		doneChan:   make(chan struct{}),
		resultChan: make(chan *PieceResult),
		errChan:    make(chan error, 50),
	}
}

// 未实现
func (ctx *Context) Deadline() (deadline time.Time, ok bool) {
	panic(ErrNoImplement)
}

// 获取任务完成的通知
func (ctx *Context) Done() <-chan struct{} {
	return ctx.doneChan
}
func (ctx *Context) Err() error {
	select {
	case err := <-ctx.errChan:
		return err
	default:
		return nil
	}
}

// 未实现
func (ctx *Context) Value(key any) any {
	panic(ErrNoImplement)
}

// 获取正在下载的 Peer 列表
func (ctx *Context) GetPeerInfos() []PeerInfo {
	ctx.rwm.RLock()
	defer ctx.rwm.RUnlock()
	return ctx.peerInfos
}

// 获取正在下载进度（已下载大小和已下载片数）
func (ctx *Context) GetProcess() (uint64, uint64) {
	successedBtye := atomic.LoadUint64(&ctx.successedBtye)
	successedPieceNum := atomic.LoadUint64(&ctx.successedPieceNum)
	return successedBtye, successedPieceNum
}

func (ctx *Context) GetResult() <-chan *PieceResult {
	return ctx.resultChan
}

// 错误信息通道
func (ctx *Context) GetErr() <-chan error {
	return ctx.errChan
}

// 结束任务
func (ctx *Context) Finish() {
	close(ctx.doneChan)
	close(ctx.resultChan)
	close(ctx.errChan)
}

var _ context.Context = (*Context)(nil)
