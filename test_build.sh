#!/bin/bash

# 切换到指定的tag
cd /data/src/ad-api-gateway/proto
git checkout main
git pull

cd ../
git checkout main
git pull

# 复制文件
echo "Copying ..."
#cp -r /data/src/ad-api-gateway/conf /data/etc/test/ad-api-gateway
go build -v -o /data/bin/splay-admin/ad-api-gateway /data/src/ad-api-gateway/main.go

echo "build Done."

echo "start"
#nohup /data/bin/test/ad-api-gateway/test_ad-api-gateway api -c=/data/etc/test/ad-api-gateway/conf -log=/data/log/test/ad-api-gateway -env=qa  > /dev/null 2>&1 &
supervisorctl restart ad-api-gateway
echo "done."