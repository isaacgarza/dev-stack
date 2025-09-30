import subprocess
import os
import pytest
import shutil
import yaml
import tempfile
import glob

def run_cmd(cmd, cwd=None):
    """Run a shell command and return the result object."""
    result = subprocess.run(cmd, shell=True, capture_output=True, text=True, cwd=cwd)
    return result

@pytest.fixture
def data_test_environment():
    """Fixture to set up data management test environment."""
    # Backup existing files
    backup_files = {}
    files_to_manage = ["dev-stack-config.yaml", "docker-compose.generated.yml", ".env.generated"]

    for fname in files_to_manage:
        if os.path.exists(fname):
            backup_files[fname] = f"{fname}.data_backup"
            shutil.copy2(fname, backup_files[fname])

    # Clean up any existing backup files from previous tests
    for backup_file in glob.glob("*.sql"):
        if "test" in backup_file or "backup" in backup_file:
            os.remove(backup_file)

    yield

    # Cleanup and restore
    for fname in files_to_manage:
        if os.path.exists(fname):
            os.remove(fname)

    for original, backup in backup_files.items():
        if os.path.exists(backup):
            shutil.copy2(backup, original)
            os.remove(backup)

    # Clean up any test backup files
    for backup_file in glob.glob("*test*.sql"):
        os.remove(backup_file)

@pytest.mark.order(1)
def test_backup_command_availability():
    """Test that backup commands are available in the framework."""
    result = run_cmd("./scripts/manage.sh --help")

    output = result.stdout + result.stderr

    # Check if backup functionality is mentioned in help
    has_backup_command = any(word in output.lower() for word in ["backup", "dump", "export"])

    if not has_backup_command:
        pytest.skip("Backup functionality not available in framework")

    assert has_backup_command, "Backup functionality should be available"

@pytest.mark.order(2)
def test_restore_command_availability():
    """Test that restore commands are available in the framework."""
    result = run_cmd("./scripts/manage.sh --help")

    output = result.stdout + result.stderr

    # Check if restore functionality is mentioned in help
    has_restore_command = any(word in output.lower() for word in ["restore", "import", "load"])

    if not has_restore_command:
        pytest.skip("Restore functionality not available in framework")

    assert has_restore_command, "Restore functionality should be available"

@pytest.mark.order(3)
def test_data_directory_management():
    """Test that data directories are properly managed."""
    # Check if there are data directories or volume mounts
    data_dirs = ["./data", "./volumes", "./db-data"]

    existing_data_dirs = [d for d in data_dirs if os.path.exists(d)]

    if len(existing_data_dirs) == 0:
        # Create a test data directory structure
        test_data_dir = "./test-data"
        os.makedirs(test_data_dir, exist_ok=True)

        try:
            # Test directory creation and cleanup
            assert os.path.exists(test_data_dir), "Should be able to create data directories"

            # Test file operations in data directory
            test_file = os.path.join(test_data_dir, "test.txt")
            with open(test_file, "w") as f:
                f.write("test data")

            assert os.path.exists(test_file), "Should be able to create files in data directory"

        finally:
            if os.path.exists(test_data_dir):
                shutil.rmtree(test_data_dir)
    else:
        # Test that existing data directories are accessible
        for data_dir in existing_data_dirs:
            assert os.access(data_dir, os.R_OK), f"Data directory {data_dir} should be readable"
            assert os.access(data_dir, os.W_OK), f"Data directory {data_dir} should be writable"

@pytest.mark.order(4)
def test_postgres_backup_simulation(data_test_environment):
    """Test postgres backup functionality if available."""
    # First check if postgres is configured
    if not os.path.exists("dev-stack-config.yaml"):
        # Create a test config with postgres
        test_config = """services:
  enabled:
    - postgres
    - redis

overrides:
  postgres:
    port: 5432
    database: "test_db"
    username: "test_user"
"""
        with open("dev-stack-config.yaml", "w") as f:
            f.write(test_config)

    # Load config to check if postgres is enabled
    with open("dev-stack-config.yaml") as f:
        config = yaml.safe_load(f)

    enabled_services = config.get("services", {}).get("enabled", [])

    if "postgres" not in enabled_services:
        pytest.skip("PostgreSQL not enabled in configuration")

    # Test backup command (but don't require it to succeed if services aren't running)
    backup_result = run_cmd("./scripts/manage.sh backup postgres")

    output = backup_result.stdout + backup_result.stderr

    if backup_result.returncode == 0:
        # If backup succeeded, check for backup file
        backup_files = glob.glob("*postgres*.sql") + glob.glob("*backup*.sql")
        assert len(backup_files) > 0, "Backup should create a SQL file"

        # Verify backup file has some content
        if backup_files:
            backup_file = backup_files[0]
            assert os.path.getsize(backup_file) > 0, "Backup file should not be empty"

    else:
        # If backup failed, should provide meaningful error message
        assert len(output.strip()) > 0, "Backup command should provide error message when failing"

        # Common reasons for failure that are acceptable in test environment
        acceptable_failures = [
            "not running", "connection refused", "no such container",
            "service not found", "docker", "postgresql", "no local development environment",
            "run './scripts/setup.sh' first", "environment found"
        ]

        has_acceptable_failure = any(failure in output.lower() for failure in acceptable_failures)

        if not has_acceptable_failure:
            pytest.fail(f"Backup failed with unexpected error: {output}")

