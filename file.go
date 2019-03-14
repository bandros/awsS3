package awsS3

import (
	"bytes"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"os"
	"strings"
)

type S3File struct {
	location     string
	filename     string
	Prefix       string
	AwsKey       string
	AwsScreetKey string
	AwsRegion    string
}

func (f *S3File) Set(file string) {
	f.location = file
	var files = strings.Split(file, "/")
	f.filename = files[len(files)-1]
}

func (f *S3File) Upload(bucket string) (string, error) {
	var bucketSlice = strings.Split(bucket, "/")
	bucket = bucketSlice[0]
	var filepath = strings.Join(bucketSlice[1:], "/")
	filepath = strings.TrimRight(filepath, "/")
	sess, _ := session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Region:      aws.String(f.AwsRegion),
			Credentials: credentials.NewStaticCredentials(f.AwsKey, f.AwsScreetKey, ""),
		},
	})
	uploader := s3manager.NewUploader(sess)
	file, err := os.Open(f.location)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Get file size and read the file content into a buffer
	fileInfo, _ := file.Stat()
	var size int64 = fileInfo.Size()
	buffer := make([]byte, size)
	file.Read(buffer)
	filepath = filepath + "/" + f.filename
	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: &bucket,
		Key:    &filepath,
		Body:   bytes.NewReader(buffer),
		ACL:    aws.String("public-read"),
	})

	if err != nil {
		return "", err
	}
	return result.Location, nil
}
