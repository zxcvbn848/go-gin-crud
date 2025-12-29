#!/bin/bash

# Swagger 文檔生成腳本

echo "🚀 開始生成 Swagger API 文檔..."

# 獲取 GOPATH
GOPATH=$(go env GOPATH)
GOBIN="${GOPATH}/bin"

# 將 GOPATH/bin 添加到 PATH（如果不在 PATH 中）
if [[ ":$PATH:" != *":${GOBIN}:"* ]]; then
    export PATH="${GOBIN}:${PATH}"
fi

# 檢查 swag 是否已安裝
SWAG_CMD="swag"
if ! command -v swag &> /dev/null; then
    echo "❌ swag 未安裝，正在安裝..."
    go install github.com/swaggo/swag/cmd/swag@latest
    if [ $? -ne 0 ]; then
        echo "❌ swag 安裝失敗，請手動執行: go install github.com/swaggo/swag/cmd/swag@latest"
        exit 1
    fi
    
    # 安裝後再次檢查
    if ! command -v swag &> /dev/null; then
        # 嘗試使用完整路徑
        if [ -f "${GOBIN}/swag" ]; then
            SWAG_CMD="${GOBIN}/swag"
            echo "✅ 使用完整路徑: ${SWAG_CMD}"
        else
            echo "❌ swag 安裝後仍無法找到，請確認 GOPATH/bin 在 PATH 中"
            echo "💡 解決方法："
            echo "   export PATH=\$PATH:\$(go env GOPATH)/bin"
            echo "   或將以下內容添加到 ~/.zshrc 或 ~/.bashrc:"
            echo "   export PATH=\$PATH:\$(go env GOPATH)/bin"
            exit 1
        fi
    else
        echo "✅ swag 安裝成功"
    fi
fi

# 生成文檔
${SWAG_CMD} init -g cmd/main.go -o docs --parseDependency --parseInternal

if [ $? -eq 0 ]; then
    echo "✅ Swagger 文檔生成成功！"
    echo ""
    echo "📄 文檔位置:"
    echo "   - docs/swagger.json"
    echo "   - docs/swagger.yaml"
    echo ""
    echo "🌐 訪問地址:"
    echo "   - Swagger UI: http://localhost:8080/swagger/index.html"
    echo "   - Swagger JSON: http://localhost:8080/swagger/doc.json"
    echo ""
    echo "💡 提示: 啟動服務後即可訪問 Swagger UI"
else
    echo "❌ Swagger 文檔生成失敗！"
    echo "💡 請檢查："
    echo "   1. 是否在專案根目錄執行"
    echo "   2. cmd/main.go 是否有正確的 Swagger 註釋"
    echo "   3. Controller 中的註釋格式是否正確"
    exit 1
fi

