# Heztner cloud with Kubernetes

## Server creation

From Hetzner Cloud UI create a server like this:

REGION: Nuremberg
OS type: Ubuntu 20.04
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
  location: Nuremberg (or Falkenstein)
  protocol: IPV4
- name: ac-mosquitto-floating-ip
  location: Nuremberg (or Falkenstein)
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


## Check open ports

Check port via telnet

```bash
sudo apt-get update
sudo apt-get upgrade
sudo apt install telnet

telnet 127.0.0.1 6443
# must return "telnet: Unable to connect to remote host: Connection refused"

```


## Installing Docker Engine

```bash
sudo apt install ca-certificates curl gnupg lsb-release

curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg

echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

sudo apt update && sudo apt install docker-ce docker-ce-cli containerd.io -y

sudo systemctl start docker && sudo systemctl enable docker 

sudo systemctl status docker
```


## Configure Cgroup Driver

To do this, you can adjust the Docker configuration using the following command on each node:

```bash
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

or more details, see configuring a cgroup driver in the official Kubernetes doc.
Once you’ve adjusted the configuration on each node, restart the Docker service and its corresponding daemon.

```bash
sudo systemctl daemon-reload && sudo systemctl restart docker
```

With Docker up and running, the next step is to install kubeadm, kubelet, and kubectl on each node.


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

Create a `metallb-config.yaml` file adding your Floating IPs:

```bash
apiVersion: v1
kind: ConfigMap
metadata:
  namespace: metallb-system
  name: config
data:
  config: |
    address-pools:
    - name: default
      protocol: layer2
      addresses:
      - <ac-gui-floating-ip_IP_ADDRESS>/24                 <----------------- add here your floating IP for the HTTP GUI
      - <ac-mosquitto-floating-ip_IP_ADDRESS>/24           <----------------- add here your floating IP for the MQTT connection
```

and apply it `kubectl apply -f metallb-config.yaml`


## Deploy application

Modify `loadBalancerIP` addresses in `kubernetes-manifests/mosquitto.yaml` and `kubernetes-manifests/gui.yaml` to use your FLoating IPs.
Finally, you can deploy your app:

```bash
kubectl apply -f kubernetes-manifests/namespace.yaml
kubectl apply -f kubernetes-manifests/mosquitto.yaml
kubectl apply -f kubernetes-manifests/api-devices.yaml
kubectl apply -f kubernetes-manifests/api-server.yaml
kubectl apply -f kubernetes-manifests/gui.yaml
```

Check kubernetes services! You should see 2 LoadBalancers with the right Floating IPs assigned.

Update DNS records of your domains:

```
A @ <ac-gui-floating-ip_IP_ADDRESS>
A wwww <ac-gui-floating-ip_IP_ADDRESS>
```

```
A @ <ac-mosquitto-floating-ip_IP_ADDRESS>
A wwww <ac-mosquitto-floating-ip_IP_ADDRESS>
```