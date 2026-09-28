# Go + MongoDB + Redis Authentication and Product Management App

## Overview
This project is a full-stack web application that demonstrates secure user authentication and product management using:

- Go for the backend REST API
- MongoDB for persistent storage of users and products
- Redis for session storage and product caching
- Vanilla HTML, CSS, and JavaScript for the frontend

The application allows users to sign up, log in, view their profile, manage a product catalog, and maintain secure sessions using HTTP-only cookies.

## Architecture

The app follows a simple layered flow:

Browser / Frontend -> Go API -> MongoDB + Redis

Core backend implementation is organized in:

- [backend/main.go](backend/main.go) — database connection, Redis setup, and server startup
- [backend/server.go](backend/server.go) — route registration and CORS middleware setup
- [backend/auth_handler.go](backend/auth_handler.go) — signup, login, profile, and logout logic
- [backend/product_handler.go](backend/product_handler.go) — product CRUD and Redis cache invalidation
- [backend/middleware.go](backend/middleware.go) — session validation and CORS handling
- [backend/models.go](backend/models.go) — data models and request schemas

Separate UI hosts:

- [frontend/shopper-ui/index.html](frontend/shopper-ui/index.html) — shopper storefront
- [frontend/admin-ui/index.html](frontend/admin-ui/index.html) — admin inventory panel

## Features

- User registration with password hashing using bcrypt
- Secure login with session tokens stored in Redis
- HTTP-only cookie-based authentication
- Protected profile route requiring valid session
- Product creation, update, deletion, and listing
- Per-user product isolation
- Redis-based product caching for faster reads
- CORS support for local frontend development

## Tech Stack

- Language: Go
- Web framework: net/http
- Database: MongoDB
- Cache: Redis
- Password hashing: bcrypt
- Frontend: HTML, CSS, JavaScript
- MongoDB driver: mongo-go-driver v2
- Redis client: go-redis/v9

## Project Structure

```text
go_login_mongodb_redis/
├── README.md
├── backend/
│   ├── auth_handler.go
│   ├── go.mod
│   ├── main.go
│   ├── middleware.go
│   ├── models.go
│   ├── product_handler.go
│   ├── response.go
│   └── server.go
├── frontend/
│   ├── shopper-ui/
│   └── admin-ui/
└── .gitignore
```

## Prerequisites

Before running the project, ensure the following are available:

- Go 1.22 or newer
- MongoDB running locally on port 27017
- Redis running locally on port 6379
- A browser for frontend access

## Local Setup

### 1. Start MongoDB
Make sure MongoDB is running on:

```text
mongodb://localhost:27017
```

### 2. Start Redis
Ensure Redis is running on:

```text
localhost:6379
```

### 3. Run the backend

```bash
cd backend
go mod tidy
go run .
```

The backend runs on port 8080 by default.

### 4. Run the frontend UIs separately

Run the shopper site in one terminal:

```powershell
cd shopper-ui
python -m http.server 5500
```

Open `http://localhost:5500`.

Run the admin site in a second terminal:

```powershell
cd frontend/admin-ui
python -m http.server 5501
```

Open `http://localhost:5501`.

Both sites use the shared backend at `http://localhost:8080/api`.

## Deploy to Render

The repository includes [render.yaml](render.yaml), which creates three separate services:

- `stockroom-api` — Go backend web service
- `stockroom-shopper` — Shopper UI static site
- `stockroom-admin` — Admin UI static site

In Render, create a new Blueprint from this repository and deploy `render.yaml`. Before the backend starts, set these environment variables on `stockroom-api`:

- `MONGO_URI` — your MongoDB Atlas connection string
- `REDIS_ADDR` — your Redis provider address, including the port when required
- `CORS_ORIGINS` — both deployed UI origins separated by commas, for example `https://stockroom-shopper.onrender.com,https://stockroom-admin.onrender.com`

The deployed frontends automatically use `https://stockroom-api.onrender.com/api`; update that hostname in the three UI `index.html` files if you choose a different Render service name.

Alternative: use VS Code Live Server on each UI folder. Set the API URL in each UI's `index.html` if the backend is deployed:

```html
<script>window.__API_URL__ = "https://your-backend.example.com/api";</script>
```

## Environment Variables

The application supports the following optional environment variables:

| Variable | Default | Purpose |
| --- | --- | --- |
| MONGO_URI | mongodb://localhost:27017 | MongoDB connection string |
| MONGO_DB | loginapp | MongoDB database name |
| REDIS_ADDR | localhost:6379 | Redis connection address |
| PORT | 8080 | Backend server port |
| CORS_ORIGIN | http://localhost:5500 | Allowed frontend origin |

Example:

```bash
export MONGO_URI=mongodb://localhost:27017
export MONGO_DB=loginapp
export REDIS_ADDR=localhost:6379
export PORT=8080
export CORS_ORIGIN=http://localhost:5500
```

## Authentication Flow

### Signup
- User submits name, email, and password
- The server validates inputs
- It checks whether the email is already registered
- Password is hashed with bcrypt
- User record is stored in MongoDB

### Login
- User submits email and password
- The app looks up the user in MongoDB
- Password is verified with bcrypt
- A random session token is generated
- Redis stores the mapping:

```text
session:<token> -> userId
```

- The token is returned as an HTTP-only cookie named `session_id`
- The session expires after 24 hours

### Profile Access
- The backend reads the `session_id` cookie
- It validates the token in Redis
- If valid, it loads the user from MongoDB and returns profile data

### Logout
- The backend deletes the Redis session key
- The cookie is cleared from the browser

## Product Management

Authenticated users can manage products associated with their account.

### Product Data Model
Each product contains:

- ID
- User ID
- Name
- Description
- Price
- Quantity
- Created timestamp
- Updated timestamp

### Product API Behavior
- Product lookup is filtered by `userId`
- Each user can only access their own products
- MongoDB is the source of truth
- Redis stores a user-specific product list cache

### Cache Strategy
The product list endpoint checks Redis first using a key like:

```text
products:<userId>
```

If found, cached products are returned immediately. If not found, the backend queries MongoDB and stores the result for 5 minutes.

The cache is invalidated after:

- Product create
- Product update
- Product delete

This allows fast reads while preserving data consistency.

## API Endpoints

### Authentication

#### POST /api/signup
Creates a new user account.

Request body:

```json
{
  "name": "John Doe",
  "email": "john@example.com",
  "password": "secret123"
}
```

#### POST /api/login
Authenticates a user and sets a session cookie.

Request body:

```json
{
  "email": "john@example.com",
  "password": "secret123"
}
```

#### GET /api/profile
Returns the logged-in user profile.

#### POST /api/logout
Clears the active session.

### Products

#### GET /api/products
Returns all products for the authenticated user.

#### POST /api/products
Creates a new product.

Request body:

```json
{
  "name": "Laptop",
  "description": "Gaming laptop",
  "price": 1200.5,
  "quantity": 5
}
```

#### PUT /api/products/:id
Updates a specific product.

#### DELETE /api/products/:id
Deletes a product by ID.

## Security Notes

- Passwords are not stored in plain text; they are hashed with bcrypt.
- Session tokens are stored in Redis and associated with a user ID.
- Cookies are configured with `HttpOnly` and `SameSite=Lax`.
- Authentication is checked before product operations.
- User access is restricted to their own products.

## Production Deployment

To deploy this project in production:

1. Set a real MongoDB connection string for production.
2. Set Redis connection details for the deployment environment.
3. Update `CORS_ORIGIN` to the deployed frontend URL.
4. Set the frontend API URL to the deployed backend service.
5. Use HTTPS for both frontend and API.

Example frontend configuration:

```html
<script>
  window.__API_URL__ = "https://your-backend-service.com/api";
</script>
```

## Known Behavior

- The frontend is intentionally simple and uses fetch-based calls.
- The backend is stateless apart from Redis session storage and MongoDB state.
- If Redis is unavailable, product reads continue to work by falling back to MongoDB.

## Troubleshooting

### Backend fails to connect to MongoDB
Check that MongoDB is running and accessible at the configured URI.

### Backend fails to connect to Redis
Verify Redis is installed and started on the configured host and port.

### Frontend cannot call the API
Confirm that:

- the backend is running on port 8080
- the frontend is pointing to the correct API URL
- the CORS configuration matches the frontend origin

### Login works but profile fails
Typical causes:

- missing or expired `session_id` cookie
- Redis session was removed or expired
- backend is running with a different Redis instance

## Summary
This project is a practical example of a secure, full-stack application built with Go and modern browser-based frontend technologies. It demonstrates core patterns used in production systems, including authentication, authorization, caching, MongoDB persistence, and session management.

It is well suited for learning, prototyping, and showcasing backend/frontend integration in a small but realistic application.
