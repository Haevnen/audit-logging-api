#!/bin/sh
set -e

echo "Creating logs index with mapping..."

curl -s -X PUT "http://opensearch:9200/logs" \
  -H 'Content-Type: application/json' \
  -d '{
    "mappings": {
      "properties": {
        "tenant_id":    { "type": "keyword" },
        "user_id":      { "type": "keyword" },
        "session_id":   { "type": "keyword" },
        "action":       { "type": "keyword" },
        "resource":     { "type": "keyword" },
        "resource_id":  { "type": "keyword" },
        "severity":     { "type": "keyword" },
        "ip_address":   { "type": "ip" },
        "user_agent":   { "type": "text" },
        "message":      { "type": "text" },
        "metadata":     { "type": "object", "enabled": true },
        "before_state": { "type": "object", "enabled": true },
        "after_state":  { "type": "object", "enabled": true },
        "event_timestamp": { "type": "date" }
      }
    }
  }'

echo "Logs index created successfully"
