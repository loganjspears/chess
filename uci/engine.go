package uci

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"sync"
)

// Engine represents a UCI compliant chess engine (e.g. Stockfish, Shredder, etc.).
// Engine is safe for concurrent use.
type Engine struct {
	cmd     *exec.Cmd
	in      *io.PipeWriter
	out     *io.PipeReader
	debug   bool
	logger  *log.Logger
	id      map[string]string
	options map[string]Option
	results SearchResults
	mu      *sync.RWMutex
}

// Debug is an option for the New function to add logging for debugging.  This will
// log all output to and from the chess engine.
func Debug(e *Engine) {
	e.debug = true
}

// Logger is an option for the New function to customize the logger.  The logger is
// only used if the Debug option is also used.
func Logger(logger *log.Logger) func(e *Engine) {
	return func(e *Engine) {
		e.logger = logger
	}
}

// New constructs an engine from the executable path (found using exec.LookPath).
// New also starts running the executable process in the background.  Once created
// the Engine can be controlled via the Run method.
func New(path string, opts ...func(e *Engine)) (*Engine, error) {
	path, err := exec.LookPath(path)
	if err != nil {
		return nil, fmt.Errorf("uci: executable not found at path %s %w", path, err)
	}
	rIn, wIn := io.Pipe()
	rOut, wOut := io.Pipe()
	cmd := exec.Command(path)
	cmd.Stdin = rIn
	cmd.Stdout = wOut
	e := &Engine{cmd: cmd, in: wIn, out: rOut, mu: &sync.RWMutex{}, logger: log.New(os.Stdout, "uci", log.LstdFlags)}
	for _, opt := range opts {
		opt(e)
	}
	err = e.cmd.Start()
	if err != nil {
		return nil, fmt.Errorf("uci: failed to start executable %s: %w", path, err)
	}
	go e.cmd.Wait()
	return e, nil
}

func (e *Engine) Renice() error {
	cmd := exec.Command("renice", "-n", "19", "-p", strconv.FormatInt(int64(e.cmd.Process.Pid), 10))

	return cmd.Run()
}

// ID returns the id values returned from the most recent CmdUCI invocation.  It includes
// key value data such as the following:
// id name Stockfish 12
// id author the Stockfish developers (see AUTHORS file)
func (e *Engine) ID() map[string]string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	cp := map[string]string{}
	for k, v := range e.id {
		cp[k] = v
	}
	return cp
}

// Options returns exposed options from the most recent CmdUCI invocation.  It includes
// data such as the following:
// option name Debug Log File type string default
// option name Contempt type spin default 24 min -100 max 100
// option name Analysis Contempt type combo default Both var Off var White var Black var Both
// option name Threads type spin default 1 min 1 max 512
// option name Hash type spin default 16 min 1 max 33554432
// option name Clear Hash type button
// option name Ponder type check default false
// option name MultiPV type spin default 1 min 1 max 500
// option name Skill Level type spin default 20 min 0 max 20
// option name Move Overhead type spin default 10 min 0 max 5000
// option name Slow Mover type spin default 100 min 10 max 1000
// option name nodestime type spin default 0 min 0 max 10000
// option name UCI_Chess960 type check default false
// option name UCI_AnalyseMode type check default false
// option name UCI_LimitStrength type check default false
// option name UCI_Elo type spin default 1350 min 1350 max 2850
// option name UCI_ShowWDL type check default false
// option name SyzygyPath type string default <empty>
// option name SyzygyProbeDepth type spin default 1 min 1 max 100
// option name Syzygy50MoveRule type check default true
// option name SyzygyProbeLimit type spin default 7 min 0 max 7
// option name Use NNUE type check default true
// option name EvalFile type string default nn-82215d0fd0df.nnue
func (e *Engine) Options() map[string]Option {
	e.mu.RLock()
	defer e.mu.RUnlock()

	cp := map[string]Option{}
	for k, v := range e.options {
		cp[k] = v
	}
	return cp
}

// SearchResults returns results from the most recent CmdGo invocation.  It includes
// data such as the following:
// info depth 21 seldepth 31 multipv 1 score cp 39 nodes 862438 nps 860716 hashfull 409 tbhits 0 time 1002 pv e2e4
// bestmove e2e4 ponder c7c5
func (e *Engine) SearchResults() SearchResults {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.results
}

// Run runs the set of Cmds in the order given and returns an error if
// any of the commands fails.  Except for CmdStop (usually paired with
// CmdGo's infinite option) all commands block via mutux until completed.
func (e *Engine) Run(cmds ...Cmd) error {
	for _, cmd := range cmds {
		if cmd.String() == CmdStop.Name {
			if err := e.processCommand(cmd); err != nil {
				return err
			}
		} else {
			if err := e.processCommandLocked(cmd); err != nil {
				return err
			}
		}
	}
	return nil
}

// Close releases readers, writers, and processes associated with the
// Engine.  It also invokes the CmdQuit to signal the engine to terminate.
func (e *Engine) Close() error {
	if err := e.Run(CmdQuit); err != nil {
		return err
	}
	if err := e.in.Close(); err != nil {
		return err
	}
	if err := e.out.Close(); err != nil {
		return err
	}
	return e.cmd.Process.Kill()
}

func (e *Engine) processCommandLocked(cmd Cmd) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.processCommand(cmd)
}

func (e *Engine) processCommand(cmd Cmd) error {
	if e.debug {
		e.logger.Println(cmd.String())
	}
	if _, err := fmt.Fprintln(e.in, cmd.String()); err != nil {
		return err
	}
	if err := cmd.ProcessResponse(e); err != nil {
		return err
	}
	return nil
}

func (e *Engine) readLine(scanner *bufio.Scanner) string {
	s := scanner.Text()
	if e.debug {
		e.logger.Println(s)
	}
	return s
}
