package profile

import (
	mtime "math-ai.com/math-ai/internal/domain/shared/time"
)

// Profile is a child managed by a parent (User). One row per child; the
// grade/semester pair is mutated as the school year advances (never inserted
// as a new row). MathTime carries the dob — zero value scans/marshals as
// NULL/null without needing a pointer.
type Profile struct {
	id            int64
	profileId     int64
	profileCode   string
	userId        int64
	name          string
	phone         *string
	email         *string
	role          string
	avatarKey     *string
	dob           mtime.MathTime
	schoolId      *int64
	programId     *int64
	gradeId       *int64
	semesterId    *int64
	isDefault     bool
	idType        *string
	teacherId     *string
	studentId     *string
	note          *string
	profileStatus *string
	status        string
	createId      *int64
	createDt      mtime.MathTime
	modifyId      *int64
	modifyDt      mtime.MathTime
}

func NewProfile() *Profile {
	return &Profile{}
}

func (p *Profile) Id() int64 {
	return p.id
}

func (p *Profile) SetId(id int64) {
	p.id = id
}

func (p *Profile) ProfileId() int64 {
	return p.profileId
}

func (p *Profile) SetProfileId(profileId int64) {
	p.profileId = profileId
}

// ProfileCode is the human-readable id (e.g. "AA-1234") minted at
// create time and unique across ma_profiles. Used by clients for
// display + search alongside the numeric profile_id.
func (p *Profile) ProfileCode() string {
	return p.profileCode
}

func (p *Profile) SetProfileCode(code string) {
	p.profileCode = code
}

func (p *Profile) UserId() int64 {
	return p.userId
}

func (p *Profile) SetUserId(userId int64) {
	p.userId = userId
}

func (p *Profile) Name() string {
	return p.name
}

func (p *Profile) SetName(name string) {
	p.name = name
}

func (p *Profile) Phone() *string {
	return p.phone
}

func (p *Profile) SetPhone(phone *string) {
	p.phone = phone
}

func (p *Profile) Email() *string {
	return p.email
}

func (p *Profile) SetEmail(email *string) {
	p.email = email
}

func (p *Profile) Role() string {
	return p.role
}

func (p *Profile) SetRole(role string) {
	p.role = role
}

func (p *Profile) AvatarKey() *string {
	return p.avatarKey
}

func (p *Profile) SetAvatarKey(avatarKey *string) {
	p.avatarKey = avatarKey
}

func (p *Profile) Dob() mtime.MathTime {
	return p.dob
}

func (p *Profile) SetDob(dob mtime.MathTime) {
	p.dob = dob
}

func (p *Profile) SchoolId() *int64 {
	return p.schoolId
}

// SetSchoolId mirrors the other reference-id setters — nil passes through
// so the model layer can flatten a NULL column to "no school assigned".
func (p *Profile) SetSchoolId(schoolId *int64) {
	if schoolId == nil {
		p.schoolId = nil
		return
	}
	p.schoolId = schoolId
}

func (p *Profile) ProgramId() *int64 {
	return p.programId
}

// SetProgramId accepts a pointer so the model layer (which stores
// *string for the now-nullable column) can pass through nil. nil is
// flattened to uuid.Nil so the domain field stays a value type.
func (p *Profile) SetProgramId(programId *int64) {
	if programId == nil {
		p.programId = nil
		return
	}
	p.programId = programId
}

func (p *Profile) GradeId() *int64 {
	return p.gradeId
}

func (p *Profile) SetGradeId(gradeId *int64) {
	if gradeId == nil {
		p.gradeId = nil
		return
	}
	p.gradeId = gradeId
}

func (p *Profile) SemesterId() *int64 {
	return p.semesterId
}

func (p *Profile) SetSemesterId(semesterId *int64) {
	if semesterId == nil {
		p.semesterId = nil
		return
	}
	p.semesterId = semesterId
}

func (p *Profile) IsDefault() bool {
	return p.isDefault
}

func (p *Profile) SetIsDefault(isDefault bool) {
	p.isDefault = isDefault
}

func (p *Profile) IdType() *string {
	return p.idType
}

func (p *Profile) SetIdType(idType *string) {
	p.idType = idType
}

func (p *Profile) TeacherId() *string {
	return p.teacherId
}

func (p *Profile) SetTeacherId(teacherId *string) {
	p.teacherId = teacherId
}

func (p *Profile) StudentId() *string {
	return p.studentId
}

func (p *Profile) SetStudentId(studentId *string) {
	p.studentId = studentId
}

func (p *Profile) Note() *string {
	return p.note
}

func (p *Profile) SetNote(note *string) {
	p.note = note
}

func (p *Profile) ProfileStatus() *string {
	return p.profileStatus
}

func (p *Profile) SetProfileStatus(profileStatus *string) {
	p.profileStatus = profileStatus
}

func (p *Profile) Status() string {
	return p.status
}

func (p *Profile) SetStatus(status string) {
	p.status = status
}

func (p *Profile) CreateId() *int64 {
	return p.createId
}

func (p *Profile) SetCreateId(createId *int64) {
	p.createId = createId
}

func (p *Profile) CreateDt() mtime.MathTime {
	return p.createDt
}

func (p *Profile) SetCreateDt(createDt mtime.MathTime) {
	p.createDt = createDt
}

func (p *Profile) ModifyId() *int64 {
	return p.modifyId
}

func (p *Profile) SetModifyId(modifyId *int64) {
	p.modifyId = modifyId
}

func (p *Profile) ModifyDt() mtime.MathTime {
	return p.modifyDt
}

func (p *Profile) SetModifyDt(modifyDt mtime.MathTime) {
	p.modifyDt = modifyDt
}
