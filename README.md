# Chirpy - Twitter-like backend in GO

---

## Overview

Chirpy is a minimalist twitter-like backend system built with go. It provides a restful api for user authentication, posting and retreiving messages (chirps), webhook support, and persistent storage using PostgreSQL.

## Features

- User authentication & authorization (JWT-based)
- CRUD operations for chirps (posts)
- Webhook event support for external integrations
- PostgreSQL database with schema migrations
- RESTful API architecture

## Tech Stack

- **Go** - Backend development
- **sqlc** - SQL query generation
- **Goose** - Database migrations
- **PostgreSQL** - Database storage

## Installation & Setup

### Prerequisites

Ensure you have the following installed:

- Go (>=1.18)
- PostgreSQL (>=13)
- `sqlc`
- `goose`

### Clone the Repository

```sh
git clone https://github.com/thetsajeet/chirpy.git
cd chirpy
cp .example.env .env
# fill the env variables
```

### Migrations

```sh
cd /sql/schema
goose postgres <connection-string> up #up
goose postgres <connection-string> down #down
```

### Generate queries

```sh
sqlc generate
```

### Start the server

```sh
go run .
```
