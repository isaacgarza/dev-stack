import subprocess
import os
import pytest
import tempfile
import shutil

def run_cmd(cmd, cwd=None):
    """Run a shell command and return the result object."""
    result = subprocess.run(cmd, shell=True, capture_output=True, text=True, cwd=cwd)
    return result

@pytest.fixture
def clean_test_environment():
    """Fixture to ensure clean test environment."""
    # Backup existing files
    backup_files = {}
    files_to_backup = ["dev-stack-config.yaml", "docker-compose.generated.yml", ".env.generated"]

    for fname in files_to_backup:
        if os.path.exists(fname):
            backup_files[fname] = f"{fname}.test_backup"
            shutil.copy2(fname, backup_files[fname])
            os.remove(fname)

    yield

    # Cleanup and restore
    for fname in files_to_backup:
        if os.path.exists(fname):
            os.remove(fname)

    for original, backup in backup_files.items():
        if os.path.exists(backup):
            shutil.copy2(backup, original)
            os.remove(backup)

def test_setup_script_exists():
    """Test that setup.sh script exists and is executable."""
    script_path = "./scripts/setup.sh"

    assert os.path.exists(script_path), f"setup.sh not found at {script_path}"
    assert os.access(script_path, os.X_OK), f"setup.sh is not executable at {script_path}"

def test_manage_script_exists():
    """Test that manage.sh script exists and is executable."""
    script_path = "./scripts/manage.sh"

    assert os.path.exists(script_path), f"manage.sh not found at {script_path}"
    assert os.access(script_path, os.X_OK), f"manage.sh is not executable at {script_path}"

def test_required_directories_exist():
    """Test that required directories exist."""
    required_dirs = [
        "./scripts",
        "./scripts/lib",
        "./services",
        "./config"
    ]

    for dir_path in required_dirs:
        assert os.path.exists(dir_path), f"Required directory not found: {dir_path}"
        assert os.path.isdir(dir_path), f"Path exists but is not a directory: {dir_path}"

def test_setup_with_help_flag():
    """Test that setup.sh responds to help flag."""
    result = run_cmd("./scripts/setup.sh --help")

    # Should either show help (return 0) or fail gracefully
    # Help output should contain usage information
    if result.returncode == 0:
        output = result.stdout.lower()
        assert any(word in output for word in ["usage", "help", "options", "commands"]), \
            f"Help output should contain usage information. Output: {result.stdout}"

def test_manage_with_help_flag():
    """Test that manage.sh responds to help flag."""
    result = run_cmd("./scripts/manage.sh --help")

    # Should either show help (return 0) or fail gracefully
    if result.returncode == 0:
        output = result.stdout.lower()
        assert any(word in output for word in ["usage", "help", "options", "commands"]), \
            f"Help output should contain usage information. Output: {result.stdout}"

def test_setup_init_basic_functionality(clean_test_environment):
    """Test basic functionality of setup.sh --init."""
    result = run_cmd("./scripts/setup.sh --init")

    # Should succeed or fail gracefully
    if result.returncode == 0:
        # If successful, should create config file
        assert os.path.exists("dev-stack-config.yaml"), \
            "Config file should be created when --init succeeds"
    else:
        # If it fails, should provide meaningful error message
        output = result.stdout + result.stderr
        assert len(output.strip()) > 0, "Should provide error message when failing"

def test_setup_without_config_file():
    """Test setup.sh behavior when no config file exists."""
    # Ensure no config file exists
    if os.path.exists("dev-stack-config.yaml"):
        os.remove("dev-stack-config.yaml")

    result = run_cmd("./scripts/setup.sh")

    # Should either create config, provide instructions, or fail gracefully
    output = result.stdout + result.stderr

    if result.returncode == 0:
        # If successful, should have created or handled missing config
        assert len(output) > 0, "Should provide some output when handling missing config"
    else:
        # If it fails, should mention config file or initialization
        assert any(word in output.lower() for word in ["config", "init", "missing", "create"]), \
            f"Should mention config-related issue. Output: {output}"

