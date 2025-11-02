#!/bin/bash

# AWS October Backend Deployment Script
# This script automates the deployment of October Backend to AWS ECS

set -e

# Configuration
AWS_REGION="us-east-1"
CLUSTER_NAME="october-cluster"
SERVICE_NAME="october-backend-service"
ECR_REPO_NAME="october-backend"
TASK_FAMILY="october-backend"
LOG_GROUP="/ecs/october-backend"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if AWS CLI is configured
check_aws_cli() {
    print_status "Checking AWS CLI configuration..."
    if ! aws sts get-caller-identity >/dev/null 2>&1; then
        print_error "AWS CLI not configured. Run 'aws configure' first."
        exit 1
    fi
    
    ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
    print_success "AWS CLI configured. Account ID: $ACCOUNT_ID"
}

# Create ECR repository if it doesn't exist
create_ecr_repo() {
    print_status "Creating ECR repository..."
    
    if aws ecr describe-repositories --repository-names $ECR_REPO_NAME --region $AWS_REGION >/dev/null 2>&1; then
        print_success "ECR repository '$ECR_REPO_NAME' already exists"
    else
        aws ecr create-repository --repository-name $ECR_REPO_NAME --region $AWS_REGION
        print_success "ECR repository '$ECR_REPO_NAME' created"
    fi
}

# Build and push Docker image
build_and_push() {
    print_status "Building and pushing Docker image..."
    
    # Get ECR login
    aws ecr get-login-password --region $AWS_REGION | docker login --username AWS --password-stdin $ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com
    
    # Build image
    print_status "Building Docker image..."
    docker build -t $ECR_REPO_NAME:latest .
    
    # Tag for ECR
    docker tag $ECR_REPO_NAME:latest $ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com/$ECR_REPO_NAME:latest
    
    # Push to ECR
    print_status "Pushing to ECR..."
    docker push $ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com/$ECR_REPO_NAME:latest
    
    print_success "Image pushed to ECR"
}

# Create IAM roles
create_iam_roles() {
    print_status "Creating IAM roles..."
    
    # Create execution role
    if aws iam get-role --role-name ecsTaskExecutionRole >/dev/null 2>&1; then
        print_success "ecsTaskExecutionRole already exists"
    else
        aws iam create-role --role-name ecsTaskExecutionRole --assume-role-policy-document '{
            "Version": "2012-10-17",
            "Statement": [
                {
                    "Effect": "Allow",
                    "Principal": {
                        "Service": "ecs-tasks.amazonaws.com"
                    },
                    "Action": "sts:AssumeRole"
                }
            ]
        }'
        
        aws iam attach-role-policy --role-name ecsTaskExecutionRole --policy-arn arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy
        aws iam attach-role-policy --role-name ecsTaskExecutionRole --policy-arn arn:aws:iam::aws:policy/AmazonSSMReadOnlyAccess
        
        print_success "ecsTaskExecutionRole created"
    fi
    
    # Create task role
    if aws iam get-role --role-name ecsTaskRole >/dev/null 2>&1; then
        print_success "ecsTaskRole already exists"
    else
        aws iam create-role --role-name ecsTaskRole --assume-role-policy-document '{
            "Version": "2012-10-17",
            "Statement": [
                {
                    "Effect": "Allow",
                    "Principal": {
                        "Service": "ecs-tasks.amazonaws.com"
                    },
                    "Action": "sts:AssumeRole"
                }
            ]
        }'
        
        print_success "ecsTaskRole created"
    fi
}

# Create CloudWatch log group
create_log_group() {
    print_status "Creating CloudWatch log group..."
    
    if aws logs describe-log-groups --log-group-name-prefix $LOG_GROUP --region $AWS_REGION | grep -q $LOG_GROUP; then
        print_success "Log group '$LOG_GROUP' already exists"
    else
        aws logs create-log-group --log-group-name $LOG_GROUP --region $AWS_REGION
        print_success "Log group '$LOG_GROUP' created"
    fi
}

# Create ECS cluster
create_cluster() {
    print_status "Creating ECS cluster..."
    
    if aws ecs describe-clusters --clusters $CLUSTER_NAME --region $AWS_REGION | grep -q $CLUSTER_NAME; then
        print_success "ECS cluster '$CLUSTER_NAME' already exists"
    else
        aws ecs create-cluster --cluster-name $CLUSTER_NAME --capacity-providers FARGATE --default-capacity-provider-strategy capacityProvider=FARGATE,weight=1 --region $AWS_REGION
        print_success "ECS cluster '$CLUSTER_NAME' created"
    fi
}

# Register task definition
register_task_definition() {
    print_status "Registering task definition..."
    
    # Update task definition with account ID
    sed "s/YOUR_ACCOUNT_ID/$ACCOUNT_ID/g" aws/ecs-task-definition.json > aws/ecs-task-definition-updated.json
    
    aws ecs register-task-definition --cli-input-json file://aws/ecs-task-definition-updated.json --region $AWS_REGION
    
    print_success "Task definition registered"
}

