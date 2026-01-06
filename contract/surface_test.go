package contract

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestContractSurfaceSnapshot(t *testing.T) {
	snapshot, err := buildContractSurface()
	if err != nil {
		t.Fatalf("build surface: %v", err)
	}

	expectedPath := filepath.Join("testdata", "contract_surface.txt")
	expectedBytes, err := os.ReadFile(expectedPath)
	if err != nil {
		if os.IsNotExist(err) || os.Getenv("UPDATE_CONTRACT_SURFACE") == "true" {
			if writeErr := os.WriteFile(expectedPath, []byte(snapshot), 0o644); writeErr != nil {
				t.Fatalf("write snapshot: %v", writeErr)
			}
			expectedBytes = []byte(snapshot)
		} else {
			t.Fatalf("read snapshot: %v", err)
		}
	}

	if snapshot != string(expectedBytes) {
		t.Fatalf("contract surface changed.\n--- got ---\n%s\n--- want ---\n%s", snapshot, string(expectedBytes))
	}
}

func buildContractSurface() (string, error) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, ".", func(info os.FileInfo) bool {
		return !strings.HasSuffix(info.Name(), "_test.go")
	}, parser.AllErrors)
	if err != nil {
		return "", err
	}

	pkg, ok := pkgs["contract"]
	if !ok {
		return "", os.ErrNotExist
	}

	var files []*ast.File
	var names []string
	for name := range pkg.Files {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		files = append(files, pkg.Files[name])
	}

	cfg := &types.Config{Importer: importer.Default()}
	checked, err := cfg.Check("contract", fset, files, nil)
	if err != nil {
		return "", err
	}

	scope := checked.Scope()
	var exported []string
	for _, name := range scope.Names() {
		if !token.IsExported(name) {
			continue
		}
		obj := scope.Lookup(name)
		switch o := obj.(type) {
		case *types.Func:
			exported = append(exported, "func "+o.Name()+" "+types.TypeString(o.Type(), types.RelativeTo(checked)))
		case *types.TypeName:
			exported = append(exported, "type "+o.Name()+" "+types.TypeString(o.Type().Underlying(), types.RelativeTo(checked)))
		}
	}

	sort.Strings(exported)

	return strings.Join(exported, "\n") + "\n", nil
}
