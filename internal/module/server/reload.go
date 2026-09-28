package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync/atomic"
	"time"

	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/config"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/shared/response"
)

// envCheckArg makes the binary validate the env file and exit instead of
// serving. HandleReload runs it in a child process before stopping anything.
const envCheckArg = "-check-env"

// envCheckTimeout bounds the pre-flight child; loading config is file I/O only.
const envCheckTimeout = 10 * time.Second

// baseEnviron is the environment the process was started with, captured
// before config.NewEnv's godotenv.Load copies the .env keys into it. Package
// variables initialise before main runs, so no .env value is in here.
//
// It is what a reload hands to the new process image. godotenv.Load never
// overrides a key that is already set, so passing os.Environ() instead would
// make the reloaded server keep every OLD .env value — the reload would
// silently change nothing.
var baseEnviron = os.Environ()

// stopping is set once a shutdown or reload has been accepted, so a second
// lifecycle request cannot race the first.
var stopping atomic.Bool

// reloadRequested tells main to re-exec itself once the graceful shutdown
// has finished.
var reloadRequested atomic.Bool

// ReloadRequested reports whether the server stopped because of
// POST /server/reload; main then calls Reexec instead of exiting.
func ReloadRequested() bool { return reloadRequested.Load() }

// IsEnvCheckRun reports whether this process was started as the reload
// pre-flight child.
func IsEnvCheckRun() bool {
	return len(os.Args) > 1 && os.Args[1] == envCheckArg
}

// RunEnvCheck loads envPath through the same loader the server boots with
// and returns the process exit code. The loader panics on a missing or
// malformed key, so the panic is the failure report.
func RunEnvCheck(envPath string) (code int) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintln(os.Stderr, r)
			code = 1
		}
	}()
	if _, err := config.NewEnv(envPath); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}

// HandleReload restarts the server in place so edited .env values take
// effect: it validates the env file, then runs the same graceful shutdown as
// HandleShutdown (sockets, jobs, session snapshot, resources), after which
// main replaces the process image with a fresh copy of the binary (same PID,
// so systemd sees no restart). Sessions survive only when
// SERIALIZED_SESSION_FILE is set — the new process reloads that snapshot.
//
// The env file is checked in a child process first because a server that
// stops and then fails to boot is an outage; a bad .env is refused here
// while the running server keeps serving.
func (h *Handler) HandleReload(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.From(ctx)

	if !reloadSupported {
		response.WriteJson(w, nil, errs.NewError(ctx, status.SERVER_RELOAD_NOT_SUPPORTED, nil,
			errors.New("reload is not supported on this OS")))
		return
	}
	if !stopping.CompareAndSwap(false, true) {
		response.WriteJson(w, nil, errs.NewError(ctx, status.SERVER_SHUTTING_DOWN, nil,
			errors.New("a shutdown or reload is already in progress")))
		return
	}

	log.Warnf("server.reload.requested uid=%d remote=%s", requesterUID(r), r.RemoteAddr)

	if err := checkEnv(ctx); err != nil {
		stopping.Store(false)
		log.Errorf("server.reload.env_invalid err=%v", err)
		// The loader's message stays in the server log: it can quote a line
		// of the env file.
		response.WriteJson(w, nil, errs.NewError(ctx, status.SERVER_RELOAD_INVALID_ENV, nil,
			errors.New("env validation failed, see server log")))
		return
	}

	reloadRequested.Store(true)
	response.WriteJsonNoContent(w, nil)

	time.AfterFunc(shutdownDelay, func() {
		if err := selfShutdown(); err != nil {
			logger.From(ctx).Errorf("server.reload.signal_failed err=%v", err)
		}
	})
}

// checkEnv runs this binary with envCheckArg under the start-up environment,
// so it validates the .env on disk exactly as the reloaded process will.
func checkEnv(ctx context.Context) error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolve executable: %w", err)
	}
	ctx, cancel := context.WithTimeout(ctx, envCheckTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, exe, envCheckArg)
	cmd.Env = baseEnviron
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}
