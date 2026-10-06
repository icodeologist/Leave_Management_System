# Business Requirements Document

## 1. Product Overview

Leave Management System is a web application that allows employees to request leave and allows administrators to review, approve, or reject those requests. The system keeps track of leave balances and gives employees a simple view of their request history.

The first release focuses on the core leave workflow. It is intentionally small so it can be deployed and used quickly.

## 2. Users and Roles

### Employee

An employee can:

- Create an account and sign in.
- View their current leave balance.
- Submit a leave request.
- Choose annual, casual, or sick leave.
- Choose a full-day or half-day request.
- View their own leave history and request status.

### Admin

An admin can:

- Create an account and sign in.
- View the admin dashboard.
- View leave requests submitted by employees.
- Filter requests by status.
- Approve or reject pending requests.
- See the employee and leave details needed to review a request.

## 3. Leave Rules

- A new employee starts with 20 days of leave balance.
- A full-day leave uses one day.
- A half-day leave uses 0.5 days.
- Half-day leave must start and end on the same date.
- Leave dates use the `YYYY-MM-DD` format.
- A leave request cannot use dates in the past.
- A leave request cannot extend outside the current year.
- A request cannot exceed the employee’s available balance.
- Overlapping pending or approved requests are rejected.
- New requests start with `PENDING` status.
- Only pending requests can be approved or rejected.
- Leave balance is deducted only when a request is approved.
- Approval and balance deduction must happen in one PostgreSQL transaction.

## 4. Main Workflows

### Employee leave request

1. The employee signs in and receives a JWT.
2. The employee opens the leave request form.
3. The frontend sends the request with the JWT in the `Authorization` header.
4. The backend validates the employee role, dates, leave type, day type, reason, and available balance.
5. A valid request is stored as `PENDING`.
6. The request appears in the admin dashboard and the employee’s history.

### Admin review

1. The admin signs in and receives a JWT.
2. The admin opens the dashboard and views employee requests.
3. The admin approves or rejects a pending request.
4. On approval, the backend locks the relevant records, checks the balance again, deducts the leave, and changes the request to `APPROVED` in one transaction.
5. On rejection, the request changes to `REJECTED` and the balance is not deducted.
6. The employee sees the updated status and balance after refreshing the application.

## 5. Functional Requirements

The system provides:

- Registration and login for employees and admins.
- JWT authentication and role-based route protection.
- Bcrypt password hashing.
- Employee balance display and leave request creation.
- Employee leave history ordered from newest to oldest.
- Admin dashboard with request status filters.
- Approve and reject actions for pending requests.
- PostgreSQL persistence through GORM.
- Automatic table setup through GORM `AutoMigrate` at startup.
- A public `/api/health` endpoint for hosting platform health checks.

## 6. API Surface

| Method | Endpoint | Access |
| --- | --- | --- |
| `GET` | `/api/health` | Public |
| `POST` | `/api/auth/register` | Public |
| `POST` | `/api/auth/login` | Public |
| `GET` | `/api/me` | Authenticated user |
| `POST` | `/api/leaves` | Employee |
| `GET` | `/api/leaves/my` | Employee |
| `GET` | `/api/admin/leaves` | Admin |
| `PATCH` | `/api/admin/leaves/:id/approve` | Admin |
| `PATCH` | `/api/admin/leaves/:id/reject` | Admin |
| `GET` | `/api/admin/dashboard` | Admin |

## 7. Technical Architecture

- **Frontend:** React, Vite, JavaScript, and Tailwind CSS.
- **Backend:** Go REST API using the standard HTTP package.
- **Database:** PostgreSQL accessed through GORM.
- **Authentication:** JWT access tokens and bcrypt password hashing.
- **Deployment:** Dockerized Go backend on Render and React frontend on Vercel.
- **Communication:** JSON REST requests from the frontend to the backend.
- **CORS:** The backend accepts requests from the configured `ALLOWED_ORIGIN` and handles `OPTIONS` preflight requests before authentication.

## 8. Deployment

### Hosting

- Frontend hosting: Vercel
- Backend hosting: Render Web Service
- Database hosting: Render PostgreSQL
- Backend packaging: Docker using `backend/Dockerfile`

Live application URLs:

- Frontend: https://leave-management-system-git-main-denz18.vercel.app/
- Backend: https://leave-management-api-xh4m.onrender.com
- Health check: https://leave-management-api-xh4m.onrender.com/api/health

### Required backend environment variables

Configure these variables in the Render Web Service:

```env
DATABASE_URL=<Render PostgreSQL internal connection URL>
JWT_SECRET=<random secret with at least 32 characters>
JWT_EXPIRY=24h
ALLOWED_ORIGIN=https://leave-management-system-git-main-denz18.vercel.app
```

Render provides `PORT` automatically. The backend uses it when it is present and uses port `8080` for local runs.

### Render setup

1. Push the repository to GitHub.
2. Create a PostgreSQL database in Render.
3. Create a Render Web Service connected to the repository.
4. Set the service Root Directory to `backend`.
5. Select Docker and use `./Dockerfile` as the Dockerfile path.
6. Add the environment variables listed above.
7. Set the health check path to `/api/health`.
8. Deploy the service and confirm that `/api/health` returns `{"status":"ok"}`.

### Vercel setup

1. Import the repository into Vercel.
2. Set the project root to `frontend`.
3. Set the build environment variable:

```env
VITE_API_URL=https://leave-management-api-xh4m.onrender.com
```

4. Deploy the frontend.
5. Copy the exact frontend URL into the backend’s `ALLOWED_ORIGIN` variable.
6. Redeploy the backend after changing `ALLOWED_ORIGIN`.

### Updating the application

1. Make and verify the code change locally.
2. Commit and push the change to GitHub.
3. Render rebuilds and redeploys the backend from the updated Dockerfile and Go source.
4. Vercel rebuilds and redeploys the frontend from the updated React source.
5. Verify `/api/health`, login, and the affected user workflow.

## 9. Non-Functional Requirements

- Protected endpoints must reject missing or invalid JWTs.
- Employees must only access their own leave history.
- Admin endpoints must reject employee requests.
- Passwords must never be stored in plain text.
- Database errors should return a clear HTTP error response without exposing secrets.
- The backend must start with an empty PostgreSQL database and create its required tables automatically.
- The application should remain simple enough to run locally with a small number of commands.

## 10. Out of Scope for the First Release

- Email notifications.
- Password reset and email verification.
- Multiple leave policies or department-specific balances.
- Scheduled yearly balance resets.
- Advanced reporting and exports.
- Frontend design customization beyond the core employee and admin screens.
