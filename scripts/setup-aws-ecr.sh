#!/bin/bash

# Setup AWS ECR Repository for October Backend
# This script creates the ECR repository if it doesn't exist

set -e

# Configuration
ECR_REPOSITORY="october-backend"
AWS_REGION="us-east-1"

echo "🚀 Setting up AWS ECR Repository: $ECR_REPOSITORY"

# Check if AWS CLI is configured
if ! aws sts get-caller-identity > /dev/null 2>&1; then
    echo "❌ AWS CLI is not configured. Please run 'aws configure' first."
    exit 1
fi

# Get AWS Account ID
ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
echo "📋 AWS Account ID: $ACCOUNT_ID"

# Check if ECR repository exists
if aws ecr describe-repositories --repository-names "$ECR_REPOSITORY" --region "$AWS_REGION" > /dev/null 2>&1; then
    echo "✅ ECR repository '$ECR_REPOSITORY' already exists"
else
    echo "🔨 Creating ECR repository '$ECR_REPOSITORY'..."
    
    # Create the repository
    aws ecr create-repository \
        --repository-name "$ECR_REPOSITORY" \
        --region "$AWS_REGION" \
        --image-scanning-configuration scanOnPush=true \
        --encryption-configuration encryptionType=AES256
    
    echo "✅ ECR repository '$ECR_REPOSITORY' created successfully"
fi

# Set lifecycle policy to manage image retention
echo "🔧 Setting up lifecycle policy..."
cat > /tmp/ecr-lifecycle-policy.json << 'EOF'
{
    "rules": [
        {
            "rulePriority": 1,
            "description": "Keep last 10 tagged images with version prefixes",
            "selection": {
                "tagStatus": "tagged",
                "tagPrefixList": ["v", "latest"],
                "countType": "imageCountMoreThan",
                "countNumber": 10
            },
            "action": {
                "type": "expire"
            }
        },
        {
            "rulePriority": 2,
            "description": "Keep last 5 commit-hash tagged images",
            "selection": {
                "tagStatus": "tagged",
                "countType": "imageCountMoreThan",
                "countNumber": 5
            },
            "action": {
                "type": "expire"
            }
        },
        {
            "rulePriority": 3,
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

echo "✅ Lifecycle policy applied"

# Clean up temporary file
rm /tmp/ecr-lifecycle-policy.json

# Display repository URI
REPOSITORY_URI="$ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com/$ECR_REPOSITORY"
echo ""
echo "🎉 ECR Setup Complete!"
echo "📍 Repository URI: $REPOSITORY_URI"
echo ""
echo "Next Steps:"
echo "1. Your GitHub Actions workflow will now be able to push images"
echo "2. Make sure these secrets are set in your GitHub repository:"
echo "   - AWS_ACCESS_KEY_ID"
echo "   - AWS_SECRET_ACCESS_KEY"
echo "3. Commit and push your code to trigger the deployment"