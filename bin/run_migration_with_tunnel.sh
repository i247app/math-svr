#!/usr/bin/env bash

# Get the script's directory and project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

SSH_HOST="m1.i247.com"
SSH_USER="mot"
SSH_KEY="${PROJECT_ROOT}/keys/kanon.pem"
LOCAL_PORT=3307  # Use different port to avoid conflicts
REMOTE_HOST="localhost"  # The database host as seen from the bastion
REMOTE_PORT=3306

# Verify SSH key exists
if [ ! -f "$SSH_KEY" ]; then
    echo "❌ SSH key not found at: $SSH_KEY"
    exit 1
fi

# Start SSH tunnel in background
echo "🔐 Establishing SSH tunnel..."
ssh -f -N -L ${LOCAL_PORT}:${REMOTE_HOST}:${REMOTE_PORT} \
    -i ${SSH_KEY} ${SSH_USER}@${SSH_HOST}

# Check if tunnel is established
sleep 2
if ! pgrep -f "ssh.*${LOCAL_PORT}:${REMOTE_HOST}:${REMOTE_PORT}" > /dev/null; then
    echo "❌ Failed to establish SSH tunnel"
    exit 1
fi

echo "✅ SSH tunnel established on localhost:${LOCAL_PORT}"

# Run the migration
migration_file=$1
if [ -z "$migration_file" ]; then
    echo "Usage: $0 <migration_file>"
    # Kill the tunnel before exiting
    pkill -f "ssh.*${LOCAL_PORT}:${REMOTE_HOST}:${REMOTE_PORT}"
    exit 1
fi

# Extract DB credentials from project root and trim whitespace
USER=$(grep "^DB_USER=" "${PROJECT_ROOT}/.env.ec2-credentials-db" | cut -d '=' -f 2- | tr -d '[:space:]')
PASSWORD=$(grep "^DB_PASSWORD=" "${PROJECT_ROOT}/.env.ec2-credentials-db" | cut -d '=' -f 2- | tr -d '[:space:]')
HOST="127.0.0.1"  # Always use 127.0.0.1 for tunnel
PORT=${LOCAL_PORT}  # Use the tunnel port
NAME=$(grep "^DB_NAME=" "${PROJECT_ROOT}/.env.ec2-credentials-db" | cut -d '=' -f 2- | tr -d '[:space:]')

if [ -z "$USER" ] || [ -z "$NAME" ]; then
    echo "❌ Error: Missing required variables in .env.ec2-credentials-db"
    echo "   Required: DB_USER, DB_NAME"
    pkill -f "ssh.*${LOCAL_PORT}:${REMOTE_HOST}:${REMOTE_PORT}"
    exit 1
fi

# Debug: Show what credentials are being used (without showing password)
echo "📋 Using credentials:"
echo "   User: $USER"
echo "   Database: $NAME"
echo "   Host: $HOST:$PORT"

# Test database connection before running migration
echo "🔍 Testing database connection..."
if [ -z "$PASSWORD" ]; then
    TEST_RESULT=$(mysql -u "$USER" -h "$HOST" -P "$PORT" -e "SELECT 1" 2>&1)
    TEST_EXIT=$?
else
    TEST_RESULT=$(mysql -u "$USER" -p"$PASSWORD" -h "$HOST" -P "$PORT" -e "SELECT 1" 2>&1)
    TEST_EXIT=$?
fi

if [ $TEST_EXIT -ne 0 ]; then
    echo "❌ Database connection failed!"
    echo ""
    echo "Error details:"
    echo "$TEST_RESULT"
    echo ""
    echo "Please verify:"
    echo "1. Credentials in .env.ec2-credentials-db are correct"
    echo "2. MySQL user '$USER'@'localhost' exists on remote server"
    echo "3. User has correct password and permissions"
    echo ""
    echo "To fix on remote server, SSH in and run:"
    echo "   ssh -i ${SSH_KEY} ${SSH_USER}@${SSH_HOST}"
    echo "   sudo mysql"
    echo "   ALTER USER '$USER'@'localhost' IDENTIFIED BY 'your_password';"
    echo "   GRANT ALL PRIVILEGES ON $NAME.* TO '$USER'@'localhost';"
    echo "   FLUSH PRIVILEGES;"
    pkill -f "ssh.*${LOCAL_PORT}:${REMOTE_HOST}:${REMOTE_PORT}"
    exit 1
fi

echo "✅ Database connection successful"

# Run migration
echo "🚀 Running migration: $migration_file"
if [ -z "$PASSWORD" ]; then
    mysql --force -u "$USER" -h "$HOST" -P "$PORT" "$NAME" < "$migration_file"
else
    mysql --force -u "$USER" -p"$PASSWORD" -h "$HOST" -P "$PORT" "$NAME" < "$migration_file"
fi

MIGRATION_EXIT_CODE=$?

# Close SSH tunnel
echo "🔒 Closing SSH tunnel..."
pkill -f "ssh.*${LOCAL_PORT}:${REMOTE_HOST}:${REMOTE_PORT}"

exit $MIGRATION_EXIT_CODE