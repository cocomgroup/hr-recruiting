# get-hubhrms-vpc-info.ps1
# Get VPC and subnet information from Hub-HRMS stack

param(
    [Parameter(Mandatory=$false)]
    [string]$HubHRMSStackName = "hub-hrms-dev",
    
    [Parameter(Mandatory=$false)]
    [string]$Region = "us-east-1"
)

Write-Host "=== Finding Hub-HRMS VPC Information ===" -ForegroundColor Green
Write-Host ""

# Get Hub-HRMS stack outputs
Write-Host "Checking Hub-HRMS stack: $HubHRMSStackName" -ForegroundColor Cyan

try {
    $outputs = aws cloudformation describe-stacks --stack-name $HubHRMSStackName --region $Region --query 'Stacks[0].Outputs' --output json | ConvertFrom-Json
    
    if ($LASTEXITCODE -ne 0) { throw }
}
catch {
    Write-Host "Could not find Hub-HRMS stack outputs." -ForegroundColor Yellow
    Write-Host "Searching for VPC with Internet Gateway..." -ForegroundColor Cyan
    Write-Host ""
    
    # Find VPCs with Internet Gateways
    $vpcs = aws ec2 describe-vpcs --region $Region --query 'Vpcs[*].[VpcId,Tags[?Key==`Name`].Value|[0],CidrBlock]' --output json | ConvertFrom-Json
    
    Write-Host "Available VPCs:" -ForegroundColor Yellow
    Write-Host ""
    
    foreach ($vpc in $vpcs) {
        $vpcId = $vpc[0]
        $vpcName = if ($vpc[1]) { $vpc[1] } else { "(no name)" }
        $cidr = $vpc[2]
        
        # Check if VPC has an Internet Gateway
        $igws = aws ec2 describe-internet-gateways --filters "Name=attachment.vpc-id,Values=$vpcId" --region $Region --query 'InternetGateways[*].[InternetGatewayId,Tags[?Key==`Name`].Value|[0]]' --output json | ConvertFrom-Json
        
        if ($igws.Count -gt 0) {
            $igwId = $igws[0][0]
            $igwName = if ($igws[0][1]) { $igws[0][1] } else { "(no name)" }
            
            Write-Host "VPC: $vpcName" -ForegroundColor Green
            Write-Host "  VPC ID: $vpcId"
            Write-Host "  CIDR: $cidr"
            Write-Host "  Internet Gateway: $igwName ($igwId)"
            
            # Get subnets
            $subnets = aws ec2 describe-subnets --filters "Name=vpc-id,Values=$vpcId" --region $Region --query 'Subnets[*].[SubnetId,Tags[?Key==`Name`].Value|[0],CidrBlock,AvailabilityZone,MapPublicIpOnLaunch]' --output json | ConvertFrom-Json
            
            if ($subnets.Count -gt 0) {
                Write-Host "  Subnets:"
                foreach ($subnet in $subnets) {
                    $subnetId = $subnet[0]
                    $subnetName = if ($subnet[1]) { $subnet[1] } else { "(no name)" }
                    $subnetCidr = $subnet[2]
                    $az = $subnet[3]
                    $isPublic = $subnet[4]
                    
                    $type = if ($isPublic) { "PUBLIC" } else { "PRIVATE" }
                    Write-Host "    - $subnetName ($type)" -ForegroundColor Cyan
                    Write-Host "      ID: $subnetId"
                    Write-Host "      CIDR: $subnetCidr"
                    Write-Host "      AZ: $az"
                }
            }
            Write-Host ""
        }
    }
    
    Write-Host "To use a VPC, note the VPC ID and at least 2 subnet IDs" -ForegroundColor Yellow
    exit 0
}

# Extract VPC information from Hub-HRMS outputs
$vpcId = ($outputs | Where-Object {$_.OutputKey -eq "VPCId" -or $_.OutputKey -eq "VPC"}).OutputValue

if (-not $vpcId) {
    Write-Host "Could not find VPC ID in Hub-HRMS outputs" -ForegroundColor Yellow
    Write-Host "Available outputs:" -ForegroundColor Cyan
    $outputs | ForEach-Object { Write-Host "  $($_.OutputKey): $($_.OutputValue)" }
    exit 1
}

Write-Host "Found Hub-HRMS VPC: $vpcId" -ForegroundColor Green
Write-Host ""

# Get VPC details
$vpcDetails = aws ec2 describe-vpcs --vpc-ids $vpcId --region $Region --query 'Vpcs[0].[VpcId,CidrBlock,Tags[?Key==`Name`].Value|[0]]' --output json | ConvertFrom-Json

