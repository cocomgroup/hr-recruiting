# HR-Recruiting ECS Deployment Script (PowerShell)
# Automated deployment to AWS with Docker image building and pushing

[CmdletBinding()]
param(
    [string]$StackName = "hr-recruiting-dev",
    [string]$Region = "us-east-1",
    [switch]$SkipImages,
    [switch]$SkipBuild,
    [string]$HubHRMSGraphQLURL,
    [string]$HubHRMSAPIKey,
    [string]$SendGridAPIKey,
    [string]$EmailFrom = "recruiting@company.com",
    [string]$S3Bucket = "",
    [switch]$Help,
    [switch]$Update,
    [switch]$Delete,
    [switch]$Status
)

# Configuration
$Environment = "development"
$BackendECRRepo = "$StackName-backend"
$FrontendECRRepo = "$StackName-frontend"
$CloudFormationTemplate = "infrastructure/cloudformation-ecs-shared-vpc.yaml"

# Helper Functions
function Write-Info {
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor Cyan
}

function Write-Success {
    param([string]$Message)
    Write-Host "[SUCCESS] $Message" -ForegroundColor Green
}

function Write-Warning {
    param([string]$Message)
    Write-Host "[WARNING] $Message" -ForegroundColor Yellow
}

function Write-Error {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor Red
}

function Write-Step {
    param([string]$Message)
    Write-Host ""
    Write-Host "==================================================" -ForegroundColor Blue
    Write-Host "  $Message" -ForegroundColor Blue
    Write-Host "==================================================" -ForegroundColor Blue
}

function Show-Help {
    Write-Host ""
    Write-Host "HR-Recruiting ECS Deployment Script"
    Write-Host "====================================="
    Write-Host ""
    Write-Host "Usage: .\deploy-ecs.ps1 [options]"
    Write-Host ""
    Write-Host "Options:"
    Write-Host "  -StackName NAME            Stack name (default: hr-recruiting)"
    Write-Host "  -Region REGION             AWS region (default: us-east-1)"
    Write-Host "  -SkipImages                Skip building and pushing Docker images"
    Write-Host "  -SkipBuild                 Skip building images (use existing)"
    Write-Host "  -HubHRMSGraphQLURL URL     Hub-HRMS GraphQL endpoint"
    Write-Host "  -HubHRMSAPIKey KEY         Hub-HRMS API key"
    Write-Host "  -SendGridAPIKey KEY        SendGrid API key for emails"
    Write-Host "  -EmailFrom EMAIL           From email address"
    Write-Host "  -S3Bucket NAME             S3 bucket for resume uploads"
    Write-Host "  -Update                    Update existing stack"
    Write-Host "  -Delete                    Delete stack"
    Write-Host "  -Status                    Show stack status"
    Write-Host "  -Help                      Show this help"
    Write-Host ""
    Write-Host "Examples:"
    Write-Host "  .\deploy-ecs.ps1"
    Write-Host "  .\deploy-ecs.ps1 -Update"
    Write-Host "  .\deploy-ecs.ps1 -Status"
    Write-Host "  .\deploy-ecs.ps1 -SkipBuild -Update"
    Write-Host ""
    exit 0
}

