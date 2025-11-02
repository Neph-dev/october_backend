#!/bin/bash
set -e

# AWS ECS Resource Diagnostic Script
# This script checks the status of your ECS resources and helps identify issues

echo "🔍 AWS ECS Resource Diagnostic"
echo "=============================="

# Configuration
CLUSTER_NAME="october-cluster"
SERVICE_NAME="october-backend-service"
REGION="us-east-1"

# Check AWS CLI access
if ! aws sts get-caller-identity >/dev/null 2>&1; then
    echo "❌ AWS CLI not configured or credentials expired"
    echo "Please run: aws configure"
    exit 1
fi

echo "✅ AWS credentials verified"
ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
echo "🔑 Account ID: $ACCOUNT_ID"
echo

# 1. Check ECS Cluster Status
echo "🏗️  Checking ECS Cluster Status..."
echo "================================="
CLUSTER_STATUS=$(aws ecs describe-clusters --clusters $CLUSTER_NAME --region $REGION --query 'clusters[0].status' --output text 2>/dev/null || echo "NOT_FOUND")
echo "Cluster: $CLUSTER_NAME"
echo "Status: $CLUSTER_STATUS"

if [ "$CLUSTER_STATUS" = "ACTIVE" ]; then
    echo "✅ Cluster is active"
elif [ "$CLUSTER_STATUS" = "INACTIVE" ]; then
    echo "⚠️  Cluster is inactive - this is why deployment failed!"
    echo "💡 Solution: The workflow will now delete and recreate inactive clusters"
elif [ "$CLUSTER_STATUS" = "NOT_FOUND" ]; then
    echo "❌ Cluster not found - needs to be created"
else
    echo "❓ Unknown cluster status: $CLUSTER_STATUS"
fi
echo

# 2. Check ECS Service Status
echo "🚀 Checking ECS Service Status..."
echo "================================="
if [ "$CLUSTER_STATUS" = "ACTIVE" ]; then
    SERVICE_INFO=$(aws ecs describe-services --cluster $CLUSTER_NAME --services $SERVICE_NAME --region $REGION 2>/dev/null || echo "NOT_FOUND")
    
    if [ "$SERVICE_INFO" != "NOT_FOUND" ]; then
        SERVICE_STATUS=$(echo "$SERVICE_INFO" | jq -r '.services[0].status // "UNKNOWN"')
        RUNNING_COUNT=$(echo "$SERVICE_INFO" | jq -r '.services[0].runningCount // 0')
        DESIRED_COUNT=$(echo "$SERVICE_INFO" | jq -r '.services[0].desiredCount // 0')
        TASK_DEFINITION=$(echo "$SERVICE_INFO" | jq -r '.services[0].taskDefinition // "NONE"')
        
        echo "Service: $SERVICE_NAME"
        echo "Status: $SERVICE_STATUS"
        echo "Running Tasks: $RUNNING_COUNT"
        echo "Desired Tasks: $DESIRED_COUNT"
        echo "Task Definition: $TASK_DEFINITION"
        
        # Get recent service events
        echo
        echo "📋 Recent Service Events:"
        echo "$SERVICE_INFO" | jq -r '.services[0].events[:5][] | "\(.createdAt) - \(.message)"'
        
    else
        echo "❌ Service not found"
    fi
else
    echo "⏭️  Skipping service check - cluster not active"
fi
echo

