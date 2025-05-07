# wgxDouYin
## 框架图

![程序框架图](/readme/程序框架图.jpg)

## Raft

### 概述

分布式系统设计的目的一般来说有三个：

1. 保证在某些机器挂掉的情况下，其他存活的机器能够确保系统的正常运行；
2. 减轻其他机器的负担，各种不修改机器状态的命令能够负载在每一台机器上；
3. 降低命令从请求到回复的网络延迟，例如一个系统将机器分布在全国各地，每条命令都选择最近的机器去执行，现实应用如CDN。

那么，根据这三个目的，就引入了一个问题：根据上述3，如何确保在海南岛上的用户能和在漠河的用户看到一样的内容？即确保一致性。

Raft是一个确保分布式系统一致性的解决方案。在分布式系统中，多台机器之间协同工作，但如果面向用户，则它们看起来就像是只有一台机器。因此它们之间要解决同步问题，例如当用户A将这个系统中的某个参数Parameter通过某台机器从“NOW”改为“AFTER”后，其他用户的Query能够查看到这个改变，而不是一些Query查看到的是“NOW”，一些Query查看到则是“AFTER”。

### 选举

为了确保系统之间能够同步，Raft为系统中的各个服务器确定了三个身份Leader、Follower、Candidate。Leader在系统中只有一个，它主要负责接收命令，然后再将命令分发给其他机器；Follower是负责接收这些命令并存储的机器；Candidate则是尝试成为Leader的Follower机器。

有了上面的介绍，则引入如下问题：

1. Follower在什么情况下尝试变为Candidate？
2. Candidate如何成为Leader？

Answer For Q1:

因为Leader在系统中只有一个，所有修改分布式系统内部状态的命令则都会通过Leader确认执行并返回，所以当他挂掉时，整个分布式系统就无法正常运行。因此Raft允许分布式系统中的每一个满足的条件Follower都能够成为Leader，借此来确保系统的可用性。Leader会为每一个Follower发送一个存活确认包Heartbeat Packet，而Follower会开启一个线程监听这个Packet，当Follower在一个容忍时限内未能接收到Packet时，他会认为这个系统的Leader已经挂掉了，它尝试成为Leader，此时它会把自己的身份变成Candidate。系统允许存在多个Candidate，但只会一个Candidate最终成为Leader。

Answer For Q2:

Follower成为Candidate后，它会为自己加上一票，然后发送一个RequestVote，其他Follower收到RequestVote后，则查看自己是否已经投过票，如果未投过，则投给这个Candidate。续…

### 同步

同时系统为每条**命令的日志**也确认了两种状态Commited、Uncommitted、Apply。Commited为那些已经提交的命令，Raft确保Commited的命令日志一定会被Apply，接回概述中的例子即现在系统能够确保Query到的一定为“AFTER”。

3.为什么不是命令而是命令的日志被提交？

### 持久化

…

### 快照

…

## 环境问题

Q1：navicat无法连接到docker内的mysql

A1：最终发现时ufw中的配置文件问题，修改/etc/default/ufw中的DEFAULT_FORWARD_POLICY解决问题

## 数据库问题

主从复制

MySQL 出于安全原因，不允许使用权限为 `777` 或其他可被所有用户写入的配置文件。这会导致配置文件未被加载，从而无法应用其中的设置，包括 `server-id`。

因此设置配置文件仅被所有者写入

chmod 644 /etc/mysql/conf.d/my.cnf

常用命令

```bash
create user 'replica'@'%' identified with mysql_native_password by '1477364283';
grant replication slave on *.* to 'replica'@'%';
FLUSH PRIVILEGES;
show master status;
binlog.000002 833

change replication source to source_host='mysql-master', source_user='replica', source_password='1477364283', source_log_file='binlog.000002', source_log_pos=833;
change replication source to source_host='mysql-master', source_user='replica', source_password='1477364283', MASTER_AUTO_POSITION = 1;

start replica;
reset replica;
stop replica;

mysql-master : 172.20.0.3
mysql-replica1: 172.18.0.2
docker exec -it mysql-replica1 bash
docker-compose -f mysql.yml down
docker-compose -f mysql.yml up -d
docker inspect mysql-master
rm -rf mysql-data-replica1/*
rm -rf mysql-data-master/*
FLUSH PRIVILEGES;
```

