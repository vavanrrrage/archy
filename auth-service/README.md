# Auth Service
Elysia + Better-auth + Postgres

This service currently support email and password sign-up/sign-in

## Available Endpoints

- `POST /api/auth/sign-up` - Register a new user
- `POST /api/auth/sign-in` - Sign in with email/password
- `GET /api/auth/token` - Get JWT token (requires authentication)
- `GET /api/auth/jwks` - JSON Web Key Set for JWT verification (public)

## JWT Configuration

The service uses EdDSA (Ed25519) for JWT signing. Public keys are available via the `/api/auth/jwks` endpoint for token verification in other services.

See `../JWT_VERIFICATION.md` for details on verifying JWT tokens in other services.