#!/bin/bash

outPath=$1
if [ -z $outPath ]; then
    outPath="./"
fi

tag=$2
if [ -z $tag ]; then
    tag="def"
fi

# 私钥
openssl genrsa -out $outPath/rsa-private-$tag.key 2048
# 公钥
openssl rsa -in $outPath/rsa-private-$tag.key -pubout -out $outPath/rsa-public-$tag.key 
# 生成证书
openssl req -new -x509 -days 365 -key $outPath/rsa-private-$tag.key -out $outPath/cert-new-$tag.crt -subj "/C=CN/ST=FJ/L=FZ/O=ucx1024/OU=dev/CN=ucx1024.com/emailAddress=110@qq.com"