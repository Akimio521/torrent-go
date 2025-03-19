package torrent

import (
	"crypto/sha1"
	"errors"
	"net"
)

const (
	RESERVED_LEN      int    = 8                                      // 保留长度
	PEER_ID_LEN       int    = 20                                     // Peer ID 长度
	HS_MSG_LEN        int    = RESERVED_LEN + sha1.Size + PEER_ID_LEN // 握手消息长度
	PORT_LEN          int    = 2                                      // 端口长度
	PEER_V4_LEN       int    = net.IPv4len + PORT_LEN                 // Peer 长度（IPv4）
	PEER_V6_LEN       int    = net.IPv6len + PORT_LEN                 // Peer 长度（IPv6）
	PEER_MSG_HEAD_LEN uint32 = 4                                      // Peer 消息头长度（消息头用于存储消息长度（不包括消息头））
	BLOCK_SIZE               = 16 * 1024                              // 块大小（16KB）
	MAX_BACKLOG              = 5                                      // 最大并发度（同一个 Peer）
)

type MsgId uint8

const (
	MsgChoke       MsgId = iota // 上传者阻塞（不提供上传，没有 Payload）
	MsgUnchoke                  // 上传者未阻塞（支持上传，没有 Payload）
	MsgInterested               // 下载者感兴趣（想要下载，没有 Payload）
	MsgNotInterest              // 下载者不感兴趣（不想下载，没有 Payload）
	MsgHave                     // 拥有（有效 Payload 是一个单独的数字，即那个下载器刚刚完成并检查了哈希值的索引）
	MsgBitfield                 // 位图（Payload 是 Bitfield）
	MsgRequest                  // 下载请求（请求消息包含索引、开始和长度。后两者是字节偏移量。长度通常是 2 的幂，除非它被文件末尾截断）
	MsgPiece                    // 下载响应（Payload 是块内容）
	MsgCancel                   // 取消消息（取消消息与请求消息具有相同的负载。它们通常只在下载的“终局模式”结束时发送。当下载接近完成时，最后几块内容往往会从单个故障调制解调器线路下载，耗时非常长。为了确保最后几块内容能快速到达，一旦给定下载器尚未拥有的所有块请求都处于挂起状态，它就会向所有正在下载的内容发送请求。为了防止这变得极其低效，每当一块内容到达时，它就会向其他人发送取消请求）
)

var (
	ErrMalformedPeersFormat = errors.New("malformed peers format") // 错误 Peers 格式
	ErrZeroPrelen           = errors.New("prelen cannot be 0")     // 握手消息中 prelen 不能为0
	ErrCheckInfoSHAFaild    = errors.New("check handshake failed") // 检查 InfoSHA 失败
	ErrNoImplement          = errors.New("no implement")           // 未实现
)
