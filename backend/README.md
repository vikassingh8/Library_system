# Library System - Backend

REST API built with Go, Azure SQL Database, and Azure Blob Storage.

## Requirements

- Go 1.21+
- Azure SQL Database
- Azure Blob Storage account

## Setup

1. Clone the repo and go to backend folder
```bash
cd library-system/backend
```

2. Install dependencies
```bash
go mod download
```

3. Create your config file
```bash
cp config/local.yaml.example config/local.yaml
```
Fill in your database, Azure Blob, and JWT values in `config/local.yaml`.

4. Run the server
```bash
go run ./cmd/server/main.go --config config/local.yaml
```

Server starts at `http://localhost:8000`. Database tables are created automatically.

## Make a User Admin

Run this in Azure SQL Query Editor:
```sql
UPDATE users SET role = 'admin' WHERE email = 'your@email.com';
```

