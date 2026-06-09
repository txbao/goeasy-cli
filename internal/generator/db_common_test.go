package generator

import (
	"strings"
	"testing"

	"github.com/txbao/goeasy-cli/internal/schema"
)

func TestAssertTableReservedMigrations(t *testing.T) {
	err := assertTableAllowed("_sqlx_migrations", nil)
	if err == nil || !strings.Contains(err.Error(), "reserved") {
		t.Fatalf("expected reserved error, got %v", err)
	}
}

func TestResolveModuleNameOverride(t *testing.T) {
	opts := DBOptions{ModuleName: "user_roles"}
	got := resolveModuleName(opts, "sys_user_roles", "sys_")
	if got != "user_roles" {
		t.Fatalf("got %q", got)
	}
}

func TestResolveModuleNameFromPhysical(t *testing.T) {
	opts := DBOptions{}
	got := resolveModuleName(opts, "ge_sys_roles", "ge_")
	if got != "sys_roles" {
		t.Fatalf("got %q", got)
	}
}

func TestEffectiveMySQLDatabase(t *testing.T) {
	dsn := "user:pass@tcp(127.0.0.1:3306)/mydb?parseTime=true"
	if got := effectiveMySQLDatabase("public", dsn); got != "mydb" {
		t.Fatalf("got %q", got)
	}
	if got := effectiveMySQLDatabase("custom", dsn); got != "custom" {
		t.Fatalf("got %q", got)
	}
}

func TestGenEntityFromFixture(t *testing.T) {
	ct := schema.Classify(schema.TableMeta{
		Name: "items",
		Columns: []schema.ColumnMeta{
			{Name: "id", DBType: "bigint", IsPrimaryKey: true, Ordinal: 1},
			{Name: "name", DBType: "varchar", Ordinal: 2},
			{Name: "created_at", DBType: "timestamp with time zone", Ordinal: 3},
		},
	}, "items", "items", schema.DefaultCodegenRules())
	meta := metaForTest("items", "items", "items")
	out := genEntity("github.com/demo/app", ct, "Items", meta)
	if !strings.Contains(out, "package items") {
		t.Fatal("missing package")
	}
	if !strings.Contains(out, "import (\n\t\"errors\"\n\t\"time\"\n") {
		t.Fatal("expected merged imports with time")
	}
	if strings.Count(out, "import (") != 1 {
		t.Fatal("expected single import block")
	}
	if !strings.Contains(out, "func RehydrateItems(") {
		t.Fatal("missing Rehydrate")
	}
}

func TestGenHandlerCrudListPagination(t *testing.T) {
	ct := schema.Classify(schema.TableMeta{
		Name: "items",
		Columns: []schema.ColumnMeta{
			{Name: "id", DBType: "bigint", IsPrimaryKey: true, Ordinal: 1},
			{Name: "name", DBType: "varchar", Ordinal: 2},
		},
	}, "items", "items", schema.DefaultCodegenRules())
	meta := metaForTest("items", "items", "items")
	out := genHandlerCrud("github.com/demo/app", "admin", ct, "Items", "github.com/txbao/goeasy", true, meta, AppStyleService)
	if !strings.Contains(out, "ShouldBindJSON") {
		t.Fatal("expected JSON page bind")
	}
	if !strings.Contains(out, "pagination") {
		t.Fatal("expected pagination meta")
	}
	if !strings.Contains(out, "ToResponseDTO") || !strings.Contains(out, "\"list\": list") {
		t.Fatal("expected client ResponseDTO list mapping")
	}
}

func TestGenRepositoryIfaceList(t *testing.T) {
	out := genRepositoryIface("items")
	if !strings.Contains(out, "List(ctx context.Context, page, pageSize int)") {
		t.Fatal("missing List on Repository")
	}
}

func TestGenListQueryNoDuplicateHandler(t *testing.T) {
	ct := schema.Classify(schema.TableMeta{Name: "items", Columns: []schema.ColumnMeta{
		{Name: "id", DBType: "bigint", IsPrimaryKey: true},
	}}, "items", "items", schema.DefaultCodegenRules())
	meta := metaForTest("items", "items", "items")
	out := genListQuery("github.com/demo/app", ct, "Items", meta)
	if strings.Contains(out, "type Handler struct") {
		t.Fatal("list.go must not redefine Handler")
	}
	if strings.Contains(out, "internal/app/items\"") {
		t.Fatal("query/list must not import parent app package")
	}
	if !strings.Contains(out, "[]*domain.Aggregate") {
		t.Fatal("expected domain aggregates return")
	}
}

