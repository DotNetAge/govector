#!/bin/bash

# Update 快速开始.md
perl -0777 -pi -e 's/### 微服务模式：启动与基础 API 调用/### 命令行工具 (CLI) 模式\nGoVector 提供了强大的交互式与非交互式命令行工具，支持缩写参数。\n\n- 启动参数与子命令\n  - \`serve\`：启动 HTTP 服务，例如 \`govector serve -p=18080 -d=govector.db\`\n  - \`ls\`：列出数据库中的所有集合，例如 \`govector ls demo.db\`\n  - \`upsert\`：写入数据，例如 \`govector upsert demo.db -c=demo_col -j='[{"id":"1","vector":[0.1]}]'\`\n  - \`search\`：相似度检索，例如 \`govector search demo.db -c=demo_col -v='[0.1]' -l=10\`\n  - \`count\`：统计点数，例如 \`govector count demo.db -c=demo_col\`\n  - \`delete\`：删除数据，例如 \`govector delete demo.db -c=demo_col -i='["1"]'\`\n\n### 微服务模式：启动与基础 API 调用/g' docs/快速开始.md

