package my9s3

import (
	"bytes"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type S3Session struct {
	Svc *s3.S3
}
type My9AWSSession session.Session

type PutObjectIn struct {
	Bucket string
	Key    string
}

func NewS3Session(sess client.ConfigProvider, region string) (s3client S3Session, err error) {
	s3client.Svc = s3.New(sess, aws.NewConfig().WithRegion(region))
	return s3client, err
}

func (s3client *S3Session) UploadFile(s3In PutObjectIn, file string) (resp *s3.PutObjectOutput, err error) {
	fd, err := os.Open(file)
	if err != nil {
		fmt.Printf("err opening file: %s", err)
	}
	fileInfo, _ := fd.Stat()
	var filesize int64 = fileInfo.Size()
	buffer := make([]byte, filesize)
	fd.Read(buffer)
	fileBytes := bytes.NewReader(buffer)

	defer fd.Close()
	params := &s3.PutObjectInput{
		Bucket:        aws.String(s3In.Bucket), // Required
		Key:           aws.String(s3In.Key),    // Required
		Body:          fileBytes,
		ContentLength: aws.Int64(filesize),
	}
	resp, err = s3client.Svc.PutObject(params)

	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		fmt.Println(err.Error())
		return
	}

	return resp, err
}
