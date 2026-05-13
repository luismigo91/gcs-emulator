# cloud-kms

## Requirements

### Requirement: Key Ring CRUD
The emulator SHALL support creating, retrieving, and listing key rings.

#### Scenario: Create key ring
- **WHEN** POST /v1/projects/{p}/locations/{l}/keyRings?keyRingId=my-kr
- **THEN** key ring is created

### Requirement: Crypto Key CRUD
The emulator SHALL support creating, retrieving, listing, and updating crypto keys.

#### Scenario: Create crypto key
- **WHEN** POST .../keyRings/{kr}/cryptoKeys?cryptoKeyId=my-key
- **THEN** crypto key is created with ENCRYPT_DECRYPT purpose

### Requirement: Version Management
The emulator SHALL support creating, retrieving, listing, destroying, and restoring crypto key versions.

#### Scenario: Create and destroy version
- **WHEN** POST .../cryptoKeyVersions
- **THEN** version ENCRYPTED; :destroy changes state to DESTROYED

### Requirement: Encrypt/Decrypt
The emulator SHALL support encrypting and decrypting data with crypto keys.

#### Scenario: Encrypt and decrypt
- **WHEN** encrypting plaintext via key:encrypt
- **THEN** ciphertext is returned; decrypting returns original plaintext
