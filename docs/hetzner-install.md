# Heztner cloud with Kubernetes

Based on https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/install-kubeadm/

## Environment

- Ubuntu 22.04 LTS
- Kubernetes 1.24
- Flannel 0.17.0
- MetalLB 0.12.1
- containerd 1.6.4
- runc 1.1.2
- cni-plugins 1.1.1


## Server creation

From Hetzner Cloud UI create a server like this:

REGION: Falkenstein
OS type: Ubuntu 22.04
Type: Standard - CPX11 - 2 vCPU - 4 GB RAM - 40 GB disk
Volume: none
Network: none
Firewalls: none
Additional features: none
SSH Keys: Create a SSH key-pair and paste the public one or enable an existing one
Name: what you like


## Create Floating IPs

**Floating IPs are required to have static public IPs to expose public Kubernetes services**

From Hetzner Cloud UI create 2 IPs:

- name: ac-gui-floating-ip
  location: Falkenstein
  protocol: IPV4
- name: ac-mosquitto-floating-ip
  location: Falkenstein
  protocol: IPV4

From "Assigned to" column you need to choose the server created above.


## SSH access

Login to your server with

```bash
ssh -i ~/.ssh/<private_key_file> root@<HETZNER_SERVER_IP>
```


## Update Ubuntu

```bash
sudo apt-get update
sudo apt-get upgrade
```


## Disable Linux swap for Kubernetes

Check with `htop` if swap is disabled. If not, run:

```bash
# disable swap right now
sudo swapoff -a
# disable swap also when you'll reboot
cp /etc/fstab /etc/fstab.backup
sudo sed -i '/ swap / s/^\(.*\)$/#\1/g' /etc/fstab
```


## Letting iptables see bridged traffic

Make sure that the br_netfilter module is loaded. This can be done by running lsmod | grep br_netfilter. To load it explicitly call sudo modprobe br_netfilter.
As a requirement for your Linux Node's iptables to correctly see bridged traffic, you should ensure net.bridge.bridge-nf-call-iptables is set to 1 in your sysctl config.

```bash
cat <<EOF | sudo tee /etc/modules-load.d/k8s.conf
overlay
br_netfilter
EOF

sudo modprobe overlay
sudo modprobe br_netfilter

# sysctl params required by setup, params persist across reboots
cat <<EOF | sudo tee /etc/sysctl.d/k8s.conf
net.bridge.bridge-nf-call-iptables  = 1
net.bridge.bridge-nf-call-ip6tables = 1
net.ipv4.ip_forward                 = 1
EOF

# Apply sysctl params without reboot
sudo sysctl --system
```


## Check open ports (optional)

Check port via telnet

```bash
sudo apt-get update
sudo apt-get upgrade
sudo apt install telnet

telnet 127.0.0.1 6443
# must return "telnet: Unable to connect to remote host: Connection refused"

```


## Installing containerd (manually)

Taken from https://github.com/containerd/containerd/blob/main/docs/getting-started.md

```bash

wget https://github.com/containerd/containerd/releases/download/v1.6.4/containerd-1.6.4-linux-amd64.tar.gz
tar Cxzvf /usr/local containerd-1.6.4-linux-amd64.tar.gz

wget https://raw.githubusercontent.com/containerd/containerd/main/containerd.service
cp containerd.service /etc/systemd/system/
chmod 664 /etc/systemd/system/containerd.service

wget https://github.com/opencontainers/runc/releases/download/v1.1.2/runc.amd64
install -m 755 runc.amd64 /usr/local/sbin/runc

wget https://github.com/containernetworking/plugins/releases/download/v1.1.1/cni-plugins-linux-amd64-v1.1.1.tgz
mkdir -p /opt/cni/bin
tar Cxzvf /opt/cni/bin cni-plugins-linux-amd64-v1.1.1.tgz

# generate default config.toml for containerd
mkdir -p /etc/containerd
containerd config default > /etc/containerd/config.toml

# start containerd
sudo systemctl enable containerd
sudo systemctl start containerd
sudo systemctl status containerd
```


