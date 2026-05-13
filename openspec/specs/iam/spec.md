# iam

## Requirements

### Requirement: Service Account CRUD
The emulator SHALL support creating, retrieving, listing, and deleting service accounts.

#### Scenario: Create service account
- **WHEN** POST /v1/projects/{p}/serviceAccounts with email and display name
- **THEN** service account is created

### Requirement: Sign JWT
The emulator SHALL support signing JWTs for service accounts.

#### Scenario: Sign JWT
- **WHEN** POST /v1/projects/{p}/serviceAccounts/{sa}:signJwt with payload
- **THEN** signed JWT is returned

### Requirement: Generate Access Token
The emulator SHALL support generating access tokens for service accounts.

#### Scenario: Generate token
- **WHEN** POST /v1/projects/{p}/serviceAccounts/{sa}:generateAccessToken
- **THEN** access token with 1-hour expiry is returned
