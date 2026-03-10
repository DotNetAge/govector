# 🚀 GoVector Launch & Promotion Guideline (产品发布与全网推广操作指南)

这份指南汇总了从代码打包、系统服务发布到全网（海内外核心开发者社区）分发的所有步骤与最佳实践。请按照顺序在今晚依次执行，确保每一步的势能最大化。

---

## 🟢 阶段一：基础设施就绪 (Pre-launch Setup)

在发任何软文之前，必须确保用户点进你的 GitHub 后能够顺滑地下载和安装。

### 1. 完善 GitHub 主仓库 (`govector`)
- [ ] 将本地所有代码 commit 并 push 到你的 GitHub 仓库 (`DotNetAge/govector`)。
- [ ] 确保 `README.md` 和 `LICENSE` 已经更新且排版正常。
- [ ] 在本地终端执行 `make release`，检查 `dist/` 目录下是否成功生成了 `.tar.gz` 和 `.zip` 压缩包。

### 2. 发布 GitHub Release
- [ ] 去 GitHub 仓库页面，点击右侧的 **Releases -> Draft a new release**。
- [ ] Tag 填写 `v0.1.0`，Title 填写 `v0.1.0: First Release - The SQLite for Vectors`。
- [ ] 把 `dist/` 目录下的 4 个平台压缩包拖拽上传到 Assets 区域，点击 Publish。

### 3. 创建 Homebrew Tap (关键引流利器)
- [ ] 在你的 GitHub 新建一个**公开仓库**，命名为 `homebrew-govector`。
- [ ] 在本地运行命令获取刚刚发布的 Mac/Linux 压缩包的 SHA256 校验码：
  ```bash
  shasum -a 256 dist/govector_v0.1.0_darwin_arm64.tar.gz
  shasum -a 256 dist/govector_v0.1.0_darwin_amd64.tar.gz
  shasum -a 256 dist/govector_v0.1.0_linux_amd64.tar.gz
  ```
- [ ] 打开项目里的 `scripts/release/govector.rb` 文件，将对应的 `URL` (替换为你真正的 Release 链接) 和刚刚算出的 `SHA256` 填入。
- [ ] 将修改好的 `govector.rb` 上传到 `homebrew-govector` 仓库的根目录。
- [ ] 自己在终端里测试一下：`brew tap DotNetAge/govector && brew install govector`，确认安装成功。

---

## 🔵 阶段二：引爆 X (Twitter) 社区

**⏰ 最佳发布时间**：北京时间 21:00 - 23:00 (覆盖欧美程序员早晨刷推时间)

### 1. 准备素材 (极其重要)
不要只发干瘪的文字。请准备两张图（二选一或都发）：
1. 终端里运行 `make test` 跑出 17000+ QPS 和 0.058ms 延迟的黑底绿字截图。
2. 极其简短的嵌入式代码截图 (推荐使用 carbon.now.sh 增加代码颜值)。

### 2. 发布文案 (直接复制)
```text
Tired of running heavy Docker containers just for a local vector search? 😫

I built GoVector — The "SQLite for Vectors" in pure Go! 🚀

Key Highlights:
✅ 100% Pure Go (CGO-free, cross-compile anywhere)
✅ Qdrant compatible (swap in/out seamlessly)
✅ HNSW Indexing (17,000+ QPS ⚡️)
✅ Local persistence via BoltDB
✅ Embeddable OR Standalone Microservice

The "Wow" Factor:
Search latency: 58 microseconds. 
Install it in seconds: `brew install DotNetAge/govector/govector`

Check it out on GitHub: 
https://github.com/DotNetAge/govector

#golang #AI #VectorDB #LocalAI #RAG #OpenSource
```

---

## 🟣 阶段三：攻克 Hacker News & Reddit (硬核英文社区)

**⏰ 最佳发布时间**：北京时间 20:00 - 22:00

### 1. 发往 Hacker News
- [ ] 标题：`Show HN: GoVector – An embeddable, pure Go vector database (SQLite for vectors)`
- [ ] 链接：填你的 GitHub 地址。
- [ ] 文本：HN 通常只留 URL，但如果你想写字，可以把阶段三下面准备的文本精简后放上去。

### 2. 发往 Reddit (`r/golang`, `r/LocalLLaMA`, `r/selfhosted`)
在这三个节点分别发帖，标题可以略有不同，重点突出“痛点”。
- [ ] 标题：`I built a pure Go embeddable vector database (Qdrant-compatible) because existing tools were too heavy for local AI apps`
- [ ] 文本：打开根目录的 `promo_post.md`，全选复制。别忘了把文中的 GitHub 占位符链接替换成真实链接。

### 📌 互动铁律
- **保持谦逊与极客精神**：有人提尖锐问题（比如质疑为什么要重新造轮子），用“解决自己的特定部署痛点（无CGO、无Docker依赖）”来回应，这在英文社区极具说服力。
- **一楼彩蛋**：发完贴后自己抢沙发：“*OP here! If you just want to test it out, the Homebrew tap takes exactly 10 seconds to install and start the daemon: `brew tap DotNetAge/govector && brew install govector`*”

---

## 🔴 阶段四：收割国内流量 (CSDN / 掘金 / 知乎)

国内社区更喜欢看“手把手造轮子”的干货教程。

- [ ] 打开项目根目录下的 `csdn_post.md`。
- [ ] 检查并替换文末的 GitHub 链接。
- [ ] 在 CSDN 和 掘金 上以文章形式发布。
- [ ] 标签打上：`Go`, `人工智能`, `向量数据库`, `开源项目`, `后端开发`。
- [ ] **配图建议**：文章开头放一张架构图或者你测试 Benchmark 的超速截图。

---

## 🏁 终极 Checklist
- [ ] 代码 Push 完成
- [ ] GitHub Release 发布完成
- [ ] Homebrew 仓库配置完毕，测试安装成功
- [ ] X (Twitter) 推文发送 (带图)
- [ ] Reddit 节点发送
- [ ] Hacker News 发布
- [ ] 国内社区 (掘金/CSDN/知乎) 发布

**祝今晚发版顺利！坐等 GitHub 涨 Star！⭐️**