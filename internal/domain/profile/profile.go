package profile

import (
	"time"

	"math-ai.com/math-ai/internal/domain/shared/audit"
)

// Profile is a child managed by a parent (User). One row per child; the
// grade/semester pair is mutated as the school year advances (never inserted
// as a new row). A zero dob means "not set"; the repository writes it as
// NULL and the API renders it as "" (empty).
type Profile struct {
	audit.Audit

	profileId     int64
	profileCode   string
	uid           int64
	name          string
	phone         *string
	email         *string
	role          *string
	identityCode  *string
	avatarKey     *string
	dob           time.Time
	schoolId      *int64
	programId     *int64
	gradeId       *int64
	semesterId    *int64
	isDefault     bool
	idType        *string
	teacherId     *string
	studentId     *string
	profileStatus *string
}

func NewProfile() *Profile {
	return &Profile{}
}

func (p *Profile) ProfileId() int64             { return p.profileId }
func (p *Profile) SetProfileId(profileId int64) { p.profileId = profileId }

// ProfileCode is the human-readable id (e.g. "AA-1234") minted at
// create time and unique across ma_profiles. Used by clients for
// display + search alongside the numeric profile_id.
func (p *Profile) ProfileCode() string        { return p.profileCode }
func (p *Profile) SetProfileCode(code string) { p.profileCode = code }

func (p *Profile) UID() int64       { return p.uid }
func (p *Profile) SetUID(uid int64) { p.uid = uid }

func (p *Profile) Name() string        { return p.name }
func (p *Profile) SetName(name string) { p.name = name }

func (p *Profile) Phone() *string         { return p.phone }
func (p *Profile) SetPhone(phone *string) { p.phone = phone }

func (p *Profile) Email() *string         { return p.email }
func (p *Profile) SetEmail(email *string) { p.email = email }

// Role mirrors ma_users.role for this child. Nullable since migration
// 031: a guest profile has no declared role.
func (p *Profile) Role() *string        { return p.role }
func (p *Profile) SetRole(role *string) { p.role = role }

// IdentityCode is GUEST / USER / VERIFIED — see enum.IdentityCodeType.
// VERIFIED is paired with profileStatus OFFICIAL; both are produced by
// DeriveIdentity, never set independently.
func (p *Profile) IdentityCode() *string                { return p.identityCode }
func (p *Profile) SetIdentityCode(identityCode *string) { p.identityCode = identityCode }

func (p *Profile) AvatarKey() *string             { return p.avatarKey }
func (p *Profile) SetAvatarKey(avatarKey *string) { p.avatarKey = avatarKey }

func (p *Profile) Dob() time.Time       { return p.dob }
func (p *Profile) SetDob(dob time.Time) { p.dob = dob }

func (p *Profile) SchoolId() *int64            { return p.schoolId }
func (p *Profile) SetSchoolId(schoolId *int64) { p.schoolId = schoolId }

func (p *Profile) ProgramId() *int64             { return p.programId }
func (p *Profile) SetProgramId(programId *int64) { p.programId = programId }

func (p *Profile) GradeId() *int64           { return p.gradeId }
func (p *Profile) SetGradeId(gradeId *int64) { p.gradeId = gradeId }

func (p *Profile) SemesterId() *int64              { return p.semesterId }
func (p *Profile) SetSemesterId(semesterId *int64) { p.semesterId = semesterId }

func (p *Profile) IsDefault() bool             { return p.isDefault }
func (p *Profile) SetIsDefault(isDefault bool) { p.isDefault = isDefault }

func (p *Profile) IdType() *string          { return p.idType }
func (p *Profile) SetIdType(idType *string) { p.idType = idType }

func (p *Profile) TeacherId() *string             { return p.teacherId }
func (p *Profile) SetTeacherId(teacherId *string) { p.teacherId = teacherId }

func (p *Profile) StudentId() *string             { return p.studentId }
func (p *Profile) SetStudentId(studentId *string) { p.studentId = studentId }

func (p *Profile) ProfileStatus() *string                 { return p.profileStatus }
func (p *Profile) SetProfileStatus(profileStatus *string) { p.profileStatus = profileStatus }