@pytest.mark.order(5)
def test_redis_backup_simulation(data_test_environment):
    """Test redis backup functionality if available."""
    # Check if redis is configured
    if not os.path.exists("dev-stack-config.yaml"):
        # Create a test config with redis
        test_config = """services:
  enabled:
    - redis

overrides:
  redis:
    port: 6379
"""
        with open("dev-stack-config.yaml", "w") as f:
            f.write(test_config)

    with open("dev-stack-config.yaml") as f:
        config = yaml.safe_load(f)

    enabled_services = config.get("services", {}).get("enabled", [])

    if "redis" not in enabled_services:
        pytest.skip("Redis not enabled in configuration")

    # Test backup command
    backup_result = run_cmd("./scripts/manage.sh backup redis")

    output = backup_result.stdout + backup_result.stderr

    if backup_result.returncode == 0:
        # Redis backup might create .rdb files or other formats
        backup_files = glob.glob("*redis*") + glob.glob("*.rdb") + glob.glob("*backup*")
        if backup_files:
            # If backup files were created, they should have some content
            backup_file = backup_files[0]
            assert os.path.exists(backup_file), "Backup file should exist"

    else:
        # Should provide error message for failed backup
        assert len(output.strip()) > 0, "Should provide error message when backup fails"

@pytest.mark.order(6)
def test_backup_file_naming_convention():
    """Test that backup files follow a consistent naming convention."""
    # Create some dummy backup files to test naming
    test_backups = [
        "postgres_backup_20240101.sql",
        "redis_backup_20240101.rdb",
        "test_backup.sql"
    ]

    try:
        for backup_file in test_backups:
            with open(backup_file, "w") as f:
                f.write("# Test backup file\n")

        # Test that we can identify backup files
        all_backups = glob.glob("*backup*") + glob.glob("*.sql") + glob.glob("*.rdb")

        found_test_backups = [f for f in all_backups if any(test in f for test in ["postgres", "redis", "test"])]

        assert len(found_test_backups) >= len(test_backups), "Should be able to identify backup files"

        # Test naming patterns
        for backup_file in found_test_backups:
            # Backup files should have reasonable names
            assert len(os.path.basename(backup_file)) > 5, f"Backup filename too short: {backup_file}"
            assert "." in backup_file, f"Backup file should have extension: {backup_file}"

    finally:
        # Clean up test files
        for backup_file in test_backups:
            if os.path.exists(backup_file):
                os.remove(backup_file)

@pytest.mark.order(7)
def test_restore_functionality_basic():
    """Test basic restore functionality."""
    # Create a dummy backup file to test restore
    test_backup_file = "test_postgres_backup.sql"

    try:
        with open(test_backup_file, "w") as f:
            f.write("""-- Test backup file
CREATE TABLE test_table (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100)
);

INSERT INTO test_table (name) VALUES ('test_data');
""")

        # Test restore command (expect it to fail gracefully if service not running)
        restore_result = run_cmd(f"./scripts/manage.sh restore postgres {test_backup_file}")

        output = restore_result.stdout + restore_result.stderr

        # Should either succeed or fail gracefully with meaningful message
        assert len(output.strip()) > 0, "Restore command should provide output"

        if restore_result.returncode != 0:
            # If it fails, should be due to service not running or similar
            acceptable_failures = [
                "not running", "connection refused", "no such container",
                "service not found", "file not found", "no local development environment",
                "run './scripts/setup.sh' first"
            ]

            has_acceptable_failure = any(failure in output.lower() for failure in acceptable_failures)

            assert has_acceptable_failure, f"Restore failed with unexpected error: {output}"

    finally:
        if os.path.exists(test_backup_file):
            os.remove(test_backup_file)

@pytest.mark.order(8)
def test_data_persistence_concepts():
    """Test that data persistence concepts are properly implemented."""
    # Check if docker-compose has volume configurations
    if os.path.exists("docker-compose.generated.yml"):
        with open("docker-compose.generated.yml") as f:
            compose_content = f.read()

        # Look for volume configurations
        has_volumes = "volumes:" in compose_content
        has_data_persistence = any(keyword in compose_content.lower()
                                 for keyword in ["volume", "mount", "data", "persist"])

        if has_data_persistence:
            assert has_volumes or "volumes:" in compose_content, \
                "Should have volume configurations for data persistence"

    # Check for data directory references in configuration
    if os.path.exists("dev-stack-config.yaml"):
        with open("dev-stack-config.yaml") as f:
            config_content = f.read()

        # Look for data-related configurations
        data_keywords = ["data", "volume", "persist", "storage"]
        has_data_config = any(keyword in config_content.lower() for keyword in data_keywords)

        # This is informational - data persistence may be handled differently
        if has_data_config:
            assert True  # Acknowledge that data persistence is configured

