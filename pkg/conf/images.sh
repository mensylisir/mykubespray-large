#!/bin/bash

# ====================================================================================
# ---                           请修改以下配置变量                                 ---
# ====================================================================================

OCI_LAYOUT_PATH="/opt/repository/kubekey/images" # OCI 布局的根目录
HARBOR_URL="dockerhub.kubekey.local"
HARBOR_PROJECT="carsio"
HARBOR_USER="admin"
HARBOR_PASSWORD="Harbor12345"
CTR_NAMESPACE="k8s.io"

# ====================================================================================
# ---                           脚本主逻辑开始                                   ---
# ====================================================================================

# 检查依赖工具
if ! command -v ctr &> /dev/null; then echo "错误: 'ctr' 命令未找到。"; exit 1; fi
if ! command -v docker &> /dev/null; then echo "错误: 'docker' 命令未找到。"; exit 1; fi # 新增 docker 检查
if ! command -v python3 &> /dev/null; then echo "错误: 'python3' 命令未找到。"; exit 1; fi
if ! command -v tar &> /dev/null; then echo "错误: 'tar' 命令未找到。"; exit 1; fi
if ! command -v sort &> /dev/null; then echo "错误: 'sort' 命令未找到。"; exit 1; fi
if ! command -v uniq &> /dev/null; then echo "错误: 'uniq' 命令未找到。"; exit 1; fi

# 新增：启用 Docker CLI 的 manifest 实验性功能
export DOCKER_CLI_EXPERIMENTAL=enabled

# --- 阶段一：打包与导入 (保留你的原始逻辑，使用 ctr) ---

# 切换到 OCI 布局的父目录
cd "$(dirname "$OCI_LAYOUT_PATH")" || { echo "错误: 无法进入目录 $(dirname "$OCI_LAYOUT_PATH")"; exit 1; }
OCI_LAYOUT_BASENAME="$(basename "$OCI_LAYOUT_PATH")"

if [ ! -d "$OCI_LAYOUT_BASENAME" ]; then echo "错误: OCI 布局目录 '${OCI_LAYOUT_BASENAME}' 不存在。"; exit 1; fi

# 使用 Python 3 解析 index.json 获取所有镜像引用
python3 -c '
import json, sys, os
index_file = os.path.join(sys.argv[1], "index.json")
output_file = "image_refs.txt"
refs = []
try:
    with open(index_file, "r") as f:
        data = json.load(f)
    for manifest in data.get("manifests", []):
        ref = manifest.get("annotations", {}).get("org.opencontainers.image.ref.name")
        if ref:
            refs.append(ref)
    with open(output_file, "w") as f:
        for ref in refs:
            f.write(ref + "\n")
except Exception as e:
    print(f"Error processing index.json: {e}", file=sys.stderr)
    sys.exit(1)
' "${OCI_LAYOUT_BASENAME}"

if [ ! -s "image_refs.txt" ]; then echo "错误: 在 index.json 中未找到带 tag 的镜像引用，或解析失败。"; exit 1; fi

# 将整个 OCI 布局目录打包成一个 tar 文件
OCI_TAR_BALL="all_images.tar"
echo "--> 步骤 1/4: 正在将 OCI 布局目录打包成 '${OCI_TAR_BALL}'..."
tar -C "${OCI_LAYOUT_BASENAME}" -cf "${OCI_TAR_BALL}" .
if [ $? -ne 0 ]; then echo "错误: 创建 tar 包失败。"; exit 1; fi

# 使用 ctr images import 导入这个完整的 tar 包
echo "--> 步骤 2/4: 正在将所有镜像从 tar 包导入到 containerd..."
ctr --namespace "${CTR_NAMESPACE}" images import - < "${OCI_TAR_BALL}"
if [ $? -ne 0 ]; then echo "错误: 从 tar 包导入镜像失败。"; rm -f "${OCI_TAR_BALL}"; exit 1; fi
rm -f "${OCI_TAR_BALL}"
echo "    所有镜像已成功导入到 containerd。"


# --- 阶段二：推送单架构镜像 (保留你的原始逻辑，使用 ctr) ---

echo -e "\n--> 步骤 3/4: 正在推送所有单架构镜像 (这是创建多架构清单的前提)..."
# **新增**：同时登录 docker，为最后一步做准备
echo "${HARBOR_PASSWORD}" | docker login "${HARBOR_URL}" -u "${HARBOR_USER}" --password-stdin