# 3. Check Running Tasks
echo "📦 Checking Running Tasks..."
echo "============================"
if [ "$CLUSTER_STATUS" = "ACTIVE" ]; then
    TASK_ARNS=$(aws ecs list-tasks --cluster $CLUSTER_NAME --service-name $SERVICE_NAME --region $REGION --query 'taskArns' --output text 2>/dev/null || echo "")
    
    if [ -n "$TASK_ARNS" ] && [ "$TASK_ARNS" != "None" ]; then
        echo "Found tasks: $TASK_ARNS"
        
        for TASK_ARN in $TASK_ARNS; do
            TASK_ID=$(basename $TASK_ARN)
            echo
            echo "📋 Task: $TASK_ID"
            
            TASK_DETAILS=$(aws ecs describe-tasks --cluster $CLUSTER_NAME --tasks $TASK_ARN --region $REGION)
            
            LAST_STATUS=$(echo "$TASK_DETAILS" | jq -r '.tasks[0].lastStatus')
            DESIRED_STATUS=$(echo "$TASK_DETAILS" | jq -r '.tasks[0].desiredStatus')
            HEALTH_STATUS=$(echo "$TASK_DETAILS" | jq -r '.tasks[0].healthStatus // "UNKNOWN"')
            CPU_UTILIZATION=$(echo "$TASK_DETAILS" | jq -r '.tasks[0].cpu // "N/A"')
            MEMORY_UTILIZATION=$(echo "$TASK_DETAILS" | jq -r '.tasks[0].memory // "N/A"')
            
            echo "  Last Status: $LAST_STATUS"
            echo "  Desired Status: $DESIRED_STATUS"
            echo "  Health Status: $HEALTH_STATUS"
            echo "  CPU: $CPU_UTILIZATION"
            echo "  Memory: $MEMORY_UTILIZATION"
            
            # Get container details
            echo "  Containers:"
            echo "$TASK_DETAILS" | jq -r '.tasks[0].containers[] | "    - \(.name): \(.lastStatus) (health: \(.healthStatus // "N/A"))"'
            
            # Try to get public IP
            NETWORK_INTERFACE=$(echo "$TASK_DETAILS" | jq -r '.tasks[0].attachments[0].details[]? | select(.name=="networkInterfaceId") | .value')
            if [ -n "$NETWORK_INTERFACE" ] && [ "$NETWORK_INTERFACE" != "null" ]; then
                PUBLIC_IP=$(aws ec2 describe-network-interfaces --network-interface-ids $NETWORK_INTERFACE --region $REGION --query "NetworkInterfaces[0].Association.PublicIp" --output text 2>/dev/null || echo "None")
                echo "  Public IP: $PUBLIC_IP"
                
                if [ "$PUBLIC_IP" != "None" ] && [ "$PUBLIC_IP" != "null" ]; then
                    echo "  🌐 Service URL: http://$PUBLIC_IP:8080"
                    echo "  🏥 Health URL: http://$PUBLIC_IP:8080/health"
                    echo "  💓 Liveness URL: http://$PUBLIC_IP:8080/liveness"
                fi
            fi
        done
    else
        echo "❌ No tasks found"
    fi
else
    echo "⏭️  Skipping task check - cluster not active"
fi
echo

# 4. Check ECR Repository
echo "🐳 Checking ECR Repository..."
echo "============================="
ECR_REPO="october-backend"
if aws ecr describe-repositories --repository-names $ECR_REPO --region $REGION >/dev/null 2>&1; then
    echo "✅ ECR repository exists: $ECR_REPO"
    
    # Get image count
    IMAGE_COUNT=$(aws ecr list-images --repository-name $ECR_REPO --region $REGION --query 'length(imageIds)')
    echo "📸 Images in repository: $IMAGE_COUNT"
    
    # Get latest image
    LATEST_IMAGE=$(aws ecr describe-images --repository-name $ECR_REPO --region $REGION --query 'sort_by(imageDetails,&imagePushedAt)[-1].imageTags[0]' --output text 2>/dev/null || echo "No tags")
    echo "🏷️  Latest image tag: $LATEST_IMAGE"
else
    echo "❌ ECR repository not found: $ECR_REPO"
fi
echo

# 5. Summary and Recommendations
echo "📋 Summary and Recommendations"
echo "==============================="

if [ "$CLUSTER_STATUS" = "INACTIVE" ]; then
    echo "🔥 CRITICAL ISSUE FOUND:"
    echo "   Your ECS cluster is INACTIVE - this explains the deployment failure!"
    echo
    echo "🚀 SOLUTION:"
    echo "   The updated GitHub workflow will now:"
    echo "   1. Detect inactive clusters"
    echo "   2. Delete and recreate them automatically"
    echo "   3. Use improved health checks with /liveness endpoint"
    echo
    echo "💡 Next steps:"
    echo "   1. Commit and push the updated workflow"
    echo "   2. The deployment should work automatically"
    
elif [ "$CLUSTER_STATUS" = "ACTIVE" ]; then
    echo "✅ Cluster is active - deployment should work"
    if [ -n "$TASK_ARNS" ]; then
        echo "✅ Tasks are running"
        echo "💡 Health check fixes have been applied to prevent cycling"
    else
        echo "⚠️  No tasks running - service might need to be recreated"
    fi
    
else
    echo "❌ Cluster not found - will be created on next deployment"
fi

echo
echo "🔧 Files updated in this session:"
echo "  ✅ .github/workflows/deploy.yml - Enhanced cluster/service handling"
echo "  ✅ aws/ecs-task-definition.json - Fixed health check to use /liveness"
echo "  ✅ Created diagnostic and deployment scripts"
echo
echo "🚀 Ready to deploy with improved error handling and health checks!"