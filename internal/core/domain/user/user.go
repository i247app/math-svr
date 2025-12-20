package domain

import (
	"github.com/google/uuid"
	"math-ai.com/math-ai/internal/driven-adapter/persistence/models"
	"math-ai.com/math-ai/internal/shared/utils/time"
)

type User struct {
	id         string
	name       string
	phone      string
	email      string
	avatarKey  *string
	dob        *time.MathTime
	gradeID    string
	semesterID string
	role       string
	roleID     string // New RBAC role ID
	password   string
	status     string
	createDT   time.MathTime
	modifyDT   time.MathTime
	deletedDT  *time.MathTime
}

func NewUserDomain() *User {
	return &User{}
}

func (u *User) ID() string {
	return u.id
}

func (u *User) GenerateID() {
	u.id = uuid.New().String()
}

func (u *User) SetID(id string) {
	u.id = id
}

func (u *User) Name() string {
	return u.name
}

func (u *User) SetName(name string) {
	u.name = name
}

func (u *User) Phone() string {
	return u.phone
}

func (u *User) SetPhone(phone string) {
	u.phone = phone
}

func (u *User) Email() string {
	return u.email
}

func (u *User) SetEmail(email string) {
	u.email = email
}

func (u *User) AvatarKey() *string {
	return u.avatarKey
}

func (u *User) SetAvatarKey(avatarKey *string) {
	u.avatarKey = avatarKey
}

func (u *User) DOB() *time.MathTime {
	return u.dob
}

func (u *User) SetDOB(dob *time.MathTime) {
	u.dob = dob
}

func (u *User) GradeID() string {
	return u.gradeID
}

func (u *User) SetGradeID(gradeID string) {
	u.gradeID = gradeID
}

func (u *User) SemesterID() string {
	return u.semesterID
}

func (u *User) SetSemesterID(semesterID string) {
	u.semesterID = semesterID
}

func (u *User) RoleID() string {
	return u.roleID
}

func (u *User) SetRoleID(roleID string) {
	u.roleID = roleID
}

func (u *User) Role() string {
	return u.role
}

func (u *User) SetRole(role string) {
	u.role = role
}

func (u *User) Password() string {
	return u.password
}

func (u *User) SetPassword(password string) {
	u.password = password
}

func (u *User) Status() string {
	return u.status
}

func (u *User) SetStatus(status string) {
	u.status = status
}

func (u *User) CreatedAt() time.MathTime {
	return u.createDT
}

func (u *User) SetCreatedAt(createAt time.MathTime) {
	u.createDT = createAt
}

func (u *User) ModifyAt() time.MathTime {
	return u.modifyDT
}

func (u *User) SetModifyAt(modifyAt time.MathTime) {
	u.modifyDT = modifyAt
}

func (u *User) DeletedAt() *time.MathTime {
	return u.deletedDT
}

func (u *User) SetDeletedAt(deletedAt *time.MathTime) {
	u.deletedDT = deletedAt
}

func BuildUserDomainFromModel(model *models.UserModel) *User {
	return &User{
		id:        model.ID,
		name:      model.Name,
		phone:     model.Phone,
		email:     model.Email,
		avatarKey: model.AvatarKey,
		dob:       model.Dob,
		role:      model.Role,
		roleID:    model.RoleID,
		password:  model.HashPassword,
		status:    model.Status,
		createDT:  model.CreateDT,
		modifyDT:  model.ModifyDT,
	}
}
