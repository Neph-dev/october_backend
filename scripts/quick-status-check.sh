#!/bin/bash
set -e

# Simple AWS ECS Status Check
# Quick script to check current state of your ECS resources

echo "🔍 Quick ECS Status Check"
echo "========================"

# Configuration
CLUSTER_NAME="october-cluster"
SERVICE_NAME="october-backend-service"
REGION="us-east-1"

# Check AWS CLI access
if ! aws sts get-caller-identity >/dev/null 2>&1; then
    echo "❌ AWS CLI not configured or credentials expired"
    exit 1
fi

echo "✅ AWS credentials OK"
echo

# Check cluster
echo "🏗️  Cluster Status:"
CLUSTER_STATUS=$(aws ecs describe-clusters --clusters $CLUSTER_NAME --region $REGION --query 'clusters[0].status' --output text 2>/dev/null || echo "NOT_FOUND")
echo "   $CLUSTER_NAME: $CLUSTER_STATUS"

# Check service
echo
echo "🚀 Service Status:"
SERVICE_RESPONSE=$(aws ecs describe-services --cluster $CLUSTER_NAME --services $SERVICE_NAME --region $REGION 2>/dev/null || echo "ERROR")

if [ "$SERVICE_RESPONSE" = "ERROR" ]; then
    echo "   ❌ Cannot check service (cluster might not exist)"
elif [ "$(echo "$SERVICE_RESPONSE" | jq -r '.services | length')" = "0" ]; then
    echo "   ❌ Service does not exist"
    echo "   💡 This explains the MISSING error - service needs to be created"
else
    SERVICE_STATUS=$(echo "$SERVICE_RESPONSE" | jq -r '.services[0].status')
    RUNNING_COUNT=$(echo "$SERVICE_RESPONSE" | jq -r '.services[0].runningCount')
    DESIRED_COUNT=$(echo "$SERVICE_RESPONSE" | jq -r '.services[0].desiredCount')
    echo "   $SERVICE_NAME: $SERVICE_STATUS"
    echo "   Tasks: $RUNNING_COUNT/$DESIRED_COUNT"
fi

echo
echo "🎯 Quick Analysis:"
if [ "$CLUSTER_STATUS" = "INACTIVE" ]; then
    echo "   🔥 ISSUE: Cluster is INACTIVE"
    echo "   💡 SOLUTION: Updated workflow will delete and recreate it"
elif [ "$CLUSTER_STATUS" = "NOT_FOUND" ]; then
    echo "   🔥 ISSUE: Cluster doesn't exist"  
    echo "   💡 SOLUTION: Workflow will create it"
elif [ "$SERVICE_RESPONSE" = "ERROR" ] || [ "$(echo "$SERVICE_RESPONSE" | jq -r '.services | length' 2>/dev/null || echo "0")" = "0" ]; then
    echo "   🔥 ISSUE: Service is MISSING (explains your error)"
    echo "   💡 SOLUTION: Updated workflow will create the service"
else
    echo "   ✅ Resources look OK, might be a timing/health check issue"
fi

echo
echo "🚀 Next Steps:"
echo "   1. The updated workflow should handle these issues automatically"
echo "   2. Push the changes to trigger redeployment"
echo "   3. The enhanced error handling will provide better diagnostics"