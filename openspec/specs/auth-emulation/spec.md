# auth-emulation

## MODIFIED Requirements

### Requirement: OAuth2 token endpoint
The emulator SHALL provide an OAuth2 token endpoint at /oauth2/v4/token that returns valid-looking tokens for SDK compatibility.

#### Scenario: Token endpoint returns mock token
- **WHEN** POST /oauth2/v4/token with any credentials
- **THEN** HTTP 200 with access_token, token_type "Bearer", expires_in 3600

#### Scenario: Token endpoint handles JSON and form bodies
- **WHEN** POST with Content-Type application/json or application/x-www-form-urlencoded
- **THEN** the response format is identical

#### Scenario: Signed URL passthrough
- **WHEN** a signed URL is presented with any signature
- **THEN** the emulator processes the request without validation
