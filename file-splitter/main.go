package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s [FILE [FILE ...]]\n", os.Args[0])
		fmt.Println("Too few arguments")
		return
	}
	chunkSize := 100 * 1024 * 1024
	for _, fileName := range os.Args[1:] {
		// File size
		file, err := os.Open(fileName)
		if err != nil {
			log.Fatal(err)
		}
		info, _ := file.Stat()
		fileSize := int(info.Size())
		file.Close()

		// Read 100mb chunk into fileName.part
		fmt.Printf("splitting: %s\n", fileName)
		split(fileName, chunkSize)
		fmt.Printf("joining: %s\n", fileName)
		err = join(fileName, fileSize, chunkSize)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func join(fileName string, fileSize, chunkSize int) error {
	joinFile, err := os.Create(fileName + ".join")
	if err != nil {
		return err
	}
	defer joinFile.Close()

	// Fill join file with parts
	buffer := make([]byte, chunkSize)
	for i := 0; i < chunks(fileSize, chunkSize); i++ {
		filePart := fileName + ".part-" + strconv.Itoa(i)
		fmt.Printf("reading: %s\n", filePart)
		part, err := os.Open(filePart)
		if err != nil {
			return err
		}
		n, err := part.Read(buffer)
		if err != nil {
			return err
		}
		fmt.Printf("%d bytes read\n", n)
		joinFile.Write(buffer[:n])
		part.Close()
	}

	return nil
}

func split(fileName string, chunkSize int) error {
	buffer := make([]byte, chunkSize)

	// Open file to read
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// File size in bytes
	info, err := file.Stat()
	if err != nil {
		return err
	}
	for i := 0; i < chunks(int(info.Size()), chunkSize); i++ {
		// Read next chunk
		n, err := file.Read(buffer)
		if err != nil {
			return err
		}
		fmt.Printf("%d bytes read\n", n)

		// Write chunk file
		outFile, err := os.Create(fileName + ".part-" + strconv.Itoa(i))
		if err != nil {
			return err
		}
		outFile.Write(buffer[:n])
		outFile.Close()
	}
	return nil
}

// Number of chunks given file size
func chunks(fileSize, chunkSize int) int {
	return int(math.Ceil(float64(fileSize) / float64(chunkSize)))
}
