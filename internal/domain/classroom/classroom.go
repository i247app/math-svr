package classroom

import (
	"time"

	"math-ai.com/math-ai/internal/domain/shared/audit"
)

// Classroom models ma_classrooms. owner_profile_id is denormalized for
// O(1) "who owns this" lookups; it is kept in sync with the matching
// OWNER row in ma_classroom_members inside the same UnitOfWork. The
// classroom_code path is opt-in: nil/empty + a null expiry means
// joining is closed (only targeted invitations work).
type Classroom struct {
	audit.Audit

	classroomId            int64
	ownerProfileId         int64
	name                   string
	description            *string
	schoolId               *int64
	gradeId                *int64
	classroomCode          *string
	classroomCodeExpiresDt time.Time
	maxMembers             *int64
	memberCount            int64
	studentCount           int64
	teacherCount           int64
	coverKey               *string
	classroomStatus        *string

	// programIds is the in-memory denormalisation of the
	// ma_classroom_programs junction rows for this classroom. It is set
	// by the query/command layer after the classroom row is fetched;
	// no column on ma_classrooms backs it. A nil slice means "not
	// hydrated"; an empty non-nil slice means "hydrated, zero programs".
	programIds []int64
}

func NewClassroom() *Classroom {
	return &Classroom{}
}

func (c *Classroom) ClassroomId() int64      { return c.classroomId }
func (c *Classroom) SetClassroomId(id int64) { c.classroomId = id }

func (c *Classroom) OwnerProfileId() int64      { return c.ownerProfileId }
func (c *Classroom) SetOwnerProfileId(id int64) { c.ownerProfileId = id }

func (c *Classroom) Name() string     { return c.name }
func (c *Classroom) SetName(n string) { c.name = n }

func (c *Classroom) Description() *string     { return c.description }
func (c *Classroom) SetDescription(d *string) { c.description = d }

func (c *Classroom) SchoolId() *int64      { return c.schoolId }
func (c *Classroom) SetSchoolId(id *int64) { c.schoolId = id }

func (c *Classroom) GradeId() *int64      { return c.gradeId }
func (c *Classroom) SetGradeId(id *int64) { c.gradeId = id }

func (c *Classroom) ClassroomCode() *string     { return c.classroomCode }
func (c *Classroom) SetClassroomCode(v *string) { c.classroomCode = v }

func (c *Classroom) ClassroomCodeExpiresDt() time.Time     { return c.classroomCodeExpiresDt }
func (c *Classroom) SetClassroomCodeExpiresDt(t time.Time) { c.classroomCodeExpiresDt = t }

func (c *Classroom) MaxMembers() *int64     { return c.maxMembers }
func (c *Classroom) SetMaxMembers(v *int64) { c.maxMembers = v }

func (c *Classroom) MemberCount() int64     { return c.memberCount }
func (c *Classroom) SetMemberCount(v int64) { c.memberCount = v }

func (c *Classroom) StudentCount() int64     { return c.studentCount }
func (c *Classroom) SetStudentCount(v int64) { c.studentCount = v }

func (c *Classroom) TeacherCount() int64     { return c.teacherCount }
func (c *Classroom) SetTeacherCount(v int64) { c.teacherCount = v }

func (c *Classroom) CoverKey() *string     { return c.coverKey }
func (c *Classroom) SetCoverKey(k *string) { c.coverKey = k }

func (c *Classroom) ClassroomStatus() *string     { return c.classroomStatus }
func (c *Classroom) SetClassroomStatus(v *string) { c.classroomStatus = v }

func (c *Classroom) ProgramIds() []int64       { return c.programIds }
func (c *Classroom) SetProgramIds(ids []int64) { c.programIds = ids }
