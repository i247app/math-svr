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
	profileId     string
	userId        string
	name          string
	avatarKey     *string
	dob           mtime.MathTime
	programId     *string
	gradeId       *string
	semesterId    *string
	isDefault     bool
	note          *string
	profileStatus *string
	status        string
	createId      *string
	createDt      mtime.MathTime
	modifyId      *string
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

func (p *Profile) ProfileId() string {
	return p.profileId
}

func (p *Profile) SetProfileId(profileId string) {
	p.profileId = profileId
}

func (p *Profile) UserId() string {
	return p.userId
}

func (p *Profile) SetUserId(userId string) {
	p.userId = userId
}

func (p *Profile) Name() string {
	return p.name
}

func (p *Profile) SetName(name string) {
	p.name = name
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

func (p *Profile) ProgramId() *string {
	return p.programId
}

// SetProgramId accepts a pointer so the model layer (which stores
// *string for the now-nullable column) can pass through nil. nil is
// flattened to uuid.Nil so the domain field stays a value type.
func (p *Profile) SetProgramId(programId *string) {
	if programId == nil {
		p.programId = nil
		return
	}
	p.programId = programId
}

func (p *Profile) GradeId() *string {
	return p.gradeId
}

func (p *Profile) SetGradeId(gradeId *string) {
	if gradeId == nil {
		p.gradeId = nil
		return
	}
	p.gradeId = gradeId
}

func (p *Profile) SemesterId() *string {
	return p.semesterId
}

func (p *Profile) SetSemesterId(semesterId *string) {
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

func (p *Profile) CreateId() *string {
	return p.createId
}

func (p *Profile) SetCreateId(createId *string) {
	p.createId = createId
}

func (p *Profile) CreateDt() mtime.MathTime {
	return p.createDt
}

func (p *Profile) SetCreateDt(createDt mtime.MathTime) {
	p.createDt = createDt
}

func (p *Profile) ModifyId() *string {
	return p.modifyId
}

func (p *Profile) SetModifyId(modifyId *string) {
	p.modifyId = modifyId
}

func (p *Profile) ModifyDt() mtime.MathTime {
	return p.modifyDt
}

func (p *Profile) SetModifyDt(modifyDt mtime.MathTime) {
	p.modifyDt = modifyDt
}
