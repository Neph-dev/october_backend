#!/bin/bash

# AWS Infrastructure Setup for October Backend
# This script sets up all required AWS resources

set -e

# Configuration
PROJECT_NAME="october-backend"
AWS_REGION="us-east-1"
ECR_REPOSITORY="october-backend"
ECS_CLUSTER="october-cluster"
ECS_SERVICE="october-backend-service"

echo "🚀 Setting up AWS infrastructure for October Backend"

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
echo "1️⃣ Setting up ECR Repository..."
if aws ecr describe-repositories --repository-names "$ECR_REPOSITORY" --region "$AWS_REGION" > /dev/null 2>&1; then
    echo "✅ ECR repository '$ECR_REPOSITORY' already exists"
else
    echo "🔨 Creating ECR repository '$ECR_REPOSITORY'..."
    aws ecr create-repository \
        --repository-name "$ECR_REPOSITORY" \
        --region "$AWS_REGION" \
        --image-scanning-configuration scanOnPush=true \
        --encryption-configuration encryptionType=AES256

    # Set lifecycle policy
    cat > /tmp/ecr-lifecycle-policy.json << 'EOF'
{
    "rules": [
        {
            "rulePriority": 1,
            "description": "Keep last 10 tagged images",
            "selection": {
                "tagStatus": "tagged",
                "countType": "imageCountMoreThan",
                "countNumber": 10
            },
            "action": {
                "type": "expire"
            }
        },
        {
            "rulePriority": 2,
            "description": "Delete untagged images older than 1 day",
            "selection": {
                "tagStatus": "untagged",
                "countType": "sinceImagePushed",
                "countUnit": "days",
                "countNumber": 1
            },
            "action": {
                "type": "expire"
            }
        }
    ]
}
EOF

    aws ecr put-lifecycle-policy \
        --repository-name "$ECR_REPOSITORY" \
        --region "$AWS_REGION" \
        --lifecycle-policy-text file:///tmp/ecr-lifecycle-policy.json

    rm /tmp/ecr-lifecycle-policy.json
    echo "✅ ECR repository created with lifecycle policy"
fi

# 2. Create ECS Cluster
echo ""
echo "2️⃣ Setting up ECS Cluster..."
if aws ecs describe-clusters --clusters "$ECS_CLUSTER" --region "$AWS_REGION" --query 'clusters[0].status' --output text 2>/dev/null | grep -q "ACTIVE"; then
    echo "✅ ECS cluster '$ECS_CLUSTER' already exists"
else
    echo "🔨 Creating ECS cluster '$ECS_CLUSTER'..."
    aws ecs create-cluster \
        --cluster-name "$ECS_CLUSTER" \
        --region "$AWS_REGION" \
        --capacity-providers FARGATE \
        --default-capacity-provider-strategy capacityProvider=FARGATE,weight=1
    echo "✅ ECS cluster created"
fi

# 3. Create Systems Manager Parameters for configuration
echo ""
echo "3️⃣ Setting up Systems Manager Parameters..."

# Database URI parameter
if aws ssm get-parameter --name "/october/DATABASE_URI" --region "$AWS_REGION" > /dev/null 2>&1; then
    echo "✅ DATABASE_URI parameter already exists"
else
    echo "🔨 Creating DATABASE_URI parameter..."
    echo "❓ Please enter your MongoDB connection string:"
    read -s DATABASE_URI
    aws ssm put-parameter \
        --name "/october/DATABASE_URI" \
        --value "$DATABASE_URI" \
        --type "SecureString" \
        --region "$AWS_REGION" \
        --description "MongoDB connection string for October Backend"
    echo "✅ DATABASE_URI parameter created"
fi

# RSS API Key parameter (optional)
if aws ssm get-parameter --name "/october/RSS_API_KEY" --region "$AWS_REGION" > /dev/null 2>&1; then
    echo "✅ RSS_API_KEY parameter already exists"
else
    echo "🔨 Creating RSS_API_KEY parameter..."
    echo "❓ Please enter your RSS API key (press Enter to skip):"
    read -s RSS_API_KEY
    if [ ! -z "$RSS_API_KEY" ]; then
        aws ssm put-parameter \
            --name "/october/RSS_API_KEY" \
            --value "$RSS_API_KEY" \
            --type "SecureString" \
            --region "$AWS_REGION" \
            --description "RSS API key for October Backend"
        echo "✅ RSS_API_KEY parameter created"
    else
        echo "⏭️  RSS_API_KEY parameter skipped"
    fi
fi

# 4. Create IAM roles if they don't exist
echo ""
echo "4️⃣ Setting up IAM Roles..."

# Task execution role
TASK_EXECUTION_ROLE_NAME="ecsTaskExecutionRole-october"
if aws iam get-role --role-name "$TASK_EXECUTION_ROLE_NAME" > /dev/null 2>&1; then
    echo "✅ Task execution role already exists"
else
    echo "🔨 Creating ECS task execution role..."
    
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
        --assume-role-policy-document file:///tmp/assume-role-policy.json

    # Attach AWS managed policy
    aws iam attach-role-policy \
        --role-name "$TASK_EXECUTION_ROLE_NAME" \
        --policy-arn "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"

    # Create custom policy for Systems Manager access
    cat > /tmp/ssm-policy.json << 'EOF'
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "ssm:GetParameter",
                "ssm:GetParameters",
                "ssm:GetParametersByPath"
            ],
            "Resource": [
                "arn:aws:ssm:*:*:parameter/october/*"
            ]
        }
    ]
}
EOF

    aws iam put-role-policy \
        --role-name "$TASK_EXECUTION_ROLE_NAME" \
        --policy-name "SSMParameterAccess" \
        --policy-document file:///tmp/ssm-policy.json

    rm /tmp/assume-role-policy.json /tmp/ssm-policy.json
    echo "✅ Task execution role created"
fi

echo ""
echo "🎉 AWS Infrastructure Setup Complete!"
echo ""
echo "📋 Summary:"
echo "  • ECR Repository: $ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com/$ECR_REPOSITORY"
echo "  • ECS Cluster: $ECS_CLUSTER"
echo "  • Task Execution Role: $TASK_EXECUTION_ROLE_NAME"
echo "  • Systems Manager Parameters: /october/*"
echo ""
echo "📝 Next Steps:"
echo "1. Set up GitHub Secrets:"
echo "   - AWS_ACCESS_KEY_ID"
echo "   - AWS_SECRET_ACCESS_KEY"
echo "2. Commit and push your code to trigger deployment"
echo "3. Monitor the deployment in AWS ECS console"
echo ""
echo "🔗 Useful URLs:"
echo "  • ECS Console: https://console.aws.amazon.com/ecs/home?region=$AWS_REGION"
echo "  • ECR Console: https://console.aws.amazon.com/ecr/repositories?region=$AWS_REGION"
echo "  • Systems Manager: https://console.aws.amazon.com/systems-manager/parameters?region=$AWS_REGION"