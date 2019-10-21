package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

type SignedURL struct {
	url    string
	fields map[string]string
}

func main() {
	if len(os.Args) < 3 {
		fmt.Printf("Usage: %s ENDPOINT [FILE [FILE ...]]\n", os.Args[0])
		fmt.Println("Too few arguments specified")
		return
	}
	endpoint := os.Args[1]
	for _, fileName := range os.Args[2:] {
		signed, err := presignedURL(endpoint, fileName)
		if err != nil {
			fmt.Printf("pre-signed URL generation failed: %s\n", fileName)
		}
		fmt.Println(signed)
	}
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
