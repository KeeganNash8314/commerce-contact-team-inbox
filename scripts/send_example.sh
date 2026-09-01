#!/bin/sh
set -eu

curl --fail-with-body --request POST http://localhost:8080/contact \
  --header 'Content-Type: application/json' \
  --data '{"kind":"fulfillment","order_id":"ORD-1043","customer":"Lin Chen","email":"lin@example.com","message":"When will this ship?"}'
