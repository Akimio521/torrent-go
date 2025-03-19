package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Akimio521/torrent-go/torrent"
)

const (
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorReset  = "\033[0m"
)

// 根据速度自动选择单位（KB/MB/GB）
func formatSpeed(speedKB float64) string {
	switch {
	case speedKB >= 1024*1024:
		return fmt.Sprintf("%s%.2f GB/s%s", colorGreen, speedKB/(1024*1024), colorReset)
	case speedKB >= 1024:
		return fmt.Sprintf("%s%.2f MB/s%s", colorYellow, speedKB/1024, colorReset)
	default:
		return fmt.Sprintf("%s%.2f KB/s%s", colorRed, speedKB, colorReset)
	}
}
func main() {
	filePath := flag.String("file", "", "Path to the torrent file")
	port := flag.Uint("port", 6881, "Port to listen on")
	flag.Parse()
	if *filePath == "" {
		fmt.Println("Error: Torrent file path is required.")
		flag.Usage()
		os.Exit(1)
	}
	var tf *torrent.TorrentFile
	{
		file, err := os.Open(*filePath)
		if err != nil {
			fmt.Println("open file error:", err.Error())
			os.Exit(1)
		}
		defer file.Close()
		tf, err = torrent.ParseFile(file)
		if err != nil {
			fmt.Println("parse file error:", err.Error())
			os.Exit(1)
		}

	}

	var peerId [torrent.PEER_ID_LEN]byte // 随机生成 Peer ID
	_, _ = rand.Read(peerId[:])

	task, err := tf.GetTask(peerId, uint16(*port))
	if err != nil {
		fmt.Println("get task error:", err.Error())
		os.Exit(1)
	}

	ctx := task.Download()
	go func() {
		startTime := time.Now()
		var succeededByte, successedPieceNum uint64
		totalPieces := len(task.PieceSHA)
		errBuffer := make([]string, 0, 3)
		for {
			select {
			case <-ctx.Done():
				fmt.Println("\033[32mDownload complete[0m")
				return
			case err := <-ctx.GetErr():
				msg := fmt.Sprintf("\033[31m[ERROR]\033[0m: %s", err.Error())
				if len(errBuffer) == 3 {
					copy(errBuffer[0:], errBuffer[1:]) // 移除最早的错误信息
					errBuffer[2] = msg
				} else {
					errBuffer = append(errBuffer, msg)
				}
			default:
				succeededByte, successedPieceNum = ctx.GetProcess()
				processPercentage := int(float64(successedPieceNum) / float64(totalPieces) * 100)
				fmt.Print("\033[H\033[2J") // 清屏
				speed := float64(succeededByte) / 1024 / time.Since(startTime).Seconds()
				fmt.Printf("\r[%-101s] %3d%% %8d/%d %s\n", strings.Repeat(">", processPercentage)+"🚀", processPercentage, successedPieceNum, totalPieces, formatSpeed(speed))
				// 打印 Error 信息
				for _, err := range errBuffer {
					fmt.Println(err)
				}
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()
	file, err := os.Create(task.FileName)
	if err != nil {
		fmt.Println("fail to create file: " + task.FileName)
		os.Exit(1)
	}
	defer file.Close()

	// 设置文件大小（预分配空间）
	if err = file.Truncate(int64(task.FileLen)); err != nil {
		fmt.Printf("fail to allocate disk space: %v\n", err)
		os.Exit(1)
	}
	for res := range ctx.GetResult() {
		begin, _ := task.GetPieceBounds(res.Index)
		// 直接将下载片段写入硬盘对应位置
		if _, err := file.WriteAt(res.Data, int64(begin)); err != nil {
			fmt.Printf("fail to write piece %d: %v\n", res.Index, err)
			os.Exit(1)
		}
	}
}
