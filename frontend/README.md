# HR Recruiting Platform

A modern recruitment platform built with TypeScript/Svelte frontend and Go backend, integrating with Hub-HRMS via GraphQL.

## Architecture

```
┌─────────────────────────────────┐
│   Svelte Frontend (TypeScript)  │
│   - Job Board                   │
│   - Application Portal          │
│   - Recruiter Dashboard         │
└────────────┬────────────────────┘
             │ GraphQL
             ├──────────────────────┐
             │                      │
┌────────────▼────────────┐  ┌─────▼──────────────┐
│  Recruiting Backend (Go)│  │  Hub-HRMS Core     │
│  - GraphQL Gateway      │──│  - Job Service     │
│  - File Upload          │  │  - Candidate Svc   │
│  - AI Integration       │  │  - AI Scoring      │
└─────────────────────────┘  └────────────────────┘
```

## Project Structure

```
hr-recruiting/
├── frontend/                # Svelte TypeScript app
│   ├── src/
│   │   ├── lib/
│   │   │   ├── api/        # GraphQL client
│   │   │   ├── components/ # Reusable components
│   │   │   └── stores/     # State management
│   │   └── routes/         # SvelteKit routes
│   ├── package.json
│   └── tsconfig.json
│
├── backend/                 # Go backend
│   ├── cmd/server/         # Main application
│   ├── internal/
│   │   ├── gateway/        # GraphQL gateway to Hub-HRMS
│   │   ├── upload/         # File upload service
│   │   └── middleware/     # Auth, CORS, etc.
│   ├── go.mod
│   └── Dockerfile
│
└── docker-compose.yml      # Local development
```

## Quick Start

### Prerequisites
- Node.js 18+
- Go 1.21+
- Docker (optional)

### Frontend Setup
```bash
cd frontend
npm install
npm run dev
```

### Backend Setup
```bash
cd backend
go mod download
go run cmd/server/main.go
```

### Using Docker
```bash
docker-compose up
```

## Environment Variables

### Frontend (.env)
```
VITE_API_URL=http://localhost:8080
VITE_HUBHRMS_API_URL=https://api.hubhrms.com/graphql
```

### Backend (.env)
```
PORT=8080
HUBHRMS_GRAPHQL_URL=https://api.hubhrms.com/graphql
HUBHRMS_API_KEY=your-api-key
AWS_S3_BUCKET=hr-recruiting-resumes
AWS_REGION=us-east-1
```

## Features

### For Candidates
- Browse and search jobs
- Apply with resume upload
- Track application status
- Build candidate profile

### For Recruiters
- Post and manage jobs
- Review applications
- AI-powered resume screening
- Analytics dashboard

## Development

```bash
# Frontend
npm run dev          # Development server
npm run build        # Production build
npm run check        # Type checking

# Backend
go run cmd/server/main.go  # Development server
go test ./...              # Run tests
go build -o server         # Production build
```

## Deployment

### Frontend (Vercel)
```bash
cd frontend
vercel deploy --prod
```

### Backend (Docker)
```bash
cd backend
docker build -t hr-recruiting-backend .
docker push your-registry/hr-recruiting-backend
```

## License

Proprietary - Part of Hub HRMS Suite
