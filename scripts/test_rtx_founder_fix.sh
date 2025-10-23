#!/bin/bash

# Test script for RTX founder query fix
# Usage: ./scripts/test_rtx_founder_fix.sh

set -e

BASE_URL="${BASE_URL:-http://localhost:8080}"

echo "Testing RTX Founder Query Fix..."
echo "Base URL: $BASE_URL"
echo

# Test 1: "Who was the founder of RTX?" (uppercase)
echo "=== Test 1: RTX Founder Query (Uppercase) ==="
echo "Query: 'Who was the founder of RTX?'"
echo "Expected: Should trigger web search and provide answer"
curl -X POST "$BASE_URL/ai/query" \
  -H "Content-Type: application/json" \
  -d '{
    "question": "Who was the founder of RTX?"
  }' | jq '.'
echo
echo

# Test 2: "Who was the founder of rtx?" (lowercase)
echo "=== Test 2: RTX Founder Query (Lowercase) ==="
echo "Query: 'Who was the founder of rtx?'"
echo "Expected: Should also trigger web search and provide answer"
curl -X POST "$BASE_URL/ai/query" \
  -H "Content-Type: application/json" \
  -d '{
    "question": "Who was the founder of rtx?"
  }' | jq '.'
echo
echo

# Test 3: Direct web search test (uppercase)
echo "=== Test 3: Direct Web Search (Uppercase RTX) ==="
echo "Query: 'Who was the founder of RTX?'"
echo "Expected: Should be accepted as defense-related"
curl -X POST "$BASE_URL/ai/web-search" \
  -H "Content-Type: application/json" \
  -d '{
    "question": "Who was the founder of RTX?"
  }' | jq '.'
echo
echo

# Test 4: Direct web search test (lowercase)
echo "=== Test 4: Direct Web Search (Lowercase rtx) ==="
echo "Query: 'Who was the founder of rtx?'"
echo "Expected: Should be accepted as defense-related"
curl -X POST "$BASE_URL/ai/web-search" \
  -H "Content-Type: application/json" \
  -d '{
    "question": "Who was the founder of rtx?"
  }' | jq '.'
echo
echo

# Test 5: Raytheon founder query
echo "=== Test 5: Raytheon Founder Query ==="
echo "Query: 'Who founded Raytheon Technologies?'"
curl -X POST "$BASE_URL/ai/web-search" \
  -H "Content-Type: application/json" \
  -d '{
    "question": "Who founded Raytheon Technologies?"
  }' | jq '.'
echo
echo

# Test 6: Non-defense query (should still fail)
echo "=== Test 6: Non-Defense Query (Should Still Fail) ==="
echo "Query: 'Who founded Apple Inc?'"
echo "Expected: Should be rejected as not defense-related"
curl -X POST "$BASE_URL/ai/web-search" \
  -H "Content-Type: application/json" \
  -d '{
    "question": "Who founded Apple Inc?"
  }' | jq '.'
echo
echo

echo "RTX founder query fix testing completed!"