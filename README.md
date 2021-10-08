
# k8s-demo
## TODO:
- 搭建 k8s 集群
- 部署一个 pod 输出 k8s 里 cpu 使用率最高的 10 个 pod 
## 安装 docker
```
sudo yum remove docker \
                  docker-client \
                  docker-client-latest \
                  docker-common \
                  docker-latest \
                  docker-latest-logrotate \
                  docker-logrotate \
                  docker-engine

sudo yum install -y yum-utils

sudo yum-config-manager \
    --add-repo \
    https://download.docker.com/linux/centos/docker-ce.repo

sudo yum install docker-ce docker-ce-cli containerd.io

sudo systemctl start docker
```
## 添加 k8s 阿里源
```
tee /etc/yum.repos.d/kubernetes.repo << EOL
[kubernetes]
name=Kubernetes
baseurl=https://mirrors.aliyun.com/kubernetes/yum/repos/kubernetes-el7-x86_64/
enabled=1
gpgcheck=1
repo_gpgcheck=1
gpgkey=https://mirrors.aliyun.com/kubernetes/yum/doc/yum-key.gpg https://mirrors.aliyun.com/kubernetes/yum/doc/rpm-package-key.gpg
EOL
```
## 配置 daemon.json
```
tee /etc/docker/daemon.json <<EOL
{
  "registry-mirrors": ["https://fl791z1h.mirror.aliyuncs.com"],
  "exec-opts": ["native.cgroupdriver=systemd"],
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "100m"
  },
  "storage-driver": "overlay2"
}
EOL

sudo systemctl enable docker
sudo systemctl daemon-reload
sudo systemctl restart docker
```
## 允许 iptables 检查桥接流量
```
cat <<EOF | sudo tee /etc/modules-load.d/k8s.conf
br_netfilter
EOF

cat <<EOF | sudo tee /etc/sysctl.d/k8s.conf
net.bridge.bridge-nf-call-ip6tables = 1
net.bridge.bridge-nf-call-iptables = 1
EOF
sudo sysctl --system
```
## 安装 k8s
将 SELinux 设置为 permissive 模式（相当于将其禁用）
```
sudo setenforce 0
sudo sed -i 's/^SELINUX=enforcing$/SELINUX=permissive/' /etc/selinux/config

sudo yum install -y kubelet kubeadm kubectl --disableexcludes=kubernetes

sudo systemctl enable --now kubelet
```
## 初始化 k8s 集群
```
kubeadm init --apiserver-advertise-address='172.18.142.162' \
--image-repository registry.aliyuncs.com/google_containers \
--pod-network-cidr=172.0.0.0/24
```
- 记录生成的最后部分内容，此内容需要在其它节点加入Kubernetes集群时执行。
根据提示创建kubectl
```
mkdir -p $HOME/.kube
sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
sudo chown $(id -u):$(id -g) $HOME/.kube/config
```
- 使kubectl可以自动补充
```
source <(kubectl completion bash)
```
### 安装 calico 网络
```
kubectl apply -f https://docs.projectcalico.org/manifests/calico.yaml
```
### 加入新的 node
```
kubeadm join 172.18.142.162:6443 --token ylwz43.7n3nrxsf7w70knkz \
--discovery-token-ca-cert-hash sha256:5c556969ccd1fee913cf0ed4c2218d7c210b55091a5c453ce00150b7b73ab496
```
## 修改 kube-apiserver 运行参数
/etc/kubernetes/manifests/kube-apiserver.yaml
```
--requestheader-allowed-names=aggregator
```
## 部署 metrics-server
- wget https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml
- 修改 components.yaml
```
spec:
   template:
     spec:
       containers:
       - args:
         - --kubelet-insecure-tls   # 不验证客户端证书
         image: registry.aliyuncs.com/google_containers/metrics-server:v0.5.1
```
- 部署 
```
kubectl apply -f components.yaml
```
## 部署 my-app
### 编译镜像
```
cd test 
go build -o ./app .

docker build -t k8s-test:v4 .
docker tag k8s-test:v4 ywdxz/k8s-test:v4
docker push  ywdxz/k8s-test:v4
```
### 部署
```
kubectl apply -f web-service.yaml
kubectl apply -f app.yaml
```
## 浏览器查看结果 
