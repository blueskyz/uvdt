# 分布式资源共享分发服务
* 分布式 bt tracker 服务
* 资源共享节点服务


<br/>

# 开发环境

* debian 8
* golang 1.8


<br/>

# 1. 分布式 bt tracker 服务器

uvdt-tracker

## 1.1 tracker 功能介绍

* 具备完整的 tracker 服务功能
* 提供 peer 节点信息
* 提供资源信息
* 多个 tracker 服务器之间可以集群工作


<br/>

# 2. 资源共享节点

uvdt-node

## 2.1 node 功能介绍

* 具备本地资源共享功能，计算共享目录的文件，创建 metainfo 文件信息
* 提供资源的分享功能，提供资源的下载功能
* 管理 peer 链接的资源传输信息


# 运行

## 3.1 tracker 服务器

bin/uvdt-tracker -clusterip="192.168.2.1:3333" -btserv "0.0.0.0:30080" -trackerserv "0.0.0.0:30081"

## 3.2 node 服务器

bin/uvdt-node -httpserv 0.0.0.0:8088 -rootpath ~/Downloads/movie

## 3.3 node tool 创建种子文件工具

bin/uvdt-node-tool -rootpath ~/Downloads/movie -respath share/walkingdead
