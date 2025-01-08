@echo off
taskkill /F /IM httpserver.exe
taskkill /F /IM gateserver.exe
taskkill /F /IM gameserver.exe
taskkill /F /IM chatserver.exe

del /a/f/q "../bin/httpserver.exe"
del /a/f/q "../bin/gateserver.exe"
del /a/f/q "../bin/gameserver.exe"
del /a/f/q "../bin/chatserver.exe"

go build -o ../bin/httpserver.exe   ../httpserver/cmd
go build -o ../bin/gateserver.exe   ../gateserver/cmd
go build -o ../bin/gameserver.exe   ../gameserver/cmd
go build -o ../bin/chatserver.exe   ../chatserver/cmd