func TestGenEntityJSONTags(t *testing.T) {
	ct := rolesLikeClassified()
	meta := metaForTest("sys_roles", "system", "roles")
	out := genEntity("github.com/demo/app", ct, "SysRoles", meta)
	if !strings.Contains(out, "`json:\"name\"`") {
		t.Fatal("expected json tag on Name")
	}
	if !strings.Contains(out, "`json:\"deleted_at,omitempty\"`") {
		t.Fatal("expected omitempty on nullable deleted_at")
	}
}

func TestGenEntityExportedFields(t *testing.T) {
	ct := rolesLikeClassified()
	meta := metaForTest("sys_roles", "system", "roles")
	out := genEntity("github.com/demo/app", ct, "SysRoles", meta)
	if strings.Contains(out, "\tname string") {
		t.Fatal("expected exported Name field")
	}
	if !strings.Contains(out, "\tName string") {
		t.Fatal("missing exported Name")
	}
	if !strings.Contains(out, "\tID int64") {
		t.Fatal("expected ID field for id column")
	}
}

func TestGenEntityNoGetterMethods(t *testing.T) {
	ct := rolesLikeClassified()
	meta := metaForTest("sys_roles", "system", "roles")
	out := genEntity("github.com/demo/app", ct, "SysRoles", meta)
	if strings.Contains(out, "func (e *SysRoles) Name()") || strings.Contains(out, "func (e *SysRoles) ID()") {
		t.Fatal("exported fields must not have same-name getters")
	}
}

func TestGenDTOUsesFieldAccess(t *testing.T) {
	ct := rolesLikeClassified()
	meta := metaForTest("sys_roles", "system", "roles")
	out := genDTO("github.com/demo/app", ct, "SysRoles", meta)
	if strings.Contains(out, "r.Name()") {
		t.Fatal("ToDTO should use field access r.Name")
	}
	if !strings.Contains(out, "r.Name,") {
		t.Fatal("expected r.Name in ToDTO")
	}
}

func TestGenRepositoryPGUsesFieldAccess(t *testing.T) {
	ct := rolesLikeClassified()
	meta := metaForTest("sys_roles", "system", "roles")
	out := genRepositoryPG("github.com/demo/app", "github.com/txbao/goeasy", ct, "SysRoles", "sysRoles", meta)
	if strings.Contains(out, "root.Name()") {
		t.Fatal("repository should use root.Name field access")
	}
	if !strings.Contains(out, "root.Name,") {
		t.Fatal("expected root.Name in repository_pg")
	}
}

func TestGenAppList(t *testing.T) {
	meta := metaForTest("sys_roles", "system", "roles")
	out := genAppList(meta, "SysRoles")
	if !strings.Contains(out, "func (a *Application) List") {
		t.Fatal("missing Application.List")
	}
	if !strings.Contains(out, "ToDTO(agg)") {
		t.Fatal("expected ToDTO mapping")
	}
}

func TestGenHandlerUsesValidator(t *testing.T) {
	ct := rolesLikeClassified()
	meta := metaForTest("sys_roles", "system", "roles")
	out := genHandler("github.com/demo/app", "admin", ct, "SysRoles", "sysRoles", "github.com/txbao/goeasy", true, meta, AppStyleLightCQRS)
	if !strings.Contains(out, "zvalid.Validate(&cmd)") {
		t.Fatal("expected validator on Create")
	}
}

func TestGenHandlerCrudUsesAppList(t *testing.T) {
	ct := rolesLikeClassified()
	meta := metaForTest("sys_roles", "system", "roles")
	out := genHandlerCrud("github.com/demo/app", "admin", ct, "SysRoles", "github.com/txbao/goeasy", true, meta, AppStyleLightCQRS)
	if strings.Contains(out, "Queries().List") {
		t.Fatal("should use Application.List to avoid import cycle")
	}
	if !strings.Contains(out, "h.app.List") {
		t.Fatal("expected h.app.List")
	}
}

