# GeeCache
一个高可用的分布式缓存系统。实现了分布式缓存节点通信机制，并发访问控制机制，缓存节点动态感知机制，缓存保护机制，TTL机制，数据定期持久化机制。

项目技术点：

● 实现基于gin框架的服务器端、基于grpc+protobuf的服务器-分布式缓存节点通信机制

● 使用一致性哈希算法解决Key路由和负载均衡问题，使用SingleFlight算法合并重复请求防止缓存击穿问题

● 实现TTL机制+LRU混合淘汰策略降低内存占用

● 结合etcd的List+Watch机制，服务端可以高效、可靠的感知节点动态变化，实现毫秒级节点发现

● 设计数据持久化策略，实现节点热启动。解决由节点冷启动造成的大量缓存击穿与MySQL慢查询问题

![image] https://github.com/PengYu-Shi/GeeCache/blob/main/%E7%BB%93%E6%9E%84.jpg

# How to use

## What U Need：

etcd

GeeServer为服务器端，GeeNode为缓存节点
