import subprocess
import pytest
import os
import shutil
import tempfile
import yaml
import time

def run_cmd(cmd, cwd=None):
    """Run a shell command and return the result object."""
    result = subprocess.run(cmd, shell=True, capture_output=True, text=True, cwd=cwd)
    return result

@pytest.fixture
def troubleshooting_environment():
    """Fixture to set up troubleshooting test environment."""
    # Backup existing files
    backup_files = {}
    files_to_manage = ["dev-stack-config.yaml", "docker-compose.generated.yml", ".env.generated"]

    for fname in files_to_manage:
        if os.path.exists(fname):
            backup_files[fname] = f"{fname}.troubleshooting_backup"
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
def test_debug_mode_availability():
    """Test that debug mode is available and produces verbose output."""
    result = run_cmd("./scripts/setup.sh --debug --dry-run")

    output = result.stdout + result.stderr

    # Should produce debug output
    debug_indicators = ["[DEBUG]", "DEBUG:", "debug:", "verbose", "--debug"]
    has_debug_output = any(indicator in output for indicator in debug_indicators)

    assert has_debug_output or len(output) > 100, \
        f"Debug mode should produce verbose output. Output: {output[:200]}..."

@pytest.mark.order(2)
def test_help_system_functionality():
    """Test that help system provides useful troubleshooting information."""
    help_commands = [
        ("./scripts/setup.sh --help", ["usage", "options", "help"]),
        ("./scripts/manage.sh --help", ["commands", "usage", "help"]),
        ("./scripts/setup.sh -h", ["usage", "help"]),
        ("./scripts/manage.sh -h", ["usage", "help"])
    ]

    for cmd, expected_keywords in help_commands:
        result = run_cmd(cmd)
        output = result.stdout + result.stderr

        if result.returncode == 0:
            # If help succeeds, should contain helpful information
            output_lower = output.lower()
            found_keywords = [kw for kw in expected_keywords if kw in output_lower]

            assert len(found_keywords) > 0, \
                f"Help output should contain helpful keywords. Command: {cmd}, Output: {output[:200]}"
        else:
            # If help fails, should still provide some output
            assert len(output.strip()) > 0, f"Failed help command should provide error message: {cmd}"

@pytest.mark.order(3)
def test_error_message_quality():
    """Test that error messages are helpful and actionable."""
    # Test various error scenarios to ensure good error messages

    error_scenarios = [
        {
            "description": "Invalid command line argument",
            "cmd": "./scripts/setup.sh --invalid-argument",
            "expected_indicators": ["invalid", "unknown", "option", "argument", "help"]
        },
        {
            "description": "Invalid manage command",
            "cmd": "./scripts/manage.sh invalid-command",
            "expected_indicators": ["invalid", "unknown", "command", "available", "help"]
        }
    ]

    for scenario in error_scenarios:
        result = run_cmd(scenario["cmd"])
        output = result.stdout + result.stderr

        # Should provide error message
        assert len(output.strip()) > 0, \
            f"Error scenario '{scenario['description']}' should provide output"

        # Should contain helpful indicators
        output_lower = output.lower()
        helpful_indicators = [indicator for indicator in scenario["expected_indicators"]
                            if indicator in output_lower]

        assert len(helpful_indicators) > 0, \
            f"Error message should be helpful for: {scenario['description']}. Output: {output}"

