# dev-stack
> Local development framework for streamlined automation, documentation, and service management.

## Overview
**dev-stack** provides a modular, maintainable, and automated environment for local development. All commands and services are defined in YAML manifests, with documentation auto-generated for accuracy and ease of use.

## Quick Start

1. **Clone the repo:**  
   `git clone <repo-url>`
2. **Install Python & PyYAML:**  
   See [Contributing Guide](docs/contributing.md#setup).
3. **Run the doc generation script:**  
   `python scripts/generate_docs.py`
4. **Explore docs:**  
   - [Reference](docs/reference.md)
   - [Services](docs/services.md)
   - [Setup Guide](docs/setup.md)

## Documentation Structure

- **Reference:** Auto-generated CLI commands ([docs/reference.md](docs/reference.md))
- **Services:** Supported services/options ([docs/services.md](docs/services.md))
- **Setup:** Environment & dependencies ([docs/setup.md](docs/setup.md))
- **Troubleshooting:** Common issues ([docs/troubleshooting.md](docs/troubleshooting.md))
- **Contributing:** Workflow & standards ([docs/contributing.md](docs/contributing.md))

## Contributor Workflow

- Use [pyenv](https://github.com/pyenv/pyenv) to manage your Python versions:
  1. Install pyenv:  
     `curl https://pyenv.run | bash`
  2. Install the required Python version (see `.python-version` or contributing guide):  
     `pyenv install <version>`
  3. Set the local Python version:  
     `pyenv local <version>`
  4. (Optional) Create and activate a virtualenv:  
     `pyenv virtualenv <version> dev-stack-env`  
     `pyenv activate dev-stack-env`
- Install dependencies:  
  `pip install -r requirements.txt`
- Update `scripts/commands.yaml` and/or `services/services.yaml` for changes.
- Run `python scripts/generate_docs.py` to update docs.
- Commit both the manifest and generated docs.
- Follow [Contributing Guide](docs/contributing.md) and PR template.

## Branding & Best Practices

- Repo name: **dev-stack** (consistent across docs and code)
- Documentation is DRY, modular, and references centralized sources.
- Manual edits to auto-generated docs are discouraged.

## See Also

- [Reference Docs](docs/reference.md)
- [Services Docs](docs/services.md)
- [Configuration Guide](docs/configuration.md)
- [Usage Guide](docs/usage.md)
- [Integration Guide](docs/integration.md)
- [Contributing Guide](docs/contributing.md)
- [Setup Guide](docs/setup.md)
- [Troubleshooting Guide](docs/troubleshooting.md)

---

> **Tip:** For troubleshooting, advanced configuration, and integration patterns, see the guides above.  
> For a full list of commands and service options, see [Reference Docs](docs/reference.md) and [Services Docs](docs/services.md).