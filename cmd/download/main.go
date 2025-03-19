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
	speedTracker := NewSpeedTracker()
	totalBytes := task.FileLen

	ctx := task.Download()
	go func() {
		startTime := time.Now()
		var currentBytes, currentPieces uint64

		errBuffer := make([]string, 0, 3)
		for {
			select {
			case <-ctx.Done():
				fmt.Printf(
					"%sDownload complete (%.2f MB/s)%s\n",
					colorGreen,
					float64(totalBytes)/1024/1024/time.Since(startTime).Seconds(),
					colorReset,
				)
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
				currentBytes, currentPieces = ctx.GetProcess()
				// 计算基于字节的进度
				bytePercentage := int(float64(currentBytes) / float64(totalBytes) * 100)
				speedKB := speedTracker.Update(currentBytes)

				// 减少清屏频率（每500ms刷新一次）
				if time.Since(startTime).Milliseconds()%500 < 100 {
					fmt.Print("\033[H\033[2J") // 清屏
					fmt.Printf("\r[%-100s] %3d%% %8d/%d %s\n",
						generateProgressBar(bytePercentage),
						currentPieces,
						totalBytes,
						len(task.PieceSHA),
						formatSpeed(speedKB),
					)
					fmt.Println(currentBytes)
					for _, err := range errBuffer {
						fmt.Println(err)
					}
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

func generateProgressBar(p int) string {
	if p <= 0 {
		return "🚀"
	} else if p > 100 {
		p = 100
	}
	return strings.Repeat(">", p-1) + "🚀"
}

type SpeedTracker struct { // 跟踪速度结构体
	lastBytes    uint64
	lastTime     time.Time
	window       [5]float64 // 5秒滑动窗口
	windowCursor int
}

func NewSpeedTracker() *SpeedTracker {
	return &SpeedTracker{
		lastTime: time.Now(),
	}
}

// 更新速度并返回当前速度（KB/s）
func (st *SpeedTracker) Update(currentBytes uint64) float64 {
	now := time.Now()
	elapsed := now.Sub(st.lastTime).Seconds()
	if elapsed < 0.1 { // 最小时间间隔
		return 0
	}

	delta := currentBytes - st.lastBytes
	speedKB := float64(delta) / 1024 / elapsed

	// 更新滑动窗口
	st.window[st.windowCursor] = speedKB
	st.windowCursor = (st.windowCursor + 1) % len(st.window)

	st.lastBytes = currentBytes
	st.lastTime = now

	// 计算窗口平均值
	sum := 0.0
	validCount := 0
	for _, v := range st.window {
		if v > 0 {
			sum += v
			validCount++
		}
	}
	if validCount == 0 {
		return 0
	}
	return sum / float64(validCount)
}

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
