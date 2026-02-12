# 🎮 Game Wallet & Ledger Service

A high-performance, double-entry ledger system built with **Go (Golang)**, **Gin**, and **PostgreSQL**. Designed to handle transactional game economy features like currency balances, inventory management, treasury audits, and referral bonuses with optimistic locking for data integrity.

---

## ⚠️ Due to limited Azure student credits, my backend server might stop soon !, please don't mind if it fails unfortunately while reviewing. But you can always run the app locally using the below steps.

## ✨ Live Link : 👉 [game-wallet ↗](https://happy-wave-06fe03900.4.azurestaticapps.net/)

## Architecture
![Architecture Diagram](https://github.com/Pramod-325/game-transactions/blob/main/game-wallet-demo/public/game-wallet-arch.png)
*(Place your architecture image in the root folder and name it `architecture-diagram.png`)*

---
## 🔗 Prisma Schema- ERD
![Architecture Diagram](https://github.com/Pramod-325/game-transactions/blob/main/game-wallet-demo/public/tables.jpg)
*(Place your architecture image in the root folder and name it `architecture-diagram.png`)*

---

## 🚀 Features

* **Double-Entry Ledger:** Every transaction is recorded (Credit/Debit) ensuring mathematical accuracy.
* **Atomic Transactions:** All DB operations (User + Treasury updates) happen in a single ACID transaction.
* **Optimistic Locking:** Prevents race conditions during high-concurrency balance updates.
* **Inventory System:** Manages virtual assets like Gold Coins and Treasure Boxes using `Int32` (supports upto ~2+ Billion In-game System Tokens/Assets).
* **JWT Authentication:** Secure stateless authentication.
* **Referral System:** Automated bonuses for both referrer and referee, deducted from the System Treasury.


---

## ⚒ Core Tasks (The Flows)
**Status**: Successfully implemented all three required functional flows. ✅

- **Wallet Top-up**: Implemented in ProcessTopUp (and exposed via /top-up). It correctly uses the "Hybrid" flow to credit the user instantly and queue the treasury update.

- **Bonus/Incentive**: Implemented in Signup (via ProcessBonus). It correctly issues referral bonuses to both the new user and the referrer.

- **Purchase/Spend**: Implemented in ProcessPurchase (and exposed via /purchase). It calculates costs and deducts from the user's wallet.

**⚠️ Critical Constraints (The "Hard" Part) Status:** PARTIALLY FAILED The current code handles some concurrency aspects but fails on strict overdraft protection and idempotency due to time constraints.

## 🛠️ Tech Stack

* **Language:** Go (Golang) 1.25+
* **Framework:** Gin Web Framework
* **Database:** PostgreSQL (Neon DB)
* **ORM:** Prisma Client Go
* **Containerization:** Optimized Docker multi-stage build (~120MB image).

---

## 💻 For Running Locally:

your have two options: build from source, or pull the docker image.

## ⚙️ Configuration (.env)

⚠ Create a `.env` file inside the `./game-wallet-demo` root directory only not in cloned home directory and using double-quotes around the values be verifed properly, while running locally it's fine, but remove double-quotes while dockerizing!

```ini
# Database Connection (Ensure it starts with postgresql://)
# For Local Docker: Use host.docker.internal instead of localhost

DATABASE_URL=postgresql://username:password@host:5432/DB-name?schema=public&sslmode=disable # example: double quotes not used here

# Security
JWT_SECRET="your-JWT-secret-string" # double quotes used

FRONTEND_URL=http://localhost:5173 #Your frontend domain
```

### Quick and easy way Docker Pull:
#### Prerequisites
- latest go version ~1.25+ to be installed, if not download from here -> [Go download](https://go.dev/)
- docker `check  using "docker version" command`, if its not there install from here as per your OS -> [Docker](https://docs.docker.com/get-started/get-docker/)
- Run a Postgress Image using docker or easy way create an account on [NeonDB](https://console.neon.tech/) then copy & paste the connection string from your project dashboard into your `.env file`
- NodeJs latest ~v24 (Only to run frontend files) download from here -> [NodeJs](https://nodejs.org/en/download)

```bash
# Order of command Steps in your terminal/cmd prompt
1. git clone https://github.com/Pramod-325/game-transactions.git
2. cd game-transactions
```

```ini
Open two terminals two run backend and front end separately
```
```bash
# To run backend in Terminal-1
# Ensure you have your .env file in this (game-wallet-demo) folder
- cd game-wallet-demo

## Option-1
1. docker pull telnozo72/game-wallet:v1
2. docker run -d --name game-wallet --env-file .env -p 8080:8080 game-wallet

# =====================( or )======================

# option-2 run from source

# download dependencies
1. go mod download

# Generate the Prisma Client Go code
2. go run github.com/steebchen/prisma-client-go generate

# Push the schema to the DB (Creates tables)
3. go run github.com/steebchen/prisma-client-go db push

# Seeding DB
4. go run cmd/seed/main.go

# Start the server
5. go run main.go

```

Backend Server will start at [http://localhost:8080](http://localhost:8080)
```
Ensure your prisma client is ready and not conflicting in "prisma/db" folder
```
```bash
# Open Terminal-2 to run frontend
- cd frontend
ensure .env file with "VITE_API_URL" set to your backend address

# install dependencies
1. npm install
2. npm run dev
```

Frontend Server will start at [localhost:5173](http://localhost:5173)


# 📚 Game Wallet Working API Documentation & Backend Processes

This section details the internal engineering ensuring data integrity, concurrency safety, and performance.

## 🔐 Authentication Process

The system uses **Stateless JWT** (JSON Web Token) authentication.

1. **Login**: User exchanges credentials for a signed `Bearer Token`
2. **Middleware**: Validates the signature using `JWT_SECRET` and injects the `UserID` into the Gin Context for handlers to use safely

---

## ⚙️ Core Engineering Concepts

### 1. Optimistic Locking (Concurrency Control)

**Problem**: "Double Spending" race conditions when multiple requests try to spend money simultaneously.

**Solution**: Every Wallet (Account) has a `version` integer.

**Execution**:
- Transaction A reads `Version: 1`
- Transaction B reads `Version: 1`
- Transaction A updates balance and sets `Version: 2` → Success
- Transaction B tries to update `WHERE version = 1` → Fails (Version is now 2)
- **Result**: Transaction B is rejected (409 Conflict), preserving data integrity

### 2. Double-Entry Ledger System

**Philosophy**: Money is never destroyed, only moved.

**Flow**: A purchase consists of two distinct movements:
- **Debit**: User Wallet (Balance decreases)
- **Credit**: System Treasury (Balance increases)

**Auditability**: `Sum(User Balances) + Sum(Treasury)` always equals the total economy size.

### 3. ⚡ Async Treasury Worker (The "Batcher")

**The Bottleneck**: In a naive system, thousands of users buying items simultaneously would all fight for a lock on the single System Treasury row, causing massive latency.

**The Solution**: A Hybrid Transaction Model implemented in `internal/worker/treasury_batcher.go`:

1. **User Side (Synchronous)**: The user's balance and inventory are updated immediately in a strict ACID transaction. The user gets a generic "Success" response instantly.
2. **Treasury Side (Asynchronous)**: The credit to the Treasury is pushed to a Go Channel (`queue`).
3. **The Batcher**: A background worker groups these queued tasks (e.g., 100 at a time) and executes a single bulk update to the Treasury.

**Performance**: Reduces database lock contention on the Treasury row from N times per second to ~2 times per second.

---

# 📡 API Endpoints Reference

## 1. System Health
- **Endpoint**: `GET /health`
- **Auth**: ❌ Public
- **Description**: Checks if the server is running and pings the Database
- **Response**: `{ "status": "UP", "database": "CONNECTED" }`

## 2. User Signup
- **Endpoint**: `POST /signup`
- **Auth**: ❌ Public
- **Body**: `{ "username": "player1", "password": "pass", "referralCode": "opt" }`
- **Process**: Atomic creation of User, Wallet, and Inventory

## 3. User Login
- **Endpoint**: `POST /login`
- **Auth**: ❌ Public
- **Body**: `{ "username": "player1", "password": "pass" }`
- **Response**: Returns JWT Bearer Token

## 4. Get Balance
- **Endpoint**: `GET /balance`
- **Auth**: ✅ Protected
- **Response**: `{ "balance": 1000, "version": 5 }`

## 5. Top-Up (Deposit)
- **Endpoint**: `POST /top-up`
- **Auth**: ✅ Protected
- **Body**: `{ "amount": 100 }`
- **Description**: Adds funds to the user's wallet via the Hybrid/Ledger flow

## 6. Purchase Item
- **Endpoint**: `POST /purchase`
- **Auth**: ✅ Protected
- **Body**: `{ "item": "gold_coin" }`
- **Description**: Atomically deducts balance and adds item. Fails if funds are insufficient or if a race condition (optimistic lock) occurs

---

# 🧠 Key Learnings

- **Developing a Go Application 😅**: I acknowledge using of gemini AI to help me code, Since I'm new to Go lang (for the project's requirement), I understand the drawbacks of AI generated code and need for manual control, as of now the project works without any problems.
- **Race Conditions**: Implemented Optimistic Locking to handle high-concurrency spending safely
- **Docker Optimization**: Reduced image size from 1.8GB to ~60MB using Multi-Stage Builds and Alpine Linux
- **Hybrid Architecture**: Learned to split critical user-facing consistency (Sync) from backend aggregation (Async) using Go Channels
- **Async Worker Pattern**: Implemented a buffered channel worker in Go to decouple high-frequency user writes from the central treasury bottleneck


Made with 💖 for Dino Ventures 😉
