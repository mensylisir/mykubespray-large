#!/bin/bash

# 用户名和密码
USERNAME="wangzhendong"
PASSWORD="Def@u1tpwd"

# 创建新用户
adduser $USERNAME

# 设置用户密码
echo "$USERNAME:$PASSWORD" | chpasswd

# 将新用户添加到 wheel 组
usermod -aG wheel $USERNAME

# 确保 sudo 权限已启用
echo "%wheel ALL=(ALL) ALL" >> /etc/sudoers.d/wheel

echo "用户 $USERNAME 已创建并赋予 sudo 权限。"