while IFS= read -r ref; do
    repo_and_tag=$(echo "$ref" | sed 's|.*/kubesphereio/||')
    HARBOR_TAG="${HARBOR_URL}/${HARBOR_PROJECT}/${repo_and_tag}"
    
    # 使用 ctr 打 tag 和推送
    ctr --namespace "${CTR_NAMESPACE}" images tag "${ref}" "${HARBOR_TAG}"
    # 使用 --skip-verify 来忽略证书错误
    ctr --namespace "${CTR_NAMESPACE}" images push -u "${HARBOR_USER}:${HARBOR_PASSWORD}" --skip-verify "${HARBOR_TAG}"
    if [ $? -ne 0 ]; then
        echo "警告: 推送单架构镜像 ${HARBOR_TAG} 失败，但将继续尝试。"
    else
        echo "    成功推送: ${HARBOR_TAG}"
    fi
done < image_refs.txt


# --- 阶段三：创建并推送多架构清单 (修改：使用 docker manifest) ---

echo -e "\n--> 步骤 4/4: 正在创建并推送多架构清单 (使用 docker manifest)..."
# 提取出不带架构后缀的基础镜像名和版本
cat image_refs.txt | sed -e 's/-amd64$//' -e 's/-arm64$//' | sort | uniq | while IFS= read -r base_ref; do
    
    AMD64_REF="${base_ref}-amd64"
    ARM64_REF="${base_ref}-arm64"

    # 检查配对的 amd64 和 arm64 镜像是否都存在
    if grep -qF -- "${AMD64_REF}" image_refs.txt && grep -qF -- "${ARM64_REF}" image_refs.txt; then
        echo "========================================================"
        echo "发现配对的多架构镜像: $(basename "$base_ref")"

        # 准备目标 Harbor 的 tag
        base_repo_and_tag=$(echo "$base_ref" | sed 's|.*/kubesphereio/||')
        TARGET_MANIFEST_LIST_TAG="${HARBOR_URL}/${HARBOR_PROJECT}/${base_repo_and_tag}"

        AMD64_HARBOR_TAG="${TARGET_MANIFEST_LIST_TAG}-amd64"
        ARM64_HARBOR_TAG="${TARGET_MANIFEST_LIST_TAG}-arm64"

        echo "    AMD64 源: ${AMD64_HARBOR_TAG}"
        echo "    ARM64 源: ${ARM64_HARBOR_TAG}"
        echo "    目标清单: ${TARGET_MANIFEST_LIST_TAG}"

        # ====================  修改的核心在这里 ====================
        # 使用 docker manifest create 创建清单列表
        # --amend 参数会自动从已存在的镜像中提取架构信息
        docker manifest create "${TARGET_MANIFEST_LIST_TAG}" \
            --amend "${AMD64_HARBOR_TAG}" \
            --amend "${ARM64_HARBOR_TAG}"
        
        if [ $? -ne 0 ]; then
            echo "错误: 'docker manifest create' 失败。请检查 Docker 版本和实验性功能是否开启。"
            continue
        fi

        # 使用 docker manifest push 推送清单
        # docker push 会自动处理证书问题（如果 daemon.json 已配置）
        docker manifest push "${TARGET_MANIFEST_LIST_TAG}"
        
        if [ $? -eq 0 ]; then
            echo "    成功创建并推送多架构清单: ${TARGET_MANIFEST_LIST_TAG}"
        else
            echo "错误: 'docker manifest push' 失败。请检查 /etc/docker/daemon.json 中是否配置了 insecure-registries。"
        fi
        echo "========================================================"
        # ==========================================================
    fi
done


# --- 清理工作 ---
echo -e "\n--> 清理 containerd 和 docker 中的镜像..."
while IFS= read -r ref; do
    repo_and_tag=$(echo "$ref" | sed 's|.*/kubesphereio/||')
    HARBOR_TAG="${HARBOR_URL}/${HARBOR_PROJECT}/${repo_and_tag}"
    # 清理 containerd
    ctr --namespace "${CTR_NAMESPACE}" images rm "${ref}" >/dev/null 2>&1
    ctr --namespace "${CTR_NAMESPACE}" images rm "${HARBOR_TAG}" >/dev/null 2>&1
done < image_refs.txt

# 清理 docker 中创建的 manifest list
cat image_refs.txt | sed -e 's/-amd64$//' -e 's/-arm64$//' | sort | uniq | while IFS= read -r base_ref; do
    base_repo_and_tag=$(echo "$base_ref" | sed 's|.*/kubesphereio/||')
    TARGET_MANIFEST_LIST_TAG="${HARBOR_URL}/${HARBOR_PROJECT}/${base_repo_and_tag}"
    docker manifest rm "${TARGET_MANIFEST_LIST_TAG}" >/dev/null 2>&1
done

rm -f image_refs.txt

echo -e "\n迁移完成！"
