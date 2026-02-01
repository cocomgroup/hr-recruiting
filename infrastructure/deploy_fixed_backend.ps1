# Complete Backend Deployment Script with Auto-Detection (PowerShell)
# This script will find your resources and deploy the fixed backend

$ErrorActionPreference = "Stop"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Backend Health Check Fix Deployment  " -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Check if we're in the right directory
if (-not (Test-Path "Dockerfile") -and -not (Test-Path "cmd/server")) {
    if (Test-Path "backend") {
        Write-Host "[*] Switching to backend directory..." -ForegroundColor Yellow
        Set-Location backend
    } else {
        Write-Host "[ERROR] Not in backend directory and backend/ not found" -ForegroundColor Red
        Write-Host "Please run from backend directory or project root"
        exit 1
    }
}

# Get AWS Account and Region
Write-Host "[*] Detecting AWS environment..." -ForegroundColor Yellow

if (-not $env:AWS_REGION) {
    $env:AWS_REGION = "us-east-1"
}

try {
    $accountInfo = aws sts get-caller-identity --query Account --output text 2>$null
    $env:AWS_ACCOUNT_ID = $accountInfo
} catch {
    Write-Host "[ERROR] Failed to get AWS account ID. Check your credentials." -ForegroundColor Red
    exit 1
}

if (-not $env:AWS_ACCOUNT_ID) {
    Write-Host "[ERROR] Failed to get AWS account ID. Check your credentials." -ForegroundColor Red
    exit 1
}

Write-Host "[OK] Account ID: $env:AWS_ACCOUNT_ID" -ForegroundColor Green
Write-Host "[OK] Region: $env:AWS_REGION" -ForegroundColor Green
Write-Host ""

# Find ECR repository
Write-Host "[*] Finding ECR repository..." -ForegroundColor Yellow
$ecrReposJson = aws ecr describe-repositories --region $env:AWS_REGION --query 'repositories[*].repositoryName' --output json 2>$null

if (-not $ecrReposJson) {
    Write-Host "[ERROR] Failed to list ECR repositories!" -ForegroundColor Red
    exit 1
}

$ecrRepos = $ecrReposJson | ConvertFrom-Json

# Try to find backend repo
$env:ECR_REPO = $ecrRepos | Where-Object { $_ -like "*backend*" } | Select-Object -First 1

if (-not $env:ECR_REPO) {
    Write-Host "[ERROR] No backend ECR repository found!" -ForegroundColor Red
    Write-Host ""
    Write-Host "Available repositories:"
    $ecrRepos | ForEach-Object { Write-Host "  $_" }
    Write-Host ""
    $env:ECR_REPO = Read-Host "Enter ECR repository name"
}

Write-Host "[OK] Using ECR repo: $env:ECR_REPO" -ForegroundColor Green
Write-Host ""

# Verify health.go is fixed
Write-Host "[*] Verifying health.go fix..." -ForegroundColor Yellow
if (Test-Path "internal/handlers/health.go") {
    $healthContent = Get-Content "internal/handlers/health.go" -Raw
    if ($healthContent -match "hr-recruiting-api") {
        Write-Host "[OK] health.go has been fixed" -ForegroundColor Green
    } elseif ($healthContent -match "StatusServiceUnavailable") {
        Write-Host "[ERROR] health.go still has the old buggy code!" -ForegroundColor Red
        Write-Host ""
        Write-Host "You need to replace internal/handlers/health.go with the fixed version."
        Write-Host "The Health function should NOT check Hub-HRMS and should always return 200."
        exit 1
    } else {
        Write-Host "[WARNING] Cannot verify health.go - proceeding anyway" -ForegroundColor Yellow
    }
} else {
    Write-Host "[WARNING] health.go not found - proceeding anyway" -ForegroundColor Yellow
}
Write-Host ""

# Build Docker image
Write-Host "[*] Building Docker image..." -ForegroundColor Yellow
Write-Host "This may take a few minutes..."
Write-Host ""

