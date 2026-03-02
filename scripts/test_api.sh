#!/bin/bash
# 注意：后端发请求前端也会展示
# 配置
API_URL="http://localhost:8080"
API_KEY="sk-dsve2JfFpBfJh6k3R5zaVmRAOgK1KTifwnI_CwKLdgopz5YE"
SESSION_ID="cc86ebe7-d949-4182-b201-4bb5a4a0eb07"

1. 测试基本查询
echo "=== 测试1: 基本知识库查询 ==="
curl -X POST "$API_URL/api/v1/knowledge-chat/$SESSION_ID" \
  -H "X-API-Key: $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "什么是机器学习"
  }' \
  -N

# echo -e "\n\n=== 测试2: 指定知识库 ==="
# curl -X POST "$API_URL/api/v1/knowledge-chat/$SESSION_ID" \
#   -H "X-API-Key: $API_KEY" \
#   -H "Content-Type: application/json" \
#   -d '{
#     "query": "董氏分型是什么",
#     "knowledge_base_ids": ["6c8f0a64-f2ae-408a-9c68-1ad47b4fd497"]
#   }' \
#   -N