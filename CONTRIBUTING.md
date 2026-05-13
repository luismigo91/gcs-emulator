# Contributing Guidelines

Thank you for your interest in contributing to GCS Emulator!

## Getting Started

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/my-feature`
3. Make your changes
4. Run tests: `make test`
5. Run linter: `make lint`
6. Commit your changes: `git commit -m 'Add my feature'`
7. Push to the branch: `git push origin feature/my-feature`
8. Submit a pull request

## Code Style

- Follow [Effective Go](https://go.dev/doc/effective_go)
- Use `gofmt` to format your code
- Run `make lint` before submitting
- Write tests for new functionality

## Commit Messages

We follow [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` New feature
- `fix:` Bug fix
- `docs:` Documentation changes
- `test:` Adding or updating tests
- `refactor:` Code refactoring
- `chore:` Maintenance tasks

## Testing

- Unit tests should be in the same package as the code being tested
- Integration tests go in the `tests/` directory
- Run `make test` to run all tests with race detection

## Pull Requests

- Keep PRs focused and minimal
- Include tests for new functionality
- Update documentation if needed
- Reference any related issues

## Questions?

Feel free to open an issue for any questions or concerns.