function Test-Prerequisites {
    Write-Step "Checking Prerequisites"
    
    $missing = @()
    
    # Check AWS CLI
    try {
        $awsVersion = aws --version 2>&1
        Write-Success "AWS CLI found: $awsVersion"
    } catch {
        $missing += "AWS CLI"
    }
    
    # Check Docker (only if not skipping images)
    if (-not $SkipImages -and -not $SkipBuild) {
        try {
            $dockerVersion = docker --version 2>&1
            Write-Success "Docker found: $dockerVersion"
        } catch {
            $missing += "Docker"
        }
    }
    
    # Check AWS credentials
    try {
        $identity = aws sts get-caller-identity 2>&1 | ConvertFrom-Json
        Write-Success "AWS Account: $($identity.Account)"
        Write-Success "AWS User: $($identity.Arn)"
        $script:AccountId = $identity.Account
    } catch {
        $missing += "AWS credentials"
    }
    
    # Check if CloudFormation template exists
    if (-not (Test-Path $CloudFormationTemplate)) {
        Write-Error "CloudFormation template not found: $CloudFormationTemplate"
        $missing += "CloudFormation template"
    } else {
        Write-Success "CloudFormation template found"
    }
    
    if ($missing.Count -gt 0) {
        Write-Error "Missing prerequisites: $($missing -join ', ')"
        Write-Host ""
        Write-Host "Install missing components:"
        if ($missing -contains "AWS CLI") {
            Write-Host "  AWS CLI: https://aws.amazon.com/cli/"
        }
        if ($missing -contains "Docker") {
            Write-Host "  Docker: https://www.docker.com/products/docker-desktop"
        }
        if ($missing -contains "AWS credentials") {
            Write-Host "  Run: aws configure"
        }
        exit 1
    }
    
    Write-Success "All prerequisites met!"
}

function Get-RandomString {
    param([int]$Length = 32)
    $chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
    -join ((1..$Length) | ForEach-Object { $chars[(Get-Random -Maximum $chars.Length)] })
}

function New-ECRRepository {
    param([string]$RepoName)
    
    Write-Info "Checking ECR repository: $RepoName"
    
    try {
        $repo = aws ecr describe-repositories --repository-names $RepoName --region $Region 2>&1 | ConvertFrom-Json
        Write-Success "Repository exists: $RepoName"
        return $repo.repositories[0].repositoryUri
    } catch {
        Write-Info "Creating ECR repository: $RepoName"
        $repo = aws ecr create-repository --repository-name $RepoName --region $Region --image-scanning-configuration scanOnPush=true | ConvertFrom-Json
        Write-Success "Repository created: $RepoName"
        return $repo.repository.repositoryUri
    }
}

function Build-AndPushImage {
    param(
        [string]$ImageName,
        [string]$DockerfilePath,
        [string]$ContextPath,
        [string]$ECRUri,
        [hashtable]$BuildArgs = @{}
    )
    
    Write-Info "Building $ImageName image..."
    
    # Build docker build command with build args
    $buildCmd = "docker build -t $ImageName -f $DockerfilePath"
    
    foreach ($key in $BuildArgs.Keys) {
        $buildCmd += " --build-arg $key=$($BuildArgs[$key])"
    }
    
    $buildCmd += " $ContextPath"
    
    # Execute build
    Invoke-Expression "$buildCmd 2>&1" | Out-Null
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Failed to build $ImageName image"
        exit 1
    }
    Write-Success "Built $ImageName image"
    
    # Tag for ECR
    $imageTag = "${ECRUri}:latest"
    docker tag $ImageName $imageTag 2>&1 | Out-Null
    Write-Success "Tagged as $imageTag"
    
    # Login to ECR
    Write-Info "Logging into ECR..."
    $loginCmd = aws ecr get-login-password --region $Region
    $loginCmd | docker login --username AWS --password-stdin $ECRUri.Split('/')[0] 2>&1 | Out-Null
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Failed to login to ECR"
        exit 1
    }
    
    # Push to ECR
    Write-Info "Pushing $ImageName to ECR..."
    docker push $imageTag 2>&1 | Out-Null
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Failed to push $ImageName image"
        exit 1
    }
    Write-Success "Pushed $ImageName to ECR"
    
    Write-Output $imageTag
}

function Get-StackOutputValue {
    param(
        [string]$OutputKey
    )
    
    try {
        $stack = aws cloudformation describe-stacks --stack-name $StackName --region $Region 2>&1 | ConvertFrom-Json
        $output = $stack.Stacks[0].Outputs | Where-Object { $_.OutputKey -eq $OutputKey }
        return $output.OutputValue
    } catch {
        return $null
    }
}

