# Scripts Directory

This directory contains essential scripts for managing and testing the October Backend.

## 🔧 Management Scripts

### `diagnose-aws-resources.sh`
**Purpose**: Comprehensive diagnostic tool for AWS ECS resources
**When to use**: When troubleshooting deployment issues or checking AWS resource status
**Usage**: `./scripts/diagnose-aws-resources.sh`

### `fix-ssm-parameters.sh`
**Purpose**: Set up or update AWS Systems Manager parameters for the application
**When to use**: When configuring API keys, database connections, or other secrets
**Usage**: `./scripts/fix-ssm-parameters.sh`

### `quick-status-check.sh`
**Purpose**: Quick check of ECS cluster and service status
**When to use**: For fast status verification before/after deployments
**Usage**: `./scripts/quick-status-check.sh`

## 🧪 Testing Scripts

### `test-deployment.sh`
**Purpose**: Test deployed application endpoints and functionality
**When to use**: After successful deployment to verify all endpoints work
**Usage**: `./scripts/test-deployment.sh`

### `test_api.sh`
**Purpose**: Test company API endpoints locally or remotely
**When to use**: For testing company-specific functionality
**Usage**: `./scripts/test_api.sh`

### `test_full_api.sh`
**Purpose**: Comprehensive API testing for all endpoints
**When to use**: For full regression testing of the API
**Usage**: `./scripts/test_full_api.sh`

## 📝 Usage Notes

- All scripts require appropriate AWS credentials when testing deployed services
- Local testing scripts assume the application is running on `localhost:8080`
- Make sure scripts are executable: `chmod +x scripts/*.sh`

## 🚀 Quick Commands

```bash
# Check AWS deployment status
./scripts/quick-status-check.sh

# Diagnose deployment issues  
./scripts/diagnose-aws-resources.sh

# Test local API
./scripts/test_full_api.sh

# Test deployed API
./scripts/test-deployment.sh
```