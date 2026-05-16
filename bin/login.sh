#!/usr/bin/env bash

if [ $# -ne 1 ]
then
  echo "Usage: $0 <host>"
  exit 1
fi


# Check if the .env.ec2-credentials file exists
if [ ! -f .env.ec2-credentials ]; then
    echo "Error: .env.ec2-credentials file is missing"
    exit 1
fi

hostIn=$1

SSH_KEY=$(cat .env.ec2-credentials | grep SSH_KEY | cut -d '=' -f 2)
USER=$(cat .env.ec2-credentials | grep USER | cut -d '=' -f 2)
HOST=$(cat .env.ec2-credentials | grep '^HOST=' | cut -d '=' -f 2)
HOST_P1=$(cat .env.ec2-credentials | grep '^HOST_P1=' | cut -d '=' -f 2)
HOST_P2=$(cat .env.ec2-credentials | grep '^HOST_P2=' | cut -d '=' -f 2)
HOST1=$(cat .env.ec2-credentials | grep '^HOST1=' | cut -d '=' -f 2)
HOST2=$(cat .env.ec2-credentials | grep '^HOST2=' | cut -d '=' -f 2)
HOST3=$(cat .env.ec2-credentials | grep '^HOST3=' | cut -d '=' -f 2)
HOST4=$(cat .env.ec2-credentials | grep '^HOST4=' | cut -d '=' -f 2)

case $hostIn in
  "t1") HOST=$HOST1
        echo "detecting $HOST" ;;
  "t2") HOST=$HOST2
        echo "detecting $HOST" ;;
  "t3") HOST=$HOST3
        echo "detecting $HOST" ;;
  "t4") HOST=$HOST4
        echo "detecting $HOST" ;;
  "p1") HOST=$HOST_P1
        echo "detecting $HOST" ;;
  "p2") HOST=$HOST_P2
        echo "detecting $HOST" ;;
  *) echo "host not found $hostIn"
     exit 1 ;;
esac

echo "logging in to $HOST..."

if [ -z "$SSH_KEY" ] || [ -z "$USER" ] || [ -z "$HOST" ]; then
    echo "Error: .env.ec2-credentials file is missing one or more required variables [SSH_KEY, USER, HOST]"
    exit 1
fi

echo "Logging in to $HOST as $USER"
ssh -i $SSH_KEY $USER@$HOST
