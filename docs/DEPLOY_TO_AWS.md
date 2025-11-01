# AWS Deployment Guide for October Backend

This comprehensive guide provides step-by-step instructions to deploy the October Backend to AWS using multiple deployment strategies, specifically tailored for your Go application with MongoDB, RSS processing, and AI features.

## 🚀 Deployment Options

### Option 1: AWS ECS with Fargate (Recommended)
- **Best for**: Production workloads, auto-scaling, managed infrastructure
- **Cost**: ~$30-50/month for small workloads
- **Complexity**: Medium
- **Features**: Auto-scaling, load balancing, service discovery

### Option 2: AWS App Runner (Easiest)
- **Best for**: Simple deployments, minimal configuration
- **Cost**: ~$25-40/month for small workloads  
- **Complexity**: Low
- **Features**: Automatic deployments, built-in load balancing

### Option 3: AWS EC2 with Docker (Most Control)
- **Best for**: Custom configurations, cost optimization
- **Cost**: ~$15-30/month for small instances
- **Complexity**: High
- **Features**: Full control, custom networking

---

## 📋 Prerequisites

### 1. AWS Account Setup
You'll need an AWS account with permissions to create:
- ECR (Elastic Container Registry)
- ECS (Elastic Container Service) 
- IAM roles and policies
- CloudWatch logs
- Systems Manager Parameter Store
- VPC and Security Groups

### 2. Local Development Environment
```bash
# Install AWS CLI (macOS)
brew install awscli

# Or Linux/Windows
curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
unzip awscliv2.zip
sudo ./aws/install

# Configure AWS credentials
aws configure
# Enter: Access Key ID, Secret Access Key, Region (us-east-1), Output format (json)

# Verify configuration
aws sts get-caller-identity
```

### 3. Required API Keys
Before deployment, gather these API keys:
- **OpenAI API Key** - For AI features
- **Google Custom Search API Key** - For web search functionality  
- **Google Custom Search Engine ID** - For web search
- **Finnhub API Key** - For market data
- **Database Connection String** - MongoDB/DocumentDB URI

### 4. Docker Requirements
Ensure Docker is installed and running:
```bash
docker --version
docker run hello-world
```

---

## 🎯 Quick Start (5-Minute Deployment)

### Automated Deployment Script
The easiest way to deploy is using our automated script:

```bash
# Make sure you're in the project root
cd /Users/nephthali.salam/Documents/October/october_backend

# Run the deployment script
./aws/deploy.sh
```

This script will:
1. ✅ Create ECR repository
2. ✅ Build and push Docker image  
3. ✅ Create IAM roles
4. ✅ Set up CloudWatch logging
5. ✅ Create ECS cluster and service
6. ✅ Store your API keys securely
7. ✅ Deploy and provide endpoint URL

**Expected output:**
```
🚀 October Backend AWS Deployment
=================================

✅ ECR repository created
✅ Docker image built and pushed
✅ ECS service deployed
🎉 Service deployed successfully!
📍 Service Endpoint: http://54.123.456.789:8080
🏥 Health Check: http://54.123.456.789:8080/health
```

### Manual Step-by-Step Deployment
If you prefer manual control, follow the detailed steps below.

---

## 🔧 Manual Deployment Steps

### Step 1: Create ECR Repository
```bash
# Create ECR repository for your Docker images
aws ecr create-repository --repository-name october-backend --region us-east-1

# Note the repository URI (you'll need this later)
# Example: 123456789012.dkr.ecr.us-east-1.amazonaws.com/october-backend
```

### Step 2: Build and Push Docker Image
```bash
# Login to ECR
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin YOUR_ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com

# Build the Docker image (your Dockerfile is already production-ready)
docker build -t october-backend:latest .

# Tag for ECR
docker tag october-backend:latest YOUR_ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com/october-backend:latest

# Push to ECR
docker push YOUR_ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com/october-backend:latest
```

### Step 3: Store API Keys Securely
```bash
# Store all your API keys in AWS Systems Manager Parameter Store
aws ssm put-parameter --name "/october/OPENAI_API_KEY" --value "your-openai-key" --type "SecureString"
aws ssm put-parameter --name "/october/CUSTOM_SEARCH_API_KEY" --value "your-search-key" --type "SecureString"  
aws ssm put-parameter --name "/october/CUSTOM_SEARCH_ENGINE_ID" --value "your-engine-id" --type "String"
aws ssm put-parameter --name "/october/FINNHUB_API_KEY" --value "your-finnhub-key" --type "SecureString"
aws ssm put-parameter --name "/october/DATABASE_URI" --value "your-mongodb-uri" --type "SecureString"
```

