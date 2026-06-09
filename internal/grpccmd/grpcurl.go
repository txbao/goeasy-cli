package grpccmd

import (
	"fmt"
	"os"
	"os/exec"
)

func runGRPCURL(plaintext bool, target string, args ...string) error {
	if _, err := exec.LookPath("grpcurl"); err != nil {
		return fmt.Errorf("grpcurl not found in PATH: install https://github.com/fullstorydev/grpcurl/releases — server must enable reflection (goeasy grpcx default)")
	}
	if target == "" {
		return fmt.Errorf("grpc target is empty")
	}
	cmdArgs := make([]string, 0, 8)
	if plaintext {
		cmdArgs = append(cmdArgs, "-plaintext")
	}
	// grpcurl: [flags] [address] [list|describe|call ...]
	// call: grpcurl -plaintext -d '{}' host:port package.Service/Method
	if len(args) >= 2 && args[0] == "-d" {
		cmdArgs = append(cmdArgs, "-d", args[1])
		cmdArgs = append(cmdArgs, target)
		if len(args) > 2 {
			cmdArgs = append(cmdArgs, args[2:]...)
		}
	} else {
		cmdArgs = append(cmdArgs, target)
		cmdArgs = append(cmdArgs, args...)
	}
	cmd := exec.Command("grpcurl", cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("grpcurl: %w", err)
	}
	return nil
}
