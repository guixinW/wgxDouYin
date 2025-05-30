#!/bin/bash

# 关闭当前目录所有子目录中的 *-docker-compose.yml 所启动的服务

for dir in */ ; do
  # 查找符合条件的 docker-compose 文件
  if composes=$(ls "${dir}"*-docker-compose.yml 2>/dev/null); then
    echo "进入目录：$dir"
    (
      cd "$dir" || exit
      echo "执行：docker-compose -f *-docker-compose.yml down"
      docker-compose -f *-docker-compose.yml down
    )
  fi
done
