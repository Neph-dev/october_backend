#!/bin/bash

# Deployment Validation Script for October Backend
# This script validates that all prerequisites are in place for successful deployment

set -e

echo "🔍 October Backend - Deployment Validation"
echo "=========================================="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Track validation status
VALIDATION_ERRORS=0

# Function to print status
print_status() {
    local status=$1
    local message=$2
    
    if [ "$status" = "pass" ]; then
        echo -e "${GREEN}✅ $message${NC}"
    elif [ "$status" = "warn" ]; then
        echo -e "${YELLOW}⚠️  $message${NC}"
    else
        echo -e "${RED}❌ $message${NC}"
        VALIDATION_ERRORS=$((VALIDATION_ERRORS + 1))
    fi
}

# Check AWS CLI configuration
echo "1️⃣ Checking AWS CLI Configuration..."
if aws sts get-caller-identity > /dev/null 2>&1; then
    ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
    print_status "pass" "AWS CLI configured - Account ID: $ACCOUNT_ID"
else
    print_status "fail" "AWS CLI not configured. Run 'aws configure' first."
fi

echo ""

# Check AWS Resources
echo "2️⃣ Checking AWS Resources..."

# ECR Repository
if aws ecr describe-repositories --repository-names october-backend > /dev/null 2>&1; then
    print_status "pass" "ECR repository 'october-backend' exists"
else
    print_status "fail" "ECR repository 'october-backend' does not exist"
fi

# ECS Cluster
if aws ecs describe-clusters --clusters october-cluster --query 'clusters[0].status' --output text 2>/dev/null | grep -q "ACTIVE"; then
    print_status "pass" "ECS cluster 'october-cluster' is active"
else
    print_status "fail" "ECS cluster 'october-cluster' does not exist or is not active"
fi

# IAM Role
if aws iam get-role --role-name ecsTaskExecutionRole-october > /dev/null 2>&1; then
    print_status "pass" "IAM role 'ecsTaskExecutionRole-october' exists"
else
    print_status "fail" "IAM role 'ecsTaskExecutionRole-october' does not exist"
fi

# CloudWatch Log Group
if aws logs describe-log-groups --log-group-name-prefix "/ecs/october-backend" --query 'logGroups[?logGroupName==`/ecs/october-backend`]' --output text | grep -q "/ecs/october-backend"; then
    print_status "pass" "CloudWatch log group '/ecs/october-backend' exists"
else
    print_status "warn" "CloudWatch log group '/ecs/october-backend' does not exist (will be created automatically)"
fi

echo ""

# Check Systems Manager Parameters
echo "3️⃣ Checking Systems Manager Parameters..."

# Required parameters
if aws ssm get-parameter --name "/october/DATABASE_URI" > /dev/null 2>&1; then
    print_status "pass" "Required parameter '/october/DATABASE_URI' exists"
else
    print_status "fail" "Required parameter '/october/DATABASE_URI' is missing"
fi

# Optional parameters
OPTIONAL_PARAMS=("OPENAI_API_KEY" "CUSTOM_SEARCH_API_KEY" "CUSTOM_SEARCH_ENGINE_ID" "FINNHUB_API_KEY")
for param in "${OPTIONAL_PARAMS[@]}"; do
    if aws ssm get-parameter --name "/october/$param" > /dev/null 2>&1; then
        print_status "pass" "Optional parameter '/october/$param' exists"
    else
        print_status "warn" "Optional parameter '/october/$param' is missing (AI features may not work)"
    fi
done

echo ""

# Check GitHub Secrets (if we can detect them)
echo "4️⃣ GitHub Repository Configuration..."
if [ -d ".git" ]; then
    REPO_URL=$(git remote get-url origin 2>/dev/null || echo "unknown")
    print_status "pass" "Git repository detected: $REPO_URL"
    print_status "warn" "Please ensure GitHub secrets are set: AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY"
else
    print_status "warn" "Not in a git repository"
fi

echo ""

# Check local files
echo "5️⃣ Checking Local Files..."

if [ -f "aws/ecs-task-definition.json" ]; then
    print_status "pass" "ECS task definition file exists"
    
    # Validate JSON
    if jq . aws/ecs-task-definition.json > /dev/null 2>&1; then
        print_status "pass" "ECS task definition is valid JSON"
    else
        print_status "fail" "ECS task definition is invalid JSON"
    fi
else
    print_status "fail" "ECS task definition file 'aws/ecs-task-definition.json' is missing"
fi

if [ -f "Dockerfile" ]; then
    print_status "pass" "Dockerfile exists"
else
    print_status "fail" "Dockerfile is missing"
fi

if [ -f ".github/workflows/deploy.yml" ]; then
    print_status "pass" "GitHub Actions workflow exists"
else
    print_status "fail" "GitHub Actions workflow '.github/workflows/deploy.yml' is missing"
fi

echo ""

# Summary
echo "📋 Validation Summary"
echo "===================="

if [ $VALIDATION_ERRORS -eq 0 ]; then
    print_status "pass" "All critical validations passed! Ready for deployment 🚀"
    echo ""
    echo "Next steps:"
    echo "1. git add . && git commit -m 'Deploy to production'"
    echo "2. git push"
    echo "3. Monitor deployment in GitHub Actions"
else
    print_status "fail" "Found $VALIDATION_ERRORS critical issues that need to be resolved"
    echo ""
    echo "Quick fixes:"
    echo "• Run infrastructure setup: ./scripts/quick-aws-setup.sh"
    echo "• Set up parameters: ./scripts/setup-ssm-parameters.sh"
    echo "• Configure AWS CLI: aws configure"
fi

echo ""
exit $VALIDATION_ERRORS