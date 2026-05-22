package main

import (
	"bufio"
	"fmt"
	"log"
	"math/rand/v2"
	"os"
	"time"
)

func main() {

	err := os.MkdirAll("logs/", 0750)
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.OpenFile("logs/auth.log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}
	writer := bufio.NewWriter(f)
	defer f.Close()
	for {
		tm := time.Now()
		msg := fmt.Sprintf("[%d] [LOG] - %s \n", rand.IntN(3), tm.Format("2006-01-02 15:04:05"))
		_, err := writer.WriteString(msg)
		if err != nil {
			fmt.Printf("Error writing string\n", err)
		}
		writer.Flush()

		time.Sleep(2 * time.Second)
	}
}
