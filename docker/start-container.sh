#!/bin/bash

# 脚本说明：递归查找当前目录下所有匹配 "*-docker-compose.yml" 的文件，并执行 docker-compose up -d

# 获取当前路径
BASE_DIR=$(pwd)

# 查找所有符合条件的 docker-compose 文件
find "$BASE_DIR" -type f -name '*-docker-compose.yml' | while read -r compose_file; do
  # 获取 docker-compose.yml 所在目录
  dir=$(dirname "$compose_file")

  echo "启动目录：$dir 中的 $(basename "$compose_file")"

  # 进入目录并启动 docker-compose
  (cd "$dir" && docker-compose -f "$(basename "$compose_file")" up -d)
done
