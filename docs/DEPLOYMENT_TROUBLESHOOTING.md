# AWS Deployment Troubleshooting Guide

## 🚨 Common Deployment Issues

### 1. ECR Repository Not Found Error
**Error:** `name unknown: The repository with name 'october-backend' does not exist`

**Solutions:**
```bash
# Option 1: Run the infrastructure setup script
./scripts/setup-aws-infrastructure.sh

# Option 2: Create just the ECR repository
./scripts/setup-aws-ecr.sh

# Option 3: Manual creation
aws ecr create-repository --repository-name october-backend --region us-east-1
```

### 2. GitHub Actions Authentication Issues
**Error:** `Unable to locate credentials`

**Solutions:**
1. Set GitHub repository secrets:
   - `AWS_ACCESS_KEY_ID`
   - `AWS_SECRET_ACCESS_KEY`

2. Verify AWS credentials have proper permissions:
   ```json
   {
     "Version": "2012-10-17",
     "Statement": [
       {
         "Effect": "Allow",
         "Action": [
           "ecr:*",
           "ecs:*",
           "iam:PassRole",
           "ec2:DescribeNetworkInterfaces",
           "ssm:GetParameter*"
         ],
         "Resource": "*"
       }
     ]
   }
   ```

### 3. ECS Task Definition Issues
**Error:** `Invalid task definition`

**Solutions:**
1. Check if ECS cluster exists:
   ```bash
   aws ecs describe-clusters --clusters october-cluster
   ```

2. Verify task execution role exists:
   ```bash
   aws iam get-role --role-name ecsTaskExecutionRole-october
   ```

3. Run the full infrastructure setup:
   ```bash
   ./scripts/setup-aws-infrastructure.sh
   ```

### 4. Health Check Failures
**Error:** `Health check failed`

**Solutions:**
1. Check ECS service logs:
   ```bash
   aws logs tail /ecs/october-backend --follow
   ```

2. Verify environment variables in task definition:
   - `DATABASE_URI` should reference Systems Manager parameter
   - Port configuration should match (8080)

3. Check security group allows inbound traffic on port 8080

### 5. Database Connection Issues
**Error:** `Failed to connect to MongoDB`

**Solutions:**
1. Verify database URI in Systems Manager:
   ```bash
   aws ssm get-parameter --name "/october/DATABASE_URI" --with-decryption
   ```

2. Check if MongoDB cluster allows connections from AWS:
   - Add AWS ECS IP ranges to MongoDB network access list
   - Verify credentials in connection string

3. Test connection from local environment:
   ```bash
   export DATABASE_URI="your-connection-string"
   go run cmd/api/main.go
   ```

## 🔧 Quick Fixes

### Reset Everything
```bash
# 1. Delete and recreate ECR repository
aws ecr delete-repository --repository-name october-backend --force
./scripts/setup-aws-ecr.sh

# 2. Update ECS service
aws ecs update-service --cluster october-cluster --service october-backend-service --force-new-deployment

# 3. Check deployment status
aws ecs describe-services --cluster october-cluster --services october-backend-service
```

### View Logs
```bash
# Get latest task ARN
TASK_ARN=$(aws ecs list-tasks --cluster october-cluster --service-name october-backend-service --query "taskArns[0]" --output text)

# Get CloudWatch logs
aws logs tail /ecs/october-backend --follow

# Or check task details
aws ecs describe-tasks --cluster october-cluster --tasks $TASK_ARN
```

### Manual Deployment Test
```bash
# Build and push image manually
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin $ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com

docker build -t october-backend .
docker tag october-backend:latest $ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com/october-backend:latest
docker push $ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com/october-backend:latest
```

## 📞 Getting Help

1. **Check GitHub Actions logs** - Most errors are visible in the workflow output
2. **Review AWS CloudWatch logs** - Application runtime errors appear here
3. **Verify AWS resources** - Ensure all required resources exist
4. **Test locally first** - Always verify the application works locally before deploying

## 🔍 Useful Commands

```bash
# Check all AWS resources
aws ecr describe-repositories --region us-east-1
aws ecs describe-clusters --clusters october-cluster
aws ecs describe-services --cluster october-cluster --services october-backend-service
aws ssm get-parameters-by-path --path "/october" --recursive

# Monitor deployment
watch aws ecs describe-services --cluster october-cluster --services october-backend-service --query 'services[0].deployments'

# Get service endpoint
aws ecs describe-services --cluster october-cluster --services october-backend-service --query 'services[0].loadBalancers'
```