package classroom

import (
	mtime "math-ai.com/math-ai/internal/domain/shared/time"
)

// Classroom models ma_classrooms. owner_profile_id is denormalized for
// O(1) "who owns this" lookups; it is kept in sync with the matching
// OWNER row in ma_classroom_members inside the same UnitOfWork. The
// invite_code path is opt-in: nil/empty + a null expiry means joining
// is closed (only targeted invitations work).
type Classroom struct {
	id                   int64
	classroomId          string
	ownerProfileId       string
	name                 string
	description          *string
	schoolId             *string
	gradeId              *string
	inviteCode           *string
	inviteCodeExpiresDt  mtime.MathTime
	maxMembers           *int64
	memberCount          int64
	studentCount         int64
	teacherCount         int64
	coverKey             *string
	note                 *string
	classroomStatus      *string
	status               string
	createId             *string
	createDt             mtime.MathTime
	modifyId             *string
	modifyDt             mtime.MathTime

	// programIds is the in-memory denormalisation of the
	// ma_classroom_programs junction rows for this classroom. It is set
	// by the query/command layer after the classroom row is fetched;
	// no column on ma_classrooms backs it. A nil slice means "not
	// hydrated"; an empty non-nil slice means "hydrated, zero programs".
	programIds []string
}

func NewClassroom() *Classroom {
	return &Classroom{}
}

func (c *Classroom) Id() int64                          { return c.id }
func (c *Classroom) SetId(id int64)                     { c.id = id }
func (c *Classroom) ClassroomId() string                { return c.classroomId }
func (c *Classroom) SetClassroomId(id string)           { c.classroomId = id }
func (c *Classroom) OwnerProfileId() string             { return c.ownerProfileId }
func (c *Classroom) SetOwnerProfileId(id string)        { c.ownerProfileId = id }
func (c *Classroom) Name() string                       { return c.name }
func (c *Classroom) SetName(n string)                   { c.name = n }
func (c *Classroom) Description() *string               { return c.description }
func (c *Classroom) SetDescription(d *string)           { c.description = d }
func (c *Classroom) SchoolId() *string                  { return c.schoolId }
func (c *Classroom) SetSchoolId(id *string)             { c.schoolId = id }
func (c *Classroom) ProgramIds() []string               { return c.programIds }
func (c *Classroom) SetProgramIds(ids []string)         { c.programIds = ids }
func (c *Classroom) GradeId() *string                   { return c.gradeId }
func (c *Classroom) SetGradeId(id *string)              { c.gradeId = id }
func (c *Classroom) InviteCode() *string                { return c.inviteCode }
func (c *Classroom) SetInviteCode(v *string)            { c.inviteCode = v }
func (c *Classroom) InviteCodeExpiresDt() mtime.MathTime { return c.inviteCodeExpiresDt }
func (c *Classroom) SetInviteCodeExpiresDt(t mtime.MathTime) {
	c.inviteCodeExpiresDt = t
}
func (c *Classroom) MaxMembers() *int64                 { return c.maxMembers }
func (c *Classroom) SetMaxMembers(v *int64)             { c.maxMembers = v }
func (c *Classroom) MemberCount() int64                 { return c.memberCount }
func (c *Classroom) SetMemberCount(v int64)             { c.memberCount = v }
func (c *Classroom) StudentCount() int64                { return c.studentCount }
func (c *Classroom) SetStudentCount(v int64)            { c.studentCount = v }
func (c *Classroom) TeacherCount() int64                { return c.teacherCount }
func (c *Classroom) SetTeacherCount(v int64)            { c.teacherCount = v }
func (c *Classroom) CoverKey() *string                  { return c.coverKey }
func (c *Classroom) SetCoverKey(k *string)              { c.coverKey = k }
func (c *Classroom) Note() *string                      { return c.note }
func (c *Classroom) SetNote(n *string)                  { c.note = n }
func (c *Classroom) ClassroomStatus() *string           { return c.classroomStatus }
func (c *Classroom) SetClassroomStatus(v *string)       { c.classroomStatus = v }
func (c *Classroom) Status() string                     { return c.status }
func (c *Classroom) SetStatus(v string)                 { c.status = v }
func (c *Classroom) CreateId() *string                  { return c.createId }
func (c *Classroom) SetCreateId(id *string)             { c.createId = id }
func (c *Classroom) CreateDt() mtime.MathTime           { return c.createDt }
func (c *Classroom) SetCreateDt(t mtime.MathTime)       { c.createDt = t }
func (c *Classroom) ModifyId() *string                  { return c.modifyId }
func (c *Classroom) SetModifyId(id *string)             { c.modifyId = id }
func (c *Classroom) ModifyDt() mtime.MathTime           { return c.modifyDt }
func (c *Classroom) SetModifyDt(t mtime.MathTime)       { c.modifyDt = t }
