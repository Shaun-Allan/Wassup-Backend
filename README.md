# Wassup Backend

A real-time chat application backend built with Go microservices architecture, featuring user authentication, friend management, direct messaging, group chats, and user search functionality.

## 🚀 Features

- **User Authentication**: Registration and login system
- **User Search**: Find users by name or email
- **Friend Management**: Add friends and manage friend relationships
- **Real-time Direct Messaging**: WebSocket-based instant messaging
- **Group Chat**: Create groups and add members
- **Message History**: Persistent message storage
- **User Profile**: Get user information by ID
- **Microservices Architecture**: Distributed service design

## 🛠️ Tech Stack

- **Language**: Go (Golang)
- **Web Framework**: Gin (HTTP) & Gorilla Mux
- **Databases**: 
  - PostgreSQL (User data, friendships, groups)
  - MongoDB (Messages, chat history)
- **Real-time Communication**: WebSockets (Gorilla WebSocket)
- **Containerization**: Docker
- **Orchestration**: Kubernetes

## 🏗️ Architecture

The application consists of 6 microservices:

| Service | Port | Description | Database |
|---------|------|-------------|----------|
| **Auth Service** | 5000 | User registration and login | PostgreSQL |
| **Search Service** | 5001 | User search functionality | PostgreSQL |
| **Friends Service** | 5002 | Friend management | PostgreSQL |
| **Messaging Service** | 5003 | Real-time messaging & history | MongoDB |
| **Groups Service** | 5004 | Group chat management | PostgreSQL + MongoDB |
| **User Service** | 5006 | User profile information | PostgreSQL |

## 📋 Prerequisites

Before running this application, make sure you have the following installed:

- Go 1.21 or higher
- PostgreSQL
- MongoDB
- Docker (optional)
- Kubernetes (optional)

## ⚙️ Installation

### 1. Clone the Repository
```bash
git clone https://github.com/Shaun-Allan/Wassup-Backend.git
cd Wassup-Backend
```

### 2. Database Setup

#### PostgreSQL Setup
```sql
-- Create database
CREATE DATABASE wassupdb;

-- Create users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL
);

-- Create friendships table
CREATE TABLE friendships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    friend_id UUID REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, friend_id)
);

-- Create groups table
CREATE TABLE groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create group memberships table
CREATE TABLE group_memberships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id UUID REFERENCES groups(id),
    user_id UUID REFERENCES users(id),
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(group_id, user_id)
);
```

#### MongoDB Setup
MongoDB will automatically create the required collections:
- `wassupdb.dms` - Direct messages
- `wassupdb.groups_meta` - Group metadata and messages

### 3. Install Dependencies

For each service, install Go dependencies:
```bash
# Initialize go modules (if not already done)
go mod init wassup-backend

# Install dependencies
go mod tidy
```

Required dependencies:
```go
github.com/gin-gonic/gin
github.com/jackc/pgx/v5
github.com/jackc/pgx/v5/pgxpool
github.com/gorilla/websocket
github.com/gorilla/mux
github.com/google/uuid
go.mongodb.org/mongo-driver/mongo
go.mongodb.org/mongo-driver/bson
```

### 4. Environment Configuration

Update database connection strings in each service:

**PostgreSQL Connection String:**
```go
"postgres://username:password@localhost:5432/wassupdb"
```

**MongoDB Connection String:**
```go
"mongodb://localhost:27017"
```

### 5. Running the Services

#### Manual Start (Development)
```bash
# Terminal 1 - Auth Service
go run auth-service.go

# Terminal 2 - Search Service  
go run search-service.go

# Terminal 3 - Friends Service
go run friends-service.go

# Terminal 4 - Messaging Service
go run messaging-service.go

# Terminal 5 - Groups Service
go run groups-service.go

# Terminal 6 - User Service
go run user-service.go
```

#### Using Docker
```bash
# Build and run with Docker
docker build -t wassup-backend .
docker run -p 5000-5006:5000-5006 wassup-backend
```

#### Using Kubernetes
```bash
# Deploy to Kubernetes cluster
kubectl apply -f k8s/
```

## 🔌 API Endpoints

### Auth Service (Port 5000)
- `POST /register` - Register new user
- `POST /login` - User login

### Search Service (Port 5001)
- `POST /search` - Search users by name/email

### Friends Service (Port 5002)
- `POST /addFriend` - Add a friend
- `GET /users/:userID/friends` - Get user's friends
- `POST /checkFriends` - Check if users are friends

### Messaging Service (Port 5003)
- `WS /ws?user_id={id}` - WebSocket connection for real-time messaging
- `GET /history?user1={id}&user2={id}` - Get message history between users

### Groups Service (Port 5004)
- `POST /createGroup` - Create a new group
- `POST /groups/{id}/addMember` - Add members to group
- `GET /users/{id}/groups` - Get user's groups

### User Service (Port 5006)
- `GET /get-name?user_id={id}` - Get user name by ID

## 📝 API Examples

### Register User
```bash
curl -X POST http://localhost:5000/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com", 
    "password": "password123"
  }'
```

### Login
```bash
curl -X POST http://localhost:5000/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "password123"
  }'
```

### Search Users
```bash
curl -X POST http://localhost:5001/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "john"
  }'
```

### WebSocket Connection (JavaScript)
```javascript
const ws = new WebSocket('ws://localhost:5003/ws?user_id=USER_ID');

// Send message
ws.send(JSON.stringify({
  sender: "SENDER_ID",
  receiver: "RECEIVER_ID", 
  content: "Hello, World!"
}));

// Receive messages
ws.onmessage = function(event) {
  const message = JSON.parse(event.data);
  console.log('Received:', message);
};
```

## 🐳 Docker Deployment

### Dockerfile Example
```dockerfile
FROM golang:1.21 as builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o app

FROM debian:bullseye-slim
WORKDIR /app
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*
COPY --from=builder /app/app .

EXPOSE 5000-5006
CMD ["./app"]
```

### Docker Compose
```yaml
version: '3.8'
services:
  postgres:
    image: postgres:13
    environment:
      POSTGRES_DB: wassupdb
      POSTGRES_USER: shaun
      POSTGRES_PASSWORD: shaun
    ports:
      - "5432:5432"

  mongodb:
    image: mongo:5
    ports:
      - "27017:27017"

  auth-service:
    build: .
    ports:
      - "5000:5000"
    depends_on:
      - postgres

  # Add other services...
```

## ☸️ Kubernetes Deployment

### Service Deployment Example
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: auth-service
spec:
  replicas: 3
  selector:
    matchLabels:
      app: auth-service
  template:
    metadata:
      labels:
        app: auth-service
    spec:
      containers:
      - name: auth-service
        image: wassup-backend:latest
        ports:
        - containerPort: 5000
---
apiVersion: v1
kind: Service
metadata:
  name: auth-service
spec:
  selector:
    app: auth-service
  ports:
  - port: 5000
    targetPort: 5000
  type: LoadBalancer
```