### Step 4: Create IAM Roles
```bash
# Create ECS Task Execution Role
aws iam create-role --role-name ecsTaskExecutionRole --assume-role-policy-document '{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow", 
      "Principal": {"Service": "ecs-tasks.amazonaws.com"},
      "Action": "sts:AssumeRole"
    }
  ]
}'

# Attach required policies
aws iam attach-role-policy --role-name ecsTaskExecutionRole --policy-arn arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy
aws iam attach-role-policy --role-name ecsTaskExecutionRole --policy-arn arn:aws:iam::aws:policy/AmazonSSMReadOnlyAccess

# Create ECS Task Role (for application permissions)
aws iam create-role --role-name ecsTaskRole --assume-role-policy-document '{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {"Service": "ecs-tasks.amazonaws.com"}, 
      "Action": "sts:AssumeRole"
    }
  ]
}'
```

### Step 5: Create ECS Cluster and Task Definition
```bash
# Create ECS cluster
aws ecs create-cluster --cluster-name october-cluster --capacity-providers FARGATE --default-capacity-provider-strategy capacityProvider=FARGATE,weight=1

# Create CloudWatch log group
aws logs create-log-group --log-group-name /ecs/october-backend

# Register task definition (update aws/ecs-task-definition.json with your account ID first)
aws ecs register-task-definition --cli-input-json file://aws/ecs-task-definition.json
```

### Step 6: Create ECS Service
```bash
# Get your default VPC and subnets
VPC_ID=$(aws ec2 describe-vpcs --filters "Name=isDefault,Values=true" --query "Vpcs[0].VpcId" --output text)
SUBNET_IDS=$(aws ec2 describe-subnets --filters "Name=vpc-id,Values=$VPC_ID" --query "Subnets[*].SubnetId" --output text | tr '\t' ',')

# Create security group
SG_ID=$(aws ec2 create-security-group --group-name october-backend-sg --description "October Backend Security Group" --vpc-id $VPC_ID --query "GroupId" --output text)

# Allow HTTP traffic
aws ec2 authorize-security-group-ingress --group-id $SG_ID --protocol tcp --port 8080 --cidr 0.0.0.0/0

# Create ECS service
aws ecs create-service \
  --cluster october-cluster \
  --service-name october-backend-service \
  --task-definition october-backend:1 \
  --desired-count 1 \
  --launch-type FARGATE \
  --network-configuration "awsvpcConfiguration={subnets=[$SUBNET_IDS],securityGroups=[$SG_ID],assignPublicIp=ENABLED}"
```

---

## 🔄 CI/CD with GitHub Actions

### Automated Deployments
The repository includes a GitHub Actions workflow (`.github/workflows/deploy.yml`) that automatically:

1. **Builds** Docker image on every push to main
2. **Pushes** to ECR with commit SHA and latest tags
3. **Updates** ECS service with new image
4. **Performs** health checks post-deployment

### Setup GitHub Secrets
Add these secrets to your GitHub repository:

```bash
# Go to GitHub repo → Settings → Secrets and variables → Actions
# Add these repository secrets:

AWS_ACCESS_KEY_ID: your-aws-access-key
AWS_SECRET_ACCESS_KEY: your-aws-secret-key
```

### Trigger Deployment
```bash
# Any push to main branch will trigger deployment
git add .
git commit -m "Deploy to production"  
git push origin main

# Or manually trigger from GitHub Actions tab
```

---

## 🗄️ Database Options

### Option 1: MongoDB Atlas (Recommended)
```bash
# 1. Sign up at https://www.mongodb.com/atlas
# 2. Create a free M0 cluster
# 3. Get connection string: mongodb+srv://username:password@cluster.mongodb.net/october
# 4. Store in Parameter Store:
aws ssm put-parameter --name "/october/DATABASE_URI" --value "your-atlas-uri" --type "SecureString"
```

### Option 2: AWS DocumentDB
```bash
# Create DocumentDB cluster (MongoDB-compatible)
aws docdb create-db-cluster \
  --db-cluster-identifier october-docdb \
  --engine docdb \
  --master-username october_admin \
  --master-user-password YourSecurePassword123! \
  --backup-retention-period 7

# Create DocumentDB instance  
aws docdb create-db-instance \
  --db-instance-identifier october-docdb-instance \
  --db-instance-class db.t3.medium \
  --engine docdb \
  --db-cluster-identifier october-docdb

# Connection string will be: 
# mongodb://october_admin:YourSecurePassword123!@october-docdb.cluster-xyz.us-east-1.docdb.amazonaws.com:27017/october?tls=true&replicaSet=rs0&readPreference=secondaryPreferred&retryWrites=false
```

