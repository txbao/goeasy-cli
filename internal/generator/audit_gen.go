package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// moduleHasAudit 检测 application 是否已接入 audit.Recorder。
func moduleHasAudit(projectDir string, meta ModuleMeta) bool {
	b, err := os.ReadFile(filepath.Join(projectDir, meta.appRel("application.go")))
	if err != nil {
		return false
	}
	s := string(b)
	return strings.Contains(s, "audit.Recorder") || strings.Contains(s, "recorder audit.Recorder")
}

func resolveModuleAudit(opts ModuleOptions, meta ModuleMeta) bool {
	return opts.WithAudit || moduleHasAudit(opts.ProjectDir, meta)
}

func auditServiceImports(goeasy string, withAudit bool) string {
	if !withAudit {
		return ""
	}
	return fmt.Sprintf("\t\"github.com/txbao/goeasy/audit\"\n\t\"github.com/txbao/goeasy/contextx\"\n")
}

func auditServiceStructField(withAudit bool) string {
	if !withAudit {
		return ""
	}
	return "\trecorder audit.Recorder\n"
}

func auditServiceCtorParams(withAudit bool) string {
	if !withAudit {
		return ""
	}
	return ", recorder audit.Recorder"
}

func auditServiceCtorBody(withAudit bool) string {
	if !withAudit {
		return ""
	}
	return "\tif recorder == nil {\n\t\trecorder = audit.NopRecorder{}\n\t}\n"
}

func auditServiceCtorAssign(withAudit bool) string {
	if !withAudit {
		return ""
	}
	return ", recorder: recorder"
}

func auditRecordStub(moduleCode, actionType, objectType, idExpr string) string {
	return fmt.Sprintf(`
	// _ = a.recorder.Record(ctx, contextx.OperatorFrom(ctx), audit.Entry{
	// 	ModuleCode: %q, ActionType: %q, ObjectType: %q,
	// 	ObjectID: %s, Result: "1",
	// })`, moduleCode, actionType, objectType, idExpr)
}

func auditCommandHandlerField(withAudit bool) string {
	if !withAudit {
		return ""
	}
	return "\trecorder audit.Recorder\n"
}

func auditCommandCtorParams(withAudit bool) string {
	if !withAudit {
		return ""
	}
	return ", recorder audit.Recorder"
}

func auditCommandCtorBody(withAudit bool) string {
	if !withAudit {
		return ""
	}
	return "\tif recorder == nil {\n\t\trecorder = audit.NopRecorder{}\n\t}\n"
}

func auditCommandCtorAssign(withAudit bool) string {
	if !withAudit {
		return ""
	}
	return ", recorder: recorder"
}

func auditCommandImports(goeasy string, withAudit bool) string {
	if !withAudit {
		return ""
	}
	return fmt.Sprintf("\t\"github.com/txbao/goeasy/audit\"\n\t\"github.com/txbao/goeasy/contextx\"\n")
}

func auditCommandRecordStub(moduleCode, actionType, objectType, idExpr string) string {
	return fmt.Sprintf(`
	// _ = h.recorder.Record(ctx, contextx.OperatorFrom(ctx), audit.Entry{
	// 	ModuleCode: %q, ActionType: %q, ObjectType: %q,
	// 	ObjectID: %s, Result: "1",
	// })`, moduleCode, actionType, objectType, idExpr)
}
