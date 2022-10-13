# go-docker-k3s-example

[English](./README.md)

这个项目用来演示将一个go服务快速部署到k3s， 包含了打包服务成docker镜像，推送镜像到本地的docker镜像仓库，启动一个单节点的k3s，然后部署到k3s。

<br/>

全部代码都是用来学习的，可能不正确，请理性参考，感谢

## 有一个go服务

例如 ```k3sdemo```

## 服务器系统选择

ubuntu 20.04

## 服务器用户配置(可选)

```bash
apt-get update
```

禁用密码登陆，启用pubkey登陆，需要编辑```/etc/ssh/sshd_config```文件
```
AuthorizedKeysFile .ssh/authorized_keys
PasswordAuthentication no
PubkeyAuthentication yes
```

添加pubkey到```.ssh/authorized_keys```文件，然后执行
```bash
service sshd restart
```

创建一个新用户，之后全部的操作都在这个用户下面进行
```bash
adduser k3suser
```

设置新用户的权限，添加下面的代码到```/etc/sudoers```
```
k3suser ALL=(ALL:ALL) ALL
```

切换到新用户，添加pubkey到新用户的```.ssh/authorized_keys```文件
```bash
su k3suser
```

完成之后，注销服务器登录然后登陆到```k3suser@YOUR_IP```

<br/>

## 服务器安装基础库

下面的命令可能要在前面添加```sudo```

```bash
apt-get update
```

安装 docker
```bash
curl https://releases.rancher.com/install-docker/20.10.sh | sh
```

如果服务器是树莓派，可能需要解决获取机器底层信息的问题，可以回到这里安装下面
```bash
apt install haveged 
systemctl enable haveged
```

安装 postgres
```bash
# 安装到 docker(推荐)
docker run --name postgres -v /var/lib/postgresql/data:/var/lib/postgresql/data -e POSTGRES_PASSWORD=postgres -p 5432:5432 -d --restart=always postgres:latest

# 或者安装到机器本地
apt install postgresql postgresql-contrib
sudo -u postgres psql
ALTER USER postgres WITH PASSWORD 'postgres';
```

推荐使用更复杂的密码

<br/>

安装 docker registry 来存 docker镜像，作为镜像仓库
```bash
docker run -d -p 5000:5000 --restart=always --name registry -v /var/lib/registry/data:/var/lib/registry registry:2
```

获取 postgres 的 docker IPAddress，然后替换[production.yaml](config/production.yaml)文件的host
```bash
# get postgres CONTAINER_ID
docker ps | grep postgres
docker inspect CONTAINER_ID
```

创建 database
```bash
# get postgres CONTAINER_ID
docker ps | grep postgres
docker exec -it CONTAINER_ID bash
psql -U postgres

# k3s 默认使用 sqlite 作为数据存储, 这里使用postgres作为数据存储
CREATE DATABASE k3s;
GRANT ALL PRIVILEGES ON DATABASE k3s TO postgres;

# go 服务的数据库
CREATE DATABASE k3sdemo;
GRANT ALL PRIVILEGES ON DATABASE k3sdemo TO postgres;

# 如果忘了密码可以重置密码
ALTER USER postgres WITH PASSWORD 'postgres';
```

安装 k3s
```bash
curl -sfL https://rancher-mirror.oss-cn-beijing.aliyuncs.com/k3s/k3s-install.sh | INSTALL_K3S_MIRROR=cn INSTALL_K3S_SKIP_START=true sh -s - server --docker --datastore-endpoint="postgres://postgres:postgres@localhost:5432/k3s?sslmode=disable" --etcd-disable-snapshots
```
如果这里有pg ssl连接的问题，可以添加```?sslmode=disable```到```/etc/systemd/system/multi-user.target.wants/k3s.service```文件的```--datastore-endpoint```中，然后重启k3s

<br/>

配置镜像的私库地址，创建```/etc/rancher/k3s/registries.yaml```文件并添加下面代码：
```
mirrors: 
  "localhost:5000": 
    endpoint: 
      - "http://127.0.0.1:5000"
```
<br/>

启动 k3s(或者重启)
```bash
systemctl daemon-reload
service k3s restart

# 检查 k3s 状态
systemctl status k3s.service
# 检查 k3s 日志
journalctl -u k3s.service -f
```

<br/>

如果想停止并卸载k3s
```
/usr/local/bin/k3s-killall.sh && /usr/local/bin/k3s-uninstall.sh
```

<br/>

## 部署go服务到k3s

在服务器上进入到go服务的文件夹
```bash
cd go-docker-k3s-example/
```

创建 k3s namespace(名称: k3sdemo)
```bash
make k3s-create-namespace
```

打包go服务成镜像并推送到本地镜像仓库
```bash
make docker-build-image-prod && make docker-push-image-prod
```
当看到```{"repositories":["k3sdemo"]}```时就说明镜像打包成功并且推送成功

<br/>

部署到 k3s 
```bash
make k3s-apply-yaml
```

检查go服务的部署结果
```bash
kubectl -n k3sdemo get pods
```

将会看到:
| NAME | READY | STATUS | RESTARTS | AGE |
| ---- | ---- | ---- | ---- | ---- |
| k3sdemo-deployment-xxx | 1/1 | Running | 0 | 13s |
| k3sdemo-deployment-xxx | 1/1 | Running | 0 | 12s |

其他的一些检查命令
```bash
# pods
kubectl -n k3sdemo describe deployments k3sdemo-deployment
kubectl -n k3sdemo logs -f POD_NAME
kubectl -n k3sdemo describe pod/POD_NAME

# services
kubectl -n k3sdemo describe services k3sdemo-service

# ingress
kubectl -n k3sdemo describe ingress k3sdemo-ingress
kubectl -n k3sdemo k3sdemo-ingress get svc 
```

如果需要重新部署
```bash
make k3s-replace-yaml && make k3s-restart-deployment
```

如果想删除pod
```bash
kubectl -n k3sdemo delete po POD_NAME
```

检查pod运行情况
```bash
kubectl describe node
```
将会看到:
| Namespace | Name | CPU Requests | CPU Limits | Memory Requests | Memory Limits | Age |
| ---- | ---- | ---- | ---- | ---- | ---- | ---- |
| k3sdemo | k3sdemo-deployment-xxx | 0 (0%) | 0 (0%) | 0 (0%) | 0 (0%) | 10m |

<br/>

现在可以访问```curl 127.0.0.1``` 或 ```curl YOUR_IP``` 或 ```curl YOUR_DOMAIN```来确认go服务是否可以正常使用，如果没有意外的话将会看到```welcome```