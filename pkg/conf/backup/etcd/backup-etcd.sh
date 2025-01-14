#!/bin/bash

if [ -f /etc/etcd.env ]; then
    source /etc/etcd.env
else
    echo "Error: /etc/etcd.env file not found!"
    exit 1
fi

# 检查必需的环境变量
if [ -z "$ETCDCTL_ENDPOINTS" ] || [ -z "$ETCDCTL_CERT_FILE" ] || [ -z "$ETCDCTL_KEY_FILE" ] || [ -z "$ETCDCTL_CA_FILE" ]; then
    echo "Error: Missing environment variables in /etc/etcd.env!"
    exit 1
fi

BACKUP_DIR="/var/backup"
if [ -z "$BACKUP_DIR" ]; then
  mkdir -p $BACKUP_DIR
fi

TIMESTAMP=$(date +\%Y-\%m-\%d-\%H-\%M-\%S)
BACKUP_FILE="${BACKUP_DIR}/etcd-backup-${TIMESTAMP}.db"


# 使用 etcdctl 进行备份
ETCDCTL_API=3 etcdctl --endpoints=$ETCDCTL_ENDPOINTS \
    --cert-file=$ETCDCTL_CERT_FILE \
    --key-file=$ETCDCTL_KEY_FILE \
    --ca-file=$ETCDCTL_CA_FILE \
    snapshot save $BACKUP_FILE

# 检查备份是否成功
if [ $? -eq 0 ]; then
    echo "Backup successful: $BACKUP_FILE"
else
    echo "Backup failed!"
    exit 1
fi

# 删除30天以前的备份
find $BACKUP_DIR -type f -name "etcd-backup-*.db" -mtime +30 -exec rm -f {} \;

# 输出删除的文件
echo "Old backups (older than 30 days) have been deleted."
