package tea

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/term"
	"github.com/muesli/cancelreader"
)

func (p *Program) suspend() {
	if err := p.ReleaseTerminal(); err != nil {
		// If we can't release input, abort.
		return
	}

	suspendProcess()

	_ = p.RestoreTerminal()
	go p.Send(ResumeMsg{})
}

func (p *Program) initTerminal() error {
	return p.initInput()
}

// restoreTerminalState restores the terminal to the state prior to running the
// Bubble Tea program.
func (p *Program) restoreTerminalState() error {
	if p.bpActive {
		p.execute(ansi.DisableBracketedPaste)
	}
	if p.renderer != nil {
		if p.renderer.Mode(hideCursor) {
			p.renderer.SetMode(hideCursor, false)
		}
	}

	if p.mouseEnabled {
		p.disableMouse()
	}
	if p.modifyOtherKeys != 0 {
		p.execute(ansi.DisableModifyOtherKeys)
	}
	if p.kittyFlags != 0 {
		p.execute(ansi.DisableKittyKeyboard)
	}
	if p.reportFocus {
		p.execute(ansi.DisableReportFocus)
	}
	if p.graphemeClustering {
		p.execute(ansi.DisableGraphemeClustering)
	}

	if p.renderer != nil {
		if p.renderer.Mode(altScreenMode) {
			p.renderer.SetMode(altScreenMode, false)

			// give the terminal a moment to catch up
			time.Sleep(time.Millisecond * 10) //nolint:gomnd
		}
	}

	return p.restoreInput()
}

// restoreInput restores the tty input to its original state.
func (p *Program) restoreInput() error {
	if p.ttyInput != nil && p.previousTtyInputState != nil {
		if err := term.Restore(p.ttyInput.Fd(), p.previousTtyInputState); err != nil {
			return fmt.Errorf("error restoring console: %w", err)
		}
	}
	if p.ttyOutput != nil && p.previousOutputState != nil {
		if err := term.Restore(p.ttyOutput.Fd(), p.previousOutputState); err != nil {
			return fmt.Errorf("error restoring console: %w", err)
		}
	}
	return nil
}

// initInputReader (re)commences reading inputs.
func (p *Program) initInputReader() error {
	var term string
	for i := len(p.environ) - 1; i >= 0; i-- {
		// We iterate backwards to find the last TERM variable set in the
		// environment. This is because the last one is the one that will be
		// used by the terminal.
		parts := strings.SplitN(p.environ[i], "=", 2)
		if len(parts) == 2 && parts[0] == "TERM" {
			term = parts[1]
			break
		}
	}

	// Initialize the input reader.
	// This need to be done after the terminal has been initialized and set to
	// raw mode.
	// On Windows, this will change the console mode to enable mouse and window
	// events.
	var flags int // TODO: make configurable through environment variables?
	drv, err := newDriver(p.input, term, flags)
	if err != nil {
		return err
	}

	p.inputReader = drv
	p.readLoopDone = make(chan struct{})
	go p.readLoop()

	return nil
}

func readInputs(ctx context.Context, msgs chan<- Msg, reader *driver) error {
	for {
		events, err := reader.ReadEvents()
		if err != nil {
			return err
		}

		for _, msg := range events {
			incomingMsgs := []Msg{msg}

			for _, m := range incomingMsgs {
				select {
				case msgs <- m:
				case <-ctx.Done():
					err := ctx.Err()
					if err != nil {
						err = fmt.Errorf("found context error while reading input: %w", err)
					}
					return err
				}
			}
		}
	}
}

func (p *Program) readLoop() {
	defer close(p.readLoopDone)

	err := readInputs(p.ctx, p.msgs, p.inputReader)
	if !errors.Is(err, io.EOF) && !errors.Is(err, cancelreader.ErrCanceled) {
		select {
		case <-p.ctx.Done():
		case p.errs <- err:
		}
	}
}

// waitForReadLoop waits for the cancelReader to finish its read loop.
func (p *Program) waitForReadLoop() {
	select {
	case <-p.readLoopDone:
	case <-time.After(500 * time.Millisecond): //nolint:gomnd
		// The read loop hangs, which means the input
		// cancelReader's cancel function has returned true even
		// though it was not able to cancel the read.
	}
}

// checkResize detects the current size of the output and informs the program
// via a WindowSizeMsg.
func (p *Program) checkResize() {
	if p.ttyOutput == nil {
		// can't query window size
		return
	}

	w, h, err := term.GetSize(p.ttyOutput.Fd())
	if err != nil {
		select {
		case <-p.ctx.Done():
		case p.errs <- err:
		}

		return
	}

	p.Send(WindowSizeMsg{
		Width:  w,
		Height: h,
	})
}
