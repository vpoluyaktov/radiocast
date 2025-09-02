# Radio Propagation Service

A Go-based service that generates daily amateur radio propagation reports by collecting data from multiple sources, analyzing it with OpenAI LLM, and producing beautiful HTML reports with interactive charts.

## Overview

The Radio Propagation Service automatically:
- Fetches solar and geomagnetic data from NOAA SWPC, N0NBH, and SIDC
- Analyzes conditions using OpenAI GPT-4
- Generates comprehensive HTML reports with go-echarts visualizations
- Stores reports in Google Cloud Storage
- Runs daily via GCP Cloud Scheduler

## Architecture

```
Data Sources → Data Fetcher → LLM Analysis → Report Generator → GCS Storage
     ↓              ↓              ↓              ↓              ↓
  NOAA SWPC     Normalize      OpenAI API    Markdown→HTML    Daily Reports
  N0NBH API      Data         GPT-4 Model    + Charts        /YYYY/MM/DD/
  SIDC RSS                                   go-echarts
```

## Versioning

The project uses automatic semantic versioning (SemVer) for Docker images:
- **Production Format**: `v{BASE_VERSION}.{COMMIT_COUNT}` (e.g., `v0.1.0.47`)
- **Staging Format**: `v{BASE_VERSION}.{COMMIT_COUNT}-rc.{TIMESTAMP}` (e.g., `v0.1.0.47-rc.1234`)
- **Automatic Generation**: 
  - Base version read from `VERSION` file
  - Commit count calculated from git history
  - Each deployment gets unique version automatically
- **Manual Version Changes**: Edit the `VERSION` file to change base version (e.g., for major/minor releases)

## Quick Start

### Prerequisites

- Go 1.21+
- Google Cloud Project with billing enabled
- OpenAI API key
- Docker (for deployment)

### Local Development

1. **Clone and setup**:
```bash
git clone <repository>
cd radiocast/service
go mod download
```

2. **Set environment variables**:
```bash
export OPENAI_API_KEY="your-openai-key"
export GCP_PROJECT_ID="your-gcp-project"
export GCS_BUCKET="your-reports-bucket"
export ENVIRONMENT="development"
```

3. **Run locally**:
```bash
go run main.go
```

4. **Test the service**:
```bash
# Health check
curl http://localhost:8080/health

# Generate report
curl -X POST http://localhost:8080/generate

# List reports
curl http://localhost:8080/reports
```

### Docker Build

```bash
cd service
docker build -t radiocast .
docker run -p 8080:8080 \
  -e OPENAI_API_KEY="your-key" \
  -e GCP_PROJECT_ID="your-project" \
  -e GCS_BUCKET="your-bucket" \
  radiocast
```

## Deployment

### Infrastructure Setup

The service uses Terraform for infrastructure management with separate environments:

1. **Initialize Terraform state buckets**:
```bash
# Create state buckets manually first
gsutil mb gs://dfh-stage-tfstate
gsutil mb gs://dfh-prod-tfstate
```

2. **Deploy staging infrastructure**:
```bash
cd terraform/stage
terraform init
terraform plan -var="openai_api_key=your-key"
terraform apply
```

3. **Deploy production infrastructure**:
```bash
cd terraform/prod
terraform init
terraform plan -var="openai_api_key=your-key"
terraform apply
```

### CI/CD Pipeline

The service uses GitHub Actions for automated deployment:

- **Stage branch** → deploys to `dfh-stage` project
- **Main branch** → deploys to `dfh-prod` project with blue/green deployment

Required GitHub Secrets:
- `GCP_SA_KEY_STAGE`: Service account key for staging
- `GCP_SA_KEY_PROD`: Service account key for production

## API Endpoints

### `GET /`
Service information and available endpoints.

### `GET /health`
Health check endpoint for monitoring.

