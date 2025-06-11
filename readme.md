# 📘 Go Payroll

A simple payroll management backend built in Go using Fiber and GORM.

---
## ▶️ How to Run the App

1. **Clone the repository**
    ```bash
    git clone https://github.com/your-username/go-payroll.git
    cd go-payroll
    ```
2. **Set up environment variables**
    Create a .env file in the root directory:

    ```bash
    DB_DSN="host=localhost user=postgres password=root dbname=payroll port=5432 sslmode=disable"
    ```
3. **Create the PostgreSQL database**

    ```bash
    CREATE DATABASE payroll;
    ```
4. **Install dependencies**

    ```bash
    go mod tidy
    ```
5. **Run the application**

    ```bash
    go run main.go
    ```

## 📦 Features

- User authentication with JWT-based role access (admin & employee)
- **Employee Functions:**
  - Submit daily attendance
  - Submit overtime and reimbursement requests
  - View individual payslips
- **Admin Functions:**
  - Create attendance periods for employees
  - Generate payslip summaries for all employees
  - Run and freeze payroll for a specific period
- Audit logging for all requests including user ID, IP address, and endpoint access
---

## 🚀 Tech Stack

- **Go (Golang)**
- **Fiber** web framework
- **GORM** ORM
- **PostgreSQL** (or your configured DB)
- **UUID** and custom middleware

---

## 🛠️ Prerequisites

Before you begin, make sure you have:

1. Go 1.18+
2. PostgreSQL installed and running
3. Git

---

## 📂 Project Structure
```bash
go-payroll/
├── config/ # DB and app config
├── controllers/ # Route handlers
├── middleware/ # JWT guard and audit logging
├── models/ # GORM models
├── routes/ # Route definitions
├── utils/ # Utility functions (rounding, etc.)
├── main.go # Entry point
├── go.mod / go.sum # Go dependencies
└── README.md # You are here
```
---

## 🔐 API Endpoints

### Auth
- `POST /api/login` – Generate JWT token for either admin or employee (based on credentials)

### Admin
> Requires `admin` JWT token
- `POST /api/admin/attendance-period` – Create attendance period for selected employees
- `GET /api/admin/payslip-summary` – View total take-home pay for all unpaid employees
- `POST /api/admin/run-payroll` – Process payslips

### Employee
> Requires `employee` JWT token
- `POST /api/employee/attendance` – Submit attendance for a given date
- `POST /api/employee/overtime` – Submit overtime request
- `POST /api/employee/reimbursement` – Submit reimbursement request
- `GET /api/employee/payslip` – View personal payslips

---

## 📫 Postman Collection

[📎 Open in Postman](https://pk-8575591.postman.co/workspace/PK's-Workspace~fd5522e8-c8ab-4d5d-85e9-6a06f33b7be8/collection/45765118-9081d3e9-c0c1-4ed9-80bf-8f7237dea03c?action=share&creator=45765118)

---