## Configure systemd cgroup driver

To use the systemd cgroup driver in `/etc/containerd/config.toml` with runc, set

```
[plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runc]
    ...
    [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runc.options]
        SystemdCgroup = true
```

If you apply this change, make sure to restart containerd:
```bash
sudo systemctl restart containerd
sudo systemctl status containerd
```

With containerd up and running, the next step is to install kubeadm, kubelet, and kubectl on each node.


## Installing kubeadm, kubelet, and kubectl

```bash
sudo apt-get update
sudo apt-get install -y apt-transport-https ca-certificates curl

sudo curl -fsSLo /usr/share/keyrings/kubernetes-archive-keyring.gpg https://packages.cloud.google.com/apt/doc/apt-key.gpg

echo "deb [signed-by=/usr/share/keyrings/kubernetes-archive-keyring.gpg] https://apt.kubernetes.io/ kubernetes-xenial main" | sudo tee /etc/apt/sources.list.d/kubernetes.list

sudo apt-get update
sudo apt-get install -y kubelet kubeadm kubectl
sudo apt-mark hold kubelet kubeadm kubectl
```

The last line with the apt-mark hold command is optional, but highly recommended. This will prevent these packages from being updated until you unhold them.


## Initialize kubeadm:

Taken from https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/create-cluster-kubeadm/

```bash
sudo kubeadm init --pod-network-cidr=10.244.0.0/16

mkdir -p $HOME/.kube
sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
sudo chown $(id -u):$(id -g) $HOME/.kube/config
```


## Deploy CNI plugin

MetalLB reports some incompatibilities with different CNI plugins, so I chose flannel, because it seems supported without issues.


```bash
kubectl apply -f https://raw.githubusercontent.com/coreos/flannel/v0.17.0/Documentation/kube-flannel.yml
```

Patch Flannel deployment to tolerate 'uninitialized' taint:

```bash
kubectl -n kube-system patch ds kube-flannel-ds --type json -p '[{"op":"add","path":"/spec/template/spec/tolerations/-","value":{"key":"node.cloudprovider.kubernetes.io/uninitialized","value":"true","effect":"NoSchedule"}}]'
```


## Fix taints

To fix error `0/1 nodes are available: 1 node(s) had taint {node-role.kubernetes.io/master: }, that the pod didn't tolerate" `when you'll deploy your app, you have to run this:

```bash
kubectl taint nodes --all node-role.kubernetes.io/master-
```


## Install MetalLB

```bash
curl -fsSL -o get_helm.sh https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3
chmod 700 get_helm.sh
./get_helm.sh

kubectl apply -f https://raw.githubusercontent.com/metallb/metallb/v0.12.1/manifests/namespace.yaml
kubectl apply -f https://raw.githubusercontent.com/metallb/metallb/v0.12.1/manifests/metallb.yaml
```

And fix taint error for MetalLB if you see error: `0/1 nodes are available: 1 node(s) had untolerated taint {node-role.kubernetes.io/control-plane: }. preemption: 0/1 nodes are available: 1 Preemption is not helpful for scheduling.`.

```bash
kubectl taint nodes --all node-role.kubernetes.io/control-plane-
```


## Prepare Persistent Volumes

Two PVs are required to store nginx.conf and SSL certificates.
Let's Encrypt certificates issued via Certbot are limited. You cannot register your domain multiple times, otherwise you'll be banned for many days.
So, you need to store certificates and re-use they.

Access to Hetzner server via SSH and prepare these two folders:

```bash
mkdir /root/lets-encrypt-certs
mkdir /root/lets-encrypt-certs-mqtt
mkdir /root/nginx-conf
```


## Update DNS records

Update DNS records of your domains:

```
A @ <ac-gui-floating-ip_IP_ADDRESS>
A wwww <ac-gui-floating-ip_IP_ADDRESS>
```

