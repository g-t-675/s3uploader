package main

// author and maintainer Goce Trenchev <goce.trenchev@gmail.com>
// (not looking for any work currently)
// https://docs.aws.amazon.com/sdk-for-go/api/aws/

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type reader struct {
	r io.Reader
}

// holder for the configuration
type S3Config struct {
	S3Region string
	S3Key    string
	S3Secret string
	S3Bucket string
	S3Folder string
}

var config *S3Config

// handling errors the easy way
func handleErr(err error) {
	if err != nil {
		log.Fatal(fmt.Sprintf("\033[1;31m%s\033[0m", err.Error()))
	}
}

// getting environment variables the easy way
func getEnv(key string) string {
	value, err := os.LookupEnv(key)
	if err == false {
		if strings.Contains(key, "FOLDER") {
			return "/" // if it's not defined then it uploads to
			// to the root of the s3 bucket, i.e. "/"
		} else if key == "REGION" {
			return "us-east-1"
			// same as above, you may change if needed
		}
		return ""
	}
	return value
}

func init() {
	// initializes the config from the environment variables
	config = &S3Config{
		S3Region: getEnv("REGION"),
		S3Bucket: getEnv("BUCKET"),
		S3Key:    getEnv("ACCESSKEY"),
		S3Secret: getEnv("SECRET"),
		S3Folder: getEnv("FOLDER"),
	}

}

func uploadFile(fileName string) (bool, error) {
	fileBody, err := ioutil.ReadFile(fileName)
	// ^^ this will not work if the file is larger than the free memory on the
	// system however it is much more efficient to read a file in a single gulp
	// rather than to split it, counting also line at a time
	// it makes quite a lot of difference in performance and memory usage, as
	// shown in the two benchmark comparisons:
	// (the runtime memstats was used)
	// https://golang.org/pkg/runtime/#MemStats
	//
	// Reading a line at a time:
	// Alloc: 651 MB, TotalAlloc: 1192 MB, Sys: 812 MB
	// Mallocs: 1166446, Frees: 518919
	// HeapAlloc: 651 MB, HeapSys: 767 MB, HeapIdle: 74 MB
	// HeapObjects: 647527
	//
	// Reading all at once:
	//
	// Alloc: 583 MB, TotalAlloc: 583 MB, Sys: 662 MB
	// Mallocs: 192, Frees: 13
	// HeapAlloc: 583 MB, HeapSys: 639 MB, HeapIdle: 55 MB
	// HeapObjects: 179
	handleErr(err)

	// create a session which provides the client with
	// shared configuration such as region, creds etc.
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(config.S3Region),
		Credentials: credentials.NewStaticCredentials(
			config.S3Key,
			config.S3Secret,
			""),
	})
	handleErr(err)

	// create new instance with a definition on how parts are kept in memory
	uploader := s3manager.NewUploader(sess, func(u *s3manager.Uploader) {
		// buffer 256 MiB in memory
		u.BufferProvider = s3manager.NewBufferedReadSeekerWriteToPool(256 * 1024 * 1024)
		// make it ~83 MB
		u.PartSize = 20 << 22
	})

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(config.S3Bucket),
		Key:    aws.String(config.S3Folder), // don't confuse this with the access key
		Body:   bytes.NewReader(fileBody),
	})
	handleErr(err)

	return true, nil
}

func printInfoScreen() {
	fmt.Println("Please define the arguments as environment variables! Ex: \n\033[1;31mACCESSKEY\033[0m=AWSACCESSKEY \033[1;31mSECRETKEY\033[0m=AWSSECRETKEY \033[1;31mBUCKET\033[0m=BUCKETNAME \033[1;31mREGION\033[0m=REGION.... \033[1;32mgo run main.go\033[0m /path/to/file")
}

func main() {

	if config.S3Key == "" || config.S3Bucket == "" || config.S3Secret == "" || config.S3Region == "" {
		printInfoScreen()
		return
	}
	var (
		fileName string
		showHelp string
	)

	flag.StringVar(&showHelp, "h", "", "Show the usage")
	flag.StringVar(&fileName, "f", "", "The file to be uploaded")
	flag.Parse()

	if showHelp != "" {
		printInfoScreen()
		os.Exit(0)
	}

	ok, err := uploadFile(fileName)
	handleErr(err)
	if ok {
		fmt.Printf("File %s uploaded at s3://%s\n", fileName, config.S3Bucket)
	}
}
