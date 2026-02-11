package docgen

import (
	"embed"
	"io/fs"
	"os"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

//go:embed templates/*.html
var templatesFS embed.FS

//go:embed all:content
var contentFS embed.FS

//go:embed static
var staticFS embed.FS

// ResolveFS returns os.DirFS(localDir) if localDir exists on disk,
// otherwise returns the embedded filesystem. This lets dev mode use
// local files (live editing) while production uses the embedded copy.
func ResolveFS(localDir string, embedded fs.FS) fs.FS {
	if info, err := os.Stat(localDir); err == nil && info.IsDir() {
		return os.DirFS(localDir)
	}
	return embedded
}

// EmbeddedMigrations returns the embedded migrations filesystem,
// rooted at the "migrations" subdirectory.
func EmbeddedMigrations() fs.FS {
	sub, _ := fs.Sub(migrationsFS, "migrations")
	return sub
}

// EmbeddedTemplates returns the embedded templates filesystem,
// rooted at the "templates" subdirectory.
func EmbeddedTemplates() fs.FS {
	sub, _ := fs.Sub(templatesFS, "templates")
	return sub
}

// EmbeddedContent returns the embedded content filesystem,
// rooted at the "content" subdirectory.
func EmbeddedContent() fs.FS {
	sub, _ := fs.Sub(contentFS, "content")
	return sub
}

// EmbeddedStatic returns the embedded static filesystem,
// rooted at the "static" subdirectory.
func EmbeddedStatic() fs.FS {
	sub, _ := fs.Sub(staticFS, "static")
	return sub
}