```
A @ <ac-mosquitto-floating-ip_IP_ADDRESS>
A wwww <ac-mosquitto-floating-ip_IP_ADDRESS>
```


## Deploy application

## Development without SLL and domain names

1. Define personal config in a private repository

Create a new private repository to store your secrets and private configurations, for instance `air-conditioner-server-config`

2. Create a custom values file in `air-conditioner-server-config/custom-values.yaml` with a specific configuration like:

```yaml
domains:
  # overwrite default http domain to don't use domain name
  # in this way you'll be able to reach this web app and rest services via `gui.publicIp`
  http: "<ac-gui-floating-ip_IP_ADDRESS>"
  mqtt: "localhost"

mosquitto:
  publicIp: "<ac-mosquitto-floating-ip_IP_ADDRESS>"

apiServer:
  oauthClientId: "<GITHUB_OAUTH_CLIENT>"
  oauthSecret: "<GITHUB_OAUTH_SECRET>"
  singleUserLoginEmail: "<GITHUB_ACCOUNT_EMAIL_TO_LOGIN>"

gui:
  publicIp: "<ac-gui-floating-ip_IP_ADDRESS>"

mongodbUrl: "mongodb+srv://<MONGODB_ATLAS_USERNAME>:<MONGODB_ATLAS_PASSWORD>@cluster0.4wies.mongodb.net"
```

3. (optional step) If you want to see all manifests processed by Helm without deploying them, you can run:

```bash
cd helm/ac
helm template -f values.yaml -f ../../air-conditioner-server-config/custom-values.yaml . > output-manifests-no-ssl.yaml
```

4. Deploy with Helm

```bash
cd helm/ac
helm install -f values.yaml -f ../../air-conditioner-server-config/custom-values.yaml ac .
```

5. Check kubernetes services! You should see 2 LoadBalancers with the right Floating IPs assigned.
   After some time, you'll be able to navigate to the website via HTTP and to the Mosquitto server via MQTT connection.



## Production with SSL and domain names

1. Define personal config in a private repository

Create a new private repository to store your secrets and private configurations, for instance `air-conditioner-server-config`

2. Create a custom values file in `air-conditioner-server-config/custom-values.yaml` with a specific configuration like:

```yaml
domains:
  http: "YOUR_DOMAIN"
  mqtt: "YOUR_MQTT_DOMAIN"

mosquitto:
  publicIp: "<ac-mosquitto-floating-ip_IP_ADDRESS>"
  ssl:
    enable: true
    certbot:
      email: "<YOUR_CERTIFICATE_EMAIL>"

apiServer:
  oauthClientId: "<GITHUB_OAUTH_CLIENT>"
  oauthSecret: "<GITHUB_OAUTH_SECRET>"
  singleUserLoginEmail: "<GITHUB_ACCOUNT_EMAIL_TO_LOGIN>"

gui:
  publicIp: "<ac-gui-floating-ip_IP_ADDRESS>"
  ssl:
    enable: true
    certbot:
      email: "<YOUR_CERTIFICATE_EMAIL>"

mongodbUrl: "mongodb+srv://<MONGODB_ATLAS_USERNAME>:<MONGODB_ATLAS_PASSWORD>@cluster0.4wies.mongodb.net"
```

3. (optional step) If you want to see all manifests processed by Helm without deploying them, you can run:

```bash
cd helm/ac
helm template -f values.yaml -f ../../../air-conditioner-server-config/custom-values.yaml . > output-manifests.yaml
```

4. Deploy with Helm

```bash
cd helm/ac
helm install -f values.yaml -f ../../../air-conditioner-server-config/custom-values.yaml  ac .
```

5. Check kubernetes services! You should see 2 LoadBalancers with the right Floating IPs assigned.
   After some time, you'll be able to navigate to the website via HTTPS and to the Mosquitto server via MQTTS connection.
   ESP32 device should already be working using secure connections.
   If you have problems with certificates, you should check if certbot is started getting SSL certificates from Let's Encrypt.
   Certbot runs on these 2 pods:
   - gui
   - mosquitto