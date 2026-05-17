package profile

import (
	"github.com/google/uuid"
	mtime "math-ai.com/math-ai/internal/domain/shared/time"
)

// Profile is a child managed by a parent (User). One row per child; the
// grade/semester pair is mutated as the school year advances (never inserted
// as a new row). MathTime carries the dob — zero value scans/marshals as
// NULL/null without needing a pointer.
type Profile struct {
	id            int64
	profileId     uuid.UUID
	userId        uuid.UUID
	name          string
	avatarKey     *string
	dob           mtime.MathTime
	programId     uuid.UUID
	gradeId       uuid.UUID
	semesterId    uuid.UUID
	note          *string
	profileStatus *string
	status        string
	createId      *uuid.UUID
	createDt      mtime.MathTime
	modifyId      *uuid.UUID
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

func (p *Profile) ProfileId() uuid.UUID {
	return p.profileId
}

func (p *Profile) SetProfileId(profileId uuid.UUID) {
	p.profileId = profileId
}

func (p *Profile) UserId() uuid.UUID {
	return p.userId
}

func (p *Profile) SetUserId(userId uuid.UUID) {
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

func (p *Profile) ProgramId() uuid.UUID {
	return p.programId
}

func (p *Profile) SetProgramId(programId uuid.UUID) {
	p.programId = programId
}

func (p *Profile) GradeId() uuid.UUID {
	return p.gradeId
}

func (p *Profile) SetGradeId(gradeId uuid.UUID) {
	p.gradeId = gradeId
}

func (p *Profile) SemesterId() uuid.UUID {
	return p.semesterId
}

func (p *Profile) SetSemesterId(semesterId uuid.UUID) {
	p.semesterId = semesterId
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

func (p *Profile) CreateId() *uuid.UUID {
	return p.createId
}

func (p *Profile) SetCreateId(createId *uuid.UUID) {
	p.createId = createId
}

func (p *Profile) CreateDt() mtime.MathTime {
	return p.createDt
}

func (p *Profile) SetCreateDt(createDt mtime.MathTime) {
	p.createDt = createDt
}

func (p *Profile) ModifyId() *uuid.UUID {
	return p.modifyId
}

func (p *Profile) SetModifyId(modifyId *uuid.UUID) {
	p.modifyId = modifyId
}

func (p *Profile) ModifyDt() mtime.MathTime {
	return p.modifyDt
}

func (p *Profile) SetModifyDt(modifyDt mtime.MathTime) {
	p.modifyDt = modifyDt
}
