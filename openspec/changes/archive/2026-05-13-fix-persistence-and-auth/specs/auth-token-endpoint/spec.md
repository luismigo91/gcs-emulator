## ADDED Requirements

### Requirement: OAuth2 Token Endpoint
The emulator SHALL expose a `POST /oauth2/v4/token` endpoint that returns a valid-looking OAuth2 token response for any input credentials, enabling SDK credential flows without real authentication.

#### Scenario: Token endpoint returns mock token
- **WHEN** a POST request is sent to `/oauth2/v4/token` with form body `grant_type=urn:ietf:params:oauth:grant-type:jwt-bearer&assertion=any-jwt`
- **THEN** the response SHALL be HTTP 200 with JSON body containing `access_token`, `token_type: "Bearer"`, and `expires_in` fields

#### Scenario: Token endpoint accepts any grant type
- **WHEN** a POST request is sent with `grant_type=authorization_code`
- **THEN** the response SHALL return HTTP 200 with the same mock token format

#### Scenario: Token endpoint handles JSON content type
- **WHEN** a POST request is sent with `Content-Type: application/json` and body containing `{"grant_type": "..."}`
- **THEN** the response SHALL return HTTP 200 with the mock token format

#### Scenario: Token endpoint returns valid expires_in
- **WHEN** the token endpoint is called
- **THEN** the `expires_in` field SHALL be a positive integer (e.g., 3600)
