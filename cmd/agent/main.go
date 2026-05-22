package main

import (
	//"encoding/binary"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

/*	Some convertion func.
	func int64TObyte(n int64) []byte {
		b := make([]byte, 8)
		binary.BigEndian.PutUint64(b, uint64(n))
		return b
	}

	func byteToint64(b []byte) int64 {
		return int64(binary.BigEndian.Uint64(b))
	}
*/

type Config struct {
	ServerAddr      string `json:"server_address"`
	LogDir          string `json:"log_dir"`
	PollIntervalMs  int    `json:"poll_interval_ms"`
	RetryIntervalMs int    `json:"retry_interval_ms"`
}

func createOffsetFile(f *os.File, offsetPath string) int64 {

	if _, err := os.Stat(offsetPath + ".offset"); err == nil {
		offB, err := os.ReadFile(offsetPath + ".offset")
		if err != nil {
			log.Fatal(err)
		}

		res, err := strconv.ParseInt(string(offB), 10, 64)
		if err != nil {
			log.Fatal(err)
		}

		return res
	} else {

		ret, err := f.Seek(0, 1) // create file, it doesn't exist
		if err != nil {
			log.Fatal(err)
		}
		//b := int64TObyte(ret)
		b := strconv.FormatInt(ret, 10)
		os.WriteFile(offsetPath+".offset", []byte(b), 0600)

		return ret
	}

}

func analyzeAndPrint(logmsg string, path string) {
	if strings.HasPrefix(logmsg, "[0]") {
		fmt.Printf("[LOG] [%s] %s", path, logmsg) //
	} else if strings.HasPrefix(logmsg, "[1]") {
		fmt.Printf("\033[33m[WARNING] [%s] %s\033[0m", path, logmsg) // prints yellow log msg
	} else if strings.HasPrefix(logmsg, "[2]") {
		fmt.Printf("\033[31m[CRITICAL] [%s] %s\033[0m", path, logmsg) // prints red log msg
	} else {
		fmt.Printf("[INVALID LOG SITUATION] [%s] %s", path, logmsg)
	}

}

func tailFile(ctx context.Context, path string, conn net.Conn, PollIntervalMs int) {

	f, err := os.OpenFile(path, os.O_RDONLY, 0600)
	if err != nil {
		panic(err)
	}
	//f.Seek(0, 2) // SEEK_END = 2

	lastPos := createOffsetFile(f, path)
	f.Seek(lastPos, 0)

	for {

		FileInfo, statERR := os.Stat(path)

		if statERR == nil && FileInfo.Size() < lastPos {
			f.Close()
			f, err = os.OpenFile(path, os.O_RDONLY, 0600)
			if err != nil {
				log.Printf("New file couldn't opened: %v", err)
				continue
			}
			lastPos = 0

			os.WriteFile(path+".offset", []byte("0"), 0600)
		}

		logmsg := make([]byte, 1024)
		n, err := f.Read(logmsg)
		if err == io.EOF {
			select {
			case <-ctx.Done():
				f.Close()
				return
			case <-time.After(time.Duration(PollIntervalMs) * time.Millisecond):
			}
		} else {
			analyzeAndPrint(string(logmsg[:n]), filepath.Base(path))
			fmt.Fprintf(conn, "[%s] %s", filepath.Base(path), string(logmsg[:n]))
			ret, err := f.Seek(0, 1)
			if err != nil {
				log.Fatal(err)
			}
			os.WriteFile(path+".offset", []byte(strconv.FormatInt(ret, 10)), 0600)
		}
	}
}

func loadConfig(path string) (*Config, error) {

	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	err = json.Unmarshal(file, &config)
	if err != nil {
		return nil, err
	}
	fmt.Printf("SERVER: %s\n", config.ServerAddr)
	return &config, nil
}

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt)
	cfg, err := loadConfig("config.json")
	if err != nil {
		log.Fatal(err)
	}

	files, err := os.ReadDir(cfg.LogDir)
	if err != nil {
		log.Fatal(err)
	}
	var conn net.Conn
	// err's already declared above
	for {
		conn, err = net.Dial("tcp", cfg.ServerAddr)
		if err != nil {
			log.Printf("Couldn't connected to the server %v. Trying again...", err)
			time.Sleep(time.Duration(cfg.RetryIntervalMs) * time.Millisecond)
			continue
		} else {
			log.Println("Connected Successfully")
			break
		}
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".log" {
			fullPath := filepath.Join(cfg.LogDir, file.Name())
			go tailFile(ctx, fullPath, conn, cfg.PollIntervalMs)
		}
	}

	<-signalChannel
	fmt.Printf("Agent shuting down...\n")
	cancel()
	conn.Close()
}