## RPC问题

### etcd

1. etcd v2采用轮询模式，使用http/1.x协议，通过长连接定时轮询server。
2. multi-version concurrency control（MVCC）是etcd的版本控制方式，每个key都有一个独立的版本号revision number，这样就最大化读写的并发度。同时etcd在同步写时，etcd确保仅有一个写操作被执行，其它的写操作将会被拒绝。
3. etcd watcher可以设置版本号，watcher则会监听所有大于等于该版本号的内容。

### grpc

1. grpc会实现一个resolver，就像域名最终会被域名服务器转换为ip地址一样。grpc client的连接名也会被转化为具体的ip地址。
2. 和etcd联合使用的问题：为了使grpc能够做动态的服务发现，可将其解析仓库变为etcd，让etcd客户端监听，当etcd中某个服务的地址发生变化时，etcd客户端能够通过其watcher通知grpc。
   - 需求：希望将etcd与grpc松耦合，因为在初始化一个grpc的resolver时传入一个ip地址非常奇怪。同时松耦合后，这个etcd还能够用于其它需要查询etcd key、value的应用。比如后面还会查询服务的公钥。
   - 方案：因此考虑设计一个interface，用于将etcd client查询到的键值对同步更新到resolver以及key manager当中。只要resolver、keyManager实现了这个interface，当使用etcd时，只需向其注册自身。
   - 问题：etcd client实现了Watch函数用于监听特定的keyPrefix在服务端发生的变化，同时etcd client会在第一次调用rpc服务时初始化。当其初始化时，etcd client会通过Watcher监听etcd服务器某个keyPrefix的变化。我的写法是当其发生变化时，会同步调用resolver、key manager的update，但是resolver的Builder仅会在gRPC发器服务解析时才会被调用，因此服务的发现会先于解析器的创建，因此需要一个暂存服务地址的空间。

常用命令：protoc --proto_path=. --go_out=./ --go-grpc_out=./ --go_opt=Muser.proto=wgxDouYin/grpc/user relation.proto

生成grpc

## ApiRouter

### Middleware

1. TokenAuthMiddleware

   参数：

   - serviceDependencyMap：服务依赖映射
   - keys：密钥管理对象
   - skipRoutes：中间件忽略列表，用于中间件判断是否跳过该请求执行中间件函数

   函数运行过程：

   - 根据请求路径以及忽略列表判断是否跳过该请求
   - 根据请求路径获取请求服务名
   - 根据请求体获取JWT，根据请求服务名根据serviceDependencyMap获取其依赖服务
   - 利用keys中的依赖服务公钥判断该JWT是否由该服务签署，即验证请求的合法性
   - 合法则保存JWT中包含的用户名，并传递给下一个中间件；不合法则返回错误

   使用公私钥的原因如下：

   - ApiRouter需要知晓每一个服务的对称密钥（通过Etcd），因此如果它发生某些安全事故，就会导致所有服务方的签发密钥泄漏，进而导致伪造请求的出现

   不使用一个密钥的原因如下：

   - 服务之间无法验证是否合法。例如某个服务A需要在进行服务B之后才能够完成操作，但是若使用一个密钥进行鉴权，则服务A要一直等到RPC请求被解析后才能够确认该服务不合法。但是如果使用公私钥进行鉴权，同时定义一套规则链给API Router，API Router按照规则链查询对应的公钥进行解析请求，如果解析失败（即服务B的公钥无法解析该请求），说明该请求不合法
   - 服务存在伪造的可能。因为所有的token都通过一个密钥进行签名，因此服务B可能知道了服务A的请求规则，于是它伪造了一个请求并进行签名，服务A接收后确认该服务合法执行服务B伪造的操作

   因此在进行服务调用前，必须使用查看该服务的依赖服务，再查询依赖服务的公钥，用此公钥验证请求中的jwt是否合法。

