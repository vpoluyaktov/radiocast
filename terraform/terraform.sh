#!/bin/bash

# Radiocast Terraform Deployment Script
# This script helps deploy the Terraform configuration for different environments

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to setup Terraform
setup_terraform() {
    # Check common Terraform locations
    local terraform_paths=(
        "/opt/homebrew/bin/terraform"
        "/usr/local/bin/terraform"
        "/usr/bin/terraform"
        "$(which terraform 2>/dev/null)"
    )
    
    local terraform_path=""
    for path in "${terraform_paths[@]}"; do
        if [ -n "$path" ] && [ -x "$path" ]; then
            terraform_path="$path"
            break
        fi
    done
    
    if [ -n "$terraform_path" ]; then
        print_status "Found Terraform at: $terraform_path"
        # Add to PATH and export for this session
        export PATH="$(dirname "$terraform_path"):$PATH"
        print_status "Added Terraform to PATH"
        return 0
    else
        print_error "Terraform is not installed or not executable"
        print_error "Please install Terraform or ensure it's in your PATH"
        return 1
    fi
}

# Function to setup GCP authentication
setup_gcp_auth() {
    print_status "Checking GCP authentication..."
    
    # Check if gcloud is authenticated
    if ! gcloud auth list --filter=status:ACTIVE --format="value(account)" | grep -q "@"; then
        print_error "No active GCP authentication found"
        print_error "Please run: gcloud auth login"
        return 1
    fi
    
    # Check application default credentials
    if ! gcloud auth application-default print-access-token >/dev/null 2>&1; then
        print_warning "Application default credentials not found"
        print_status "Setting up application default credentials..."
        gcloud auth application-default login
    fi
    
    print_status "GCP authentication verified"
    return 0
}

# Function to validate OpenAI API key
validate_openai_key() {
    if [ -z "$OPENAI_API_KEY" ]; then
        print_error "OPENAI_API_KEY environment variable is required"
        print_error "Please set it with: export OPENAI_API_KEY=your_api_key"
        return 1
    fi
    
    if [[ ! "$OPENAI_API_KEY" =~ ^sk- ]]; then
        print_warning "OpenAI API key should start with 'sk-'"
    fi
    
    print_status "OpenAI API key validated"
    return 0
}

# Check if environment is provided
if [ -z "$1" ]; then
    print_error "Usage: $0 <environment> [action]"
    echo "Environment options: stage, prod"
    echo "Action options: plan, apply, destroy (default: plan)"
    exit 1
fi

ENVIRONMENT=$1
ACTION=${2:-plan}

# Validate environment
if [[ ! "$ENVIRONMENT" =~ ^(stage|prod)$ ]]; then
    print_error "Invalid environment: $ENVIRONMENT"
    echo "Valid environments: stage, prod"
    exit 1
fi

# Validate action
if [[ ! "$ACTION" =~ ^(plan|apply|destroy)$ ]]; then
    print_error "Invalid action: $ACTION"
    echo "Valid actions: plan, apply, destroy"
    exit 1
fi

print_status "Radiocast Terraform Deployment"
print_status "Environment: $ENVIRONMENT"
print_status "Action: $ACTION"

# Setup Terraform
if ! setup_terraform; then
    exit 1
fi

# Setup GCP authentication
if ! setup_gcp_auth; then
    exit 1
fi

# Validate OpenAI API key for apply/destroy actions
if [[ "$ACTION" =~ ^(apply|destroy)$ ]]; then
    if ! validate_openai_key; then
        exit 1
    fi
fi

# Navigate to terraform directory
if [ ! -d "terraform" ]; then
    print_error "Terraform directory not found. Please run from project root."
    exit 1
fi

cd terraform

# Check if tfvars file exists
TFVARS_FILE="${ENVIRONMENT}.tfvars"
if [ ! -f "$TFVARS_FILE" ]; then
    print_error "Configuration file not found: $TFVARS_FILE"
    exit 1
fi

# Check if backend configuration exists
BACKEND_FILE="${ENVIRONMENT}/backend.tf"
if [ ! -f "$BACKEND_FILE" ]; then
    print_error "Backend configuration not found: $BACKEND_FILE"
    exit 1
fi

# Copy the appropriate backend configuration
print_status "Setting up backend configuration for $ENVIRONMENT..."
cp "$BACKEND_FILE" backend.tf

# Initialize Terraform
print_status "Initializing Terraform..."
terraform init -reconfigure

# Execute Terraform command
case $ACTION in
    "plan")
        print_status "Planning Terraform deployment..."
        if [ -n "$OPENAI_API_KEY" ]; then
            terraform plan -var-file="$TFVARS_FILE" -var="openai_api_key=$OPENAI_API_KEY"
        else
            terraform plan -var-file="$TFVARS_FILE" -var="openai_api_key=placeholder"
        fi
        ;;
    "apply")
        print_status "Applying Terraform configuration..."
        terraform apply -var-file="$TFVARS_FILE" -var="openai_api_key=$OPENAI_API_KEY" -auto-approve
        
        # Get service URL after successful apply
        if [ $? -eq 0 ]; then
            SERVICE_URL=$(terraform output -raw service_url 2>/dev/null || echo "N/A")
            print_status "Deployment completed successfully!"
            print_status "Service URL: $SERVICE_URL"
            
            # Test health endpoint if URL is available
            if [ "$SERVICE_URL" != "N/A" ]; then
                print_status "Testing health endpoint..."
                sleep 10  # Wait for service to be ready
                if curl -f "$SERVICE_URL/health" >/dev/null 2>&1; then
                    print_status "Health check passed!"
                else
                    print_warning "Health check failed - service may still be starting"
                fi
            fi
        fi
        ;;
    "destroy")
        print_warning "This will destroy all resources in the $ENVIRONMENT environment!"
        read -p "Are you sure? Type 'yes' to confirm: " confirm
        if [ "$confirm" = "yes" ]; then
            print_status "Destroying resources..."
            terraform destroy -var-file="$TFVARS_FILE" -var="openai_api_key=$OPENAI_API_KEY" -auto-approve
        else
            print_status "Destroy cancelled"
        fi
        ;;
esac

print_status "Operation completed!"
