#!/usr/bin/env bash

# Check if the .env.ec2-credentials file exists
if [ ! -f .env.ec2-credentials ]; then
    echo "Error: .env.ec2-credentials file is missing"
    exit 1
fi


SSH_KEY=$(cat .env.ec2-credentials | grep SSH_KEY | cut -d '=' -f 2)
USER=$(cat .env.ec2-credentials | grep USER | cut -d '=' -f 2)
HOST=$(cat .env.ec2-credentials | grep '^HOST=' | cut -d '=' -f 2)

if [ -z "$SSH_KEY" ] || [ -z "$USER" ] || [ -z "$HOST" ]; then
    echo "Error: .env.ec2-credentials file is missing one or more required variables [SSH_KEY, USER, HOST]"
    exit 1
fi

# 0. Define constants

SVR_DIR="/var/www/math_svr"

# 1. Create Go server web directory:

ssh -i "$SSH_KEY" "$USER@$HOST" \
"
cd /var/www/ && \
sudo mkdir math_svr && \
sudo chgrp -R math math_svr/ && \
sudo chmod 775 -R math_svr/
"

# 2. Upload JWT handshake keys:

rsync -e "ssh -i $SSH_KEY" resources/certs/*.pem "$USER@$HOST":"$SVR_DIR/keys"

# 3. Create .env file:

# First, upload the file
scp -i "$SSH_KEY" ./.env "$USER@$HOST":"$SVR_DIR/.env"
# Then, open it in vim if you need to make further edits
ssh -t -i "$SSH_KEY" "$USER@$HOST" "sudo vim $SVR_DIR/.env"

# 4. Create 'ez' utility command

cat << 'EOF' | ssh -i "$SSH_KEY" "$USER@$HOST" "sudo tee /usr/local/bin/ez >/dev/null && sudo chmod +x /usr/local/bin/ez"
#!/usr/bin/env bash

# ez - TNMonex EZSend server control script

# Define paths
SVR_DIR="$SVR_DIR"
PID_FILE="$SVR_DIR/mathsvr.pid"
LOG_FILE="$SVR_DIR/mathsvr.log"

# Usage info
usage() {
    echo "Usage: $0 {start|stop|restart}"
    echo "  start   - Start the mathsvr server"
    echo "  stop    - Stop the mathsvr server"
    echo "  restart - Restart the mathsvr server"
    exit 1
}

# Start the server
start() {
    if [ -f "$PID_FILE" ] && kill -0 "$(cat "$PID_FILE")" 2>/dev/null; then
        echo "Server is already running with PID $(cat "$PID_FILE")"
        exit 1
    fi
    cd "$SVR_DIR" || {
        echo "Error: Cannot cd to $SVR_DIR"
        exit 1
    }
    sudo ./dist/mathsvr > "$LOG_FILE" 2>&1 &
    echo $! > "$PID_FILE"
    echo "Server started with PID $(cat "$PID_FILE")"
}

# Stop the server
stop() {
    if [ ! -f "$PID_FILE" ]; then
        echo "No PID file found - server may not be running"
        exit 1
    fi
    PID=$(cat "$PID_FILE" 2>/dev/null)
    if [ -z "$PID" ]; then
        echo "PID file is empty - server may not be running"
        exit 1
    fi
    kill "$PID" 2>/dev/null
    if [ $? -eq 0 ]; then
        rm -f "$PID_FILE"
        echo "Server stopped (PID $PID)"
    else
        echo "Failed to stop server - process $PID may not exist"
        exit 1
    fi
}

# Restart the server
restart() {
    stop
    sleep 1  # Give it a moment to fully stop
    start
}

# Check for command argument
if [ $# -ne 1 ]; then
    usage
fi

# Handle commands
case "$1" in
    start)
        start
        ;;
    stop)
        stop
        ;;
    restart)
        restart
        ;;
    *)
        usage
        ;;
esac

exit 0
EOF