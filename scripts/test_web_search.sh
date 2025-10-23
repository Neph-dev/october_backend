#!/bin/bash

# Test script for AI web search functionality
# Usage: ./scripts/test_web_search.sh

set -e

BASE_URL="${BASE_URL:-http://localhost:8080}"

echo "Testing AI Web Search API..."
echo "Base URL: $BASE_URL"
echo

# Test 1: Defense-related query
echo "=== Test 1: Defense Contract Query ==="
echo "Query: 'Latest defense contracts for Raytheon'"
curl -X POST "$BASE_URL/ai/web-search" \
  -H "Content-Type: application/json" \
  -d '{
    "question": "Latest defense contracts for Raytheon",
    "companies": ["Raytheon Technologies"]
  }' | jq '.'
echo
echo

# Test 2: Aeronautics query
echo "=== Test 2: Aeronautics Query ==="
echo "Query: 'New fighter jet developments 2024'"
curl -X POST "$BASE_URL/ai/web-search" \
  -H "Content-Type: application/json" \
  -d '{
    "question": "New fighter jet developments 2024",
    "companies": ["RTX", "Lockheed Martin"]
  }' | jq '.'
echo
echo

# Test 3: RTX performance query
echo "=== Test 3: RTX Performance Query ==="
echo "Query: 'RTX earnings and financial performance'"
curl -X POST "$BASE_URL/ai/web-search" \
  -H "Content-Type: application/json" \
  -d '{
    "question": "RTX earnings and financial performance",
    "companies": ["RTX"]
  }' | jq '.'
echo
echo

# Test 4: Military technology query
echo "=== Test 4: Military Technology Query ==="
echo "Query: 'Advanced military radar systems'"
curl -X POST "$BASE_URL/ai/web-search" \
  -H "Content-Type: application/json" \
  -d '{
    "question": "Advanced military radar systems",
    "companies": ["Raytheon Technologies"]
  }' | jq '.'
echo
echo

# Test 5: Non-defense query (should fail)
echo "=== Test 5: Non-Defense Query (Should Fail) ==="
echo "Query: 'Best pizza recipes'"
curl -X POST "$BASE_URL/ai/web-search" \
  -H "Content-Type: application/json" \
  -d '{
    "question": "Best pizza recipes"
  }' | jq '.'
echo
echo

# Test 6: Compare with regular AI query
echo "=== Test 6: Regular AI Query for Comparison ==="
echo "Query: 'How did RTX perform this quarter?' (using full RAG pipeline)"
curl -X POST "$BASE_URL/ai/query" \
  -H "Content-Type: application/json" \
  -d '{
    "question": "How did RTX perform this quarter?",
    "company_context": ["Raytheon Technologies"]
  }' | jq '.'
echo
echo

echo "Web search testing completed!"