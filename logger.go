package main

import (
	"context"
	"fmt"
	"log"
	"os"
)

type (
	loggerOutKey struct{}
	loggerErrKey struct{}
)

var (
	verbose = true
	debug   = false
)

func infof(ctx context.Context, format string, args ...any) {
	if verbose {
		ctx.Value(loggerOutKey{}).(*log.Logger).Output(2, fmt.Sprintf(format, args...))
	}
}

func debugf(ctx context.Context, format string, args ...any) {
	if debug {
		ctx.Value(loggerOutKey{}).(*log.Logger).Output(2, fmt.Sprintf("[DBG] "+format, args...))
	}
}

func warnf(ctx context.Context, format string, args ...any) {
	ctx.Value(loggerErrKey{}).(*log.Logger).Output(2, fmt.Sprintf(format, args...))
}

func fatal(ctx context.Context, err error) {
	ctx.Value(loggerErrKey{}).(*log.Logger).Output(2, err.Error())
	os.Exit(1)
}
