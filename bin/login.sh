#!/usr/bin/env bash

# if [ $# -ne 1 ]
# then
#   echo "Usage: $0 <host>"
#   exit 1
# fi


# Check if the .env.ec2-credentials file exists
if [ ! -f .env.ec2-credentials ]; then
    echo "Error: .env.ec2-credentials file is missing"
    exit 1
fi

SSH_KEY=$(cat .env.ec2-credentials | grep SSH_KEY | cut -d '=' -f 2)
USER=$(cat .env.ec2-credentials | grep USER | cut -d '=' -f 2)
HOST=$(cat .env.ec2-credentials | grep HOST | cut -d '=' -f 2)

echo "Logging in to $HOST as $USER"
ssh -i $SSH_KEY $USER@$HOST