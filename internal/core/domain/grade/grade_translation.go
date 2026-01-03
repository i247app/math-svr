package domain

import (
	"github.com/google/uuid"
	"math-ai.com/math-ai/internal/driven-adapter/persistence/models"
	"math-ai.com/math-ai/internal/shared/utils/time"
)

type GradeTranslation struct {
	id          string
	gradeID     string
	language    string
	label       string
	description *string
	note        *string
	gtStatus    string
	status      string
	createID    *int64
	createDT    time.MathTime
	modifyID    *int64
	modifyDT    time.MathTime
	deletedDT   *time.MathTime
}

func NewGradeTranslationDomain() *GradeTranslation {
	return &GradeTranslation{}
}

func (gt *GradeTranslation) ID() string {
	return gt.id
}

func (gt *GradeTranslation) GenerateID() {
	gt.id = uuid.New().String()
}

func (gt *GradeTranslation) SetID(id string) {
	gt.id = id
}

func (gt *GradeTranslation) GradeID() string {
	return gt.gradeID
}

func (gt *GradeTranslation) SetGradeID(gradeID string) {
	gt.gradeID = gradeID
}

func (gt *GradeTranslation) Language() string {
	return gt.language
}

func (gt *GradeTranslation) SetLanguage(language string) {
	gt.language = language
}

func (gt *GradeTranslation) Label() string {
	return gt.label
}

func (gt *GradeTranslation) SetLabel(label string) {
	gt.label = label
}

func (gt *GradeTranslation) Description() *string {
	return gt.description
}

func (gt *GradeTranslation) SetDescription(description *string) {
	gt.description = description
}

func (gt *GradeTranslation) Note() *string {
	return gt.note
}

func (gt *GradeTranslation) SetNote(note *string) {
	gt.note = note
}

func (gt *GradeTranslation) GTStatus() string {
	return gt.gtStatus
}

func (gt *GradeTranslation) SetGTStatus(gtStatus string) {
	gt.gtStatus = gtStatus
}

func (gt *GradeTranslation) Status() string {
	return gt.status
}

func (gt *GradeTranslation) SetStatus(status string) {
	gt.status = status
}

func (gt *GradeTranslation) CreateID() *int64 {
	return gt.createID
}

func (gt *GradeTranslation) SetCreateID(createID *int64) {
	gt.createID = createID
}

func (gt *GradeTranslation) CreateDT() time.MathTime {
	return gt.createDT
}

func (gt *GradeTranslation) SetCreateDT(createDT time.MathTime) {
	gt.createDT = createDT
}

func (gt *GradeTranslation) ModifyID() *int64 {
	return gt.modifyID
}

func (gt *GradeTranslation) SetModifyID(modifyID *int64) {
	gt.modifyID = modifyID
}

func (gt *GradeTranslation) ModifyDT() time.MathTime {
	return gt.modifyDT
}

func (gt *GradeTranslation) SetModifyDT(modifyDT time.MathTime) {
	gt.modifyDT = modifyDT
}

func (gt *GradeTranslation) DeletedDT() *time.MathTime {
	return gt.deletedDT
}

func (gt *GradeTranslation) SetDeletedDT(deletedDT *time.MathTime) {
	gt.deletedDT = deletedDT
}

func BuildGradeTranslationFromModel(model *models.GradeTranslationModel) *GradeTranslation {
	return &GradeTranslation{
		id:          model.ID,
		gradeID:     model.GradeID,
		language:    model.Language,
		label:       model.Label,
		description: model.Description,
		note:        model.Note,
		gtStatus:    model.GTStatus,
		status:      model.Status,
		createID:    model.CreateID,
		createDT:    model.CreateDT,
		modifyID:    model.ModifyID,
		modifyDT:    model.ModifyDT,
		deletedDT:   model.DeletedDT,
	}
}
