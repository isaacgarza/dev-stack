---
title: "Contributing"
description: "Guide for contributing to the dev-stack project"
weight: 50
---

# Contributing to dev-stack

Thank you for your interest in contributing to dev-stack! This guide will help you get started with contributing to our Go-based development stack management tool.

## Code of Conduct

By participating in this project, you agree to abide by our [Code of Conduct](https://github.com/isaacgarza/dev-stack/blob/main/CODE_OF_CONDUCT.md). Please read it before contributing.

## Getting Started

### Prerequisites

- **Go 1.21+** - Required for building the project
- **Docker** - Required for testing service integrations
- **Git** - For version control
- **Make** - For build automation

### Development Setup

1. **Fork and clone the repository:**
   ```bash
   git clone https://github.com/your-username/dev-stack.git
   cd dev-stack
   ```

2. **Install dependencies:**
   ```bash
   go mod download
   ```

3. **Build the project:**
   ```bash
   task build
   ```

4. **Run tests:**
   ```bash
   task test
   ```

5. **Start development:**
   ```bash
   # Run with hot reload during development
   task dev
   ```

## Project Structure

Understanding the project structure will help you navigate the codebase:

```
dev-stack/
├── cmd/
│   └── dev-stack/           # CLI entry point
├── internal/
│   ├── cli/                 # CLI commands and interfaces
│   ├── core/
│   │   ├── docker/          # Docker integration
│   │   └── services/        # Service management
│   └── pkg/
│       ├── config/          # Configuration management
│       ├── docs/            # Documentation generation
│       ├── logger/          # Logging utilities
│       ├── types/           # Type definitions
│       └── utils/           # Utility functions
├── services/                # Service definitions (YAML)
├── scripts/                 # Build and automation scripts
├── docs/                    # Legacy documentation
├── content/                 # Hugo documentation content
└── themes/                  # Hugo themes
```

## Ways to Contribute

### 1. Reporting Bugs

Before reporting a bug:
- Check the [existing issues](https://github.com/isaacgarza/dev-stack/issues)
- Ensure you're using the latest version
- Gather relevant information about your environment

When reporting a bug, include:

```markdown
**Environment:**
- OS: [e.g., Ubuntu 22.04, macOS 13.0]
- dev-stack version: [run `dev-stack version`]
- Docker version: [run `docker version`]
- Go version: [run `go version`]

**Steps to Reproduce:**
1. Run command `dev-stack ...`
2. Observe error

**Expected Behavior:**
[What you expected to happen]

**Actual Behavior:**
[What actually happened]

**Additional Context:**
[Any additional information, logs, screenshots]
```

### 2. Suggesting Features

Feature requests are welcome! Before suggesting:
- Check if the feature already exists
- Review [existing feature requests](https://github.com/isaacgarza/dev-stack/issues?q=is%3Aissue+is%3Aopen+label%3Aenhancement)
- Consider if it fits the project's goals

Use the feature request template and include:
- Clear description of the feature
- Use cases and examples
- Potential implementation ideas
- Any relevant mockups or designs

### 3. Contributing Code

#### Getting Your Environment Ready

1. **Set up pre-commit hooks:**
   ```bash
   # Install pre-commit (if not already installed)
   pip install pre-commit
   # or
   brew install pre-commit

   # Install hooks
   pre-commit install
   ```

2. **Configure your editor:**
   - Install Go extension/plugin
   - Enable format on save
   - Configure linter integration

#### Making Changes

1. **Create a feature branch:**
   ```bash
   git checkout -b feature/your-feature-name
   # or
   git checkout -b bugfix/issue-number
   ```

2. **Make your changes:**
   - Write clean, readable code
   - Follow Go conventions and idioms
   - Add/update tests for your changes
   - Update documentation if needed

3. **Test your changes:**
   ```bash
   # Run all tests
   task test

   # Run linting
   task lint

   # Run security checks
   task vet

   # Test specific functionality
   go test ./internal/...
   ```

4. **Commit your changes:**
   ```bash
   # Use conventional commit format
   git commit -m "feat: add new service template system"
   git commit -m "fix: resolve docker network configuration issue"
   git commit -m "docs: update CLI reference documentation"
   ```

5. **Push and create a pull request:**
   ```bash
   git push origin feature/your-feature-name
   ```

### 4. Adding New Services

Adding a new service involves creating a service definition file:

1. **Create service definition:**
   ```bash
   # Create a new service file
   touch services/newservice.yaml
   ```

2. **Define the service:**
   ```yaml
   name: newservice
   description: Description of the new service
   category: database  # or cache, queue, monitoring, etc.
   version: "1.0.0"

   container:
     image: newservice/newservice:latest
     ports:
       - "8080:8080"
     environment:
       - SERVICE_ENV=development
     volumes:
       - newservice-data:/data
     
   health_check:
     endpoint: /health
     port: 8080
     interval: 30s
     timeout: 10s
     retries: 3

   configuration:
     port:
       description: Service port
       default: 8080
       type: integer
       validation: "1024-65535"
     
     database:
       description: Database name
       default: "{{.ProjectName}}"
       type: string
       required: true

   documentation:
     description: |
       Comprehensive description of what this service does,
       how to configure it, and how to use it.
     
     examples:
       - description: Basic setup
         command: dev-stack service add newservice
       
       - description: Custom configuration
         command: dev-stack service add newservice --port 9090 --database mydb
     
     links:
       - title: Official Documentation
         url: https://newservice.io/docs
       - title: Docker Hub
         url: https://hub.docker.com/r/newservice/newservice
   ```

3. **Add tests:**
   ```go
   // internal/core/services/newservice_test.go
   func TestNewServiceConfiguration(t *testing.T) {
       // Test service configuration
   }

   func TestNewServiceHealthCheck(t *testing.T) {
       // Test health check functionality
   }
   ```

4. **Update documentation:**
   - Add service to `content/services.md`
   - Include usage examples
   - Document configuration options

### 5. Improving Documentation

Documentation improvements are always welcome:

- **Fix typos and grammar**
- **Add examples and use cases**
- **Improve clarity and organization**
- **Add missing documentation**

Documentation is built with Hugo and located in the `content/` directory.

## Development Guidelines

### Code Style

- **Follow Go conventions:** Use `gofmt`, `golint`, and `go vet`
- **Write clear variable names:** Prefer descriptive names over short ones
- **Add comments:** Document exported functions and complex logic
- **Keep functions small:** Aim for single responsibility
- **Handle errors properly:** Don't ignore errors, handle them appropriately

### Testing

- **Write tests for new functionality**
- **Maintain or improve test coverage**
- **Include integration tests for services**
- **Test error conditions and edge cases**

```bash
# Run tests with coverage
task test

# Run integration tests
task test-integration

# Run specific test
go test -run TestSpecificFunction ./internal/...
```

### Performance

- **Profile performance-critical code**
- **Avoid unnecessary allocations**
- **Use efficient data structures**
- **Consider memory usage in long-running operations**

### Security

- **Validate all inputs**
- **Use secure defaults**
- **Avoid hardcoded secrets**
- **Follow security best practices**

```bash
# Run security checks
task vet

# Check for vulnerabilities
task lint
```

## Pull Request Process

### Before Submitting

1. **Ensure tests pass:**
   ```bash
   task test
   ```

2. **Check code quality:**
   ```bash
   task lint
   task vet
   ```

3. **Update documentation:**
   - Update relevant docs in `content/`
   - Add/update CLI help text
   - Include examples

4. **Write clear commit messages:**
   ```
   feat: add PostgreSQL service template
   
   - Add PostgreSQL service definition with health checks
   - Include configuration options for port, database, and credentials
   - Add comprehensive documentation and examples
   - Include integration tests for service lifecycle
   
   Closes #123
   ```

### Pull Request Template

When creating a pull request, include:

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix (non-breaking change that fixes an issue)
- [ ] New feature (non-breaking change that adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update

## Testing
- [ ] I have added tests that prove my fix is effective or that my feature works
- [ ] New and existing unit tests pass locally with my changes
- [ ] I have tested this change manually

## Documentation
- [ ] I have updated the documentation accordingly
- [ ] I have added/updated examples where appropriate

## Checklist
- [ ] My code follows the style guidelines of this project
- [ ] I have performed a self-review of my own code
- [ ] I have commented my code, particularly in hard-to-understand areas
- [ ] My changes generate no new warnings
```

### Review Process

1. **Automated checks** must pass (tests, linting, security)
2. **Manual review** by maintainers
3. **Address feedback** and update PR as needed
4. **Final approval** and merge

## Release Process

Releases follow semantic versioning (SemVer):
- **Patch** (1.0.1): Bug fixes
- **Minor** (1.1.0): New features, backward compatible
- **Major** (2.0.0): Breaking changes

Releases are automated using:
- **Release Please** for changelog generation
- **GitHub Actions** for building and publishing
- **Conventional Commits** for automatic versioning

## Community and Support

### Getting Help

- **GitHub Discussions:** For questions and community support
- **GitHub Issues:** For bug reports and feature requests
- **Documentation:** Check existing docs first

### Communication

- **Be respectful and constructive**
- **Ask questions if you're unsure**
- **Provide context and examples**
- **Help others when you can**

## Recognition

We value all contributions! Contributors will be:
- **Listed in release notes** for significant contributions
- **Mentioned in documentation** for new features
- **Added to contributors list** in the repository

## Development Tips

### Useful Commands

```bash
# Development workflow
task dev          # Build and run with hot reload
task build        # Build binary
task test         # Run tests
task lint         # Run linting
task clean        # Clean build artifacts

# Documentation
task docs         # Generate documentation
task validate-docs # Validate documentation

# Release
task build-all    # Create release builds
```

### Debugging

```bash
# Run with debug logging
dev-stack --debug [command]

# Enable verbose output
dev-stack --verbose [command]

# Use Go debugger
dlv debug ./cmd/dev-stack
```

### IDE Setup

**VS Code:**
- Install Go extension
- Configure format on save
- Enable Go linting

**GoLand:**
- Import project as Go module
- Configure Go tools integration
- Set up run configurations

## Questions?

If you have questions about contributing:

1. **Check the [FAQ]({{< ref "/usage#troubleshooting" >}})**
2. **Search [existing issues](https://github.com/isaacgarza/dev-stack/issues)**
3. **Start a [discussion](https://github.com/isaacgarza/dev-stack/discussions)**
4. **Ask in your pull request**

Thank you for contributing to dev-stack! 🚀

---

> **Remember**: No contribution is too small. Whether it's fixing a typo, reporting a bug, or adding a major feature, every contribution helps make dev-stack better for everyone.