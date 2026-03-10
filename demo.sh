#!/bin/bash

echo "🚀 [1/4] 正在后台启动 GoVector 数据库服务..."
go run cmd/govector-server/main.go > server.log 2>&1 &
SERVER_PID=$!
sleep 2 # 等待服务启动完成

echo "✅ 服务已启动 (PID: $SERVER_PID)"
echo ""

echo "📦 [2/4] 正在通过 HTTP API 插入测试数据 (包含水果和动物)..."
curl -s -X PUT http://localhost:18080/collections/test_collection/points -H "Content-Type: application/json" -d '{
  "points": [
    {"id": "1", "vector": [1.0, 0.0, 0.0], "payload": {"category": "fruit", "name": "Apple"}},
    {"id": "2", "vector": [0.9, 0.1, 0.0], "payload": {"category": "fruit", "name": "Orange"}},
    {"id": "3", "vector": [0.0, 1.0, 0.0], "payload": {"category": "animal", "name": "Dog"}},
    {"id": "4", "vector": [0.0, 0.9, 0.1], "payload": {"category": "animal", "name": "Cat"}}
  ]
}' | jq .
echo ""

echo "🔍 [3/4] 正在执行无过滤的全局搜索 (寻找最接近 [0.8, 0.2, 0.0] 的前3个事物)..."
curl -s -X POST http://localhost:18080/collections/test_collection/points/search -H "Content-Type: application/json" -d '{
  "vector": [0.8, 0.2, 0.0],
  "limit": 3
}' | jq .
echo ""

echo "🎯 [4/4] 正在执行带 Payload 过滤的搜索 (条件: 必须是 animal)..."
curl -s -X POST http://localhost:18080/collections/test_collection/points/search -H "Content-Type: application/json" -d '{
  "vector": [0.8, 0.2, 0.0],
  "limit": 2,
  "filter": {
    "must": [
      { "key": "category", "match": { "value": "animal" } }
    ]
  }
}' | jq .
echo ""

echo "🛑 测试完成，正在关闭服务..."
kill $SERVER_PID
