#!/bin/sh
#程序启动脚本
#注意一定要保持前台运行，不然docker会退出
ulimit -n 65536

# 设置日志目录
log_dir="/home/web_server/kuaishou-runner/logs"
timestamp=$(date +%s)  # 获取当前 Unix 时间戳

# 查找并重命名所有 .log 文件
if ls "$log_dir"/*.log 1> /dev/null 2>&1; then
    echo "正在重命名以下 .log 文件："
    for log_file in "$log_dir"/*.log; do
        # 获取文件名和扩展名
        base_name=$(basename "$log_file" .log)
        new_name="${base_name}_${timestamp}.log"
        
        # 重命名文件
        mv "$log_file" "$log_dir/$new_name"
        echo "已重命名: $log_file -> $new_name"
    done
    echo "所有 .log 文件已重命名。"
else
    echo "在目录 $log_dir 下没有找到任何 .log 文件。"
fi

/home/web_server/B/bin/B -log_dir=/home/web_server/kuaishou-runner/logs -hp=${AUTO_PORT0} -gp=${AUTO_PORT1} -debug=false -zk_addr=yanlian-test-zk.internal:2181 -cpu_num=4 > /home/web_server/kuaishou-runner/logs/stdout.log 2>&1

# /home/web_server/B/bin/B -log_dir=/home/web_server/kuaishou-runner/logs -hp=${AUTO_PORT0} -gp=${AUTO_PORT1} -debug=true -safety_factor=1.5 -max_exec_time=1000000 -zk_addr=yanlian-test-zk.internal:2181
