#!/bin/bash
#
# Gemeente Vlierdam - setup #2
# ============================
#
# Relatie-model met gecombineerde regels.
#

export FGA_STORE_ID=$(curl -s -q "$FGA_API_URL/stores?name=vlierdam2" | jq -r '.stores[].id')
echo "store=$FGA_STORE_ID"

if [ ! -z ${FGA_STORE_ID+x} ]; then
  curl -X DELETE "$FGA_API_URL/stores/$FGA_STORE_ID"
  FGA_STORE_ID=$(curl -s -X POST "$FGA_API_URL/stores" -H "content-type: application/json" -d '{"name":"vlierdam2"}' | jq -r '.id')
  echo "store=$FGA_STORE_ID"
fi

export FGA_MODEL_ID=$(curl -s -X POST "$FGA_API_URL/stores/$FGA_STORE_ID/authorization-models" \
  -H "content-type: application/json" --data-binary "@vlierdam2.model.json" \
  | jq -r '.authorization_model_id')
echo "model=$FGA_MODEL_ID"

JSON_DATA=$(envsubst < vlierdam2.tuples.json)
curl -s -X POST "$FGA_API_URL/stores/$FGA_STORE_ID/write" -H "content-type: application/json" -d "$JSON_DATA"

curl -s -X PUT "$FGA_API_URL/stores/${FGA_STORE_ID}/assertions/$FGA_MODEL_ID" \
  -H "content-type: application/json" --data-binary "@vlierdam2.assertions.json"
