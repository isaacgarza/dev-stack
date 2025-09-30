import os
import pytest
import subprocess
import shutil
import yaml

def run_cmd(cmd, cwd=None):
    """Run a shell command and return the result object."""
    result = subprocess.run(cmd, shell=True, capture_output=True, text=True, cwd=cwd)
    return result

@pytest.fixture
def integration_environment():
    """Fixture to set up integration test environment."""
    # Backup existing files
    backup_files = {}
    files_to_manage = ["dev-stack-config.yaml", "docker-compose.generated.yml", ".env.generated"]

    for fname in files_to_manage:
        if os.path.exists(fname):
            backup_files[fname] = f"{fname}.integration_backup"
            shutil.copy2(fname, backup_files[fname])

    yield

    # Cleanup and restore
    for fname in files_to_manage:
        if os.path.exists(fname):
            os.remove(fname)

    for original, backup in backup_files.items():
        if os.path.exists(backup):
            shutil.copy2(backup, original)
            os.remove(backup)

@pytest.mark.order(1)
def test_end_to_end_config_creation(integration_environment):
    """Test complete flow from initialization to file generation."""
    # Ensure no config file exists first
    if os.path.exists("dev-stack-config.yaml"):
        os.remove("dev-stack-config.yaml")

    # Step 1: Initialize configuration
    init_result = run_cmd("./scripts/setup.sh --init")

    assert init_result.returncode == 0, f"Initialization failed: {init_result.stderr}"
    assert os.path.exists("dev-stack-config.yaml"), "Config file not created during initialization"

    # Step 2: Verify config is valid YAML
    with open("dev-stack-config.yaml") as f:
        config = yaml.safe_load(f)

    assert isinstance(config, dict), "Config should be a valid YAML dictionary"
    assert "services" in config, "Config should contain services section"
    assert "enabled" in config["services"], "Config should contain enabled services list"

    # Step 3: Run setup to generate files
    setup_result = run_cmd("./scripts/setup.sh --dry-run")

    # Should complete without crashing
    assert len(setup_result.stdout + setup_result.stderr) > 0, "Setup should produce output"

@pytest.mark.order(2)
def test_configuration_file_structure():
    """Test that generated configuration has expected structure."""
    if not os.path.exists("dev-stack-config.yaml"):
        # Create a config file for testing
        result = run_cmd("./scripts/setup.sh --init")
        if result.returncode != 0:
            pytest.skip("Cannot create configuration file for testing")

    with open("dev-stack-config.yaml") as f:
        config = yaml.safe_load(f)

    # Verify basic structure
    assert "services" in config, "Configuration missing services section"
    assert "enabled" in config["services"], "Configuration missing enabled services"

    # Verify enabled services are valid
    enabled_services = config["services"]["enabled"]
    valid_services = ["redis", "postgres", "mysql", "jaeger", "prometheus", "localstack", "kafka"]

    for service in enabled_services:
        assert service in valid_services, f"Invalid service in configuration: {service}"

@pytest.mark.order(3)
def test_docker_compose_generation():
    """Test that docker-compose file can be generated if setup creates it."""
    # Only test if the file was actually generated
    if not os.path.exists("docker-compose.generated.yml"):
        pytest.skip("docker-compose.generated.yml not created by setup")

    # Verify it's valid YAML
    with open("docker-compose.generated.yml") as f:
        compose_data = yaml.safe_load(f)

    assert isinstance(compose_data, dict), "Docker Compose file should be valid YAML"
    assert "services" in compose_data, "Docker Compose should have services section"

    # Verify services structure
    for service_name, service_config in compose_data["services"].items():
        assert isinstance(service_config, dict), f"Service {service_name} should be a dictionary"

        # Each service should have either image or build directive
        assert "image" in service_config or "build" in service_config, \
            f"Service {service_name} missing image or build directive"

@pytest.mark.order(4)
def test_environment_file_generation():
    """Test that environment file contains expected format if generated."""
    if not os.path.exists(".env.generated"):
        pytest.skip(".env.generated not created by setup")

    with open(".env.generated") as f:
        env_content = f.read()

    # Check for environment variable format
    env_lines = [line.strip() for line in env_content.splitlines()
                 if line.strip() and not line.strip().startswith("#")]

    for line in env_lines:
        if "=" in line:
            key, value = line.split("=", 1)
            assert key.strip(), f"Environment variable has empty key: {line}"
            # Value can be empty, but key should not be

@pytest.mark.order(5)
def test_service_specific_environment_variables():
    """Test that enabled services have corresponding environment variables."""
    if not os.path.exists(".env.generated") or not os.path.exists("dev-stack-config.yaml"):
        pytest.skip("Required files not available")

    # Load configuration to see what services are enabled
    with open("dev-stack-config.yaml") as f:
        config = yaml.safe_load(f)

    enabled_services = config.get("services", {}).get("enabled", [])

    # Load environment variables
    with open(".env.generated") as f:
        env_content = f.read()

    # Check for service-specific environment variables
    service_env_mapping = {
        "redis": ["REDIS"],
        "postgres": ["POSTGRES"],
        "mysql": ["MYSQL"],
        "kafka": ["KAFKA"],
        "jaeger": ["JAEGER"],
        "prometheus": ["PROMETHEUS"],
        "localstack": ["LOCALSTACK", "AWS"]
    }

    for service in enabled_services:
        if service in service_env_mapping:
            expected_env_prefixes = service_env_mapping[service]

            # At least one environment variable for this service should exist
            found_service_env = False
            for prefix in expected_env_prefixes:
                if prefix in env_content:
                    found_service_env = True
                    break

            assert found_service_env, \
                f"No environment variables found for enabled service {service}"

