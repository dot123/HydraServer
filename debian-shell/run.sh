#!/bin/bash

# 杀死进程
pkill -f httpserver
pkill -f gateserver
pkill -f gameserver
pkill -f chatserver

# 暂停一会
sleep 2

cd ../bin

# 启动服务器进程
./httpserver -port 9001 -serverid 1 &
./gateserver -ip 127.0.0.1 -port 8001 -serializer msgpack -grpc 0 -grpchost 127.0.0.1 -grpcport 3001 -serverid 1 &
./gameserver -serializer msgpack -grpc 0 -grpchost 127.0.0.1 -grpcport 5001 -serverid 1 &
./chatserver -serializer msgpack -grpc 0 -grpchost 127.0.0.1 -grpcport 6001 -serverid 1 &