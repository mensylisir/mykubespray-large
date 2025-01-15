#!/bin/bash

# 私钥路径
PRIVATE_KEY="/var/tmp/ssh_config/localhuoyun"

# 读取 host.info 中的 IP 列表，按行循环处理
while IFS= read -r ip; do
    # 跳过空行
    if [[ -z "$ip" ]]; then
        continue
    fi

    echo "Processing host: $ip"

    # SSH 命令：使用私钥连接远程主机
    ssh -i "$PRIVATE_KEY" -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null "root@$ip" << EOF
        # 开始硬盘分区操作
        echo "Starting partitioning on $ip..."

        # 执行 fdisk 分区操作
        (echo -e "n\n\n\n\nw" | fdisk /dev/sda) && echo "Partitioning completed on $ip" || echo "Partitioning failed on $ip"

        # 等待系统识别新分区
        sleep 5

        # 执行 pvcreate 操作
        pvcreate /dev/sda4 && echo "Physical volume created on $ip" || echo "Failed to create physical volume on $ip"

        # 执行 vgextend 操作
        vgextend uos /dev/sda4 && echo "Volume group extended on $ip" || echo "Failed to extend volume group on $ip"

        # 执行 lvextend 操作
        lvextend -l +100%FREE /dev/mapper/uos-var && echo "Logical volume extended on $ip" || echo "Failed to extend logical volume on $ip"

        # 执行 xfs_growfs 操作
        xfs_growfs /dev/mapper/uos-var && echo "Filesystem resized on $ip" || echo "Failed to resize filesystem on $ip"
EOF

    # 检查 SSH 是否成功执行
    if [ $? -eq 0 ]; then
        echo "Successfully completed operations on $ip"
    else
        echo "Failed to complete operations on $ip"
    fi
done < "host.info"
