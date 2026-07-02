package seq

import (
	"math-ai.com/math-ai/internal/domain/shared/mtime"
)

// Sequence is one row of ma_seqs — the central registry that drives
// per-entity ID generation. CurrentValue advances on every Next call;
// Prefix + Padding shape the rendered ID (e.g. "CH" + 8 → "CH00000001").
type Sequence struct {
	name         string
	currentValue uint64
	prefix       string
	padding      uint8
	incrementBy  uint32
	modifyDt     mtime.MathTime
}

func NewSequence() *Sequence { return &Sequence{} }

func (s *Sequence) Name() string                 { return s.name }
func (s *Sequence) SetName(n string)             { s.name = n }
func (s *Sequence) CurrentValue() uint64         { return s.currentValue }
func (s *Sequence) SetCurrentValue(v uint64)     { s.currentValue = v }
func (s *Sequence) Prefix() string               { return s.prefix }
func (s *Sequence) SetPrefix(p string)           { s.prefix = p }
func (s *Sequence) Padding() uint8               { return s.padding }
func (s *Sequence) SetPadding(p uint8)           { s.padding = p }
func (s *Sequence) IncrementBy() uint32          { return s.incrementBy }
func (s *Sequence) SetIncrementBy(i uint32)      { s.incrementBy = i }
func (s *Sequence) ModifyDt() mtime.MathTime     { return s.modifyDt }
func (s *Sequence) SetModifyDt(t mtime.MathTime) { s.modifyDt = t }
