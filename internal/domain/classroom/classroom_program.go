package classroom

import (
	"math-ai.com/math-ai/internal/domain/shared/mtime"
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
	classroomProgramId int64
	classroomId        int64
	programId          int64
	note               *string
	status             string
	createId           *int64
	createDt           mtime.MathTime
	modifyId           *int64
	modifyDt           mtime.MathTime
}

func NewClassroomProgram() *ClassroomProgram {
	return &ClassroomProgram{}
}

func (cp *ClassroomProgram) Id() int64                     { return cp.id }
func (cp *ClassroomProgram) SetId(v int64)                 { cp.id = v }
func (cp *ClassroomProgram) ClassroomProgramId() int64     { return cp.classroomProgramId }
func (cp *ClassroomProgram) SetClassroomProgramId(v int64) { cp.classroomProgramId = v }
func (cp *ClassroomProgram) ClassroomId() int64            { return cp.classroomId }
func (cp *ClassroomProgram) SetClassroomId(v int64)        { cp.classroomId = v }
func (cp *ClassroomProgram) ProgramId() int64              { return cp.programId }
func (cp *ClassroomProgram) SetProgramId(v int64)          { cp.programId = v }
func (cp *ClassroomProgram) Note() *string                 { return cp.note }
func (cp *ClassroomProgram) SetNote(v *string)             { cp.note = v }
func (cp *ClassroomProgram) Status() string                { return cp.status }
func (cp *ClassroomProgram) SetStatus(v string)            { cp.status = v }
func (cp *ClassroomProgram) CreateId() *int64              { return cp.createId }
func (cp *ClassroomProgram) SetCreateId(v *int64)          { cp.createId = v }
func (cp *ClassroomProgram) CreateDt() mtime.MathTime      { return cp.createDt }
func (cp *ClassroomProgram) SetCreateDt(t mtime.MathTime)  { cp.createDt = t }
func (cp *ClassroomProgram) ModifyId() *int64              { return cp.modifyId }
func (cp *ClassroomProgram) SetModifyId(v *int64)          { cp.modifyId = v }
func (cp *ClassroomProgram) ModifyDt() mtime.MathTime      { return cp.modifyDt }
func (cp *ClassroomProgram) SetModifyDt(t mtime.MathTime)  { cp.modifyDt = t }
