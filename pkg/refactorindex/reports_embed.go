package refactorindex

import "embed"

//go:embed reports/queries/*.sql reports/templates/*.md.tmpl
var reportsFS embed.FS
