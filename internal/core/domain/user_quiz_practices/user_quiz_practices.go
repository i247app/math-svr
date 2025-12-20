package domain

import (
	"github.com/google/uuid"
	"math-ai.com/math-ai/internal/driven-adapter/persistence/models"
	"math-ai.com/math-ai/internal/shared/constant/enum"
	"math-ai.com/math-ai/internal/shared/utils/time"
)

type UserQuizPractices struct {
	id        string
	uid       string
	questions string
	answers   string
	aiReview  string
	status    string
	createID  *int64
	createDT  time.MathTime
	modifyID  *int64
	modifyDT  time.MathTime
	deletedDT *time.MathTime
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

func (u *UserQuizPractices) CreatedAt() time.MathTime {
	return u.createDT
}

func (u *UserQuizPractices) SetCreatedAt(createDT time.MathTime) {
	u.createDT = createDT
}

func (u *UserQuizPractices) ModifyID() *int64 {
	return u.modifyID
}

func (u *UserQuizPractices) SetModifyID(modifyID *int64) {
	u.modifyID = modifyID
}

func (u *UserQuizPractices) ModifiedAt() time.MathTime {
	return u.modifyDT
}

func (u *UserQuizPractices) SetModifiedAt(modifyDT time.MathTime) {
	u.modifyDT = modifyDT
}

func (u *UserQuizPractices) DeletedAt() *time.MathTime {
	return u.deletedDT
}

func (u *UserQuizPractices) SetDeletedAt(deletedDT *time.MathTime) {
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
