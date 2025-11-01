#!/bin/bash

# Quick test script to verify your deployment is working
# Run this after deploying to AWS

set -e

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

print_status() { echo -e "${BLUE}[TEST]${NC} $1"; }
print_success() { echo -e "${GREEN}[PASS]${NC} $1"; }
print_error() { echo -e "${RED}[FAIL]${NC} $1"; }

# Get service endpoint
echo "🧪 October Backend Deployment Test"
echo "=================================="
echo

print_status "Getting service endpoint..."
TASK_ARN=$(aws ecs list-tasks --cluster october-cluster --service-name october-backend-service --query "taskArns[0]" --output text)

if [ "$TASK_ARN" = "None" ] || [ "$TASK_ARN" = "" ]; then
    print_error "No running tasks found"
    exit 1
fi

PUBLIC_IP=$(aws ecs describe-tasks --cluster october-cluster --tasks $TASK_ARN --query "tasks[0].attachments[0].details[?name=='networkInterfaceId'].value" --output text | xargs -I {} aws ec2 describe-network-interfaces --network-interface-ids {} --query "NetworkInterfaces[0].Association.PublicIp" --output text)

if [ "$PUBLIC_IP" = "None" ] || [ "$PUBLIC_IP" = "" ]; then
    print_error "Could not get public IP"
    exit 1
fi

ENDPOINT="http://$PUBLIC_IP:8080"
print_success "Service endpoint: $ENDPOINT"
echo

# Test health endpoint
print_status "Testing health endpoint..."
if curl -f -s "$ENDPOINT/health" | jq . >/dev/null 2>&1; then
    print_success "Health check passed"
    curl -s "$ENDPOINT/health" | jq .
else
    print_error "Health check failed"
    exit 1
fi
echo

# Test companies endpoint
print_status "Testing companies endpoint..."
if curl -f -s "$ENDPOINT/companies" >/dev/null 2>&1; then
    print_success "Companies endpoint accessible"
    COMPANY_COUNT=$(curl -s "$ENDPOINT/companies" | jq -r '.pagination.count')
    echo "Found $COMPANY_COUNT companies"
else
    print_error "Companies endpoint failed"
fi
echo

# Test news endpoint
print_status "Testing news endpoint..."
if curl -f -s "$ENDPOINT/news?limit=1" >/dev/null 2>&1; then
    print_success "News endpoint accessible"
    NEWS_COUNT=$(curl -s "$ENDPOINT/news?limit=1" | jq -r '.pagination.count')
    echo "Found $NEWS_COUNT news articles"
else
    print_error "News endpoint failed"
fi
echo

# Test market data endpoint (may fail if no API key)
print_status "Testing market tickers endpoint..."
if curl -f -s "$ENDPOINT/market/tickers" >/dev/null 2>&1; then
    print_success "Market tickers endpoint accessible"
else
    print_error "Market tickers endpoint failed (may need valid API keys)"
fi
echo

# Performance test
print_status "Running basic performance test..."
echo "Testing response times (5 requests)..."
for i in {1..5}; do
    RESPONSE_TIME=$(curl -w "%{time_total}" -s -o /dev/null "$ENDPOINT/health")
    echo "Request $i: ${RESPONSE_TIME}s"
done
echo

print_success "🎉 Deployment test completed!"
echo
echo "Your October Backend is running at: $ENDPOINT"
echo
echo "Available endpoints:"
echo "  🏥 Health: $ENDPOINT/health"
echo "  🏢 Companies: $ENDPOINT/companies"
echo "  📰 News: $ENDPOINT/news"
echo "  📈 Market: $ENDPOINT/market/tickers"
echo "  🤖 AI Query: $ENDPOINT/ai/query (POST)"
echo
echo "Next steps:"
echo "  1. Set up a custom domain"
echo "  2. Configure HTTPS with SSL certificate"
echo "  3. Set up monitoring and alerting"
echo "  4. Configure auto-scaling"