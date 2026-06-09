package generator

import (
	"fmt"
	"os"
)

// maybeRunGenProtoAfterDB 在 add db proto 后尝试 protoc；失败时提示手动 gen proto。
func maybeRunGenProtoAfterDB(projectDir string, protoRels []string, skip bool) {
	for _, rel := range protoRels {
		if skip {
			fmt.Fprintf(os.Stderr, "info: next: goeasy-cli gen proto --file %s\n", rel)
			continue
		}
		if err := GenerateProtoGo(GenProtoOptions{ProjectDir: projectDir, ProtoFile: rel}); err != nil {
			fmt.Fprintf(os.Stderr, "warn: auto gen proto for %s: %v\n", rel, err)
			fmt.Fprintf(os.Stderr, "info: run manually: goeasy-cli gen proto --file %s\n", rel)
		}
	}
}
