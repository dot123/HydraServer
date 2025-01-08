@echo off
swag fmt -g ../httpserver/cmd/main.go --exclude ./pkg/rabbitMQ
swag init -g ../httpserver/cmd/main.go -o ../httpserver/docs

wire ../httpserver/cmd
wire ../gameserver/cmd
wire ../chatserver/cmd
pause