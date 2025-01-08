#!/bin/bash

./etcdctl del --prefix ""

# 杀死进程
pkill httpserver
pkill gateserver
pkill gameserver
pkill chatserver