### Option 3: Self-Managed MongoDB on EC2
```bash
# Launch EC2 instance with MongoDB (for development/testing)
# Not recommended for production due to operational overhead
```

---

## 📊 Monitoring and Alerting

### CloudWatch Dashboards
```bash
# Create custom dashboard for monitoring
aws cloudwatch put-dashboard --dashboard-name "October-Backend" --dashboard-body '{
  "widgets": [
    {
      "type": "metric",
      "properties": {
        "metrics": [
          ["AWS/ECS", "CPUUtilization", "ServiceName", "october-backend-service"],
          ["AWS/ECS", "MemoryUtilization", "ServiceName", "october-backend-service"]
        ],
        "period": 300,
        "stat": "Average",
        "region": "us-east-1",
        "title": "ECS Service Metrics"
      }
    }
  ]
}'
```

### CloudWatch Alarms
```bash
# High CPU alarm
aws cloudwatch put-metric-alarm \
  --alarm-name "October-Backend-High-CPU" \
  --alarm-description "Alert when CPU exceeds 80%" \
  --metric-name CPUUtilization \
  --namespace AWS/ECS \
  --statistic Average \
  --period 300 \
  --threshold 80 \
  --comparison-operator GreaterThanThreshold \
  --dimensions Name=ServiceName,Value=october-backend-service

# High memory alarm  
aws cloudwatch put-metric-alarm \
  --alarm-name "October-Backend-High-Memory" \
  --alarm-description "Alert when memory exceeds 80%" \
  --metric-name MemoryUtilization \
  --namespace AWS/ECS \
  --statistic Average \
  --period 300 \
  --threshold 80 \
  --comparison-operator GreaterThanThreshold \
  --dimensions Name=ServiceName,Value=october-backend-service
```

### Application Logs
```bash
# View real-time logs
aws logs tail /ecs/october-backend --follow

# Search logs for errors
aws logs filter-log-events --log-group-name /ecs/october-backend --filter-pattern "ERROR"
```

---

## 🔧 Troubleshooting

### Common Issues

#### 1. Container Won't Start
```bash
# Check ECS service events
aws ecs describe-services --cluster october-cluster --services october-backend-service

# Check task logs
aws logs get-log-events --log-group-name /ecs/october-backend --log-stream-name ecs/october-backend/TASK_ID

# Common causes:
# - Missing environment variables
# - Image pull errors (check ECR permissions)
# - Health check failures
```

#### 2. Health Check Failures
```bash
# Test health endpoint directly
curl -v http://TASK_PUBLIC_IP:8080/health

# Common causes:
# - App not binding to 0.0.0.0:8080
# - Health endpoint not implemented
# - Security group blocking traffic
```

#### 3. Database Connection Issues
```bash
# Test database connectivity
# - Check DATABASE_URI parameter exists in Parameter Store
# - Verify security groups allow MongoDB traffic (port 27017)
# - Check VPC routing and NAT gateway configuration
```

#### 4. API Key Issues
```bash
# Check parameter store values
aws ssm get-parameter --name "/october/OPENAI_API_KEY" --with-decryption

# Common causes:
# - Parameter Store values not set
# - IAM role missing SSM permissions
# - Incorrect parameter names
```

### Health Check Script
Create a health check script to verify deployment:

```bash
#!/bin/bash
# health-check.sh
ENDPOINT="http://YOUR_PUBLIC_IP:8080"

echo "🏥 Testing health endpoint..."
if curl -f -s "$ENDPOINT/health"; then
    echo "✅ Health check passed"
else
    echo "❌ Health check failed"
    exit 1
fi

echo "🏢 Testing companies endpoint..."
curl -s "$ENDPOINT/companies" | jq .

echo "📰 Testing news endpoint..."  
curl -s "$ENDPOINT/news?limit=1" | jq .
```

---

## 💰 Cost Optimization

### Current Setup Costs (Estimated Monthly)
- **ECS Fargate**: ~$15-25 (0.5 vCPU, 1GB memory)
- **DocumentDB**: ~$50-80 (db.t3.medium)
- **MongoDB Atlas**: ~$0-9 (M0 free tier to M2)
- **CloudWatch Logs**: ~$1-3
- **Data Transfer**: ~$1-5
- **Total**: ~$17-113/month depending on database choice

