package workbenchapi

import "net/http"

func registerBaseRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", nil)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
	})
}