func rolesLikeClassified() schema.ClassifiedTable {
	return schema.Classify(schema.TableMeta{
		Name: "sys_roles",
		Columns: []schema.ColumnMeta{
			{Name: "id", DBType: "bigint", IsPrimaryKey: true, Ordinal: 1},
			{Name: "created_at", DBType: "timestamptz", Ordinal: 2},
			{Name: "updated_at", DBType: "timestamptz", Ordinal: 3},
			{Name: "deleted_at", DBType: "timestamptz", Nullable: true, Ordinal: 4},
			{Name: "name", DBType: "varchar", Ordinal: 5},
			{Name: "code", DBType: "varchar", Ordinal: 6},
			{Name: "status", DBType: "smallint", Ordinal: 7},
		},
	}, "sys_roles", "sys_roles", schema.DefaultCodegenRules())
}

func TestGenDTOImportsTime(t *testing.T) {
	ct := rolesLikeClassified()
	meta := metaForTest("sys_roles", "system", "roles")
	out := genDTO("github.com/demo/app", ct, "SysRoles", meta)
	if !strings.Contains(out, "\"time\"") {
		t.Fatal("expected time import in dto")
	}
}

func TestGenRepositoryPGImportsTime(t *testing.T) {
	ct := rolesLikeClassified()
	meta := metaForTest("sys_roles", "system", "roles")
	out := genRepositoryPG("github.com/demo/app", "github.com/txbao/goeasy", ct, "SysRoles", "sysRoles", meta)
	if !strings.Contains(out, "zcachekey.GetEntityBytes") {
		t.Fatal("expected entity cache in FindByID")
	}
	if !strings.Contains(out, "invalidateEntityCache") {
		t.Fatal("expected cache invalidation on Update/Delete")
	}
	if !strings.Contains(out, "\"time\"") {
		t.Fatal("expected time import in repository_pg")
	}
}

func TestGenEntityNoDuplicateSetter(t *testing.T) {
	ct := rolesLikeClassified()
	meta := metaForTest("sys_roles", "system", "roles")
	out := genEntity("github.com/demo/app", ct, "SysRoles", meta)
	if strings.Count(out, "func (e *SysRoles) SetName") != 1 {
		t.Fatalf("SetName should appear once, count=%d", strings.Count(out, "func (e *SysRoles) SetName"))
	}
}

func TestGenHandlerBindsCreateCommand(t *testing.T) {
	ct := rolesLikeClassified()
	meta := metaForTest("sys_roles", "system", "roles")
	out := genHandler("github.com/demo/app", "admin", ct, "SysRoles", "sysRoles", "github.com/txbao/goeasy", true, meta, AppStyleLightCQRS)
	if strings.Contains(out, "var body struct") {
		t.Fatal("should not use anonymous body struct")
	}
	if !strings.Contains(out, "var cmd command.CreateCommand") {
		t.Fatal("expected bind to CreateCommand")
	}
}

func TestGenHandlerCrudBindsUpdateCommand(t *testing.T) {
	ct := rolesLikeClassified()
	meta := metaForTest("sys_roles", "system", "roles")
	out := genHandlerCrud("github.com/demo/app", "admin", ct, "SysRoles", "github.com/txbao/goeasy", true, meta, AppStyleLightCQRS)
	if strings.Contains(out, "var body struct") {
		t.Fatal("should not use anonymous body struct")
	}
	if !strings.Contains(out, "var cmd command.UpdateCommand") {
		t.Fatal("expected bind to UpdateCommand")
	}
	if !strings.Contains(out, "cmd.ID = id") {
		t.Fatal("expected path id assignment")
	}
}

func TestGenCreateCommandJSONTags(t *testing.T) {
	ct := rolesLikeClassified()
	meta := metaForTest("sys_roles", "system", "roles")
	out := genCreateCommand("github.com/demo/app", ct, "SysRoles", "sysRoles", meta)
	if !strings.Contains(out, `json:"name"`) || !strings.Contains(out, "validate:") {
		t.Fatal("expected json and validate tags on CreateCommand")
	}
}
