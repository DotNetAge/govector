# GoVector CAPI 构建指南

**您不需要懂 C++ 或 SWIG！** 本指南将带您一步步完成构建过程。

---

## 📋 目录

- [快速开始](#快速开始)
- [环境准备](#环境准备)
- [构建步骤](#构建步骤)
- [使用示例](#使用示例)
- [故障排除](#故障排除)

---

## 🚀 快速开始

### 最简单的构建方式（推荐）

```bash
# 1. 进入 capi 目录
cd govector/capi

# 2. 一键构建 Python 绑定
./build.sh python

# 3. 运行测试
python3 examples/simple_example.py
```

就这么简单！🎉

---

## 📦 环境准备

### macOS

```bash
# 安装 Homebrew（如果还没有）
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# 安装依赖
brew install swig
brew install go
brew install python3
```

### Ubuntu/Debian

```bash
# 更新包列表
sudo apt-get update

# 安装依赖
sudo apt-get install swig
sudo apt-get install golang
sudo apt-get install python3-dev
```

### Windows (WSL)

```bash
# 在 WSL 中安装
sudo apt-get update
sudo apt-get install swig golang python3-dev
```

### 验证安装

```bash
# 检查 Go
go version

# 检查 SWIG
swig -version

# 检查 GCC
gcc --version

# 检查 Python3
python3 --version
```

---

## 🔨 构建步骤

### 方法 1: 使用构建脚本（最简单）

```bash
cd govector/capi

# 构建 Python 绑定
./build.sh python

# 或者构建所有支持的语言
./build.sh all
```

### 方法 2: 使用 Makefile

```bash
cd govector/capi

# 构建 Python 绑定
make python

# 构建 Java 绑定
make java

# 构建 C# 绑定
make csharp

# 清理
make clean
```

### 方法 3: 手动构建（不推荐，仅用于调试）

```bash
cd govector/capi

# 1. 构建 Go 静态库
go build -buildmode=c-archive -o libgovector_c.a ./capi

# 2. 生成 SWIG 包装文件
swig -cgo -python -py3 -o govector_wrap.cxx govector.i

# 3. 编译 Python 模块
g++ -fPIC -shared govector_wrap.cxx -o _govector.so \
    $(python3-config --includes) \
    -L. -lgovector_c
```

---

## 📁 构建产物

构建成功后会生成以下文件：

```
capi/
├── libgovector_c.a      # Go 静态库
├── govector_wrap.cxx    # SWIG 生成的 C++ 包装文件
├── govector.py          # Python 包装模块
├── _govector.so         # Python 共享库
└── examples/
    └── simple_example.py # Python 示例
```

---

## 💻 使用示例

### Python 使用示例

#### 1. 简单示例

```python
import govector

# 初始化
error = govector.ErrorInfo()
storage = govector.govector_storage_new("test.db", error)

# 创建集合
params = govector.HNSWParams()
params.m = 16
collection = govector.govector_collection_create(
    "my_col", 128, govector.DISTANCE_COSINE,
    True, params, storage, error
)

# 获取数量
count = govector.govector_collection_count(collection)
print(f"Count: {count}")

# 清理
govector.govector_collection_free(collection)
govector.govector_storage_free(storage)
```

#### 2. 运行内置示例

```bash
# 确保在 capi 目录
cd govector/capi

# 运行简单示例
python3 examples/simple_example.py

# 运行完整演示
python3 examples/demo.py
```

---

## ❓ 故障排除

### 问题 1: `swig: command not found`

**解决方案**：
```bash
# macOS
brew install swig

# Ubuntu
sudo apt-get install swig
```

### 问题 2: `Python.h: No such file or directory`

**解决方案**：
```bash
# macOS
brew install python3

# Ubuntu
sudo apt-get install python3-dev
```

### 问题 3: `_govector.so: undefined symbol`

**解决方案**：
```bash
# 清理并重新构建
cd govector/capi
./build.sh clean
./build.sh python
```

### 问题 4: `import govector: module not found`

**解决方案**：
```bash
# 确保在正确的目录
cd govector/capi

# 或者添加路径
export PYTHONPATH=$PYTHONPATH:$(pwd)
```

### 问题 5: 段错误 (Segmentation Fault)

**解决方案**：
1. 确保先调用了 `govector_storage_new`
2. 确保最后调用了 `*_free` 函数释放资源
3. 检查错误返回值

---

## 📚 更多信息

- **快速入门**: [CGO_QUICKSTART.md](../docs/CGO_QUICKSTART.md)
- **架构设计**: [CGO_ISOLATION_DESIGN.md](../docs/CGO_ISOLATION_DESIGN.md)
- **详细示例**: [SWIG_EXAMPLES_AND_BEST_PRACTICES.md](../docs/SWIG_EXAMPLES_AND_BEST_PRACTICES.md)
- **集成指南**: [SWIG_INTEGRATION_GUIDE.md](../docs/SWIG_INTEGRATION_GUIDE.md)

---

## 🎯 下一步

构建成功后，您可以：

1. ✅ 运行示例程序：`python3 examples/simple_example.py`
2. ✅ 修改示例代码，尝试不同的功能
3. ✅ 阅读更多文档了解高级用法
4. ✅ 将 GoVector 集成到您的 Python 项目中

---

## 🆘 获取帮助

如果遇到问题：

1. 查看本文档的"故障排除"部分
2. 阅读相关文档
3. 检查错误日志
4. 提交 Issue 到 GitHub 仓库

---

**版本**: 1.0  
**更新日期**: 2026-03-31  
**维护者**: Raya Info co.,Ltd
