// Package web embeds the client templates and static assets (CSS/JS) into
// the compiled binary, so the server has no runtime dependency on files
// existing on disk. Import this package from handlers and use the exported
// FS variables to build template sets and a static file server.
package web

import "embed"

//go:embed templates
var TemplatesFS embed.FS

//go:embed static
var StaticFS embed.FS