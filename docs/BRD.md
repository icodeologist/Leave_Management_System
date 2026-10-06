# Business Requirements Document

## Product

Leave Management System lets employees request leave and admins review those requests. The first release focuses on a small, reliable leave workflow.

## Users

### Employee

- Register and log in.
- See the current leave balance.
- Submit annual, casual, or sick leave.
- Submit full-day or half-day leave.
- See personal request history and status.

### Admin

- Register and log in.
- View employee leave requests.
- Filter requests by status.
- Approve or reject pending requests.

## Leave rules

- New employees start with 20 days.
- Full day uses 1 day; half day uses 0.5 days.
- Half-day dates must be the same.
- Dates use YYYY-MM-DD and cannot be in the past.
- Requests cannot exceed the available balance or overlap pending/approved requests.
- New requests are PENDING.
- Only pending requests can be approved or rejected.
- Rejected requests do not change the balance.

## Data consistency

Approval is atomic: it is handled in one PostgreSQL transaction. The leave request row and employee row are locked while the balance is checked and updated. This serializes conflicting admin actions and prevents the same request from being approved twice or the balance being deducted twice.

## Technical requirements

- React/Vite/Tailwind frontend.
- Go REST API with GORM and PostgreSQL.
- JWT authentication and bcrypt password hashing.
- Role-based employee and admin routes.
- GORM AutoMigrate on startup.
- Public GET /api/health endpoint.
- CORS preflight support through OPTIONS requests.

## Deployment

The frontend is hosted on Vercel:

https://leave-management-system-git-main-denz18.vercel.app/

The backend and PostgreSQL database are hosted on Render:

https://leave-management-api-xh4m.onrender.com

Health check:

https://leave-management-api-xh4m.onrender.com/api/health

### Deploy or update

1. Push the repository to GitHub.
2. In Render, set the service root directory to backend and use backend/Dockerfile.
3. Connect a Render PostgreSQL database.
4. Set DATABASE_URL, JWT_SECRET, JWT_EXPIRY, and ALLOWED_ORIGIN in Render.
5. Set ALLOWED_ORIGIN to the exact Vercel origin without a trailing slash.
6. Set Render health check path to /api/health.
7. In Vercel, set VITE_API_URL to the Render backend URL.
8. Push future changes to GitHub and verify the health endpoint after redeployment.
