# Heztner cloud with Kubernetes

Tested on `VM CX21 - Ubuntu 20.04 - 2 vCPU - 4 GB RAM - 40 GB disk`

```
sudo apt-get update
sudo apt-get upgrade
```

## Floating IP (NOT USED IN CURRENT SETUP AND NOT REQUIRED)

1. Create Floating IP via Heztner web interface and copy that value (don't run the temporary command as suggested)
2. Set the IP following [this tutorial](https://docs.hetzner.com/cloud/floating-ips/persistent-configuration/) or

    ```
    touch /etc/netplan/60-floating-ip.yaml
    nano /etc/netplan/60-floating-ip.yaml
    ```

    Add this content:
    ```
    network:
       version: 2
       renderer: networkd
       ethernets:
         eth0:
           addresses:
           - your.float.ing.ip/32          <------ add your IPv4 floating IP here
    ```

    Finally, apply changes:
    `sudo netplan apply`


## Letting iptables see bridged traffic

Make sure that the br_netfilter module is loaded. This can be done by running lsmod | grep br_netfilter. To load it explicitly call sudo modprobe br_netfilter.
As a requirement for your Linux Node's iptables to correctly see bridged traffic, you should ensure net.bridge.bridge-nf-call-iptables is set to 1 in your sysctl config, e.g.

```
sudo modprobe br_netfilter

cat <<EOF | sudo tee /etc/modules-load.d/k8s.conf
br_netfilter
EOF

cat <<EOF | sudo tee /etc/sysctl.d/k8s.conf
net.bridge.bridge-nf-call-ip6tables = 1
net.bridge.bridge-nf-call-iptables = 1
EOF

sudo sysctl --system

```

Check port via telnet

```
sudo apt-get update
sudo apt-get upgrade
sudo apt install telnet

telnet 127.0.0.1 6443
# must return "telnet: Unable to connect to remote host: Connection refused"

```


## Installing Docker Engine

```
sudo apt install ca-certificates curl gnupg lsb-release

curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg

echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

sudo apt update && sudo apt install docker-ce docker-ce-cli containerd.io -y

sudo systemctl start docker && sudo systemctl enable docker 

sudo systemctl status docker
```


## Configuring Cgroup Driver

You can adjust the Docker configuration using the following command on each node:

```
cat <<EOF | sudo tee /etc/docker/daemon.json
{
  "exec-opts": ["native.cgroupdriver=systemd"],
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "100m"
  },
  "storage-driver": "overlay2"
}
EOF
```

or more details, see configuring a cgroup driver.
Once you’ve adjusted the configuration on each node, restart the Docker service and its corresponding daemon.

`sudo systemctl daemon-reload && sudo systemctl restart docker`

With Docker up and running, the next step is to install kubeadm, kubelet, and kubectl on each node.


## Installing kubeadm, kubelet, and kubectl

```
sudo apt-get update
sudo apt-get install -y apt-transport-https ca-certificates curl

sudo curl -fsSLo /usr/share/keyrings/kubernetes-archive-keyring.gpg https://packages.cloud.google.com/apt/doc/apt-key.gpg

echo "deb [signed-by=/usr/share/keyrings/kubernetes-archive-keyring.gpg] https://apt.kubernetes.io/ kubernetes-xenial main" | sudo tee /etc/apt/sources.list.d/kubernetes.list

sudo apt-get update
sudo apt-get install -y kubelet kubeadm kubectl
sudo apt-mark hold kubelet kubeadm kubectl
```

The last line with the apt-mark hold command is optional, but highly recommended. This will prevent these packages from being updated until you unhold them.


## Prepare for hcloud-cloud-controller-manager

The cloud controller manager adds its labels when a node is added to the cluster.
For Kubernetes versions prior to 1.23, this means we have to add the --cloud-provider=external flag to the kubelet before initializing the cluster master with kubeadm init

`nano /etc/systemd/system/kubelet.service.d/20-hcloud.conf` with this content:

```
Service]
Environment="KUBELET_EXTRA_ARGS=--cloud-provider=external"
```

```
sudo kubeadm init --pod-network-cidr=10.244.0.0/16

mkdir -p $HOME/.kube
sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
sudo chown $(id -u):$(id -g) $HOME/.kube/config
```


## Deploy CNI plugin

hcloud suggests flannel, so follow this:

```
kubectl apply -f https://raw.githubusercontent.com/coreos/flannel/v0.17.0/Documentation/kube-flannel.yml
```

### Patch Flannel deployment

To tolerate 'uninitialized' taint:

```
kubectl -n kube-system patch ds kube-flannel-ds --type json -p '[{"op":"add","path":"/spec/template/spec/tolerations/-","value":{"key":"node.cloudprovider.kubernetes.io/uninitialized","value":"true","effect":"NoSchedule"}}]'
```

### Create secret with Hetzner Cloud API token

```
kubectl -n kube-system create secret generic hcloud --from-literal=token=<hcloud API token>            <------ add Hetzner API token here from the UI
```

### Install hcloud-cloud-controller-manager

```
kubectl apply -f  https://github.com/hetznercloud/hcloud-cloud-controller-manager/releases/latest/download/ccm.yaml
```


## Deploy your application

Deploy everything. If they are Pending with error "default-scheduler  0/1 nodes are available: 1 node(s) had taint {node-role.kubernetes.io/master: }, that the pod didn't tolerate" run this:

```
kubectl taint nodes --all node-role.kubernetes.io/master-
```

From the UI of Hetzner, you'll find the Load Balancer public IPs.