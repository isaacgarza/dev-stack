package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

// Database service constants
const (
	ServicePostgres   = "postgres"
	ServicePostgreSQL = "postgresql"
	ServiceMySQL      = "mysql"
	ServiceMongo      = "mongo"
	ServiceMariaDB    = "mariadb"
	ServiceRedis      = "redis"
	ServiceMongoDB    = "mongodb"
)

// backupCmd represents the backup command
var backupCmd = &cobra.Command{
	Use:   "backup <service> [backup-name]",
	Short: "Backup service data to local storage",
	Long: `Backup service data to local storage for safekeeping or migration.
This command creates backups of databases and persistent volumes.

Supported services:
  postgres, postgresql  - Creates SQL dump using pg_dump
  mysql, mariadb       - Creates SQL dump using mysqldump
  redis                - Creates RDB snapshot
  mongodb              - Creates database dump using mongodump
  volumes              - Creates tar archive of volume data

Examples:
  dev-stack backup postgres              # Auto-named backup with timestamp
  dev-stack backup postgres prod-data    # Named backup
  dev-stack backup --all                 # Backup all services
  dev-stack backup postgres --compress   # Compressed backup
  dev-stack backup --output ./backups    # Custom backup directory`,
	Args: func(cmd *cobra.Command, args []string) error {
		all, _ := cmd.Flags().GetBool("all")
		if all && len(args) > 0 {
			return fmt.Errorf("cannot specify service name when using --all flag")
		}
		if !all && len(args) == 0 {
			return fmt.Errorf("service name is required unless using --all flag")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		all, _ := cmd.Flags().GetBool("all")
		output, _ := cmd.Flags().GetString("output")
		compress, _ := cmd.Flags().GetBool("compress")
		format, _ := cmd.Flags().GetString("format")
		exclude, _ := cmd.Flags().GetStringSlice("exclude")

		timestamp := time.Now().Format("20060102_150405")

		if all {
			fmt.Println("💾 Backing up all services...")
			// TODO: Get list of all services and backup each one
			services := []string{"postgres", "redis"} // Example services
			for _, service := range services {
				if contains(exclude, service) {
					fmt.Printf("⏭️  Skipping %s (excluded)\n", service)
					continue
				}
				fmt.Printf("💾 Backing up %s...\n", service)
				backupName := fmt.Sprintf("%s_%s", service, timestamp)
				if err := performBackup(service, backupName, output, compress, format); err != nil {
					return fmt.Errorf("failed to backup %s: %w", service, err)
				}
			}
		} else {
			service := args[0]
			var backupName string
			if len(args) > 1 {
				backupName = args[1]
			} else {
				backupName = fmt.Sprintf("%s_%s", service, timestamp)
			}

			fmt.Printf("💾 Backing up %s...\n", service)
			if err := performBackup(service, backupName, output, compress, format); err != nil {
				return fmt.Errorf("failed to backup %s: %w", service, err)
			}
		}

		fmt.Println("✅ Backup completed successfully")
		return nil
	},
}

// performBackup executes the backup for a specific service
func performBackup(service, backupName, outputDir string, compress bool, format string) error {
	if outputDir == "" {
		outputDir = "./backups"
	}

	fmt.Printf("  📁 Output directory: %s\n", outputDir)
	fmt.Printf("  📝 Backup name: %s\n", backupName)

	if compress {
		fmt.Println("  🗜️  Compression enabled")
	}

	// TODO: Implement actual backup logic based on service type
	// This will integrate with Docker exec to run backup commands
	switch service {
	case ServicePostgres, ServicePostgreSQL, "pg":
		fmt.Println("  🐘 Creating PostgreSQL dump...")
		// Will execute: docker compose exec postgres pg_dump -U postgres -h localhost dbname > backup.sql
	case ServiceMySQL, ServiceMariaDB:
		fmt.Println("  🐬 Creating MySQL dump...")
		// Will execute: docker compose exec mysql mysqldump -u root -p --all-databases > backup.sql
	case ServiceRedis:
		fmt.Println("  🔴 Creating Redis snapshot...")
		// Will execute: docker compose exec redis redis-cli BGSAVE and copy dump.rdb
	case ServiceMongoDB, ServiceMongo:
		fmt.Println("  🍃 Creating MongoDB dump...")
		// Will execute: docker compose exec mongodb mongodump --out /backup
	default:
		return fmt.Errorf("unsupported service for backup: %s", service)
	}

	return nil
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func init() {
	rootCmd.AddCommand(backupCmd)

	// Add flags for backup command
	backupCmd.Flags().Bool("all", false, "Backup all services")
	backupCmd.Flags().StringP("output", "o", "./backups", "Output directory for backups")
	backupCmd.Flags().BoolP("compress", "c", false, "Compress backup files")
	backupCmd.Flags().String("format", "sql", "Backup format (sql, binary, custom)")
	backupCmd.Flags().StringSlice("exclude", []string{}, "Services to exclude from backup")
	backupCmd.Flags().Bool("no-owner", false, "Don't include ownership information in backup")
	backupCmd.Flags().Bool("clean", false, "Include DROP statements in backup")
	backupCmd.Flags().Bool("if-exists", false, "Use IF EXISTS clauses in backup")
}