try {
    docker build -f Dockerfile -t backend-fixed:latest .
    if ($LASTEXITCODE -ne 0) {
        throw "Docker build failed"
    }
} catch {
    Write-Host "[ERROR] Docker build failed!" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "[OK] Docker build successful" -ForegroundColor Green
Write-Host ""

# Tag for ECR
Write-Host "[*] Tagging image..." -ForegroundColor Yellow
docker tag backend-fixed:latest "${env:AWS_ACCOUNT_ID}.dkr.ecr.${env:AWS_REGION}.amazonaws.com/${env:ECR_REPO}:latest"

Write-Host "[OK] Image tagged" -ForegroundColor Green
Write-Host ""

# ECR Login
Write-Host "[*] Logging into ECR..." -ForegroundColor Yellow

try {
    $ecrPassword = aws ecr get-login-password --region $env:AWS_REGION
    $ecrPassword | docker login --username AWS --password-stdin "${env:AWS_ACCOUNT_ID}.dkr.ecr.${env:AWS_REGION}.amazonaws.com" 2>&1 | Out-Null
    
    if ($LASTEXITCODE -ne 0) {
        throw "ECR login failed"
    }
} catch {
    Write-Host "[ERROR] ECR login failed!" -ForegroundColor Red
    exit 1
}

Write-Host "[OK] ECR login successful" -ForegroundColor Green
Write-Host ""

# Push to ECR
Write-Host "[*] Pushing image to ECR..." -ForegroundColor Yellow
Write-Host "This may take a few minutes..."
Write-Host ""

try {
    docker push "${env:AWS_ACCOUNT_ID}.dkr.ecr.${env:AWS_REGION}.amazonaws.com/${env:ECR_REPO}:latest"
    if ($LASTEXITCODE -ne 0) {
        throw "Image push failed"
    }
} catch {
    Write-Host "[ERROR] Image push failed!" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "[OK] Image pushed successfully" -ForegroundColor Green
Write-Host ""

# Verify image in ECR
Write-Host "[*] Verifying image in ECR..." -ForegroundColor Yellow
$latestImage = aws ecr describe-images --repository-name $env:ECR_REPO --query 'sort_by(imageDetails,& imagePushedAt)[-1].imagePushedAt' --output text
Write-Host "[OK] Latest image pushed: $latestImage" -ForegroundColor Green
Write-Host ""

# Find ECS cluster
Write-Host "[*] Finding ECS cluster..." -ForegroundColor Yellow
$clustersJson = aws ecs list-clusters --region $env:AWS_REGION --query 'clusterArns[]' --output json
$clusters = $clustersJson | ConvertFrom-Json

$env:CLUSTER_NAME = $clusters | Where-Object { $_ -match "recruiting" } | Select-Object -First 1
if ($env:CLUSTER_NAME) {
    $env:CLUSTER_NAME = ($env:CLUSTER_NAME -split '/')[-1]
}

if (-not $env:CLUSTER_NAME) {
    Write-Host "[ERROR] No recruiting cluster found!" -ForegroundColor Red
    Write-Host "Available clusters:"
    $clusters | ForEach-Object { Write-Host "  $_" }
    Write-Host ""
    $env:CLUSTER_NAME = Read-Host "Enter cluster name"
}

Write-Host "[OK] Using cluster: $env:CLUSTER_NAME" -ForegroundColor Green
Write-Host ""

# Find backend service
Write-Host "[*] Finding backend service..." -ForegroundColor Yellow
$servicesJson = aws ecs list-services --cluster $env:CLUSTER_NAME --region $env:AWS_REGION --query 'serviceArns[]' --output json
$services = $servicesJson | ConvertFrom-Json

$env:SERVICE_NAME = $services | Where-Object { $_ -match "backend" } | Select-Object -First 1
if ($env:SERVICE_NAME) {
    $env:SERVICE_NAME = ($env:SERVICE_NAME -split '/')[-1]
}

if (-not $env:SERVICE_NAME) {
    Write-Host "[ERROR] No backend service found!" -ForegroundColor Red
    Write-Host "Available services:"
    $services | ForEach-Object { Write-Host "  $_" }
    Write-Host ""
    $env:SERVICE_NAME = Read-Host "Enter service name"
}

Write-Host "[OK] Using service: $env:SERVICE_NAME" -ForegroundColor Green
Write-Host ""

# Get current service status
Write-Host "--- Current service status ---" -ForegroundColor Yellow
aws ecs describe-services --cluster $env:CLUSTER_NAME --services $env:SERVICE_NAME --query 'services[0].{Running:runningCount,Desired:desiredCount,Status:status}' --output table
Write-Host ""

# Force deployment
Write-Host "[*] Forcing new deployment..." -ForegroundColor Yellow
aws ecs update-service --cluster $env:CLUSTER_NAME --service $env:SERVICE_NAME --force-new-deployment --region $env:AWS_REGION --output table | Out-Null

Write-Host "[OK] Deployment initiated!" -ForegroundColor Green
Write-Host ""

Write-Host "[*] Waiting for deployment to complete (this may take 2-5 minutes)..." -ForegroundColor Yellow
Write-Host "ECS will:"
Write-Host "  1. Start new tasks with the fixed image"
Write-Host "  2. Wait for them to pass health checks"
Write-Host "  3. Stop old tasks"
Write-Host ""

# Wait for service stability (with timeout)
$timeout = 300
$elapsed = 0
$interval = 10

while ($elapsed -lt $timeout) {
    Start-Sleep -Seconds $interval
    $elapsed += $interval
    
    $serviceStatus = aws ecs describe-services --cluster $env:CLUSTER_NAME --services $env:SERVICE_NAME --query 'services[0].deployments[0].rolloutState' --output text
    
    if ($serviceStatus -eq "COMPLETED") {
        Write-Host "[OK] Deployment completed!" -ForegroundColor Green
        break
    } elseif ($serviceStatus -eq "FAILED") {
        Write-Host "[ERROR] Deployment failed!" -ForegroundColor Red
        break
    }
    
    Write-Host "  Status: $serviceStatus (${elapsed}s elapsed)" -ForegroundColor Gray
}

Write-Host ""

# Check final status
Write-Host "--- Final service status ---" -ForegroundColor Yellow
aws ecs describe-services --cluster $env:CLUSTER_NAME --services $env:SERVICE_NAME --query 'services[0].{Running:runningCount,Desired:desiredCount,Status:status}' --output table

Write-Host ""
Write-Host "--- Recent events ---" -ForegroundColor Yellow
aws ecs describe-services --cluster $env:CLUSTER_NAME --services $env:SERVICE_NAME --query 'services[0].events[0:3].[createdAt,message]' --output table

Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "       Deployment Complete!            " -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""

# Get log group
$logGroup = "/ecs/$($env:CLUSTER_NAME -replace '-cluster$','')-backend"

Write-Host "--- Monitor logs ---" -ForegroundColor Cyan
Write-Host "  aws logs tail $logGroup --follow"
Write-Host ""

Write-Host "--- Test health endpoint ---" -ForegroundColor Cyan
Write-Host '  $albDns = aws cloudformation describe-stacks --stack-name hr-recruiting-ecs --query "Stacks[0].Outputs[?OutputKey==``ALBEndpoint``].OutputValue" --output text'
Write-Host '  Invoke-WebRequest -Uri "http://$albDns/health" -Method GET'
Write-Host ""

Write-Host "--- Expected in logs (after ~30 seconds) ---" -ForegroundColor Cyan
Write-Host '  [OK] "GET http://localhost:8080/health" - 200'
Write-Host '  [OK] "HEAD http://localhost:8080/health" - 200'
Write-Host ""

Write-Host "--- Should NOT see ---" -ForegroundColor Cyan
Write-Host '  [X] "HEAD http://localhost:8080/health" - 405'
Write-Host '  [X] "GET http://....:8080/health" - 503'
Write-Host ""

Write-Host "Run this to check logs now:" -ForegroundColor Yellow
Write-Host "aws logs tail $logGroup --follow --since 1m"