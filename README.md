# 🎮 Game Wallet & Ledger Service

A high-performance, double-entry ledger system built with **Go (Golang)**, **Gin**, and **PostgreSQL**. Designed to handle transactional game economy features like currency balances, inventory management, treasury audits, and referral bonuses with optimistic locking for data integrity.

![Architecture Diagram](./architecture-diagram.png)
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
