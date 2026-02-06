# lib-common

Shared Go library for Kilat Pet Runner microservices platform.

**Organization:** `github.com/Kilat-Pet-Delivery`

## What it Provides

- **domain/** — Base entities, aggregate root, value objects, money type, repository interfaces, domain errors
- **auth/** — JWT token management (access and refresh tokens)
- **middleware/** — Auth, CORS, logger, rate limiter, recovery, request ID, security headers
- **kafka/** — Producer, consumer, CloudEvent envelope support
- **database/** — PostgreSQL with PostGIS via GORM
- **config/** — Viper-based configuration loader
- **health/** — Health check and readiness endpoints
- **response/** — Standard HTTP response helpers
- **logger/** — Zap logger factory
- **resilience/** — Retry with exponential backoff

## Installation

```bash
go get github.com/Kilat-Pet-Delivery/lib-common
```

## Requirements

- Go 1.24 or higher

## License

Proprietary - Kilat Pet Delivery
