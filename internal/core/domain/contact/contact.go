package domain

import (
	"github.com/google/uuid"
	"math-ai.com/math-ai/internal/driven-adapter/persistence/models"
	"math-ai.com/math-ai/internal/shared/utils/time"
)

type Contact struct {
	id             string
	uid            *string
	contactName    string
	contactEmail   *string
	contactPhone   *string
	contactMessage string
	isRead         bool
	note           *string
	contactStatus  string
	status         string
	createID       *int64
	createDT       time.MathTime
	modifyID       *int64
	modifyDT       time.MathTime
	deletedDT      *time.MathTime
}

func NewContactDomain() *Contact {
	return &Contact{}
}

func (l *Contact) ID() string {
	return l.id
}

func (l *Contact) SetID(id string) {
	l.id = id
}

func (l *Contact) GenerateID() {
	l.id = uuid.New().String()
}

func (l *Contact) UID() *string {
	return l.uid
}

func (l *Contact) SetUID(uid *string) {
	l.uid = uid
}

func (l *Contact) ContactName() string {
	return l.contactName
}

func (l *Contact) SetContactName(contactName string) {
	l.contactName = contactName
}

func (l *Contact) ContactEmail() *string {
	return l.contactEmail
}

func (l *Contact) SetContactEmail(contactEmail *string) {
	l.contactEmail = contactEmail
}

func (l *Contact) ContactPhone() *string {
	return l.contactPhone
}

func (l *Contact) SetContactPhone(contactPhone *string) {
	l.contactPhone = contactPhone
}

func (l *Contact) ContactMessage() string {
	return l.contactMessage
}

func (l *Contact) SetContactMessage(contactMessage string) {
	l.contactMessage = contactMessage
}

func (l *Contact) SetIsRead(isRead bool) {
	l.isRead = isRead
}

func (l *Contact) IsRead() bool {
	return l.isRead
}

func (l *Contact) Note() *string {
	return l.note
}

func (l *Contact) SetNote(note *string) {
	l.note = note
}

func (l *Contact) ContactStatus() string {
	return l.contactStatus
}

func (l *Contact) SetContactStatus(contactStatus string) {
	l.contactStatus = contactStatus
}

func (l *Contact) Status() string {
	return l.status
}

func (l *Contact) SetStatus(status string) {
	l.status = status
}

func (l *Contact) CreatedBy() *int64 {
	return l.createID
}

func (l *Contact) SetCreatedBy(createID *int64) {
	l.createID = createID
}

func (l *Contact) CreateDT() time.MathTime {
	return l.createDT
}

func (l *Contact) SetCreateDT(createDT time.MathTime) {
	l.createDT = createDT
}

func (l *Contact) ModifiedBy() *int64 {
	return l.modifyID
}

func (l *Contact) SetModifiedBy(modifyID *int64) {
	l.modifyID = modifyID
}

func (l *Contact) ModifyDT() time.MathTime {
	return l.modifyDT
}

func (l *Contact) SetModifyDT(modifyDT time.MathTime) {
	l.modifyDT = modifyDT
}

func (l *Contact) DeletedDT() *time.MathTime {
	return l.deletedDT
}

func (l *Contact) SetDeletedDT(deletedDT *time.MathTime) {
	l.deletedDT = deletedDT
}

func BuildContactDomainFromModel(model *models.ContactModel) *Contact {
	isRead := false
	if model.IsRead != nil {
		isRead = *model.IsRead
	}

	return &Contact{
		id:             model.ID,
		uid:            model.UID,
		contactName:    model.ContactName,
		contactEmail:   model.ContactEmail,
		contactPhone:   model.ContactPhone,
		contactMessage: model.ContactMessage,
		isRead:         isRead,
		note:           model.Note,
		contactStatus:  model.ContactStatus,
		status:         model.Status,
		createID:       model.CreateID,
		createDT:       model.CreateDT,
		modifyID:       model.ModifyID,
		modifyDT:       model.ModifyDT,
		deletedDT:      model.DeletedDT,
	}
}
