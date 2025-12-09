package domain

import (
	"time"

	"github.com/google/uuid"
	"math-ai.com/math-ai/internal/driven-adapter/persistence/models"
	"math-ai.com/math-ai/internal/shared/constant/enum"
)

type UserQuizAssessment struct {
	id            string
	uid           string
	questions     string
	answers       string
	aiReview      string
	aiDetectGrade string
	status        string
	createID      *int64
	createDT      time.Time
	modifyID      *int64
	modifyDT      time.Time
	deletedDT     *time.Time
}

func NewUserQuizAssessmentDomain() *UserQuizAssessment {
	return &UserQuizAssessment{}
}

func (u *UserQuizAssessment) ID() string {
	return u.id
}

func (u *UserQuizAssessment) GenerateID() {
	u.id = uuid.New().String()
}

func (u *UserQuizAssessment) SetID(id string) {
	u.id = id
}

func (u *UserQuizAssessment) UID() string {
	return u.uid
}

func (u *UserQuizAssessment) SetUID(uid string) {
	u.uid = uid
}

func (u *UserQuizAssessment) Questions() string {
	return u.questions
}

func (u *UserQuizAssessment) SetQuestions(questions string) {
	u.questions = questions
}

func (u *UserQuizAssessment) Answers() string {
	return u.answers
}

func (u *UserQuizAssessment) SetAnswers(answers string) {
	u.answers = answers
}

func (u *UserQuizAssessment) AIReview() string {
	return u.aiReview
}

func (u *UserQuizAssessment) SetAIReview(aiReview string) {
	u.aiReview = aiReview
}

func (u *UserQuizAssessment) AIDetectGrade() string {
	return u.aiDetectGrade
}

func (u *UserQuizAssessment) SetAIDetectGrade(aiDetectGrade string) {
	u.aiDetectGrade = aiDetectGrade
}

func (u *UserQuizAssessment) Status() string {
	return u.status
}

func (u *UserQuizAssessment) SetStatus(status string) {
	if status == "" {
		status = string(enum.StatusActive)
	}
	u.status = status
}

func (u *UserQuizAssessment) CreateID() *int64 {
	return u.createID
}

func (u *UserQuizAssessment) SetCreateID(createID *int64) {
	u.createID = createID
}

func (u *UserQuizAssessment) CreatedAt() time.Time {
	return u.createDT
}

func (u *UserQuizAssessment) SetCreatedAt(createDT time.Time) {
	u.createDT = createDT
}

func (u *UserQuizAssessment) ModifyID() *int64 {
	return u.modifyID
}

func (u *UserQuizAssessment) SetModifyID(modifyID *int64) {
	u.modifyID = modifyID
}

func (u *UserQuizAssessment) ModifiedAt() time.Time {
	return u.modifyDT
}

func (u *UserQuizAssessment) SetModifiedAt(modifyDT time.Time) {
	u.modifyDT = modifyDT
}

func (u *UserQuizAssessment) DeletedAt() *time.Time {
	return u.deletedDT
}

func (u *UserQuizAssessment) SetDeletedAt(deletedDT *time.Time) {
	u.deletedDT = deletedDT
}

func BuildUserQuizAssessmentDomainFromModel(model *models.UserQuizAssessmentModel) *UserQuizAssessment {
	return &UserQuizAssessment{
		id:            model.ID,
		uid:           model.UID,
		questions:     model.Questions,
		answers:       model.Answers,
		aiReview:      model.AIReview,
		aiDetectGrade: model.AIDetectGrade,
		status:        model.Status,
		createID:      model.CreateID,
		createDT:      model.CreateDT,
		modifyID:      model.ModifyID,
		modifyDT:      model.ModifyDT,
		deletedDT:     model.DeletedDT,
	}
}
