package awsS3

import (
	"bytes"
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/disintegration/imaging"
	"image"
	"image/png"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

type S3img struct {
	img          image.Image
	imgMulti     []image.Image
	file         *multipart.FileHeader
	fileMulti    []*multipart.FileHeader
	Prefix       string
	Width        int
	Height       int
	AwsKey       string
	AwsScreetKey string
	AwsRegion    string
}

type ListObject struct{
	Fulpath string
	Folder string
	File string
	LastModif time.Time
	Size uint64
	IsFolder bool
}

func (img *S3img) Set(file *multipart.FileHeader) error {
	f, err := file.Open()
	if err != nil {
		return err
	}
	defer f.Close()
	src, err := imaging.Decode(f)
	if err != nil {
		return err
	}
	img.img = src
	img.file = file
	return nil
}

func (img *S3img) SetMulti(files []*multipart.FileHeader) error {
	for _, file := range files {
		f, err := file.Open()
		if err != nil {
			return err
		}
		defer f.Close()
		src, err := imaging.Decode(f)
		if err != nil {
			return err
		}
		img.imgMulti = append(img.imgMulti, src)
	}
	img.fileMulti = files
	return nil
}

func (img *S3img) Upload(bucket string) ([]string, error) {
	if img.file == nil && img.fileMulti == nil {
		return nil, errors.New("file not found")
	} else if img.file != nil && img.fileMulti != nil {
		return nil, errors.New("file or fileMulti not nil")
	}
	var bucketSlice = strings.Split(bucket, "/")
	bucket = bucketSlice[0]
	var filepath = strings.Join(bucketSlice[1:], "/")
	filepath = strings.TrimRight(filepath, "/")
	if img.file != nil {
		//single upload
		var width, height = getSize(img.img.Bounds())
		if (img.Width > 0 || img.Height > 0) && img.Width < width && img.Height < height {
			img.img = imaging.Resize(img.img, img.Width, img.Height, imaging.Lanczos)
		}
	} else if img.fileMulti != nil {
		//multiupload
		for i, v := range img.imgMulti {
			var width, height = getSize(v.Bounds())
			if (img.Width > 0 || img.Height > 0) && img.Width < width && img.Height < height {
				img.imgMulti[i] = imaging.Resize(v, img.Width, img.Height, imaging.Lanczos)
			}
		}
	}
	sess, _ := session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Region:      aws.String(img.AwsRegion),
			Credentials: credentials.NewStaticCredentials(img.AwsKey, img.AwsScreetKey, ""),
		},
	})
	uploader := s3manager.NewUploader(sess)
	if img.file != nil {
		location, err := upload(uploader, img.file, img.img, img.Prefix, filepath, bucket)
		var locationSlice = []string{0: location}
		return locationSlice, err
	} else {
		var locationSlice = []string{}
		var location string
		var err error
		for i, imgs := range img.imgMulti {
			location, err = upload(uploader, img.fileMulti[i], imgs, img.Prefix, filepath, bucket)
			locationSlice = append(locationSlice, location)
		}
		return locationSlice, err
	}

}

func (img *S3img) Delete(file string) error {
	var bucketSlice = strings.Split(file, "/")
	var bucket = bucketSlice[0]
	var filepath = strings.Join(bucketSlice[1:], "/")
	filepath = strings.TrimRight(filepath, "/")
	sess, _ := session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Region:      aws.String(img.AwsRegion),
			Credentials: credentials.NewStaticCredentials(img.AwsKey, img.AwsScreetKey, ""),
		},
	})
	svc := s3.New(sess)
	params := &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(filepath),
	}
	_, err := svc.DeleteObject(params)
	if err != nil {
		return err
	}
	return nil
}

func getSize(img image.Rectangle) (int, int) {
	b := img.Bounds()
	width := b.Max.X
	height := b.Max.Y
	return width, height
}

func getFileContentType(out multipart.File) (string, error) {
	buffer := make([]byte, 512)
	_, err := out.Read(buffer)
	if err != nil {
		return "", err
	}
	contentType := http.DetectContentType(buffer)
	return contentType, nil
}

func upload(uploader *s3manager.Uploader, file *multipart.FileHeader, img image.Image, prefix, filepath, bucket string) (string, error) {
	var err error
	var key = filepath + "/" + prefix + file.Filename
	var f, _ = file.Open()
	contentType, _ := getFileContentType(f)
	buf := new(bytes.Buffer)
	switch contentType {
	case "image/png":
		err = imaging.Encode(buf, img, imaging.PNG, imaging.PNGCompressionLevel(png.BestCompression))
	case "image/jpeg":
		err = imaging.Encode(buf, img, imaging.JPEG,imaging.JPEGQuality(75))
	case "image/gif":
		err = imaging.Encode(buf, img, imaging.GIF)
	case "image/bmp":
		err = imaging.Encode(buf, img, imaging.BMP)
	}

	if err != nil {
		return "", err
	}

	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket:      &bucket,
		Key:         &key,
		Body:        bytes.NewReader(buf.Bytes()),
		ACL:         aws.String("public-read"),
		ContentType: &contentType,
	})

	if err != nil {
		return "", err
	}

	return result.Location, nil
}


func (img *S3img) List(bucket string) ([]ListObject, error) {
	bucket = strings.TrimRight(bucket,"/")
	var bucketSlice = strings.Split(bucket, "/")
	bucket = bucketSlice[0]
	var filepath = strings.Join(bucketSlice[1:], "/")
	filepath = strings.TrimRight(filepath, "/")

	sess, _ := session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Region:      aws.String(img.AwsRegion),
			Credentials: credentials.NewStaticCredentials(img.AwsKey, img.AwsScreetKey, ""),
		},
	})
	svc := s3.New(sess)
	params := &s3.ListObjectsInput{
		Bucket: aws.String(bucket),
		Prefix:   aws.String(filepath),
	}
	resp,err := svc.ListObjects(params)

	if err != nil {
		return []ListObject{} , err
	}
	var list = []ListObject{}
	for _, key := range resp.Contents {
		var isFolder = false
		if *key.Size == 0 {
			isFolder = true
		}
		var folder = *key.Key
		if !isFolder {
			var fullpath = strings.Split(folder,"/")
			folder = strings.Join(fullpath[:len(fullpath)-1],"/")
		}
		var fileSlice = strings.Split(*key.Key,"/")
		list = append(list,ListObject{
			Fulpath: *key.Key,
			Folder: folder,
			File: fileSlice[len(fileSlice)-1],
			Size: uint64(*key.Size),
			LastModif: *key.LastModified,
			IsFolder: isFolder,
		})
	}
	return list,nil
}