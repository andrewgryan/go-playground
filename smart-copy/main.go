package main

import (
	"flag"
	"fmt"
	"github.com/cheggaaa/pb/v3"
	"io"
	"log"
	"os"
)

// Compile-time variable hidden from end-user
// use go build -ldflags "-X main.endpoint=$ENDPOINT"
var endpoint string

func main() {
	// Example supporting -user flag
	user := flag.String("user", "default", "user name")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options] [FILE [FILE ...]]\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		flag.Usage()
		fmt.Printf("please specify at least one file\n")
		return
	}
	fmt.Printf("user: %s\n", *user)
	for _, fileName := range args {
		fmt.Println(fileName)

		// Read fileName
		reader, err := os.Open(fileName)
		if err != nil {
			log.Fatal(err)
		}
		defer reader.Close()

		// File size in bytes
		info, err := reader.Stat()
		if err != nil {
			log.Fatal(err)
		}
		fileSize64 := info.Size()

		// Write fileName.copy
		writer, err := os.Create(fileName + ".copy")
		if err != nil {
			log.Fatal(err)
		}
		defer writer.Close()

		// Play with progress bar
		bar := pb.Full.Start(int(fileSize64))
		barReader := bar.NewProxyReader(reader)

		// Copy file from fileName to fileName.copy
		io.Copy(writer, barReader)

		bar.Finish()
	}
}
