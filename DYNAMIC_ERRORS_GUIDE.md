# Dynamic Errors with Localization - Complete Guide

## Overview

This system handles error messages with **dynamic arguments** (runtime values like email addresses, user IDs, etc.) while maintaining **full localization** support across all languages (EN, VN, FR).

## The Problem & Solution

### ❌ Before (Static Messages Only)

```go
// Static message - user doesn't know WHICH email exists
return status.USER_EMAIL_ALREADY_EXISTS, nil, err_svc.ErrEmailAlreadyExists

// Response:
// EN: "Email already exists."
// VN: "Email đã tồn tại."
```

### ✅ After (Dynamic Messages!)

```go
// Dynamic message with actual email value
return status.USER_EMAIL_ALREADY_EXISTS, nil, err_svc.NewEmailAlreadyExistsError("john@example.com")

// Response (automatically localized):
// EN: "Email john@example.com already exists."
// VN: "Email john@example.com đã tồn tại."
// FR: "L'email john@example.com existe déjà."
```

## Quick Start

### 1. Use Helper Functions (Easiest)

```go
import err_svc "math-ai.com/math-ai/internal/shared/error"

// In your service
func (s *UserService) CreateUser(ctx context.Context, req *dto.CreateUserRequest) (status.Code, *dto.UserResponse, error) {
    // Check if email exists
    existingUser, _ := s.repo.FindByEmail(ctx, req.Email)
    if existingUser != nil {
        // ✅ Return dynamic error with email value
        return status.USER_EMAIL_ALREADY_EXISTS, nil, err_svc.NewEmailAlreadyExistsError(req.Email)
    }

    // Check if phone exists
    existingUser, _ = s.repo.FindByPhone(ctx, req.Phone)
    if existingUser != nil {
        // ✅ Return dynamic error with phone value
        return status.USER_PHONE_ALREADY_EXISTS, nil, err_svc.NewPhoneAlreadyExistsError(req.Phone)
    }

    // Business logic...
    return status.SUCCESS, userResponse, nil
}
```

### 2. Client Receives Localized Message

**Request with locale = "en":**
```json
{
  "status": 10010,
  "error": "email already exists",
  "message": "Email john@example.com already exists."
}
```

**Request with locale = "vn":**
```json
{
  "status": 10010,
  "error": "email already exists",
  "message": "Email john@example.com đã tồn tại."
}
```

## Available Helper Functions

### User Errors

```go
// Email already exists
err_svc.NewEmailAlreadyExistsError(email string)
// EN: "Email {email} already exists."
// VN: "Email {email} đã tồn tại."

// Phone already exists
err_svc.NewPhoneAlreadyExistsError(phone string)
// EN: "Phone number {phone} already exists."
// VN: "Số điện thoại {phone} đã tồn tại."

// User not found
err_svc.NewUserNotFoundError(userID string)
// EN: "User with ID {user_id} not found."
// VN: "Không tìm thấy người dùng có ID {user_id}."
```

### Grade Errors

```go
// Grade already exists
err_svc.NewGradeAlreadyExistsError(label string)
// EN: "Grade '{label}' already exists."
// VN: "Cấp học '{label}' đã tồn tại."

// Grade not found
err_svc.NewGradeNotFoundError(label string)
// EN: "Grade '{label}' not found."
// VN: "Không tìm thấy cấp học '{label}'."
```

## Advanced Usage

### Custom Dynamic Errors

For cases not covered by helper functions:

```go
import err_svc "math-ai.com/math-ai/internal/shared/error"

// Create custom dynamic error
err := err_svc.NewDynamicError(
    status.YOUR_STATUS_CODE,
    map[string]interface{}{
        "key1": value1,
        "key2": value2,
    },
)

return statusCode, nil, err
```

### Multiple Arguments

```go
// Example: Name too long with max length
err := err_svc.NewDynamicError(
    status.CONTACT_NAME_TOO_LONG,
    map[string]interface{}{
        "max_length": 100,
    },
)
```

## Real-World Examples

### Example 1: User Registration

```go
func (s *UserService) CreateUser(ctx context.Context, req *dto.CreateUserRequest) (status.Code, *dto.UserResponse, error) {
    // Validate email format (static message)
    if !isValidEmail(req.Email) {
        return status.USER_INVALID_EMAIL, nil, err_svc.ErrInvalidEmail
    }

    // Check duplicate email (dynamic message with email value)
    existingUser, _ := s.repo.FindByEmail(ctx, req.Email)
    if existingUser != nil {
        return status.USER_EMAIL_ALREADY_EXISTS, nil, err_svc.NewEmailAlreadyExistsError(req.Email)
    }

    // Check duplicate phone (dynamic message with phone value)
    existingUser, _ = s.repo.FindByPhone(ctx, req.Phone)
    if existingUser != nil {
        return status.USER_PHONE_ALREADY_EXISTS, nil, err_svc.NewPhoneAlreadyExistsError(req.Phone)
    }

    // Create user...
    return status.SUCCESS, userResponse, nil
}
```

