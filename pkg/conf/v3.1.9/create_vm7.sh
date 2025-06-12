#!/bin/bash

# --- 请根据你的环境修改以下变量 ---

# 主机的网络接口 (用于桥接)
# 使用 `ip a` 命令查看你的物理网卡名
HOST_INTERFACE="enp0s25"

# Ubuntu Server ISO 镜像的完整路径
ISO_PATH="$(pwd)/ubuntu-22.04.4-live-server-amd64.iso" # 假设ISO在当前目录

# 你的网络网关
NETWORK_GATEWAY="10.144.169.1"

# 子网掩码 (CIDR格式)
NETWORK_CIDR="24" # 24 代表 255.255.255.0

# DNS 服务器 (用空格分隔)
DNS_SERVERS="172.30.1.13 172.17.0.13"

# 虚拟机的 root 密码
ROOT_PASSWORD="Def@u1tpwd"

# 虚拟机通用资源配置
CPUS=2
MEM_MB=4096    # VirtualBox 使用 MB 作为单位
DISK_MB=20480  # 20G

# --- 脚本主逻辑 ---

# 检查 ISO 文件是否存在
if [ ! -f "$ISO_PATH" ]; then
    echo "错误: Ubuntu ISO 文件 '$ISO_PATH' 未找到!"
    echo "请先下载: wget https://releases.ubuntu.com/22.04/ubuntu-22.04.4-live-server-amd64.iso"
    exit 1
fi

echo "准备批量创建3台虚拟机 (vm1, vm2, vm3)..."

# 将空格分隔的DNS服务器列表转换为YAML格式的列表
DNS_YAML_LIST=$(echo "$DNS_SERVERS" | awk '{for(i=1;i<=NF;i++) printf "          - %s\n", $i}')

# 循环创建3台虚拟机
for i in {1..3}
do
    VM_NAME="vm$i" # <<< 已修改为 vm1, vm2, vm3
    # IP地址从 10.144.169.2 开始
    VM_IP="10.144.169.$((i+1))"
    USER_DATA_FILE="user-data-for-$VM_NAME"

    echo "-----------------------------------------------------"
    echo "正在创建虚拟机: $VM_NAME"
    echo "IP 地址: $VM_IP"
    echo "Root 密码: $ROOT_PASSWORD"
    echo "-----------------------------------------------------"

    # 检查同名虚拟机是否已存在
    if VBoxManage showvminfo "$VM_NAME" >/dev/null 2>&1; then
        echo "虚拟机 $VM_NAME 已存在，正在删除..."
        VBoxManage unregistervm "$VM_NAME" --delete
    fi

    # 1. 生成 user-data 应答文件
    # 注意: Ubuntu自动化安装需要一个初始用户，我们创建它但不主要使用
    INITIAL_USER="tempuser"
    cat > "$USER_DATA_FILE" <<EOF
#cloud-config
autoinstall:
  version: 1
  identity:
    hostname: $VM_NAME
    username: $INITIAL_USER
    password: "temporarypassword"
  network:
    version: 2
    ethernets:
      enp0s3: # VirtualBox虚拟机的第一个网卡通常是enp0s3
        dhcp4: no
        addresses:
          - $VM_IP/$NETWORK_CIDR
        gateway4: $NETWORK_GATEWAY
        nameservers:
          addresses:
$(echo -e "$DNS_YAML_LIST")
  ssh:
    install-server: yes
    allow-pw: yes
  refresh-installer:
    update: no
EOF

    # 2. 定义安装后立即执行的命令：
    #    a) 设置root密码
    #    b) 允许root通过SSH登录
    #    c) 重启SSH服务使配置生效
    POST_INSTALL_COMMAND="echo 'root:$ROOT_PASSWORD' | sudo chpasswd && sudo sed -i 's/^#?PermitRootLogin.*/PermitRootLogin yes/' /etc/ssh/sshd_config && sudo systemctl restart sshd"

    # 3. 使用 VBoxManage unattended install 创建并开始安装
    VBoxManage unattended install "$VM_NAME" \
      --iso="$ISO_PATH" \
      --user-data="$USER_DATA_FILE" \
      --install-additions \
      --hostname="$VM_NAME" \
      --user="$INITIAL_USER" \
      --password="temporarypassword" \
      --full-user-name="$INITIAL_USER" \
      --time-zone="auto" \
      --post-install-command="$POST_INSTALL_COMMAND" # <<< 执行我们定义好的后期命令

    # 检查命令是否成功
    if [ $? -ne 0 ]; then
        echo "为虚拟机 $VM_NAME 启动无人值守安装失败! 中止脚本。"
        rm -f "$USER_DATA_FILE"
        exit 1
    fi

    # 4. 修改虚拟机的硬件配置
    echo "正在为 $VM_NAME 配置硬件..."
    VBoxManage modifyvm "$VM_NAME" \
      --cpus "$CPUS" \
      --memory "$MEM_MB" \
      --nic1 bridged --bridgeadapter1 "$HOST_INTERFACE"

    # 5. 修改虚拟磁盘大小
    VDI_FILE=$(VBoxManage showvminfo "$VM_NAME" --machinereadable | grep "SATA-0-0" | cut -d'"' -f2)
    if [ -n "$VDI_FILE" ]; then
        echo "正在调整磁盘大小至 ${DISK_MB}MB..."
        VBoxManage modifymedium disk "$VDI_FILE" --resize "$DISK_MB"
    fi

    # 6. 启动虚拟机开始真正的安装过程
    echo "启动虚拟机 $VM_NAME 开始后台安装..."
    VBoxManage startvm "$VM_NAME" --type headless

    # 清理临时的应答文件
    rm -f "$USER_DATA_FILE"

    echo "虚拟机 $VM_NAME 已启动并在后台进行无人值守安装。"
done

echo "-----------------------------------------------------"
echo "所有虚拟机创建命令已发出!"
echo "安装过程将在后台进行，需要一些时间。"
echo "你可以使用 'VBoxManage list runningvms' 查看正在运行的虚拟机。"
echo "安装完成后，就可以通过 'ssh root@<ip_address>' 并使用密码 '$ROOT_PASSWORD' 连接了。"