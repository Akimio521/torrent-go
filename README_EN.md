<div align="center">

# Akimio521/torrent-go
**A BitTorrent Client Library Written in Go**  

[简体中文](./README.md) •
English

</div>

## Overview  
This repository contains a Go implementation of a BitTorrent client library, supporting torrent file parsing, peer communication, and piece-based downloading. The project is organized into three main packages: `bencode`, `torrent`, and `cmd`.

---

## Features  
- **Bencode Support**: Full encoding/decoding of Bencode data structures.  
- **Torrent Parsing**: Extract metadata and piece hashes from `.torrent` files.  
- **Peer Discovery**: Query trackers for peer lists.  
- **Parallel Downloading**: Multi-piece downloading with SHA-1 validation.  
- **Progress Tracking**: Real-time progress monitoring via `Context`.  

---

## Usage Example  
```go  
file, _ := os.Open("example.torrent")  
tf, _ := torrent.ParseFile(file)
var peerId [torrent.PEER_ID_LEN]byte // radom Generate Peer ID
_, _ = rand.Read(peerId[:])
task, _ := tf.GetTask(peerID, 6881)
file.close()

file, err := os.Create(task.FileName)
if err != nil {
  fmt.Println("fail to create file: " + task.FileName)
  os.Exit(1)
}
defer file.Close()

if err = file.Truncate(int64(task.FileLen)); err != nil {
    fmt.Printf("fail to allocate disk space: %v\n", err)
    os.Exit(1)
}

ctx := task.Download()
for res := range ctx.GetResult() {
    begin, _ := task.GetPieceBounds(res.Index)
    if _, err := file.WriteAt(res.Data, int64(begin)); err != nil {
        fmt.Printf("fail to write piece %d: %v\n", res.Index, err)
        os.Exit(1)
    }
}
```

---

## Packages  

### `bencode` Package  
Implements Bencode encoding/decoding (BitTorrent's data serialization format).  

#### Core Components:  
- **`BObject`**: Base type for storing Bencode data (supports strings, integers, lists, and dictionaries).  
  - `func GetBObject[T allowedTypes](v T) *BObject`: Creates a valid `BObject` from primitive/complex types.  
  - `func (o *BObject) GetBType() BType`: Returns the stored data type.  
  - `func GetValue[T allowedTypes](o *BObject, dest *T) error`: Extracts the stored value into `dest`.  
  - `func (o *BObject) Bencode(w io.Writer) (int, error)`: Encodes the object to Bencode format.  
  - `func Parse(r io.Reader) (*BObject, error)`: Parses Bencode data from a reader.  

- **Marshal/Unmarshal**:  
  - `func Unmarshal(r io.Reader, s any) error`: Binds Bencode data to a Go struct (uses `bencode`:"<name>"` tags).  
  - `func Marshal(w io.Writer, s any) (int, error)`: Serializes a Go struct to Bencode format.  

- **Utilities**:  
  - `EncodeString`, `DecodeString`, `EncodeInt`, `DecodeInt`: Low-level encoding/decoding primitives.  

---

### `torrent` Package  
Handles torrent file parsing, peer discovery, and download management.  

#### Key Structs:  
- **`TorrentFile`**: Represents torrent metadata.  
  - `func ParseFile(r io.Reader) (*TorrentFile, error)`: Parses a `.torrent` file.  
  - `func (tf *TorrentFile) GetInfoSHA1() [sha1.Size]byte`: Returns SHA-1 hash of the `info` dictionary.  
  - `func (tf *TorrentFile) GetAllPieceSHA() [][sha1.Size]byte`: Returns SHA-1 hashes for all pieces.  
  - `func (tf *TorrentFile) GetTask(peerID [PEER_ID_LEN]byte, port uint16) (*TorrentTask, error)`: Creates a download task using tracker-provided peers.  
  - `func (tf *TorrentFile) FindPeers(...) ([]PeerInfo, error)`: Queries trackers for active peers.  

- **`PeerInfo`**: Stores peer connection details (IP/Port).  
  - `func (pi *PeerInfo) GetConnAddr() string`: Returns address as `IP`:Port`.  

- **`TorrentTask`**: Manages download tasks.  
  - `func (t *TorrentTask) GetPieceBounds(index int) (begin, end int)`: Calculates byte range for a piece.  
  - `func (task *TorrentTask) Download() *Context`: Starts the download process.  

- **`Context`**: Tracks download progress and results.  
  - `Done() <-chan struct{}`: Channel signaling task completion.  
  - `GetProcess() (uint64, uint64)`: Returns downloaded bytes/pieces count.  
  - `GetResult() <-chan *PieceResult`: Stream of downloaded pieces.  

---

### `cmd` Package  
Contains example implementations and CLI tools demonstrating library usage.  