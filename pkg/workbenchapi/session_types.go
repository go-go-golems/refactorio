package workbenchapi

type SessionRuns struct {
	Commits    *int64 `json:"commits,omitempty"`
	Diff       *int64 `json:"diff,omitempty"`
	Symbols    *int64 `json:"symbols,omitempty"`
	CodeUnits  *int64 `json:"code_units,omitempty"`
	DocHits    *int64 `json:"doc_hits,omitempty"`
	GoplsRefs  *int64 `json:"gopls_refs,omitempty"`
	TreeSitter *int64 `json:"tree_sitter,omitempty"`
}

type Session struct {
	ID           string          `json:"id"`
	WorkspaceID  string          `json:"workspace_id,omitempty"`
	RootPath     string          `json:"root_path,omitempty"`
	GitFrom      string          `json:"git_from,omitempty"`
	GitTo        string          `json:"git_to,omitempty"`
	Runs         SessionRuns     `json:"runs"`
	Availability map[string]bool `json:"availability"`
	LastUpdated  string          `json:"last_updated,omitempty"`
}

type SessionOverride struct {
	ID       string      `json:"id"`
	RootPath string      `json:"root_path,omitempty"`
	GitFrom  string      `json:"git_from,omitempty"`
	GitTo    string      `json:"git_to,omitempty"`
	Runs     SessionRuns `json:"runs"`
}
