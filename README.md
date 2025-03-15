# wgxDouYin
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

## 安全问题

### KeyManager

实现了一个KeyManager，KeyManager实现了QueryUpdater接口，用于监听服务公钥的变化。同时在后续的扩展工作中，也能作为各个微服务的KeyManager用于管理微服务之间的通信密钥。

### 权限管理

当一个用户登录时，根据其用户信息，使用**用户服务**的私钥根据用户名以及登录日期进行签名，签名形成一个**JWT**。这个JWT在用于在用户登陆后进行其它服务请求时的身份验证。

项目采用公私钥系统来进行鉴权操作。因为所有服务都使用一个密钥进行鉴权会产生如下问题：

1.服务之间无法验证是否合法。例如某个服务A需要在进行服务B之后才能够完成操作，但是若使用一个密钥进行鉴权，则服务A要一直等到RPC请求被解析后才能够确认该服务不合法。但是如果使用公私钥进行鉴权，同时定义一套规则链给API Router，API Router按照规则链查询对应的公钥进行解析请求，如果解析失败（即服务B的公钥无法解析该请求），说明该请求不合法。

2.服务存在伪造的可能。因为所有的token都通过一个密钥进行签名，因此服务B可能知道了服务A的请求规则，于是它伪造了一个请求并进行签名，服务A接收后确认该服务合法执行服务B伪造的操作。

因此在进行服务调用前，必须使用查看该服务的依赖服务，再查询依赖服务的公钥，用此公钥验证请求中的jwt是否合法。

## 类说明

### QueryUpdater （interface）

该接口包含一个Update函数，用于实时更新Etcd的键值对变化给实现该接口的类。

## 业务逻辑

### relation

原本项目的逻辑如下：

1.针对用户的关注、取关操作，relation服务首先查询是否

relation服务主要包含用户的关系操作，主要是关注、取关两个操作，对应关注数量以及粉丝数量。主要的逻辑如下：

1.用户不能够关注自己

2.某些用户可能为热点用户，即涨粉速度很快，

### User

**密码存储**

密码采用Argon+Salt加密后存储，避免了彩虹表攻击

### DAL(Data Access Layer)

所有的数据操作使用DAL中定义的函数对数据库实例进行操作。拿MySQL来说，各个微服务都会启动一个连接，为了避免他们的操作出现冲突，可以对一些写操作以事务的方式进行。

## Package

### RabbitMQ

RabbitMQ包用于创建RabbitMQ连接，然后通过连接中的交换机进行信息的交换。使用Qos可以用于限流。

同时RabbitMQ用户各个微服务向Redis客户端推送信息，当Redis客户端消费时信息时则采用