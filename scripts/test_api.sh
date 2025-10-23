#!/bin/bash

# API Testing Script for October Backend
# Tests the company API endpoints

set -e

BASE_URL="http://localhost:8080"
echo "Testing October Backend Company API"
echo "=================================="

# Test health endpoint
echo "1. Testing health endpoint..."
curl -s "$BASE_URL/health" | jq .
echo ""

# Test company endpoint - Lockheed Martin
echo "2. Testing company endpoint - Lockheed Martin..."
response=$(curl -s "$BASE_URL/company/Lockheed%20Martin")
if echo "$response" | jq -e '.name' > /dev/null 2>&1; then
    echo "✅ Successfully retrieved Lockheed Martin data"
    echo "$response" | jq .
else
    echo "❌ Failed to retrieve Lockheed Martin data"
    echo "$response"
fi
echo ""

# Test company endpoint - Raytheon Technologies
echo "3. Testing company endpoint - Raytheon Technologies..."
response=$(curl -s "$BASE_URL/company/Raytheon%20Technologies")
if echo "$response" | jq -e '.name' > /dev/null 2>&1; then
    echo "✅ Successfully retrieved Raytheon Technologies data"
    echo "$response" | jq .
else
    echo "❌ Failed to retrieve Raytheon Technologies data"
    echo "$response"
fi
echo ""

# Test non-existent company
echo "4. Testing non-existent company..."
response=$(curl -s "$BASE_URL/company/NonExistent")
if echo "$response" | jq -e '.error' > /dev/null 2>&1; then
    echo "✅ Correctly returned error for non-existent company"
    echo "$response" | jq .
else
    echo "❌ Unexpected response for non-existent company"
    echo "$response"
fi
echo ""

# Test rate limiting (if applicable)
echo "5. Testing rate limiting..."
echo "Making 5 rapid requests to test rate limiting..."
for i in {1..5}; do
    echo -n "Request $i: "
    status_code=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/company/Lockheed%20Martin")
    echo "HTTP $status_code"
    sleep 0.1
done
echo ""

# Test invalid endpoint
echo "6. Testing invalid endpoint..."
response=$(curl -s "$BASE_URL/company/")
echo "Response for empty company name: $response"
echo ""

echo "Testing completed!"
echo "=================="