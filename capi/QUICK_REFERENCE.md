# GoVector CAPI - 快速参考卡片

## 🚀 30 秒快速开始

```bash
# 1. 进入目录
cd govector/capi

# 2. 构建
./build.sh python

# 3. 运行示例
python3 examples/simple_example.py
```

---

## 📦 安装依赖

### macOS
```bash
brew install swig go python3
```

### Ubuntu
```bash
sudo apt-get install swig golang python3-dev
```

---

## 🔧 常用命令

### 构建脚本
```bash
./build.sh python      # 构建 Python 绑定
./build.sh java        # 构建 Java 绑定
./build.sh csharp      # 构建 C# 绑定
./build.sh clean       # 清理
./build.sh help        # 帮助
```

### Makefile
```bash
make python            # 构建 Python
make java              # 构建 Java
make csharp            # 构建 C#
make clean             # 清理
make help              # 帮助
```

---

## 💻 Python 使用模板

```python
import govector

# 1. 初始化
error = govector.ErrorInfo()
storage = govector.govector_storage_new("test.db", error)

# 2. 创建集合
params = govector.HNSWParams()
params.m = 16
collection = govector.govector_collection_create(
    "my_col", 128, govector.DISTANCE_COSINE,
    True, params, storage, error
)

# 3. 使用...
count = govector.govector_collection_count(collection)

# 4. 清理
govector.govector_collection_free(collection)
govector.govector_storage_free(storage)
```

---

## ⚠️ 重要提醒

### ✅ 必做
- [ ] 总是检查返回值
- [ ] 释放所有资源（try-finally）
- [ ] 在 capi 目录下运行

### ❌ 禁止
- [ ] 不要忘记释放内存
- [ ] 不要忽略错误
- [ ] 不要在 core 包中使用 CGO

---

## 🆘 故障排除

| 问题 | 解决方案 |
|------|---------|
| `swig: command not found` | `brew install swig` |
| `Python.h: No such file` | `brew install python3` |
| `module not found` | `cd govector/capi` |
| `undefined symbol` | `./build.sh clean && ./build.sh python` |

---

## 📚 文档索引

1. **BUILD_GUIDE.md** - 详细构建指南
2. **README.md** - 目录说明
3. **../docs/CGO_QUICKSTART.md** - 快速入门
4. **../docs/CGO_ISOLATION_DESIGN.md** - 架构设计

---

## 🎯 下一步

1. ✅ 运行 `python3 examples/simple_example.py`
2. ✅ 修改示例代码进行实验
3. ✅ 阅读更多文档了解高级用法

---

**完整文档**: 查看 [`COMPLETION_SUMMARY.md`](COMPLETION_SUMMARY.md)

**Happy Coding! 🚀**
