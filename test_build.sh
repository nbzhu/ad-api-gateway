#!/bin/bash

#/data/bin/test/ad-api-gateway/test_ad-api-gateway api -c=/data/etc/test/ad-api-gateway/conf -log=/data/log/test/ad-api-gateway -env=qa
# 检查是否提供了tag参数
if [ "$#" -ne 1 ]; then
    echo "Usage: $0 <tag>"
    exit 1
fi

tag=$1

# 切换到指定的tag
cd /data/src/ad-api-gateway
git checkout main
git pull
echo "Switching to tag $tag..."
git checkout $tag
git pull

# 注意：通常不需要在tag上执行git pull，因为tag是固定的。
# 如果你确实需要更新，可能需要先切换到某个分支，拉取更新，再切换回tag。
# git pull

# 复制文件
echo "Copying ..."
#cp -r /data/src/ad-api-gateway/conf /data/etc/test/ad-api-gateway
go build -v -o /data/bin/splay-admin/ad-api-gateway /data/src/ad-api-gateway/main.go

echo "build Done."

echo "start"
#nohup /data/bin/test/ad-api-gateway/test_ad-api-gateway api -c=/data/etc/test/ad-api-gateway/conf -log=/data/log/test/ad-api-gateway -env=qa  > /dev/null 2>&1 &
supervisorctl restart ad-api-gateway
echo "done."