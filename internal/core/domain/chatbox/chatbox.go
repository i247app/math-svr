package domain

import (
	"time"

	"github.com/sashabaranov/go-openai"
)

// Message represents a single message in a chat conversation
type Message struct {
	role      string
	content   string
	timestamp time.Time
}

func NewMessage(role, content string) *Message {
	return &Message{
		role:      role,
		content:   content,
		timestamp: time.Now(),
	}
}

func (m *Message) Role() string {
	return m.role
}

func (m *Message) SetRole(role string) {
	m.role = role
}

func (m *Message) Content() string {
	return m.content
}

func (m *Message) SetContent(content string) {
	m.content = content
}

func (m *Message) Timestamp() time.Time {
	return m.timestamp
}

func (m *Message) SetTimestamp(timestamp time.Time) {
	m.timestamp = timestamp
}

// Conversation represents a chat conversation
type Conversation struct {
	id           string
	messages     []*Message
	model        string
	temperature  float32
	maxTokens    int
	systemPrompt *string
	createdAt    time.Time
	updatedAt    time.Time
}

func NewConversation() *Conversation {
	now := time.Now()
	return &Conversation{
		messages:    make([]*Message, 0),
		model:       openai.GPT4oMini, // default model
		temperature: 0.7,              // default temperature
		maxTokens:   2000,             // default max tokens (enough for ~1500 words)
		createdAt:   now,
		updatedAt:   now,
	}
}

func (c *Conversation) ID() string {
	return c.id
}

func (c *Conversation) SetID(id string) {
	c.id = id
}

func (c *Conversation) Messages() []*Message {
	return c.messages
}

func (c *Conversation) AddMessage(message *Message) {
	c.messages = append(c.messages, message)
	c.updatedAt = time.Now()
}

func (c *Conversation) SetMessages(messages []*Message) {
	c.messages = messages
	c.updatedAt = time.Now()
}

func (c *Conversation) Model() string {
	return c.model
}

func (c *Conversation) SetModel(model string) {
	c.model = model
}

func (c *Conversation) Temperature() float32 {
	return c.temperature
}

func (c *Conversation) SetTemperature(temperature float32) {
	c.temperature = temperature
}

func (c *Conversation) MaxTokens() int {
	return c.maxTokens
}

func (c *Conversation) SetMaxTokens(maxTokens int) {
	c.maxTokens = maxTokens
}

func (c *Conversation) SystemPrompt() *string {
	return c.systemPrompt
}

func (c *Conversation) SetSystemPrompt(systemPrompt *string) {
	c.systemPrompt = systemPrompt
}

func (c *Conversation) CreatedAt() time.Time {
	return c.createdAt
}

func (c *Conversation) SetCreatedAt(createdAt time.Time) {
	c.createdAt = createdAt
}

func (c *Conversation) UpdatedAt() time.Time {
	return c.updatedAt
}

func (c *Conversation) SetUpdatedAt(updatedAt time.Time) {
	c.updatedAt = updatedAt
}
