#!/usr/bin/env bash

# Check if migration name is provided
if [ -z "$1" ]; then
    echo "❌ Error: Migration name is required."
    echo "Usage: $0 <migration_name>"
    exit 1
fi

# Generate a timestamp for the migration version
VERSION=$(date +%Y%m%d%H%M%S)
NAME=$1
UP_DIR="migrations/up"
DOWN_DIR="migrations/down"

# Create directories if they don't exist
mkdir -p $UP_DIR
mkdir -p $DOWN_DIR

# Define file paths for up and down migrations
UP_FILE="${UP_DIR}/${VERSION}_${NAME}.sql"
DOWN_FILE="${DOWN_DIR}/${VERSION}_${NAME}.sql"

# Create empty migration files
touch $UP_FILE
touch $DOWN_FILE

echo "✨ Created migration files:"
echo "   -> ${UP_FILE}"
echo "   -> ${DOWN_FILE}"