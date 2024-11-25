#!/bin/bash

# 要查找的关键字
keywords=("/home/web_server/A/bin/A" "/home/web_server/B/bin/B")

# 初始化一个空数组来存储找到的进程ID
pids=()

# 查找并杀死进程
for keyword in "${keywords[@]}"; do
    # 使用 ps 命令查找进程ID，并将其添加到 pids 数组中
    while IFS= read -r pid; do
        pids+=("$pid")
    done < <(ps -ef | grep "$keyword" | grep -v grep | awk '{print $2}')
done

if [ ${#pids[@]} -eq 0 ]; then
    echo "没有找到包含 '${keywords[*]}' 的进程."
else
    echo "找到以下进程，将被杀死:"
    printf '%s\n' "${pids[@]}"

    # 杀死进程
    for pid in "${pids[@]}"; do
        kill -9 "$pid"
        echo "已杀死进程: $pid"
    done
fi

sleep 3