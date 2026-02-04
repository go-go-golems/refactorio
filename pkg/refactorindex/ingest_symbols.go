package refactorindex

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"go/ast"
	"go/token"
	"go/types"
	"golang.org/x/tools/go/packages"
)

// IngestSymbolsConfig controls AST symbol ingestion.
type IngestSymbolsConfig struct {
	DBPath     string
	RootDir    string
	SourcesDir string
}

// IngestSymbolsResult reports counts for symbol ingestion.
type IngestSymbolsResult struct {
	RunID       int64
	Symbols     int
	Occurrences int
	Packages    int
	Files       int
}

func IngestSymbols(ctx context.Context, cfg IngestSymbolsConfig) (*IngestSymbolsResult, error) {
	if strings.TrimSpace(cfg.DBPath) == "" {
		return nil, errors.New("db path is required")
	}
	if strings.TrimSpace(cfg.RootDir) == "" {
		return nil, errors.New("root dir is required")
	}
	rootDir, err := filepath.Abs(cfg.RootDir)
	if err != nil {
		return nil, errors.Wrap(err, "resolve root dir")
	}

	db, err := OpenDB(ctx, cfg.DBPath)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = db.Close()
	}()

	store := NewStore(db)
	if err := store.InitSchema(ctx); err != nil {
		return nil, err
	}

	argsJSON, err := EncodeArgsJSON(map[string]string{
		"root": rootDir,
	})
	if err != nil {
		return nil, err
	}

	runID, err := store.CreateRun(ctx, RunConfig{
		ToolVersion: ToolVersion,
		RootPath:    rootDir,
		SourcesDir:  cfg.SourcesDir,
		ArgsJSON:    argsJSON,
	})
	if err != nil {
		return nil, err
	}

	pkgConfig := &packages.Config{
		Mode: packages.NeedName | packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedFiles | packages.NeedCompiledGoFiles,
		Dir:  rootDir,
	}
	pkgs, err := packages.Load(pkgConfig, "./...")
	if err != nil {
		return nil, errors.Wrap(err, "load packages")
	}
	if packages.PrintErrors(pkgs) > 0 {
		return nil, errors.New("package load errors")
	}

	tx, err := store.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	fileIDs := make(map[string]int64)
	symbolCount := 0
	occurrenceCount := 0
	fileCount := 0

	for _, pkg := range pkgs {
		if pkg.Types == nil || pkg.TypesInfo == nil || pkg.Fset == nil {
			continue
		}
		qualifier := types.RelativeTo(pkg.Types)
		pkgPath := pkg.PkgPath
		for _, file := range pkg.Syntax {
			filePath := pkg.Fset.Position(file.Pos()).Filename
			if filePath == "" {
				continue
			}
			relPath, err := filepath.Rel(rootDir, filePath)
			if err != nil {
				return nil, errors.Wrap(err, "relativize file path")
			}
			relPath = filepath.ToSlash(relPath)
			fileID, ok := fileIDs[relPath]
			if !ok {
				id, err := store.GetOrCreateFile(ctx, tx, relPath)
				if err != nil {
					return nil, err
				}
				fileID = id
				fileIDs[relPath] = id
				fileCount++
			}

			for _, decl := range file.Decls {
				switch d := decl.(type) {
				case *ast.FuncDecl:
					obj := pkg.TypesInfo.Defs[d.Name]
					if obj == nil {
						continue
					}
					def, occ, err := buildSymbolDef(pkg.Fset, pkgPath, qualifier, obj)
					if err != nil {
						return nil, err
					}
					symbolID, err := store.GetOrCreateSymbolDef(ctx, tx, def)
					if err != nil {
						return nil, err
					}
					symbolCount++
					if err := store.InsertSymbolOccurrence(ctx, tx, runID, fileID, symbolID, occ.Line, occ.Col, occ.Exported); err != nil {
						return nil, err
					}
					occurrenceCount++
				case *ast.GenDecl:
					for _, spec := range d.Specs {
						switch s := spec.(type) {
						case *ast.TypeSpec:
							obj := pkg.TypesInfo.Defs[s.Name]
							if obj == nil {
								continue
							}
							def, occ, err := buildSymbolDef(pkg.Fset, pkgPath, qualifier, obj)
							if err != nil {
								return nil, err
							}
							symbolID, err := store.GetOrCreateSymbolDef(ctx, tx, def)
							if err != nil {
								return nil, err
							}
							symbolCount++
							if err := store.InsertSymbolOccurrence(ctx, tx, runID, fileID, symbolID, occ.Line, occ.Col, occ.Exported); err != nil {
								return nil, err
							}
							occurrenceCount++
						case *ast.ValueSpec:
							for _, name := range s.Names {
								obj := pkg.TypesInfo.Defs[name]
								if obj == nil {
									continue
								}
								def, occ, err := buildSymbolDef(pkg.Fset, pkgPath, qualifier, obj)
								if err != nil {
									return nil, err
								}
								symbolID, err := store.GetOrCreateSymbolDef(ctx, tx, def)
								if err != nil {
									return nil, err
								}
								symbolCount++
								if err := store.InsertSymbolOccurrence(ctx, tx, runID, fileID, symbolID, occ.Line, occ.Col, occ.Exported); err != nil {
									return nil, err
								}
								occurrenceCount++
							}
						}
					}
				}
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "commit symbol ingestion")
	}
	if err := store.FinishRun(ctx, runID); err != nil {
		return nil, err
	}

	return &IngestSymbolsResult{
		RunID:       runID,
		Symbols:     symbolCount,
		Occurrences: occurrenceCount,
		Packages:    len(pkgs),
		Files:       fileCount,
	}, nil
}

type symbolOccurrence struct {
	Line     int
	Col      int
	Exported bool
}

func buildSymbolDef(fset *token.FileSet, pkgPath string, qualifier types.Qualifier, obj types.Object) (SymbolDef, symbolOccurrence, error) {
	pos := fset.Position(obj.Pos())
	if pos.Line == 0 {
		return SymbolDef{}, symbolOccurrence{}, errors.New("symbol position not found")
	}

	kind := symbolKind(obj)
	signature := types.TypeString(obj.Type(), qualifier)
	recv := receiverString(obj, qualifier)

	def := SymbolDef{
		Pkg:       pkgPath,
		Name:      obj.Name(),
		Kind:      kind,
		Recv:      recv,
		Signature: signature,
	}
	def.Hash = hashSymbol(def)

	occ := symbolOccurrence{
		Line:     pos.Line,
		Col:      pos.Column,
		Exported: obj.Exported(),
	}

	return def, occ, nil
}

func symbolKind(obj types.Object) string {
	switch o := obj.(type) {
	case *types.Func:
		if o.Type() != nil {
			if sig, ok := o.Type().(*types.Signature); ok {
				if sig.Recv() != nil {
					return "method"
				}
			}
		}
		return "func"
	case *types.TypeName:
		return "type"
	case *types.Const:
		return "const"
	case *types.Var:
		return "var"
	default:
		return "symbol"
	}
}

func receiverString(obj types.Object, qualifier types.Qualifier) string {
	fn, ok := obj.(*types.Func)
	if !ok {
		return ""
	}
	sig, ok := fn.Type().(*types.Signature)
	if !ok || sig.Recv() == nil {
		return ""
	}
	return types.TypeString(sig.Recv().Type(), qualifier)
}

func hashSymbol(def SymbolDef) string {
	parts := []string{def.Pkg, def.Name, def.Kind, def.Recv, def.Signature}
	normalized := strings.Join(parts, "|")
	sum := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(sum[:])
}
