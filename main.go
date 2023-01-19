package main

import (
	"fmt"
	"log"
	"regexp"
	"test/db"

	"github.com/hpcloud/tail"
	"github.com/joho/godotenv"
	"os"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("error: no .env file found")
	}
}

func main() {

	mongoURI := os.Getenv("MONGO_URI")
	mongoDB := os.Getenv("MONGO_DB")
	mongoErrCollection := os.Getenv("MONGO_ERROR_COLLECTION")
	mongoAccCollection := os.Getenv("MONGO_ACCESS_COLLECTION")
	nginxAccFilePath := os.Getenv("NGINX_ACCESS_FILE_PATH")
	nginxErrFilePath := os.Getenv("NGINX_ERROR_FILE_PATH")

	logsAccessFormat := `$remote_addr - $remote_user \[$time_stamp\]  $response_code \"$http_method $request_path $http_version\" $body_bytes_sent \"$http_referer\" \"$http_user_agent\" $X_Forwarded_For`
	regexAccessFormat := regexp.MustCompile(`\$([\w_]*)`).ReplaceAllString(logsAccessFormat, `(?P<$1>.*)`)
	AccessRe := regexp.MustCompile(regexAccessFormat)

	logsErrorFormat := `$time_stamp \[$level\] $pid#$tid: \*$message, client: $client, server: $server, request: \"$http_method $request_path $http_version\", upstream: \"$upstream\", host: \"$host\"`
	regexErrorFormat := regexp.MustCompile(`\$([\w_]*)`).ReplaceAllString(logsErrorFormat, `(?P<$1>.*)`)
	ErrorRe := regexp.MustCompile(regexErrorFormat)

	conToErr, err := db.ConnectToCollection(mongoURI, mongoDB, mongoErrCollection)

	if err != nil {
		log.Fatal(err)
	}

	errLogger := db.NewLogger(conToErr)

	conToAcc, err := db.ConnectToCollection(mongoURI, mongoDB, mongoAccCollection)

	if err != nil {
		log.Fatal(err)
	}

	accLogger := db.NewLogger(conToAcc)

	error_file, err := tail.TailFile(nginxErrFilePath, tail.Config{Follow: true, Location: &tail.SeekInfo{Offset: 0, Whence: 2}})
	if err != nil {
		log.Fatal(err)
	}

	access_file, err := tail.TailFile(nginxAccFilePath, tail.Config{Follow: true, Location: &tail.SeekInfo{Offset: 0, Whence: 2}})
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case line := <-access_file.Lines:
			matches := AccessRe.FindStringSubmatch(line.Text)
			if len(matches) != 0 {
				logs := make(map[string]string)
				for i, k := range AccessRe.SubexpNames() {
					if i == 0 {
						continue
					}
					fmt.Printf("%-15s => %s\n", k, matches[i])
					logs[k] = matches[i]
				}
				accLogger.SendLog(logs)
			}
		case line := <-error_file.Lines:
			matches := ErrorRe.FindStringSubmatch(line.Text)
			if len(matches) != 0 {
				logs := make(map[string]string)
				for i, k := range ErrorRe.SubexpNames() {
					if i == 0 {
						continue
					}
					fmt.Printf("%-15s => %s\n", k, matches[i])
					logs[k] = matches[i]
				}
				errLogger.SendLog(logs)
			}
		}
	}

}
