package domain

import (
	"time"

	"github.com/google/uuid"
	"math-ai.com/math-ai/internal/driven-adapter/persistence/models"
	"math-ai.com/math-ai/internal/shared/constant/enum"
)

type Profile struct {
	id         string
	uid        string
	name       string
	email      string
	phone      string
	gradeID    string
	grade      string
	semesterID string
	semester   string
	avatarKey  *string
	dob        *time.Time
	status     string
	createID   *int64
	createDT   time.Time
	modifyID   *int64
	modifyDT   time.Time
	deletedDT  *time.Time
}

func NewProfileDomain() *Profile {
	return &Profile{}
}

func (p *Profile) ID() string {
	return p.id
}

func (p *Profile) GenerateID() {
	p.id = uuid.New().String()
}

func (p *Profile) SetID(id string) {
	p.id = id
}

func (p *Profile) UID() string {
	return p.uid
}

func (p *Profile) SetUID(uid string) {
	p.uid = uid
}

func (p *Profile) Name() string {
	return p.name
}

func (p *Profile) SetName(name string) {
	p.name = name
}

func (p *Profile) Email() string {
	return p.email
}

func (p *Profile) SetEmail(email string) {
	p.email = email
}

func (p *Profile) Phone() string {
	return p.phone
}

func (p *Profile) SetPhone(phone string) {
	p.phone = phone
}

func (p *Profile) AvatarKey() *string {
	return p.avatarKey
}

func (p *Profile) SetAvatarKey(avatarKey *string) {
	p.avatarKey = avatarKey
}

func (p *Profile) Dob() *time.Time {
	return p.dob
}

func (p *Profile) SetDob(dob *time.Time) {
	p.dob = dob
}

func (p *Profile) SemesterID() string {
	return p.semesterID
}

func (p *Profile) SetSemesterID(semesterID string) {
	p.semesterID = semesterID
}

func (p *Profile) Semester() string {
	return p.semester
}

func (p *Profile) SetSemester(semester string) {
	p.semester = semester
}

func (p *Profile) GradeID() string {
	return p.gradeID
}

func (p *Profile) SetGradeID(gradeID string) {
	p.gradeID = gradeID
}

func (p *Profile) Grade() string {
	return p.grade
}

func (p *Profile) SetGrade(grade string) {
	p.grade = grade
}

func (p *Profile) Status() string {
	return p.status
}

func (p *Profile) SetStatus(status string) {
	if status == "" {
		status = string(enum.StatusActive)
	}
	p.status = status
}

func (p *Profile) CreateID() *int64 {
	return p.createID
}

func (p *Profile) SetCreateID(createID *int64) {
	p.createID = createID
}

func (p *Profile) CreatedAt() time.Time {
	return p.createDT
}

func (p *Profile) SetCreatedAt(createDT time.Time) {
	p.createDT = createDT
}

func (p *Profile) ModifyID() *int64 {
	return p.modifyID
}

func (p *Profile) SetModifyID(modifyID *int64) {
	p.modifyID = modifyID
}

func (p *Profile) ModifiedAt() time.Time {
	return p.modifyDT
}

func (p *Profile) SetModifiedAt(modifyDT time.Time) {
	p.modifyDT = modifyDT
}

func (p *Profile) DeletedAt() *time.Time {
	return p.deletedDT
}

func (p *Profile) SetDeletedAt(deletedDT *time.Time) {
	p.deletedDT = deletedDT
}

func BuildProfileDomainFromModel(model *models.ProfileModel) *Profile {
	return &Profile{
		id:        model.ID,
		uid:       model.UID,
		name:      model.Name,
		email:     model.Email,
		phone:     model.Phone,
		grade:     model.Grade,
		semester:  model.Semester,
		dob:       model.Dob,
		avatarKey: model.AvatarKey,
		status:    model.Status,
		createID:  model.CreateID,
		createDT:  model.CreateDT,
		modifyID:  model.ModifyID,
		modifyDT:  model.ModifyDT,
		deletedDT: model.DeletedDT,
	}
}