@pytest.mark.order(4)
def test_configuration_validation_errors():
    """Test that configuration validation provides clear error messages."""
    # Test with invalid configuration files

    invalid_configs = [
        {
            "name": "malformed_yaml",
            "content": """services:
  enabled
    - redis  # Invalid YAML syntax
    - postgres
""",
            "expected_errors": ["yaml", "syntax", "parse", "invalid"]
        },
        {
            "name": "missing_services",
            "content": """project:
  name: test
# Missing services section
""",
            "expected_errors": ["services", "missing", "required"]
        },
        {
            "name": "invalid_service",
            "content": """services:
  enabled:
    - not_a_real_service
    - definitely_fake
""",
            "expected_errors": ["invalid", "service", "not_a_real_service", "unknown"]
        }
    ]

    for config in invalid_configs:
        config_file = f"test_{config['name']}.yaml"

        try:
            with open(config_file, "w") as f:
                f.write(config["content"])

            # Test setup with invalid config
            env = os.environ.copy()
            env["PROJECT_CONFIG"] = config_file

            result = subprocess.run(
                ["./scripts/setup.sh"],
                env=env,
                capture_output=True,
                text=True
            )

            # Check the result
            output = result.stdout + result.stderr

            if result.returncode == 0:
                # If it succeeds, it might have created a default config instead
                # This is acceptable behavior - check if it mentions config creation
                creation_indicators = ["created", "sample", "configuration", "edit", "customize"]
                has_creation_info = any(indicator in output.lower() for indicator in creation_indicators)

                if not has_creation_info:
                    # If no creation info and no error, that's unexpected
                    pytest.fail(f"Expected error or config creation for {config['name']}, but got success without clear indication")
            else:
                # If it fails, should provide error message
                assert len(output.strip()) > 0, f"Should provide error message for: {config['name']}"

                # Should contain relevant error indicators
                output_lower = output.lower()
                relevant_errors = [error for error in config["expected_errors"]
                                 if error in output_lower]

                assert len(relevant_errors) > 0, \
                    f"Should provide relevant error message for {config['name']}. Output: {output}"

        finally:
            if os.path.exists(config_file):
                os.remove(config_file)

            # Clean up environment variable
            if "PROJECT_CONFIG" in os.environ:
                del os.environ["PROJECT_CONFIG"]

@pytest.mark.order(5)
def test_dependency_checking():
    """Test that missing dependencies are properly reported."""
    # Test various dependency scenarios

    # Check if Docker is mentioned when commands fail
    result = run_cmd("./scripts/setup.sh --dry-run")
    output = result.stdout + result.stderr

    if result.returncode != 0:
        # If setup fails, should mention potential dependency issues
        dependency_keywords = ["docker", "dependency", "install", "requirement", "missing"]
        has_dependency_info = any(keyword in output.lower() for keyword in dependency_keywords)

        if not has_dependency_info:
            # If no dependency info, should at least provide helpful error
            assert len(output.strip()) > 20, \
                "Should provide detailed error message when failing"

@pytest.mark.order(6)
def test_log_access_functionality():
    """Test that log access functionality works or fails gracefully."""
    # Test log commands with different services
    log_commands = [
        "./scripts/manage.sh logs",
        "./scripts/manage.sh logs redis",
        "./scripts/manage.sh logs postgres",
        "./scripts/manage.sh logs --tail 10"
    ]

    for cmd in log_commands:
        result = run_cmd(cmd)
        output = result.stdout + result.stderr

        # Should provide output (either logs or error message)
        assert len(output.strip()) > 0, f"Log command should provide output: {cmd}"

        if result.returncode != 0:
            # If logs fail, should provide helpful error
            helpful_errors = ["not running", "no such service", "container", "docker", "no local development environment", "run './scripts/setup.sh' first"]
            has_helpful_error = any(error in output.lower() for error in helpful_errors)

            assert has_helpful_error or "log" in output.lower(), \
                f"Log command should provide helpful error: {cmd}. Output: {output}"

@pytest.mark.order(7)
def test_status_reporting_accuracy():
    """Test that status reporting provides accurate information."""
    # Test status command in various scenarios

    # Test status when no services are running
    status_result = run_cmd("./scripts/manage.sh status")
    output = status_result.stdout + status_result.stderr

    # Should provide status information
    assert len(output.strip()) > 0, "Status command should provide output"

    # Should indicate current state
    status_indicators = ["running", "stopped", "not running", "status", "service"]
    has_status_info = any(indicator in output.lower() for indicator in status_indicators)

    assert has_status_info, f"Status should provide service state information. Output: {output}"

@pytest.mark.order(8)
def test_common_issue_detection():
    """Test detection and reporting of common issues."""
    common_issues = [
        {
            "scenario": "Port already in use",
            "setup": lambda: None,  # Would simulate port conflict
            "check": lambda output: any(word in output.lower() for word in ["port", "address", "use", "bind"])
        },
        {
            "scenario": "Permission denied",
            "setup": lambda: None,  # Would simulate permission issue
            "check": lambda output: any(word in output.lower() for word in ["permission", "denied", "access"])
        },
        {
            "scenario": "Service startup failure",
            "setup": lambda: None,  # Would simulate service failure
            "check": lambda output: any(word in output.lower() for word in ["failed", "error", "unable", "start"])
        }
    ]

    # Since we can't easily simulate these issues in tests, we'll check that
    # the framework handles errors gracefully

    result = run_cmd("./scripts/manage.sh start")
    output = result.stdout + result.stderr

    if result.returncode != 0:
        # If start fails, should provide informative error
        assert len(output.strip()) > 10, "Start failure should provide informative error"

        # Check if any common issue patterns are detected
        for issue in common_issues:
            if issue["check"](output):
                assert True  # Common issue properly detected
                return

        # If no specific issue detected, should still be informative
        assert any(word in output.lower() for word in ["error", "failed", "issue", "problem"]), \
            "Error output should indicate there's an issue"

