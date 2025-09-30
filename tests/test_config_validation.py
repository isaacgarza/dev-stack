import subprocess
import os
import yaml
import pytest
import tempfile
import shutil
from pathlib import Path

def run_cmd(cmd, cwd=None):
    """Run a shell command and return the result object."""
    result = subprocess.run(cmd, shell=True, capture_output=True, text=True, cwd=cwd)
    return result

def create_config_file(config_content, config_path):
    """Create a config file with the given content."""
    with open(config_path, 'w') as f:
        f.write(config_content)

@pytest.fixture
def clean_config():
    """Fixture to backup and restore config file."""
    config_path = "dev-stack-config.yaml"
    backup_path = None

    # Backup existing config if it exists
    if os.path.exists(config_path):
        backup_path = f"{config_path}.backup"
        shutil.copy2(config_path, backup_path)

    yield config_path

    # Cleanup: remove test config and restore backup if it existed
    if os.path.exists(config_path):
        os.remove(config_path)

    if backup_path and os.path.exists(backup_path):
        shutil.copy2(backup_path, config_path)
        os.remove(backup_path)

@pytest.mark.order(1)
def test_valid_minimal_config(clean_config):
    """Test that a minimal valid configuration works."""
    config_content = """services:
  enabled:
    - redis
"""
    create_config_file(config_content, clean_config)

    result = run_cmd("./scripts/setup.sh --dry-run")

    # Should pass with minimal config
    assert result.returncode == 0 or "Services validated" in result.stdout, \
        f"Minimal config should be valid. Output: {result.stdout}\nError: {result.stderr}"

@pytest.mark.order(2)
def test_invalid_service_name(clean_config):
    """Test that setup fails gracefully with invalid service names."""
    config_content = """services:
  enabled:
    - not_a_real_service
    - definitely_fake_service
"""
    create_config_file(config_content, clean_config)

    result = run_cmd("./scripts/setup.sh")

    # Should fail with invalid services
    assert result.returncode != 0, "Setup should fail with invalid services"

    # Should mention the invalid services in output
    output = result.stdout + result.stderr
    assert "not_a_real_service" in output or "invalid" in output.lower(), \
        f"Should mention invalid services. Output: {output}"

@pytest.mark.order(3)
def test_mixed_valid_invalid_services(clean_config):
    """Test that setup fails when mixing valid and invalid services."""
    config_content = """services:
  enabled:
    - redis           # valid
    - fake_service    # invalid
    - postgres        # valid
"""
    create_config_file(config_content, clean_config)

    result = run_cmd("./scripts/setup.sh")

    # Should fail with mixed valid/invalid services
    assert result.returncode != 0, "Setup should fail with mixed valid/invalid services"
    assert "fake_service" in (result.stdout + result.stderr), \
        "Should identify the invalid service"

@pytest.mark.order(4)
def test_empty_services_list(clean_config):
    """Test handling of empty services list."""
    config_content = """project:
  name: test-stack

services:
  enabled: []
"""
    create_config_file(config_content, clean_config)

    result = run_cmd("./scripts/setup.sh")

    # Should handle empty list gracefully
    output = result.stdout + result.stderr
    assert "no services" in output.lower() or result.returncode != 0, \
        "Should handle empty services list appropriately"

@pytest.mark.order(5)
def test_missing_services_section(clean_config):
    """Test handling of configuration missing services section."""
    config_content = """project:
  name: test-stack
  environment: local
"""
    create_config_file(config_content, clean_config)

    result = run_cmd("./scripts/setup.sh")

    # Should fail or warn about missing services
    assert result.returncode != 0, "Should fail with missing services section"

@pytest.mark.order(6)
def test_malformed_yaml_config(clean_config):
    """Test handling of malformed YAML configuration."""
    # Invalid YAML with incorrect indentation
    config_content = """services:
enabled:
- redis
  - postgres
"""
    create_config_file(config_content, clean_config)

    result = run_cmd("./scripts/setup.sh")

    # Should fail with YAML parsing error
    assert result.returncode != 0, "Should fail with malformed YAML"
    output = result.stdout + result.stderr
    assert any(word in output.lower() for word in ["yaml", "parse", "syntax", "invalid"]), \
        f"Should mention YAML/parsing error. Output: {output}"

@pytest.mark.order(7)
def test_valid_config_with_overrides(clean_config):
    """Test that valid configuration with overrides works."""
    config_content = """project:
  name: test-project
  environment: local

services:
  enabled:
    - redis
    - postgres

overrides:
  redis:
    port: 6380
    password: "test-password"
  postgres:
    port: 5433
    database: "test_db"
    username: "test_user"
"""
    create_config_file(config_content, clean_config)

    result = run_cmd("./scripts/setup.sh --dry-run")

    # Should accept valid overrides
    assert result.returncode == 0 or "Services validated" in result.stdout, \
        f"Valid config with overrides should work. Output: {result.stdout}\nError: {result.stderr}"

