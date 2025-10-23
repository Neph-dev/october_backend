#!/bin/bash

# Test script for AI/RAG functionality
echo "=== October Backend AI/RAG Test Script ==="

API_BASE="http://localhost:8080"

echo ""
echo "1. Health Check"
curl -s "$API_BASE/health" | head -5

echo ""
echo ""
echo "2. Test AI Query - RTX Financial Performance"
curl -X POST "$API_BASE/ai/query" \
  -H "Content-Type: application/json" \
  -d '{
    "question": "How did RTX perform this quarter?",
    "company_context": ["Raytheon Technologies"]
  }' | head -20

echo ""
echo ""
echo "3. Test AI Query - Defense Contracts"
curl -X POST "$API_BASE/ai/query" \
  -H "Content-Type: application/json" \
  -d '{
    "question": "What defense contracts did RTX recently win?",
    "company_context": ["Raytheon Technologies"]
  }' | head -20

echo ""
echo ""
echo "4. Test AI Query - US War Department"
curl -X POST "$API_BASE/ai/query" \
  -H "Content-Type: application/json" \
  -d '{
    "question": "What are the latest military training developments?",
    "company_context": ["US War Department"]
  }' | head -20

echo ""
echo ""
echo "5. Test Query Analysis Only"
curl -X POST "$API_BASE/ai/analyze" \
  -H "Content-Type: application/json" \
  -d '{
    "question": "What were RTX earnings this quarter?"
  }' | head -10

echo ""
echo ""
echo "6. Test General Defense Industry Question"
curl -X POST "$API_BASE/ai/query" \
  -H "Content-Type: application/json" \
  -d '{
    "question": "What are the recent developments in defense technology?"
  }' | head -20

echo ""
echo ""
echo "=== AI/RAG Test Complete ==="

echo ""
echo "Note: Make sure to:"
echo "1. Set OPENAI_API_KEY in your environment"
echo "2. Ensure the server is running (make run-server)"
echo "3. Have processed some RSS feeds (make process-feeds)"