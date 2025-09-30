PYTHON ?= python3
VENV ?= dev-stack-env
PYTHON_VERSION := $(shell cat .python-version)

.PHONY: setup docs lint clean help

setup:
	@echo "Setting up Python environment with pyenv and virtualenv..."
	@if ! command -v pyenv >/dev/null 2>&1; then \
		echo "Error: pyenv is not installed. See https://github.com/pyenv/pyenv"; \
		exit 1; \
	fi
	@if ! pyenv versions >/dev/null 2>&1; then \
		echo "Error: pyenv-virtualenv is not installed. See https://github.com/pyenv/pyenv-virtualenv"; \
		exit 1; \
	fi
	@if [ ! -f requirements.txt ]; then \
		echo "Error: requirements.txt not found."; \
		exit 1; \
	fi
	pyenv install -s $(PYTHON_VERSION)
	pyenv local $(PYTHON_VERSION)
	pyenv virtualenv $(PYTHON_VERSION) $(VENV) || true
	@echo "Installing dependencies in dev-stack-env virtualenv..."
	$(HOME)/.pyenv/versions/$(VENV)/bin/pip install --upgrade pip
	$(HOME)/.pyenv/versions/$(VENV)/bin/pip install -r requirements.txt

docs:
	@which $(PYTHON) >/dev/null || (echo "$(PYTHON) not found. Run 'make setup' first."; exit 1)
	@$(PYTHON) -c "import yaml" 2>/dev/null || (echo "pyyaml not installed. Run 'make setup' first."; exit 1)
	@if [ ! -f scripts/generate_docs.py ]; then \
		echo "Error: scripts/generate_docs.py not found."; \
		exit 1; \
	fi
	@echo "Generating documentation from YAML manifests..."
	$(PYTHON) scripts/generate_docs.py

lint:
	@echo "Linting Python scripts..."
	@if ! command -v flake8 >/dev/null 2>&1; then \
		echo "Error: flake8 is not installed. Install it with 'pip install flake8'."; \
		exit 1; \
	fi
	@flake8 scripts/ > lint.log 2>&1; \
	if [ $$? -eq 0 ]; then \
		echo "No lint errors found."; \
	else \
		echo "Lint errors found:"; \
		cat lint.log; \
		exit 1; \
	fi

clean:
	@echo "Cleaning up generated files..."
	rm -f docs/reference.md docs/services.md lint.log

help:
	@echo "Available targets:"
	@echo "  setup   - Set up Python environment and install dependencies"
	@echo "  docs    - Generate documentation from YAML manifests"
	@echo "  lint    - Lint Python scripts in scripts/"
	@echo "  clean   - Remove generated documentation and lint files"
	@echo "  help    - Show this help message"
	@echo ""
	@echo "Usage examples:"
	@echo "  make setup"
	@echo "  make docs"
	@echo "  make lint"
	@echo "  make clean"
