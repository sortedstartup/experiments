# Environment Variables Setup

To use the JWT authentication for the paid content pages, you need to configure the following environment variables:

## Required Variables

Create a `.env.local` file in the root of the project with the following variables:

```bash
# JWT Configuration
JWT_SECRET=your-secret-key-change-this-in-production
JWT_ISSUER=sortedauth
```

### Description:

- **JWT_SECRET**: The secret key used to verify JWT tokens. This should be a strong, random string in production.
- **JWT_ISSUER**: The issuer claim to verify in JWT tokens. Must match the issuer that signs the tokens.

## Installation

Install dependencies using pnpm:

```bash
pnpm install
```

## Running the Application

```bash
pnpm dev
```

## How It Works

1. **Middleware** (`middleware.ts`): Intercepts requests to `/paid-content-1` and `/paid-content-2`
2. **JWT Verification**: Validates the JWT token from the `Authorization` header using `JWT_SECRET` and `JWT_ISSUER`
3. **Access Control**: Checks if the user has paid and has access to the specific product
4. **Server-Side Pages**: Both pages are server-side rendered (no client-side code)
5. **No Cookies**: Authentication is handled via Authorization header only

## JWT Token Structure

The JWT token should contain:
- `iss` (issuer): Must match `JWT_ISSUER`
- `paid`: Boolean indicating if user has paid
- `products`: Array of products the user has access to
- `email`: User's email (optional, for display)
- `user_id` or `sub`: User identifier

Example payload:
```json
{
  "iss": "sortedauth",
  "sub": "user123",
  "email": "user@example.com",
  "paid": true,
  "products": [
    { "product_id": "1", "name": "Product 1" },
    { "product_id": "2", "name": "Product 2" }
  ]
}
```

## Testing

To test the protected pages:
1. Obtain a valid JWT token from your auth service
2. Make a request with the `Authorization: Bearer <token>` header
3. Access `/paid-content-1` or `/paid-content-2`

Without a valid token, you'll receive a 401 Unauthorized response.

