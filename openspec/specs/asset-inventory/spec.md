# asset-inventory

## Requirements

### Requirement: List Assets
The emulator SHALL support listing all emulated resources across services.

#### Scenario: List all assets
- **WHEN** GET /v1/assets
- **THEN** all emulated resources (buckets, topics, secrets, etc.) are returned

### Requirement: Export Assets
The emulator SHALL support exporting assets filtered by type.

#### Scenario: Export by asset type
- **WHEN** POST /v1/assets with assetType filter
- **THEN** only matching assets are returned