### Cost Optimization Tips
```bash
# 1. Use MongoDB Atlas free tier for development
# 2. Set up billing alerts
aws budgets create-budget --account-id YOUR_ACCOUNT_ID --budget '{
  "BudgetName": "October-Backend-Monthly",
  "BudgetLimit": {"Amount": "50", "Unit": "USD"},
  "TimeUnit": "MONTHLY",
  "BudgetType": "COST"
}'

# 3. Use Spot instances for non-production environments
# 4. Set CloudWatch logs retention to 7 days for cost savings
aws logs put-retention-policy --log-group-name /ecs/october-backend --retention-in-days 7

# 5. Clean up old ECR images automatically
aws ecr put-lifecycle-policy --repository-name october-backend --lifecycle-policy-text '{
  "rules": [
    {
      "rulePriority": 1,
      "selection": {"tagStatus": "untagged", "countType": "sinceImagePushed", "countUnit": "days", "countNumber": 7},
      "action": {"type": "expire"}
    }
  ]
}'
```

---

## 10) Create ECS Service

* Launch type: Fargate.
* Desired count: 1 (increase for production).
* Service discovery: optional.
* Configure auto-scaling policies later.
* For zero-downtime deployments choose **Deployment type: Rolling or Blue/Green** (if using CodeDeploy).

---

## 11) Secrets & config management

* **AWS Secrets Manager** — store DB passwords, API keys.
* **AWS Systems Manager Parameter Store** — for non-sensitive config or lower cost.

In ECS task definition you can reference secrets via `secrets` mapping (key: ENV_VAR, value: full ARN of secret or parameter).

---

## 12) CI/CD with GitHub Actions (sample)

Create `.github/workflows/ci-cd.yml` in your repo. The workflow builds the Docker image, pushes to ECR, and triggers an ECS service update.

```yaml
name: CI/CD
on:
  push:
    branches: [ main ]

jobs:
  build-and-deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - name: Login to ECR
        uses: aws-actions/amazon-ecr-login@v2
        with:
          region: ${{ secrets.AWS_REGION }}
      - name: Build and push image
        env:
          ECR_REG: ${{ secrets.ECR_REG }}
        run: |
          docker build -t $ECR_REG:latest .
          docker push $ECR_REG:latest
      - name: Deploy to ECS
        uses: aws-actions/amazon-ecs-deploy-task-definition@v1
        with:
          aws-region: ${{ secrets.AWS_REGION }}
          cluster: my-cluster
          service: my-service
          task-definition: taskdef.json
          wait-for-service-stability: true
```

* **Secrets to set in GitHub**: `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_REGION`, `ECR_REG` (like `<acct>.dkr.ecr.<region>.amazonaws.com/myapp:latest`).
* `taskdef.json`: either commit a templated task definition or generate/patch it in workflow before deploying.

---

## 13) DNS and HTTPS

* Use **AWS Certificate Manager (ACM)** to request an SSL cert (for `example.com` / `*.example.com`).
* Attach the cert to the ALB HTTPS listener (port 443) and redirect HTTP to HTTPS.
* Add Route53 record (A/ALIAS) pointing to your ALB.

---

## 14) Autoscaling & reliability

* Configure ECS Service Auto Scaling based on CPU / memory or custom CloudWatch metrics.
* Use at least 2 AZs for high availability.
* Set minimum desired count to 2 for redundancy.

---

## 15) Logging & monitoring

* **CloudWatch Logs**: use log group `/ecs/myapp` and set retention.
* **CloudWatch Alarms**: set alarms on high error rate (5xx), CPU, memory, target group unhealthy host count.
* Consider adding structured logs (JSON) and integrate with CloudWatch Logs Insights or a log sink (ElasticSearch, Datadog).

---

## 16) Database & migrations

* Use RDS with private subnets, security group that allows connections from task ENIs only.
* Run DB migrations as an ECS task or via CI step before deploy.
* For secrets, use Secrets Manager and map DB creds into env vars of the task.

---

## 17) Blue/Green / Canary deployments (optional)

* For safer deploys, use **CodeDeploy + ECS** for blue/green or configure canary deployments in your CI to gradually shift traffic.

---

## 18) Rollback plan

* Git tag/releases for production images.
* Keep previous image tags in ECR (don’t `latest` only) so you can redeploy previous tag quickly.
* Use ECS service update to roll back to an older task definition.

---

## 19) Cost & cleanup

* Fargate costs scale with vCPU & memory. Monitor usage.
* Remove unused ECR images, old log groups, and test clusters to save cost.

---

## 20) Terraform / IaC (recommended)

Create Terraform modules for:

* VPC + subnets
* ECS cluster + service + task definition
* ALB and listeners
* IAM roles
* ECR repository
* CloudWatch resources

This allows repeatable infra and safer PR review.

---

## 21) Example troubleshooting checklist

* Container fails health check -> check app binds to 0.0.0.0 and health path.
* Cannot pull image -> ECR auth or wrong image URI.
* 502 from ALB -> security groups (allow traffic), target group port mismatch, or container crash.
* Secrets not present -> check task role permissions and correct secret ARNs.

---

## 22) Handy commands

* Update ECS service (force new deployment):

  ```bash
  aws ecs update-service --cluster my-cluster --service my-service --force-new-deployment
  ```
* View logs:

  ```bash
  aws logs tail /ecs/myapp --follow
  ```

---

## 23) Next steps / advanced

* Add health-checked readiness probes in your app for graceful shutdown.
* Implement graceful shutdown in Go on SIGTERM.
* Use ECR lifecycle policies to garbage collect old images.
* Integrate Observability: distributed tracing (X-Ray / OpenTelemetry), metrics (Prometheus + CloudWatch or hosted solution).

---

If you want, I can now:

* Add a ready-to-copy **Terraform module** for ECS + ALB + ECR, or
* Provide a complete **GitHub Actions** workflow that templates the task definition dynamically and deploys, or
* Produce sample `taskdef.json` and `ecs-service` spec ready to paste to the AWS CLI.

---

## 🔒 Production Best Practices

### Security Checklist
- ✅ **API Keys**: Stored in Parameter Store with encryption
- ✅ **IAM Roles**: Principle of least privilege
- ✅ **Network**: Tasks in private subnets (if using ALB)
- ✅ **Container**: Non-root user in Docker
- ⚠️ **HTTPS**: Set up ACM certificate and ALB HTTPS listener
- ⚠️ **WAF**: Consider AWS WAF for API protection

### High Availability Setup
```bash
# 1. Run in multiple AZs
aws ecs update-service --cluster october-cluster --service october-backend-service --desired-count 2

# 2. Set up Application Load Balancer for high availability
# 3. Configure auto-scaling based on CPU/memory metrics
# 4. Set up health checks and circuit breakers
```

### Performance Optimization
```bash
# 1. Implement Redis caching layer
# 2. Optimize database queries and add indexes
# 3. Enable gzip compression
# 4. Use CDN for static assets
# 5. Implement connection pooling
```

---

## 🚀 Next Steps

### Production Enhancements
1. **Custom Domain**: Set up Route 53 and ACM certificate
2. **Load Balancer**: Add Application Load Balancer for HA
3. **Auto Scaling**: Configure ECS auto-scaling policies  
4. **Monitoring**: Set up comprehensive monitoring and alerting
5. **Backup**: Implement database backup strategy
6. **Security**: Add WAF, VPC endpoints, and security scanning

### Scaling Considerations
- **Horizontal Scaling**: Increase desired count for more instances
- **Vertical Scaling**: Increase CPU/memory allocation
- **Database Scaling**: Consider read replicas for read-heavy workloads
- **Caching**: Implement Redis/ElastiCache for improved performance

---

## 📞 Support and Resources

### Quick Commands Reference
```bash
# Check service status
aws ecs describe-services --cluster october-cluster --services october-backend-service

# View logs
aws logs tail /ecs/october-backend --follow

# Update service (force new deployment)
aws ecs update-service --cluster october-cluster --service october-backend-service --force-new-deployment

# Scale service
aws ecs update-service --cluster october-cluster --service october-backend-service --desired-count 2

# Check parameter store values
aws ssm get-parameters-by-path --path "/october" --recursive --with-decryption
```

### Useful Links
- [AWS ECS Documentation](https://docs.aws.amazon.com/ecs/)
- [AWS ECR Documentation](https://docs.aws.amazon.com/ecr/)
- [AWS Systems Manager Parameter Store](https://docs.aws.amazon.com/systems-manager/latest/userguide/systems-manager-parameter-store.html)
- [MongoDB Atlas](https://www.mongodb.com/atlas)

---

## 🎯 Deployment Summary

Your October Backend is now ready for AWS deployment with:

- ✅ **Production-ready Docker container**
- ✅ **Automated deployment script** (`aws/deploy.sh`)
- ✅ **CI/CD pipeline** (GitHub Actions)
- ✅ **Secure secrets management** (Parameter Store)
- ✅ **Monitoring and logging** (CloudWatch)
- ✅ **Cost optimization** strategies
- ✅ **Troubleshooting** guides

**Get started now:**
```bash
./aws/deploy.sh
```