2. ServiceAvailabilityMiddleware

参数：

- serviceNameMap：服务依赖映射
- keys：密钥管理对象

函数运行过程：

- 根据请求路径适用serviceNameMap获取服务名ServiceName
- 查看keys中对应ServiceName的公钥是否存在，依据此判断该服务是否上线，若未上线，返回错误，否则传入下一个中间件

## User

### cmd

1. UserRegister

   函数运行过程：

   - 读取RPC Request发送来的请求体，获取用户名、用户密码
   - 为密码加密，主要通过Argon2+Salt的方式预防暴力破解与彩虹表攻击，盐值拼接在密码前（安全性待考量），方便后续比较密码
   - 调用数据库包的CreateUser函数在持久化数据库MySQL创建用户
   - 返回RPC Response

   Argon2：可以指定函数的生成密钥的时间与空间复杂度，降低破解成本

2. Login

   函数运行过程：

   - 读取RPC Request发送来的请求体，获取用户名、用户密码
   - 使用用户名调用MySQL查询函数查询用户是否存在
   - 若存在则比较密码，主要利用上面的密码生成函数进行比较
   - 返回RPC Response

   改进：针对粉丝增长最近比较快的用户，将其用户ID常驻Redis，避免频繁调用MySQL查询。利用粉丝增长数量加上ZSet做一个热点用户排行榜（未完成）

3. UserInfo

   函数运行过程：

   - 读取RPC Request发送来的请求体，获取用户名
   - 使用用户名调用MySQL查询函数查询用户是否存在
   - 若存在则返回User信息

## Relation

### cmd

1. RelationAction

函数运行过程：

- 读取RPC Request发送来的请求体，获取TokenUserId, ToUserId，ActionType
- 判断TokenUserId是否与ToUserId相同，相同则说明用户试图关注自己，返回错误
- 根据RelationActionType作出如下操作：
   - 若RelationActionType为关注，则使用MySQL事务函数CreateRelation，若存在相同记录，则返回RPC Response提示用户不能重复关注
   - 若RelationActionType为取消关注，则调用事务函数DelRelationByUserId，若数据库不存在对应Relation记录则返回一个RPC Response提示用户无法取消关注一个未关注的用户
   - 若RelationActionType类型错误，则返回错误RPC Response
- 将上述操作组成RelationCache对象，放入RabbitMQ中让Redis消费

注释：由于使用Redis实现了24小时热点用户，因此Redis只需要消费RelationCache时间大于当前时间-24h的消息。同时Redis还会自动为KeyValue队设置24小时过期时间，并使用过期触发事件实时更新热点用户ZSet以及各个用户的关注、被关注信息。一种场景是当用户A关注了用户B，消息进入Redis，热点用户ZSet根据这条记录将对应的FansCount加1，然后不超过24小时用户A取关了用户B，Redis消费了这条消息，并将FansCount减1，这很好。但是，若24小时后，用户A取关了用户B，这条消息由于产生的时间最新，因此会通过消息队列传入Redis，但Redis并不存在相应的关注记录。解决办法是在删除时判断删除记录的CreateAt时间，若CreateAt+24h大于当前时间，则不加入消息队列，反之则说明该记录在Redis还未过期，需要传入消息队列让Redis修改。

1. RelationFollowList

函数运行过程：

- 读取RPC Request发送来的请求体，获取TokenUserId, ToUserId，ActionType
- 判断TokenUserId是否与ToUserId相同，相同则说明用户试图关注自己，返回错误
- 根据userId调用Redis数据库函数GetFollowingIDs用于从Redis中获取userId所属用户关注的其他用户followingIds
- 使用followingIds中的userId调用Redis数据库函数中的GetFollowerCount、GetFollowingCount，获取这些用户的被关注数以及关注数
- 返回用户TokenUserId关注用户的详细信息

