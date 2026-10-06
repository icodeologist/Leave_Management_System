# Leave Management System

Leave Management System is a small full-stack application for submitting, reviewing, and tracking employee leave. It has a Go API, a PostgreSQL database, and a React frontend.

The project deliberately keeps the architecture simple:

- `backend/` contains the Go API, GORM models, database setup, authentication, and Dockerfile.
- `frontend/` contains the React/Vite application.
- The backend creates the required tables with GORM when it starts. There is no separate migration command.

## Features

Employees can register, sign in, see their leave balance, submit leave requests, and view their request history. A request can be annual, casual, or sick leave, and can be either a full day or a half day.

Admins can sign in to a separate dashboard, view employee requests, filter them by status, and approve or reject pending requests. The API checks the employee’s balance and deducts leave only after approval. Approval and deduction happen in the same database transaction.

The API uses JWTs for authentication and bcrypt for password hashing. Employee and admin access is enforced by the JWT role claim and backend middleware.

## API Reference

| Method | Endpoint | Who can use it |
| --- | --- | --- |
| `GET` | `/api/health` | Anyone |
| `POST` | `/api/auth/register` | Anyone |
| `POST` | `/api/auth/login` | Anyone |
| `GET` | `/api/me` | Signed-in users |
| `POST` | `/api/leaves` | Employees |
| `GET` | `/api/leaves/my` | Employees |
| `GET` | `/api/admin/leaves` | Admins |
| `PATCH` | `/api/admin/leaves/:id/approve` | Admins |
| `PATCH` | `/api/admin/leaves/:id/reject` | Admins |
| `GET` | `/api/admin/dashboard` | Admins |

Protected requests must include:

```text
Authorization: Bearer <jwt>
```

## Run Locally

### Requirements

- Go 1.23 or newer
- PostgreSQL
- Node.js and npm

### 1. Create the backend environment file

From the repository root:

```bash
cp backend/.env.example backend/.env
```

Open `backend/.env` and set a real local PostgreSQL connection and a JWT secret with at least 32 characters. `DATABASE_URL` and `JWT_SECRET` are required. The API uses port `8080` by default.

### 2. Start the backend

```bash
cd backend
go run ./cmd/api
```

The first startup connects to PostgreSQL and creates or updates the `users` and `leave_requests` tables through GORM.

Check that it is running:

```bash
curl http://localhost:8080/api/health
```

Expected response:

```json
{"status":"ok"}
```

### 3. Start the frontend

Create `frontend/.env.local`:

```env
VITE_API_URL=http://localhost:8080
```

Then run:

```bash
cd frontend
npm install
npm run dev
```

Open the Vite URL shown in the terminal, normally `http://localhost:5173`.

## Run with Docker

The Docker build context is `backend/`. Build and run it from that directory:

```bash
cd backend
docker build -t leave-backend .
docker run --rm -p 8081:8080 --env-file .env leave-backend
```

The container reads configuration from environment variables. It uses `PORT` when it is provided and falls back to `8080` for local runs.

## Deploy the Backend

The repository is already arranged for a Docker deployment:

1. Create a Render PostgreSQL database.
2. Create a Render Web Service connected to this repository.
3. Set the service Root Directory to `backend`.
4. Select Docker as the runtime and use `./Dockerfile`.
5. Add the environment variables below.
6. Set the health check path to `/api/health`.

Set these variables in Render:

```env
DATABASE_URL=<Render PostgreSQL internal connection URL>
JWT_SECRET=<random secret with at least 32 characters>
JWT_EXPIRY=24h
ALLOWED_ORIGIN=https://leave-management-system-git-main-denz18.vercel.app
```

Render supplies `PORT` automatically. Do not hardcode it in Render.

The backend fails at startup if `DATABASE_URL` or `JWT_SECRET` is missing. Once the service is live, check:

```bash
curl https://leave-management-api-xh4m.onrender.com/api/health
```

## Deploy the Frontend

Set the frontend build variable to the public Render API URL:

```env
VITE_API_URL=https://leave-management-api-xh4m.onrender.com
```

For the current deployment, the API URL is:

```env
VITE_API_URL=https://leave-management-api-xh4m.onrender.com
```

The value of `ALLOWED_ORIGIN` on Render must exactly match the URL in the browser address bar. For example, these are different origins:

```text
https://leave-management-system-denz18.vercel.app
https://leave-management-system-git-main-denz18.vercel.app
```

Use one exact origin, without a trailing slash. If you move between Vercel preview and production URLs, update `ALLOWED_ORIGIN` and redeploy the backend.

## Troubleshoot CORS

The backend handles `OPTIONS` preflight requests before authentication and returns `204` with the CORS headers for the configured origin.

If the browser says that `Access-Control-Allow-Origin` is missing:

1. Look at the browser error and copy the exact value after `origin`.
2. Set that exact value as Render’s `ALLOWED_ORIGIN`.
3. Remove any old `CORS_ORIGIN` variable from Render.
4. Save the environment variable and redeploy the backend.
5. Confirm that `VITE_API_URL` points to the Render API, not `localhost`.

You can test a preflight request yourself:

```bash
curl -i -X OPTIONS \
  https://leave-management-api-xh4m.onrender.com/api/auth/login \
  -H "Origin: https://leave-management-system-git-main-denz18.vercel.app" \
  -H "Access-Control-Request-Method: POST" \
  -H "Access-Control-Request-Headers: content-type"
```

The response should be `204` and include:

```text
Access-Control-Allow-Origin: https://leave-management-system-git-main-denz18.vercel.app
Access-Control-Allow-Methods: GET, POST, PATCH, OPTIONS
Access-Control-Allow-Headers: Authorization, Content-Type
```

## Live URLs

Frontend:

```text
https://leave-management-system-git-main-denz18.vercel.app
```

Backend:

```text
https://leave-management-api-xh4m.onrender.com
```

Health check:

```text
https://leave-management-api-xh4m.onrender.com/api/health
```

## Project Documentation

The full requirements are documented in [`docs/BRD.md`](docs/BRD.md). Keeping the BRD in the repository means the requirements are versioned with the code and easy to review on GitHub.
