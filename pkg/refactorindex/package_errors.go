package refactorindex

import (
	"context"
	"fmt"

	"golang.org/x/tools/go/packages"
)

type goPackagesErrorMeta struct {
	Severity string `json:"severity"`
	Package  string `json:"package"`
	Position string `json:"position,omitempty"`
	Kind     string `json:"kind,omitempty"`
	Message  string `json:"message"`
}

func recordGoPackagesErrors(ctx context.Context, store *Store, runID int64, pkgs []*packages.Package) error {
	for _, pkg := range pkgs {
		for _, perr := range pkg.Errors {
			meta := goPackagesErrorMeta{
				Severity: "warning",
				Package:  pkg.PkgPath,
				Position: perr.Pos,
				Kind:     fmt.Sprint(perr.Kind),
				Message:  perr.Msg,
			}
			if err := store.InsertRunMetadataJSON(ctx, runID, "go_packages_error", meta); err != nil {
				return err
			}
		}
	}
	return nil
}