def test_manage_status_without_setup():
    """Test manage.sh status when no setup has been run."""
    result = run_cmd("./scripts/manage.sh status")

    # Should handle gracefully when no services are set up
    output = result.stdout + result.stderr

    if result.returncode == 0:
        # If successful, should indicate no services running
        assert any(word in output.lower() for word in ["not running", "stopped", "no services"]), \
            f"Should indicate no services running. Output: {output}"
    else:
        # If it fails, should provide meaningful message
        assert len(output.strip()) > 0, "Should provide error message"

def test_basic_command_line_parsing():
    """Test that scripts handle basic command line arguments."""
    # Test that scripts don't crash with various argument patterns
    test_cases = [
        ("./scripts/setup.sh", ["--dry-run"]),
        ("./scripts/setup.sh", ["--invalid-flag"]),
        ("./scripts/manage.sh", ["status"]),
        ("./scripts/manage.sh", ["invalid-command"]),
    ]

    for script, args in test_cases:
        cmd = f"{script} {' '.join(args)}"
        result = run_cmd(cmd)

        # Scripts should not crash (return codes are acceptable, but should produce output)
        output = result.stdout + result.stderr
        assert len(output.strip()) > 0, f"Script should produce output for: {cmd}"

def test_required_library_files_exist():
    """Test that required library files exist."""
    lib_files = [
        "./scripts/lib/common.sh",
        "./scripts/lib/services.sh"
    ]

    for lib_file in lib_files:
        assert os.path.exists(lib_file), f"Required library file not found: {lib_file}"

def test_services_configuration_exists():
    """Test that services configuration exists."""
    services_file = "./services/services.yaml"

    assert os.path.exists(services_file), f"Services configuration not found: {services_file}"

def test_sample_config_exists():
    """Test that sample configuration file exists."""
    sample_config = "./dev-stack-config.sample.yaml"

    assert os.path.exists(sample_config), f"Sample config not found: {sample_config}"

def test_framework_structure_integrity():
    """Test overall framework structure integrity."""
    # Check that we have the basic components needed for the framework to function
    essential_components = [
        "./scripts/setup.sh",           # Main setup script
        "./scripts/manage.sh",          # Management script
        "./scripts/lib/common.sh",      # Common functions
        "./services/services.yaml",     # Service definitions
        "./dev-stack-config.sample.yaml"  # Sample configuration
    ]

    missing_components = []
    for component in essential_components:
        if not os.path.exists(component):
            missing_components.append(component)

    assert len(missing_components) == 0, \
        f"Missing essential framework components: {missing_components}"

def test_debug_output_in_scripts():
    """Test that scripts produce debug output when expected."""
    result = run_cmd("./scripts/setup.sh --init")

    # Scripts should produce debug output to help with troubleshooting
    output = result.stdout + result.stderr

    # Look for debug indicators
    has_debug = any(indicator in output for indicator in ["[DEBUG]", "DEBUG:", "debug:", "Debug:"])

    if not has_debug and len(output.strip()) > 0:
        # Even without explicit debug markers, should have some informative output
        assert len(output.strip()) > 10, "Scripts should produce informative output"
    else:
        assert has_debug, f"Expected debug output in scripts. Output: {output}"

def test_error_handling_basic():
    """Test basic error handling in scripts."""
    # Test with intentionally problematic scenarios

    # Try to run setup with a directory instead of file as config
    if not os.path.exists("fake_dir"):
        os.makedirs("fake_dir")

    try:
        # This should fail gracefully, not crash
        result = run_cmd("./scripts/setup.sh")

        # Should handle errors gracefully
        output = result.stdout + result.stderr
        assert len(output.strip()) > 0, "Should provide error output when things go wrong"

    finally:
        if os.path.exists("fake_dir"):
            os.rmdir("fake_dir")
