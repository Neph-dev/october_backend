#!/bin/bash

# ECS Service-Linked Role Setup
# This script creates the required ECS service-linked role

set -e

echo "🔗 Setting up ECS Service-Linked Role"
echo ""

# Check if AWS CLI is configured
if ! aws sts get-caller-identity > /dev/null 2>&1; then
    echo "❌ AWS CLI is not configured. Please run 'aws configure' first."
    exit 1
fi

# Get AWS Account ID for reference
ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
echo "📋 AWS Account ID: $ACCOUNT_ID"
echo ""

# Check if ECS service-linked role exists
echo "🔍 Checking for ECS service-linked role..."

if aws iam get-role --role-name AWSServiceRoleForECS > /dev/null 2>&1; then
    echo "✅ ECS service-linked role already exists"
    
    # Show role details
    ROLE_ARN=$(aws iam get-role --role-name AWSServiceRoleForECS --query 'Role.Arn' --output text)
    echo "📍 Role ARN: $ROLE_ARN"
else
    echo "🔨 Creating ECS service-linked role..."
    
    # Create the service-linked role
    aws iam create-service-linked-role --aws-service-name ecs.amazonaws.com
    
    # Wait for the role to be available
    echo "⏳ Waiting for role to be available..."
    sleep 15
    
    # Verify creation
    if aws iam get-role --role-name AWSServiceRoleForECS > /dev/null 2>&1; then
        ROLE_ARN=$(aws iam get-role --role-name AWSServiceRoleForECS --query 'Role.Arn' --output text)
        echo "✅ ECS service-linked role created successfully"
        echo "📍 Role ARN: $ROLE_ARN"
    else
        echo "❌ Failed to create service-linked role"
        exit 1
    fi
fi

echo ""
echo "🎉 ECS Service-Linked Role Setup Complete!"
echo ""
echo "ℹ️  What is a service-linked role?"
echo "   A service-linked role is a unique type of IAM role that is linked"
echo "   directly to an AWS service. ECS uses this role to make calls to"
echo "   other AWS services on your behalf."
echo ""
echo "📝 Next Steps:"
echo "   • You can now create ECS clusters and services"
echo "   • Run your deployment pipeline"
echo "   • The role will be automatically used by ECS"