# push-initial-images.ps1
# Build and push initial Docker images to ECR before ECS services start

param(
    [Parameter(Mandatory=$false)]
    [string]$StackName = "hr-recruiting",
    
    [Parameter(Mandatory=$false)]
    [string]$Region = "us-east-1",
    
    [Parameter(Mandatory=$false)]
    [string]$FrontendDir = "./frontend",
    
    [Parameter(Mandatory=$false)]
    [string]$BackendDir = "./backend"
)

$ErrorActionPreference = "Stop"

Write-Host "=== Building and Pushing Initial Docker Images ===" -ForegroundColor Green
Write-Host ""

# ==========================================
# Get Stack Outputs
# ==========================================
Write-Host "Step 1: Getting ECR repository URLs..." -ForegroundColor Cyan

try {
    $outputs = aws cloudformation describe-stacks `
        --stack-name $StackName `
        --region $Region `
        --query 'Stacks[0].Outputs' `
        --output json | ConvertFrom-Json
    
    if ($LASTEXITCODE -ne 0) { throw }
}
catch {
    Write-Host "ERROR: Could not get stack outputs" -ForegroundColor Red
    Write-Host "Make sure the CloudFormation stack is deployed first:" -ForegroundColor Yellow
    Write-Host "  aws cloudformation describe-stacks --stack-name $StackName" -ForegroundColor Gray
    exit 1
}

$frontendRepo = ($outputs | Where-Object {$_.OutputKey -eq "ECRFrontendRepository"}).OutputValue
$backendRepo = ($outputs | Where-Object {$_.OutputKey -eq "ECRBackendRepository"}).OutputValue
$albEndpoint = ($outputs | Where-Object {$_.OutputKey -eq "ALBEndpoint"}).OutputValue

if (-not $frontendRepo -or -not $backendRepo) {
    Write-Host "ERROR: Could not find ECR repository URLs in stack outputs" -ForegroundColor Red
    exit 1
}

Write-Host "  Frontend Repository: $frontendRepo" -ForegroundColor Green
Write-Host "  Backend Repository: $backendRepo" -ForegroundColor Green
Write-Host ""

$registryUrl = $frontendRepo.Split('/')[0]

# ==========================================
# Login to ECR
# ==========================================
Write-Host "Step 2: Logging into Amazon ECR..." -ForegroundColor Cyan

try {
    aws ecr get-login-password --region $Region | docker login --username AWS --password-stdin $registryUrl
    if ($LASTEXITCODE -ne 0) { throw }
    Write-Host "  Successfully logged into ECR" -ForegroundColor Green
}
catch {
    Write-Host "ERROR: Failed to login to ECR" -ForegroundColor Red
    exit 1
}
Write-Host ""

# ==========================================
# Build and Push Frontend
# ==========================================
Write-Host "Step 3: Building frontend image..." -ForegroundColor Cyan

if (-not (Test-Path $FrontendDir)) {
    Write-Host "ERROR: Frontend directory not found: $FrontendDir" -ForegroundColor Red
    exit 1
}

Push-Location $FrontendDir

# Create production .env
$apiUrl = "http://$albEndpoint/api"
Write-Host "  Configuring API URL: $apiUrl" -ForegroundColor Yellow

@"
VITE_API_URL=$apiUrl
VITE_ENVIRONMENT=production
"@ | Out-File -FilePath ".env" -Encoding UTF8

Write-Host "  Building Docker image..." -ForegroundColor Yellow
docker build -t hr-recruiting-frontend:latest .

if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Frontend Docker build failed" -ForegroundColor Red
    Pop-Location
    exit 1
}

Write-Host "  Tagging image for ECR..." -ForegroundColor Yellow
docker tag hr-recruiting-frontend:latest ${frontendRepo}:latest

Write-Host "  Pushing to ECR..." -ForegroundColor Yellow
docker push ${frontendRepo}:latest

if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Failed to push frontend image" -ForegroundColor Red
    Pop-Location
    exit 1
}

Write-Host "  Frontend image pushed successfully" -ForegroundColor Green
Pop-Location
Write-Host ""

# ==========================================
# Build and Push Backend
# ==========================================
Write-Host "Step 4: Building backend image..." -ForegroundColor Cyan

if (-not (Test-Path $BackendDir)) {
    Write-Host "ERROR: Backend directory not found: $BackendDir" -ForegroundColor Red
    exit 1
}

Push-Location $BackendDir

Write-Host "  Building Docker image..." -ForegroundColor Yellow
docker build -t hr-recruiting-backend:latest .

if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Backend Docker build failed" -ForegroundColor Red
    Pop-Location
    exit 1
}

Write-Host "  Tagging image for ECR..." -ForegroundColor Yellow
docker tag hr-recruiting-backend:latest ${backendRepo}:latest

Write-Host "  Pushing to ECR..." -ForegroundColor Yellow
docker push ${backendRepo}:latest

