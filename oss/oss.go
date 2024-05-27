package oss

import (
	"context"
	"fmt"
	"time"

	minIo "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type OssUtilConfig struct {
	EndPoint        string `json:"end_point" yaml:"endPoint"`
	Region          string `json:"region" yaml:"region"`
	AccessKeyID     string `json:"access_key_id" yaml:"accessKeyId"`
	SecretAccessKey string `json:"secret_access_key" yaml:"secretAccessKey"`
	BucketName      string `json:"bucket_name" yaml:"bucketName"`
	Dir             string `json:"dir" yaml:"dir"`
	UseSSL          bool   `json:"use_ssl" yaml:"useSsl"`
}

type OssUtil struct {
	Config OssUtilConfig
	Client *minIo.Client
}

func NewOssUtil(config OssUtilConfig) *OssUtil {
	client, err := minIo.New(config.EndPoint, &minIo.Options{
		Creds:  credentials.NewStaticV4(config.AccessKeyID, config.SecretAccessKey, ""),
		Secure: config.UseSSL,
		Region: config.Region,
	})
	if err != nil {
		panic(err)
	}
	return &OssUtil{
		Config: config,
		Client: client,
	}
}

type PolicyToken struct {
	AccessKeyId string `json:"accessKeyId"`
	Bucket      string `json:"bucket"`
	Region      string `json:"region"`
	Path        string `json:"path"`
	Expire      int64  `json:"expire"`
	Key         string `json:"key,omitempty"`
}

// 创建一个bucket
func (o *OssUtil) MakeBucket(bucket string) bool {
	err := o.Client.MakeBucket(context.Background(), bucket, minIo.MakeBucketOptions{Region: o.Config.Region})
	if err != nil {
		// Check to see if we already own this bucket (which happens if you run this twice)
		exists, errBucketExists := o.Client.BucketExists(context.Background(), bucket)
		if errBucketExists == nil && exists {
			return true
		} else {
			return false
		}
	} else {
		return true
	}
}

// 返回一个临时的上传链接
func (o *OssUtil) GenerateAWSPutTempUrl(fileName string, expires time.Duration) (PolicyToken, error) {
	bucket := o.Config.BucketName
	dir := fmt.Sprintf("%s/%s", o.Config.Dir, fileName)
	u, err := o.Client.PresignedPutObject(context.Background(), bucket, dir, expires)
	if err != nil {
		return PolicyToken{}, err
	}
	return PolicyToken{
		AccessKeyId: o.Config.AccessKeyID,
		Bucket:      bucket,
		Region:      o.Config.Region,
		Expire:      time.Now().Add(expires).Unix(),
		Path:        dir,
		Key:         u.String(),
	}, nil
}
