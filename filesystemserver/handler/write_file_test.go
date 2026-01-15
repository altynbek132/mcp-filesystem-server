package handler

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleWriteFile(t *testing.T) {
	// Setup a temporary directory for the test
	tmpDir := t.TempDir()

	// Create a handler with the temp dir as an allowed path
	allowedDirs := resolveAllowedDirs(t, tmpDir)
	fsHandler, err := NewFilesystemHandler(allowedDirs)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("write to a new file", func(t *testing.T) {
		filePath := filepath.Join(tmpDir, "new_file.txt")
		content := "Hello, world!"
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{
					"path":    filePath,
					"content": content,
				},
			},
		}

		res, err := fsHandler.HandleWriteFile(ctx, req)
		require.NoError(t, err)
		require.False(t, res.IsError)

		// Verify the file was created with correct content
		readContent, err := os.ReadFile(filePath)
		require.NoError(t, err)
		assert.Equal(t, content, string(readContent))
	})

	t.Run("write to nested non-existent directory (mkdir -p behavior)", func(t *testing.T) {
		filePath := filepath.Join(tmpDir, "a", "b", "c", "nested_file.txt")
		content := "Nested content"
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{
					"path":    filePath,
					"content": content,
				},
			},
		}

		res, err := fsHandler.HandleWriteFile(ctx, req)
		require.NoError(t, err)
		require.False(t, res.IsError)

		// Verify the file and directories were created
		require.DirExists(t, filepath.Join(tmpDir, "a", "b", "c"))
		readContent, err := os.ReadFile(filePath)
		require.NoError(t, err)
		assert.Equal(t, content, string(readContent))
	})

	t.Run("overwrite existing file", func(t *testing.T) {
		filePath := filepath.Join(tmpDir, "existing_file.txt")
		err := os.WriteFile(filePath, []byte("original"), 0644)
		require.NoError(t, err)

		newContent := "overwritten"
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{
					"path":    filePath,
					"content": newContent,
				},
			},
		}

		res, err := fsHandler.HandleWriteFile(ctx, req)
		require.NoError(t, err)
		require.False(t, res.IsError)

		readContent, err := os.ReadFile(filePath)
		require.NoError(t, err)
		assert.Equal(t, newContent, string(readContent))
	})

	t.Run("try to write to a directory", func(t *testing.T) {
		dirPath := filepath.Join(tmpDir, "some_directory")
		err := os.Mkdir(dirPath, 0755)
		require.NoError(t, err)

		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{
					"path":    dirPath,
					"content": "content",
				},
			},
		}

		res, err := fsHandler.HandleWriteFile(ctx, req)
		require.NoError(t, err)
		require.True(t, res.IsError)
		assert.Contains(t, res.Content[0].(mcp.TextContent).Text, "Cannot write to a directory")
	})

	t.Run("path in non-allowed directory", func(t *testing.T) {
		otherDir := t.TempDir()
		filePath := filepath.Join(otherDir, "forbidden.txt")

		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{
					"path":    filePath,
					"content": "content",
				},
			},
		}

		res, err := fsHandler.HandleWriteFile(ctx, req)
		require.NoError(t, err)
		require.True(t, res.IsError)
		assert.Contains(t, res.Content[0].(mcp.TextContent).Text, "access denied")
	})
}
