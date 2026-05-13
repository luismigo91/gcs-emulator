# artifact-registry

## Requirements

### Requirement: Repository CRUD
The emulator SHALL support creating, retrieving, listing, and deleting Docker repositories.

#### Scenario: Create Docker repo
- **WHEN** POST /v1/projects/{p}/locations/{l}/repositories?repositoryId=my-repo with format DOCKER
- **THEN** repository is created

### Requirement: Docker Image Listing
The emulator SHALL support listing Docker images in a repository.

#### Scenario: List images
- **WHEN** GET .../repositories/{r}/dockerImages
- **THEN** images with tags are returned
