package tools

import "fmt"

var ErrNotExist = fmt.Errorf("not exist")

var FailedToCreateTool = func(toolName string, err error) error { return fmt.Errorf("Failed to create %s tool: %w", err) }
