package refactorindex

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"go/ast"
	"go/token"
	"go/types"
	"golang.org/x/tools/go/packages"
)

// IngestCodeUnitsConfig controls code unit snapshot ingestion.
type IngestCodeUnitsConfig struct {
	DBPath     string
	RootDir    string
	SourcesDir string
}

// IngestCodeUnitsResult reports counts for code unit ingestion.
type IngestCodeUnitsResult struct {
	RunID      int64
	CodeUnits  int
	Snapshots  int
	Packages   int
	Files      int
	BodyBytes  int
	DocEntries int
}

func IngestCodeUnits(ctx context.Context, cfg IngestCodeUnitsConfig) (*IngestCodeUnitsResult, error) {
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
	seenCodeUnits := make(map[string]struct{})
	codeUnitCount := 0
	snapshotCount := 0
	fileCount := 0
	docCount := 0
	bodyBytes := 0

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

			fileBytes, err := os.ReadFile(filePath)
			if err != nil {
				return nil, errors.Wrap(err, "read file")
			}

			for _, decl := range file.Decls {
				switch d := decl.(type) {
				case *ast.FuncDecl:
					obj := pkg.TypesInfo.Defs[d.Name]
					if obj == nil {
						continue
					}
					def, err := buildCodeUnitDef(pkgPath, qualifier, obj)
					if err != nil {
						return nil, err
					}
					codeUnitID, err := store.GetOrCreateCodeUnit(ctx, tx, def)
					if err != nil {
						return nil, err
					}
					if _, seen := seenCodeUnits[def.Hash]; !seen {
						seenCodeUnits[def.Hash] = struct{}{}
						codeUnitCount++
					}

					bodyText, err := extractNodeText(pkg.Fset, fileBytes, d)
					if err != nil {
						return nil, err
					}
					normalized := normalizeBodyText(bodyText)
					bodyHash := hashText(normalized)
					bodyBytes += len(bodyText)

					docText := commentText(d.Doc)
					if docText != "" {
						docCount++
					}

					startLine, startCol, endLine, endCol, err := nodeSpan(pkg.Fset, d)
					if err != nil {
						return nil, err
					}
					if err := store.InsertCodeUnitSnapshot(ctx, tx, runID, fileID, codeUnitID, startLine, startCol, endLine, endCol, bodyHash, bodyText, docText); err != nil {
						return nil, err
					}
					snapshotCount++
				case *ast.GenDecl:
					for _, spec := range d.Specs {
						s, ok := spec.(*ast.TypeSpec)
						if !ok {
							continue
						}
						obj := pkg.TypesInfo.Defs[s.Name]
						if obj == nil {
							continue
						}
						def, err := buildCodeUnitDef(pkgPath, qualifier, obj)
						if err != nil {
							return nil, err
						}
						codeUnitID, err := store.GetOrCreateCodeUnit(ctx, tx, def)
						if err != nil {
							return nil, err
						}
						if _, seen := seenCodeUnits[def.Hash]; !seen {
							seenCodeUnits[def.Hash] = struct{}{}
							codeUnitCount++
						}

						node := ast.Node(s)
						if len(d.Specs) == 1 {
							node = d
						}
						bodyText, err := extractNodeText(pkg.Fset, fileBytes, node)
						if err != nil {
							return nil, err
						}
						normalized := normalizeBodyText(bodyText)
						bodyHash := hashText(normalized)
						bodyBytes += len(bodyText)

						docText := commentText(s.Doc)
						if docText == "" {
							docText = commentText(d.Doc)
						}
						if docText != "" {
							docCount++
						}

						startLine, startCol, endLine, endCol, err := nodeSpan(pkg.Fset, node)
						if err != nil {
							return nil, err
						}
						if err := store.InsertCodeUnitSnapshot(ctx, tx, runID, fileID, codeUnitID, startLine, startCol, endLine, endCol, bodyHash, bodyText, docText); err != nil {
							return nil, err
						}
						snapshotCount++
					}
				}
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "commit code unit ingestion")
	}
	if err := store.FinishRun(ctx, runID); err != nil {
		return nil, err
	}

	return &IngestCodeUnitsResult{
		RunID:      runID,
		CodeUnits:  codeUnitCount,
		Snapshots:  snapshotCount,
		Packages:   len(pkgs),
		Files:      fileCount,
		BodyBytes:  bodyBytes,
		DocEntries: docCount,
	}, nil
}

func buildCodeUnitDef(pkgPath string, qualifier types.Qualifier, obj types.Object) (CodeUnitDef, error) {
	if obj == nil {
		return CodeUnitDef{}, errors.New("nil symbol object")
	}

	kind := codeUnitKind(obj)
	signature := types.TypeString(obj.Type(), qualifier)
	recv := receiverString(obj, qualifier)

	def := CodeUnitDef{
		Pkg:       pkgPath,
		Name:      obj.Name(),
		Kind:      kind,
		Recv:      recv,
		Signature: signature,
	}
	def.Hash = hashCodeUnit(def)
	return def, nil
}

func extractNodeText(fset *token.FileSet, fileBytes []byte, node ast.Node) (string, error) {
	if node == nil {
		return "", errors.New("nil node")
	}
	startPos := fset.PositionFor(node.Pos(), false)
	endPos := fset.PositionFor(node.End(), false)
	if startPos.Offset < 0 || endPos.Offset < 0 {
		return "", errors.New("invalid node offsets")
	}
	if startPos.Offset > len(fileBytes) || endPos.Offset > len(fileBytes) || startPos.Offset > endPos.Offset {
		return "", errors.New("node offsets out of range")
	}
	return string(fileBytes[startPos.Offset:endPos.Offset]), nil
}

func nodeSpan(fset *token.FileSet, node ast.Node) (int, int, int, int, error) {
	if node == nil {
		return 0, 0, 0, 0, errors.New("nil node")
	}
	start := fset.PositionFor(node.Pos(), false)
	end := fset.PositionFor(node.End(), false)
	if start.Line == 0 || end.Line == 0 {
		return 0, 0, 0, 0, errors.New("invalid node positions")
	}
	return start.Line, start.Column, end.Line, end.Column, nil
}

func normalizeBodyText(text string) string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t")
	}
	return strings.Join(lines, "\n")
}

func hashText(text string) string {
	sum := sha256.Sum256([]byte(text))
	return hex.EncodeToString(sum[:])
}

func codeUnitKind(obj types.Object) string {
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
	default:
		return "code_unit"
	}
}

func commentText(doc *ast.CommentGroup) string {
	if doc == nil {
		return ""
	}
	return strings.TrimSpace(doc.Text())
}

func hashCodeUnit(def CodeUnitDef) string {
	parts := []string{def.Pkg, def.Name, def.Kind, def.Recv, def.Signature}
	normalized := strings.Join(parts, "|")
	sum := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(sum[:])
}
