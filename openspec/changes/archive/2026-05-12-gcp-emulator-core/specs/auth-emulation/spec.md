## ADDED Requirements

### Requirement: Accept any credentials
The emulator SHALL accept any authentication credentials without performing real validation, allowing developers to test without GCP service accounts.

#### Scenario: Arbitrary access key accepted
- **WHEN** a request is made with `Authorization: Bearer any-token-string`
- **THEN** the emulator processes the request without rejecting it for authentication failure

#### Scenario: No credentials required
- **WHEN** a request is made without any Authorization header
- **THEN** the emulator processes the request as an authenticated request

### Requirement: OAuth2 token endpoint emulation
The emulator SHALL provide a minimal OAuth2 token endpoint that returns valid-looking tokens for SDK compatibility.

#### Scenario: Token endpoint returns mock token
- **WHEN** a POST request is sent to the OAuth2 token endpoint with any service account credentials
- **THEN** the emulator returns a JSON response with `access_token`, `token_type: "Bearer"`, `expires_in`, and `scope` fields

### Requirement: Service account JSON acceptance
The emulator SHALL accept service account JSON files as credentials without validating their contents.

#### Scenario: Fake service account JSON
- **WHEN** a client is configured with a service account JSON containing `{"client_email": "test@test.iam.gserviceaccount.com", "private_key": "fake-key"}`
- **THEN** the emulator accepts the credentials and processes requests normally

### Requirement: Application Default Credentials (ADC) compatibility
The emulator SHALL work with GCP Application Default Credentials patterns when `STORAGE_EMULATOR_HOST` is set.

#### Scenario: STORAGE_EMULATOR_HOST environment variable
- **WHEN** the `STORAGE_EMULATOR_HOST` environment variable is set to `http://localhost:9090`
- **THEN** the Go `cloud.google.com/go/storage` client routes requests to the emulator without authentication

#### Scenario: WithoutAuthentication option
- **WHEN** a Go storage client is created with `option.WithoutAuthentication()` and `option.WithEndpoint()`
- **THEN** the client connects to the emulator successfully

### Requirement: Signed URL passthrough
The emulator SHALL accept signed URLs without validating the signature, expiration, or other query parameters.

#### Scenario: Signed URL with fake signature
- **WHEN** a GET request includes `?GoogleAccessId=test&Signature=fake&Expires=9999999999` query parameters
- **THEN** the emulator processes the request and returns the object if it exists

#### Scenario: Signed URL host replacement
- **WHEN** a signed URL originally pointing to `storage.googleapis.com` is rewritten to point to the emulator host
- **THEN** the emulator processes the request correctly
