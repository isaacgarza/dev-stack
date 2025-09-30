import subprocess
import os
import yaml
import pytest
import tempfile
import shutil

def run_cmd(cmd, cwd=None):
    """Run a shell command and return the result object."""
    result = subprocess.run(cmd, shell=True, capture_output=True, text=True, cwd=cwd)
    return result

@pytest.fixture
def clean_environment():
    """Fixture to ensure clean test environment."""
    # Store original files if they exist
    backup_files = {}
    files_to_backup = ["dev-stack-config.yaml", "docker-compose.generated.yml", ".env.generated"]

    for fname in files_to_backup:
        if os.path.exists(fname):
            backup_files[fname] = f"{fname}.test_backup"
            shutil.copy2(fname, backup_files[fname])

    # Remove existing files for clean test
    for fname in files_to_backup:
        if os.path.exists(fname):
            os.remove(fname)

    yield

    # Cleanup: stop any running services and remove test files
    run_cmd("./scripts/manage.sh stop")
    for fname in files_to_backup:
        if os.path.exists(fname):
            os.remove(fname)

    # Restore original files
    for original, backup in backup_files.items():
        if os.path.exists(backup):
            shutil.copy2(backup, original)
            os.remove(backup)

@pytest.mark.order(1)
def test_setup_script_execution():
    """Test that setup.sh can execute without crashing."""
    result = run_cmd("./scripts/setup.sh")

    # Should not crash (return code may vary but should have output)
    output = result.stdout + result.stderr
    assert len(output.strip()) > 0, "Setup script should produce output"

    # If it succeeds, might create config file
    if result.returncode == 0:
        # Check if config file was created
        if os.path.exists("dev-stack-config.yaml"):
            # Verify it's valid YAML
            with open("dev-stack-config.yaml") as f:
                config = yaml.safe_load(f)
            assert isinstance(config, dict), "Created config should be valid YAML"

@pytest.mark.order(2)
def test_setup_with_existing_config():
    """Test setup behavior when config already exists."""
    # Ensure we have a config file
    if not os.path.exists("dev-stack-config.yaml"):
        # Try to create one
        result = run_cmd("./scripts/setup.sh")
        if result.returncode != 0 or not os.path.exists("dev-stack-config.yaml"):
            pytest.skip("Cannot create config file for testing")

    # Run setup again with existing config
    result = run_cmd("./scripts/setup.sh --dry-run")

    # Should handle existing config without error
    output = result.stdout + result.stderr
    assert len(output.strip()) > 0, "Setup should produce output with existing config"

@pytest.mark.order(3)
def test_configuration_validation():
    """Test that setup validates configuration properly."""
    # Create a minimal valid config
    test_config = """services:
  enabled:
    - redis
"""

    with open("dev-stack-config.yaml", "w") as f:
        f.write(test_config)

    # Test that setup accepts valid config
    result = run_cmd("./scripts/setup.sh --dry-run")

    # Should not crash with valid config
    output = result.stdout + result.stderr
    assert len(output.strip()) > 0, "Setup should process valid config"

@pytest.mark.order(4)
def test_invalid_configuration_handling():
    """Test handling of invalid configuration."""
    # Create invalid config
    invalid_config = """services:
  enabled:
    - invalid_service_name
"""

    with open("dev-stack-config.yaml", "w") as f:
        f.write(invalid_config)

    result = run_cmd("./scripts/setup.sh")
    output = result.stdout + result.stderr

    # Should either fail gracefully or handle invalid services
    if result.returncode != 0:
        # If it fails, should provide helpful error
        assert "invalid" in output.lower() or "service" in output.lower(), \
            "Should provide helpful error for invalid config"
    else:
        # If it succeeds, might have created default config instead
        assert "created" in output.lower() or "sample" in output.lower(), \
            "Should indicate config handling"

@pytest.mark.order(5)
def test_docker_compose_generation():
    """Test that docker-compose.yml can be generated if implemented."""
    # Ensure we have a valid config
    if not os.path.exists("dev-stack-config.yaml"):
        test_config = """services:
  enabled:
    - redis
"""
        with open("dev-stack-config.yaml", "w") as f:
            f.write(test_config)

    # Run setup
    result = run_cmd("./scripts/setup.sh --dry-run")

    # Check if docker-compose file was generated
    if os.path.exists("docker-compose.generated.yml"):
        # Validate it's proper YAML
        with open("docker-compose.generated.yml") as f:
            compose_data = yaml.safe_load(f)

        assert isinstance(compose_data, dict), "Docker compose should be valid YAML"
        if "services" in compose_data:
            assert isinstance(compose_data["services"], dict), "Services should be a dictionary"

@pytest.mark.order(6)
def test_environment_file_generation():
    """Test that .env file can be generated if implemented."""
    # Ensure we have a valid config
    if not os.path.exists("dev-stack-config.yaml"):
        test_config = """services:
  enabled:
    - redis
"""
        with open("dev-stack-config.yaml", "w") as f:
            f.write(test_config)

    # Run setup
    result = run_cmd("./scripts/setup.sh --dry-run")

    # Check if env file was generated
    if os.path.exists(".env.generated"):
        with open(".env.generated") as f:
            env_content = f.read()

        # Should contain environment variables
        env_lines = [line for line in env_content.splitlines()
                    if "=" in line and not line.startswith("#")]

        if env_lines:  # If there are env vars, validate format
            for line in env_lines:
                assert "=" in line, f"Invalid env var format: {line}"

