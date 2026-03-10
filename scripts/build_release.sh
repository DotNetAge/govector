#!/bin/bash

# build_release.sh - 自动编译所有平台的 Release 压缩包

VERSION=${1:-"v0.1.0"}
APP_NAME="govector"
OUTPUT_DIR="dist"

echo "🚀 开始构建 GoVector $VERSION 发布包..."

rm -rf $OUTPUT_DIR
mkdir -p $OUTPUT_DIR

# 需要编译的平台矩阵
PLATFORMS=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
)

for PLATFORM in "${PLATFORMS[@]}"; do
    # 分割 OS 和 ARCH
    GOOS=${PLATFORM%/*}
    GOARCH=${PLATFORM#*/}
    
    # 定义输出二进制文件名
    OUTPUT_NAME=$APP_NAME
    if [ "$GOOS" = "windows" ]; then
        OUTPUT_NAME+='.exe'
    fi

    echo "⚙️  编译 $GOOS/$GOARCH ..."
    
    # 创建临时打包目录
    ARCHIVE_DIR="${APP_NAME}_${VERSION}_${GOOS}_${GOARCH}"
    mkdir -p "$OUTPUT_DIR/$ARCHIVE_DIR"
    
    # 执行跨平台编译
    env CGO_ENABLED=0 GOOS=$GOOS GOARCH=$GOARCH go build \
        -ldflags "-s -w" \
        -o "$OUTPUT_DIR/$ARCHIVE_DIR/$OUTPUT_NAME" \
        ./cmd/govector-server
    
    # 拷贝 README 和 LICENSE 到压缩包
    cp README.md "$OUTPUT_DIR/$ARCHIVE_DIR/"
    cp LICENSE "$OUTPUT_DIR/$ARCHIVE_DIR/"

    # 打包压缩
    pushd $OUTPUT_DIR > /dev/null
    if [ "$GOOS" = "windows" ]; then
        zip -r -q "${ARCHIVE_DIR}.zip" "$ARCHIVE_DIR"
    else
        tar -czf "${ARCHIVE_DIR}.tar.gz" "$ARCHIVE_DIR"
    fi
    # 清理临时目录
    rm -rf "$ARCHIVE_DIR"
    popd > /dev/null

    echo "✅ 成功打包: $ARCHIVE_DIR"
done

echo "🎉 所有平台编译完成！发布文件位于 $OUTPUT_DIR/ 目录："
ls -lh $OUTPUT_DIR
