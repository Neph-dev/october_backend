#!/bin/bash

# Setup Systems Manager Parameters for October Backend
# This script helps set up the required environment variables in AWS Systems Manager

set -e

echo "🔧 Setting up Systems Manager Parameters for October Backend"
echo ""

# Check if AWS CLI is configured
if ! aws sts get-caller-identity > /dev/null 2>&1; then
    echo "❌ AWS CLI is not configured. Please run 'aws configure' first."
    exit 1
fi

# Get AWS Account ID
ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
echo "📋 AWS Account ID: $ACCOUNT_ID"
echo "🌍 AWS Region: us-east-1"
echo ""

# Function to create or update parameter
create_or_update_parameter() {
    local param_name=$1
    local param_description=$2
    local is_required=$3
    
    if aws ssm get-parameter --name "/october/$param_name" > /dev/null 2>&1; then
        echo "✅ Parameter /october/$param_name already exists"
        return
    fi
    
    if [ "$is_required" = "true" ]; then
        echo "🔑 Setting up REQUIRED parameter: $param_name"
        echo "💡 $param_description"
        echo -n "Please enter the value: "
        read -s param_value
        echo ""
        
        if [ -z "$param_value" ]; then
            echo "❌ Value cannot be empty for required parameter"
            return 1
        fi
    else
        echo "🔑 Setting up OPTIONAL parameter: $param_name"
        echo "💡 $param_description"
        echo -n "Please enter the value (press Enter to skip): "
        read -s param_value
        echo ""
        
        if [ -z "$param_value" ]; then
            echo "⏭️  Skipped optional parameter: $param_name"
            return
        fi
    fi
    
    aws ssm put-parameter \
        --name "/october/$param_name" \
        --value "$param_value" \
        --type "SecureString" \
        --description "$param_description"
    
    echo "✅ Parameter /october/$param_name created successfully"
    echo ""
}

# Required parameters
echo "📋 Setting up REQUIRED parameters:"
echo ""

create_or_update_parameter "DATABASE_URI" "MongoDB connection string (e.g., mongodb+srv://<your_user>:<your_password>@cluster.mongodb.net/<your_db>)" "true"

# Optional parameters
echo ""
echo "📋 Setting up OPTIONAL parameters (for AI features):"
echo ""

create_or_update_parameter "OPENAI_API_KEY" "OpenAI API key for AI features" "false"
create_or_update_parameter "CUSTOM_SEARCH_API_KEY" "Google Custom Search API key" "false"
create_or_update_parameter "CUSTOM_SEARCH_ENGINE_ID" "Google Custom Search Engine ID" "false"
create_or_update_parameter "FINNHUB_API_KEY" "Finnhub API key for financial data" "false"

echo ""
echo "🎉 Systems Manager Parameters Setup Complete!"
echo ""
echo "📋 Summary of created parameters:"
aws ssm get-parameters-by-path --path "/october" --query 'Parameters[*].Name' --output table

echo ""
echo "📝 Next Steps:"
echo "1. Your parameters are now securely stored in AWS Systems Manager"
echo "2. The ECS task definition will automatically load these as environment variables"
echo "3. Deploy your application using the GitHub Actions workflow"
echo ""
echo "🔧 To update a parameter later:"
echo "aws ssm put-parameter --name '/october/PARAMETER_NAME' --value 'NEW_VALUE' --type SecureString --overwrite"