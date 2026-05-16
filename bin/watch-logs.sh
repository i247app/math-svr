#!/usr/bin/env bash
SSH_KEY=$(cat .env.ec2-credentials | grep SSH_KEY | cut -d '=' -f 2)
USER=$(cat .env.ec2-credentials | grep USER | cut -d '=' -f 2)
HOST=$(cat .env.ec2-credentials | grep '^HOST=' | cut -d '=' -f 2)

if [ -z "$SSH_KEY" ] || [ -z "$USER" ] || [ -z "$HOST" ]; then
    echo "Error: .env.ec2-credentials file is missing one or more required variables [SSH_KEY, USER, HOST]"
    exit 1
fi

echo "Watching logs for $HOST as $USER"
ssh -i $SSH_KEY $USER@$HOST "tail -n 5000 -f /apps/monex/monexsvr.log"
