package torrent

import "strconv"

type Bitfield []byte // 其中下载器已发送的每个索引都设置为 1，其余设置为 0。尚未有任何内容的下载器可以跳过“位字段”消息。位字段的第一个字节对应于从高到低位依次为 0-7 的索引。下一个字节对应于 8-15，等等。末尾的空位被设置为 0。

// 检查是否拥有指定索引的片段
func (field Bitfield) HasPiece(index int) bool {
	byteIndex := index / 8
	offset := index % 8
	if byteIndex < 0 || byteIndex >= len(field) {
		return false
	}
	return field[byteIndex]>>uint(7-offset)&1 != 0
}

// 设置指定索引的片段
func (field Bitfield) SetPiece(index int) {
	byteIndex := index / 8
	offset := index % 8
	if byteIndex < 0 || byteIndex >= len(field) {
		return
	}
	field[byteIndex] |= 1 << uint(7-offset)
}

// 将拥有的片段转换为字符串
func (field Bitfield) String() string {
	str := "piece# "
	for i := 0; i < len(field)*8; i++ {
		if field.HasPiece(i) {
			str = str + strconv.Itoa(i) + " "
		}
	}
	return str
}
