#!/bin/bash

# 检查目标文件夹是否存在，如果存在则删除
if [ -d "../bin/data/" ]; then
    rm -rf "../bin/data/"
fi

# 复制源文件夹到目标文件夹
cp -r "../data/" "../bin/data/"

echo "复制完成"
