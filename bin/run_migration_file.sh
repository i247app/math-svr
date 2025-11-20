#!/usr/bin/env bash
migration_file=$1
if [ -z "$migration_file" ]; then
    echo "Usage: $0 <migration_file>"
    exit 1
fi

# Extract DB credentials from .env file
USER=$(cat .env | grep DB_USER | cut -d '=' -f 2)
PASSWORD=$(cat .env | grep DB_PASSWORD | cut -d '=' -f 2)
HOST=$(cat .env | grep DB_HOST | cut -d '=' -f 2)
PORT=$(cat .env | grep DB_PORT | cut -d '=' -f 2)
NAME=$(cat .env | grep DB_NAME | cut -d '=' -f 2)

if [ -z "$USER" ] || [ -z "$HOST" ] || [ -z "$PORT" ] || [ -z "$NAME" ]; then
    echo "Error: .env file is missing one or more required variables [DB_USER, DB_HOST, DB_PORT, DB_NAME]"
    exit 1
fi

if [ -z "$PASSWORD" ]; then
    mysql --force -u $USER -h $HOST -P $PORT $NAME < $migration_file
else
    # echo "mysql --force -u $USER -p$PASSWORD -h $HOST -P $PORT $NAME < $migration_file"
    mysql --force -u $USER -p$PASSWORD -h $HOST -P $PORT $NAME < $migration_file
fi