package classroom

import (
	mtime "math-ai.com/math-ai/internal/domain/shared/time"
)

// ClassroomProgram models a row in ma_classroom_programs — the junction
// linking a classroom to a curriculum book/program. Many-to-many: one
// classroom can carry several programs, and a program can be present in
// many classrooms.
//
// Unlike ma_classroom_members, the edge has no business state worth
// keeping after removal, so the create/remove pair is plain
// INSERT / DELETE (not soft-delete-then-reactivate). The dual-status
// columns are present for shape consistency with the rest of the schema
// but the repository only emits ACTIVE rows.
type ClassroomProgram struct {
	id                 int64
	classroomProgramId string
	classroomId        string
	programId          string
	note               *string
	status             string
	createId           *string
	createDt           mtime.MathTime
	modifyId           *string
	modifyDt           mtime.MathTime
}

func NewClassroomProgram() *ClassroomProgram {
	return &ClassroomProgram{}
}

func (cp *ClassroomProgram) Id() int64                       { return cp.id }
func (cp *ClassroomProgram) SetId(v int64)                   { cp.id = v }
func (cp *ClassroomProgram) ClassroomProgramId() string      { return cp.classroomProgramId }
func (cp *ClassroomProgram) SetClassroomProgramId(v string)  { cp.classroomProgramId = v }
func (cp *ClassroomProgram) ClassroomId() string             { return cp.classroomId }
func (cp *ClassroomProgram) SetClassroomId(v string)         { cp.classroomId = v }
func (cp *ClassroomProgram) ProgramId() string               { return cp.programId }
func (cp *ClassroomProgram) SetProgramId(v string)           { cp.programId = v }
func (cp *ClassroomProgram) Note() *string                   { return cp.note }
func (cp *ClassroomProgram) SetNote(v *string)               { cp.note = v }
func (cp *ClassroomProgram) Status() string                  { return cp.status }
func (cp *ClassroomProgram) SetStatus(v string)              { cp.status = v }
func (cp *ClassroomProgram) CreateId() *string               { return cp.createId }
func (cp *ClassroomProgram) SetCreateId(v *string)           { cp.createId = v }
func (cp *ClassroomProgram) CreateDt() mtime.MathTime        { return cp.createDt }
func (cp *ClassroomProgram) SetCreateDt(t mtime.MathTime)    { cp.createDt = t }
func (cp *ClassroomProgram) ModifyId() *string               { return cp.modifyId }
func (cp *ClassroomProgram) SetModifyId(v *string)           { cp.modifyId = v }
func (cp *ClassroomProgram) ModifyDt() mtime.MathTime        { return cp.modifyDt }
func (cp *ClassroomProgram) SetModifyDt(t mtime.MathTime)    { cp.modifyDt = t }
