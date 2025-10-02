package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// connectCmd represents the connect command
var connectCmd = &cobra.Command{
	Use:   "connect <service>",
	Short: "Connect directly to a database or service",
	Long: `Connect directly to a database or service using its native client.
This command provides quick access to common services like databases
without needing to remember the exact connection parameters.

Supported services:
  postgres, postgresql, pg  - Connect to PostgreSQL using psql
  redis                     - Connect to Redis using redis-cli
  mysql, mariadb           - Connect to MySQL/MariaDB using mysql client
  mongo, mongodb           - Connect to MongoDB using mongosh
  elastic, elasticsearch   - Connect to Elasticsearch

Examples:
  dev-stack connect postgres     # Connect to PostgreSQL
  dev-stack connect redis        # Connect to Redis CLI
  dev-stack connect mysql        # Connect to MySQL
  dev-stack connect --user admin postgres  # Connect as specific user`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		service := args[0]
		user, _ := cmd.Flags().GetString("user")
		database, _ := cmd.Flags().GetString("database")
		host, _ := cmd.Flags().GetString("host")
		port, _ := cmd.Flags().GetString("port")

		fmt.Printf("Connecting to %s...\n", service)

		if user != "" {
			fmt.Printf("User: %s\n", user)
		}

		if database != "" {
			fmt.Printf("Database: %s\n", database)
		}

		if host != "" {
			fmt.Printf("Host: %s\n", host)
		}

		if port != "" {
			fmt.Printf("Port: %s\n", port)
		}

		// TODO: Implement service connection logic
		// This will map service names to appropriate connection commands
		// and execute them with the correct parameters

		switch service {
		case "postgres", "postgresql", "pg":
			fmt.Println("Connecting to PostgreSQL...")
			// Will execute: docker compose exec postgres psql -U <user> -d <database>
		case "redis":
			fmt.Println("Connecting to Redis...")
			// Will execute: docker compose exec redis redis-cli
		case "mysql", "mariadb":
			fmt.Println("Connecting to MySQL/MariaDB...")
			// Will execute: docker compose exec mysql mysql -u <user> -p <database>
		case "mongo", "mongodb":
			fmt.Println("Connecting to MongoDB...")
			// Will execute: docker compose exec mongodb mongosh
		case "elastic", "elasticsearch":
			fmt.Println("Connecting to Elasticsearch...")
			// Will execute: docker compose exec elasticsearch curl localhost:9200
		default:
			return fmt.Errorf("unsupported service: %s", service)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(connectCmd)

	// Add flags for connect command
	connectCmd.Flags().StringP("user", "u", "", "Username to connect with")
	connectCmd.Flags().StringP("database", "d", "", "Database name to connect to")
	connectCmd.Flags().StringP("host", "h", "", "Host to connect to (overrides service discovery)")
	connectCmd.Flags().StringP("port", "p", "", "Port to connect to (overrides service discovery)")
	connectCmd.Flags().String("password", "", "Password to use (if not using environment variables)")
	connectCmd.Flags().Bool("read-only", false, "Connect in read-only mode (if supported)")
}