1. RelationFollowerList

函数运行过程：

- 读取RPC Request发送来的请求体，获取TokenUserId, ToUserId，ActionType
- 判断TokenUserId是否与ToUserId相同，相同则说明用户试图关注自己，返回错误
- 根据userId调用Redis数据库函数GetFollowerIDs用于从Redis中获取userId所属用户关注的其他用户followingIds
- 使用followerIds中的userId调用Redis数据库函数中的GetFollowerCount、GetFollowingCount，获取这些用户的被关注数以及关注数
- 返回用户TokenUserId关注用户的详细信息

1. RelationFriendList

函数运行过程：

- 读取RPC Request发送来的请求体，获取TokenUserId, ToUserId，ActionType
- 判断TokenUserId是否与ToUserId相同，相同则说明用户试图关注自己，返回错误
- 根据userId调用Redis数据库函数GetFriends用于从Redis中获取userId的好友信息
- 使用friendsIds中的userId调用Redis数据库函数中的GetFollowerCount、GetFollowingCount，获取这些用户的被关注数以及关注数
- 返回用户TokenUserId好友的详细信息

## Video

## Favorite

### cmd

1. FavoriteAction
- 读取RPC Request发送来的请求体，获取TokenUserId, ToUserId，ActionType
- 根据FavoriteActionType作出如下操作：
   - 若FavoriteActionType为LIKE，则使用MySQL事务函数CreateVideoFavorite，若存在相同记录，则返回RPC Response提示用户不能重复点赞
   - 若ActionType为CANCEL_LIKE，则调用事务函数DelFavoriteByUserVideoID，若数据库不存在对应Relation记录则返回一个RPC Response提示用户无法取消点赞一个未点赞的视频
   - 若FavoriteActionType为DISLIKE，则先查看MySQL事务函数CreateVideoFavorite，若存在相同记录，则返回RPC Response提示用户不能重复点赞
   - 若ActionType为CANCEL_DISLIKE，则调用事务函数DelFavoriteByUserVideoID，若数据库不存在对应Relation记录则返回一个RPC Response提示用户无法取消关注一个未关注的用户
- 将上述操作组成RelationCache对象，放入RabbitMQ中让Redis消费

# 测试报告

机器：Macbook M2

CPU核心数：8

内存大小：16G

固态硬盘大小：256G

## ApiRouter

测试工具：wrk

### UserLogin

去掉微服务请求后，测试结果如下：

```bash
wrk -t16 -c800 -d30s -s post_login.lua http://127.0.0.1:8089/wgxdouyin/user/login/16 threads and 800 connections
Thread Stats   Avg      Stdev     Max   +/- Stdev
Latency    67.52ms   41.29ms 431.42ms   68.65%
Req/Sec   773.11    820.98     8.15k    96.42%
371239 requests in 30.09s, 69.39MB read
Socket errors: connect 0, read 1917, write 0, timeout 0
Requests/sec:  12337.55
Transfer/sec:      2.31MB

wrk -t8 -c8 -d30s -s post_login.lua http://127.0.0.1:8089/wgxdouyin/user/login/
8 threads and 8 connections
Thread Stats   Avg      Stdev     Max   +/- Stdev
Latency     6.82ms   21.28ms 223.57ms   91.67%
Req/Sec     1.57k     1.69k   14.39k    96.62%
372646 requests in 30.05s, 69.66MB read
Requests/sec:  12401.52
Transfer/sec:      2.32MB

wrk -t1 -c1 -d30s -s post_login.lua http://127.0.0.1:8089/wgxdouyin/user/login/
1 threads and 1 connections
Thread Stats   Avg      Stdev     Max   +/- Stdev
Latency     4.77ms   17.00ms 124.87ms   92.66%
Req/Sec    12.58k     8.08k   35.31k    78.33%
375814 requests in 30.04s, 70.25MB read
Requests/sec:  12511.33
Transfer/sec:      2.34MB 
```

