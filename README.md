# Workshop BRIN - IT Support Chatbot Demo

## Overview

This project is a demonstration for the BRIN workshop, showcasing an intelligent IT support chatbot system built with modern AI technologies. The chatbot leverages LLM (Large Language Models) and RAG (Retrieval-Augmented Generation) to provide automated IT support assistance.

### Key Components

- **Backend WA Service**: Go-based API service for WhatsApp integration
- **N8N Workflow Automation**: Orchestrates the chatbot workflow, integrating LLM and RAG capabilities
- **PostgreSQL with pgvector**: Database with vector search support for RAG implementation

### Architecture

The system uses Docker Compose to manage microservices with isolated networking and resource management. The chatbot workflow processes user queries through N8N, retrieves relevant context using RAG, and generates responses using LLM.

## How to Run

### Prerequisites

- Docker & Docker Compose installed
- Available ports: 5466 (PostgreSQL), 8095 (Backend API), 5676 (N8N)

### Quick Start

1. **Clone repository**
   ```bash
   git clone <repository-url>
   cd workshop-brin
   ```

2. **Start all services**
   ```bash
   docker-compose up -d
   ```

3. **Check service status**
   ```bash
   docker-compose ps
   ```

4. **View logs**
   ```bash
   # All services
   docker-compose logs -f

   # Specific service
   docker-compose logs -f backend-wa_service
   ```

### Access Services

- **Backend API**: http://localhost:8095
  - Health check: http://localhost:8095/health
- **N8N Dashboard**: http://localhost:5676
  - Username: `admin` (default)
  - Password: `admin123` (default)
- **PostgreSQL**: `localhost:5466`
  - User: `postgres` (default)
  - Password: `workshop2025` (default)

### Environment Configuration

Create a `.env` file in the project root for custom configuration (optional):

```env
# PostgreSQL
POSTGRES_USER=postgres
POSTGRES_PASSWORD=workshop2025

# WA Service Database
POSTGRES_WA_SERVICE_NAME=wa_service
POSTGRES_WA_SERVICE_USER=wa_service
POSTGRES_WA_SERVICE_PASSWORD=workshop2025

# N8N Database
POSTGRES_N8N_NAME=n8n
POSTGRES_N8N_USER=n8n
POSTGRES_N8N_PASSWORD=workshop2025

# N8N Auth
N8N_USER=admin
N8N_PASSWORD=admin123

# Backend Service
JWT_SECRET=workshop2025
N8N_WEBHOOK_URL=https://workshop.gosignal.id/webhook/6c69b572-9c71-4a6b-8827-a31ce8fa6408
```

### Stop Services

```bash
# Stop services
docker-compose down

# Stop and remove volumes
docker-compose down -v
```