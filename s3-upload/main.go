package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/cheggaaa/pb/v3"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
)

type SignedURL struct {
	url    string
	fields map[string]string
}

// Compile-time variable hidden from end-user
// use go build -ldflags "-X main.endpoint=$ENDPOINT"
var endpoint string

type Namespace struct {
	apikey    string
	fileNames []string
}

func parseArgs(argc []string) (Namespace, error) {
	flagSet := flag.NewFlagSet(argc[0], flag.ContinueOnError)
	apikey := flagSet.String("apikey", "", "AWS Lambda API key")
	err := flagSet.Parse(argc[1:])
	if err != nil {
		return Namespace{}, err
	}
	return Namespace{*apikey, flagSet.Args()}, nil
}

func Usage() {
	fmt.Printf("Usage: %s -apikey APIKEY FILE [FILE ...]]\n", os.Args[0])
}

func main() {
	args, err := parseArgs(os.Args)
	if err != nil {
		log.Fatal(err)
	}
	if args.apikey == "" {
		Usage()
		fmt.Println("Please specify -apikey")
		return
	}
	if len(args.fileNames) == 0 {
		Usage()
		fmt.Println("Too few arguments specified")
		return
	}
	for _, fileName := range args.fileNames {
		fmt.Printf("pre-sign URL: %s\n", fileName)
		signed, err := presignedURL(endpoint, fileName)
		if err != nil {
			fmt.Printf("pre-signed URL generation failed: %s\n", fileName)
		}
		err = fileUpload(fileName, signed.url, signed.fields)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func fileUpload(fileName string, url string, params map[string]string) error {
	fmt.Printf("upload: %s to %s\n", fileName, url)

	// Read/write buffer to store file content
	rw := &bytes.Buffer{}

	// Multipart Writer
	writer := multipart.NewWriter(rw)

	// Add pre-signed form-fields
	for k, v := range params {
		writer.WriteField(k, v)
	}

	// Add file form-field at the end (AWS peculiarity)
	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return err
	}

	// Open file
	reader, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer reader.Close()

	// File size in bytes
	info, err := reader.Stat()
	if err != nil {
		return err
	}
	fileSize64 := info.Size()

	// Progress bar
	bar := pb.Full.Start(int(fileSize64))
	barReader := bar.NewProxyReader(reader)

	// Copy file using multipart.Writer
	if _, err = io.Copy(part, barReader); err != nil {
		return err
	}
	writer.Close()
	bar.Finish()

	// POST request and Println response (consider refactor)
	response, err := http.Post(url, writer.FormDataContentType(), rw)
	if err != nil {
		return err
	} else {
		body := &bytes.Buffer{}
		_, err := body.ReadFrom(response.Body)
		if err != nil {
			return err
		}
		response.Body.Close()
		if true {
			fmt.Println(response.StatusCode)
			fmt.Println(response.Header)
			fmt.Println(body)
		}
	}
	return nil
}

func presignedURL(endpoint, fileName string) (SignedURL, error) {
	// Ask Lambda for pre-signed URL and parse response
	url := endpoint + "?file=" + fileName
	res, err := http.Get(url)
	if err != nil {
		return SignedURL{}, err
	}
	defer res.Body.Close()
	content, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return SignedURL{}, err
	}
	var data interface{}
	err = json.Unmarshal(content, &data)
	if err != nil {
		return SignedURL{}, err
	}
	mapping := data.(map[string]interface{})
	body := mapping["body"].(string)

	var bodyData interface{}
	err = json.Unmarshal([]byte(body), &bodyData)
	if err != nil {
		return SignedURL{}, err
	}
	b := bodyData.(map[string]interface{})
	s3url := b["url"].(string)
	fields := b["fields"].(map[string]interface{})
	s3fields := make(map[string]string)
	for k, v := range fields {
		s3fields[k] = v.(string)
	}
	return SignedURL{s3url, s3fields}, nil
}
