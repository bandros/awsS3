## AWS S3 Library for Go
How to using S3 in AWS

## Upload
file,_ := c.FormFile("file")\
img := awsS3.S3img{}\
img.Width = 200\
img.Prefix = "data_"\
img.AwsRegion = "ap-southeast-1"\
img.AwsKey = "Aws Key"\
img.AwsScreetKey = "Aws Screet Key"\
err := img.SetMulti(file)\
if err != nil {\
    fmt.Println("Set Img: ",err.Error())\
    return\
}\
location, err := img.Upload("bucket/folder")\
if err != nil {\
    fmt.Println("Upload: ",err.Error())\
    return\
}\
fmt.Println("Location: ",location)


## Multi Upload
form, _ := c.MultipartForm()\
files := form.File["file[]"]\
img := awsS3.S3img{}\
img.Width = 200\
img.Prefix = "multi_"\
img.AwsRegion = "ap-southeast-1"\
img.AwsKey = "Aws Key"\
img.AwsScreetKey = "Aws Screet Key"\
err := img.SetMulti(files)\
if err != nil {\
    fmt.Println("Set Img: ",err.Error())\
    return\
}
location, err := img.Upload("bucket/folder")\
if err != nil {
  fmt.Println("Upload: ",err.Error())\
  return\
}\
fmt.Println("Location: ",location)


**Router Framework**
https://github.com/gin-gonic/gin

**Creator**
https://github.com/tss182