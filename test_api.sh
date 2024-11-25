#!/bin/bash

# 定义请求的 URL
URL="http://localhost:10001/sha256/calc"

# 定义请求体
DATA='{
    "inputValue": "KSChallenge2024",
"number":10000
}'

# 使用 curl 发送 POST 请求
response=$(curl -s -X POST "$URL" \
-H "Content-Type: application/json" \
-d "$DATA")

# 输出响应
echo "Response:"
echo "$response"