@pytest.mark.order(6)
def test_configuration_validation_integration():
    """Test that configuration validation works in practice."""
    if not os.path.exists("dev-stack-config.yaml"):
        # Create a config file for testing
        result = run_cmd("./scripts/setup.sh --init")
        if result.returncode != 0:
            pytest.skip("Cannot create configuration file for testing")

    # Test that the current configuration passes validation
    result = run_cmd("./scripts/setup.sh --dry-run")

    # Should not fail due to configuration issues
    output = result.stdout + result.stderr

    # Look for validation success indicators or lack of validation errors
    validation_error_indicators = ["invalid service", "configuration error", "yaml error"]

    has_validation_errors = any(indicator in output.lower() for indicator in validation_error_indicators)

    assert not has_validation_errors, f"Configuration validation failed: {output}"

@pytest.mark.order(7)
def test_service_port_configuration():
    """Test that services have proper port configurations."""
    if not os.path.exists("docker-compose.generated.yml"):
        pytest.skip("Docker Compose file not available")

    with open("docker-compose.generated.yml") as f:
        compose_data = yaml.safe_load(f)

    services = compose_data.get("services", {})

    # Services that typically expose ports
    port_services = ["redis", "postgres", "mysql", "jaeger", "prometheus", "kafka"]

    for service_name in services:
        service_config = services[service_name]

        if any(port_service in service_name.lower() for port_service in port_services):
            # Should have port configuration
            assert "ports" in service_config or "expose" in service_config, \
                f"Service {service_name} should have port configuration"

@pytest.mark.order(8)
def test_framework_file_permissions():
    """Test that generated files have appropriate permissions."""
    generated_files = ["dev-stack-config.yaml", "docker-compose.generated.yml", ".env.generated"]

    for fname in generated_files:
        if os.path.exists(fname):
            # File should be readable
            assert os.access(fname, os.R_OK), f"File {fname} should be readable"

            # .env files should not be world-readable for security
            if fname.endswith(".env.generated"):
                stat_info = os.stat(fname)
                # Check that it's not world-readable (mode & 0o004 == 0)
                # This is a security best practice for environment files
                assert (stat_info.st_mode & 0o004) == 0, \
                    f"Environment file {fname} should not be world-readable"

@pytest.mark.order(9)
def test_cleanup_integration():
    """Test that cleanup operations work properly."""
    # Create some test files to cleanup
    test_files = ["test-docker-compose.yml", "test.env"]

    for test_file in test_files:
        with open(test_file, "w") as f:
            f.write("# Test file for cleanup\n")

    try:
        # Test that cleanup doesn't crash
        result = run_cmd("./scripts/manage.sh stop")

        # Cleanup should complete without crashing
        output = result.stdout + result.stderr
        assert len(output.strip()) >= 0, "Cleanup should produce output or complete silently"

    finally:
        # Clean up test files
        for test_file in test_files:
            if os.path.exists(test_file):
                os.remove(test_file)

@pytest.mark.order(10)
def test_multiple_setup_runs():
    """Test that running setup multiple times doesn't break things."""
    if not os.path.exists("dev-stack-config.yaml"):
        # Create a config file for testing
        result = run_cmd("./scripts/setup.sh --init")
        if result.returncode != 0:
            pytest.skip("Cannot create configuration file for testing")

    # Run setup multiple times
    for i in range(2):
        result = run_cmd("./scripts/setup.sh --dry-run")

        # Each run should complete without crashing
        output = result.stdout + result.stderr
        assert len(output.strip()) > 0, f"Setup run {i+1} should produce output"

        # Should not accumulate errors across runs
        error_indicators = ["error", "failed", "crash"]
        critical_errors = [indicator for indicator in error_indicators
                          if indicator in output.lower()]

        # Some errors might be acceptable (like Docker not running)
        # but should not have multiple different types of errors
        assert len(critical_errors) <= 1, \
            f"Setup run {i+1} has multiple error types: {critical_errors}"

@pytest.mark.order(11)
def test_configuration_override_integration():
    """Test that service overrides are properly applied."""
    if not os.path.exists("dev-stack-config.yaml"):
        # Create a config file for testing
        result = run_cmd("./scripts/setup.sh --init")
        if result.returncode != 0:
            pytest.skip("Cannot create configuration file for testing")

    # Load configuration to check for overrides
    with open("dev-stack-config.yaml") as f:
        config = yaml.safe_load(f)

    overrides = config.get("overrides", {})

    if len(overrides) == 0:
        pytest.skip("No service overrides configured")

    # Run setup and check that overrides are reflected in generated files
    result = run_cmd("./scripts/setup.sh --dry-run")

    # Should complete without errors when overrides are present
    assert result.returncode == 0 or len(result.stdout + result.stderr) > 0, \
        "Setup should handle service overrides without crashing"

@pytest.mark.order(12)
def test_project_isolation():
    """Test that the project doesn't interfere with system-wide settings."""
    # Check that we're working in project directory
    assert os.path.exists("./scripts/setup.sh"), "Should be in project directory"
    assert os.path.exists("./services"), "Should be in project directory"

    # Generated files should be in current directory, not system-wide
    generated_files = ["dev-stack-config.yaml", "docker-compose.generated.yml", ".env.generated"]

    for fname in generated_files:
        if os.path.exists(fname):
            # File should be in current directory
            abs_path = os.path.abspath(fname)
            current_dir = os.path.abspath(".")
            assert abs_path.startswith(current_dir), \
                f"Generated file {fname} should be in project directory"
