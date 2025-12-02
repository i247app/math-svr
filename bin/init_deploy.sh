#!/usr/bin/env bash

# Check if the .env.ec2-credentials file exists
if [ ! -f .env.ec2-credentials ]; then
    echo "Error: .env.ec2-credentials file is missing"
    exit 1
fi

# Extract variables and trim whitespace
SSH_KEY=$(grep '^SSH_KEY=' .env.ec2-credentials | cut -d '=' -f 2 | tr -d '[:space:]')
USER=$(grep '^USER=' .env.ec2-credentials | cut -d '=' -f 2 | tr -d '[:space:]')
HOST=$(grep '^HOST=' .env.ec2-credentials | cut -d '=' -f 2 | tr -d '[:space:]')

if [ -z "$SSH_KEY" ] || [ -z "$USER" ] || [ -z "$HOST" ]; then
    echo "Error: .env.ec2-credentials file is missing one or more required variables [SSH_KEY, USER, HOST]"
    exit 1
fi

DEST_DIR="/apps/math"

echo "🚀 Initializing deployment environment on ${USER}@${HOST}..."

# Create the application directory
echo "📁 Creating application directory: $DEST_DIR"
ssh -i "$SSH_KEY" "${USER}@${HOST}" "sudo mkdir -p $DEST_DIR && sudo chown -R $USER:$USER $DEST_DIR"

if [ $? -ne 0 ]; then
    echo "❌ Failed to create application directory"
    exit 1
fi

# Create subdirectories
echo "📁 Creating subdirectories..."
ssh -i "$SSH_KEY" "${USER}@${HOST}" "mkdir -p $DEST_DIR/{pre-deploy,post-deploy,migrations,keys,dist}"

if [ $? -ne 0 ]; then
    echo "❌ Failed to create subdirectories"
    exit 1
fi

# Install dependencies (if needed)
echo "📦 Installing system dependencies..."
ssh -i "$SSH_KEY" "${USER}@${HOST}" << 'ENDSSH'
# Update package list
sudo apt-get update

# Install MySQL client (if not already installed)
if ! command -v mysql &> /dev/null; then
    echo "Installing MySQL client..."
    sudo apt-get install -y mysql-client
fi

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "⚠️  Go is not installed. You may need to install it manually."
fi

echo "✅ System dependencies checked"
ENDSSH

if [ $? -ne 0 ]; then
    echo "⚠️  Warning: Some dependencies may not have installed correctly"
fi

echo "✅ Initialization complete!"
echo ""
echo "Next steps:"
echo "  1. Run: make deploy-ec2-remote"
echo "  2. SSH to server and start the application"