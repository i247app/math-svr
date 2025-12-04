package db

import (
	"context"
	"fmt"
	"path"
	"runtime"
	"strings"

	"math-ai.com/math-ai/internal/shared/logger"
)

func (*Database) logInputSQL(ctx context.Context, query string, args ...any) {
	logger := logger.GetLogger(ctx)
	var tag string
	pc, file, line, ok := runtime.Caller(2) // Get information about the caller 3 frames up
	if ok {
		funcName := runtime.FuncForPC(pc).Name()
		// Extract just the function name from the full path
		if lastDot := strings.LastIndex(funcName, "."); lastDot != -1 {
			funcName = funcName[lastDot+1:]
		}

		// Extract package/directory name
		dir := path.Dir(file)
		if lastSlash := strings.LastIndex(dir, "/"); lastSlash != -1 {
			dir = dir[lastSlash+1:]
		}

		// Extract filename
		fileName := path.Base(file)

		tag = fmt.Sprintf("# %s/%s %s:%d", dir, fileName, funcName, line) // Updated log format
	}
	logger.Infof("# SQL START: %s\n", tag)

	msg := fmt.Sprintf("%s, args: %v", query, args)
	logger.Infof("%s\n", msg)

	logger.Info("# SQL END")
}
