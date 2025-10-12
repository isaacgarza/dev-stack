package docker

import (
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name        string
		logger      *slog.Logger
		expectError bool
	}{
		{
			name:        "create client with valid logger",
			logger:      slog.New(slog.NewTextHandler(os.Stdout, nil)),
			expectError: false,
		},
		{
			name:        "create client with nil logger",
			logger:      nil,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.logger)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, client)
			} else {
				if err != nil {
					// Docker might not be available in test environment
					t.Skipf("Docker not available: %v", err)
				}
				require.NotNil(t, client)
				assert.NotNil(t, client.cli)
				assert.Equal(t, tt.logger, client.logger)

				// Clean up
				_ = client.Close()
			}
		})
	}
}

func TestClient_Close(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	client, err := NewClient(logger)
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}

	t.Run("close client", func(t *testing.T) {
		err := client.Close()
		assert.NoError(t, err)
	})

	t.Run("close client twice", func(t *testing.T) {
		// Should not panic on double close
		assert.NotPanics(t, func() {
			_ = client.Close()
		})
	})
}

func TestClient_GetCli(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	client, err := NewClient(logger)
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}
	defer func() { _ = client.Close() }()

	t.Run("get underlying docker client", func(t *testing.T) {
		// Test that the underlying client is accessible
		assert.NotNil(t, client.cli)
	})
}

func TestClient_GetLogger(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	client, err := NewClient(logger)
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}
	defer func() { _ = client.Close() }()

	t.Run("get logger", func(t *testing.T) {
		assert.Equal(t, logger, client.logger)
	})
}

func TestClient_WithNilLogger(t *testing.T) {
	client, err := NewClient(nil)
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}
	defer func() {
		if client != nil {
			_ = client.Close()
		}
	}()

	t.Run("client works with nil logger", func(t *testing.T) {
		assert.NotNil(t, client)
		assert.NotNil(t, client.cli)
		assert.Nil(t, client.logger)
	})
}
