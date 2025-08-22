#!/bin/bash
set -e

echo "Creating OpenSearch index 'logs' with mapping..."

curl -s -X PUT "http://opensearch:9200/logs" \
  -H 'Content-Type: application/json' \
  -d '{
    "mappings": {
      "properties": {
        "id":          { "type": "keyword" },
        "tenant_id":   { "type": "keyword" },
        "user_id":     { "type": "keyword" },
        "session_id":  { "type": "keyword" },
        "action":      { "type": "keyword" },
        "resource":    { "type": "keyword" },
        "resource_id": { "type": "keyword" },
        "severity":    { "type": "keyword" },
        "ip_address":  { "type": "ip" },
        "user_agent":  { "type": "text" },
        "message": {
          "type": "text",
          "fields": {
            "raw": { "type": "keyword" }
          }
        },
        "before":      { "type": "object", "enabled": true },
        "after":       { "type": "object", "enabled": true },
        "metadata":    { "type": "object", "enabled": true },
        "event_timestamp": { "type": "date" }
      }
    }
  }'

echo "✅ OpenSearch index 'logs' created"
