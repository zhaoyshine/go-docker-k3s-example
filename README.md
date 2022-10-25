# go-docker-k3s-example

[简体中文](./README_zh.md)

This project is used to demonstrate the quick deploy of go service to k3s, including building a docker image, pushing it to a local docker image repository, starting k3s on a single node, and then deploying to k3s.

<br/>

All codes are for learning only and may be incorrect, please refer to them rationally, thank you

## Go service

such as k3sdemo

## Server selection

ubuntu 20.04

## Server user configuration(optional)

```bash
apt-get update
```

disable password login, enable pubkey login, edit in the ```/etc/ssh/sshd_config``` file
```
AuthorizedKeysFile .ssh/authorized_keys
PasswordAuthentication no
PubkeyAuthentication yes
```

add your pubkey to ```.ssh/authorized_keys``` file, and then exec
```bash
service sshd restart
```

create a new user, all operations are performed under this user
```bash
adduser k3suser
```

set new user permissions, add the following to ```/etc/sudoers```
```
k3suser ALL=(ALL:ALL) ALL
```

add your pubkey to new user ```.ssh/authorized_keys``` file, switch to new user
```bash
su k3suser
```

and then, log out of the server and log in to k3suser@YOUR_IP

<br/>

## Server installation dependencies

the following commands maybe need to add ```sudo```

```bash
apt-get update
```

Install docker
```bash
curl https://releases.rancher.com/install-docker/20.10.sh | sh
```

if your server is a Raspberry Pi, you may need to solve some low-level machine configuration access problems, you can go back here and install:
```bash
apt install haveged 
systemctl enable haveged
```

Install postgres
```bash
# install to docker(recommend)
docker run --name postgres -v /var/lib/postgresql/data:/var/lib/postgresql/data -e POSTGRES_PASSWORD=postgres -p 5432:5432 -d --restart=always postgres:latest

# or install to local
apt install postgresql postgresql-contrib
sudo -u postgres psql
ALTER USER postgres WITH PASSWORD 'postgres';
```
more complex passwords are recommended

<br/>

Install docker registry to save docker images as a image repository
```bash
docker run -d -p 5000:5000 --restart=always --name registry -v /var/lib/registry/data:/var/lib/registry registry:2
```

get the IPAddress and replace the host address in [production.yaml](config/production.yaml)
```bash
# get postgres CONTAINER_ID
docker ps | grep postgres
docker inspect CONTAINER_ID
```

create database
```bash
# get postgres CONTAINER_ID
docker ps | grep postgres
docker exec -it CONTAINER_ID bash
psql -U postgres

# k3s uses sqlite by default, here postgres is used as data storage
CREATE DATABASE k3s;
GRANT ALL PRIVILEGES ON DATABASE k3s TO postgres;

# go service database
CREATE DATABASE k3sdemo;
GRANT ALL PRIVILEGES ON DATABASE k3sdemo TO postgres;

# reset password if you forget
ALTER USER postgres WITH PASSWORD 'postgres';
```

Install k3s
```bash
curl -sfL https://rancher-mirror.oss-cn-beijing.aliyuncs.com/k3s/k3s-install.sh | INSTALL_K3S_MIRROR=cn INSTALL_K3S_SKIP_START=true sh -s - server --docker --datastore-endpoint="postgres://postgres:postgres@localhost:5432/k3s?sslmode=disable" --etcd-disable-snapshots
```

if there is a issue with pg ssl connection when k3s start, you can add ```?sslmode=disable``` to ```/etc/systemd/system/multi-user.target.wants/k3s.service``` file ```--datastore-endpoint``` options and restart

<br/>

configure the address of the docker images private repository, create ```/etc/rancher/k3s/registries.yaml``` file with following:
```
mirrors: 
  "localhost:5000": 
    endpoint: 
      - "http://127.0.0.1:5000"
```
<br/>

start k3s(restart)
```bash
systemctl daemon-reload
service k3s restart

# check k3s status
systemctl status k3s.service
# check k3s logs
journalctl -u k3s.service -f
```

<br/>

killall k3s and uninstall k3s if you want
```
/usr/local/bin/k3s-killall.sh && /usr/local/bin/k3s-uninstall.sh
```

<br/>

## Https

Install cert manager
```bash
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.10.0/cert-manager.yaml
```

replace YOUR_EMAIL and then run following
```bash
kubectl apply -f deploy/letsencrypt-prod.yaml
```

```bash
kubectl apply -f deploy/redirect-https.yaml
```

update [ingress.yaml](./deploy/ingress.yaml)
```
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: k3sdemo-ingress
  namespace: k3sdemo
  annotations:
    kubernetes.io/ingress.class: traefik
    cert-manager.io/cluster-issuer: letsencrypt-prod
    traefik.ingress.kubernetes.io/router.middlewares: default-redirect-https@kubernetescrd
spec:
  rules:
  - host: YOUR_HOST.com
    http: 
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: k3sdemo-service
            port:
              name: api
  tls:
  - secretName: k3sdemo-tls
    hosts:
      - YOUR_HOST.com
```

<br/>

## Deploy the go service to k3s

go to the go service folder on the server
```bash
cd go-docker-k3s-example/
```

create k3s namespace(name: k3sdemo)
```bash
make k3s-create-namespace
```

build docker image and push image to registry
```bash
make docker-build-image-prod && make docker-push-image-prod
```
if you can see ```{"repositories":["k3sdemo"]}```, the image build and push successfully

<br/>

deploy to k3s 
```bash
make k3s-apply-yaml
```

check deploy results
```bash
kubectl -n k3sdemo get pods
```
will get table like:
| NAME | READY | STATUS | RESTARTS | AGE |
| ---- | ---- | ---- | ---- | ---- |
| k3sdemo-deployment-xxx | 1/1 | Running | 0 | 13s |
| k3sdemo-deployment-xxx | 1/1 | Running | 0 | 12s |

some other check commands
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

redeploy pods if you need
```bash
make k3s-replace-yaml && make k3s-restart-deployment
```

delete pods if you want
```bash
kubectl -n k3sdemo delete po POD_NAME
```

node info if you want to see
```bash
kubectl describe node
```
will get table like:
| Namespace | Name | CPU Requests | CPU Limits | Memory Requests | Memory Limits | Age |
| ---- | ---- | ---- | ---- | ---- | ---- | ---- |
| k3sdemo | k3sdemo-deployment-xxx | 0 (0%) | 0 (0%) | 0 (0%) | 0 (0%) | 10m |

<br/>

now you can test ```curl 127.0.0.1``` or ```curl YOUR_IP``` or ```curl YOUR_DOMAIN``` to confirm whether the go service can be used normally, you can see: ```welcome```.