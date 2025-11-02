#!/bin/bash

# Quick AWS Setup for October Backend
# This script sets up the minimal required AWS resources for deployment

set -e

# Configuration
PROJECT_NAME="october-backend"
AWS_REGION="us-east-1"
ECR_REPOSITORY="october-backend"
ECS_CLUSTER="october-cluster"
ECS_SERVICE="october-backend-service"

echo "🚀 Quick AWS Setup for October Backend"
echo "This will create the minimal resources needed for deployment."
echo ""

# Check if AWS CLI is configured
if ! aws sts get-caller-identity > /dev/null 2>&1; then
    echo "❌ AWS CLI is not configured. Please run 'aws configure' first."
    exit 1
fi

# Get AWS Account ID
ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
echo "📋 AWS Account ID: $ACCOUNT_ID"
echo "🌍 AWS Region: $AWS_REGION"
echo ""

# 1. Create ECR Repository
echo "1️⃣ Creating ECR Repository..."
if aws ecr describe-repositories --repository-names "$ECR_REPOSITORY" --region "$AWS_REGION" > /dev/null 2>&1; then
    echo "✅ ECR repository already exists"
else
    aws ecr create-repository \
        --repository-name "$ECR_REPOSITORY" \
        --region "$AWS_REGION" \
        --image-scanning-configuration scanOnPush=true \
        --encryption-configuration encryptionType=AES256 > /dev/null
    echo "✅ ECR repository created"
fi

# 2. Create ECS Cluster
echo ""
echo "2️⃣ Creating ECS Cluster..."
if aws ecs describe-clusters --clusters "$ECS_CLUSTER" --region "$AWS_REGION" --query 'clusters[0].status' --output text 2>/dev/null | grep -q "ACTIVE"; then
    echo "✅ ECS cluster already exists"
else
    aws ecs create-cluster \
        --cluster-name "$ECS_CLUSTER" \
        --region "$AWS_REGION" \
        --capacity-providers FARGATE \
        --default-capacity-provider-strategy capacityProvider=FARGATE,weight=1 > /dev/null
    echo "✅ ECS cluster created"
fi

# 3. Create IAM Role
echo ""
echo "3️⃣ Creating IAM Role..."
TASK_EXECUTION_ROLE_NAME="ecsTaskExecutionRole-october"
if aws iam get-role --role-name "$TASK_EXECUTION_ROLE_NAME" > /dev/null 2>&1; then
    echo "✅ Task execution role already exists"
else
    # Create assume role policy
    cat > /tmp/assume-role-policy.json << 'EOF'
{
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
}
EOF

    aws iam create-role \
        --role-name "$TASK_EXECUTION_ROLE_NAME" \
        --assume-role-policy-document file:///tmp/assume-role-policy.json > /dev/null

    aws iam attach-role-policy \
        --role-name "$TASK_EXECUTION_ROLE_NAME" \
        --policy-arn "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy" > /dev/null

    rm /tmp/assume-role-policy.json
    echo "✅ Task execution role created"
fi

echo ""
echo "🎉 Quick AWS Setup Complete!"
echo ""
echo "📋 Created Resources:"
echo "  • ECR Repository: $ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com/$ECR_REPOSITORY"
echo "  • ECS Cluster: $ECS_CLUSTER"
echo "  • IAM Role: $TASK_EXECUTION_ROLE_NAME"
echo ""
echo "📝 Next Steps:"
echo "1. Set GitHub repository secrets:"
echo "   - AWS_ACCESS_KEY_ID"
echo "   - AWS_SECRET_ACCESS_KEY"
echo "2. Push your code to trigger deployment"
echo ""
echo "💡 The GitHub Actions workflow will handle:"
echo "   - ECS Service creation"
echo "   - Security Group setup"
echo "   - Network configuration"
echo "   - Database connection setup"