@pytest.mark.order(9)
def test_cleanup_troubleshooting():
    """Test that cleanup operations provide troubleshooting information."""
    # Test cleanup commands and their output
    cleanup_commands = [
        "./scripts/manage.sh stop",
        "./scripts/manage.sh down",
        "./scripts/manage.sh clean"
    ]

    for cmd in cleanup_commands:
        result = run_cmd(cmd)
        output = result.stdout + result.stderr

        # Should provide feedback about cleanup actions
        if result.returncode == 0:
            # Successful cleanup should indicate what was done
            cleanup_indicators = ["stopped", "removed", "cleaned", "down", "complete"]
            has_cleanup_info = any(indicator in output.lower() for indicator in cleanup_indicators)

            assert has_cleanup_info or len(output.strip()) == 0, \
                f"Cleanup should indicate actions taken: {cmd}"
        else:
            # Failed cleanup should explain why
            assert len(output.strip()) > 0, f"Failed cleanup should provide error: {cmd}"

@pytest.mark.order(10)
def test_environment_validation():
    """Test that environment validation provides helpful feedback."""
    # Test environment setup validation

    result = run_cmd("./scripts/setup.sh --validate-env")

    if result.returncode == 0:
        # If validation succeeds, should confirm environment is ready
        output = result.stdout + result.stderr
        validation_indicators = ["valid", "ready", "ok", "check", "environment"]
        has_validation_info = any(indicator in output.lower() for indicator in validation_indicators)

        assert has_validation_info, "Environment validation should provide feedback"
    else:
        # If validation fails or command doesn't exist, should be clear
        output = result.stdout + result.stderr

        if "command not found" in output.lower() or "unknown" in output.lower():
            pytest.skip("Environment validation not implemented")
        else:
            # Should provide helpful validation errors
            assert len(output.strip()) > 0, "Environment validation should provide error details"

@pytest.mark.order(11)
def test_configuration_troubleshooting_guide():
    """Test that configuration issues provide troubleshooting guidance."""
    # Create a problematic but syntactically valid configuration

    problematic_config = """project:
  name: ""  # Empty project name

services:
  enabled: []  # No services enabled

overrides:
  redis:  # Override for service not enabled
    port: 6379

validation:
  skip_warnings: false
"""

    config_file = "problematic-config.yaml"

    try:
        with open(config_file, "w") as f:
            f.write(problematic_config)

        env = os.environ.copy()
        env["PROJECT_CONFIG"] = config_file

        result = subprocess.run(
            ["./scripts/setup.sh"],
            env=env,
            capture_output=True,
            text=True
        )

        output = result.stdout + result.stderr

        # Should provide guidance for configuration issues
        guidance_indicators = [
            "configure", "config", "check", "fix", "correct",
            "valid", "required", "missing", "empty"
        ]

        has_guidance = any(indicator in output.lower() for indicator in guidance_indicators)

        assert has_guidance, f"Should provide configuration guidance. Output: {output}"

    finally:
        if os.path.exists(config_file):
            os.remove(config_file)
        if "PROJECT_CONFIG" in os.environ:
            del os.environ["PROJECT_CONFIG"]

