# RINP: RINP Is Not a Proxy

## 原则

- 切换代理期间服务不中断（无感知）
- 不要改变原有业务代码（只能新增）
- 不要引入新的安全问题

## 待办清单

- [x] 实现 client 与 sidecar 的封包后的报文抓取（采用UDP作为L2）
- [x] tun 实现 user 与 service 基于封包的正常通信
- [x] 多用户支持
- [x] 性能分析（docker run uber/go-torch -u http://<ip>:8080/debug/pprof -p -t=8 > torch.svg)
- [x] 经由 proxy 实现 user 与 service 基于封包的正常通信
- [x] 经由动态切换的 proxy 实现 user 与 service 基于封包的正常通信
- [x] 通过 auth 模块接入验证
- [ ] 通过 scheduler 与 controller 的调度生成相关 proxy
- [x] proxy 定期切换，用户服务无感知，防御住僵尸流量
- [ ] 引入洗牌算法+评分机制，筛选出间谍用户 （采用现成算法）
- [ ] 包传送过程中仿照 JWT进行加密、解密：签名算法、数据、签名算法
- [ ] 跨平台支持 (tun)
- [ ] 采用自定义方法实现洗牌算法+评分机制，高效率筛选出间谍用户

## 论文列表

- [ ] **网络区**: 移动目标防御的一种工程实现
- [ ] **网络区**: 筛选间谍用户的算法更新
- [ ] **软工区**: 实现机制性能损耗分析
- [ ] **软工区**: 结伴编程

## Redis Reference

### DB0

Stores information about clients:

- Virtual IP (key)
- Valid Proxy IP **list** (Note: when implementing proxy, we should take network latency into consideration, i.e. from the schedulers' message to proxy)

Note: 

- [x] Auth module will be setting expiration time according to the expiration time of the JWT token when clients logging in. When the client renews its token, the expiry time should be updated.
- [x] Proxies will be watching this, so they know which clients are valid. Also, proxies should use client side caching to reduce the number of requests to Redis when inspecting packets. TODO: currently we use built-in client side caching. This is not what we want
- [ ] Scheduler will update the proxy information when it reschedules clients.

### DB1

Stores information about proxies:

- Proxy name (key)
- Public IP address : port

Note: 

- [x] Proxies should update their key every 1s when they are alive. Keys should have a TTL of 2s, so that we can remove the proxy when it is down. 
- [ ] Scheduler will be watching this, so it knows which proxy is alive and assign clients to them.
- [ ] Auth module will also use this information to choose the first proxy.

### DB2

Stores information about sidecars:

- Virtual IP (key)
- Actual IP : port

Note:

- [x] Sidecar should update their key every 5s when they are alive. Keys should have a TTL of 10s, so that we can remove the sidecar when it is down. 
- [ ] Proxies will be watching this, so they know how to route traffic.
