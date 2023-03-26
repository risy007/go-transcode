package hls

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/m1k1o/go-transcode/internal/utils"
)

// how often should be cleanup called
const cleanupPeriod = 4 * time.Second

// timeout for first playlist, when it waits for new data
const playlistTimeout = 60 * time.Second

// minimum segments available to consider stream as active
const hlsMinimumSegments = 2

// how long must be active stream idle to be considered as dead
const activeIdleTimeout = 12 * time.Second

// how long must be iactive stream idle to be considered as dead
const inactiveIdleTimeout = 24 * time.Second

type ManagerCtx struct {
	logger     zerolog.Logger
	mu         sync.Mutex
	cmdFactory func() *exec.Cmd
	active     bool
	events     struct {
		onStart  func()
		onCmdLog func(message string)
		onStop   func(err error)
	}

	cmd         *exec.Cmd
	tempdir     string
	lastRequest time.Time

	shutdown chan interface{}
}

func New(cmdFactory func() *exec.Cmd) *ManagerCtx {
	return &ManagerCtx{
		logger:     log.With().Str("module", "hls").Str("submodule", "manager").Logger(),
		cmdFactory: cmdFactory,
		shutdown:   make(chan interface{}),
	}
}

func (m *ManagerCtx) SetRunPath(cmdPath string) error {
	err := os.MkdirAll(cmdPath, 0777)
	m.tempdir = cmdPath
	return err
}

func (m *ManagerCtx) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.cmd != nil {
		return errors.New("has already started")
	}

	m.logger.Debug().Msg("开始启动程序......")

	var err error

	m.cmd = m.cmdFactory()
	m.cmd.Dir = m.tempdir

	if m.events.onCmdLog != nil {
		m.cmd.Stderr = utils.LogEvent(m.events.onCmdLog)
	} else {
		m.cmd.Stderr = utils.LogWriter(m.logger)
	}

	_, write := io.Pipe()
	m.cmd.Stdout = write

	// create a new process group
	m.cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	m.active = false
	m.lastRequest = time.Now()
	m.shutdown = make(chan interface{})

	if m.events.onStart != nil {
		m.events.onStart()
	}

	// start program
	err = m.cmd.Start()

	// wait for program to exit
	go func() {
		err = m.cmd.Wait()
		if err != nil {
			if exiterr, ok := err.(*exec.ExitError); ok {
				// The program has exited with an exit code != 0

				// This works on both Unix and Windows. Although package
				// syscall is generally platform dependent, WaitStatus is
				// defined for both Unix and Windows and in both cases has
				// an ExitStatus() method with the same signature.
				if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
					m.logger.Warn().Int("exit-status", status.ExitStatus()).Msg("转码脚本执行时返回非0码!")
				}
			} else {
				m.logger.Err(err).Msg("转码脚本执行时错误退出了！")
			}
		} else {
			m.logger.Info().Msg("the program has successfully exited")
		}

		close(m.shutdown)

		if m.events.onStop != nil {
			m.events.onStop(err)
		}

		err := os.RemoveAll(m.tempdir)
		m.logger.Err(err).Msg("清理临时文件......")

		m.mu.Lock()
		m.cmd = nil
		m.mu.Unlock()
	}()

	return err
}

func (m *ManagerCtx) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.cmd != nil && m.cmd.Process != nil {
		m.logger.Debug().Msg("performing stop")

		pgid, err := syscall.Getpgid(m.cmd.Process.Pid)
		if err == nil {
			err := syscall.Kill(-pgid, syscall.SIGKILL)
			m.logger.Err(err).Msg("killing process group")
		} else {
			m.logger.Err(err).Msg("could not get process group id")
			err := m.cmd.Process.Kill()
			m.logger.Err(err).Msg("killing process")
		}
	}
}

func (m *ManagerCtx) Cleanup() {
	m.mu.Lock()
	diff := time.Since(m.lastRequest)
	stop := m.active && diff > activeIdleTimeout || !m.active && diff > inactiveIdleTimeout
	m.mu.Unlock()

	m.logger.Debug().
		Time("last_request", m.lastRequest).
		Dur("diff", diff).
		Bool("active", m.active).
		Bool("stop", stop).
		Msg("performing cleanup")

	if stop {
		m.Stop()
	}
}

func (m *ManagerCtx) OnStart(event func()) {
	m.events.onStart = event
}

func (m *ManagerCtx) OnCmdLog(event func(message string)) {
	m.events.onCmdLog = event
}

func (m *ManagerCtx) OnStop(event func(err error)) {
	m.events.onStop = event
}
