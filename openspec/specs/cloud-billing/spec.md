# cloud-billing

## Requirements

### Requirement: Budget CRUD
The emulator SHALL support creating, retrieving, listing, and deleting budgets.

#### Scenario: Create budget
- **WHEN** POST /v1/billingAccounts/{id}/budgets with amount and threshold rules
- **THEN** budget is created and returned
