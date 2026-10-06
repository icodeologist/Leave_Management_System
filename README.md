# Leave Management System

A simple leave request system with a React frontend and Go/PostgreSQL backend.

**Requirements document:** [docs/BRD.md](docs/BRD.md)

## Features

- Employee and admin registration/login with JWT and bcrypt.
- Employees see their balance, submit annual/casual/sick leave, choose full or half day, and view history.
- Admins view requests, filter by status, and approve or reject pending requests.
- Employee balance starts at 20 days and is deducted only after approval.
- Leave balance validation and overlapping-date validation.
- Public `/api/health` endpoint for Render.

### Demo admin registration

For this demo, the registration form allows a user to select `ADMIN`. This is intentionally public for testing and is not safe for a production release because anyone could create an admin account.

To create a test admin, open registration, enter the admin details, select `Admin`, and submit the form. Use those credentials to open the admin dashboard.

In a future iteration, public admin registration must be removed. Admins should then be created by a private seed process or by an existing authorized administrator.

## Data consistency

Approval uses one PostgreSQL transaction. The leave request row and employee row are locked while approval is checked and written. This prevents two admins from approving the same request or deducting the same employee balance twice. Admins can still view the dashboard at the same time; only conflicting updates are serialized.

The backend handles CORS `OPTIONS` preflight requests before authentication and returns the required headers with a `204` response.

## Stack

- Frontend: React, Vite, JavaScript, Tailwind CSS
- Backend: Go REST API, GORM, PostgreSQL
- Deployment: Vercel frontend, Render Docker backend and PostgreSQL

## Architecture

```text
Browser
  └─ React/Vite frontend on Vercel
       └─ API client sends JSON requests and Bearer JWTs
            ↓ HTTPS REST API
Go API on Render
  ├─ HTTP handlers: decode input, validate requests, return JSON responses
  ├─ Auth middleware: validate JWT and attach user ID and role to request context
  ├─ Store layer: keep GORM database operations outside the HTTP handlers
  ├─ Auth package: bcrypt password checks and JWT creation/validation
  └─ Database layer: open PostgreSQL connection and run GORM AutoMigrate
            ↓ GORM
PostgreSQL on Render
  ├─ users
  └─ leave_requests
```

Public routes handle health checks, registration, and login. Protected routes require a Bearer token; role checks then limit employee and admin operations. Employees can only query their own history, while admins can review and update leave requests.

Leave approval locks both the leave request and employee rows inside one PostgreSQL transaction. The balance check, balance deduction, and status update either all succeed or all roll back. This prevents concurrent admin actions from approving one request twice or deducting the same balance twice.

## Run locally

Requirements: Go 1.23+, PostgreSQL, Node.js, and npm.

```bash
cp backend/.env.example backend/.env
cd backend
go run ./cmd/api
```

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

GORM creates the required tables on backend startup.

## Environment variables

Backend:

```env
DATABASE_URL=<PostgreSQL connection URL>
JWT_SECRET=<at least 32 characters>
JWT_EXPIRY=24h
ALLOWED_ORIGIN=<exact frontend origin>
```

Render supplies `PORT`; local runs default to `8080`.

Frontend:

```env
VITE_API_URL=<backend URL>
```

## API

| Method | Endpoint | Access |
| --- | --- | --- |
| GET | `/api/health` | Public |
| POST | `/api/auth/register` | Public |
| POST | `/api/auth/login` | Public |
| GET | `/api/me` | Authenticated |
| POST | `/api/leaves` | Employee |
| GET | `/api/leaves/my` | Employee |
| GET | `/api/admin/leaves` | Admin |
| PATCH | `/api/admin/leaves/:id/approve` | Admin |
| PATCH | `/api/admin/leaves/:id/reject` | Admin |
| GET | `/api/admin/dashboard` | Admin |

## Deploy

### Render backend

Set the service root directory to `backend`, use `./Dockerfile`, and configure:

```env
DATABASE_URL=<Render PostgreSQL internal URL>
JWT_SECRET=<production secret>
JWT_EXPIRY=24h
ALLOWED_ORIGIN=https://leave-management-system-git-main-denz18.vercel.app
```

Health check path:

```text
/api/health
```

### Vercel frontend

Set:

```env
VITE_API_URL=https://leave-management-api-xh4m.onrender.com
```

Live URLs:

- Frontend: https://leave-management-system-git-main-denz18.vercel.app/
- Backend: https://leave-management-api-xh4m.onrender.com
- Health: https://leave-management-api-xh4m.onrender.com/api/health

For CORS, `ALLOWED_ORIGIN` must exactly match the browser origin, without a trailing slash. See [docs/BRD.md](docs/BRD.md) for the short business requirements.
