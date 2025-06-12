sudo apt install -y qemu-kvm libvirt-daemon-system bridge-utils virt-manager cloud-image-utils
mkdir -p /var/lib/libvirt/images/templates
cd /var/lib/libvirt/images/templates
wget https://cloud-images.ubuntu.com/jammy/current/jammy-server-cloudimg-amd64.img
qemu-img create -f qcow2 -o backing_fmt=qcow2 -b jammy-server-cloudimg-amd64.img ubuntu2204-base.qcow2 20G

#!/bin/bash
# --- create_vm_nat_static_ip.sh ---

# --- 变量配置 ---
# libvirt存储池和网络名称
STORAGE_POOL_NAME="default" # 脚本会自动检测并使用活跃的池
NETWORK_NAME="default"

# 模板卷的名称
TEMPLATE_VOL_NAME="ubuntu2204-base.qcow2"

# 虚拟机的root密码
ROOT_PASSWORD="Def@u1tpwd"

# 虚拟机通用资源配置
CPUS=2
MEM_MB=4096

# --- 准备工作：智能地检测和准备环境 ---
# 1. 自动检测活跃的存储池
DETECTED_POOL=$(virsh pool-list --all | grep -i 'active' | awk '{print $1}' | head -n 1)
if [ -n "$DETECTED_POOL" ]; then
    STORAGE_POOL_NAME=$DETECTED_POOL
fi
echo "成功找到并使用活跃存储池: '$STORAGE_POOL_NAME'"
STORAGE_POOL_PATH=$(virsh pool-dumpxml "$STORAGE_POOL_NAME" | grep -oP '(?<=<path>).*(?=</path>)')

# 2. 检查模板是否存在
if ! virsh vol-info --pool "$STORAGE_POOL_NAME" "$TEMPLATE_VOL_NAME" >/dev/null 2>&1; then
    echo "错误: 模板卷 '$TEMPLATE_VOL_NAME' 在存储池 '$STORAGE_POOL_NAME' 中未找到!"
    exit 1
fi

# 3. 激活默认NAT网络
if ! virsh net-info "$NETWORK_NAME" | grep -q 'Active: *yes'; then
    echo "正在激活 '$NETWORK_NAME' 网络..."
    virsh net-start "$NETWORK_NAME"
    virsh net-autostart "$NETWORK_NAME"
fi

# --- 脚本主逻辑 ---
echo "准备通过QEMU/KVM(NAT+静态IP)批量创建3台虚拟机..."

for i in {1..3}
do
    VM_NAME="vm$i"
    # 我们为每台VM预先定义好IP和MAC地址
    VM_IP="192.168.122.10${i}" # vm1->.101, vm2->.102, ...
    VM_MAC="52:54:00:00:01:$(printf '%02x' $i)" # 生成唯一的MAC地址
    VM_VOL_NAME="${VM_NAME}.qcow2"
    CLOUD_INIT_ISO="${STORAGE_POOL_PATH}/${VM_NAME}-cloud-init.iso"

    echo "-----------------------------------------------------"
    echo "正在创建虚拟机: $VM_NAME"
    echo "  - 固定IP: $VM_IP"
    echo "  - 固定MAC: $VM_MAC"
    echo "-----------------------------------------------------"

    # 如果同名虚拟机已存在，则先销毁并删除
    if virsh dominfo "$VM_NAME" >/dev/null 2>&1; then
        echo "虚拟机 $VM_NAME 已存在，正在销毁并删除..."
        virsh destroy "$VM_NAME" --graceful || virsh destroy "$VM_NAME"
        virsh undefine "$VM_NAME" --remove-all-storage
    fi
    rm -f "$CLOUD_INIT_ISO"

    # 步骤 1: 在libvirt网络中添加/更新DHCP静态租约
    echo "步骤 1: 为 $VM_MAC -> $VM_IP 添加DHCP静态租约..."
    virsh net-update "$NETWORK_NAME" add ip-dhcp-host \
      "<host mac='$VM_MAC' name='$VM_NAME' ip='$VM_IP' />" \
      --live --config

    if [ $? -ne 0 ]; then
        echo "添加DHCP租约失败! 请检查libvirt网络配置。中止脚本。"
        exit 1
    fi

    # 步骤 2: 克隆磁盘卷并调整大小
    echo "步骤 2: 使用 virsh 克隆并调整磁盘卷..."
    virsh vol-clone --pool "$STORAGE_POOL_NAME" "$TEMPLATE_VOL_NAME" "$VM_VOL_NAME"
    virsh vol-resize --pool "$STORAGE_POOL_NAME" "$VM_VOL_NAME" 20G

    # 步骤 3: 生成 cloud-init 配置文件 (不再需要网络配置)
    USER_DATA=$(cat <<EOF
#cloud-config
# 网络配置由libvirt的DHCP静态租约管理，这里留空
hostname: $VM_NAME
manage_etc_hosts: true
chpasswd:
  list: |
    root:$ROOT_PASSWORD
  expire: False
ssh_pwauth: True
runcmd:
  - [ sed, -i, 's/^#?PermitRootLogin.*/PermitRootLogin yes/', /etc/ssh/sshd_config ]
  - [ systemctl, restart, sshd ]
EOF
)
    META_DATA="instance-id: ${VM_NAME}-$(uuidgen)\nlocal-hostname: ${VM_NAME}"

    # 步骤 4: 创建配置盘
    echo "步骤 3: 创建Cloud-Init配置盘..."
    cloud-localds "$CLOUD_INIT_ISO" <(echo "$USER_DATA") <(echo "$META_DATA")

    # 步骤 5: 使用 virt-install 定义并启动虚拟机 (指定了MAC地址)
    echo "步骤 4: 使用virt-install定义并启动虚拟机..."
    virt-install --name "$VM_NAME" \
      --virt-type kvm \
      --memory "$MEM_MB" \
      --vcpus "$CPUS" \
      --os-variant ubuntu22.04 \
      --disk vol="${STORAGE_POOL_NAME}/${VM_VOL_NAME}",device=disk,bus=virtio \
      --disk path="$CLOUD_INIT_ISO",device=cdrom \
      --network network="$NETWORK_NAME",mac="$VM_MAC",model=virtio \
      --graphics none \
      --noautoconsole \
      --import

    if [ $? -ne 0 ]; then
        echo "创建虚拟机 $VM_NAME 失败! 中止脚本。"
        exit 1
    fi

    echo "虚拟机 $VM_NAME 已成功创建并启动，IP地址应为 $VM_IP 。"
done

echo "-----------------------------------------------------"
echo "所有虚拟机已创建并启动!"
echo "等待约30-60秒让cloud-init完成配置后，即可通过固定IP SSH连接:"
echo "  vm1: ssh root@192.168.122.101"
echo "  vm2: ssh root@192.168.122.102"
echo "  vm3: ssh root@192.168.122.103"