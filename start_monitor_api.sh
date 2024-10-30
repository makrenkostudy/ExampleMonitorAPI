#!/bin/bash
start() {
    cd /opt/monitor_api || { echo "Failed to change directory to /opt/monitor_api"; exit 1; }
    /opt/go/bin/go run counter.go --createdb > /opt/monitor_api/monitor_api.log 2>&1
    echo "Starting the API server..." >> /opt/monitor_api/monitor_api.log 
    /opt/go/bin/go run counter.go --start >> /opt/monitor_api/monitor_api.log 2>&1 &
}

stop() {
    kill -s SIGKILL `ps -ef | grep -i "go run counter.go --start" | awk '{print $2;}'`
}

case $1 in
    start|stop) "$1" ;;
esac
echo ""

