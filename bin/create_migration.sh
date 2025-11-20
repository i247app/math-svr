#!/usr/bin/env bash

# Kiểm tra xem tên migration đã được cung cấp chưa
if [ -z "$1" ]; then
    echo "❌ Error: Migration name is required."
    echo "Usage: $0 <migration_name>"
    exit 1
fi

# Tạo version dựa trên timestamp (YYYYMMDDHHMMSS)
VERSION=$(date +%Y%m%d%H%M%S)
NAME=$1
UP_DIR="migrations/up"
DOWN_DIR="migrations/down"

# Tạo các thư mục up/down nếu chúng chưa tồn tại
mkdir -p $UP_DIR
mkdir -p $DOWN_DIR

# Tạo đường dẫn file đầy đủ
UP_FILE="${UP_DIR}/${VERSION}_${NAME}.sql"
DOWN_FILE="${DOWN_DIR}/${VERSION}_${NAME}.sql"

# Tạo các file rỗng
touch $UP_FILE
touch $DOWN_FILE

echo "✨ Created migration files:"
echo "   -> ${UP_FILE}"
echo "   -> ${DOWN_FILE}"