### Example 2: Grade Management

```go
func (s *GradeService) CreateGrade(ctx context.Context, req *dto.CreateGradeRequest) (status.Code, *dto.GradeResponse, error) {
    // Check if grade exists (dynamic message with grade label)
    existingGrade, _ := s.repo.FindByLabel(ctx, req.Label)
    if existingGrade != nil {
        return status.GRADE_ALREADY_EXISTS, nil, err_svc.NewGradeAlreadyExistsError(req.Label)
    }

    // Create grade...
    return status.SUCCESS, gradeResponse, nil
}

func (s *GradeService) GetGradeByLabel(ctx context.Context, label string) (status.Code, *dto.GradeResponse, error) {
    grade, err := s.repo.FindByLabel(ctx, label)
    if err != nil || grade == nil {
        return status.GRADE_NOT_FOUND, nil, err_svc.NewGradeNotFoundError(label)
    }

    return status.SUCCESS, gradeResponse, nil
}
```

## Adding New Dynamic Errors

### Step 1: Add Message Templates

Add templates for each language:

**`internal/shared/utils/locales/templates_en.go`:**
```go
case status.YOUR_NEW_ERROR:
    return "Your message with {placeholder}"
```

**`internal/shared/utils/locales/templates_vn.go`:**
```go
case status.YOUR_NEW_ERROR:
    return "Thông báo của bạn với {placeholder}"
```

**`internal/shared/utils/locales/templates_fr.go`:**
```go
case status.YOUR_NEW_ERROR:
    return "Votre message avec {placeholder}"
```

### Step 2: Create Helper Function (Optional)

**`internal/shared/error/your_domain.go`:**
```go
func NewYourCustomError(value string) error {
    return &DynamicError{
        StatusCode: status.YOUR_NEW_ERROR,
        Args:       map[string]interface{}{"placeholder": value},
        BaseError:  errors.New("base error message"),
    }
}
```

### Step 3: Use It

```go
return status.YOUR_NEW_ERROR, nil, err_svc.NewYourCustomError(dynamicValue)
```

## How It Works Internally

1. **Service returns DynamicError** with status code and args
2. **Response handler** detects it's a DynamicError (via `IsDynamicError()`)
3. **Locale is extracted** from context (from metadata)
4. **Template is loaded** for the user's language
5. **Arguments are interpolated** into the template using `{key}` placeholders
6. **Localized message is sent** to the client

## Template Placeholders

| Placeholder | Description | Example |
|-------------|-------------|---------|
| `{email}` | Email address | `john@example.com` |
| `{phone}` | Phone number | `+84123456789` |
| `{user_id}` | User ID | `abc-123` |
| `{label}` | Grade/Level label | `Grade 5` |
| `{name}` | Semester/Entity name | `Semester 1` |
| `{max_length}` | Maximum length | `100` |

## Benefits

✅ **Better UX** - Users see exactly what value caused the error
✅ **Full Localization** - Messages in user's language (EN/VN/FR)
✅ **Type-Safe** - Compile-time safety with Go types
✅ **Maintainable** - Templates defined once, used everywhere
✅ **Backward Compatible** - Old static errors still work
✅ **Clean Code** - Helper functions make it simple

## Migration from Static Errors

**Before:**
```go
if existingUser != nil {
    return status.USER_EMAIL_ALREADY_EXISTS, nil, err_svc.ErrEmailAlreadyExists
}
```

**After:**
```go
if existingUser != nil {
    return status.USER_EMAIL_ALREADY_EXISTS, nil, err_svc.NewEmailAlreadyExistsError(req.Email)
}
```

## Summary

Use `DynamicError` when you want to show the actual value that caused the problem in error messages. The system handles localization automatically based on the user's locale from the request metadata.

**Simple pattern:**
1. Use helper functions: `err_svc.NewEmailAlreadyExistsError(email)`
2. Pass runtime values as arguments
3. System handles localization automatically
4. Client receives message in their language with actual values

Perfect for errors like:
- "Email john@example.com already exists"
- "User with ID abc-123 not found"
- "Grade 'Grade 5' already exists"
- Any error where showing the actual value helps the user
