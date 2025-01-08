# Hydra Game Server

> 一个基于 [Pitaya](https://github.com/topfreegames/pitaya) 开发的高性能、可扩展的游戏服务器脚手架。
> 
> 类似九头蛇，多个服务器协同工作，组成强大的游戏服务生态。该架构包括网关服务器、HTTP服务器（用于用户注册）、游戏服务器和聊天服务器。

## 系统架构

### 功能说明

1. **Gateway Server（网关服务器）**
   - 负责请求转发和负载均衡
   - 维护客户端连接
   - 路由分发到不同的服务器
   - 数据压缩传输（msgpack）

2. **HTTP Server（用户注册服务器）**
   - 提供用户注册接口
   - 用户数据持久化到 MySQL
   - 注册成功后返回用户凭证

3. **Game Server（游戏服务器）**
   - 处理核心游戏逻辑
   - 游戏状态同步
   - 数据持久化

4. **Chat Server（聊天服务器）**
   - 处理即时通讯功能
   - 消息转发
   - 在线状态管理

### 架构图

```
                    客户端
                      │
          ┌──────────┴───────────┐
          │                      │
          ▼                      ▼
     HTTP服务器              网关服务器 ◄─────┐
          │                   ▲  │           │
          │                   │  │           │
          ▼                   │  ▼           │
        MySQL            游戏服务器 ◄────► 聊天服务器
                            │  │           │
                            │  │           │
                            ▼  ▼           ▼
                          MySQL Redis     Redis

```

- **Gateway Server**: 网关服务器，负责请求转发和负载均衡
- **HTTP Server**: Web API服务器，处理HTTP请求
- **Game Server**: 游戏服务器，处理核心游戏逻辑
- **Chat Server**: 聊天服务器，处理即时通讯功能

### 数据流说明

1. **客户端接入**
   - 注册流程：
     - 客户端 → HTTP服务器：发送注册请求
     - HTTP服务器 → MySQL：存储用户数据
     - HTTP服务器 → 客户端：返回注册结果
   - 游戏流程：
     - 客户端 → 网关服务器：游戏和聊天功能
     - 网关服务器：负责转发到对应的服务器

2. **服务层通信**
   - HTTP服务器：
     - 只处理用户注册
     - 与 MySQL 交互存储用户数据
     - 不参与游戏逻辑通信
   - 游戏相关服务器（通过 NATS）：
     - 网关服务器 ←→ 游戏服务器：实时游戏数据、玩家状态同步
     - 网关服务器 ←→ 聊天服务器：聊天消息、在线状态
     - 游戏服务器 ←→ 聊天服务器：游戏内聊天、队伍通信

3. **存储层**
   - MySQL：
     - 存储用户注册信息（HTTP服务器）
     - 存储游戏数据和玩家信息（游戏服务器）
   - Redis：
     - 游戏服务器：游戏状态缓存、临时数据
     - 聊天服务器：会话管理、消息缓存

4. **消息队列**
   - NATS：作为核心消息队列，实现所有服务器之间的互相通信

## 环境要求

### 必需组件

1. **Go 1.22**
   - Windows: 下载安装 [go1.22.windows-amd64.msi](https://golang.org/dl/)
   - Linux: `wget https://golang.org/dl/go1.22.linux-amd64.tar.gz`

2. **MySQL 5.7**
   - Windows: [MySQL Community Server 5.7](https://downloads.mysql.com/archives/community/)
   - Linux: `sudo apt-get install mysql-server-5.7`

3. **NATS 2.9.23**
   - Windows: [nats-server-v2.9.23-windows-amd64.zip](https://github.com/nats-io/nats-server/releases/tag/v2.9.23)
   - Linux: `wget https://github.com/nats-io/nats-server/releases/download/v2.9.23/nats-server-v2.9.23-linux-amd64.tar.gz`

4. **Redis 7.2.5**
   - Windows: [Redis-Windows](https://github.com/zkteco-home/redis-windows/archive/refs/tags/7.2.5.0.zip)
   - Linux: `sudo apt-get install redis`

5. **etcd 3.5**
   - Windows: [etcd-v3.5.10-windows-amd64.zip](https://github.com/etcd-io/etcd/releases/tag/v3.5.10)
   - Linux: `sudo apt-get install etcd`

### 配置说明

1. MySQL 配置
   - 默认端口：3306
   - 创建数据库和用户
   - 导入初始化SQL脚本

2. NATS 配置
   - 默认端口：4222
   - 集群配置（可选）

3. Redis 配置
   - 默认端口：6379
   - 配置密码（推荐）

## 快速开始

### Windows
```bash
# 构建
./shell/build.cmd

# 运行
./shell/run.cmd

# 停止服务
./shell/kill.cmd
```

### Linux
```bash
# 构建
./debian-shell/build.sh

# 运行
./debian-shell/run.sh

# 停止服务
./debian-shell/kill.sh
```

## 配置文件

所有配置文件位于 `data/conf/` 目录下：
- `gateserver.toml`: 网关服务器配置
- `httpserver.toml`: HTTP服务器配置
- `gameserver.toml`: 游戏服务器配置
- `chatserver.toml`: 聊天服务器配置
