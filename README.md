# Akimio521/torrent-go
**一个使用 go-lang 编写的种子解析/下载库**

# 说明
## bencode 包
- **BObject**: bencode 存储基元，用于存储基本 bencode 对象或复杂 bencode 对象
  - **func GetBObject[T allowedTypes](v T) *BObject**: 用于生成 BObject 对象，保证 GetBObject 中值的合法性
  - **func (o *BObject) GetBType() BType**: 可以得到 BObject 对象存储的类型
  - **func GetValue[T allowedTypes](o *BObject, dest *T) error**: 可以得到BObject 对象存储的值
  - **func (o *BObject) Bencode(w io.Writer) (int, error)**: 将 BObject 编码写入 Writer 中，返回写入的字节长度
  - **func Parse(r io.Reader) (*BObject, error)**: 从 io.Reader 中读取流并解析成 BObject 对象
- Marshal: 将数据结构转换成 bencode 输出到 io.Wirter 中或者从 io.Reader 读取 bencode 流转换成数据结构，使用标签 `bencode:"<name>"` 进行解析
  - **func Unmarshal(r io.Reader, s any) error**: 从 io.Reader 读 bencode 数据绑定在数据结构上
  - **func Marshal(w io.Writer, s any) (int, error)**: 将数据结构转换成 bencode 输出到 io.Wirter 中
- other: 工具包，底层函数
  - **func EncodeString(w io.Writer, val string) (int, error)**: 编码一个字符串写入 io.Writer
  - **func DecodeString(r io.Reader) (string, error)**: 从 io.Reader 读数据并解码成字符串
  - **func EncodeInt(w io.Writer, val int) (int, error)**: 编码一个整数写入 io.Writer
  - **func DecodeInt(r io.Reader) (int, error)**: 从 io.Reader 读数据并解码成整数
