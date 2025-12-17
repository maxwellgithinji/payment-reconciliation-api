# Insurance Management Portal - POC

A domain-driven design implementation for an integrated insurance management and payment reconciliation system.

## Project Structure

```
insurance-portal-poc/
├── cmd/
│   └── api/                    # Application entry point
├── internal/
│   ├── domain/                 # Domain layer (business logic)
│   │   ├── customer/
│   │   ├── policy/
│   │   ├── payment/
│   │   ├── reconciliation/
│   │   └── user/
│   ├── application/            # Application services
│   │   ├── commands/
│   │   └── queries/
│   ├── infrastructure/         # External concerns
│   │   ├── persistence/
│   │   ├── payment/
│   │   └── auth/
│   └── interfaces/             # API/UI layer
│       ├── http/
│       └── dto/
├── pkg/                        # Shared utilities
└── migrations/                 # Database migrations
```

## Domain Model

### Bounded Contexts

1. **Customer Management** - Customer profiles and information
2. **Policy Management** - Insurance policies and configurations
3. **Payment Processing** - Payment collection and tracking
4. **Reconciliation** - Automated payment matching
5. **User & Access Management** - Authentication and authorization

## Tech Stack

- **Language**: Go 1.21+
- **Framework**: Chi (HTTP router)
- **Database**: PostgreSQL
- **Payment Gateway**: Paystack/PesaLink integration
- **Authentication**: JWT

## Getting Started

```bash
# Install dependencies
go mod download

# Run migrations
make migrate-up

# Run the application
go run cmd/api/main.go
```

## Key Features (POC Phase)

- [x] User Management (Admin)
- [x] Customer & Policy Management (Accounts Manager)
- [x] Payment Tracking Dashboard
- [x] Automated Payment Reconciliation
- [x] Real-time Status Updates
- [x] Reconciliation Reporting