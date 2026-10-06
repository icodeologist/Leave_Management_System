# Business Requirements Document

## 1. Purpose

Leave Management System gives employees a simple way to request leave and gives administrators a reliable way to review those requests and manage balances.

The first release is intentionally focused on the complete request-to-approval workflow.

## 2. Users

### Employee

An employee needs to:

- Sign up and sign in.
- See how many leave days remain.
- Submit a leave request.
- Choose the leave type and duration.
- See previous requests and their current status.

### Admin

An admin needs to:

- Sign up and sign in.
- See leave requests from employees.
- Review the employee, dates, reason, and requested duration.
- Approve or reject pending requests.
- See the effect of approval on the employee’s balance.

## 3. Business Rules

- Every new employee receives 20 leave days.
- Full-day leave uses 1 day.
- Half-day leave uses 0.5 days and must use one date.
- Requests use YYYY-MM-DD dates.
- Past dates, invalid date ranges, overlapping pending/approved requests, and requests over the available balance are rejected.
- Every new request starts as PENDING.
- Only a PENDING request can be approved or rejected.
- Rejected requests do not reduce the employee’s balance.
- Approved requests reduce the balance by the requested duration.
- Approval and balance deduction are one atomic operation.
- The leave request and employee balance rows are locked while approval is processed. This prevents two admins from approving the same request or deducting the same balance twice.

## 4. Main User Journeys

### Employee submits leave

1. The employee signs in.
2. The employee checks the current balance.
3. The employee submits the leave type, dates, duration, and reason.
4. The system validates the request and records it as PENDING.
5. The request appears in the employee history and admin dashboard.

### Admin reviews leave

1. The admin signs in.
2. The admin opens the request list.
3. The admin reviews the employee and request details.
4. The admin approves or rejects the request.
5. The employee sees the updated status and balance.

## 5. Release Acceptance Criteria

The release is complete when:

- An employee can register, sign in, submit leave, and view history.
- An admin can register, sign in, view requests, and approve or reject them.
- The employee balance changes only after approval.
- Two concurrent approval attempts cannot approve one request twice.
- Employees cannot access admin actions.
- Admins cannot access another employee’s private history.
- The deployed frontend can communicate with the deployed backend.
- The backend health check returns successfully.

## 6. Deployment

The live application uses:

- Frontend: Vercel — https://leave-management-system-git-main-denz18.vercel.app/
- Backend: Render — https://leave-management-api-xh4m.onrender.com
- Database: Render PostgreSQL
- Backend packaging: Docker from `backend/Dockerfile`

Deployment requires:

- `DATABASE_URL`
- `JWT_SECRET`
- `JWT_EXPIRY`
- `ALLOWED_ORIGIN`

To update the application, push the changes to GitHub. Vercel rebuilds the frontend and Render rebuilds the Docker backend. After deployment, verify the backend health check and complete one employee-to-admin leave workflow.
