#!/bin/bash

if [[ $# -eq 0 ]] ; then
    echo "Usage: $0 FAHC_VIEWER_URL"
    exit 1
fi

URL=$1

echo -e "Timestamp\tPod\tNode\tProgress\tState"

while true;do
    sleep 1
    curl --silent $URL | jq -r '.[] | .sc_hostname +"\t"+ .sc_nodename +"\t"+ .percentdone +"\t"+ .state' | sed -e "s/^/$(date '+%Y-%m-%d %H:%M:%S')\t/" &
done