可见在该机器中，最大可承载的请求数量约在12500 Requests/sec，这也是UserLogin接口的上限，后续可根据此请求数作为Baseline，逐一添加微服务、数据库、消息队列各个部件一一对比接口的请求数量

下面时加上微服务、数据库操作后的测试结果：

```lua
wrk.method = "POST"
wrk.headers["Content-Type"] = "application/x-www-form-urlencoded"
users = {
    { username = "test1", password = "12345" },
    { username = "test2", password = "12345" },
    { username = "test3", password = "12345" },
    { username = "test4", password = "12345" },
    { username = "test5", password = "12345" },
}

function request()
    local user = users[math.random(#users)]
    local body = "username=" .. user.username .. "&password=" .. user.password
    return wrk.format(nil, nil, nil, body)
end
```

```bash
wrk -t8 -c8 -d30s -s post_login.lua http://127.0.0.1:8089/wgxdouyin/user/login/
  8 threads and 8 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency    40.32ms    8.59ms 110.60ms   73.50%
    Req/Sec    24.62      5.67    90.00     97.22%
  5956 requests in 30.10s, 2.22MB read
Requests/sec:    197.86
Transfer/sec:     75.55KB

wrk -t16 -c100 -d30s -s post_login.lua http://127.0.0.1:8089/wgxdouyin/user/login/
  16 threads and 100 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   505.08ms  303.90ms   1.99s    73.78%
    Req/Sec    13.19      8.02    50.00     76.04%
  5819 requests in 30.10s, 2.17MB read
  Socket errors: connect 0, read 0, write 0, timeout 11
Requests/sec:    193.31
Transfer/sec:     73.81KB

wrk -t8 -c50 -d30s -s post_login.lua http://127.0.0.1:8089/wgxdouyin/user/login/ 
  8 threads and 50 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   242.89ms  128.13ms   1.25s    74.37%
    Req/Sec    25.50     12.28    70.00     58.62%
  6052 requests in 30.10s, 2.26MB read
Requests/sec:    201.07
Transfer/sec:     76.78KB
```

可见整个系统的对于登陆接口，最多允许一秒内同时登陆200次。同时，根据后台报错，发现如下问题，即一些查询超过200ms，被标记为慢查询。

```bash
2025/04/07 15:30:53 /Users/wangguixin/wgx/Project/wgxDouYinPage/dal/db/user.go:52 SLOW SQL >= 200ms
[289.575ms] [rows:1] SELECT id, user_name, password FROM `users` WHERE user_name = 'test2' AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/04/07 15:30:53 /Users/wangguixin/wgx/Project/wgxDouYinPage/dal/db/user.go:52 SLOW SQL >= 200ms
[294.447ms] [rows:1] SELECT id, user_name, password FROM `users` WHERE user_name = 'test5' AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1
```
Login rpc server端主要是使用了三个函数分别是`GetUserByName` 、`PasswordCompare` 、`CreateToken` 。下面是这三个函数单次Query的执行时间：

数据库查询耗时：2.896834ms
比较密钥耗时：27.551584ms
签发token耗时：210.625µs

根据gorm查询转译后的原SQL语句，使用该SQL语句使用sysbench发现其QPS高达5万，因此`GetUserByName` 不是影响QPS的主要原因。签发token也不需要多少时间。主要性能瓶颈则是比较密钥函数，原因则是我使用了Argon2这种慢哈希函数来根据原密码生成密钥，该函数可以指定CPU核心数、Memory占用来控制时间和空间复杂度，因此当请求数量过多时，该函数大量占用CPU时间，进而导致CPU只能分配少量的计算资源给数据库引擎，最后导致了数据库Query变慢（猜测）。