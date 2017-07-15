#!/bin/bash -e

go_file=${1?Missing go file}

echo "$(date) Building $go_file";
GOARM=7 GOARCH=arm GOOS=linux go build "$go_file";

service=sensors.service

echo "$(date) Stopping $service service";
set +e
ssh -n -f pi "sh -c 'sudo systemctl stop $service'"
set -e

bin_file="${1%%.*}"
echo "$(date) Transferring $bin_file";
scp "$bin_file" pi:bin/

echo "$(date) Restarting $service";
ssh -n -f pi "sh -c 'sudo mv ~pi/bin/sensors /usr/local/bin && sudo systemctl start $service'"

echo "$(date) Done"
