package workbenchapi

import (
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"
)

func decodeJSON(w http.ResponseWriter, r *http.Request, dest interface{}) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dest); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "failed to decode JSON body", map[string]string{"error": err.Error()})
		return errors.Wrap(err, "decode json")
	}
	if dec.More() {
		writeError(w, http.StatusBadRequest, "invalid_json", "unexpected extra JSON values", nil)
		return errors.New("extra json values")
	}
	return nil
}
