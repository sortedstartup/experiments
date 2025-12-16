# Auth.js REST API Documentation

This document provides a high-level overview of the internal REST API endpoints for Auth.js (formerly NextAuth.js), based on the core action files.

## 1. Sign Out (`/signout`)

**Source:** `packages/core/src/lib/actions/signout.ts`

### Description
Handles the user sign-out process. It terminates the session by clearing cookies and, if applicable, removing the session from the database.

### Input
- **Cookies**: Current request cookies.
- **SessionStore**: Helper to manage session cookies.
- **Options**: Internal configuration options (contains adapter, jwt, events, etc.).

### Return Value
- **ResponseInternal**:
  - `redirect`: URL to redirect to after sign-out (default is the configured callback URL).
  - `cookies`: List of cookies to set (specifically, cookies to delete/expire the session).

### High-Level Logic
1.  **Retrieve Session Token**: Extracts the session token from the cookies.
2.  **Strategy Check**:
    -   **JWT Strategy**: Decodes the token to verify it. Emits the `signOut` event with the token.
    -   **Database Strategy**: Calls the adapter's `deleteSession` method to remove the session entry from the database. Emits the `signOut` event with the session object.
3.  **Clear Cookies**: Generates "expired" cookies to clear the session token from the client's browser.
4.  **Redirect**: Returns the redirect URL and the cleared cookies.

---

## 2. Session (`/session`)

**Source:** `packages/core/src/lib/actions/session.ts`

### Description
Retrieves the current session data. This is used by the client to check if a user is logged in and to get user details.

### Input
- **Options**: Internal configuration options.
- **SessionStore**: Helper to manage session cookies.
- **Cookies**: Current request cookies.
- **isUpdate**: (Optional) Boolean indicating if the session is being updated.
- **newSession**: (Optional) New session data to merge/update.

### Return Value
- **ResponseInternal**:
  - `body`: The Session object (e.g., `{ user: { ... }, expires: "..." }`) or `null` if not authenticated.
  - `cookies`: Updated cookies (e.g., for rolling sessions).

### High-Level Logic
1.  **Token Retrieval**: Checks for a session token in the cookies. If missing, returns `null`.
2.  **Strategy Check**:
    -   **JWT Strategy**:
        -   Decodes and verifies the JWT.
        -   Calls the `jwt` callback to allow customization.
        -   Calls the `session` callback to format the final session object.
        -   **Rolling Session**: If enabled, re-signs the JWT with a new expiry and updates the cookie.
    -   **Database Strategy**:
        -   Retrieves the session and user from the database via the adapter.
        -   **Expiry Check**: If the session is expired, deletes it and returns `null`.
        -   **Update**: If the session is active but close to expiry (based on `updateAge`), updates the expiry time in the database.
        -   Calls the `session` callback.
3.  **Response**: Returns the session object and any updated cookies.

---

## 3. Sign In (`/signin`)

**Source:** `packages/core/src/lib/actions/signin/index.ts`
**Related:** `authorization-url.ts`, `send-token.ts`

### Description
Initiates the sign-in flow. This endpoint handles various providers, including OAuth/OIDC and Email (Magic Links).

### Input
- **Request**: The incoming request object (contains query params, body).
- **Cookies**: Current request cookies.
- **Options**: Internal configuration options (includes provider details).

### Return Value
- **ResponseInternal**:
  - `redirect`: The URL to redirect the user to (e.g., OAuth provider's login page or a verification page).
  - `cookies`: Cookies to set (e.g., OAuth state, nonce, PKCE code verifier).

### High-Level Logic
The logic branches based on the `provider.type`:

#### A. OAuth / OIDC
1.  **Authorization URL**: Calls `getAuthorizationUrl`.
2.  **Discovery**: If needed, fetches OIDC discovery document to find the authorization endpoint.
3.  **Parameter Construction**: Builds the OAuth URL with `client_id`, `redirect_uri`, `scope`, `response_type`, etc.
4.  **Security Checks**:
    -   **State**: Generates a random state string and stores it in a cookie to prevent CSRF.
    -   **PKCE**: Generates a code verifier and challenge (S256) if supported, storing the verifier in a cookie.
    -   **Nonce**: Generates a nonce for OIDC flows.
5.  **Redirect**: Returns the constructed provider URL and the security cookies.

#### B. Email (Magic Link)
1.  **Send Token**: Calls `sendToken`.
2.  **Normalization**: Normalizes the email address.
3.  **User Lookup**: Checks if the user exists via the adapter; creates a default user object if not.
4.  **SignIn Callback**: Calls the `signIn` callback to allow/deny access.
5.  **Token Generation**: Generates a random verification token.
6.  **Persistence**:
    -   Calls `provider.sendVerificationRequest` to email the link to the user.
    -   Calls `adapter.createVerificationToken` to save the hashed token in the database.
7.  **Redirect**: Returns a redirect to the `/verify-request` page.

---

## 4. WebAuthn Options (`/webauthn-options`)

**Source:** `packages/core/src/lib/actions/webauthn-options.ts`

### Description
Generates the necessary options (challenge, public key parameters) for WebAuthn (Passkeys) registration or authentication.

### Input
- **Request**: Incoming request (contains query params like `action`).
- **Options**: Internal configuration options.
- **SessionStore**: Helper to manage session cookies.
- **Cookies**: Current request cookies.

### Return Value
- **ResponseInternal**:
  - `body`: JSON object containing WebAuthn options (e.g., `publicKey` challenge).
  - `cookies`: Cookies to set.
  - `status`: HTTP status code (e.g., 200 or 400).

### High-Level Logic
1.  **Action Extraction**: Reads the `action` query parameter (`register` or `authenticate`).
2.  **User Context**: Checks if a user is already logged in (from session) or tries to identify the user from the request.
3.  **Decision**: Calls `inferWebAuthnOptions` to decide the flow.
4.  **Flow Execution**:
    -   **Authenticate**: Calls `getAuthenticationResponse`. Generates a challenge for the user to sign with their authenticator.
    -   **Register**: Calls `getRegistrationResponse`. Generates a challenge and registration options (e.g., relying party info, user info) to create a new passkey.
5.  **Error Handling**: Returns 400 if the action is invalid or if required user info (like email for registration) is missing.