function Deploy-Stack {
    param(
        [string]$Action = "create"
    )
    
    Write-Step "$(if ($Action -eq 'create') { 'Deploying' } else { 'Updating' }) CloudFormation Stack"
    
    # Check if stack parameters are needed
    $needsParams = $false
    if ($Action -eq "create") {
        $needsParams = $true
    }
    
    # Build parameters array
    $parametersArray = @(
        @{
            ParameterKey = "EnvironmentName"
            ParameterValue = $Environment
        }
    )
    
    # Only add VPC parameters for create (not for update)
    if ($Action -eq "create") {
        # Get VPC info from Hub-HRMS or user input
        Write-Info "Stack requires VPC parameters for initial creation"
        Write-Info "Run .\get-hubhrms-vpc-info.ps1 to find your VPC and subnet IDs"
        
        # Try to get from existing hr-recruiting stack if updating
        $existingVpcId = Get-StackOutputValue -OutputKey "VPCId"
        
        if (-not $existingVpcId) {
            Write-Error "VPC parameters required for stack creation"
            Write-Host ""
            Write-Host "Run this command first:"
            Write-Host "  .\get-hubhrms-vpc-info.ps1"
            Write-Host ""
            Write-Host "Then use the output to deploy with parameters:"
            Write-Host "  See QUICKSTART-SHARED-VPC.md for details"
            exit 1
        }
    }
    
    # Convert parameters to JSON
    $parametersJson = $parametersArray | ConvertTo-Json -Compress
    $parametersJson | Out-File -FilePath "stack-parameters.json" -Encoding UTF8
    
    try {
        if ($Action -eq "create") {
            Write-Info "Creating stack: $StackName"
            
            aws cloudformation create-stack `
                --stack-name $StackName `
                --template-body file://$CloudFormationTemplate `
                --parameters file://stack-parameters.json `
                --capabilities CAPABILITY_NAMED_IAM `
                --region $Region
                
            if ($LASTEXITCODE -ne 0) {
                Write-Error "Failed to create stack"
                Remove-Item "stack-parameters.json" -ErrorAction SilentlyContinue
                return $false
            }
            
            Write-Success "Stack creation initiated"
            Write-Info "Waiting for stack creation to complete (this may take 5-10 minutes)..."
            
            aws cloudformation wait stack-create-complete --stack-name $StackName --region $Region
            
            if ($LASTEXITCODE -eq 0) {
                Write-Success "Stack created successfully!"
                Remove-Item "stack-parameters.json" -ErrorAction SilentlyContinue
                return $true
            } else {
                Write-Error "Stack creation failed or timed out"
                Show-StackEvents
                Remove-Item "stack-parameters.json" -ErrorAction SilentlyContinue
                return $false
            }
        } else {
            Write-Info "Updating stack: $StackName"
            
            aws cloudformation update-stack `
                --stack-name $StackName `
                --template-body file://$CloudFormationTemplate `
                --capabilities CAPABILITY_NAMED_IAM `
                --region $Region
                
            if ($LASTEXITCODE -ne 0) {
                $errorOutput = $Error[0].Exception.Message
                if ($errorOutput -match "No updates are to be performed") {
                    Write-Warning "No changes to update"
                    Remove-Item "stack-parameters.json" -ErrorAction SilentlyContinue
                    return $true
                } else {
                    Write-Error "Failed to update stack"
                    Remove-Item "stack-parameters.json" -ErrorAction SilentlyContinue
                    return $false
                }
            }
            
            Write-Success "Stack update initiated"
            Write-Info "Waiting for stack update to complete..."
            
            aws cloudformation wait stack-update-complete --stack-name $StackName --region $Region
            
            if ($LASTEXITCODE -eq 0) {
                Write-Success "Stack updated successfully!"
                Remove-Item "stack-parameters.json" -ErrorAction SilentlyContinue
                return $true
            } else {
                Write-Error "Stack update failed or timed out"
                Show-StackEvents
                Remove-Item "stack-parameters.json" -ErrorAction SilentlyContinue
                return $false
            }
        }
    } catch {
        Write-Error "Failed to $Action stack: $_"
        Remove-Item "stack-parameters.json" -ErrorAction SilentlyContinue
        return $false
    }
}

function Show-StackEvents {
    Write-Host ""
    Write-Host "Recent Stack Events:" -ForegroundColor Yellow
    try {
        aws cloudformation describe-stack-events `
            --stack-name $StackName `
            --region $Region `
            --max-items 10 `
            --query 'StackEvents[*].[Timestamp,ResourceStatus,ResourceType,LogicalResourceId,ResourceStatusReason]' `
            --output table
    } catch {
        Write-Warning "Could not retrieve stack events"
    }
}

function Get-StackOutputs {
    Write-Step "Stack Outputs"
    
    try {
        $stack = aws cloudformation describe-stacks --stack-name $StackName --region $Region | ConvertFrom-Json
        $outputs = $stack.Stacks[0].Outputs
        
        if ($outputs.Count -eq 0) {
            Write-Warning "No outputs found"
            return
        }
        
        foreach ($output in $outputs) {
            Write-Host "$($output.OutputKey):" -ForegroundColor Cyan -NoNewline
            Write-Host " $($output.OutputValue)"
            
            # Store important values
            switch ($output.OutputKey) {
                "ApplicationURL" { $script:AppURL = $output.OutputValue }
                "APIURL" { $script:ApiURL = $output.OutputValue }
                "ALBEndpoint" { $script:ALBEndpoint = $output.OutputValue }
            }
        }
        
        if ($script:AppURL) {
            Write-Host ""
            Write-Host "Application URL: $script:AppURL" -ForegroundColor Yellow
            Write-Host "API Endpoint: $script:ApiURL" -ForegroundColor Yellow
            Write-Host ""
            Write-Host "Deployment complete! Your HR-Recruiting application is ready." -ForegroundColor Green
        }
        
    } catch {
        Write-Warning "Could not retrieve stack outputs: $_"
    }
}

function Show-StackStatus {
    try {
        $stack = aws cloudformation describe-stacks --stack-name $StackName --region $Region | ConvertFrom-Json
        $stackInfo = $stack.Stacks[0]
        
        Write-Host ""
        Write-Host "Stack Status" -ForegroundColor Cyan
        Write-Host "=" * 60
        Write-Host "Stack Name:   $($stackInfo.StackName)"
        Write-Host "Status:       $($stackInfo.StackStatus)" -ForegroundColor $(
            switch ($stackInfo.StackStatus) {
                { $_ -match "COMPLETE" } { "Green" }
                { $_ -match "PROGRESS" } { "Yellow" }
                { $_ -match "FAILED" } { "Red" }
                default { "White" }
            }
        )
        Write-Host "Created:      $($stackInfo.CreationTime)"
        if ($stackInfo.LastUpdatedTime) {
            Write-Host "Last Updated: $($stackInfo.LastUpdatedTime)"
        }
        Write-Host "=" * 60
        
        # Get ECS service status
        try {
            $cluster = Get-StackOutputValue -OutputKey "ECSCluster"
            if ($cluster) {
                Write-Host ""
                Write-Host "ECS Services:" -ForegroundColor Cyan
                
                $backendService = aws ecs describe-services --cluster $cluster --services "$StackName-backend" --region $Region 2>&1 | ConvertFrom-Json
                if ($backendService.services) {
                    $svc = $backendService.services[0]
                    Write-Host "  Backend:  $($svc.runningCount)/$($svc.desiredCount) running"
                }
                
                $frontendService = aws ecs describe-services --cluster $cluster --services "$StackName-frontend" --region $Region 2>&1 | ConvertFrom-Json
                if ($frontendService.services) {
                    $svc = $frontendService.services[0]
                    Write-Host "  Frontend: $($svc.runningCount)/$($svc.desiredCount) running"
                }
            }
        } catch {
            # Ignore ECS status errors
        }
        
        Get-StackOutputs
        
    } catch {
        Write-Error "Stack not found: $StackName"
        exit 1
    }
}

function Remove-Stack {
    Write-Warning "This will delete the entire stack including:"
    Write-Host "  - ECS Cluster and Services"
    Write-Host "  - ECR Repositories (images will be deleted)"
    Write-Host "  - Load Balancer and Target Groups"
    Write-Host "  - CloudWatch Log Groups"
    Write-Host "  - Security Groups"
    
    Write-Host ""
    $confirmation = Read-Host "Are you sure you want to delete stack '$StackName'? (yes/no)"
    
    if ($confirmation -ne "yes") {
        Write-Info "Deletion cancelled"
        exit 0
    }
    
    # Delete ECR images first
    Write-Info "Deleting ECR images..."
    try {
        aws ecr batch-delete-image --repository-name $BackendECRRepo --image-ids imageTag=latest --region $Region 2>&1 | Out-Null
        aws ecr batch-delete-image --repository-name $FrontendECRRepo --image-ids imageTag=latest --region $Region 2>&1 | Out-Null
    } catch {
        # Ignore errors if images don't exist
    }
    
    Write-Info "Deleting stack: $StackName"
    
    try {
        aws cloudformation delete-stack --stack-name $StackName --region $Region
        Write-Success "Stack deletion initiated"
        
        Write-Info "Waiting for stack deletion to complete..."
        aws cloudformation wait stack-delete-complete --stack-name $StackName --region $Region
        
        if ($LASTEXITCODE -eq 0) {
            Write-Success "Stack deleted successfully!"
        } else {
            Write-Error "Stack deletion failed or timed out"
            Show-StackEvents
        }
    } catch {
        Write-Error "Failed to delete stack: $_"
    }
}

function Update-ECSServices {
    Write-Step "Updating ECS Services"
    
    $cluster = Get-StackOutputValue -OutputKey "ECSCluster"
    $frontendService = Get-StackOutputValue -OutputKey "FrontendService"
    $backendService = Get-StackOutputValue -OutputKey "BackendService"
    
    if (-not $cluster) {
        Write-Error "Could not find ECS cluster. Is the stack deployed?"
        return $false
    }
    
    Write-Info "Updating frontend service..."
    aws ecs update-service --cluster $cluster --service $frontendService --force-new-deployment --region $Region 2>&1 | Out-Null
    if ($LASTEXITCODE -eq 0) {
        Write-Success "Frontend service update initiated"
    }
    
    Write-Info "Updating backend service..."
    aws ecs update-service --cluster $cluster --service $backendService --force-new-deployment --region $Region 2>&1 | Out-Null
    if ($LASTEXITCODE -eq 0) {
        Write-Success "Backend service update initiated"
    }
    
    Write-Info "Waiting for services to stabilize (this may take 3-5 minutes)..."
    aws ecs wait services-stable --cluster $cluster --services $backendService $frontendService --region $Region
    
    if ($LASTEXITCODE -eq 0) {
        Write-Success "All services are stable"
        return $true
    } else {
        Write-Warning "Service stabilization timed out"
        return $false
    }
}

# ==========================================
# Main Script
# ==========================================

if ($Help) {
    Show-Help
}

if ($Status) {
    Show-StackStatus
    exit 0
}

if ($Delete) {
    Remove-Stack
    exit 0
}

# Check prerequisites
Test-Prerequisites

# Build and push Docker images
if (-not $SkipImages -and -not $SkipBuild) {
    Write-Step "Building and Pushing Docker Images"
    
    # Create ECR repositories
    $backendECRUri = New-ECRRepository -RepoName $BackendECRRepo
    $frontendECRUri = New-ECRRepository -RepoName $FrontendECRRepo
    
    # Check if backend and frontend directories exist
    if (-not (Test-Path "backend")) {
        Write-Error "Backend directory not found."
        exit 1
    }
    
    if (-not (Test-Path "frontend")) {
        Write-Error "Frontend directory not found."
        exit 1
    }
    
    # Get ALB endpoint for frontend API URL
    $albEndpoint = Get-StackOutputValue -OutputKey "ALBEndpoint"
    if (-not $albEndpoint) {
        Write-Warning "ALB endpoint not found. Using placeholder."
        $albEndpoint = "PLACEHOLDER_ALB_ENDPOINT"
    }
    
    # Build and push frontend
    $apiUrl = "http://$albEndpoint/api"
    
    # Create frontend .env
    Push-Location frontend
    @"
VITE_API_URL=$apiUrl
VITE_ENVIRONMENT=$Environment
"@ | Out-File -FilePath ".env" -Encoding UTF8
    Pop-Location
    
    $script:FrontendImageURI = Build-AndPushImage `
        -ImageName "hr-recruiting-frontend" `
        -DockerfilePath "frontend/Dockerfile" `
        -ContextPath "frontend" `
        -ECRUri $frontendECRUri
    
    # Build and push backend
    $script:BackendImageURI = Build-AndPushImage `
        -ImageName "hr-recruiting-backend" `
        -DockerfilePath "backend/Dockerfile" `
        -ContextPath "backend" `
        -ECRUri $backendECRUri
    
    Write-Success "All images built and pushed successfully!"
} elseif ($SkipBuild) {
    Write-Info "Skipping image build - using existing images in ECR"
    
    $script:BackendImageURI = "$script:AccountId.dkr.ecr.$Region.amazonaws.com/${BackendECRRepo}:latest"
    $script:FrontendImageURI = "$script:AccountId.dkr.ecr.$Region.amazonaws.com/${FrontendECRRepo}:latest"
    
    Write-Info "Using backend image: $script:BackendImageURI"
    Write-Info "Using frontend image: $script:FrontendImageURI"
}

# Deploy or update stack
if ($Update) {
    # For updates, we can skip stack deployment and just update services
    if ($SkipImages) {
        Write-Info "Skipping stack update, updating services only..."
        $success = Update-ECSServices
    } else {
        $success = Deploy-Stack -Action "update"
        if ($success) {
            $success = Update-ECSServices
        }
    }
} else {
    # Check if stack already exists
    try {
        aws cloudformation describe-stacks --stack-name $StackName --region $Region 2>&1 | Out-Null
        if ($LASTEXITCODE -eq 0) {
            Write-Warning "Stack already exists. Use -Update to update it."
            Write-Info "Or use -Delete to remove it first."
            exit 1
        }
    } catch {}
    
    Write-Error "Stack does not exist. Deploy the CloudFormation stack first."
    Write-Host ""
    Write-Host "See QUICKSTART-SHARED-VPC.md for deployment instructions."
    exit 1
}

if ($success) {
    Get-StackOutputs
    
    Write-Host ""
    Write-Host ("=" * 60) -ForegroundColor Green
    Write-Host "   HR-Recruiting Deployment Complete! " -ForegroundColor Green
    Write-Host ("=" * 60) -ForegroundColor Green
    
    Write-Host ""
    Write-Host "Next Steps:" -ForegroundColor Cyan
    Write-Host "  1. Wait 2-3 minutes for services to register with load balancer"
    Write-Host "  2. Visit your application URL above"
    Write-Host "  3. Test the job listings and application flow"
    
    Write-Host ""
    Write-Host "Monitor deployment:" -ForegroundColor Cyan
    Write-Host "   .\deploy-ecs.ps1 -Status" -ForegroundColor Gray
    Write-Host "   aws logs tail /ecs/$StackName-backend --follow" -ForegroundColor Gray
    Write-Host "   aws logs tail /ecs/$StackName-frontend --follow" -ForegroundColor Gray
}