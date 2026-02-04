package refactorindex

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/dop251/goja"
	"github.com/go-go-golems/refactorio/pkg/refactor/js/modules"
	"github.com/go-go-golems/refactorio/pkg/refactorindex"
	"github.com/pkg/errors"
)

type Module struct {
	store      *refactorindex.Store
	runID      int64
	maxResults int
	ctx        context.Context
	traceSeq   int
	traceEnc   *json.Encoder
	traceClose io.Closer
}

var _ modules.NativeModule = (*Module)(nil)

func NewModule(store *refactorindex.Store, runID int64) *Module {
	return &Module{
		store:      store,
		runID:      runID,
		maxResults: 5000,
		ctx:        context.Background(),
	}
}

func (m *Module) EnableTrace(writer io.Writer) {
	if writer == nil {
		m.traceEnc = nil
		m.traceClose = nil
		return
	}
	m.traceEnc = json.NewEncoder(writer)
}

func (m *Module) EnableTraceFile(path string) error {
	if path == "" {
		return errors.New("trace path is required")
	}
	f, err := os.Create(path)
	if err != nil {
		return errors.Wrap(err, "create trace file")
	}
	m.traceClose = f
	m.traceEnc = json.NewEncoder(f)
	return nil
}

func (m *Module) CloseTrace() error {
	if m.traceClose == nil {
		return nil
	}
	err := m.traceClose.Close()
	m.traceClose = nil
	return err
}

func (m *Module) Name() string { return "refactor-index" }

func (m *Module) Doc() string {
	return `
Refactor index module exposes read-only query helpers.

Functions:
  querySymbols(filter)
  queryRefs(symbolHash)
  queryDocHits(terms, fileset)
  queryFiles(fileset)
`
}

func (m *Module) Loader(vm *goja.Runtime, moduleObj *goja.Object) {
	exports := moduleObj.Get("exports").(*goja.Object)

	exports.Set("querySymbols", func(call goja.FunctionCall) goja.Value {
		records, err := m.querySymbols(vm, call)
		if err != nil {
			panic(vm.ToValue(err.Error()))
		}
		return vm.ToValue(records)
	})

	exports.Set("queryRefs", func(call goja.FunctionCall) goja.Value {
		records, err := m.queryRefs(vm, call)
		if err != nil {
			panic(vm.ToValue(err.Error()))
		}
		return vm.ToValue(records)
	})

	exports.Set("queryDocHits", func(call goja.FunctionCall) goja.Value {
		records, err := m.queryDocHits(vm, call)
		if err != nil {
			panic(vm.ToValue(err.Error()))
		}
		return vm.ToValue(records)
	})

	exports.Set("queryFiles", func(call goja.FunctionCall) goja.Value {
		records, err := m.queryFiles(vm, call)
		if err != nil {
			panic(vm.ToValue(err.Error()))
		}
		return vm.ToValue(records)
	})
}

type symbolFilter struct {
	Pkg          string `json:"pkg"`
	Name         string `json:"name"`
	Kind         string `json:"kind"`
	Path         string `json:"path"`
	ExportedOnly bool   `json:"exported_only"`
	Limit        int    `json:"limit"`
	Offset       int    `json:"offset"`
}

type fileset struct {
	Include []string `json:"include"`
	Exclude []string `json:"exclude"`
}

