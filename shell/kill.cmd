@echo off
cd ../bin
cd ./nginx-1.24.0
nginx -s stop
cd ../
taskkill /F /IM httpserver.exe
taskkill /F /IM gateserver.exe
taskkill /F /IM gameserver.exe
taskkill /F /IM chatserver.exe
taskkill /F /IM nats-server.exe
taskkill /F /IM etcd.exe
taskkill /F /IM redis-server.exe
del /f /s /q *.log