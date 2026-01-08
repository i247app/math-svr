# Payment Gateway Architecture Design

## Table of Contents
1. [Overview](#overview)
2. [Domain Layer](#domain-layer)
3. [Dependency Injection Interfaces](#dependency-injection-interfaces)
4. [Application Layer](#application-layer)
5. [Adapter Layer](#adapter-layer)
6. [Provider Implementations](#provider-implementations)
7. [Error Handling & Normalization](#error-handling--normalization)
8. [Idempotency & Retries](#idempotency--retries)
9. [Security Considerations](#security-considerations)
10. [Design Decisions & Tradeoffs](#design-decisions--tradeoffs)

---

## Overview

This architecture provides a **provider-agnostic payment gateway** that integrates with multiple payment processors (Stripe, CyberSource, Plaid, Goat, etc.) while maintaining clean separation of concerns.

### Key Principles
- **Provider Independence**: Business logic never depends on specific provider implementations
- **Unified Interface**: Single API for all payment operations regardless of provider
- **Normalized Responses**: All provider responses are mapped to common domain models
- **Extensibility**: Adding new providers requires zero changes to existing business logic
- **Testability**: Easy to mock providers for testing

---

## Domain Layer

### File Structure
```
internal/core/domain/payment/
├── customer.go
├── payment_method.go
├── payment.go
├── payment_state.go
├── transaction_type.go
├── provider_type.go
└── errors.go
```

### Domain Entities

#### 1. Customer Entity (`customer.go`)
```go
package payment

import "time"

// Customer represents a payment customer
type Customer struct {
	id           string    // Internal UUID
	providerID   string    // Provider-specific customer ID (e.g., Stripe cus_xxx)
	provider     Provider  // Which provider this customer belongs to
	userID       string    // Link to your user system
	email        string
	name         string
	phone        *string
	metadata     map[string]string // Additional provider-specific data
	createdAt    time.Time
	updatedAt    time.Time
}

// Constructor
func NewCustomer(
	id string,
	providerID string,
	provider Provider,
	userID string,
	email string,
	name string,
) *Customer {
	now := time.Now()
	return &Customer{
		id:         id,
		providerID: providerID,
		provider:   provider,
		userID:     userID,
		email:      email,
		name:       name,
		metadata:   make(map[string]string),
		createdAt:  now,
		updatedAt:  now,
	}
}

// Getters
func (c *Customer) ID() string           { return c.id }
func (c *Customer) ProviderID() string   { return c.providerID }
func (c *Customer) Provider() Provider   { return c.provider }
func (c *Customer) UserID() string       { return c.userID }
func (c *Customer) Email() string        { return c.email }
func (c *Customer) Name() string         { return c.name }
func (c *Customer) Phone() *string       { return c.phone }
func (c *Customer) Metadata() map[string]string { return c.metadata }
func (c *Customer) CreatedAt() time.Time { return c.createdAt }
func (c *Customer) UpdatedAt() time.Time { return c.updatedAt }

// Setters (business logic)
func (c *Customer) UpdateEmail(email string) {
	c.email = email
	c.updatedAt = time.Now()
}

func (c *Customer) UpdatePhone(phone string) {
	c.phone = &phone
	c.updatedAt = time.Now()
}

func (c *Customer) SetMetadata(key, value string) {
	c.metadata[key] = value
	c.updatedAt = time.Now()
}
```

#### 2. Payment Method Entity (`payment_method.go`)
```go
package payment

import "time"

// PaymentMethodType represents the type of payment method
type PaymentMethodType string

const (
	PaymentMethodTypeCreditCard  PaymentMethodType = "credit_card"
	PaymentMethodTypeDebitCard   PaymentMethodType = "debit_card"
	PaymentMethodTypeBankAccount PaymentMethodType = "bank_account"
	PaymentMethodTypeWallet      PaymentMethodType = "wallet"
)

// PaymentMethod represents a stored payment method
type PaymentMethod struct {
	id           string
	customerID   string            // FK to customer
	providerID   string            // Provider's payment method ID
	provider     Provider
	methodType   PaymentMethodType
	isDefault    bool

	// Card information (tokenized)
	cardLast4    *string
	cardBrand    *string // visa, mastercard, amex, etc.
	cardExpMonth *int
	cardExpYear  *int

	// Bank account information (tokenized)
	bankLast4    *string
	bankName     *string
	accountType  *string // checking, savings

	metadata     map[string]string
	createdAt    time.Time
	updatedAt    time.Time
}

// Constructor for credit/debit card
func NewCardPaymentMethod(
	id string,
	customerID string,
	providerID string,
	provider Provider,
	methodType PaymentMethodType,
	last4 string,
	brand string,
	expMonth int,
	expYear int,
) *PaymentMethod {
	now := time.Now()
	return &PaymentMethod{
		id:           id,
		customerID:   customerID,
		providerID:   providerID,
		provider:     provider,
		methodType:   methodType,
		isDefault:    false,
		cardLast4:    &last4,
		cardBrand:    &brand,
		cardExpMonth: &expMonth,
		cardExpYear:  &expYear,
		metadata:     make(map[string]string),
		createdAt:    now,
		updatedAt:    now,
	}
}

// Constructor for bank account
func NewBankAccountPaymentMethod(
	id string,
	customerID string,
	providerID string,
	provider Provider,
	last4 string,
	bankName string,
	accountType string,
) *PaymentMethod {
	now := time.Now()
	return &PaymentMethod{
		id:          id,
		customerID:  customerID,
		providerID:  providerID,
		provider:    provider,
		methodType:  PaymentMethodTypeBankAccount,
		isDefault:   false,
		bankLast4:   &last4,
		bankName:    &bankName,
		accountType: &accountType,
		metadata:    make(map[string]string),
		createdAt:   now,
		updatedAt:   now,
	}
}

// Getters
func (pm *PaymentMethod) ID() string                  { return pm.id }
func (pm *PaymentMethod) CustomerID() string          { return pm.customerID }
func (pm *PaymentMethod) ProviderID() string          { return pm.providerID }
func (pm *PaymentMethod) Provider() Provider          { return pm.provider }
func (pm *PaymentMethod) MethodType() PaymentMethodType { return pm.methodType }
func (pm *PaymentMethod) IsDefault() bool             { return pm.isDefault }
func (pm *PaymentMethod) CardLast4() *string          { return pm.cardLast4 }
func (pm *PaymentMethod) CardBrand() *string          { return pm.cardBrand }
func (pm *PaymentMethod) CardExpMonth() *int          { return pm.cardExpMonth }
func (pm *PaymentMethod) CardExpYear() *int           { return pm.cardExpYear }
func (pm *PaymentMethod) BankLast4() *string          { return pm.bankLast4 }
func (pm *PaymentMethod) BankName() *string           { return pm.bankName }
func (pm *PaymentMethod) AccountType() *string        { return pm.accountType }
func (pm *PaymentMethod) Metadata() map[string]string { return pm.metadata }
func (pm *PaymentMethod) CreatedAt() time.Time        { return pm.createdAt }
func (pm *PaymentMethod) UpdatedAt() time.Time        { return pm.updatedAt }

// Business logic
func (pm *PaymentMethod) SetAsDefault() {
	pm.isDefault = true
	pm.updatedAt = time.Now()
}

func (pm *PaymentMethod) UnsetDefault() {
	pm.isDefault = false
	pm.updatedAt = time.Now()
}

func (pm *PaymentMethod) IsCard() bool {
	return pm.methodType == PaymentMethodTypeCreditCard ||
		pm.methodType == PaymentMethodTypeDebitCard
}

func (pm *PaymentMethod) IsBankAccount() bool {
	return pm.methodType == PaymentMethodTypeBankAccount
}

func (pm *PaymentMethod) IsExpired() bool {
	if !pm.IsCard() || pm.cardExpMonth == nil || pm.cardExpYear == nil {
		return false
	}
	now := time.Now()
	expDate := time.Date(*pm.cardExpYear, time.Month(*pm.cardExpMonth+1), 0, 0, 0, 0, 0, time.UTC)
	return now.After(expDate)
}
```

#### 3. Payment State Machine (`payment_state.go`)
```go
package payment

import "fmt"

// PaymentState represents the state of a payment
type PaymentState string

const (
	PaymentStatePending    PaymentState = "pending"
	PaymentStateAuthorized PaymentState = "authorized"
	PaymentStateCaptured   PaymentState = "captured"
	PaymentStateFailed     PaymentState = "failed"
	PaymentStateRefunded   PaymentState = "refunded"
	PaymentStateVoided     PaymentState = "voided"
	PaymentStateReversed   PaymentState = "reversed"
)

// Valid state transitions
var validTransitions = map[PaymentState][]PaymentState{
	PaymentStatePending: {
		PaymentStateAuthorized,
		PaymentStateFailed,
	},
	PaymentStateAuthorized: {
		PaymentStateCaptured,
		PaymentStateVoided,
		PaymentStateFailed,
	},
	PaymentStateCaptured: {
		PaymentStateRefunded,
		PaymentStateReversed,
	},
	PaymentStateFailed:   {},
	PaymentStateRefunded: {},
	PaymentStateVoided:   {},
	PaymentStateReversed: {},
}

// CanTransitionTo checks if a state transition is valid
func (s PaymentState) CanTransitionTo(next PaymentState) bool {
	allowedStates, exists := validTransitions[s]
	if !exists {
		return false
	}

	for _, allowed := range allowedStates {
		if allowed == next {
			return true
		}
	}
	return false
}

// ValidateTransition returns an error if transition is invalid
func (s PaymentState) ValidateTransition(next PaymentState) error {
	if !s.CanTransitionTo(next) {
		return fmt.Errorf("invalid state transition from %s to %s", s, next)
	}
	return nil
}

// IsTerminal returns true if this is a final state
func (s PaymentState) IsTerminal() bool {
	return s == PaymentStateFailed ||
		s == PaymentStateRefunded ||
		s == PaymentStateVoided ||
		s == PaymentStateReversed
}
```

#### 4. Transaction Type (`transaction_type.go`)
```go
package payment

// TransactionType represents the type of payment transaction
type TransactionType string

const (
	TransactionTypeAuthorize TransactionType = "authorize"
	TransactionTypeCapture   TransactionType = "capture"
	TransactionTypeRefund    TransactionType = "refund"
	TransactionTypeVoid      TransactionType = "void"
	TransactionTypeReverse   TransactionType = "reverse"
	TransactionTypePurchase  TransactionType = "purchase" // Auth + Capture in one
)

// RequiresAmount returns true if this transaction type requires an amount
func (t TransactionType) RequiresAmount() bool {
	return t == TransactionTypeAuthorize ||
		t == TransactionTypeCapture ||
		t == TransactionTypeRefund ||
		t == TransactionTypePurchase
}

// AllowsPartialAmount returns true if partial amounts are allowed
func (t TransactionType) AllowsPartialAmount() bool {
	return t == TransactionTypeCapture || t == TransactionTypeRefund
}
```

#### 5. Provider Type (`provider_type.go`)
```go
package payment

// Provider represents a payment provider
type Provider string

const (
	ProviderStripe      Provider = "stripe"
	ProviderCyberSource Provider = "cybersource"
	ProviderPlaid       Provider = "plaid"
	ProviderGoat        Provider = "goat"
	ProviderMock        Provider = "mock" // For testing
)

// IsValid checks if provider is supported
func (p Provider) IsValid() bool {
	switch p {
	case ProviderStripe, ProviderCyberSource, ProviderPlaid, ProviderGoat, ProviderMock:
		return true
	default:
		return false
	}
}

// String returns string representation
func (p Provider) String() string {
	return string(p)
}
```

#### 6. Payment Entity (`payment.go`)
```go
package payment

import (
	"time"
)

// Payment represents a payment transaction
type Payment struct {
	id                string
	customerID        string
	paymentMethodID   string
	provider          Provider
	providerPaymentID string        // Provider's transaction ID

	// Transaction details
	transactionType   TransactionType
	state             PaymentState
	amount            int64  // Amount in smallest currency unit (cents)
	currency          string // ISO 4217 currency code

	// Authorization tracking
	authorizedAmount  int64
	capturedAmount    int64
	refundedAmount    int64

	// Idempotency
	idempotencyKey    string

	// Parent transaction (for captures, refunds, etc.)
	parentPaymentID   *string

	// Provider response
	providerResponse  map[string]interface{}

	// Error tracking
	errorCode         *string
	errorMessage      *string

	// Metadata
	description       string
	metadata          map[string]string

	// Timestamps
	authorizedAt      *time.Time
	capturedAt        *time.Time
	failedAt          *time.Time
	createdAt         time.Time
	updatedAt         time.Time
}

// Constructor
func NewPayment(
	id string,
	customerID string,
	paymentMethodID string,
	provider Provider,
	transactionType TransactionType,
	amount int64,
	currency string,
	idempotencyKey string,
) *Payment {
	now := time.Now()
	return &Payment{
		id:              id,
		customerID:      customerID,
		paymentMethodID: paymentMethodID,
		provider:        provider,
		transactionType: transactionType,
		state:           PaymentStatePending,
		amount:          amount,
		currency:        currency,
		idempotencyKey:  idempotencyKey,
		metadata:        make(map[string]string),
		createdAt:       now,
		updatedAt:       now,
	}
}

// Getters
func (p *Payment) ID() string                          { return p.id }
func (p *Payment) CustomerID() string                  { return p.customerID }
func (p *Payment) PaymentMethodID() string             { return p.paymentMethodID }
func (p *Payment) Provider() Provider                  { return p.provider }
func (p *Payment) ProviderPaymentID() string           { return p.providerPaymentID }
func (p *Payment) TransactionType() TransactionType    { return p.transactionType }
func (p *Payment) State() PaymentState                 { return p.state }
func (p *Payment) Amount() int64                       { return p.amount }
func (p *Payment) Currency() string                    { return p.currency }
func (p *Payment) AuthorizedAmount() int64             { return p.authorizedAmount }
func (p *Payment) CapturedAmount() int64               { return p.capturedAmount }
func (p *Payment) RefundedAmount() int64               { return p.refundedAmount }
func (p *Payment) IdempotencyKey() string              { return p.idempotencyKey }
func (p *Payment) ParentPaymentID() *string            { return p.parentPaymentID }
func (p *Payment) ProviderResponse() map[string]interface{} { return p.providerResponse }
func (p *Payment) ErrorCode() *string                  { return p.errorCode }
func (p *Payment) ErrorMessage() *string               { return p.errorMessage }
func (p *Payment) Description() string                 { return p.description }
func (p *Payment) Metadata() map[string]string         { return p.metadata }
func (p *Payment) AuthorizedAt() *time.Time            { return p.authorizedAt }
func (p *Payment) CapturedAt() *time.Time              { return p.capturedAt }
func (p *Payment) FailedAt() *time.Time                { return p.failedAt }
func (p *Payment) CreatedAt() time.Time                { return p.createdAt }
func (p *Payment) UpdatedAt() time.Time                { return p.updatedAt }

// Business logic methods
func (p *Payment) MarkAuthorized(providerPaymentID string, amount int64) error {
	if err := p.state.ValidateTransition(PaymentStateAuthorized); err != nil {
		return err
	}

	now := time.Now()
	p.state = PaymentStateAuthorized
	p.providerPaymentID = providerPaymentID
	p.authorizedAmount = amount
	p.authorizedAt = &now
	p.updatedAt = now
	return nil
}

func (p *Payment) MarkCaptured(amount int64) error {
	if err := p.state.ValidateTransition(PaymentStateCaptured); err != nil {
		return err
	}

	if amount > p.authorizedAmount {
		return fmt.Errorf("capture amount %d exceeds authorized amount %d", amount, p.authorizedAmount)
	}

	now := time.Now()
	p.state = PaymentStateCaptured
	p.capturedAmount = amount
	p.capturedAt = &now
	p.updatedAt = now
	return nil
}

func (p *Payment) MarkRefunded(amount int64) error {
	if err := p.state.ValidateTransition(PaymentStateRefunded); err != nil {
		return err
	}

	if amount > p.capturedAmount {
		return fmt.Errorf("refund amount %d exceeds captured amount %d", amount, p.capturedAmount)
	}

	p.state = PaymentStateRefunded
	p.refundedAmount = amount
	p.updatedAt = time.Now()
	return nil
}

func (p *Payment) MarkVoided() error {
	if err := p.state.ValidateTransition(PaymentStateVoided); err != nil {
		return err
	}

	p.state = PaymentStateVoided
	p.updatedAt = time.Now()
	return nil
}

func (p *Payment) MarkFailed(errorCode, errorMessage string) {
	now := time.Now()
	p.state = PaymentStateFailed
	p.errorCode = &errorCode
	p.errorMessage = &errorMessage
	p.failedAt = &now
	p.updatedAt = now
}

func (p *Payment) SetProviderResponse(response map[string]interface{}) {
	p.providerResponse = response
	p.updatedAt = time.Now()
}

func (p *Payment) SetDescription(description string) {
	p.description = description
	p.updatedAt = time.Now()
}

func (p *Payment) SetMetadata(key, value string) {
	p.metadata[key] = value
	p.updatedAt = time.Now()
}

func (p *Payment) SetParentPaymentID(parentID string) {
	p.parentPaymentID = &parentID
}

// Helper methods
func (p *Payment) CanCapture() bool {
	return p.state == PaymentStateAuthorized
}

func (p *Payment) CanRefund() bool {
	return p.state == PaymentStateCaptured
}

func (p *Payment) CanVoid() bool {
	return p.state == PaymentStateAuthorized
}

func (p *Payment) RemainingCaptureAmount() int64 {
	return p.authorizedAmount - p.capturedAmount
}

func (p *Payment) RemainingRefundAmount() int64 {
	return p.capturedAmount - p.refundedAmount
}
```

#### 7. Domain Errors (`errors.go`)
```go
package payment

import "fmt"

// Domain-level errors
type DomainError struct {
	Code    string
	Message string
	Err     error
}

func (e *DomainError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *DomainError) Unwrap() error {
	return e.Err
}

// Common domain errors
var (
	ErrInvalidStateTransition = &DomainError{
		Code:    "INVALID_STATE_TRANSITION",
		Message: "invalid payment state transition",
	}

	ErrInsufficientAmount = &DomainError{
		Code:    "INSUFFICIENT_AMOUNT",
		Message: "amount exceeds available balance",
	}

	ErrPaymentMethodExpired = &DomainError{
		Code:    "PAYMENT_METHOD_EXPIRED",
		Message: "payment method has expired",
	}

	ErrInvalidCurrency = &DomainError{
		Code:    "INVALID_CURRENCY",
		Message: "unsupported currency code",
	}

	ErrDuplicateIdempotencyKey = &DomainError{
		Code:    "DUPLICATE_IDEMPOTENCY_KEY",
		Message: "idempotency key already used",
	}
)
```

---

## Dependency Injection Interfaces

### File Structure
```
internal/core/di/
├── repositories/
│   ├── i_payment_repository.go
│   ├── i_customer_repository.go
│   └── i_payment_method_repository.go
└── services/
    ├── i_payment_service.go
    └── i_payment_provider.go
```

### Repository Interfaces

#### Payment Repository (`i_payment_repository.go`)
```go
package repositories

import (
	"context"
	"database/sql"
	"math-ai.com/math-ai/internal/core/domain/payment"
)

type IPaymentRepository interface {
	// Create a new payment record
	Create(ctx context.Context, tx *sql.Tx, p *payment.Payment) (int64, error)

	// Find payment by ID
	FindByID(ctx context.Context, id string) (*payment.Payment, error)

	// Find payment by idempotency key
	FindByIdempotencyKey(ctx context.Context, key string) (*payment.Payment, error)

	// Find payment by provider payment ID
	FindByProviderPaymentID(ctx context.Context, providerPaymentID string) (*payment.Payment, error)

	// Find all payments for a customer
	FindByCustomerID(ctx context.Context, customerID string, limit, offset int) ([]*payment.Payment, error)

	// Update payment
	Update(ctx context.Context, tx *sql.Tx, p *payment.Payment) (int64, error)

	// Find child payments (captures, refunds) of a parent
	FindChildPayments(ctx context.Context, parentPaymentID string) ([]*payment.Payment, error)
}
```

#### Customer Repository (`i_customer_repository.go`)
```go
package repositories

import (
	"context"
	"database/sql"
	"math-ai.com/math-ai/internal/core/domain/payment"
)

type ICustomerRepository interface {
	// Create a new customer
	Create(ctx context.Context, tx *sql.Tx, customer *payment.Customer) (int64, error)

	// Find customer by ID
	FindByID(ctx context.Context, id string) (*payment.Customer, error)

	// Find customer by provider ID
	FindByProviderID(ctx context.Context, provider payment.Provider, providerID string) (*payment.Customer, error)

	// Find customer by user ID and provider
	FindByUserIDAndProvider(ctx context.Context, userID string, provider payment.Provider) (*payment.Customer, error)

	// Update customer
	Update(ctx context.Context, tx *sql.Tx, customer *payment.Customer) (int64, error)

	// Delete customer (soft delete)
	Delete(ctx context.Context, tx *sql.Tx, id string) (int64, error)
}
```

#### Payment Method Repository (`i_payment_method_repository.go`)
```go
package repositories

import (
	"context"
	"database/sql"
	"math-ai.com/math-ai/internal/core/domain/payment"
)

type IPaymentMethodRepository interface {
	// Create a new payment method
	Create(ctx context.Context, tx *sql.Tx, method *payment.PaymentMethod) (int64, error)

	// Find payment method by ID
	FindByID(ctx context.Context, id string) (*payment.PaymentMethod, error)

	// Find all payment methods for a customer
	FindByCustomerID(ctx context.Context, customerID string) ([]*payment.PaymentMethod, error)

	// Find default payment method for a customer
	FindDefaultByCustomerID(ctx context.Context, customerID string) (*payment.PaymentMethod, error)

	// Update payment method
	Update(ctx context.Context, tx *sql.Tx, method *payment.PaymentMethod) (int64, error)

	// Delete payment method (soft delete)
	Delete(ctx context.Context, tx *sql.Tx, id string) (int64, error)

	// Set a payment method as default (unsets others)
	SetAsDefault(ctx context.Context, tx *sql.Tx, customerID string, methodID string) error
}
```

### Service Interfaces

#### Payment Provider Interface (`i_payment_provider.go`)
```go
package services

import (
	"context"
	"math-ai.com/math-ai/internal/core/domain/payment"
)

// IPaymentProvider defines the interface that all payment providers must implement
type IPaymentProvider interface {
	// Provider identification
	GetProviderName() payment.Provider

	// Customer Management
	CreateCustomer(ctx context.Context, req *CreateCustomerRequest) (*CustomerResponse, error)
	GetCustomer(ctx context.Context, providerCustomerID string) (*CustomerResponse, error)
	UpdateCustomer(ctx context.Context, providerCustomerID string, req *UpdateCustomerRequest) (*CustomerResponse, error)

	// Payment Method Management
	AddPaymentMethod(ctx context.Context, req *AddPaymentMethodRequest) (*PaymentMethodResponse, error)
	GetPaymentMethod(ctx context.Context, providerMethodID string) (*PaymentMethodResponse, error)
	DeletePaymentMethod(ctx context.Context, providerMethodID string) error

	// Payment Operations
	Authorize(ctx context.Context, req *AuthorizeRequest) (*PaymentResponse, error)
	Capture(ctx context.Context, req *CaptureRequest) (*PaymentResponse, error)
	Refund(ctx context.Context, req *RefundRequest) (*PaymentResponse, error)
	Void(ctx context.Context, req *VoidRequest) (*PaymentResponse, error)

	// Query Operations
	GetPayment(ctx context.Context, providerPaymentID string) (*PaymentResponse, error)

	// Webhook verification
	VerifyWebhookSignature(payload []byte, signature string) (bool, error)
	ParseWebhookEvent(payload []byte) (*WebhookEvent, error)
}

// Request/Response DTOs for provider interface

type CreateCustomerRequest struct {
	Email    string
	Name     string
	Phone    *string
	Metadata map[string]string
}

type UpdateCustomerRequest struct {
	Email    *string
	Name     *string
	Phone    *string
	Metadata map[string]string
}

type CustomerResponse struct {
	ProviderCustomerID string
	Email              string
	Name               string
	Phone              *string
	Metadata           map[string]string
	RawResponse        map[string]interface{}
}

type AddPaymentMethodRequest struct {
	ProviderCustomerID string
	Token              string // Provider-specific token (e.g., Stripe token, payment method ID)
	SetAsDefault       bool
	Metadata           map[string]string
}

type PaymentMethodResponse struct {
	ProviderMethodID string
	MethodType       payment.PaymentMethodType

	// Card details
	CardLast4    *string
	CardBrand    *string
	CardExpMonth *int
	CardExpYear  *int

	// Bank account details
	BankLast4   *string
	BankName    *string
	AccountType *string

	IsDefault   bool
	Metadata    map[string]string
	RawResponse map[string]interface{}
}

type AuthorizeRequest struct {
	ProviderCustomerID string
	ProviderMethodID   string
	Amount             int64
	Currency           string
	Description        string
	IdempotencyKey     string
	Metadata           map[string]string
}

type CaptureRequest struct {
	ProviderPaymentID string
	Amount            *int64 // Nil = capture full amount
	IdempotencyKey    string
}

type RefundRequest struct {
	ProviderPaymentID string
	Amount            *int64 // Nil = refund full amount
	Reason            *string
	IdempotencyKey    string
}

type VoidRequest struct {
	ProviderPaymentID string
	IdempotencyKey    string
}

type PaymentResponse struct {
	ProviderPaymentID string
	State             payment.PaymentState
	Amount            int64
	Currency          string
	AuthorizedAmount  int64
	CapturedAmount    int64
	RefundedAmount    int64
	ErrorCode         *string
	ErrorMessage      *string
	RawResponse       map[string]interface{}
}

type WebhookEvent struct {
	EventType         string // e.g., "payment.succeeded", "payment.failed"
	ProviderPaymentID string
	State             payment.PaymentState
	RawPayload        map[string]interface{}
}
```

#### Payment Service Interface (`i_payment_service.go`)
```go
package services

import (
	"context"
	"math-ai.com/math-ai/internal/applications/dto"
	"math-ai.com/math-ai/internal/core/domain/payment"
	"math-ai.com/math-ai/internal/shared/constant/status"
)

type IPaymentService interface {
	// Customer operations
	CreateCustomer(ctx context.Context, req *dto.CreatePaymentCustomerRequest) (status.Code, *dto.PaymentCustomerResponse, error)
	GetCustomer(ctx context.Context, req *dto.GetPaymentCustomerRequest) (status.Code, *dto.PaymentCustomerResponse, error)

	// Payment method operations
	AddCreditCard(ctx context.Context, req *dto.AddCreditCardRequest) (status.Code, *dto.PaymentMethodResponse, error)
	AddDebitCard(ctx context.Context, req *dto.AddDebitCardRequest) (status.Code, *dto.PaymentMethodResponse, error)
	AddBankAccount(ctx context.Context, req *dto.AddBankAccountRequest) (status.Code, *dto.PaymentMethodResponse, error)
	GetPaymentMethods(ctx context.Context, req *dto.GetPaymentMethodsRequest) (status.Code, []*dto.PaymentMethodResponse, error)
	DeletePaymentMethod(ctx context.Context, req *dto.DeletePaymentMethodRequest) (status.Code, error)
	SetDefaultPaymentMethod(ctx context.Context, req *dto.SetDefaultPaymentMethodRequest) (status.Code, error)

	// Payment operations
	Authorize(ctx context.Context, req *dto.AuthorizePaymentRequest) (status.Code, *dto.PaymentResponse, error)
	Capture(ctx context.Context, req *dto.CapturePaymentRequest) (status.Code, *dto.PaymentResponse, error)
	Refund(ctx context.Context, req *dto.RefundPaymentRequest) (status.Code, *dto.PaymentResponse, error)
	Void(ctx context.Context, req *dto.VoidPaymentRequest) (status.Code, *dto.PaymentResponse, error)

	// Query operations
	GetPayment(ctx context.Context, req *dto.GetPaymentRequest) (status.Code, *dto.PaymentResponse, error)
	GetPaymentHistory(ctx context.Context, req *dto.GetPaymentHistoryRequest) (status.Code, []*dto.PaymentResponse, error)

	// Webhook handling
	HandleWebhook(ctx context.Context, provider payment.Provider, payload []byte, signature string) error
}
```

---

## Application Layer

### File Structure
```
internal/applications/
├── dto/
│   └── payment_dto.go
├── services/
│   └── payment_service.go
└── validators/
    └── payment_validator.go
```

### DTOs (`payment_dto.go`)
```go
package dto

// Create customer request
type CreatePaymentCustomerRequest struct {
	UserID   string            `json:"user_id" validate:"required"`
	Provider string            `json:"provider" validate:"required,oneof=stripe cybersource plaid goat"`
	Email    string            `json:"email" validate:"required,email"`
	Name     string            `json:"name" validate:"required"`
	Phone    *string           `json:"phone,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// Add credit card request
type AddCreditCardRequest struct {
	CustomerID   string            `json:"customer_id" validate:"required"`
	Token        string            `json:"token" validate:"required"` // Provider token
	SetAsDefault bool              `json:"set_as_default"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// Add bank account request
type AddBankAccountRequest struct {
	CustomerID   string            `json:"customer_id" validate:"required"`
	Token        string            `json:"token" validate:"required"` // Provider token
	SetAsDefault bool              `json:"set_as_default"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// Authorize payment request
type AuthorizePaymentRequest struct {
	CustomerID      string            `json:"customer_id" validate:"required"`
	PaymentMethodID *string           `json:"payment_method_id,omitempty"` // Use default if nil
	Amount          int64             `json:"amount" validate:"required,gt=0"`
	Currency        string            `json:"currency" validate:"required,len=3"`
	Description     string            `json:"description"`
	IdempotencyKey  string            `json:"idempotency_key" validate:"required"`
	Metadata        map[string]string `json:"metadata,omitempty"`
}

// Capture payment request
type CapturePaymentRequest struct {
	PaymentID      string  `json:"payment_id" validate:"required"`
	Amount         *int64  `json:"amount,omitempty"` // Capture full if nil
	IdempotencyKey string  `json:"idempotency_key" validate:"required"`
}

// Refund payment request
type RefundPaymentRequest struct {
	PaymentID      string  `json:"payment_id" validate:"required"`
	Amount         *int64  `json:"amount,omitempty"` // Refund full if nil
	Reason         *string `json:"reason,omitempty"`
	IdempotencyKey string  `json:"idempotency_key" validate:"required"`
}

// Void payment request
type VoidPaymentRequest struct {
	PaymentID      string `json:"payment_id" validate:"required"`
	IdempotencyKey string `json:"idempotency_key" validate:"required"`
}

// Payment response
type PaymentResponse struct {
	ID                string                 `json:"id"`
	CustomerID        string                 `json:"customer_id"`
	PaymentMethodID   string                 `json:"payment_method_id"`
	Provider          string                 `json:"provider"`
	ProviderPaymentID string                 `json:"provider_payment_id"`
	TransactionType   string                 `json:"transaction_type"`
	State             string                 `json:"state"`
	Amount            int64                  `json:"amount"`
	Currency          string                 `json:"currency"`
	AuthorizedAmount  int64                  `json:"authorized_amount"`
	CapturedAmount    int64                  `json:"captured_amount"`
	RefundedAmount    int64                  `json:"refunded_amount"`
	Description       string                 `json:"description"`
	ErrorCode         *string                `json:"error_code,omitempty"`
	ErrorMessage      *string                `json:"error_message,omitempty"`
	Metadata          map[string]string      `json:"metadata,omitempty"`
	CreatedAt         string                 `json:"created_at"`
	UpdatedAt         string                 `json:"updated_at"`
}

// Payment method response
type PaymentMethodResponse struct {
	ID           string            `json:"id"`
	CustomerID   string            `json:"customer_id"`
	Provider     string            `json:"provider"`
	MethodType   string            `json:"method_type"`
	IsDefault    bool              `json:"is_default"`
	CardLast4    *string           `json:"card_last4,omitempty"`
	CardBrand    *string           `json:"card_brand,omitempty"`
	CardExpMonth *int              `json:"card_exp_month,omitempty"`
	CardExpYear  *int              `json:"card_exp_year,omitempty"`
	BankLast4    *string           `json:"bank_last4,omitempty"`
	BankName     *string           `json:"bank_name,omitempty"`
	AccountType  *string           `json:"account_type,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	CreatedAt    string            `json:"created_at"`
}

// Customer response
type PaymentCustomerResponse struct {
	ID         string            `json:"id"`
	UserID     string            `json:"user_id"`
	Provider   string            `json:"provider"`
	ProviderID string            `json:"provider_id"`
	Email      string            `json:"email"`
	Name       string            `json:"name"`
	Phone      *string           `json:"phone,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
	CreatedAt  string            `json:"created_at"`
}
```

### Payment Service Implementation (`payment_service.go`)
```go
package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"math-ai.com/math-ai/internal/applications/dto"
	"math-ai.com/math-ai/internal/applications/validators"
	diRepo "math-ai.com/math-ai/internal/core/di/repositories"
	diSvc "math-ai.com/math-ai/internal/core/di/services"
	"math-ai.com/math-ai/internal/core/domain/payment"
	"math-ai.com/math-ai/internal/shared/constant/status"
	"math-ai.com/math-ai/internal/shared/logger"
)

type paymentService struct {
	validator           validators.IPaymentValidator
	paymentRepo         diRepo.IPaymentRepository
	customerRepo        diRepo.ICustomerRepository
	paymentMethodRepo   diRepo.IPaymentMethodRepository
	providerFactory     IPaymentProviderFactory
}

func NewPaymentService(
	validator validators.IPaymentValidator,
	paymentRepo diRepo.IPaymentRepository,
	customerRepo diRepo.ICustomerRepository,
	paymentMethodRepo diRepo.IPaymentMethodRepository,
	providerFactory IPaymentProviderFactory,
) diSvc.IPaymentService {
	return &paymentService{
		validator:         validator,
		paymentRepo:       paymentRepo,
		customerRepo:      customerRepo,
		paymentMethodRepo: paymentMethodRepo,
		providerFactory:   providerFactory,
	}
}

// Authorize creates an authorization hold on the payment method
func (s *paymentService) Authorize(ctx context.Context, req *dto.AuthorizePaymentRequest) (status.Code, *dto.PaymentResponse, error) {
	log := logger.GetLogger(ctx)

	// 1. Validate request
	if statusCode, err := s.validator.ValidateAuthorizeRequest(req); err != nil {
		return statusCode, nil, err
	}

	// 2. Check idempotency
	existingPayment, err := s.paymentRepo.FindByIdempotencyKey(ctx, req.IdempotencyKey)
	if err == nil && existingPayment != nil {
		log.Infof("Returning cached payment for idempotency key: %s", req.IdempotencyKey)
		return status.OK, s.buildPaymentResponse(existingPayment), nil
	}

	// 3. Load customer
	customer, err := s.customerRepo.FindByID(ctx, req.CustomerID)
	if err != nil {
		log.Errorf("Customer not found: %v", err)
		return status.CUSTOMER_NOT_FOUND, nil, fmt.Errorf("customer not found")
	}

	// 4. Load payment method (or use default)
	var paymentMethod *payment.PaymentMethod
	if req.PaymentMethodID != nil {
		paymentMethod, err = s.paymentMethodRepo.FindByID(ctx, *req.PaymentMethodID)
		if err != nil {
			return status.PAYMENT_METHOD_NOT_FOUND, nil, fmt.Errorf("payment method not found")
		}
	} else {
		paymentMethod, err = s.paymentMethodRepo.FindDefaultByCustomerID(ctx, req.CustomerID)
		if err != nil || paymentMethod == nil {
			return status.PAYMENT_METHOD_NOT_FOUND, nil, fmt.Errorf("no default payment method found")
		}
	}

	// 5. Check if payment method is expired
	if paymentMethod.IsExpired() {
		return status.PAYMENT_METHOD_EXPIRED, nil, fmt.Errorf("payment method has expired")
	}

	// 6. Get provider
	provider, err := s.providerFactory.GetProvider(customer.Provider())
	if err != nil {
		log.Errorf("Failed to get provider: %v", err)
		return status.INTERNAL_ERROR, nil, err
	}

	// 7. Create domain payment entity
	paymentID := uuid.New().String()
	domainPayment := payment.NewPayment(
		paymentID,
		customer.ID(),
		paymentMethod.ID(),
		customer.Provider(),
		payment.TransactionTypeAuthorize,
		req.Amount,
		req.Currency,
		req.IdempotencyKey,
	)
	domainPayment.SetDescription(req.Description)
	for k, v := range req.Metadata {
		domainPayment.SetMetadata(k, v)
	}

	// 8. Call provider to authorize
	providerReq := &diSvc.AuthorizeRequest{
		ProviderCustomerID: customer.ProviderID(),
		ProviderMethodID:   paymentMethod.ProviderID(),
		Amount:             req.Amount,
		Currency:           req.Currency,
		Description:        req.Description,
		IdempotencyKey:     req.IdempotencyKey,
		Metadata:           req.Metadata,
	}

	providerResp, err := provider.Authorize(ctx, providerReq)
	if err != nil {
		log.Errorf("Provider authorization failed: %v", err)
		domainPayment.MarkFailed("provider_error", err.Error())

		// Save failed payment
		if _, saveErr := s.paymentRepo.Create(ctx, nil, domainPayment); saveErr != nil {
			log.Errorf("Failed to save payment: %v", saveErr)
		}

		return status.PAYMENT_FAILED, s.buildPaymentResponse(domainPayment), err
	}

	// 9. Update domain entity with provider response
	if err := domainPayment.MarkAuthorized(providerResp.ProviderPaymentID, providerResp.AuthorizedAmount); err != nil {
		log.Errorf("Failed to mark payment as authorized: %v", err)
		return status.INTERNAL_ERROR, nil, err
	}
	domainPayment.SetProviderResponse(providerResp.RawResponse)

	// 10. Persist payment
	if _, err := s.paymentRepo.Create(ctx, nil, domainPayment); err != nil {
		log.Errorf("Failed to save payment: %v", err)
		return status.INTERNAL_ERROR, nil, err
	}

	log.Infof("Payment authorized successfully: %s", paymentID)
	return status.OK, s.buildPaymentResponse(domainPayment), nil
}

// Capture captures a previously authorized payment
func (s *paymentService) Capture(ctx context.Context, req *dto.CapturePaymentRequest) (status.Code, *dto.PaymentResponse, error) {
	log := logger.GetLogger(ctx)

	// 1. Validate request
	if statusCode, err := s.validator.ValidateCaptureRequest(req); err != nil {
		return statusCode, nil, err
	}

	// 2. Load payment
	domainPayment, err := s.paymentRepo.FindByID(ctx, req.PaymentID)
	if err != nil {
		return status.PAYMENT_NOT_FOUND, nil, fmt.Errorf("payment not found")
	}

	// 3. Verify payment can be captured
	if !domainPayment.CanCapture() {
		return status.INVALID_PAYMENT_STATE, nil, fmt.Errorf("payment cannot be captured in state: %s", domainPayment.State())
	}

	// 4. Determine capture amount
	captureAmount := domainPayment.RemainingCaptureAmount()
	if req.Amount != nil {
		if *req.Amount > captureAmount {
			return status.INVALID_AMOUNT, nil, fmt.Errorf("capture amount exceeds authorized amount")
		}
		captureAmount = *req.Amount
	}

	// 5. Get provider
	provider, err := s.providerFactory.GetProvider(domainPayment.Provider())
	if err != nil {
		return status.INTERNAL_ERROR, nil, err
	}

	// 6. Call provider to capture
	providerReq := &diSvc.CaptureRequest{
		ProviderPaymentID: domainPayment.ProviderPaymentID(),
		Amount:            &captureAmount,
		IdempotencyKey:    req.IdempotencyKey,
	}

	providerResp, err := provider.Capture(ctx, providerReq)
	if err != nil {
		log.Errorf("Provider capture failed: %v", err)
		return status.PAYMENT_FAILED, nil, err
	}

	// 7. Update domain entity
	if err := domainPayment.MarkCaptured(providerResp.CapturedAmount); err != nil {
		return status.INTERNAL_ERROR, nil, err
	}
	domainPayment.SetProviderResponse(providerResp.RawResponse)

	// 8. Persist changes
	if _, err := s.paymentRepo.Update(ctx, nil, domainPayment); err != nil {
		log.Errorf("Failed to update payment: %v", err)
		return status.INTERNAL_ERROR, nil, err
	}

	log.Infof("Payment captured successfully: %s", req.PaymentID)
	return status.OK, s.buildPaymentResponse(domainPayment), nil
}

// Refund refunds a captured payment
func (s *paymentService) Refund(ctx context.Context, req *dto.RefundPaymentRequest) (status.Code, *dto.PaymentResponse, error) {
	log := logger.GetLogger(ctx)

	// Similar implementation pattern to Capture...
	// Load payment, validate state, call provider, update domain, persist

	return status.OK, nil, nil
}

// Void voids an authorized (but not captured) payment
func (s *paymentService) Void(ctx context.Context, req *dto.VoidPaymentRequest) (status.Code, *dto.PaymentResponse, error) {
	log := logger.GetLogger(ctx)

	// Similar implementation pattern...

	return status.OK, nil, nil
}

// Helper: Build DTO response from domain entity
func (s *paymentService) buildPaymentResponse(p *payment.Payment) *dto.PaymentResponse {
	return &dto.PaymentResponse{
		ID:                p.ID(),
		CustomerID:        p.CustomerID(),
		PaymentMethodID:   p.PaymentMethodID(),
		Provider:          p.Provider().String(),
		ProviderPaymentID: p.ProviderPaymentID(),
		TransactionType:   string(p.TransactionType()),
		State:             string(p.State()),
		Amount:            p.Amount(),
		Currency:          p.Currency(),
		AuthorizedAmount:  p.AuthorizedAmount(),
		CapturedAmount:    p.CapturedAmount(),
		RefundedAmount:    p.RefundedAmount(),
		Description:       p.Description(),
		ErrorCode:         p.ErrorCode(),
		ErrorMessage:      p.ErrorMessage(),
		Metadata:          p.Metadata(),
		CreatedAt:         p.CreatedAt().Format(time.RFC3339),
		UpdatedAt:         p.UpdatedAt().Format(time.RFC3339),
	}
}

// IPaymentProviderFactory creates provider instances
type IPaymentProviderFactory interface {
	GetProvider(provider payment.Provider) (diSvc.IPaymentProvider, error)
}
```

---

## Adapter Layer

### File Structure
```
internal/driven-adapter/external/payment_provider/
├── factory.go                    # Provider factory
├── base_provider.go              # Base provider with common logic
├── stripe/
│   ├── stripe_provider.go
│   ├── stripe_mapper.go          # Maps Stripe responses to domain
│   └── stripe_client.go          # HTTP client wrapper
├── cybersource/
│   ├── cybersource_provider.go
│   ├── cybersource_mapper.go
│   └── cybersource_client.go
├── plaid/
│   ├── plaid_provider.go
│   ├── plaid_mapper.go
│   └── plaid_client.go
└── mock/
    └── mock_provider.go          # For testing
```

### Provider Factory (`factory.go`)
```go
package payment_provider

import (
	"fmt"
	"math-ai.com/math-ai/internal/core/domain/payment"
	diSvc "math-ai.com/math-ai/internal/core/di/services"
	"math-ai.com/math-ai/internal/driven-adapter/external/payment_provider/stripe"
	"math-ai.com/math-ai/internal/driven-adapter/external/payment_provider/cybersource"
	"math-ai.com/math-ai/internal/shared/config"
	"math-ai.com/math-ai/internal/shared/http_client"
)

type PaymentProviderFactory struct {
	config     *config.Env
	httpClient *http_client.Client
	providers  map[payment.Provider]diSvc.IPaymentProvider
}

func NewPaymentProviderFactory(config *config.Env, httpClient *http_client.Client) *PaymentProviderFactory {
	return &PaymentProviderFactory{
		config:     config,
		httpClient: httpClient,
		providers:  make(map[payment.Provider]diSvc.IPaymentProvider),
	}
}

func (f *PaymentProviderFactory) GetProvider(provider payment.Provider) (diSvc.IPaymentProvider, error) {
	// Return cached provider if exists
	if p, exists := f.providers[provider]; exists {
		return p, nil
	}

	// Create new provider instance
	var p diSvc.IPaymentProvider
	var err error

	switch provider {
	case payment.ProviderStripe:
		p, err = stripe.NewStripeProvider(f.config.StripeConfig, f.httpClient)
	case payment.ProviderCyberSource:
		p, err = cybersource.NewCyberSourceProvider(f.config.CyberSourceConfig, f.httpClient)
	case payment.ProviderPlaid:
		p, err = plaid.NewPlaidProvider(f.config.PlaidConfig, f.httpClient)
	// ... other providers
	default:
		return nil, fmt.Errorf("unsupported payment provider: %s", provider)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create provider %s: %w", provider, err)
	}

	// Cache provider
	f.providers[provider] = p
	return p, nil
}
```

---

## Provider Implementations

### Stripe Provider Example (`stripe/stripe_provider.go`)
```go
package stripe

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"math-ai.com/math-ai/internal/core/domain/payment"
	diSvc "math-ai.com/math-ai/internal/core/di/services"
	"math-ai.com/math-ai/internal/shared/config"
	httpClient "math-ai.com/math-ai/internal/shared/http_client"
)

type StripeProvider struct {
	config     *config.StripeConfig
	httpClient *httpClient.Client
	mapper     *StripeMapper
}

func NewStripeProvider(config *config.StripeConfig, client *httpClient.Client) (*StripeProvider, error) {
	return &StripeProvider{
		config:     config,
		httpClient: client,
		mapper:     NewStripeMapper(),
	}, nil
}

func (p *StripeProvider) GetProviderName() payment.Provider {
	return payment.ProviderStripe
}

// Authorize creates a payment intent with manual capture
func (p *StripeProvider) Authorize(ctx context.Context, req *diSvc.AuthorizeRequest) (*diSvc.PaymentResponse, error) {
	// Build Stripe API request
	stripeReq := map[string]interface{}{
		"customer":        req.ProviderCustomerID,
		"payment_method":  req.ProviderMethodID,
		"amount":          req.Amount,
		"currency":        req.Currency,
		"description":     req.Description,
		"capture_method":  "manual", // Authorization only
		"confirm":         true,
		"metadata":        req.Metadata,
	}

	// Call Stripe API using your http_client
	httpReq := &httpClient.Request{
		Method:  "POST",
		URL:     "https://api.stripe.com/v1/payment_intents",
		Headers: map[string]string{
			"Authorization": "Bearer " + p.config.SecretKey,
			"Content-Type":  "application/x-www-form-urlencoded",
			"Idempotency-Key": req.IdempotencyKey,
		},
		Body: p.buildFormBody(stripeReq),
	}

	httpResp, err := p.httpClient.Do(ctx, httpReq)
	if err != nil {
		return nil, fmt.Errorf("stripe API error: %w", err)
	}

	if httpResp.StatusCode != http.StatusOK {
		return nil, p.handleStripeError(httpResp)
	}

	// Parse Stripe response
	var stripeResp StripePaymentIntentResponse
	if err := json.Unmarshal(httpResp.Body, &stripeResp); err != nil {
		return nil, fmt.Errorf("failed to parse stripe response: %w", err)
	}

	// Map to domain response
	return p.mapper.MapPaymentIntentToResponse(&stripeResp), nil
}

// Capture captures a payment intent
func (p *StripeProvider) Capture(ctx context.Context, req *diSvc.CaptureRequest) (*diSvc.PaymentResponse, error) {
	captureReq := map[string]interface{}{}
	if req.Amount != nil {
		captureReq["amount_to_capture"] = *req.Amount
	}

	httpReq := &httpClient.Request{
		Method:  "POST",
		URL:     fmt.Sprintf("https://api.stripe.com/v1/payment_intents/%s/capture", req.ProviderPaymentID),
		Headers: map[string]string{
			"Authorization": "Bearer " + p.config.SecretKey,
			"Content-Type":  "application/x-www-form-urlencoded",
			"Idempotency-Key": req.IdempotencyKey,
		},
		Body: p.buildFormBody(captureReq),
	}

	httpResp, err := p.httpClient.Do(ctx, httpReq)
	if err != nil {
		return nil, fmt.Errorf("stripe capture error: %w", err)
	}

	if httpResp.StatusCode != http.StatusOK {
		return nil, p.handleStripeError(httpResp)
	}

	var stripeResp StripePaymentIntentResponse
	if err := json.Unmarshal(httpResp.Body, &stripeResp); err != nil {
		return nil, err
	}

	return p.mapper.MapPaymentIntentToResponse(&stripeResp), nil
}

// CreateCustomer creates a Stripe customer
func (p *StripeProvider) CreateCustomer(ctx context.Context, req *diSvc.CreateCustomerRequest) (*diSvc.CustomerResponse, error) {
	// Similar implementation...
	return nil, nil
}

// AddPaymentMethod adds a payment method to a customer
func (p *StripeProvider) AddPaymentMethod(ctx context.Context, req *diSvc.AddPaymentMethodRequest) (*diSvc.PaymentMethodResponse, error) {
	// Implementation...
	return nil, nil
}

// Helper: Build form-encoded body for Stripe
func (p *StripeProvider) buildFormBody(data map[string]interface{}) []byte {
	// Convert map to application/x-www-form-urlencoded
	// Implementation details...
	return nil
}

// Helper: Handle Stripe errors
func (p *StripeProvider) handleStripeError(resp *httpClient.Response) error {
	// Parse error response and return normalized error
	return fmt.Errorf("stripe error: status %d", resp.StatusCode)
}
```

### Stripe Mapper (`stripe/stripe_mapper.go`)
```go
package stripe

import (
	"math-ai.com/math-ai/internal/core/domain/payment"
	diSvc "math-ai.com/math-ai/internal/core/di/services"
)

type StripeMapper struct{}

func NewStripeMapper() *StripeMapper {
	return &StripeMapper{}
}

// Stripe API response structures
type StripePaymentIntentResponse struct {
	ID              string                 `json:"id"`
	Status          string                 `json:"status"`
	Amount          int64                  `json:"amount"`
	AmountCaptured  int64                  `json:"amount_captured"`
	AmountRefunded  int64                  `json:"amount_refunded"`
	Currency        string                 `json:"currency"`
	Customer        string                 `json:"customer"`
	PaymentMethod   string                 `json:"payment_method"`
	Description     string                 `json:"description"`
	LastPaymentError *StripeError          `json:"last_payment_error"`
	Metadata        map[string]string      `json:"metadata"`
	RawData         map[string]interface{} `json:"-"` // Store full response
}

type StripeError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// MapPaymentIntentToResponse normalizes Stripe PaymentIntent to domain response
func (m *StripeMapper) MapPaymentIntentToResponse(stripeResp *StripePaymentIntentResponse) *diSvc.PaymentResponse {
	resp := &diSvc.PaymentResponse{
		ProviderPaymentID: stripeResp.ID,
		Amount:            stripeResp.Amount,
		Currency:          stripeResp.Currency,
		AuthorizedAmount:  stripeResp.Amount,
		CapturedAmount:    stripeResp.AmountCaptured,
		RefundedAmount:    stripeResp.AmountRefunded,
		State:             m.mapStripeStatusToState(stripeResp.Status),
		RawResponse:       stripeResp.RawData,
	}

	if stripeResp.LastPaymentError != nil {
		code := stripeResp.LastPaymentError.Code
		msg := stripeResp.LastPaymentError.Message
		resp.ErrorCode = &code
		resp.ErrorMessage = &msg
	}

	return resp
}

// Map Stripe status to domain payment state
func (m *StripeMapper) mapStripeStatusToState(stripeStatus string) payment.PaymentState {
	switch stripeStatus {
	case "requires_payment_method", "requires_confirmation", "processing":
		return payment.PaymentStatePending
	case "requires_capture":
		return payment.PaymentStateAuthorized
	case "succeeded":
		return payment.PaymentStateCaptured
	case "canceled":
		return payment.PaymentStateVoided
	case "payment_failed":
		return payment.PaymentStateFailed
	default:
		return payment.PaymentStatePending
	}
}
```

---

## Error Handling & Normalization

### Normalized Error Response (`internal/shared/errors/payment_errors.go`)
```go
package errors

import "fmt"

// PaymentError represents a normalized payment error
type PaymentError struct {
	Code           string // Normalized error code
	Message        string // Human-readable message
	ProviderCode   string // Original provider error code
	ProviderMsg    string // Original provider message
	HTTPStatus     int    // HTTP status code
	Retryable      bool   // Can this be retried?
	DeclineReason  string // For declined transactions
}

func (e *PaymentError) Error() string {
	return fmt.Sprintf("[%s] %s (provider: %s)", e.Code, e.Message, e.ProviderCode)
}

// Common normalized error codes
const (
	ErrCodeInsufficientFunds    = "insufficient_funds"
	ErrCodeCardDeclined         = "card_declined"
	ErrCodeExpiredCard          = "expired_card"
	ErrCodeInvalidCard          = "invalid_card"
	ErrCodeAuthenticationFailed = "authentication_failed"
	ErrCodeDuplicateTransaction = "duplicate_transaction"
	ErrCodeProviderError        = "provider_error"
	ErrCodeNetworkError         = "network_error"
	ErrCodeInvalidRequest       = "invalid_request"
	ErrCodeRateLimitExceeded    = "rate_limit_exceeded"
)

// Error mapper for each provider
type IErrorMapper interface {
	MapError(providerError error, httpStatus int) *PaymentError
}
```

### Stripe Error Mapper
```go
package stripe

import (
	"math-ai.com/math-ai/internal/shared/errors"
)

type StripeErrorMapper struct{}

func (m *StripeErrorMapper) MapError(providerErr error, httpStatus int) *errors.PaymentError {
	// Parse Stripe error
	stripeErr := parseStripeError(providerErr)

	paymentErr := &errors.PaymentError{
		ProviderCode: stripeErr.Code,
		ProviderMsg:  stripeErr.Message,
		HTTPStatus:   httpStatus,
	}

	// Map Stripe error codes to normalized codes
	switch stripeErr.Code {
	case "insufficient_funds":
		paymentErr.Code = errors.ErrCodeInsufficientFunds
		paymentErr.Message = "Insufficient funds on payment method"
		paymentErr.Retryable = false

	case "card_declined":
		paymentErr.Code = errors.ErrCodeCardDeclined
		paymentErr.Message = "Card was declined"
		paymentErr.Retryable = false
		paymentErr.DeclineReason = stripeErr.DeclineCode

	case "expired_card":
		paymentErr.Code = errors.ErrCodeExpiredCard
		paymentErr.Message = "Card has expired"
		paymentErr.Retryable = false

	case "rate_limit":
		paymentErr.Code = errors.ErrCodeRateLimitExceeded
		paymentErr.Message = "Too many requests, please try again later"
		paymentErr.Retryable = true

	default:
		paymentErr.Code = errors.ErrCodeProviderError
		paymentErr.Message = "Payment provider error"
		paymentErr.Retryable = false
	}

	return paymentErr
}
```

---

## Idempotency & Retries

### Idempotency Implementation

```go
// In PaymentService.Authorize()

// Check idempotency before processing
existingPayment, err := s.paymentRepo.FindByIdempotencyKey(ctx, req.IdempotencyKey)
if err == nil && existingPayment != nil {
	// Return cached response
	log.Infof("Idempotent request detected, returning existing payment: %s", existingPayment.ID())
	return status.OK, s.buildPaymentResponse(existingPayment), nil
}

// Continue with payment processing...
```

### Retry Strategy

```go
package payment_provider

import (
	"context"
	"time"
	"math-ai.com/math-ai/internal/shared/errors"
)

// RetryConfig defines retry behavior
type RetryConfig struct {
	MaxRetries     int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
	Multiplier     float64
}

// WithRetry wraps a provider call with retry logic
func WithRetry(ctx context.Context, cfg *RetryConfig, fn func() error) error {
	backoff := cfg.InitialBackoff

	for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
		err := fn()
		if err == nil {
			return nil
		}

		// Check if error is retryable
		if paymentErr, ok := err.(*errors.PaymentError); ok {
			if !paymentErr.Retryable {
				return err // Don't retry non-retryable errors
			}
		}

		// Last attempt, return error
		if attempt == cfg.MaxRetries {
			return err
		}

		// Wait before retry
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff):
			backoff = time.Duration(float64(backoff) * cfg.Multiplier)
			if backoff > cfg.MaxBackoff {
				backoff = cfg.MaxBackoff
			}
		}
	}

	return nil
}

// Usage in provider:
func (p *StripeProvider) Authorize(ctx context.Context, req *diSvc.AuthorizeRequest) (*diSvc.PaymentResponse, error) {
	var resp *diSvc.PaymentResponse
	var callErr error

	retryConfig := &RetryConfig{
		MaxRetries:     3,
		InitialBackoff: 100 * time.Millisecond,
		MaxBackoff:     2 * time.Second,
		Multiplier:     2.0,
	}

	err := WithRetry(ctx, retryConfig, func() error {
		resp, callErr = p.doAuthorize(ctx, req)
		return callErr
	})

	if err != nil {
		return nil, err
	}

	return resp, nil
}
```

---

## Security Considerations

### 1. No Raw Card Storage
```go
// NEVER store raw card numbers
type PaymentMethod struct {
	// ❌ Bad
	cardNumber string

	// ✅ Good - only store tokenized data
	cardLast4    *string  // Last 4 digits for display
	providerID   string   // Provider's token (e.g., pm_xxx)
}
```

### 2. PCI Compliance
```
┌─────────────────────────────────────────────────────────┐
│                    Your Frontend                        │
│  ┌──────────────────────────────────────────────────┐  │
│  │   Card Input Form (hosted by Stripe/provider)   │  │
│  │   - User enters card details                     │  │
│  │   - Provider tokenizes card                      │  │
│  │   - Returns token (e.g., tok_xxx or pm_xxx)     │  │
│  └─────────────────────┬────────────────────────────┘  │
└────────────────────────┼───────────────────────────────┘
                         │ Token only
┌────────────────────────▼───────────────────────────────┐
│               Your Backend (math-srv)                   │
│  - Receives token only, never raw card data            │
│  - Stores token reference in database                  │
│  - Uses token for all payment operations               │
└─────────────────────────────────────────────────────────┘
```

### 3. Webhook Signature Verification
```go
func (p *StripeProvider) VerifyWebhookSignature(payload []byte, signature string) (bool, error) {
	// Verify webhook signature to prevent spoofing
	expectedSig := computeHMAC(payload, p.config.WebhookSecret)
	return signature == expectedSig, nil
}

// In webhook handler
func (s *paymentService) HandleWebhook(ctx context.Context, provider payment.Provider, payload []byte, signature string) error {
	// Get provider
	prov, err := s.providerFactory.GetProvider(provider)
	if err != nil {
		return err
	}

	// Verify signature
	valid, err := prov.VerifyWebhookSignature(payload, signature)
	if err != nil || !valid {
		return fmt.Errorf("invalid webhook signature")
	}

	// Process webhook...
	return nil
}
```

### 4. Configuration Security
```go
// config.go
type StripeConfig struct {
	PublicKey     string // Used in frontend
	SecretKey     string // NEVER expose to frontend
	WebhookSecret string // For webhook verification
}

// Load from environment
func NewEnv(envpath string) (*Env, error) {
	// ...
	result.StripeConfig = &StripeConfig{
		PublicKey:     getConfig("STRIPE_PUBLIC_KEY"),
		SecretKey:     getConfig("STRIPE_SECRET_KEY"),
		WebhookSecret: getConfig("STRIPE_WEBHOOK_SECRET"),
	}
	// ...
}
```

---

## Design Decisions & Tradeoffs

### 1. **Provider Interface vs. Abstract Base Class**
**Decision**: Use interface (`IPaymentProvider`)

**Pros**:
- Go idiomatic (interfaces are implicit)
- Easy to mock for testing
- No inheritance complexity
- Clear contract

**Cons**:
- Some code duplication across providers
- No shared implementation

**Mitigation**: Create helper functions for common operations (e.g., `buildFormBody()`, error handling)

### 2. **Synchronous vs. Asynchronous Processing**
**Decision**: Synchronous for critical path, async for webhooks

**Critical Path (Synchronous)**:
```go
// User waits for response
Authorize() -> Call Provider API -> Return result
```

**Pros**: Simple, immediate feedback
**Cons**: Slower API response times

**Webhook Path (Asynchronous)**:
```go
// Provider sends webhook later
Webhook -> Queue -> Background Job -> Update payment state
```

**Pros**: Handles delayed notifications (e.g., ACH transfers)
**Cons**: More complex state management

### 3. **State Machine Enforcement**
**Decision**: Enforce valid state transitions in domain layer

```go
// Domain prevents invalid transitions
payment.state = Captured
payment.MarkVoided() // ERROR: Can't void a captured payment
```

**Pros**:
- Prevents data corruption
- Business rules in one place
- Easy to audit

**Cons**:
- Requires careful state mapping from providers
- Provider-specific states may not fit cleanly

### 4. **Idempotency Key Strategy**
**Decision**: Client-provided idempotency keys

```go
type AuthorizeRequest struct {
	IdempotencyKey string `json:"idempotency_key" validate:"required"`
	// ...
}
```

**Pros**:
- Client controls retry logic
- Prevents duplicate charges
- Provider-agnostic

**Cons**:
- Requires client implementation
- Key storage overhead

**Alternative**: Server-generated keys (not recommended for payments)

### 5. **Partial Captures/Refunds**
**Decision**: Support partial amounts

```go
Authorize $100
Capture $75  // Capture partial
Void $25     // Void remaining authorization
```

**Pros**:
- Flexible for split shipments, tips, etc.
- Common in e-commerce

**Cons**:
- More complex state tracking
- Not all providers support partial operations

### 6. **Provider Response Storage**
**Decision**: Store raw provider responses in `provider_response` field

**Pros**:
- Audit trail
- Debugging support
- Can extract additional data later

**Cons**:
- Database storage overhead
- May contain sensitive data (requires careful scrubbing)

### 7. **Multi-Tenancy Support**
**Decision**: Provider per customer, not per tenant

```go
// Each customer can use a different provider
customer1 -> Stripe
customer2 -> CyberSource
```

**Pros**:
- Flexible provider selection
- A/B testing different providers
- Geographic optimization

**Cons**:
- More complex configuration
- Harder to consolidate reporting

### 8. **Error Normalization Depth**
**Decision**: Two-level error codes (normalized + provider-specific)

```go
type PaymentError struct {
	Code         string // "card_declined" (normalized)
	ProviderCode string // "card_declined" (Stripe) or "102" (CyberSource)
}
```

**Pros**:
- Consistent error handling in business logic
- Preserve provider details for debugging

**Cons**:
- Mapping complexity
- May lose nuance

### 9. **HTTP Client Usage**
**Decision**: Use your existing `internal/shared/http_client`

**Pros**:
- Consistent HTTP handling
- Built-in observability (traces, metrics)
- Retry logic
- Connection pooling

**Implementation**:
```go
func (p *StripeProvider) Authorize(ctx context.Context, req *AuthorizeRequest) (*PaymentResponse, error) {
	httpReq := &http_client.Request{
		Method: "POST",
		URL:    "https://api.stripe.com/v1/payment_intents",
		Headers: map[string]string{
			"Authorization": "Bearer " + p.config.SecretKey,
		},
		Body: buildBody(req),
	}

	resp, err := p.httpClient.Do(ctx, httpReq)
	// Handle response...
}
```

### 10. **Testing Strategy**
**Decision**: Mock provider for unit tests, real providers for integration tests

**Unit Tests**:
```go
mockProvider := &MockPaymentProvider{
	AuthorizeFunc: func(ctx context.Context, req *AuthorizeRequest) (*PaymentResponse, error) {
		return &PaymentResponse{
			ProviderPaymentID: "test_123",
			State: PaymentStateAuthorized,
		}, nil
	},
}

service := NewPaymentService(validator, repo, mockProvider)
```

**Integration Tests**:
```go
// Use Stripe test mode
stripeProvider := stripe.NewStripeProvider(&config.StripeConfig{
	SecretKey: "sk_test_...",
})
```

---

## Database Schema

```sql
-- Customers table
CREATE TABLE payment_customers (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    provider VARCHAR(50) NOT NULL,
    provider_customer_id VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    phone VARCHAR(50),
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,

    UNIQUE(provider, provider_customer_id),
    INDEX idx_user_id (user_id),
    INDEX idx_provider_customer_id (provider, provider_customer_id)
);

-- Payment methods table
CREATE TABLE payment_methods (
    id VARCHAR(36) PRIMARY KEY,
    customer_id VARCHAR(36) NOT NULL,
    provider VARCHAR(50) NOT NULL,
    provider_method_id VARCHAR(255) NOT NULL,
    method_type VARCHAR(50) NOT NULL,
    is_default BOOLEAN NOT NULL DEFAULT FALSE,

    -- Card details (tokenized)
    card_last4 VARCHAR(4),
    card_brand VARCHAR(50),
    card_exp_month INT,
    card_exp_year INT,

    -- Bank account details (tokenized)
    bank_last4 VARCHAR(4),
    bank_name VARCHAR(255),
    account_type VARCHAR(50),

    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,

    FOREIGN KEY (customer_id) REFERENCES payment_customers(id),
    UNIQUE(provider, provider_method_id),
    INDEX idx_customer_id (customer_id)
);

-- Payments table
CREATE TABLE payments (
    id VARCHAR(36) PRIMARY KEY,
    customer_id VARCHAR(36) NOT NULL,
    payment_method_id VARCHAR(36) NOT NULL,
    parent_payment_id VARCHAR(36),
    provider VARCHAR(50) NOT NULL,
    provider_payment_id VARCHAR(255),

    transaction_type VARCHAR(50) NOT NULL,
    state VARCHAR(50) NOT NULL,

    amount BIGINT NOT NULL,
    currency VARCHAR(3) NOT NULL,
    authorized_amount BIGINT NOT NULL DEFAULT 0,
    captured_amount BIGINT NOT NULL DEFAULT 0,
    refunded_amount BIGINT NOT NULL DEFAULT 0,

    idempotency_key VARCHAR(255) NOT NULL,

    description TEXT,
    error_code VARCHAR(100),
    error_message TEXT,

    provider_response JSONB,
    metadata JSONB,

    authorized_at TIMESTAMP,
    captured_at TIMESTAMP,
    failed_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    FOREIGN KEY (customer_id) REFERENCES payment_customers(id),
    FOREIGN KEY (payment_method_id) REFERENCES payment_methods(id),
    FOREIGN KEY (parent_payment_id) REFERENCES payments(id),

    UNIQUE(idempotency_key),
    INDEX idx_customer_id (customer_id),
    INDEX idx_provider_payment_id (provider, provider_payment_id),
    INDEX idx_state (state),
    INDEX idx_created_at (created_at DESC)
);
```

---

## API Endpoints

```go
// HTTP Routes (internal/app/routes/app_routes.go)

func RegisterPaymentRoutes(router *gex.Router, paymentController *controller.PaymentController) {
	// Customer management
	router.POST("/payment/customers", paymentController.CreateCustomer)
	router.GET("/payment/customers/:id", paymentController.GetCustomer)

	// Payment method management
	router.POST("/payment/methods/credit-card", paymentController.AddCreditCard)
	router.POST("/payment/methods/debit-card", paymentController.AddDebitCard)
	router.POST("/payment/methods/bank-account", paymentController.AddBankAccount)
	router.GET("/payment/methods", paymentController.GetPaymentMethods)
	router.DELETE("/payment/methods/:id", paymentController.DeletePaymentMethod)
	router.PUT("/payment/methods/:id/default", paymentController.SetDefaultPaymentMethod)

	// Payment operations
	router.POST("/payment/authorize", paymentController.Authorize)
	router.POST("/payment/capture", paymentController.Capture)
	router.POST("/payment/refund", paymentController.Refund)
	router.POST("/payment/void", paymentController.Void)

	// Query
	router.GET("/payment/:id", paymentController.GetPayment)
	router.GET("/payment/history", paymentController.GetPaymentHistory)

	// Webhooks
	router.POST("/payment/webhook/stripe", paymentController.HandleStripeWebhook)
	router.POST("/payment/webhook/cybersource", paymentController.HandleCyberSourceWebhook)
}
```

---

## Complete Integration Example

```go
// Example: Complete payment flow

// 1. Create customer
createCustomerReq := &dto.CreatePaymentCustomerRequest{
	UserID:   "user_123",
	Provider: "stripe",
	Email:    "customer@example.com",
	Name:     "John Doe",
}
statusCode, customer, err := paymentService.CreateCustomer(ctx, createCustomerReq)

// 2. Add payment method
addCardReq := &dto.AddCreditCardRequest{
	CustomerID:   customer.ID,
	Token:        "tok_visa", // From Stripe.js tokenization
	SetAsDefault: true,
}
statusCode, paymentMethod, err := paymentService.AddCreditCard(ctx, addCardReq)

// 3. Authorize payment
authReq := &dto.AuthorizePaymentRequest{
	CustomerID:     customer.ID,
	Amount:         10000, // $100.00
	Currency:       "usd",
	Description:    "Order #12345",
	IdempotencyKey: "order_12345_auth_" + uuid.New().String(),
}
statusCode, payment, err := paymentService.Authorize(ctx, authReq)

// 4. Capture payment (e.g., after shipping)
captureReq := &dto.CapturePaymentRequest{
	PaymentID:      payment.ID,
	Amount:         nil, // Capture full amount
	IdempotencyKey: "order_12345_capture_" + uuid.New().String(),
}
statusCode, capturedPayment, err := paymentService.Capture(ctx, captureReq)

// 5. Refund (if needed)
refundReq := &dto.RefundPaymentRequest{
	PaymentID:      capturedPayment.ID,
	Amount:         &refundAmount,
	IdempotencyKey: "order_12345_refund_" + uuid.New().String(),
}
statusCode, refundedPayment, err := paymentService.Refund(ctx, refundReq)
```

---

## Summary

This architecture provides:

✅ **Provider Independence**: Business logic never depends on specific providers
✅ **Unified Interface**: Single API for all payment operations
✅ **Normalized Responses**: All provider responses mapped to domain models
✅ **Extensibility**: Add new providers by implementing `IPaymentProvider`
✅ **State Machine**: Enforces valid payment state transitions
✅ **Idempotency**: Prevents duplicate transactions
✅ **Retry Logic**: Handles transient failures gracefully
✅ **Error Normalization**: Consistent error handling across providers
✅ **Security**: PCI-compliant, no raw card storage
✅ **Testability**: Easy to mock and test
✅ **Observability**: Uses existing telemetry infrastructure

**Next Steps**:
1. Review and approve architecture
2. Implement database migrations
3. Implement Stripe provider (reference implementation)
4. Add unit and integration tests
5. Implement remaining providers (CyberSource, Plaid, etc.)
6. Add webhook handlers
7. Create API documentation