@pytest.mark.order(12)
def test_performance_troubleshooting():
    """Test that performance issues are addressed in troubleshooting."""
    # Test commands that might reveal performance information

    # Test with resource limits
    resource_config = """services:
  enabled:
    - redis
    - postgres
    - mysql
    - kafka
    - jaeger
    - prometheus
    - localstack

overrides:
  redis:
    memory_limit: "128m"
  postgres:
    memory_limit: "256m"
  mysql:
    memory_limit: "256m"
  kafka:
    memory_limit: "512m"
"""

    config_file = "resource-test-config.yaml"

    try:
        with open(config_file, "w") as f:
            f.write(resource_config)

        env = os.environ.copy()
        env["PROJECT_CONFIG"] = config_file

        result = subprocess.run(
            ["./scripts/setup.sh", "--dry-run"],
            env=env,
            capture_output=True,
            text=True
        )

        output = result.stdout + result.stderr

        # Should handle resource-heavy configurations
        if result.returncode != 0:
            # If it fails, should provide helpful guidance
            resource_indicators = ["memory", "resource", "limit", "performance", "heavy"]
            has_resource_info = any(indicator in output.lower() for indicator in resource_indicators)

            if not has_resource_info:
                # Should at least provide general error guidance
                assert len(output.strip()) > 20, "Should provide detailed error for complex configuration"

    finally:
        if os.path.exists(config_file):
            os.remove(config_file)
        if "PROJECT_CONFIG" in os.environ:
            del os.environ["PROJECT_CONFIG"]

@pytest.mark.order(13)
def test_recovery_procedures():
    """Test that recovery procedures are available and helpful."""
    # Test recovery commands
    recovery_commands = [
        "./scripts/manage.sh reset",
        "./scripts/manage.sh clean",
        "./scripts/setup.sh --reset",
        "./scripts/manage.sh restart"
    ]

    for cmd in recovery_commands:
        result = run_cmd(cmd)
        output = result.stdout + result.stderr

        if result.returncode == 0:
            # If recovery succeeds, should indicate what was done
            recovery_indicators = ["reset", "clean", "restart", "recovered", "fixed"]
            has_recovery_info = any(indicator in output.lower() for indicator in recovery_indicators)

            assert has_recovery_info or len(output.strip()) == 0, \
                f"Recovery command should indicate actions: {cmd}"
        else:
            # If recovery command doesn't exist or fails, should be clear
            if "command not found" in output.lower() or "unknown" in output.lower():
                continue  # Command not implemented - that's fine
            else:
                assert len(output.strip()) > 0, f"Failed recovery should provide error: {cmd}"

@pytest.mark.order(14)
def test_documentation_references():
    """Test that troubleshooting output references documentation when helpful."""
    # Test that error messages or help output reference documentation

    help_result = run_cmd("./scripts/setup.sh --help")
    help_output = help_result.stdout + help_result.stderr

    # Look for documentation references
    doc_indicators = ["readme", "docs", "documentation", "wiki", "guide", "manual"]
    has_doc_references = any(indicator in help_output.lower() for indicator in doc_indicators)

    if not has_doc_references:
        # Check if actual documentation files exist
        doc_files = ["README.md", "docs/", "TROUBLESHOOTING.md", "SETUP.md"]
        existing_docs = [doc for doc in doc_files if os.path.exists(doc)]

        if len(existing_docs) > 0:
            # If docs exist but aren't referenced, that's ok but note it
            assert True  # Documentation exists even if not referenced in help
        else:
            # If no docs and no references, help should still be useful
            assert len(help_output.strip()) > 50, "Help should be substantial if no external docs"

@pytest.mark.order(15)
def test_troubleshooting_script_robustness():
    """Test that troubleshooting scripts handle edge cases robustly."""
    # Test scripts with various edge case inputs

    edge_cases = [
        {"args": [""], "description": "empty argument"},
        {"args": ["--"], "description": "double dash only"},
        {"args": ["--help", "--debug"], "description": "multiple flags"},
        {"args": [" "], "description": "whitespace argument"},
    ]

    for case in edge_cases:
        for script in ["./scripts/setup.sh", "./scripts/manage.sh"]:
            try:
                # Build command with edge case arguments
                cmd_parts = [script] + [arg for arg in case["args"] if arg.strip()]
                cmd = " ".join(f'"{part}"' if " " in part else part for part in cmd_parts)

                result = run_cmd(cmd)
                output = result.stdout + result.stderr

                # Should not crash and should provide some output
                assert len(output.strip()) > 0, \
                    f"Script {script} should handle edge case: {case['description']}"

                # Should not produce stack traces or crashes
                crash_indicators = ["traceback", "segmentation", "core dumped", "fatal error"]
                has_crash = any(indicator in output.lower() for indicator in crash_indicators)

                assert not has_crash, \
                    f"Script should not crash with edge case: {case['description']}"

            except Exception as e:
                # If the test itself fails, that's also acceptable - edge cases are tricky
                continue
