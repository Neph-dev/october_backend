#!/bin/bash

# Test script for company-based web search validation
# Usage: ./scripts/test_company_based_search.sh

set -e

BASE_URL="${BASE_URL:-http://localhost:8080}"

echo "Testing Company-Based Web Search Validation..."
echo "Base URL: $BASE_URL"
echo

# Test 1: RTX founder (should work - RTX is in database)
echo "=== Test 1: RTX Founder Query ==="
echo "Query: 'Who was the founder of RTX?'"
echo "Expected: Should work because RTX is in our database"
curl -X POST "$BASE_URL/ai/query" \
  -H "Content-Type: application/json" \
  -d '{
    "question": "Who was the founder of RTX?"
  }' | jq '.'
echo
echo

# Test 2: RTX random question (should work)
echo "=== Test 2: RTX Random Question ==="
echo "Query: 'What is RTX favorite color?'"
echo "Expected: Should work because RTX is in our database (OpenAI will handle the logic)"
curl -X POST "$BASE_URL/ai/query" \
  -H "Content-Type: application/json" \
  -d '{
    "question": "What is RTX favorite color?"
  }' | jq '.'
echo
echo

# Test 3: Raytheon Technologies question (should work)
echo "=== Test 3: Raytheon Technologies Question ==="
echo "Query: 'What does Raytheon Technologies sell?'"
echo "Expected: Should work because Raytheon Technologies is in our database"
curl -X POST "$BASE_URL/ai/query" \
  -H "Content-Type: application/json" \
  -d '{
    "question": "What does Raytheon Technologies sell?"
  }' | jq '.'
echo
echo

# Test 4: US War Department question (should work)
echo "=== Test 4: US War Department Question ==="
echo "Query: 'What is the mission of the US War Department?'"
echo "Expected: Should work because US War Department is in our database"
curl -X POST "$BASE_URL/ai/query" \
  -H "Content-Type: application/json" \
  -d '{
    "question": "What is the mission of the US War Department?"
  }' | jq '.'
echo
echo

# Test 5: Non-database company (should fail)
echo "=== Test 5: Non-Database Company ==="
echo "Query: 'Who founded Apple Inc?'"
echo "Expected: Should fail because Apple is not in our database"
curl -X POST "$BASE_URL/ai/query" \
  -H "Content-Type: application/json" \
  -d '{
    "question": "Who founded Apple Inc?"
  }' | jq '.'
echo
echo

# Test 6: Direct web search for RTX
echo "=== Test 6: Direct Web Search for RTX ==="
echo "Query: 'Who was the founder of RTX?'"
echo "Expected: Should be accepted"
curl -X POST "$BASE_URL/ai/web-search" \
  -H "Content-Type: application/json" \
  -d '{
    "question": "Who was the founder of RTX?"
  }' | jq '.'
echo
echo

# Test 7: Direct web search for non-database company
echo "=== Test 7: Direct Web Search for Non-Database Company ==="
echo "Query: 'Who founded Apple Inc?'"
echo "Expected: Should be rejected"
curl -X POST "$BASE_URL/ai/web-search" \
  -H "Content-Type: application/json" \
  -d '{
    "question": "Who founded Apple Inc?"
  }' | jq '.'
echo
echo

echo "Company-based web search testing completed!"