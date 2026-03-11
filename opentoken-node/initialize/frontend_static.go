package initialize

import (
	"errors"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// RegisterFrontendStaticRoutes registers handlers to serve a Vite-built SPA (frontend/dist).
// This is intended for the Windows deliverable where the backend serves the UI without nginx.
func RegisterFrontendStaticRoutes(r *gin.Engine) {
	distDir, err := locateFrontendDistDir()
	if err != nil {
		return
	}

	indexPath := filepath.Join(distDir, "index.html")
	r.GET("/", func(c *gin.Context) {
		c.File(indexPath)
	})

	// Serve other static files and provide SPA fallback (e.g. deep links).
	r.NoRoute(func(c *gin.Context) {
		requestPath := c.Request.URL.Path
		if strings.HasPrefix(requestPath, "/api/") {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		relative := strings.TrimPrefix(path.Clean(requestPath), "/")
		filePath := filepath.Join(distDir, filepath.FromSlash(relative))
		if fileExists(filePath) {
			c.File(filePath)
			return
		}
		c.File(indexPath)
	})
}

func locateFrontendDistDir() (string, error) {
	if env := strings.TrimSpace(os.Getenv("DREAM_AI_FRONTEND_DIST")); env != "" {
		if hasIndexHTML(env) {
			return env, nil
		}
	}

	candidates := make([]string, 0, 10)

	if exePath, err := os.Executable(); err == nil && exePath != "" {
		exeDir := filepath.Dir(exePath)
		candidates = append(candidates,
			filepath.Join(exeDir, "dist"),
			filepath.Join(exeDir, "web", "dist"),
		)
	}

	if cwd, err := os.Getwd(); err == nil && cwd != "" {
		candidates = append(candidates,
			filepath.Join(cwd, "dist"),
			filepath.Join(cwd, "frontend", "dist"),
			filepath.Join(cwd, "..", "frontend", "dist"),
		)
	}

	for _, dir := range candidates {
		if hasIndexHTML(dir) {
			return dir, nil
		}
	}
	return "", errors.New("frontend dist not found")
}

func hasIndexHTML(dir string) bool {
	if dir == "" {
		return false
	}
	_, err := os.Stat(filepath.Join(dir, "index.html"))
	return err == nil
}

func fileExists(p string) bool {
	if p == "" {
		return false
	}
	info, err := os.Stat(p)
	if err != nil {
		return false
	}
	return !info.IsDir()
}
