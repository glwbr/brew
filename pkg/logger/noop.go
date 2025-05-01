package logger

import "context"

// NoOp is a no-operation logger that implements the Logger interface but silently discards all log entries.
// All methods are no-ops and return the same logger, ignoring all input.
// Useful for testing or disabling logging.
type NoOp struct{}

// Debug is a no-op.
func (n NoOp) Debug(_ string, _ ...any) {}

// Info is a no-op.
func (n NoOp) Info(_ string, _ ...any) {}

// Warn is a no-op.
func (n NoOp) Warn(_ string, _ ...any) {}

// Error is a no-op.
func (n NoOp) Error(_ string, _ ...any) {}

// WithContext returns the same NoOp logger (ignores context).
func (n NoOp) WithContext(_ context.Context) Logger { return n }

// WithField returns the same NoOp logger (ignores field).
func (n NoOp) WithField(_ string, _ any) Logger { return n }

// WithFields returns the same NoOp logger (ignores fields).
func (n NoOp) WithFields(_ map[string]any) Logger { return n }
