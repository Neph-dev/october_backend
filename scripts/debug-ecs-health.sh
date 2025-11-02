#!/bin/bash
set -e

# Debug ECS Health Check Issues
# This script helps diagnose ECS task health check failures

echo "🔍 Debugging ECS Health Check Issues"
echo "===================================="

# Configuration
CLUSTER_NAME="october-cluster"
SERVICE_NAME="october-service"
REGION=${AWS_DEFAULT_REGION:-us-east-1}

# Check if AWS CLI is configured
if ! aws sts get-caller-identity >/dev/null 2>&1; then
    echo "❌ AWS CLI not configured or credentials expired"
    echo "Please run: aws configure"
    exit 1
fi

echo "✅ AWS credentials verified"
echo

# Get running tasks
echo "📋 Getting running tasks..."
TASK_ARNS=$(aws ecs list-tasks --cluster $CLUSTER_NAME --service-name $SERVICE_NAME --desired-status RUNNING --query "taskArns[]" --output text)

if [ -z "$TASK_ARNS" ]; then
    echo "❌ No running tasks found for service $SERVICE_NAME"
    exit 1
fi

echo "Found tasks: $TASK_ARNS"
echo

# Check task health for each task
for TASK_ARN in $TASK_ARNS; do
    echo "🔍 Checking task: $(basename $TASK_ARN)"
    
    # Get task details
    TASK_DETAILS=$(aws ecs describe-tasks --cluster $CLUSTER_NAME --tasks $TASK_ARN)
    
    # Extract health status
    HEALTH_STATUS=$(echo "$TASK_DETAILS" | jq -r '.tasks[0].healthStatus // "UNKNOWN"')
    LAST_STATUS=$(echo "$TASK_DETAILS" | jq -r '.tasks[0].lastStatus')
    DESIRED_STATUS=$(echo "$TASK_DETAILS" | jq -r '.tasks[0].desiredStatus')
    
    echo "  Health Status: $HEALTH_STATUS"
    echo "  Last Status: $LAST_STATUS"
    echo "  Desired Status: $DESIRED_STATUS"
    
    # Get container details
    echo "  Containers:"
    echo "$TASK_DETAILS" | jq -r '.tasks[0].containers[] | "    - \(.name): \(.lastStatus) (health: \(.healthStatus // "N/A"))"'
    
    # Get task network configuration
    NETWORK_CONFIG=$(echo "$TASK_DETAILS" | jq -r '.tasks[0].attachments[0].details[]? | select(.name=="networkInterfaceId") | .value')
    
    if [ -n "$NETWORK_CONFIG" ] && [ "$NETWORK_CONFIG" != "null" ]; then
        echo "  Network Interface: $NETWORK_CONFIG"
        
        # Get public IP
        PUBLIC_IP=$(aws ec2 describe-network-interfaces --network-interface-ids $NETWORK_CONFIG --query "NetworkInterfaces[0].Association.PublicIp" --output text 2>/dev/null || echo "None")
        PRIVATE_IP=$(aws ec2 describe-network-interfaces --network-interface-ids $NETWORK_CONFIG --query "NetworkInterfaces[0].PrivateIpAddress" --output text 2>/dev/null || echo "None")
        
        echo "  Public IP: $PUBLIC_IP"
        echo "  Private IP: $PRIVATE_IP"
        
        # Test health endpoint if accessible
        if [ "$PUBLIC_IP" != "None" ] && [ "$PUBLIC_IP" != "null" ]; then
            echo "  🏥 Testing health endpoint..."
            if curl -f -s --max-time 10 "http://$PUBLIC_IP:8080/health" >/dev/null 2>&1; then
                echo "  ✅ Health endpoint accessible"
                echo "  Response:"
                curl -s --max-time 5 "http://$PUBLIC_IP:8080/health" | jq . || echo "    (Not JSON or no response)"
            else
                echo "  ❌ Health endpoint not accessible"
                
                # Try to test if port is open
                if timeout 5 bash -c "</dev/tcp/$PUBLIC_IP/8080" 2>/dev/null; then
                    echo "  ℹ️  Port 8080 is open but health endpoint not responding"
                else
                    echo "  ❌ Port 8080 is not accessible"
                fi
            fi
        fi
    fi
    
    echo
done

# Check CloudWatch logs
echo "📝 Checking CloudWatch logs..."
LOG_GROUP="/ecs/october-task"

# Check if log group exists
if aws logs describe-log-groups --log-group-name-prefix "$LOG_GROUP" --query "logGroups[?logGroupName=='$LOG_GROUP']" --output text | grep -q "$LOG_GROUP"; then
    echo "✅ Log group exists: $LOG_GROUP"
    
    # Get recent log streams
    echo "Recent log streams:"
    aws logs describe-log-streams --log-group-name "$LOG_GROUP" --order-by LastEventTime --descending --max-items 3 --query "logStreams[*].[logStreamName,lastEventTime]" --output table
    
    echo
    echo "🔍 Getting recent logs (last 50 entries)..."
    aws logs filter-log-events --log-group-name "$LOG_GROUP" --start-time $(($(date +%s - 300) * 1000)) --query "events[*].[timestamp,message]" --output table
    
else
    echo "❌ Log group not found: $LOG_GROUP"
    echo "Available log groups:"
    aws logs describe-log-groups --log-group-name-prefix "/ecs/" --query "logGroups[*].logGroupName" --output table
fi

echo
echo "🔧 Troubleshooting Tips:"
echo "1. Check if application is starting correctly in logs"
echo "2. Verify database connection (MongoDB Atlas)"
echo "3. Check if health endpoint is responding on port 8080"
echo "4. Ensure security groups allow inbound traffic on port 8080"
echo "5. Verify Systems Manager parameters are set correctly"

echo
echo "🚀 Quick fixes to try:"
echo "1. Check Systems Manager parameters:"
echo "   aws ssm get-parameters --names 'october-backend-database-uri' 'october-backend-openai-api-key' 'october-backend-serper-api-key' --with-decryption"
echo
echo "2. Test database connectivity from local machine"
echo "3. Redeploy with updated health check configuration"