### `POST /generate`
Generates a new propagation report. Returns:
```json
{
  "status": "success",
  "report_url": "https://storage.googleapis.com/bucket/2024/01/15/PropagationReport-2024-01-15-12-00-00.html",
  "timestamp": "2024-01-15T12:00:00Z",
  "duration_ms": 45000,
  "data_summary": {
    "solar_flux": 150.2,
    "k_index": 2.3,
    "sunspot_number": 45,
    "activity_level": "Moderate"
  }
}
```

### `GET /reports?limit=10`
Lists recent reports with metadata.

## Configuration

All configuration is via environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | HTTP server port | `8080` |
| `OPENAI_API_KEY` | OpenAI API key | Required |
| `OPENAI_MODEL` | OpenAI model to use | `gpt-4.1` |
| `GCP_PROJECT_ID` | GCP project ID | Required |
| `GCS_BUCKET` | GCS bucket for reports | Required |
| `ENVIRONMENT` | Environment name | `development` |
| `LOG_LEVEL` | Logging level | `info` |

## Data Sources

### NOAA Space Weather Prediction Center
- **K-index**: `https://services.swpc.noaa.gov/json/planetary_k_index_1m.json`
- **Solar data**: `https://services.swpc.noaa.gov/json/solar-cycle/observed-solar-cycle-indices.json`

### N0NBH Solar Data API
- **Band conditions**: `https://www.hamqsl.com/solarapi.php?format=json`

### SIDC Solar Events
- **RSS feed**: `https://www.sidc.be/products/meu`

## Report Structure

Generated reports include:

1. **Executive Summary** - Current conditions overview
2. **Solar Activity Analysis** - SFI, sunspot numbers, flare activity
3. **Geomagnetic Conditions** - K-index, A-index, storm levels
4. **Band-by-Band Analysis** - Specific recommendations for each amateur band
5. **Interactive Charts** - Solar trends, K-index history, band conditions
6. **3-Day Forecast** - Predicted conditions and recommendations
7. **DX Opportunities** - Enhanced propagation paths
8. **Technical Explanation** - Educational content for operators

## Development

### Project Structure

```
radiocast/
├── service/                 # Go application
│   ├── main.go             # HTTP server and main logic
│   ├── config/             # Configuration management
│   ├── fetchers/           # Data source integrations
│   ├── llm/                # OpenAI integration
│   ├── models/             # Data structures
│   ├── reports/            # HTML generation and charts
│   ├── storage/            # GCS integration
│   ├── go.mod              # Go dependencies
│   └── Dockerfile          # Container build
├── terraform/              # Terraform infrastructure
│   ├── stage/              # Staging environment
│   └── prod/               # Production environment
├── .github/workflows/      # CI/CD pipelines
└── README.md
```

### Testing

```bash
cd service
go test -v ./...
go vet ./...
```

### Adding New Data Sources

1. Add fetcher function in `fetchers/fetcher.go`
2. Update data models in `models/data.go`
3. Modify normalization logic
4. Update LLM prompt in `llm/openai.go`

## Monitoring

### Cloud Run Metrics
- Request count and latency
- Error rates
- Memory and CPU usage

### Custom Metrics
- Report generation success/failure
- Data source availability
- LLM API response times

### Alerts
- Failed report generation
- High error rates
- Service unavailability

## Troubleshooting

### Common Issues

**Service won't start**:
- Check environment variables are set
- Verify GCP credentials and permissions
- Ensure GCS bucket exists and is accessible

**Report generation fails**:
- Check OpenAI API key and quota
- Verify data source URLs are accessible
- Review logs for specific error messages

**Charts not rendering**:
- Ensure go-echarts dependencies are included
- Check HTML template syntax
- Verify chart data format

### Logs

```bash
# Local development
go run main.go

# Cloud Run logs
gcloud logs read --service=radiocast-prod --limit=100

# Structured logging format
{"level":"info","msg":"Starting data fetch","timestamp":"2024-01-15T12:00:00Z"}
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make changes and add tests
4. Submit a pull request

## License

This project is licensed under the MIT License.

## Support

For issues and questions:
- Create GitHub issues for bugs and feature requests
- Check logs for troubleshooting
- Review the specification document for detailed requirements

---

**73!** 📡
