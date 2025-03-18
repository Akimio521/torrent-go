package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"os"

	"github.com/Akimio521/torrent-go/torrent"
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

	// random peerId
	var peerId [torrent.PEER_ID_LEN]byte
	_, _ = rand.Read(peerId[:])
	fmt.Println("peerId:", peerId)
	//connect tracker & find peers
	task, err := tf.GetTask(peerId, uint16(*port))
	if err != nil {
		fmt.Println("get task error:", err.Error())
		os.Exit(1)
	}
	//download from peers & make file
	if err = task.Download(); err != nil {
		fmt.Println("download error:", err.Error())
		os.Exit(1)
	}
}
