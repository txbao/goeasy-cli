package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeTestConfig(t *testing.T, dir, content string) string {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(dir, "configs"), 0755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "configs", "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func validDBRedisConfig() string {
	return `database:
  enabled: true
  driver: postgres
  dsn: "postgres://user:pass@127.0.0.1:5432/demo?sslmode=disable"
redis:
  enabled: true
  addr: "127.0.0.1:6379"
`
}

func TestValidateProjectDBRedisForCodegen(t *testing.T) {
	dir := t.TempDir()
	path := writeTestConfig(t, dir, validDBRedisConfig())
	if err := validateProjectDBRedisForCodegen(dir, path); err != nil {
		t.Fatalf("expected valid config: %v", err)
	}

	cases := []struct {
		name    string
		cfg     string
		wantErr string
	}{
		{
			name: "db disabled",
			cfg: `database:
  enabled: false
  dsn: "postgres://localhost/db"
redis:
  enabled: true
  addr: "127.0.0.1:6379"
`,
			wantErr: "database.enabled is false",
		},
		{
			name: "empty dsn",
			cfg: `database:
  enabled: true
  dsn: ""
redis:
  enabled: true
  addr: "127.0.0.1:6379"
`,
			wantErr: "database.dsn is empty",
		},
		{
			name: "redis disabled",
			cfg: `database:
  enabled: true
  dsn: "postgres://localhost/db"
redis:
  enabled: false
  addr: "127.0.0.1:6379"
`,
			wantErr: "redis.enabled is false",
		},
		{
			name: "empty redis addr",
			cfg: `database:
  enabled: true
  dsn: "postgres://localhost/db"
redis:
  enabled: true
  addr: ""
`,
			wantErr: "redis.addr is empty",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := t.TempDir()
			p := writeTestConfig(t, d, tc.cfg)
			err := validateProjectDBRedisForCodegen(d, p)
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("got %v, want contains %q", err, tc.wantErr)
			}
		})
	}
}

func TestGenerateCRUDWithoutDatabaseConfig(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/demo\n"), 0644); err != nil {
		t.Fatal(err)
	}
	cfg := `codegen:
  layout: domain
  domains:
    iam:
      modules:
        sys_apis:
          resource: apis
`
	writeTestConfig(t, dir, cfg)
	if err := GenerateCRUD(ModuleOptions{
		ProjectDir: dir,
		ModuleName: "sys_apis",
		Domain:     "iam",
		Resource:   "apis",
		ConfigPath: filepath.Join(dir, "configs", "config.yaml"),
		Force:      true,
	}); err != nil {
		t.Fatal(err)
	}
	meta := metaForTest("sys_apis", "iam", "apis")
	regPath := filepath.Join(dir, filepath.Join("internal", "bootstrap", "register_iam.go"))
	reg, err := os.ReadFile(regPath)
	if err != nil {
		t.Fatal(err)
	}
	pgPath := filepath.Join(dir, persistenceRepoRel(meta, "repository_pg.go"))
	pg, err := os.ReadFile(pgPath)
	if err != nil {
		t.Fatal(err)
	}
	rs, ps := string(reg), string(pg)
	if !strings.Contains(rs, "NewPGRepository(sqlxDB, infra.DBDriver, table, infra.Cache") {
		t.Fatal("register must use cache-aware NewPGRepository")
	}
	if !strings.Contains(ps, "func NewPGRepository(sqlxDB *sqlx.DB, driver, table string, c zcache.Cache") {
		t.Fatal("repository_pg must expose cache-aware NewPGRepository")
	}
}
