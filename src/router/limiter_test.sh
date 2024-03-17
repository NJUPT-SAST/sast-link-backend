#!/bin/sh

current_seconds=`date +%s`
for i in {1..3}; do
    for j in {1..60}; do
        # Get http code
        curl -w "\n%{http_code}\n" -s 'http://127.0.0.1:8080/api/v1/user/info' | grep '200'
        #echo ""
        sleep 1
    done
done

echo 'time used:' $((`date +%s` - $current_seconds))