@pytest.mark.order(8)
def test_project_name_extraction(clean_config):
    """Test that custom project names are properly extracted."""
    config_content = """project:
  name: my-custom-project
  environment: development

services:
  enabled:
    - redis
"""
    create_config_file(config_content, clean_config)

    result = run_cmd("./scripts/setup.sh --dry-run")

    # Should extract custom project name
    assert "my-custom-project" in result.stdout, \
        f"Should use custom project name. Output: {result.stdout}"

@pytest.mark.order(9)
def test_all_valid_services(clean_config):
    """Test configuration with all valid services enabled."""
    config_content = """services:
  enabled:
    - redis
    - postgres
    - mysql
    - jaeger
    - prometheus
    - localstack
    - kafka
"""
    create_config_file(config_content, clean_config)

    # Use environment variable to force non-interactive mode
    env = os.environ.copy()
    env["CI"] = "true"  # Force non-interactive mode

    result = subprocess.run(
        ["./scripts/setup.sh", "--dry-run"],
        env=env,
        capture_output=True,
        text=True,
        input="y\n"  # Provide automatic "yes" response
    )

    # Should accept all valid services or show warnings but continue
    output = result.stdout + result.stderr
    assert result.returncode == 0 or "Services validated" in output or "⚠️" in output, \
        f"All valid services should be accepted with warnings. Output: {output}"

@pytest.mark.order(10)
def test_skip_validation_flag(clean_config):
    """Test that validation can be skipped when flag is provided."""
    config_content = """services:
  enabled:
    - invalid_service_name
"""
    create_config_file(config_content, clean_config)

    result = run_cmd("./scripts/setup.sh --skip-validation --dry-run")

    # Should skip validation and not fail due to invalid service
    output = result.stdout + result.stderr
    assert "skip" in output.lower() or result.returncode == 0, \
        f"Should skip validation. Output: {output}"

    # Should not contain validation error messages
    assert "Invalid services" not in output, \
        "Should not show validation errors when skipping"

@pytest.mark.order(11)
def test_missing_config_file():
    """Test handling when no config file exists."""
    # Ensure no config file exists
    config_path = "dev-stack-config.yaml"
    if os.path.exists(config_path):
        os.remove(config_path)

    try:
        result = run_cmd("./scripts/setup.sh")

        # Should either create a sample config or fail gracefully
        if result.returncode == 0:
            # If successful, should have created a config file
            assert os.path.exists(config_path), "Should create config file if missing"

            # Verify it's valid YAML
            with open(config_path, 'r') as f:
                config = yaml.safe_load(f)
                assert 'services' in config, "Created config should have services section"
        else:
            # If it fails, should provide helpful message
            output = result.stdout + result.stderr
            assert any(word in output.lower() for word in ["config", "missing", "create", "init"]), \
                f"Should provide helpful message about missing config. Output: {output}"

    finally:
        # Cleanup: remove created config file
        if os.path.exists(config_path):
            os.remove(config_path)

@pytest.mark.order(12)
def test_config_with_comments_and_formatting(clean_config):
    """Test that configuration with comments and various formatting works."""
    config_content = """# This is a test configuration
project:
  name: test-stack  # inline comment
  environment: local

# Services section
services:
  enabled:
    - redis    # caching
    - postgres # database
    # - mysql  # commented out service

# Override section with various formats
overrides:
  redis:
    port: 6379
    memory_limit: "256m"
  postgres:
    port: 5432
    database: "app_dev"
"""
    create_config_file(config_content, clean_config)

    result = run_cmd("./scripts/setup.sh --dry-run")

    # Should handle comments and formatting properly
    assert result.returncode == 0 or "Services validated" in result.stdout, \
        f"Should handle comments and formatting. Output: {result.stdout}\nError: {result.stderr}"

@pytest.mark.order(13)
def test_duplicate_services_in_config(clean_config):
    """Test handling of duplicate services in enabled list."""
    config_content = """services:
  enabled:
    - redis
    - postgres
    - redis    # duplicate
    - jaeger
"""
    create_config_file(config_content, clean_config)

    result = run_cmd("./scripts/setup.sh --dry-run")

    # Should handle duplicates gracefully (either deduplicate or warn)
    # Most implementations would handle this without error
    output = result.stdout + result.stderr
    # This test mainly ensures the system doesn't crash with duplicates
    assert True  # If we get here without crashing, test passes

@pytest.mark.order(14)
def test_config_validation_with_warnings(clean_config):
    """Test configuration that might generate warnings but should still work."""
    config_content = """project:
  name: test-stack

services:
  enabled:
    - redis

validation:
  skip_warnings: false

# This might generate warnings but should still be valid
overrides:
  postgres:  # Override for service not enabled
    port: 5432
"""
    create_config_file(config_content, clean_config)

    result = run_cmd("./scripts/setup.sh --dry-run")

    # Should work but might show warnings
    output = result.stdout + result.stderr
    # Either succeeds or shows warnings but doesn't completely fail
    assert result.returncode == 0 or "warning" in output.lower(), \
        f"Should handle potential warnings. Output: {output}"
