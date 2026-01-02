package domain

import (
	"github.com/google/uuid"
	"math-ai.com/math-ai/internal/driven-adapter/persistence/models"
	"math-ai.com/math-ai/internal/shared/constant/enum"
	"math-ai.com/math-ai/internal/shared/utils/time"
)

type SemesterTranslation struct {
	id          string
	semesterID  string
	language    string
	name        string
	description *string
	note        *string
	stStatus    string
	status      string
	createID    *int64
	createDT    time.MathTime
	modifyID    *int64
	modifyDT    time.MathTime
	deletedDT   *time.MathTime
}

func NewSemesterTranslation() *SemesterTranslation {
	return &SemesterTranslation{}
}

func (st *SemesterTranslation) ID() string {
	return st.id
}

func (st *SemesterTranslation) GenerateID() {
	st.id = uuid.New().String()
}

func (st *SemesterTranslation) SetID(id string) {
	st.id = id
}

func (st *SemesterTranslation) SemesterID() string {
	return st.semesterID
}

func (st *SemesterTranslation) SetSemesterID(semesterID string) {
	st.semesterID = semesterID
}

func (st *SemesterTranslation) Language() string {
	return st.language
}

func (st *SemesterTranslation) SetLanguage(language string) {
	st.language = language
}

func (st *SemesterTranslation) Name() string {
	return st.name
}

func (st *SemesterTranslation) SetName(name string) {
	st.name = name
}

func (st *SemesterTranslation) Description() *string {
	return st.description
}

func (st *SemesterTranslation) SetDescription(description *string) {
	st.description = description
}

func (st *SemesterTranslation) Note() *string {
	return st.note
}

func (st *SemesterTranslation) SetNote(note *string) {
	st.note = note
}

func (st *SemesterTranslation) STStatus() string {
	return st.stStatus
}

func (st *SemesterTranslation) SetSTStatus(stStatus string) {
	if stStatus == "" {
		stStatus = string(enum.StatusActive)
	}
	st.stStatus = stStatus
}

func (st *SemesterTranslation) Status() string {
	return st.status
}

func (st *SemesterTranslation) SetStatus(status string) {
	st.status = status
}

func (st *SemesterTranslation) CreatedBy() *int64 {
	return st.createID
}

func (st *SemesterTranslation) SetCreatedBy(createID *int64) {
	st.createID = createID
}

func (st *SemesterTranslation) CreateDT() time.MathTime {
	return st.createDT
}

func (st *SemesterTranslation) SetCreateDT(createDT time.MathTime) {
	st.createDT = createDT
}

func (st *SemesterTranslation) ModifiedBy() *int64 {
	return st.modifyID
}

func (st *SemesterTranslation) SetModifiedBy(modifyID *int64) {
	st.modifyID = modifyID
}

func (st *SemesterTranslation) ModifyDT() time.MathTime {
	return st.modifyDT
}

func (st *SemesterTranslation) SetModifyDT(modifyDT time.MathTime) {
	st.modifyDT = modifyDT
}

func (st *SemesterTranslation) DeletedDT() *time.MathTime {
	return st.deletedDT
}

func BuildSemesterTranslationFromModel(model *models.SemesterTranslationModel) *SemesterTranslation {
	return &SemesterTranslation{
		id:          model.ID,
		semesterID:  model.SemesterID,
		language:    model.Language,
		name:        model.Name,
		description: model.Description,
		note:        model.Note,
		stStatus:    model.STStatus,
		status:      model.Status,
		createID:    model.CreateID,
		createDT:    model.CreateDT,
		modifyID:    model.ModifyID,
		modifyDT:    model.ModifyDT,
		deletedDT:   model.DeletedDT,
	}
}
