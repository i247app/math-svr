package domain

import (
	"github.com/google/uuid"
	"math-ai.com/math-ai/internal/driven-adapter/persistence/models"
)

type Contact struct {
	id        string
	uid 	  string
	contactName string
	contactEmail string
	contactPhone string
	contactMessage string
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

func (l *Contact) UID() string {
	return l.uid
}

func (l *Contact) SetUID(uid string) {
	l.uid = uid
}

func (l *Contact) ContactName() string {
	return l.contactName
}

func (l *Contact) SetContactName(contactName string) {
	l.contactName = contactName
}

func (l *Contact) ContactEmail() string {
	return l.contactEmail
}

func (l *Contact) SetContactEmail(contactEmail string) {
	l.contactEmail = contactEmail
}

func (l *Contact) ContactPhone() string {
	return l.contactPhone
}

func (l *Contact) SetContactPhone(contactPhone string) {
	l.contactPhone = contactPhone
}

func (l *Contact) ContactMessage() string {
	return l.contactMessage
}

func (l *Contact) SetContactMessage(contactMessage string) {
	l.contactMessage = contactMessage
}

func BuildContactDomainFromModel(model *models.ContactModel) *Contact {
	return &Contact{
		id:        model.ID,
		uid: 	   *model.UID,
		contactName:  model.ContactName,
		contactEmail: *model.ContactEmail,	
		contactPhone: *model.ContactPhone,	
		contactMessage: *model.ContactMessage,	
	}
}
