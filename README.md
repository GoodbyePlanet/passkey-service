### Passkey Service

A simple Go + Gin service implementing passkey (WebAuthn) registration and authentication flows.
PostgreSQL is used as a persistence layer.

This service exposes the following REST API endpoints:

```
POST /api/register/begin – Start user registration
POST /api/register/finish – Complete user registration
POST /api/authenticate/begin – Start user authentication
POST /api/authenticate/finish – Complete user authentication
```

#### Features

```
User registration with passkeys (WebAuthn)
User authentication with registered passkeys
PostgreSQL persistence for users and credentials
Dockerized service and database for easy setup
Makefile for local development and Docker commands
```

#### Setup

Copy the example environment file:

```shell
cp .env.example .env.development
```

Edit .env.development with your local database settings (host, user, password, db name, port).

Start PostgreSQL only (optional for local development):

```shell
make db-up
```

Run the service locally:

```shell
make run
```


The API will start on http://localhost:8085.

Alternatively, run everything with Docker:

```shell
make up
```

#### Makefile Commands
```
make build	        - Compile Go binary locally
make run	        - Run the app locally
make docker-build	- Build Docker image
make up	            - Start API + DB containers in background
make db-up	        - Start only PostgreSQL container
make logs	        - Tail API logs live
make db	            - Open PostgreSQL shell inside container
make clean	        - Stop containers, remove binary and volumes
```

#### Testing passkey registration and authentication flow
```
cd manual-tests
go run main.go

In your browser:
http://localhost:6334/api/passkeys 
```