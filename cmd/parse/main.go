package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/Akimio521/torrent-go/torrent"
)

func main() {
	filePath := flag.String("file", "", "Path to the torrent file")
	flag.Parse()

	if *filePath == "" {
		fmt.Println("Error: Torrent file path is required.")
		flag.Usage()
		os.Exit(1)
	}

	// 2. 打开文件
	file, err := os.Open(*filePath)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	// 3. 解析文件
	torrent, err := torrent.ParseFile(file)
	if err != nil {
		fmt.Printf("Error parsing torrent file: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(torrent.Info.Name)
	fmt.Println(torrent.Announce)
	fmt.Println(torrent.AnnounceList)
	fmt.Println(torrent.Info.PiceLength)
	if b, err := json.MarshalIndent(torrent.Info.Files, "", "  "); b != nil && err == nil {
		fmt.Println(string(b))
	}
	if torrent.Info.Files != nil {
		num := 0
		for _, f := range torrent.Info.Files {
			num += f.Length
		}
		fmt.Println("total size: ", num)
	}
	fmt.Println(torrent.GetInfoSHA1())
}
