package router

import (
	"embed"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/controller"
	"github.com/QuantumNous/new-api/middleware"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

// ThemeAssets holds the embedded frontend assets for both themes.
type ThemeAssets struct {
	DefaultBuildFS   embed.FS
	DefaultIndexPage []byte
	ClassicBuildFS   embed.FS
	ClassicIndexPage []byte
}

func SetWebRouter(router *gin.Engine, assets ThemeAssets) {
	// Fallback: serve index.html from disk when embed FS is empty
	if assets.DefaultIndexPage == nil {
		data, err := os.ReadFile(filepath.Join("web", "default", "dist", "index.html"))
		if err == nil {
			assets.DefaultIndexPage = data
		}
	}
	if assets.ClassicIndexPage == nil {
		data, err := os.ReadFile(filepath.Join("web", "classic", "dist", "index.html"))
		if err == nil {
			assets.ClassicIndexPage = data
		}
	}

	// Determine the static file root directory based on theme
	getStaticDir := func() string {
		if common.GetTheme() == "classic" {
			return filepath.Join("web", "classic", "dist")
		}
		return filepath.Join("web", "default", "dist")
	}

	router.Use(gzip.Gzip(gzip.DefaultCompression))

	// Static file handler runs BEFORE rate limit to avoid 429s on page load.
	router.Use(func(c *gin.Context) {
		path := c.Request.URL.Path

		// Serve index.html directly from disk (fresh on every request)
		if path == "/" || path == "/index.html" {
			indexPath := filepath.Join(getStaticDir(), "index.html")
			if info, err := os.Stat(indexPath); err == nil && !info.IsDir() {
				c.Header("Cache-Control", "no-cache, must-revalidate")
				http.ServeFile(c.Writer, c.Request, indexPath)
				c.Abort()
				return
			}
		}

		if !strings.HasPrefix(path, "/static/") && !strings.HasPrefix(path, "/logo.") && !strings.HasPrefix(path, "/favicon.") && !strings.HasPrefix(path, "/umin-") && path != "/logo.png" && path != "/favicon.ico" && !strings.HasPrefix(path, "/pay-") && !strings.HasPrefix(path, "/waffo-") && !strings.HasPrefix(path, "/yoomoney-") {
			c.Next()
			return
		}
		fullPath := filepath.Join(getStaticDir(), path)
		if info, err := os.Stat(fullPath); err == nil && !info.IsDir() {
			c.Header("Cache-Control", "max-age=86400")
			http.ServeFile(c.Writer, c.Request, fullPath)
			c.Abort()
			return
		}
		// For /static/ paths: return 404 instead of falling through to SPA index.html
		// This prevents SyntaxError from returned HTML when old JS files are missing
		if strings.HasPrefix(path, "/static/") {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		c.Next()
	})

	router.Use(middleware.GlobalWebRateLimit())
	router.Use(middleware.Cache())

	router.NoRoute(func(c *gin.Context) {
		c.Set(middleware.RouteTagKey, "web")
		if strings.HasPrefix(c.Request.RequestURI, "/v1") || strings.HasPrefix(c.Request.RequestURI, "/api") || strings.HasPrefix(c.Request.RequestURI, "/assets") {
			controller.RelayNotFound(c)
			return
		}
		c.Header("Cache-Control", "no-cache")
		if common.GetTheme() == "classic" {
			c.Data(http.StatusOK, "text/html; charset=utf-8", assets.ClassicIndexPage)
		} else {
			c.Data(http.StatusOK, "text/html; charset=utf-8", assets.DefaultIndexPage)
		}
	})
}
