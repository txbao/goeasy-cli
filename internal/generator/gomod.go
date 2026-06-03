package generator

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func initGoModule(projectDir, modulePath, goEasyModule, replacePath string) error {
	gomod := filepath.Join(projectDir, "go.mod")
	_ = os.Remove(gomod)

	cmd := exec.Command("go", "mod", "init", modulePath)
	cmd.Dir = projectDir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("go mod init: %w: %s", err, out)
	}

	if replacePath != "" {
		replaceArg := goEasyModule + "=" + replacePath
		cmd := exec.Command("go", "mod", "edit", "-replace="+replaceArg)
		cmd.Dir = projectDir
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("go mod edit -replace: %w: %s", err, out)
		}
	}
	return nil
}