$vpcName = if ($vpcDetails -and $vpcDetails[2]) { $vpcDetails[2] } else { "Hub-HRMS VPC" }
$vpcCidr = if ($vpcDetails -and $vpcDetails[1]) { $vpcDetails[1] } else { "Unknown" }

Write-Host "VPC Information:" -ForegroundColor Cyan
Write-Host "  Name: $vpcName"
Write-Host "  VPC ID: $vpcId"
Write-Host "  CIDR Block: $vpcCidr"
Write-Host ""

# Get Internet Gateway
$igw = aws ec2 describe-internet-gateways --filters "Name=attachment.vpc-id,Values=$vpcId" --region $Region --query 'InternetGateways[0].[InternetGatewayId,Tags[?Key==`Name`].Value|[0]]' --output json | ConvertFrom-Json

if ($igw) {
    $igwId = $igw[0]
    $igwName = if ($igw[1]) { $igw[1] } else { "(no name)" }
    Write-Host "Internet Gateway:" -ForegroundColor Cyan
    Write-Host "  Name: $igwName"
    Write-Host "  IGW ID: $igwId"
    Write-Host ""
}

# Get subnets
$subnets = aws ec2 describe-subnets --filters "Name=vpc-id,Values=$vpcId" --region $Region --query 'Subnets[*].[SubnetId,Tags[?Key==`Name`].Value|[0],CidrBlock,AvailabilityZone,MapPublicIpOnLaunch]' --output json | ConvertFrom-Json

if (-not $subnets -or $subnets.Count -eq 0) {
    Write-Host "No subnets found in VPC!" -ForegroundColor Red
    exit 1
}

Write-Host "Subnets:" -ForegroundColor Cyan
$publicSubnets = @()

foreach ($subnet in $subnets) {
    $subnetId = $subnet[0]
    $subnetName = if ($subnet[1]) { $subnet[1] } else { "(no name)" }
    $subnetCidr = $subnet[2]
    $az = $subnet[3]
    $isPublic = $subnet[4]
    
    if ($isPublic) {
        $publicSubnets += $subnetId
        $type = "PUBLIC"
        $color = "Green"
    }
    else {
        $type = "PRIVATE"
        $color = "Yellow"
    }
    
    Write-Host "  - $subnetName ($type)" -ForegroundColor $color
    Write-Host "    ID: $subnetId"
    Write-Host "    CIDR: $subnetCidr"
    Write-Host "    AZ: $az"
    Write-Host ""
}

# Check if we have at least 2 public subnets
if ($publicSubnets.Count -lt 2) {
    Write-Host "WARNING: Need at least 2 public subnets for ALB" -ForegroundColor Yellow
    Write-Host "Found $($publicSubnets.Count) public subnet(s)" -ForegroundColor Yellow
    
    if ($publicSubnets.Count -eq 1) {
        Write-Host ""
        Write-Host "You can either:" -ForegroundColor Cyan
        Write-Host "  1. Use the single public subnet for both parameters (not recommended)"
        Write-Host "  2. Create another public subnet in a different AZ"
    }
}

Write-Host ""
Write-Host "=== CloudFormation Parameters ===" -ForegroundColor Green
Write-Host ""
Write-Host "ExistingVpcId=$vpcId" -ForegroundColor White

if ($publicSubnets.Count -ge 2) {
    Write-Host "ExistingSubnet1Id=$($publicSubnets[0])" -ForegroundColor White
    Write-Host "ExistingSubnet2Id=$($publicSubnets[1])" -ForegroundColor White
}
elseif ($publicSubnets.Count -eq 1) {
    Write-Host "ExistingSubnet1Id=$($publicSubnets[0])" -ForegroundColor White
    Write-Host "ExistingSubnet2Id=$($publicSubnets[0])" -ForegroundColor Yellow
}
else {
    Write-Host "ERROR: No public subnets found!" -ForegroundColor Red
}

Write-Host ""
Write-Host "=== Deployment Command ===" -ForegroundColor Green
Write-Host ""

if ($publicSubnets.Count -ge 2) {
    Write-Host "aws cloudformation create-stack --stack-name hr-recruiting --template-body file://cloudformation-ecs-shared-vpc.yaml --capabilities CAPABILITY_NAMED_IAM --parameters ParameterKey=ExistingVpcId,ParameterValue=$vpcId ParameterKey=ExistingSubnet1Id,ParameterValue=$($publicSubnets[0]) ParameterKey=ExistingSubnet2Id,ParameterValue=$($publicSubnets[1]) ParameterKey=EnvironmentName,ParameterValue=production" -ForegroundColor White
}
else {
    Write-Host "Fix subnet configuration first!" -ForegroundColor Red
}

Write-Host ""