#!/usr/bin/env bash

# Check if the .env.ec2-credentials file exists
if [ ! -f .env.ec2-credentials ]; then
    echo "Error: .env.ec2-credentials file is missing"
    exit 1
fi

hostIn=$1

DEST_DIR="/apps/math"

# Extract variables and trim whitespace
SSH_KEY=$(grep '^SSH_KEY=' .env.ec2-credentials | cut -d '=' -f 2 | tr -d '[:space:]')
USER=$(grep '^USER=' .env.ec2-credentials | cut -d '=' -f 2 | tr -d '[:space:]')
HOST=$(grep '^HOST=' .env.ec2-credentials | cut -d '=' -f 2 | tr -d '[:space:]')

echo "deploy to $HOST..."

if [ -z "$SSH_KEY" ] || [ -z "$USER" ] || [ -z "$HOST" ]; then
    echo "Error: .env.ec2-credentials file is missing one or more required variables [SSH_KEY, USER, HOST]"
    exit 1
fi

GIT_BRANCH=$(git rev-parse --abbrev-ref HEAD)

# Require user confirmation
read -p "deploy branch=$GIT_BRANCH to ec2=${USER}@${HOST} [y/n]? " ansYN

case "${ansYN:-N}" in
  "y"|"Y")
    echo "* Deploying..."
    echo "* git branch=$GIT_BRANCH"
    echo "* ec2=${USER}@${HOST}"
    ;;
  *)
    echo "Deploy cancelled. exiting..."
    exit 1 ;;
esac

echo "* Running pre-deploy scripts..."
ssh -i "$SSH_KEY" "${USER}@${HOST}" 'for i in '"\"$DEST_DIR\""'/pre-deploy/*.sh; do [ -f "$i" ] && { echo "Running $i..." && bash "$i"; }; done'
if [ $? -eq 0 ]; then
    echo "* Pre-Deployed successfully!"
fi

echo "* Syncing remote files..."
# First: upload .env.ec2 → .env on the server (this overwrites any old .env)
rsync -avz --progress -e "ssh -i $SSH_KEY" \
    .env.ec2 "${USER}@${HOST}:${DEST_DIR}/.env"

if [ $? -ne 0 ]; then
    echo "Error: Failed to deploy .env file"
    exit 1
fi

echo "* .env.ec2 successfully deployed as .env on the server"

# Then: sync all other files/folders as before
rsync -avz -R --progress --no-perms -e "ssh -i $SSH_KEY" \
    hmac.key \
    migrations/ \
    keys/ \
    pre-deploy/ \
    post-deploy/ \
    dist/server \
    "${USER}@${HOST}:${DEST_DIR}/"

if [ $? -ne 0 ]; then
    echo "Error: Failed to deploy files"
    exit 1
fi

echo "* Running post-deploy scripts..."
ssh -i "$SSH_KEY" "${USER}@${HOST}" 'for i in '"\"$DEST_DIR\""'/post-deploy/*.sh; do [ -f "$i" ] && { echo "Running $i..." && bash "$i"; }; done'
if [ $? -eq 0 ]; then
    echo "* Post-Deployed successfully!"
fi

echo "* Done!"