func (m *Module) querySymbols(vm *goja.Runtime, call goja.FunctionCall) ([]map[string]interface{}, error) {
	var filter symbolFilter
	if len(call.Arguments) > 0 && !goja.IsUndefined(call.Arguments[0]) && !goja.IsNull(call.Arguments[0]) {
		if err := vm.ExportTo(call.Arguments[0], &filter); err != nil {
			return nil, errors.Wrap(err, "export symbol filter")
		}
	}

	records, err := m.store.ListSymbolInventory(m.ctx, refactorindex.SymbolInventoryFilter{
		RunID:        m.runID,
		ExportedOnly: filter.ExportedOnly,
		Kind:         filter.Kind,
		Name:         filter.Name,
		Pkg:          filter.Pkg,
		Path:         filter.Path,
		Limit:        clampLimit(filter.Limit, m.maxResults),
		Offset:       filter.Offset,
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(records, func(i, j int) bool {
		if records[i].Pkg != records[j].Pkg {
			return records[i].Pkg < records[j].Pkg
		}
		if records[i].Name != records[j].Name {
			return records[i].Name < records[j].Name
		}
		if records[i].Kind != records[j].Kind {
			return records[i].Kind < records[j].Kind
		}
		if records[i].FilePath != records[j].FilePath {
			return records[i].FilePath < records[j].FilePath
		}
		if records[i].Line != records[j].Line {
			return records[i].Line < records[j].Line
		}
		return records[i].Col < records[j].Col
	})

	results := make([]map[string]interface{}, 0, len(records))
	for _, record := range records {
		results = append(results, map[string]interface{}{
			"symbol_hash": record.SymbolHash,
			"pkg":         record.Pkg,
			"name":        record.Name,
			"kind":        record.Kind,
			"recv":        record.Recv,
			"signature":   record.Signature,
			"def_span":    fmt.Sprintf("%s:%d:%d", record.FilePath, record.Line, record.Col),
			"file":        record.FilePath,
			"line":        record.Line,
			"col":         record.Col,
			"is_exported": record.IsExported,
		})
	}
	m.writeTrace("querySymbols", filter, len(results))
	return results, nil
}

func (m *Module) queryRefs(vm *goja.Runtime, call goja.FunctionCall) ([]map[string]interface{}, error) {
	if len(call.Arguments) == 0 {
		return nil, errors.New("queryRefs requires symbol hash")
	}
	arg := call.Arguments[0]
	if goja.IsUndefined(arg) || goja.IsNull(arg) {
		return nil, errors.New("queryRefs requires symbol hash")
	}
	symbolHash := arg.String()
	if symbolHash == "" {
		return nil, errors.New("queryRefs requires symbol hash")
	}

	records, err := m.store.ListSymbolRefs(m.ctx, refactorindex.SymbolRefFilter{
		RunID:      m.runID,
		SymbolHash: symbolHash,
		Limit:      m.maxResults,
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(records, func(i, j int) bool {
		if records[i].FilePath != records[j].FilePath {
			return records[i].FilePath < records[j].FilePath
		}
		if records[i].Line != records[j].Line {
			return records[i].Line < records[j].Line
		}
		if records[i].Col != records[j].Col {
			return records[i].Col < records[j].Col
		}
		return records[i].SymbolHash < records[j].SymbolHash
	})

	results := make([]map[string]interface{}, 0, len(records))
	for _, record := range records {
		results = append(results, map[string]interface{}{
			"symbol_hash": record.SymbolHash,
			"path":        record.FilePath,
			"line":        record.Line,
			"col":         record.Col,
			"is_decl":     record.IsDecl,
			"source":      record.Source,
			"commit_hash": record.CommitHash,
		})
	}
	m.writeTrace("queryRefs", map[string]interface{}{"symbol_hash": symbolHash}, len(results))
	return results, nil
}

func (m *Module) queryDocHits(vm *goja.Runtime, call goja.FunctionCall) ([]map[string]interface{}, error) {
	var terms []string
	if len(call.Arguments) > 0 && !goja.IsUndefined(call.Arguments[0]) && !goja.IsNull(call.Arguments[0]) {
		if err := vm.ExportTo(call.Arguments[0], &terms); err != nil {
			return nil, errors.Wrap(err, "export terms")
		}
	}
	var fs fileset
	if len(call.Arguments) > 1 && !goja.IsUndefined(call.Arguments[1]) && !goja.IsNull(call.Arguments[1]) {
		parsed, err := parseFileset(vm, call.Arguments[1])
		if err != nil {
			return nil, err
		}
		fs = parsed
	}

	records, err := m.store.ListDocHits(m.ctx, refactorindex.DocHitFilter{
		RunID: m.runID,
		Terms: terms,
		Limit: m.maxResults,
	})
	if err != nil {
		return nil, err
	}

	filtered := make([]refactorindex.DocHitRecord, 0, len(records))
	for _, record := range records {
		if ok, err := matchFileset(record.FilePath, fs); err != nil {
			return nil, err
		} else if !ok {
			continue
		}
		filtered = append(filtered, record)
	}

	sort.Slice(filtered, func(i, j int) bool {
		if filtered[i].FilePath != filtered[j].FilePath {
			return filtered[i].FilePath < filtered[j].FilePath
		}
		if filtered[i].Line != filtered[j].Line {
			return filtered[i].Line < filtered[j].Line
		}
		if filtered[i].Col != filtered[j].Col {
			return filtered[i].Col < filtered[j].Col
		}
		return filtered[i].Term < filtered[j].Term
	})

	results := make([]map[string]interface{}, 0, len(filtered))
	for _, record := range filtered {
		results = append(results, map[string]interface{}{
			"term":       record.Term,
			"path":       record.FilePath,
			"line":       record.Line,
			"col":        record.Col,
			"match_text": record.MatchText,
		})
	}
	m.writeTrace("queryDocHits", map[string]interface{}{"terms": terms, "fileset": fs}, len(results))
	return results, nil
}

func (m *Module) queryFiles(vm *goja.Runtime, call goja.FunctionCall) ([]map[string]interface{}, error) {
	var fs fileset
	if len(call.Arguments) > 0 && !goja.IsUndefined(call.Arguments[0]) && !goja.IsNull(call.Arguments[0]) {
		parsed, err := parseFileset(vm, call.Arguments[0])
		if err != nil {
			return nil, err
		}
		fs = parsed
	}

	records, err := m.store.ListFiles(m.ctx, refactorindex.FileFilter{Limit: m.maxResults})
	if err != nil {
		return nil, err
	}

	filtered := make([]refactorindex.FileRecord, 0, len(records))
	for _, record := range records {
		if ok, err := matchFileset(record.Path, fs); err != nil {
			return nil, err
		} else if !ok {
			continue
		}
		filtered = append(filtered, record)
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Path < filtered[j].Path
	})

	results := make([]map[string]interface{}, 0, len(filtered))
	for _, record := range filtered {
		results = append(results, map[string]interface{}{
			"path":      record.Path,
			"ext":       record.Ext,
			"exists":    record.Exists,
			"is_binary": record.IsBinary,
		})
	}
	m.writeTrace("queryFiles", fs, len(results))
	return results, nil
}

func clampLimit(value, max int) int {
	if max <= 0 {
		return value
	}
	if value <= 0 || value > max {
		return max
	}
	return value
}

func parseFileset(vm *goja.Runtime, value goja.Value) (fileset, error) {
	var fs fileset
	if value == nil || goja.IsUndefined(value) || goja.IsNull(value) {
		return fs, nil
	}
	if err := vm.ExportTo(value, &fs); err == nil && (len(fs.Include) > 0 || len(fs.Exclude) > 0) {
		return fs, nil
	}
	obj := value.ToObject(vm)
	if obj == nil {
		return fs, nil
	}
	includeValue := obj.Get("include")
	if includeValue != nil && !goja.IsUndefined(includeValue) && !goja.IsNull(includeValue) {
		if err := vm.ExportTo(includeValue, &fs.Include); err != nil {
			return fs, errors.Wrap(err, "export fileset.include")
		}
	}
	excludeValue := obj.Get("exclude")
	if excludeValue != nil && !goja.IsUndefined(excludeValue) && !goja.IsNull(excludeValue) {
		if err := vm.ExportTo(excludeValue, &fs.Exclude); err != nil {
			return fs, errors.Wrap(err, "export fileset.exclude")
		}
	}
	return fs, nil
}

type traceEntry struct {
	Seq         int         `json:"seq"`
	Action      string      `json:"action"`
	Args        interface{} `json:"args"`
	ResultCount int         `json:"result_count"`
}

func (m *Module) writeTrace(action string, args interface{}, count int) {
	if m.traceEnc == nil {
		return
	}
	m.traceSeq++
	_ = m.traceEnc.Encode(traceEntry{
		Seq:         m.traceSeq,
		Action:      action,
		Args:        args,
		ResultCount: count,
	})
}

func matchFileset(path string, fs fileset) (bool, error) {
	included := len(fs.Include) == 0
	for _, pattern := range fs.Include {
		match, err := doublestar.Match(pattern, path)
		if err != nil {
			return false, errors.Wrapf(err, "match include pattern %s", pattern)
		}
		if match {
			included = true
			break
		}
	}
	if !included {
		return false, nil
	}
	for _, pattern := range fs.Exclude {
		match, err := doublestar.Match(pattern, path)
		if err != nil {
			return false, errors.Wrapf(err, "match exclude pattern %s", pattern)
		}
		if match {
			return false, nil
		}
	}
	return true, nil
}
