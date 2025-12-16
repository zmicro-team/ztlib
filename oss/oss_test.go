//go-:build ignore

package oss

import (
	"context"
	"embed"
	_ "embed"
	"io"
	"os"
	"testing"
	"time"

	minIo "github.com/minio/minio-go/v7"
	"github.com/stretchr/testify/assert"
)

// TODO: 测试配置修改为环境变量或者测试配置文件

var TestCfg = OssUtilConfig{
	EndPoint:        "s3.local.uc1024.com", // aws s3 endpoint
	AccessKeyID:     "didong",
	SecretAccessKey: "didong123",
	BucketName:      "local-dev",
	Dir:             "temp",
	UseSSL:          true,
	// Region:          "ap-northeast-1",
}

//go:embed image.png
var tempFile embed.FS

func TestNewOss(t *testing.T) {
	ossUtil := NewOssUtil(TestCfg)
	assert.NotNil(t, ossUtil)
}

func TestMakeBucket(t *testing.T) {
	ossUtil := NewOssUtil(TestCfg)
	bucketName := "local-dev"
	res := ossUtil.MakeBucket(bucketName)
	t.Logf("%v", res)
}

func TestS3GenerateAWSPutTempUrl1(t *testing.T) {
	ossUtil := NewOssUtil(TestCfg)
	dir := "image_02.png"
	token, err := ossUtil.GenerateAWSPutTempUrl(dir, time.Hour)
	if err != nil {
		t.Errorf("failed to generate token: %v", err)
		return
	}
	t.Logf("token: %v", token)
}

func TestS3GenerateAWSGetTempUrl1(t *testing.T) {
	ossUtil := NewOssUtil(TestCfg)
	ob, err := ossUtil.Client.GetObject(context.Background(), ossUtil.Config.BucketName, "temp/image_02.png", minIo.GetObjectOptions{})
	if err != nil {
		t.Errorf("failed to get object: %v", err)
		return
	}
	bus, err := io.ReadAll(ob)
	if err != nil {
		t.Errorf("failed to read object: %v", err)
		return
	}
	os.WriteFile("image_02.png", bus, 0644)
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
