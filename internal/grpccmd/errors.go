package grpccmd

import "errors"

func errServiceOrTargetRequired() error {
	return errors.New("one of --service or --target is required")
}
