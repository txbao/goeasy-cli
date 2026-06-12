package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const minimalSysApisOpenAPI = `openapi: 3.0.3
info:
  title: SysApis API
  version: 1.0.0
paths:
  /api/v1/admin/system/apis:
    get:
      tags: [sys_apis]
      operationId: listSysApis
    post:
      tags: [sys_apis]
      operationId: createSysApis
  /api/v1/admin/system/apis/{id}:
    get:
      tags: [sys_apis]
      operationId: getSysApis
    put:
      tags: [sys_apis]
      operationId: updateSysApis
    delete:
      tags: [sys_apis]
      operationId: deleteSysApis
components:
  schemas:
    SysApisDTO:
      type: object
      properties:
        id: { type: string }
        name: { type: string }
    CreateSysApisRequest:
      type: object
      properties:
        name: { type: string }
    UpdateSysApisRequest:
      type: object
      properties:
        name: { type: string }
`

func TestGenerateHTTPFromOpenAPIIdempotent(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/demo\n"), 0644); err != nil {
		t.Fatal(err)
	}
	cfgDir := filepath.Join(dir, "configs")
	if err := os.MkdirAll(cfgDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte(`codegen:
  domains:
    iam:
      table_prefix: sys_
      modules:
        sys_apis:
          resource: apis
`), 0644); err != nil {
		t.Fatal(err)
	}
	apiDir := filepath.Join(dir, "api", "openapi", "admin", "iam")
	if err := os.MkdirAll(apiDir, 0755); err != nil {
		t.Fatal(err)
	}
	openapiPath := filepath.Join(apiDir, "sys-apis.openapi.yaml")
	if err := os.WriteFile(openapiPath, []byte(minimalSysApisOpenAPI), 0644); err != nil {
		t.Fatal(err)
	}

	opts := GenHTTPOptions{
		ModuleOptions: ModuleOptions{
			ProjectDir: dir,
			ConfigPath: "configs/config.yaml",
			Domain:     "system",
			Resource:   "apis",
		},
		OpenAPIFile: openapiPath,
	}

	if err := GenerateHTTPFromOpenAPI(opts); err != nil {
		t.Fatalf("first gen http: %v", err)
	}
	wirePath := filepath.Join(dir, "internal", "bootstrap", "snippets", "sys_apis_wire.md")
	if _, err := os.Stat(wirePath); err == nil {
		t.Fatal("sys_apis_wire.md should not be generated")
	}

	if err := GenerateHTTPFromOpenAPI(opts); err != nil {
		t.Fatalf("second gen http without --force should not error: %v", err)
	}
}

func TestGenerateHTTPForceBlockedWhenRepositoryPGExists(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/demo\n"), 0644); err != nil {
		t.Fatal(err)
	}
	cfgDir := filepath.Join(dir, "configs")
	if err := os.MkdirAll(cfgDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte(`codegen:
  domains:
    system:
      modules:
        sys_apis:
          resource: apis
`), 0644); err != nil {
		t.Fatal(err)
	}
	openapiPath := filepath.Join(dir, "sys-apis.openapi.yaml")
	if err := os.WriteFile(openapiPath, []byte(minimalSysApisOpenAPI), 0644); err != nil {
		t.Fatal(err)
	}
	pgDir := filepath.Join(dir, "internal", "infrastructure", "system", "persistence", "apis")
	if err := os.MkdirAll(pgDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pgDir, "repository_pg.go"), []byte("package apis\n"), 0644); err != nil {
		t.Fatal(err)
	}

	opts := GenHTTPOptions{
		ModuleOptions: ModuleOptions{
			ProjectDir: dir,
			Force:      true,
			Domain:     "system",
			Resource:   "apis",
		},
		OpenAPIFile: openapiPath,
	}
	if err := GenerateHTTPFromOpenAPI(opts); err == nil {
		t.Fatal("expected error when --force with repository_pg and no --allow-overwrite")
	}

	opts.AllowOverwrite = true
	if err := GenerateHTTPFromOpenAPI(opts); err != nil {
		t.Fatalf("allow-overwrite should permit force: %v", err)
	}
}

const openAPIWithExtraRoute = `openapi: 3.0.3
info:
  title: SysApis API
  version: 1.0.0
paths:
  /api/v1/admin/system/apis:
    get:
      tags: [sys_apis]
      operationId: listSysApis
    post:
      tags: [sys_apis]
      operationId: createSysApis
  /api/v1/admin/system/apis/{id}:
    get:
      tags: [sys_apis]
      operationId: getSysApis
    put:
      tags: [sys_apis]
      operationId: updateSysApis
    delete:
      tags: [sys_apis]
      operationId: deleteSysApis
  /api/v1/admin/system/apis/batch:
    post:
      tags: [sys_apis]
      operationId: batchSysApis
      summary: Batch import APIs
components:
  schemas:
    SysApisDTO:
      type: object
      properties:
        id: { type: string }
        name: { type: string }
    CreateSysApisRequest:
      type: object
      properties:
        name: { type: string }
    UpdateSysApisRequest:
      type: object
      properties:
        name: { type: string }
`

func TestGenerateHTTPMergeOpenAPIExtras(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/demo\n"), 0644); err != nil {
		t.Fatal(err)
	}
	cfgDir := filepath.Join(dir, "configs")
	if err := os.MkdirAll(cfgDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte(`codegen:
  domains:
    system:
      modules:
        sys_apis:
          resource: apis
`), 0644); err != nil {
		t.Fatal(err)
	}
	openapiPath := filepath.Join(dir, "sys-apis.openapi.yaml")
	if err := os.WriteFile(openapiPath, []byte(openAPIWithExtraRoute), 0644); err != nil {
		t.Fatal(err)
	}

	baseOpts := GenHTTPOptions{
		ModuleOptions: ModuleOptions{
			ProjectDir: dir,
			ConfigPath: "configs/config.yaml",
			Domain:     "system",
			Resource:   "apis",
		},
		OpenAPIFile: openapiPath,
	}
	if err := GenerateHTTPFromOpenAPI(baseOpts); err != nil {
		t.Fatalf("initial gen: %v", err)
	}
	handlerPath := filepath.Join(dir, "internal", "interface", "http", "admin", "system", "apis", "handler.go")
	if _, err := os.Stat(handlerPath); err != nil {
		t.Fatalf("handler.go missing: %v", err)
	}

	mergeOpts := baseOpts
	mergeOpts.MergeHTTP = true
	if err := GenerateHTTPFromOpenAPI(mergeOpts); err != nil {
		t.Fatalf("merge-http: %v", err)
	}
	extraHandler := filepath.Join(dir, "internal", "interface", "http", "admin", "system", "apis", "handler_openapi.go")
	b, err := os.ReadFile(extraHandler)
	if err != nil {
		t.Fatalf("handler_openapi.go: %v", err)
	}
	if !strings.Contains(string(b), "BatchSysApis") {
		t.Fatalf("expected BatchSysApis handler, got:\n%s", b)
	}
}
