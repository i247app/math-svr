package domain

import (
	"time"

	"github.com/google/uuid"
	"math-ai.com/math-ai/internal/driven-adapter/persistence/models"
	"math-ai.com/math-ai/internal/shared/constant/enum"
)

type UserLatestQuiz struct {
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

func NewUserLatestQuizDomain() *UserLatestQuiz {
	return &UserLatestQuiz{}
}

func (u *UserLatestQuiz) ID() string {
	return u.id
}

func (u *UserLatestQuiz) GenerateID() {
	u.id = uuid.New().String()
}

func (u *UserLatestQuiz) SetID(id string) {
	u.id = id
}

func (u *UserLatestQuiz) UID() string {
	return u.uid
}

func (u *UserLatestQuiz) SetUID(uid string) {
	u.uid = uid
}

func (u *UserLatestQuiz) Questions() string {
	return u.questions
}

func (u *UserLatestQuiz) SetQuestions(questions string) {
	u.questions = questions
}

func (u *UserLatestQuiz) Answers() string {
	return u.answers
}

func (u *UserLatestQuiz) SetAnswers(answers string) {
	u.answers = answers
}

func (u *UserLatestQuiz) AIReview() string {
	return u.aiReview
}

func (u *UserLatestQuiz) SetAIReview(aiReview string) {
	u.aiReview = aiReview
}

func (u *UserLatestQuiz) Status() string {
	return u.status
}

func (u *UserLatestQuiz) SetStatus(status string) {
	if status == "" {
		status = string(enum.StatusActive)
	}
	u.status = status
}

func (u *UserLatestQuiz) CreateID() *int64 {
	return u.createID
}

func (u *UserLatestQuiz) SetCreateID(createID *int64) {
	u.createID = createID
}

func (u *UserLatestQuiz) CreatedAt() time.Time {
	return u.createDT
}

func (u *UserLatestQuiz) SetCreatedAt(createDT time.Time) {
	u.createDT = createDT
}

func (u *UserLatestQuiz) ModifyID() *int64 {
	return u.modifyID
}

func (u *UserLatestQuiz) SetModifyID(modifyID *int64) {
	u.modifyID = modifyID
}

func (u *UserLatestQuiz) ModifiedAt() time.Time {
	return u.modifyDT
}

func (u *UserLatestQuiz) SetModifiedAt(modifyDT time.Time) {
	u.modifyDT = modifyDT
}

func (u *UserLatestQuiz) DeletedAt() *time.Time {
	return u.deletedDT
}

func (u *UserLatestQuiz) SetDeletedAt(deletedDT *time.Time) {
	u.deletedDT = deletedDT
}

func BuildUserLatestQuizDomainFromModel(model *models.UserLatestQuizModel) *UserLatestQuiz {
	return &UserLatestQuiz{
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
