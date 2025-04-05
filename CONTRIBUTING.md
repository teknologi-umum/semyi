# Contributing to Semyi

Thank you for your interest in contributing to Semyi! This document provides guidelines and instructions for contributing to the project.

## Code of Conduct

By participating in this project, you agree to abide by our Code of Conduct. Please be respectful and considerate of others.

## Development Workflow

1. **Fork the Repository**: Start by forking the repository to your GitHub account.

2. **Create a Branch**: Always create a new branch for your changes. Never commit directly to `master` or `main`.
   ```bash
   git checkout -b feat/your-feature-name
   ```

3. **Make Changes**: Implement your changes following the project's coding standards.

4. **Commit Your Changes**: Follow the Conventional Commits specification:
   ```
   <type>[optional scope]: <description>

   [optional body]

   [optional footer]
   ```
   Where `type` is one of:
   - `fix`: A bug fix
   - `feat`: A new feature
   - `chore`: Changes to the build process or auxiliary tools
   - `docs`: Documentation only changes
   - `style`: Changes that do not affect the meaning of the code
   - `refactor`: A code change that neither fixes a bug nor adds a feature
   - `perf`: A code change that improves performance
   - `test`: Adding missing tests or correcting existing tests

   Example:
   ```bash
   git commit -m "feat(monitor): Add support for custom HTTP headers" -m "This change allows users to specify custom HTTP headers for their monitors, enabling authentication and other custom requirements."
   ```

5. **Push Your Changes**: Push your branch to your fork:
   ```bash
   git push origin feat/your-feature-name
   ```

6. **Create a Pull Request**: Open a pull request from your fork to the main repository. Provide a clear description of your changes and reference any related issues.

## Development Environment Setup

See [DEVELOPMENT.md](./DEVELOPMENT.md) for detailed instructions on setting up your development environment.

## Testing

- Write tests for new features and bug fixes
- Ensure all tests pass before submitting a pull request
- Follow the existing test patterns and conventions

## Documentation

- Update documentation for any changes that affect functionality
- Include clear comments in your code explaining complex logic
- Update the README if necessary

## Review Process

- Pull requests will be reviewed by maintainers
- Be responsive to feedback and requested changes
- Keep your pull request up to date with the main branch

## Questions?

If you have any questions, feel free to open an issue or reach out to the maintainers. 