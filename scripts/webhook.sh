#!/bin/bash

usage() {
    echo "Usage: $0 [-u <workflow_url>] [-w <webhook_url>] [-s <service_name>] [-c <commit_user>] [-m <commit_url>] [-f <flow_status>]" 1>&2
    exit 1
}

while getopts "hu:w:s:c:m:f:" arg; do
  case $arg in
    h)
      usage
      ;;
    u)
      workflow_url="$OPTARG"
      ;;
    w)
      webhook_url="$OPTARG"
      ;;
    s)
      service_name="$OPTARG"
      ;;
    c)
      commit_user="$OPTARG"
      ;;
    m)
      commit_url="$OPTARG"
      ;;
    f)
      flow_status="$OPTARG"
      ;;
    ?)
      usage
      ;;
  esac
done

if [[ -z "$workflow_url" || -z "$webhook_url" || -z "$service_name" || -z "$commit_user" || -z "$commit_url" || -z "$flow_status" ]]; then
  usage
  exit 1
fi

status="success"
if systemctl is-active --quiet "$service_name"; then
  echo "service $service_name is running"
  status="success"
else
  echo "service $service_name is not running"
  status="failed"
fi

# GMT time
cur_date=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

card_msg='{
    "msg_type": "interactive",
    "card": {
        "elements": [{
                "tag": "div",
                "text": {
                "content": "[Commit]('$commit_url') by ['$commit_user'](https://github.com/NJUPT-SAST/sast-link-backend/commits?author='$commit_user')",
                        "tag": "lark_md"
                }
        }, {
                "actions": [{
                        "tag": "button",
                        "text": {
                                "content": "see details",
                                "tag": "lark_md"
                        },
                        "url": "'$workflow_url'",
                        "type": "default",
                        "value": {}
                }],
                "tag": "action"
        }],
        "header": {
                "title": {
                        "content": "'$service_name' workflows run '$flow_status' at '$cur_date'\nService status: '$status'",
                        "tag": "plain_text"
                }
        }
    }
}'

curl -X POST -H "Content-Type: application/json" \
     -d "$card_msg" \
     "$webhook_url"