# Store secrets in Parameter Store
store_secrets() {
    print_status "Storing secrets in Parameter Store..."
    
    echo "Please enter your API keys (they will be stored securely in AWS Systems Manager):"
    
    read -s -p "OpenAI API Key: " OPENAI_KEY
    echo
    read -s -p "Custom Search API Key: " SEARCH_KEY
    echo
    read -p "Custom Search Engine ID: " ENGINE_ID
    read -s -p "Finnhub API Key: " FINNHUB_KEY
    echo
    read -p "Database URI (e.g., mongodb://username:password@host:port/database): " DB_URI
    
    aws ssm put-parameter --name "/october/OPENAI_API_KEY" --value "$OPENAI_KEY" --type "SecureString" --overwrite --region $AWS_REGION
    aws ssm put-parameter --name "/october/CUSTOM_SEARCH_API_KEY" --value "$SEARCH_KEY" --type "SecureString" --overwrite --region $AWS_REGION
    aws ssm put-parameter --name "/october/CUSTOM_SEARCH_ENGINE_ID" --value "$ENGINE_ID" --type "String" --overwrite --region $AWS_REGION
    aws ssm put-parameter --name "/october/FINNHUB_API_KEY" --value "$FINNHUB_KEY" --type "SecureString" --overwrite --region $AWS_REGION
    aws ssm put-parameter --name "/october/DATABASE_URI" --value "$DB_URI" --type "SecureString" --overwrite --region $AWS_REGION
    
    print_success "Secrets stored in Parameter Store"
}

# Create ECS service
create_service() {
    print_status "Creating ECS service..."
    
    # Get default VPC and subnets
    VPC_ID=$(aws ec2 describe-vpcs --filters "Name=isDefault,Values=true" --query "Vpcs[0].VpcId" --output text --region $AWS_REGION)
    SUBNET_IDS=$(aws ec2 describe-subnets --filters "Name=vpc-id,Values=$VPC_ID" --query "Subnets[*].SubnetId" --output text --region $AWS_REGION | tr '\t' ',')
    
    # Create security group
    SG_ID=$(aws ec2 create-security-group --group-name october-backend-sg --description "Security group for October Backend" --vpc-id $VPC_ID --region $AWS_REGION --query "GroupId" --output text 2>/dev/null || aws ec2 describe-security-groups --filters "Name=group-name,Values=october-backend-sg" --query "SecurityGroups[0].GroupId" --output text --region $AWS_REGION)
    
    # Allow HTTP traffic
    aws ec2 authorize-security-group-ingress --group-id $SG_ID --protocol tcp --port 8080 --cidr 0.0.0.0/0 --region $AWS_REGION 2>/dev/null || true
    
    if aws ecs describe-services --cluster $CLUSTER_NAME --services $SERVICE_NAME --region $AWS_REGION | grep -q $SERVICE_NAME; then
        print_success "ECS service '$SERVICE_NAME' already exists"
    else
        aws ecs create-service \
            --cluster $CLUSTER_NAME \
            --service-name $SERVICE_NAME \
            --task-definition $TASK_FAMILY:1 \
            --desired-count 1 \
            --launch-type FARGATE \
            --network-configuration "awsvpcConfiguration={subnets=[$SUBNET_IDS],securityGroups=[$SG_ID],assignPublicIp=ENABLED}" \
            --region $AWS_REGION
        
        print_success "ECS service '$SERVICE_NAME' created"
    fi
}

# Wait for service to be stable
wait_for_service() {
    print_status "Waiting for service to become stable..."
    aws ecs wait services-stable --cluster $CLUSTER_NAME --services $SERVICE_NAME --region $AWS_REGION
    print_success "Service is stable"
}

# Get service endpoint
get_endpoint() {
    print_status "Getting service endpoint..."
    
    TASK_ARN=$(aws ecs list-tasks --cluster $CLUSTER_NAME --service-name $SERVICE_NAME --region $AWS_REGION --query "taskArns[0]" --output text)
    
    if [ "$TASK_ARN" != "None" ] && [ "$TASK_ARN" != "" ]; then
        PUBLIC_IP=$(aws ecs describe-tasks --cluster $CLUSTER_NAME --tasks $TASK_ARN --region $AWS_REGION --query "tasks[0].attachments[0].details[?name=='networkInterfaceId'].value" --output text | xargs -I {} aws ec2 describe-network-interfaces --network-interface-ids {} --region $AWS_REGION --query "NetworkInterfaces[0].Association.PublicIp" --output text)
        
        if [ "$PUBLIC_IP" != "None" ] && [ "$PUBLIC_IP" != "" ]; then
            print_success "Service deployed successfully!"
            echo
            echo "🎉 Deployment Complete!"
            echo "📍 Service Endpoint: http://$PUBLIC_IP:8080"
            echo "🏥 Health Check: http://$PUBLIC_IP:8080/health"
            echo "🏢 Companies API: http://$PUBLIC_IP:8080/companies"
            echo
            echo "Testing endpoints..."
            echo "Health check:"
            curl -s "http://$PUBLIC_IP:8080/health" || echo "Health check failed"
            echo
        else
            print_warning "Could not retrieve public IP. Check AWS console for service status."
        fi
    else
        print_warning "No running tasks found. Check service status in AWS console."
    fi
}

# Main deployment function
main() {
    echo "🚀 October Backend AWS Deployment"
    echo "================================="
    echo
    
    check_aws_cli
    create_ecr_repo
    build_and_push
    create_iam_roles
    create_log_group
    create_cluster
    
    # Ask if user wants to store secrets
    read -p "Do you want to store API keys in Parameter Store? (y/n): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        store_secrets
    else
        print_warning "Skipping secret storage. Make sure to set up secrets manually."
    fi
    
    register_task_definition
    create_service
    wait_for_service
    get_endpoint
    
    echo
    print_success "Deployment script completed!"
    echo
    echo "Next steps:"
    echo "1. Set up a load balancer for production"
    echo "2. Configure a custom domain"
    echo "3. Set up monitoring and alerting"
    echo "4. Configure auto-scaling"
}

# Run main function
main "$@"