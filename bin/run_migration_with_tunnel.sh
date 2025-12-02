#!/usr/bin/env bash

# Get the script's directory and project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

SSH_HOST="a1.i247.com"
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

# Extract DB credentials from project root
USER=$(grep DB_USER "${PROJECT_ROOT}/.env.ec2-credentials" | cut -d '=' -f 2)
PASSWORD=$(grep DB_PASSWORD "${PROJECT_ROOT}/.env.ec2-credentials" | cut -d '=' -f 2)
HOST="127.0.0.1"  # Always use 127.0.0.1 for tunnel
PORT=${LOCAL_PORT}  # Use the tunnel port
NAME=$(grep DB_NAME "${PROJECT_ROOT}/.env.ec2-credentials" | cut -d '=' -f 2)

if [ -z "$USER" ] || [ -z "$NAME" ]; then
    echo "Error: Missing required variables in .env.ec2-credentials"
    pkill -f "ssh.*${LOCAL_PORT}:${REMOTE_HOST}:${REMOTE_PORT}"
    exit 1
fi

# Run migration
echo "🚀 Running migration: $migration_file"
if [ -z "$PASSWORD" ]; then
    mysql --force -u $USER -h $HOST -P $PORT $NAME < $migration_file
else
    mysql --force -u $USER -p$PASSWORD -h $HOST -P $PORT $NAME < $migration_file
fi

MIGRATION_EXIT_CODE=$?

# Close SSH tunnel
echo "🔒 Closing SSH tunnel..."
pkill -f "ssh.*${LOCAL_PORT}:${REMOTE_HOST}:${REMOTE_PORT}"

exit $MIGRATION_EXIT_CODE