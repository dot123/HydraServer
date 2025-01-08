@echo off

cd ../bin

taskkill /F /IM httpserver.exe
taskkill /F /IM gateserver.exe
taskkill /F /IM gameserver.exe
taskkill /F /IM chatserver.exe
taskkill /F /IM nats-server.exe
taskkill /F /IM etcd.exe
taskkill /F /IM redis-server.exe

del /f /s /q .\dump.rdb
rmdir .\default.etcd\ /s /q

start nats-server.exe
start etcd.exe
start redis-server.exe ./redis.conf

timeout /t 2 >nul

start httpserver.exe -port 9001 -serverid 1
start gateserver.exe -ip 127.0.0.1 -port 8001 -serializer json -grpc 0 -grpchost 127.0.0.1 -grpcport 3001 -serverid 1
start gameserver.exe -serializer json -grpc 0 -grpchost 127.0.0.1 -grpcport 5001 -serverid 1
start chatserver.exe -serializer json -grpc 0 -grpchost 127.0.0.1 -grpcport 6001 -serverid 1