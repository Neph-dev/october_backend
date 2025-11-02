#!/bin/bash
set -e

# Fix Systems Manager Parameters for October Backend
# This script helps set up the required parameters for ECS deployment

echo "🔧 Setting up Systems Manager Parameters"
echo "========================================"

# Check if AWS CLI is configured
if ! aws sts get-caller-identity >/dev/null 2>&1; then
    echo "❌ AWS CLI not configured or credentials expired"
    echo "Please run: aws configure"
    exit 1
fi

echo "✅ AWS credentials verified"
echo

# Function to check if parameter exists
check_parameter() {
    local param_name=$1
    if aws ssm get-parameter --name "$param_name" >/dev/null 2>&1; then
        echo "✅ Parameter exists: $param_name"
        return 0
    else
        echo "❌ Parameter missing: $param_name"
        return 1
    fi
}

# Function to create or update parameter
create_parameter() {
    local param_name=$1
    local param_value=$2
    local param_type=${3:-SecureString}
    
    echo "Creating/updating parameter: $param_name"
    aws ssm put-parameter \
        --name "$param_name" \
        --value "$param_value" \
        --type "$param_type" \
        --overwrite \
        --description "October Backend configuration parameter"
    
    if [ $? -eq 0 ]; then
        echo "✅ Successfully set: $param_name"
    else
        echo "❌ Failed to set: $param_name"
    fi
}

echo "🔍 Checking required parameters..."

# Required parameters
REQUIRED_PARAMS=(
    "/october/DATABASE_URI"
    "/october/OPENAI_API_KEY"
    "/october/CUSTOM_SEARCH_API_KEY"
    "/october/CUSTOM_SEARCH_ENGINE_ID"
    "/october/SERPER_API_KEY"
)

# Check existing parameters
echo "Current parameter status:"
for param in "${REQUIRED_PARAMS[@]}"; do
    check_parameter "$param" || true
done

echo
echo "🚀 Parameter Setup Options:"
echo "1. Set all parameters interactively"
echo "2. Set parameters from .env file"
echo "3. Show current parameter values (masked)"
echo "4. Exit"

read -p "Choose an option (1-4): " choice

case $choice in
    1)
        echo
        echo "📝 Setting parameters interactively..."
        echo "Enter values for each parameter (leave empty to skip):"
        
        read -p "MongoDB DATABASE_URI: " -s db_uri
        echo
        if [ -n "$db_uri" ]; then
            create_parameter "/october/DATABASE_URI" "$db_uri"
        fi
        
        read -p "OpenAI API Key: " -s openai_key
        echo
        if [ -n "$openai_key" ]; then
            create_parameter "/october/OPENAI_API_KEY" "$openai_key"
        fi
        
        read -p "Google Custom Search API Key: " -s search_key
        echo
        if [ -n "$search_key" ]; then
            create_parameter "/october/CUSTOM_SEARCH_API_KEY" "$search_key"
        fi
        
        read -p "Google Custom Search Engine ID: " search_engine_id
        if [ -n "$search_engine_id" ]; then
            create_parameter "/october/CUSTOM_SEARCH_ENGINE_ID" "$search_engine_id" "String"
        fi
        
        read -p "Serper API Key: " -s serper_key
        echo
        if [ -n "$serper_key" ]; then
            create_parameter "/october/SERPER_API_KEY" "$serper_key"
        fi
        ;;
        
    2)
        if [ -f ".env" ]; then
            echo
            echo "📂 Reading from .env file..."
            source .env
            
            if [ -n "$DATABASE_URI" ]; then
                create_parameter "/october/DATABASE_URI" "$DATABASE_URI"
            fi
            
            if [ -n "$OPENAI_API_KEY" ]; then
                create_parameter "/october/OPENAI_API_KEY" "$OPENAI_API_KEY"
            fi
            
            if [ -n "$CUSTOM_SEARCH_API_KEY" ]; then
                create_parameter "/october/CUSTOM_SEARCH_API_KEY" "$CUSTOM_SEARCH_API_KEY"
            fi
            
            if [ -n "$CUSTOM_SEARCH_ENGINE_ID" ]; then
                create_parameter "/october/CUSTOM_SEARCH_ENGINE_ID" "$CUSTOM_SEARCH_ENGINE_ID" "String"
            fi
            
            if [ -n "$SERPER_API_KEY" ]; then
                create_parameter "/october/SERPER_API_KEY" "$SERPER_API_KEY"
            fi
        else
            echo "❌ .env file not found"
        fi
        ;;
        
    3)
        echo
        echo "📋 Current parameter values:"
        for param in "${REQUIRED_PARAMS[@]}"; do
            if aws ssm get-parameter --name "$param" --with-decryption >/dev/null 2>&1; then
                value=$(aws ssm get-parameter --name "$param" --with-decryption --query "Parameter.Value" --output text)
                masked_value="${value:0:8}..."
                echo "  $param: $masked_value"
            else
            echo "  $param: NOT SET"
            fi
        done
        ;;
        
    4)
        echo "Exiting..."
        exit 0
        ;;
        
    *)
        echo "Invalid option"
        exit 1
        ;;
esac

echo
echo "✅ Parameter setup complete!"
echo
echo "🔍 Verification:"
echo "Run this command to verify all parameters are set:"
echo "aws ssm get-parameters --names ${REQUIRED_PARAMS[*]} --with-decryption"
echo
echo "🚀 Next steps:"
echo "1. Redeploy your ECS service to pick up new parameters"
echo "2. Check ECS task logs for any remaining issues"
echo "3. Test health endpoints manually"