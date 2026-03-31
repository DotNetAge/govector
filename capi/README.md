# GoVector CAPI - CGO/SWIG  bindings

[![Go Vector](https://img.shields.io/badge/Go-Vector-blue)](https://github.com/DotNetAge/govector)
[![Python](https://img.shields.io/badge/Python-3.8+-blue.svg)](https://www.python.org/)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

**您不需要懂 C++ 或 SWIG！** 

GoVector 的跨语言绑定层，让您能够从 Python、Java、C# 等语言调用 GoVector 向量数据库。

---

## 🎯 特点

- ✅ **零配置** - 一键构建，无需手动配置
- ✅ **多语言支持** - Python、Java、C# 等
- ✅ **高性能** - 直接调用 Go 代码，无 HTTP 开销
- ✅ **类型安全** - SWIG 自动生成类型安全的绑定
- ✅ **易于使用** - 提供丰富的示例代码

---

## 🚀 快速开始

### 1. 安装依赖

```bash
# macOS
brew install swig go

# Ubuntu
sudo apt-get install swig golang
```

### 2. 构建 Python 绑定

```bash
cd govector/capi
./build.sh python
```

### 3. 运行示例

```bash
python3 examples/simple_example.py
```

就这么简单！🎉

---

## 📦 目录结构

```
capi/
├── govector_c.h          # C API 头文件
├── govector_c.go         # Go 导出实现
├── govector.i            # SWIG 接口配置
├── build.sh              # 构建脚本（一键构建）
├── Makefile              # Make 构建配置
├── BUILD_GUIDE.md        # 详细构建指南
├── README.md             # 本文件
├── examples/             # 示例代码
│   ├── simple_example.py # 简单示例
│   └── demo.py           # 完整演示
└── [构建产物]
    ├── libgovector_c.a   # Go 静态库
    ├── govector_wrap.cxx # SWIG 生成的包装
    ├── govector.py       # Python 包装模块
    └── _govector.so      # Python 共享库
```

---

## 💻 使用示例

### Python 示例

```python
import govector

# 初始化
error = govector.ErrorInfo()
storage = govector.govector_storage_new("test.db", error)

# 创建集合
params = govector.HNSWParams()
params.m = 16
collection = govector.govector_collection_create(
    "my_col", 
    128, 
    govector.DISTANCE_COSINE,
    True, 
    params, 
    storage, 
    error
)

# 搜索
query = [0.1] * 128
results = govector.govector_collection_search(
    collection, 
    query, 
    len(query), 
    10,
    ...
)

# 清理资源
govector.govector_collection_free(collection)
govector.govector_storage_free(storage)
```

更多示例请查看 [`examples/`](examples/) 目录。

---

## 🔧 构建命令

### 构建脚本（推荐）

```bash
# 构建 Python 绑定
./build.sh python

# 构建 Java 绑定
./build.sh java

# 构建 C# 绑定
./build.sh csharp

# 清理
./build.sh clean

# 显示帮助
./build.sh help
```

### Makefile

```bash
# 构建 Python 绑定
make python

# 构建所有
make all

# 清理
make clean
```

---

## 📚 文档

| 文档 | 说明 |
|------|------|
| [BUILD_GUIDE.md](BUILD_GUIDE.md) | 详细构建指南（推荐先看这个） |
| [../docs/CGO_QUICKSTART.md](../docs/CGO_QUICKSTART.md) | 快速入门指南 |
| [../docs/CGO_ISOLATION_DESIGN.md](../docs/CGO_ISOLATION_DESIGN.md) | 架构设计说明 |
| [../docs/SWIG_INTEGRATION_GUIDE.md](../docs/SWIG_INTEGRATION_GUIDE.md) | SWIG 集成指南 |
| [../docs/SWIG_EXAMPLES_AND_BEST_PRACTICES.md](../docs/SWIG_EXAMPLES_AND_BEST_PRACTICES.md) | 示例与最佳实践 |

---

## ⚠️ 重要提示

### 内存管理

⚠️ **必须手动释放资源！**

```python
# ✓ 正确
storage = govector.govector_storage_new(...)
try:
    # 使用 storage...
finally:
    govector.govector_storage_free(storage)

# ✗ 错误 - 内存泄漏！
storage = govector.govector_storage_new(...)
# 忘记释放
```

### 错误处理

⚠️ **总是检查返回值！**

```python
# ✓ 正确
error = govector.ErrorInfo()
handle = govector.function(error)
if not handle:
    print(f"Error: {error.message}")

# ✗ 错误
handle = govector.function(None)
# 不检查是否成功
```

---

## 🆘 故障排除

### 常见问题

**Q: `swig: command not found`**
```bash
brew install swig  # macOS
sudo apt-get install swig  # Ubuntu
```

**Q: `Python.h: No such file or directory`**
```bash
brew install python3  # macOS
sudo apt-get install python3-dev  # Ubuntu
```

**Q: `_govector.so: undefined symbol`**
```bash
cd govector/capi
./build.sh clean
./build.sh python
```

更多问题解答请查看 [BUILD_GUIDE.md](BUILD_GUIDE.md#故障排除)。

---

## 📊 性能对比

| 调用方式 | 延迟 | 适用场景 |
|---------|------|---------|
| Go 直接调用 | 1x | Go 应用内部 ⭐⭐⭐⭐⭐ |
| CGO 批量调用 | 2-5x | Python/C++ 应用 ⭐⭐⭐⭐ |
| HTTP API | 100x+ | 远程服务 ⭐ |

---

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

---

## 📄 许可证

与 GoVector 主项目相同。

---

## 🙏 致谢

感谢以下开源项目：

- [SWIG](http://www.swig.org/) - Simplified Wrapper and Interface Generator
- [Go](https://golang.org/) - The Go Programming Language
- [GoVector](https://github.com/DotNetAge/govector) - Go 向量数据库

---

**Happy Coding! 🚀**