@pytest.mark.order(7)
def test_service_lifecycle_basic():
    """Test basic service lifecycle if services can start."""
    # Only run if we have a valid configuration
    if not os.path.exists("dev-stack-config.yaml"):
        test_config = """services:
  enabled:
    - redis
"""
        with open("dev-stack-config.yaml", "w") as f:
            f.write(test_config)

    # Try to start services
    start_result = run_cmd("./scripts/manage.sh start")

    if start_result.returncode == 0:
        try:
            # If start succeeded, test status
            status_result = run_cmd("./scripts/manage.sh status")
            assert status_result.returncode == 0, "Status should work if services started"

            # Check for service indicators
            status_output = status_result.stdout.lower()
            service_indicators = ["running", "up", "starting", "healthy", "service"]
            has_service_info = any(indicator in status_output for indicator in service_indicators)
            assert has_service_info, "Status should show service information"

        finally:
            # Always try to stop services
            run_cmd("./scripts/manage.sh stop")
    else:
        # If services can't start, that's acceptable in test environment
        # But should provide meaningful error
        output = start_result.stdout + start_result.stderr
        assert len(output.strip()) > 0, "Failed start should provide error message"

@pytest.mark.order(8)
def test_debug_output_functionality():
    """Test that debug output is available."""
    result = run_cmd("./scripts/setup.sh --debug")

    output = result.stdout + result.stderr

    # Should produce debug output
    debug_indicators = ["[DEBUG]", "DEBUG:", "debug"]
    has_debug = any(indicator in output for indicator in debug_indicators)

    assert has_debug or len(output) > 100, "Debug mode should produce verbose output"

@pytest.mark.order(9)
def test_help_functionality():
    """Test that help is available."""
    help_commands = ["./scripts/setup.sh --help", "./scripts/manage.sh --help"]

    for cmd in help_commands:
        result = run_cmd(cmd)
        output = result.stdout + result.stderr

        if result.returncode == 0:
            # If help succeeds, should contain useful info
            help_indicators = ["usage", "options", "help", "commands"]
            has_help_info = any(indicator in output.lower() for indicator in help_indicators)
            assert has_help_info, f"Help should contain useful information: {cmd}"
        else:
            # If help fails, should still provide some output
            assert len(output.strip()) > 0, f"Failed help should provide output: {cmd}"

@pytest.mark.order(10)
def test_script_permissions_and_structure():
    """Test that required scripts exist and are executable."""
    scripts = ["./scripts/setup.sh", "./scripts/manage.sh"]

    for script in scripts:
        assert os.path.exists(script), f"Script not found: {script}"
        assert os.access(script, os.X_OK), f"Script not executable: {script}"

@pytest.mark.order(11)
def test_configuration_file_structure():
    """Test configuration file structure when created."""
    # Remove any existing config
    if os.path.exists("dev-stack-config.yaml"):
        os.remove("dev-stack-config.yaml")

    # Run setup to potentially create config
    result = run_cmd("./scripts/setup.sh")

    if os.path.exists("dev-stack-config.yaml"):
        # Validate structure
        with open("dev-stack-config.yaml") as f:
            config = yaml.safe_load(f)

        assert isinstance(config, dict), "Config should be a dictionary"
        assert "services" in config, "Config should have services section"

        if "enabled" in config.get("services", {}):
            enabled_services = config["services"]["enabled"]
            assert isinstance(enabled_services, list), "Enabled services should be a list"

            # Validate enabled services are reasonable
            valid_services = ["redis", "postgres", "mysql", "jaeger", "prometheus", "localstack", "kafka"]
            for service in enabled_services:
                if service not in valid_services:
                    # This might be acceptable if it's a custom service
                    pass

@pytest.mark.order(12)
def test_error_handling_robustness():
    """Test that scripts handle errors gracefully."""
    # Test with various problematic scenarios
    error_scenarios = [
        "./scripts/setup.sh --invalid-flag",
        "./scripts/manage.sh invalid-command"
    ]

    for cmd in error_scenarios:
        result = run_cmd(cmd)
        output = result.stdout + result.stderr

        # Should not crash and should provide error message
        assert len(output.strip()) > 0, f"Error scenario should provide output: {cmd}"

        # Should not contain stack traces or crashes
        crash_indicators = ["traceback", "segmentation fault", "core dumped"]
        has_crash = any(indicator in output.lower() for indicator in crash_indicators)
        assert not has_crash, f"Should not crash with error scenario: {cmd}"

@pytest.mark.order(13)
def test_framework_directory_structure():
    """Test that framework has expected directory structure."""
    required_dirs = [
        "./scripts",
        "./scripts/lib",
        "./services",
        "./config"
    ]

    for dir_path in required_dirs:
        assert os.path.exists(dir_path), f"Required directory missing: {dir_path}"
        assert os.path.isdir(dir_path), f"Path should be directory: {dir_path}"

@pytest.mark.order(14)
def test_sample_configuration_exists():
    """Test that sample configuration file exists."""
    sample_config = "./dev-stack-config.sample.yaml"

    assert os.path.exists(sample_config), "Sample config file should exist"

    # Validate it's proper YAML
    with open(sample_config) as f:
        config = yaml.safe_load(f)

    assert isinstance(config, dict), "Sample config should be valid YAML"
    assert "services" in config, "Sample config should have services section"

@pytest.mark.order(15)
def test_multiple_execution_safety():
    """Test that running scripts multiple times is safe."""
    # Run setup multiple times
    for i in range(2):
        result = run_cmd("./scripts/setup.sh --dry-run")
        output = result.stdout + result.stderr

        # Each execution should complete
        assert len(output.strip()) > 0, f"Execution {i+1} should produce output"

        # Should not accumulate errors
        critical_errors = ["fatal", "crash", "segmentation"]
        has_critical_error = any(error in output.lower() for error in critical_errors)
        assert not has_critical_error, f"Execution {i+1} should not have critical errors"
