#!/bin/bash

# 杀死进程
pkill -f httpserver
pkill -f gateserver
pkill -f loginserver
pkill -f gameserver
pkill -f chatserver

# 删除可执行文件
rm -f "../bin/httpserver"
rm -f "../bin/gateserver"
rm -f "../bin/gameserver"
rm -f "../bin/chatserver"

# 编译Go程序
go build -o ../bin/httpserver ../httpserver/cmd ; \
go build -o ../bin/gateserver ../gateserver/cmd ; \
go build -o ../bin/gameserver ../gameserver/cmd ; \
go build -o ../bin/chatserver ../chatserver/cmd