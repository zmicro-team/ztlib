package oss

import (
	"context"
	"embed"
	_ "embed"
	"testing"
	"time"

	minIo "github.com/minio/minio-go/v7"
	"github.com/stretchr/testify/assert"
)

var TestCfg = OssUtilConfig{
	EndPoint:        "s3.amazonaws.com", // aws s3 endpoint
	AccessKeyID:     "your_access_key_here",
	SecretAccessKey: "your_secret_key_here",
	BucketName:      "zlbgame",
	Dir:             "temp",
	UseSSL:          true,
	Region:          "us-east-1",
}

//go:embed image.png
var tempFile embed.FS

func TestNewOss(t *testing.T) {
	ossUtil := NewOssUtil(TestCfg)
	assert.NotNil(t, ossUtil)
}

func TestMakeBucket(t *testing.T) {
	ossUtil := NewOssUtil(TestCfg)
	bucketName := "zlbgame"
	res := ossUtil.MakeBucket(bucketName)
	t.Logf("%v", res)
}

func TestS3GenerateAWSPutTempUrl(t *testing.T) {
	ossUtil := NewOssUtil(TestCfg)
	dir := "temp/temp_img.png"
	token, err := ossUtil.GenerateAWSPutTempUrl(dir, time.Hour)
	if err != nil {
		t.Errorf("failed to generate token: %v", err)
		return
	}
	t.Logf("token: %v", token)
}

func TestS3Put(t *testing.T) {
	ossUtil := NewOssUtil(TestCfg)
	bucketName := ossUtil.Config.BucketName
	contentType := "image/png"
	info, err := ossUtil.Client.FPutObject(context.Background(), bucketName, "temp/image_02.png", "./image.png", minIo.PutObjectOptions{ContentType: contentType})
	if err != nil {
		t.Errorf("failed to upload file: %v", err)
		return
	}
	t.Logf("upload success, %v", info)
}
