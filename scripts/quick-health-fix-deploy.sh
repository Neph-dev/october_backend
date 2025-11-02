#!/bin/bash
set -e

# Quick ECS Health Check Fix Deployment
# Run this script from any machine with AWS CLI access

echo "🚀 Deploying ECS Health Check Fix"
echo "================================="

# Configuration
CLUSTER_NAME="october-cluster"
SERVICE_NAME="october-backend-service"
TASK_FAMILY="october-backend"   # Update if different

# Check AWS CLI access
if ! aws sts get-caller-identity >/dev/null 2>&1; then
    echo "❌ AWS CLI not configured or credentials expired"
    echo "Please run: aws configure"
    exit 1
fi

echo "✅ AWS credentials verified"

# Get current task definition
echo "📋 Getting current task definition..."
CURRENT_TASK_DEF=$(aws ecs describe-task-definition --task-definition $TASK_FAMILY --query 'taskDefinition')

# Create new task definition JSON with fixed health check
echo "🔧 Creating updated task definition..."
echo "$CURRENT_TASK_DEF" | jq '
  .containerDefinitions[0].healthCheck = {
    "command": [
      "CMD-SHELL",
      "curl -f http://localhost:8080/liveness || wget --no-verbose --tries=1 --spider http://localhost:8080/liveness || exit 1"
    ],
    "interval": 30,
    "timeout": 15,
    "retries": 5,
    "startPeriod": 180
  } |
  del(.taskDefinitionArn, .revision, .status, .requiresAttributes, .placementConstraints, .compatibilities, .registeredAt, .registeredBy)
' > /tmp/updated-task-def.json

# Register new task definition
echo "📝 Registering new task definition..."
NEW_TASK_DEF_ARN=$(aws ecs register-task-definition --cli-input-json file:///tmp/updated-task-def.json --query 'taskDefinition.taskDefinitionArn' --output text)

echo "✅ New task definition registered: $NEW_TASK_DEF_ARN"

# Update service to use new task definition
echo "🔄 Updating service..."
aws ecs update-service \
    --cluster $CLUSTER_NAME \
    --service $SERVICE_NAME \
    --task-definition $NEW_TASK_DEF_ARN

echo "⏳ Waiting for service to stabilize..."
aws ecs wait services-stable --cluster $CLUSTER_NAME --services $SERVICE_NAME

echo "✅ Deployment complete!"
echo
echo "🔍 Checking service status..."
aws ecs describe-services --cluster $CLUSTER_NAME --services $SERVICE_NAME --query 'services[0].{
    serviceName: serviceName,
    taskDefinition: taskDefinition,
    runningCount: runningCount,
    desiredCount: desiredCount,
    status: status
}'

echo
echo "🎉 Health check fix deployed successfully!"
echo "Your ECS tasks should now show HEALTHY status instead of cycling through unhealthy states."

# Cleanup
rm -f /tmp/updated-task-def.json