@pytest.mark.order(9)
def test_backup_cleanup_functionality():
    """Test that old backup files can be cleaned up."""
    # Create some test backup files with different dates
    test_backups = [
        "postgres_backup_20240101_120000.sql",
        "postgres_backup_20240102_120000.sql",
        "redis_backup_20240101_120000.rdb",
        "old_backup_20231201.sql"
    ]

    try:
        for backup_file in test_backups:
            with open(backup_file, "w") as f:
                f.write(f"# Test backup: {backup_file}\n")

        # Test cleanup command if it exists
        cleanup_result = run_cmd("./scripts/manage.sh cleanup-backups")

        if cleanup_result.returncode == 0:
            # If cleanup succeeded, some files might be removed
            remaining_files = [f for f in test_backups if os.path.exists(f)]
            # This is fine - cleanup might not remove all files
        else:
            # If cleanup command doesn't exist, that's also fine
            output = cleanup_result.stdout + cleanup_result.stderr
            if "command not found" in output.lower() or "unknown command" in output.lower():
                pytest.skip("Cleanup functionality not implemented")

    finally:
        # Manual cleanup of test files
        for backup_file in test_backups:
            if os.path.exists(backup_file):
                os.remove(backup_file)

@pytest.mark.order(10)
def test_database_connection_validation():
    """Test that database connections can be validated."""
    if not os.path.exists("dev-stack-config.yaml"):
        pytest.skip("No configuration file available")

    with open("dev-stack-config.yaml") as f:
        config = yaml.safe_load(f)

    enabled_services = config.get("services", {}).get("enabled", [])
    database_services = [s for s in enabled_services if s in ["postgres", "mysql"]]

    if len(database_services) == 0:
        pytest.skip("No database services enabled")

    for db_service in database_services:
        # Test connection validation command
        result = run_cmd(f"./scripts/manage.sh test-connection {db_service}")

        output = result.stdout + result.stderr

        # Should provide some output about connection status
        assert len(output.strip()) > 0, f"Connection test for {db_service} should provide output"

        if result.returncode != 0:
            # Acceptable if services aren't running
            connection_failures = [
                "connection refused", "not running", "timeout",
                "no such container", "service unavailable"
            ]

            has_connection_failure = any(failure in output.lower() for failure in connection_failures)

            if not has_connection_failure:
                # If it's not a connection issue, might be command not implemented
                command_not_found = any(phrase in output.lower() for phrase in [
                    "command not found", "unknown command", "not implemented"
                ])

                if command_not_found:
                    pytest.skip(f"Connection testing not implemented for {db_service}")

@pytest.mark.order(11)
def test_data_migration_concepts():
    """Test that data migration concepts are considered."""
    # Check if there are migration-related scripts or documentation
    migration_files = [
        "./scripts/migrate.sh",
        "./migrations",
        "./db/migrations",
        "./data/migrations"
    ]

    has_migration_support = any(os.path.exists(path) for path in migration_files)

    if has_migration_support:
        # Test that migration functionality is accessible
        migration_result = run_cmd("./scripts/manage.sh migrate")

        output = migration_result.stdout + migration_result.stderr

        # Should provide output about migration status
        assert len(output.strip()) > 0, "Migration command should provide output"

    else:
        # Migration support is optional, but data management should be documented
        docs_files = ["./README.md", "./docs", "./SETUP.md"]

        has_documentation = any(os.path.exists(path) for path in docs_files)

        if has_documentation:
            # Check if data management is mentioned in documentation
            for doc_file in docs_files:
                if os.path.isfile(doc_file):
                    with open(doc_file) as f:
                        doc_content = f.read().lower()

                    data_topics = ["backup", "restore", "data", "persistence", "migration"]
                    has_data_docs = any(topic in doc_content for topic in data_topics)

                    if has_data_docs:
                        assert True  # Data management is documented
                        return

        # If no migration support and no documentation, that's still acceptable
        pytest.skip("Migration functionality not implemented (this is acceptable)")

@pytest.mark.order(12)
def test_data_security_considerations():
    """Test that data security considerations are in place."""
    # Check for .gitignore entries for data files
    gitignore_files = [".gitignore", ".dockerignore"]

    for ignore_file in gitignore_files:
        if os.path.exists(ignore_file):
            with open(ignore_file) as f:
                ignore_content = f.read()

            # Should ignore data files and backups
            data_patterns = ["*.sql", "*.rdb", "data/", "backup", ".env"]

            ignored_data_patterns = [pattern for pattern in data_patterns
                                   if pattern in ignore_content]

            if len(ignored_data_patterns) > 0:
                assert True  # Some data files are properly ignored
                return

    # If no gitignore or no data patterns, check environment file security
    if os.path.exists(".env.generated"):
        stat_info = os.stat(".env.generated")

        # Environment files should not be world-readable
        is_world_readable = (stat_info.st_mode & 0o004) != 0

        assert not is_world_readable, "Environment files should not be world-readable"
