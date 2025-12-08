package domain

import (
	"time"

	"github.com/google/uuid"
	"math-ai.com/math-ai/internal/driven-adapter/persistence/models"
	"math-ai.com/math-ai/internal/shared/constant/enum"
)

type UserQuizPractices struct {
	id        string
	uid       string
	questions string
	answers   string
	aiReview  string
	status    string
	createID  *int64
	createDT  time.Time
	modifyID  *int64
	modifyDT  time.Time
	deletedDT *time.Time
}

func NewUserQuizPracticesDomain() *UserQuizPractices {
	return &UserQuizPractices{}
}

func (u *UserQuizPractices) ID() string {
	return u.id
}

func (u *UserQuizPractices) GenerateID() {
	u.id = uuid.New().String()
}

func (u *UserQuizPractices) SetID(id string) {
	u.id = id
}

func (u *UserQuizPractices) UID() string {
	return u.uid
}

func (u *UserQuizPractices) SetUID(uid string) {
	u.uid = uid
}

func (u *UserQuizPractices) Questions() string {
	return u.questions
}

func (u *UserQuizPractices) SetQuestions(questions string) {
	u.questions = questions
}

func (u *UserQuizPractices) Answers() string {
	return u.answers
}

func (u *UserQuizPractices) SetAnswers(answers string) {
	u.answers = answers
}

func (u *UserQuizPractices) AIReview() string {
	return u.aiReview
}

func (u *UserQuizPractices) SetAIReview(aiReview string) {
	u.aiReview = aiReview
}

func (u *UserQuizPractices) Status() string {
	return u.status
}

func (u *UserQuizPractices) SetStatus(status string) {
	if status == "" {
		status = string(enum.StatusActive)
	}
	u.status = status
}

func (u *UserQuizPractices) CreateID() *int64 {
	return u.createID
}

func (u *UserQuizPractices) SetCreateID(createID *int64) {
	u.createID = createID
}

func (u *UserQuizPractices) CreatedAt() time.Time {
	return u.createDT
}

func (u *UserQuizPractices) SetCreatedAt(createDT time.Time) {
	u.createDT = createDT
}

func (u *UserQuizPractices) ModifyID() *int64 {
	return u.modifyID
}

func (u *UserQuizPractices) SetModifyID(modifyID *int64) {
	u.modifyID = modifyID
}

func (u *UserQuizPractices) ModifiedAt() time.Time {
	return u.modifyDT
}

func (u *UserQuizPractices) SetModifiedAt(modifyDT time.Time) {
	u.modifyDT = modifyDT
}

func (u *UserQuizPractices) DeletedAt() *time.Time {
	return u.deletedDT
}

func (u *UserQuizPractices) SetDeletedAt(deletedDT *time.Time) {
	u.deletedDT = deletedDT
}

func BuildUserQuizPracticesDomainFromModel(model *models.UserQuizPracticesModel) *UserQuizPractices {
	return &UserQuizPractices{
		id:        model.ID,
		uid:       model.UID,
		questions: model.Questions,
		answers:   model.Answers,
		aiReview:  model.AIReview,
		status:    model.Status,
		createID:  model.CreateID,
		createDT:  model.CreateDT,
		modifyID:  model.ModifyID,
		modifyDT:  model.ModifyDT,
		deletedDT: model.DeletedDT,
	}
}