if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Failed to push backend image" -ForegroundColor Red
    Pop-Location
    exit 1
}

Write-Host "  Backend image pushed successfully" -ForegroundColor Green
Pop-Location
Write-Host ""

# ==========================================
# Update ECS Services
# ==========================================
Write-Host "Step 5: Triggering ECS service updates..." -ForegroundColor Cyan

$cluster = ($outputs | Where-Object {$_.OutputKey -eq "ECSCluster"}).OutputValue
$frontendService = ($outputs | Where-Object {$_.OutputKey -eq "FrontendService"}).OutputValue
$backendService = ($outputs | Where-Object {$_.OutputKey -eq "BackendService"}).OutputValue

Write-Host "  Updating frontend service..." -ForegroundColor Yellow
aws ecs update-service `
    --cluster $cluster `
    --service $frontendService `
    --force-new-deployment `
    --region $Region `
    --output json | Out-Null

if ($LASTEXITCODE -eq 0) {
    Write-Host "  Frontend service update initiated" -ForegroundColor Green
}

Write-Host "  Updating backend service..." -ForegroundColor Yellow
aws ecs update-service `
    --cluster $cluster `
    --service $backendService `
    --force-new-deployment `
    --region $Region `
    --output json | Out-Null

if ($LASTEXITCODE -eq 0) {
    Write-Host "  Backend service update initiated" -ForegroundColor Green
}

Write-Host ""

# ==========================================
# Wait for Services
# ==========================================
Write-Host "Step 6: Waiting for services to stabilize..." -ForegroundColor Cyan
Write-Host "  This may take 3-5 minutes..." -ForegroundColor Yellow
Write-Host ""

Write-Host "  Waiting for backend service..." -ForegroundColor Yellow
aws ecs wait services-stable `
    --cluster $cluster `
    --services $backendService `
    --region $Region

if ($LASTEXITCODE -eq 0) {
    Write-Host "  Backend service is stable" -ForegroundColor Green
}
else {
    Write-Host "  Backend service stabilization timed out" -ForegroundColor Yellow
    Write-Host "  Check ECS console for details" -ForegroundColor Yellow
}

Write-Host "  Waiting for frontend service..." -ForegroundColor Yellow
aws ecs wait services-stable `
    --cluster $cluster `
    --services $frontendService `
    --region $Region

if ($LASTEXITCODE -eq 0) {
    Write-Host "  Frontend service is stable" -ForegroundColor Green
}
else {
    Write-Host "  Frontend service stabilization timed out" -ForegroundColor Yellow
    Write-Host "  Check ECS console for details" -ForegroundColor Yellow
}

Write-Host ""

# ==========================================
# Verify Deployment
# ==========================================
Write-Host "Step 7: Verifying deployment..." -ForegroundColor Cyan

# Get running task counts
$backendTasks = aws ecs describe-services `
    --cluster $cluster `
    --services $backendService `
    --region $Region `
    --query 'services[0].runningCount' `
    --output text

$frontendTasks = aws ecs describe-services `
    --cluster $cluster `
    --services $frontendService `
    --region $Region `
    --query 'services[0].runningCount' `
    --output text

Write-Host "  Backend tasks running: $backendTasks" -ForegroundColor Cyan
Write-Host "  Frontend tasks running: $frontendTasks" -ForegroundColor Cyan
Write-Host ""

# Test endpoints
Write-Host "  Testing backend health..." -ForegroundColor Yellow
try {
    $healthResponse = Invoke-WebRequest -Uri "http://$albEndpoint/health" -UseBasicParsing -TimeoutSec 10
    if ($healthResponse.StatusCode -eq 200) {
        Write-Host "  Backend health check: PASSED" -ForegroundColor Green
    }
}
catch {
    Write-Host "  Backend health check: FAILED (may still be starting)" -ForegroundColor Yellow
}

Write-Host "  Testing frontend..." -ForegroundColor Yellow
try {
    $frontendResponse = Invoke-WebRequest -Uri "http://$albEndpoint/" -UseBasicParsing -TimeoutSec 10
    if ($frontendResponse.StatusCode -eq 200) {
        Write-Host "  Frontend: ACCESSIBLE" -ForegroundColor Green
    }
}
catch {
    Write-Host "  Frontend: NOT ACCESSIBLE (may still be starting)" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "============================================" -ForegroundColor Green
Write-Host "       Initial Deployment Complete!        " -ForegroundColor Green
Write-Host "============================================" -ForegroundColor Green
Write-Host ""

Write-Host "Application URL: http://$albEndpoint" -ForegroundColor Cyan
Write-Host "API Endpoint: http://$albEndpoint/api" -ForegroundColor Cyan
Write-Host ""

Write-Host "Next Steps:" -ForegroundColor Yellow
Write-Host "  1. Open application URL in browser"
Write-Host "  2. For future deployments, use: .\deploy-ecs.ps1"
Write-Host "  3. View logs: aws logs tail /ecs/$StackName-backend --follow"
Write-Host ""