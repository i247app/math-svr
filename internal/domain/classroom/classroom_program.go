package classroom

import (
	"math-ai.com/math-ai/internal/domain/shared/audit"
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
	audit.Audit

	classroomProgramId int64
	classroomId        int64
	programId          int64
}

func NewClassroomProgram() *ClassroomProgram {
	return &ClassroomProgram{}
}

func (cp *ClassroomProgram) ClassroomProgramId() int64     { return cp.classroomProgramId }
func (cp *ClassroomProgram) SetClassroomProgramId(v int64) { cp.classroomProgramId = v }

func (cp *ClassroomProgram) ClassroomId() int64     { return cp.classroomId }
func (cp *ClassroomProgram) SetClassroomId(v int64) { cp.classroomId = v }

func (cp *ClassroomProgram) ProgramId() int64     { return cp.programId }
func (cp *ClassroomProgram) SetProgramId(v int64) { cp.programId = v }
