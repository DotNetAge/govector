#!/bin/bash

# build_release.sh - 自动编译所有平台的 Release 压缩包并更新 Homebrew Formula

VERSION=${1:-"v0.1.3"}
APP_NAME="govector"
OUTPUT_DIR="dist"
BREW_FORMULA="scripts/release/govector.rb"

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

# 用于存储 SHA256 信息的临时文件
CHECKSUM_FILE="$OUTPUT_DIR/checksums.txt"
touch "$CHECKSUM_FILE"

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
    FILE_NAME=""
    if [ "$GOOS" = "windows" ]; then
        FILE_NAME="${ARCHIVE_DIR}.zip"
        zip -r -q "$FILE_NAME" "$ARCHIVE_DIR"
    else
        FILE_NAME="${ARCHIVE_DIR}.tar.gz"
        tar -czf "$FILE_NAME" "$ARCHIVE_DIR"
    fi
    
    # 计算校验和并存入临时文件
    if command -v shasum >/dev/null 2>&1; then
        SHA=$(shasum -a 256 "$FILE_NAME" | cut -d' ' -f1)
    else
        SHA=$(sha256sum "$FILE_NAME" | cut -d' ' -f1)
    fi
    echo "$GOOS|$GOARCH|$FILE_NAME|$SHA" >> checksums.txt
    
    # 清理临时目录
    rm -rf "$ARCHIVE_DIR"
    popd > /dev/null

    echo "✅ 成功打包: $ARCHIVE_DIR"
done

# 更新 Homebrew Formula
if [ -f "$BREW_FORMULA" ]; then
    echo "📝 正在更新 Homebrew Formula: $BREW_FORMULA ..."
    
    # 获取各个平台的 SHA
    SHA_DARWIN_ARM64=$(grep "darwin|arm64" "$CHECKSUM_FILE" | cut -d'|' -f4)
    SHA_DARWIN_AMD64=$(grep "darwin|amd64" "$CHECKSUM_FILE" | cut -d'|' -f4)
    SHA_LINUX_AMD64=$(grep "linux|amd64" "$CHECKSUM_FILE" | cut -d'|' -f4)

    # 使用 sed 更新版本号和校验和
    # 注意：这里假设了 Formula 的结构，使用 MacOS 的 sed 语法兼容性处理
    if [[ "$OSTYPE" == "darwin"* ]]; then
        sed -i '' "s/version \".*\"/version \"${VERSION/v/}\"/" "$BREW_FORMULA"
        sed -i '' "s|/download/v.*/|/download/${VERSION}/|g" "$BREW_FORMULA"
        sed -i '' "s/govector_v.*_darwin_arm64.tar.gz/govector_${VERSION}_darwin_arm64.tar.gz/g" "$BREW_FORMULA"
        sed -i '' "s/govector_v.*_darwin_amd64.tar.gz/govector_${VERSION}_darwin_amd64.tar.gz/g" "$BREW_FORMULA"
        sed -i '' "s/govector_v.*_linux_amd64.tar.gz/govector_${VERSION}_linux_amd64.tar.gz/g" "$BREW_FORMULA"
        
        # 定向更新特定平台的 SHA
        # 这里通过行号定位可能更稳妥，或者利用上下文
        perl -i -pe "BEGIN{undef $/;} s/(if OS.mac\? && Hardware::CPU.arm\?.*?sha256 \").*?(\")/\${1}${SHA_DARWIN_ARM64}\${2}/s" "$BREW_FORMULA"
        perl -i -pe "BEGIN{undef $/;} s/(if OS.mac\? && Hardware::CPU.intel\?.*?sha256 \").*?(\")/\${1}${SHA_DARWIN_AMD64}\${2}/s" "$BREW_FORMULA"
        perl -i -pe "BEGIN{undef $/;} s/(if OS.linux\? && Hardware::CPU.intel\?.*?sha256 \").*?(\")/\${1}${SHA_LINUX_AMD64}\${2}/s" "$BREW_FORMULA"
    else
        sed -i "s/version \".*\"/version \"${VERSION/v/}\"/" "$BREW_FORMULA"
        sed -i "s|/download/v.*/|/download/${VERSION}/|g" "$BREW_FORMULA"
        sed -i "s/govector_v.*_darwin_arm64.tar.gz/govector_${VERSION}_darwin_arm64.tar.gz/g" "$BREW_FORMULA"
        sed -i "s/govector_v.*_darwin_amd64.tar.gz/govector_${VERSION}_darwin_amd64.tar.gz/g" "$BREW_FORMULA"
        sed -i "s/govector_v.*_linux_amd64.tar.gz/govector_${VERSION}_linux_amd64.tar.gz/g" "$BREW_FORMULA"
        
        perl -i -pe "BEGIN{undef $/;} s/(if OS.mac\? && Hardware::CPU.arm\?.*?sha256 \").*?(\")/\${1}${SHA_DARWIN_ARM64}\${2}/s" "$BREW_FORMULA"
        perl -i -pe "BEGIN{undef $/;} s/(if OS.mac\? && Hardware::CPU.intel\?.*?sha256 \").*?(\")/\${1}${SHA_DARWIN_AMD64}\${2}/s" "$BREW_FORMULA"
        perl -i -pe "BEGIN{undef $/;} s/(if OS.linux\? && Hardware::CPU.intel\?.*?sha256 \").*?(\")/\${1}${SHA_LINUX_AMD64}\${2}/s" "$BREW_FORMULA"
    fi
    echo "✅ Homebrew Formula 已同步更新！"
fi

echo ""
echo "🎉 所有平台编译完成！发布文件位于 $OUTPUT_DIR/ 目录："
echo "--------------------------------------------------------------------------------------------------------------------------------"
printf "| %-10s | %-10s | %-40s | %-64s |\n" "OS" "ARCH" "FILENAME" "SHA256 CHECKSUM"
echo "--------------------------------------------------------------------------------------------------------------------------------"

while IFS='|' read -r os arch file sha; do
    printf "| %-10s | %-10s | %-40s | %-64s |\n" "$os" "$arch" "$file" "$sha"
done < "$CHECKSUM_FILE"

echo "--------------------------------------------------------------------------------------------------------------------------------"
rm "$CHECKSUM_FILE"

ls -lh $OUTPUT_DIR
