#!/usr/bin/env bash

# Check if the .env.ec2-credentials file exists
if [ ! -f .env.ec2-credentials ]; then
    echo "Error: .env.ec2-credentials file is missing"
    exit 1
fi

hostIn=$1

SSH_KEY=$(cat .env.ec2-credentials | grep SSH_KEY | cut -d '=' -f 2)
USER=$(cat .env.ec2-credentials | grep USER | cut -d '=' -f 2)
HOST=$(cat .env.ec2-credentials | grep '^HOST=' | cut -d '=' -f 2)


echo "logging in to $HOST..."

if [ -z "$SSH_KEY" ] || [ -z "$USER" ] || [ -z "$HOST" ]; then
    echo "Error: .env.ec2-credentials file is missing one or more required variables [SSH_KEY, USER, HOST]"
    exit 1
fi

ssh -t -i $SSH_KEY $USER@$HOST '
  set -a; source /apps/math/.env; set +a
  mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASS" "$DB_NAME"
'