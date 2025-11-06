# October Backend - AWS Deployment Guide

This guide walks you through deploying the October backend to AWS ECS Fargate using the automated GitHub Actions workflow.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Initial AWS Setup](#initial-aws-setup)
- [GitHub Configuration](#github-configuration)
- [Deployment Process](#deployment-process)
- [Custom Domain Setup](#custom-domain-setup)
- [Monitoring and Troubleshooting](#monitoring-and-troubleshooting)
- [Manual Deployment Steps](#manual-deployment-steps)

---

## Prerequisites

Before deploying, ensure you have:

1. **AWS Account** with appropriate permissions
2. **GitHub Account** with repository access
3. **AWS CLI** installed locally (for initial setup)
4. **Docker** installed locally (for testing builds)
5. **Domain Name** (optional, for custom domain setup)

### Required AWS Permissions

Your AWS IAM user needs permissions for:
- ECS (Elastic Container Service)
- ECR (Elastic Container Registry)
- EC2 (for VPC, Security Groups, Subnets)
- ELB/ALB (Application Load Balancer)
- IAM (for role creation)
- SSM Parameter Store (for secrets)
- CloudWatch Logs
- Route53 (for custom domain)
- ACM (AWS Certificate Manager for HTTPS)

---

## Initial AWS Setup

### Step 1: Create AWS Access Keys

1. Sign in to the [AWS Console](https://console.aws.amazon.com)
2. Navigate to **IAM** → **Users** → Select your user
3. Go to **Security credentials** tab
4. Click **Create access key**
5. Choose **Application running outside AWS**
6. Save the **Access Key ID** and **Secret Access Key** securely

### Step 2: Configure AWS Secrets in SSM Parameter Store

The application requires several secrets stored in AWS Systems Manager Parameter Store. Create these parameters in the **us-east-1** region:

```bash
# Database connection string
aws ssm put-parameter \
  --name "/october/DATABASE_URI" \
  --value "your-mongodb-connection-string" \
  --type "SecureString" \
  --region us-east-1

# OpenAI API Key
aws ssm put-parameter \
  --name "/october/OPENAI_API_KEY" \
  --value "your-openai-api-key" \
  --type "SecureString" \
  --region us-east-1

# Google Custom Search API Key
aws ssm put-parameter \
  --name "/october/CUSTOM_SEARCH_API_KEY" \
  --value "your-google-api-key" \
  --type "SecureString" \
  --region us-east-1

# Google Custom Search Engine ID
aws ssm put-parameter \
  --name "/october/CUSTOM_SEARCH_ENGINE_ID" \
  --value "your-search-engine-id" \
  --type "SecureString" \
  --region us-east-1

# Finnhub API Key
aws ssm put-parameter \
  --name "/october/FINNHUB_API_KEY" \
  --value "your-finnhub-api-key" \
  --type "SecureString" \
  --region us-east-1
```

**Verify parameters:**
```bash
aws ssm get-parameters-by-path --path "/october" --region us-east-1
```

### Step 3: Verify Default VPC and Subnets

The deployment uses the default VPC and public subnets. Verify they exist:

```bash
# Check default VPC
aws ec2 describe-vpcs --filters "Name=is-default,Values=true" --region us-east-1

# Check public subnets (map-public-ip-on-launch=true)
aws ec2 describe-subnets \
  --filters "Name=vpc-id,Values=<YOUR_VPC_ID>" "Name=map-public-ip-on-launch,Values=true" \
  --region us-east-1
```

If no default VPC exists, create one:
```bash
aws ec2 create-default-vpc --region us-east-1
```

---

## GitHub Configuration

### Step 4: Add GitHub Secrets

1. Go to your GitHub repository
2. Navigate to **Settings** → **Secrets and variables** → **Actions**
3. Click **New repository secret** and add the following:

| Secret Name | Value | Description |
|-------------|-------|-------------|
| `AWS_ACCESS_KEY_ID` | Your AWS access key ID | From Step 1 |
| `AWS_SECRET_ACCESS_KEY` | Your AWS secret access key | From Step 1 |
| `OCTOBER_CUSTOM_DOMAIN` | `oct.circuit-x.com` | Custom domain (optional) |
| `HOSTED_ZONE_NAME` | `circuit-x.com` | Route53 hosted zone (optional) |

**Note:** The custom domain secrets are only needed if you want to configure a custom domain with HTTPS.

---

## Deployment Process

### Step 5: Trigger Deployment

The workflow automatically deploys when you push to the `main` branch:

```bash
git add .
git commit -m "Deploy to AWS"
git push origin main
```

### Step 6: Monitor Deployment

1. Go to your GitHub repository
2. Click **Actions** tab
3. Select the running workflow
4. Watch the deployment progress through each job

The workflow consists of two jobs:

#### **Job 1: Deploy**

1. **Checkout repo** - Clones the repository
2. **Configure AWS credentials** - Authenticates with AWS
3. **Get AWS Account ID** - Retrieves account ID for resource ARNs
4. **Build and push Docker image** - Builds container and pushes to ECR
5. **Ensure IAM execution role exists** - Creates `ecsTaskExecutionRole` with SSM/KMS/CloudWatch permissions
6. **Prepare ECS task definition** - Injects image URI and role ARNs
7. **Register ECS task definition** - Registers new task version
8. **Get VPC and Subnet IDs** - Discovers default VPC and public subnets
9. **Setup networking and load balancer** - Creates:
   - Security groups (`october-alb-sg`, `october-tasks-sg`)
   - Application Load Balancer (`october-alb`)
   - Target Group (`october-backend-tg`) with health check on `/health`
   - HTTP listener on port 80
10. **Ensure ECS cluster exists** - Creates `october-cluster` if needed
11. **Ensure ECS service-linked role exists** - Creates `AWSServiceRoleForECS`
12. **Create or update ECS service** - Deploys `october-backend-service` with:
    - Fargate launch type
    - 1 desired task
    - Network configuration (public subnets + security group)
    - Load balancer integration
13. **Output ALB DNS** - Displays backend URL

#### **Job 2: Domain** (if custom domain configured)

1. **Configure AWS credentials** - Re-authenticates for domain job
2. **Resolve ALB info** - Gets ALB ARN, DNS, and hosted zone ID
3. **Ensure Route53 hosted zone exists** - Validates DNS zone
4. **Ensure ACM certificate** - Requests/validates SSL certificate via DNS
5. **Ensure HTTPS listener** - Adds listener on port 443 with certificate
6. **Create/Update Route53 ALIAS** - Points custom domain to ALB

### Step 7: Verify Deployment

Once the workflow completes:

1. **Check ALB DNS** (from workflow output):
   ```
   http://<ALB_DNS>/health
   ```

2. **Check Custom Domain** (if configured):
   ```
   http://oct.circuit-x.com/health
   https://oct.circuit-x.com/health
   ```

3. **Verify health endpoint response**:
   ```json
   {
     "status": "healthy",
     "timestamp": "2025-11-07T12:00:00Z"
   }
   ```

---

## Custom Domain Setup

### Step 8: Configure Custom Domain (Optional)

To use a custom domain like `oct.circuit-x.com`:

#### Option A: Using Route53

1. **Register domain in Route53** or transfer existing domain
2. **Create hosted zone** (if not exists):
   ```bash
   aws route53 create-hosted-zone --name circuit-x.com --caller-reference $(date +%s)
   ```
3. **Add GitHub secrets** (from Step 4):
   - `OCTOBER_CUSTOM_DOMAIN`: `oct.circuit-x.com`
   - `HOSTED_ZONE_NAME`: `circuit-x.com`
4. **Push to main** - Workflow will automatically:
   - Request ACM certificate for `oct.circuit-x.com`
   - Create DNS validation record
   - Wait for certificate issuance (up to 10 minutes)
   - Add HTTPS listener on ALB port 443
   - Create Route53 ALIAS record pointing to ALB

#### Option B: Using External DNS Provider

1. **Complete deployment** without custom domain secrets
2. **Get ALB DNS** from workflow output
3. **Add CNAME or ALIAS record** in your DNS provider:
   ```
   oct.circuit-x.com → <ALB_DNS> (e.g., october-alb-123456.us-east-1.elb.amazonaws.com)
   ```
4. **Request ACM certificate manually**:
   ```bash
   aws acm request-certificate --domain-name oct.circuit-x.com --validation-method DNS
   ```
5. **Add validation CNAME** in your DNS provider
6. **Wait for certificate issuance**
7. **Add HTTPS listener manually**:
   ```bash
   aws elbv2 create-listener \
     --load-balancer-arn <ALB_ARN> \
     --protocol HTTPS --port 443 \
     --certificates CertificateArn=<CERT_ARN> \
     --default-actions Type=forward,TargetGroupArn=<TG_ARN>
   ```

### Certificate Validation Timeline

- **DNS Validation**: Usually 5-10 minutes after DNS record creation
- **Total Wait Time**: Workflow polls for up to 10 minutes (20 × 30s)
- **Fallback**: If certificate not issued during workflow, HTTPS will work once ACM completes validation

---

## Monitoring and Troubleshooting

### Viewing Logs

**CloudWatch Logs:**
```bash
# View log streams
aws logs describe-log-streams --log-group-name /ecs/october-backend --region us-east-1

# Tail latest logs
aws logs tail /ecs/october-backend --follow --region us-east-1
```

**In AWS Console:**
1. Go to **CloudWatch** → **Log groups**
2. Select `/ecs/october-backend`
3. View log streams (one per task)

### Checking ECS Service Status

```bash
# Service status
aws ecs describe-services \
  --cluster october-cluster \
  --services october-backend-service \
  --region us-east-1

# Running tasks
aws ecs list-tasks --cluster october-cluster --region us-east-1

# Task details
aws ecs describe-tasks \
  --cluster october-cluster \
  --tasks <TASK_ARN> \
  --region us-east-1
```

### Common Issues

#### Issue: Tasks failing to start with "ResourceInitializationError"

**Cause:** ECS task cannot access SSM parameters

**Solution:**
- Verify SSM parameters exist: `aws ssm get-parameters-by-path --path "/october"`
- Check `ecsTaskExecutionRole` has inline policy `ecsTaskExecutionRoleParameterRead`
- Ensure KMS permissions if using SecureString parameters

#### Issue: ALB health checks failing

**Cause:** Tasks not responding on `/health` endpoint

**Solution:**
- Check task logs in CloudWatch
- Verify security group `october-tasks-sg` allows traffic from `october-alb-sg` on port 8080
- Confirm application listens on `0.0.0.0:8080` (not `localhost`)
- Test directly: `curl http://<TASK_IP>:8080/health`

#### Issue: "No public subnets found"

**Cause:** VPC has no subnets with `map-public-ip-on-launch=true`

**Solution:**
```bash
# Enable auto-assign public IP on subnet
aws ec2 modify-subnet-attribute \
  --subnet-id <SUBNET_ID> \
  --map-public-ip-on-launch
```

#### Issue: Certificate stuck in PENDING_VALIDATION

**Cause:** DNS validation record not created or propagated

**Solution:**
- Check Route53 for validation CNAME record
- Verify hosted zone matches certificate domain
- Wait 5-10 minutes for DNS propagation
- Manually verify: `dig _<validation_name>.circuit-x.com CNAME`

#### Issue: HTTPS listener not created

**Cause:** Certificate not issued before listener creation

**Solution:**
- Workflow continues even if cert pending
- HTTPS will work once certificate issues
- Manually check cert status: `aws acm describe-certificate --certificate-arn <ARN>`
- Re-run workflow after certificate issues

### Health Check Endpoints

- **`/health`**: Full health check (may depend on database/external services)
- **`/liveness`**: Minimal liveness probe (application running check)

**Recommendation:** For ALB target group health checks, use `/liveness` to avoid false negatives from temporary database issues.

To update target group health check:
```bash
aws elbv2 modify-target-group \
  --target-group-arn <TG_ARN> \
  --health-check-path /liveness \
  --region us-east-1
```

### Scaling

**Manual scaling:**
```bash
# Update desired count
aws ecs update-service \
  --cluster october-cluster \
  --service october-backend-service \
  --desired-count 3 \
  --region us-east-1
```

**Auto-scaling:** Configure Application Auto Scaling for ECS service based on CPU/memory metrics.

---

## Manual Deployment Steps

If you need to deploy manually without GitHub Actions:

### 1. Build and Push Docker Image

```bash
# Authenticate with ECR
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin <ACCOUNT_ID>.dkr.ecr.us-east-1.amazonaws.com

# Create ECR repository
aws ecr create-repository --repository-name october-backend --region us-east-1

# Build image
docker build -t october-backend:latest .

# Tag image
docker tag october-backend:latest <ACCOUNT_ID>.dkr.ecr.us-east-1.amazonaws.com/october-backend:latest

# Push image
docker push <ACCOUNT_ID>.dkr.ecr.us-east-1.amazonaws.com/october-backend:latest
```

### 2. Create IAM Execution Role

```bash
# Create role
aws iam create-role \
  --role-name ecsTaskExecutionRole \
  --assume-role-policy-document file://trust-policy.json

# Attach AWS managed policy
aws iam attach-role-policy \
  --role-name ecsTaskExecutionRole \
  --policy-arn arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy

# Add SSM/KMS permissions
aws iam put-role-policy \
  --role-name ecsTaskExecutionRole \
  --policy-name ecsTaskExecutionRoleParameterRead \
  --policy-document file://ssm-policy.json
```

### 3. Register Task Definition

```bash
# Update task definition with image URI
jq --arg img "<ACCOUNT_ID>.dkr.ecr.us-east-1.amazonaws.com/october-backend:latest" \
   '.containerDefinitions[0].image=$img' \
   aws/ecs-task-definition.json > /tmp/task-def.json

# Register
aws ecs register-task-definition --cli-input-json file:///tmp/task-def.json
```

### 4. Create ECS Cluster

```bash
aws ecs create-cluster --cluster-name october-cluster --region us-east-1
```

### 5. Create Security Groups, ALB, and Target Group

Follow the commands in `.github/workflows/deploy.yml` steps 9-10, or use the AWS Console.

### 6. Create ECS Service

```bash
aws ecs create-service \
  --cluster october-cluster \
  --service-name october-backend-service \
  --task-definition october-backend \
  --desired-count 1 \
  --launch-type FARGATE \
  --network-configuration "awsvpcConfiguration={subnets=[<SUBNET_IDS>],securityGroups=[<TASK_SG>],assignPublicIp=ENABLED}" \
  --load-balancers "targetGroupArn=<TG_ARN>,containerName=october-backend,containerPort=8080" \
  --region us-east-1
```

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                         Internet                                 │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             │ HTTP/HTTPS
                             ▼
                  ┌──────────────────────┐
                  │  Route53 (optional)  │
                  │  oct.circuit-x.com   │
                  └──────────┬───────────┘
                             │
                             │ ALIAS
                             ▼
            ┌────────────────────────────────┐
            │   Application Load Balancer    │
            │        october-alb             │
            │  - Port 80 (HTTP)              │
            │  - Port 443 (HTTPS/ACM cert)   │
            └────────────┬───────────────────┘
                         │
                         │ Target Group: october-backend-tg
                         │ Health Check: /health or /liveness
                         ▼
         ┌──────────────────────────────────────┐
         │        ECS Service (Fargate)         │
         │     october-backend-service          │
         │  ┌────────────────────────────────┐  │
         │  │  Task: october-backend         │  │
         │  │  - Container Port: 8080        │  │
         │  │  - CPU: 512, Memory: 1024      │  │
         │  │  - Secrets from SSM            │  │
         │  │  - Logs to CloudWatch          │  │
         │  └────────────────────────────────┘  │
         └──────────────────────────────────────┘
                         │
                         ├─── SSM Parameter Store (/october/*)
                         ├─── CloudWatch Logs (/ecs/october-backend)
                         └─── MongoDB (external DATABASE_URI)
```

---

## Cost Estimates

**Monthly AWS costs** (us-east-1, approximate):

- **ECS Fargate (1 task, 0.5 vCPU, 1 GB RAM)**: ~$15-20
- **Application Load Balancer**: ~$16-20
- **ECR Storage (< 10 GB)**: ~$1
- **CloudWatch Logs (< 5 GB)**: ~$0.50
- **Data Transfer (first 100 GB free)**: $0
- **Route53 Hosted Zone (optional)**: $0.50/month
- **ACM Certificate**: Free

**Total**: ~$35-45/month for basic setup

---

## Security Best Practices

1. **Use SecureString** for all SSM parameters
2. **Rotate secrets** regularly (AWS Secrets Manager for automation)
3. **Enable container insights** for enhanced monitoring
4. **Use least-privilege IAM policies** - review and scope down permissions
5. **Enable ALB access logs** for audit trail
6. **Configure WAF** (Web Application Firewall) on ALB for production
7. **Use HTTPS only** - add redirect rule from HTTP to HTTPS
8. **Enable VPC Flow Logs** for network monitoring
9. **Set up CloudWatch alarms** for task failures, high CPU/memory
10. **Enable AWS Config** for compliance monitoring

---

## Next Steps

After successful deployment:

1. **Set up CI/CD** for automatic deployments on code changes (already configured)
2. **Configure auto-scaling** based on CPU/memory metrics
3. **Add CloudWatch alarms** for monitoring
4. **Set up backup strategy** for MongoDB
5. **Implement blue/green deployments** for zero-downtime updates
6. **Add CloudFront CDN** for better performance
7. **Configure custom error pages** on ALB
8. **Enable AWS X-Ray** for distributed tracing

---

## Support and Resources

- **AWS ECS Documentation**: https://docs.aws.amazon.com/ecs/
- **GitHub Actions**: https://docs.github.com/en/actions
- **AWS CLI Reference**: https://docs.aws.amazon.com/cli/
- **Docker Documentation**: https://docs.docker.com/

For issues or questions, check the GitHub repository issues section.

---

**Last Updated**: November 7, 2025
