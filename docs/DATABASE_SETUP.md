# Database Setup Guide

## Overview

October Backend supports both **local MongoDB** (for development) and **remote MongoDB clusters** (for production).

## 🏠 Local Development Setup

### Option 1: With Local MongoDB (Full Stack)
```bash
# Starts backend + local MongoDB + Mongo Express
docker-compose --profile local-db up
```

### Option 2: With Remote MongoDB (Backend Only)
```bash
# Starts only the backend, connects to remote MongoDB
docker-compose up october-backend
```

## ☁️ Remote MongoDB Cluster Setup

### Initial Setup (Run Once)

1. **Set your connection string:**
```bash
export DATABASE_URI="mongodb+srv://username:password@cluster.mongodb.net/october"
```

2. **Run the database setup script:**
```bash
./scripts/setup-remote-db.sh
```

This script will:
- ✅ Create optimized indexes for all collections
- ✅ Set up text search capabilities
- ✅ Configure performance indexes for AI operations
- ✅ Handle existing indexes gracefully

### Production Deployment

For AWS ECS deployment, the database URI is automatically loaded from AWS Systems Manager Parameter Store at `/october/DATABASE_URI`.

## 📊 Database Schema

### Collections

#### `companies`
- **Indexes:**
  - `name` (unique)
  - `ticker` (unique, sparse)
  - `industry`

#### `news`
- **Indexes:**
  - `title`
  - `published_date` (descending)
  - `companies`
  - `guid` (unique)
  - `companies + published_date` (compound)
  - `relevance_score` (descending)
  - `feed_source`
  - Full-text search on `title`, `summary`, `content`

## 🔧 Maintenance

### View Index Status
```bash
mongosh "$DATABASE_URI" --eval "
db.companies.getIndexes();
db.news.getIndexes();
"
```

### Check Collection Stats
```bash
mongosh "$DATABASE_URI" --eval "
db.companies.stats();
db.news.stats();
"
```

## 🚀 Development Workflows

### New Developer Setup
```bash
# Clone the repository
git clone <repo-url>
cd october_backend

# Start with local database
docker-compose --profile local-db up
```

### Using Remote Database in Development
```bash
# Set your remote database URI
export DATABASE_URI="your-remote-connection-string"

# Start only the backend
docker-compose up october-backend
```

### Production Deployment
```bash
# Deploy to AWS (uses remote database automatically)
./aws/deploy.sh
```

## 📝 Files Overview

- `docker/mongo-init.js` - Local development database initialization
- `scripts/setup-remote-db.sh` - Remote cluster setup script
- `docker-compose.yml` - Container orchestration with profiles
- Database indexes are created automatically on first connection in the application code