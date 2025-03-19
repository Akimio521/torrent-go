# Akimio521/torrent-go
**一个使用 go-lang 编写的种子解析/下载库**

# 说明
## bencode 包
- **BObject**: bencode 存储基元，用于存储基本 bencode 对象或复杂 bencode 对象
  - **func GetBObject[T allowedTypes](v T) \*BObject**: 用于生成 BObject 对象，保证 GetBObject 中值的合法性
  - **func (o *BObject) GetBType() BType**: 可以得到 BObject 对象存储的类型
  - **func GetValue[T allowedTypes](o \*BObject, dest *T) error**: 可以得到BObject 对象存储的值
  - **func (o \*BObject) Bencode(w io.Writer) (int, error)**: 将 BObject 编码写入 Writer 中，返回写入的字节长度
  - **func Parse(r io.Reader) (\*BObject, error)**: 从 io.Reader 中读取流并解析成 BObject 对象
- Marshal: 将数据结构转换成 bencode 输出到 io.Wirter 中或者从 io.Reader 读取 bencode 流转换成数据结构，使用标签 `bencode:"<name>"` 进行解析
  - **func Unmarshal(r io.Reader, s any) error**: 从 io.Reader 读 bencode 数据绑定在数据结构上
  - **func Marshal(w io.Writer, s any) (int, error)**: 将数据结构转换成 bencode 输出到 io.Wirter 中
- other: 工具包，底层函数
  - **func EncodeString(w io.Writer, val string) (int, error)**: 编码一个字符串写入 io.Writer
  - **func DecodeString(r io.Reader) (string, error)**: 从 io.Reader 读数据并解码成字符串
  - **func EncodeInt(w io.Writer, val int) (int, error)**: 编码一个整数写入 io.Writer
  - **func DecodeInt(r io.Reader) (int, error)**: 从 io.Reader 读数据并解码成整数
## torrent 包
- **TorrentFile**: 种子文件基本数据结构
  - **func ParseFile(r io.Reader) (\*TorrentFile, error)**: 传入一个 io.Reader 读取解析其中的种子文件信息
  - **func (tf \*TorrentFile) GetInfoSHA1() [sha1.Size]byte**: 获得种子的 info 信息的 SHA-1 哈希值
  - **func (tf \*TorrentFile) GetAllPieceSHA() [][sha1.Size]byte**: 获取所有的 SHA1 (使用 sha1.Size 分片存储方便分片下载后进行校验)
  - **func (tf \*TorrentFile) GetPiecesNum() int**: 计算种子的分片数
  - **func (tf \*TorrentFile) GetTask(peerID [PEER_ID_LEN]byte, port uint16) (\*TorrentTask, error)**: 传入本机的 PeerID 和 监听端口，自动向 tracker 查询 peer，返回一个 TorrentTask 对象
  - **func (tf \*TorrentFile) FindPeers(peerID [PEER_ID_LEN]byte, port uint16) ([]PeerInfo, error)**: 向 tracker 发起请求，查询 Peer
- **PeerInfo**: 对端信息结构，包含对端 IP、监听端口
  - **func (pi \*PeerInfo) GetConnAddr() string**: 获取对端的连接地址
- **TorrentTask**: 种子下载任务
  - **func (t \*TorrentTask) GetPieceBounds(index int) (bengin int, end int)**: 根据分片的索引值计算分片的起始位置
  - **func (task \*TorrentTask) Download() \*Context**: 开始下载任务，返回一个 *Context
- **PieceTask**: 分片下载任务，包含索引值，SHA-1 哈希校验值，分片长度信息
  - **func (pt \*PieceTask) CheckPiece(res \*PieceResult) bool**: 检查校验下载的分片内容是否匹配
- **PieceResult**: 分片下载结构，包含索引值，下载数据
- **Context**: 下载任务上下文
  - **func (ctx \*Context) Done() <-chan struct{}**: 等待任务结束
  - **func (ctx \*Context) GetPeerInfos() []PeerInfo**: 返回连接成功的对端 Peer 信息
  - **func (ctx \*Context) GetProcess() (uint64, uint64)**: 获取正在下载进度（已下载大小和已下载片数）
  - **func (ctx \*Context) GetResult() <-chan *PieceResult**: 分片下载结果通道
  - **func (ctx \*Context) GetErr() <-chan error**: 错误信息通道
## cmd 包
一些简单的实例代码