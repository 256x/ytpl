#!/bin/bash
set -e  # エラーが発生したらスクリプトを終了
set -x  # デバッグ情報を表示

# カレントディレクトリを保存
PROJECT_ROOT=$(pwd)

# バージョン情報を取得
VERSION=$(grep 'const Version' cmd/root.go | awk -F'"' '{print $2}')
OUTPUT_DIR="${PROJECT_ROOT}/release"

# 出力ディレクトリの作成
mkdir -p ${OUTPUT_DIR}

echo "Building ytpl version: ${VERSION}"

# ビルド対象のプラットフォームを定義（mpvプレーヤーを使用しているためLinuxのみ）
PLATFORMS=(
    "linux/amd64"
    "linux/arm64"
)

# 各プラットフォーム向けにビルド
for platform in "${PLATFORMS[@]}"; do
    # プラットフォームをOSとアーキテクチャに分割
    OS=$(echo ${platform} | cut -d'/' -f1)
    ARCH=$(echo ${platform} | cut -d'/' -f2)
    
    # 出力ファイル名を設定
    OUTPUT_NAME="ytpl-${VERSION}-${OS}-${ARCH}"
    if [ "$OS" = "windows" ]; then
        OUTPUT_NAME="${OUTPUT_NAME}.exe"
    fi
    
    echo "\n=== Building for ${OS}/${ARCH} ==="
    echo "Output: ${OUTPUT_DIR}/${OUTPUT_NAME}"
    
    # ビルドコマンドを実行
    cd "${PROJECT_ROOT}"  # 必ずプロジェクトルートに戻る
    set +e  # エラーを一時的に無効化
    env CGO_ENABLED=0 GOOS=${OS} GOARCH=${ARCH} go build -v -ldflags="-s -w" -o "${OUTPUT_DIR}/${OUTPUT_NAME}" .
    BUILD_STATUS=$?
    set -e  # エラー検出を再度有効化
    
    # ビルドが成功したか確認
    if [ ${BUILD_STATUS} -ne 0 ]; then
        echo "!! Error building for ${OS}/${ARCH} (status: ${BUILD_STATUS}) !!"
        continue  # エラーが発生しても次のプラットフォームをビルド
    fi
    
    # 圧縮
    echo "Compressing..."
    cd "${OUTPUT_DIR}"  # 出力ディレクトリに移動
    
    if [ "$OS" = "windows" ]; then
        zip "${OUTPUT_NAME}.zip" "${OUTPUT_NAME}" && \
        rm -f "${OUTPUT_NAME}" && \
        echo "Created: ${OUTPUT_DIR}/${OUTPUT_NAME}.zip"
    else
        tar -czf "${OUTPUT_NAME}.tar.gz" "${OUTPUT_NAME}" && \
        rm -f "${OUTPUT_NAME}" && \
        echo "Created: ${OUTPUT_DIR}/${OUTPUT_NAME}.tar.gz"
    fi
    
    cd - > /dev/null  # 元のディレクトリに戻る
done

echo "\n=== Build Summary ==="
echo "Build completed. Output files in ${OUTPUT_DIR}:"
ls -lh ${OUTPUT_DIR}/ | grep -v "^total"

echo "\nTo create a release, run the following commands:"
echo "git tag -a v${VERSION} -m \"Release ${VERSION}\""
echo "git push origin v${VERSION}"
