# Service Layer Tracing Examples

This document shows how to use the trace helpers for custom service layer instrumentation.

## Table of Contents
- [Basic Usage](#basic-usage)
- [Adding Attributes](#adding-attributes)
- [Error Handling](#error-handling)
- [Recording Events](#recording-events)
- [Async Operations](#async-operations)
- [Common Patterns](#common-patterns)

## Basic Usage

### Simple Service Method with Tracing

```go
package services

import (
    "context"

    "math-ai.com/math-ai/internal/applications/dto"
    "math-ai.com/math-ai/internal/shared/telemetry/trace"
)

func (s *UserService) ProcessUserAction(ctx context.Context, userID string, action string) error {
    // Start a service span
    ctx, span := trace.StartServiceSpan(ctx, "UserService.ProcessUserAction",
        trace.StringAttr("user_id", userID),
        trace.StringAttr("action", action),
    )
    defer span.End()

    // Your business logic here
    result, err := s.performAction(ctx, userID, action)
    if err != nil {
        trace.RecordServiceError(span, err)
        return err
    }

    // Add result information to the span
    trace.AddServiceAttribute(span, "result", result)
    trace.MarkServiceSuccess(span)

    return nil
}
```

## Adding Attributes

### Using Pre-built Attribute Builders

```go
func (s *ProfileService) UpdateUserProfile(ctx context.Context, req *dto.UpdateProfileRequest) error {
    // Use pre-built attribute builders
    attrs := trace.ProfileAttributes(
        req.ProfileID,
        req.GradeLevel,
        req.Semester,
    )

    ctx, span := trace.StartServiceSpan(ctx, "ProfileService.UpdateUserProfile", attrs...)
    defer span.End()

    // Business logic...

    trace.MarkServiceSuccess(span)
    return nil
}
```

### Adding Attributes During Execution

```go
func (s *QuizService) GenerateQuiz(ctx context.Context, req *dto.GenerateQuizRequest) (*dto.Quiz, error) {
    ctx, span := trace.StartServiceSpan(ctx, "QuizService.GenerateQuiz")
    defer span.End()

    // Add initial attributes
    trace.AddServiceAttributes(span,
        trace.StringAttr("user_id", req.UID),
        trace.StringAttr("grade", req.Grade),
    )

    // Generate quiz
    quiz, err := s.generateQuizContent(ctx, req)
    if err != nil {
        trace.RecordServiceError(span, err)
        return nil, err
    }

    // Add attributes after generation
    trace.AddServiceAttributes(span,
        trace.IntAttr("question_count", len(quiz.Questions)),
        trace.StringAttr("quiz_id", quiz.ID),
    )

    trace.MarkServiceSuccess(span)
    return quiz, nil
}
```

## Error Handling

### Recording Errors with Custom Messages

```go
func (s *PaymentService) ProcessPayment(ctx context.Context, userID string, amount float64) error {
    ctx, span := trace.StartServiceSpan(ctx, "PaymentService.ProcessPayment",
        trace.StringAttr("user_id", userID),
        trace.Float64Attr("amount", amount),
    )
    defer span.End()

    // Validate amount
    if amount <= 0 {
        err := fmt.Errorf("invalid amount: %.2f", amount)
        trace.RecordServiceErrorWithMessage(span, err, "Payment validation failed")
        return err
    }

    // Process payment
    err := s.chargePayment(ctx, userID, amount)
    if err != nil {
        // Record error with context-specific message
        trace.RecordServiceErrorWithMessage(span, err,
            fmt.Sprintf("Failed to charge payment for user %s", userID))
        return err
    }

    trace.MarkServiceSuccess(span)
    return nil
}
```

### Handling Partial Failures

```go
func (s *BatchService) ProcessBatch(ctx context.Context, items []string) error {
    ctx, span := trace.StartServiceSpan(ctx, "BatchService.ProcessBatch",
        trace.IntAttr("total_items", len(items)),
    )
    defer span.End()

    successCount := 0
    failureCount := 0

    for _, item := range items {
        err := s.processItem(ctx, item)
        if err != nil {
            failureCount++
            // Record individual failures as events
            trace.RecordServiceEvent(span, "item.failed",
                trace.StringAttr("item_id", item),
                trace.StringAttr("error", err.Error()),
            )
        } else {
            successCount++
        }
    }

    // Add final statistics
    trace.AddServiceAttributes(span,
        trace.IntAttr("success_count", successCount),
        trace.IntAttr("failure_count", failureCount),
    )

    if failureCount > 0 {
        trace.RecordServiceErrorWithMessage(span,
            fmt.Errorf("%d items failed", failureCount),
            "Batch processing completed with errors")
    } else {
        trace.MarkServiceSuccess(span)
    }

    return nil
}
```

## Recording Events

### Marking Important Milestones

```go
func (s *QuizService) GenerateComplexQuiz(ctx context.Context, req *dto.GenerateQuizRequest) error {
    ctx, span := trace.StartServiceSpan(ctx, "QuizService.GenerateComplexQuiz")
    defer span.End()

    // Step 1: Validation
    if err := s.validateRequest(ctx, req); err != nil {
        trace.RecordServiceError(span, err)
        return err
    }
    trace.RecordServiceEvent(span, "validation.completed")

    // Step 2: Fetch user profile
    profile, err := s.fetchProfile(ctx, req.UID)
    if err != nil {
        trace.RecordServiceError(span, err)
        return err
    }
    trace.RecordServiceEvent(span, "profile.fetched",
        trace.StringAttr("grade_level", profile.Grade))

    // Step 3: Generate questions
    trace.RecordServiceEvent(span, "question_generation.started")
    questions, err := s.generateQuestions(ctx, req, profile)
    if err != nil {
        trace.RecordServiceError(span, err)
        return err
    }
    trace.RecordServiceEvent(span, "question_generation.completed",
        trace.IntAttr("question_count", len(questions)))

    // Step 4: AI Review
    trace.RecordServiceEvent(span, "ai_review.started")
    reviewed, err := s.reviewWithAI(ctx, questions)
    if err != nil {
        trace.RecordServiceError(span, err)
        return err
    }
    trace.RecordServiceEvent(span, "ai_review.completed")

    // Step 5: Save to database
    trace.RecordServiceEvent(span, "database.save.started")
    if err := s.saveQuiz(ctx, reviewed); err != nil {
        trace.RecordServiceError(span, err)
        return err
    }
    trace.RecordServiceEvent(span, "database.save.completed")

    trace.MarkServiceSuccess(span)
    return nil
}
```

## Async Operations

### Tracing Background Jobs

```go
func (s *NotificationService) SendBulkNotifications(ctx context.Context, userIDs []string, message string) error {
    ctx, span := trace.StartServiceSpan(ctx, "NotificationService.SendBulkNotifications",
        trace.IntAttr("user_count", len(userIDs)),
    )
    defer span.End()

    // Launch async notification sending
    for _, userID := range userIDs {
        go func(uid string) {
            // Create a new async span for each notification
            asyncCtx, asyncSpan := trace.StartAsyncOperation(ctx, "NotificationService.SendNotification",
                trace.StringAttr("user_id", uid),
            )
            defer asyncSpan.End()

            err := s.sendNotification(asyncCtx, uid, message)
            if err != nil {
                trace.RecordServiceError(asyncSpan, err)
            } else {
                trace.MarkServiceSuccess(asyncSpan)
            }
        }(userID)
    }

    trace.RecordServiceEvent(span, "async_notifications.dispatched")
    trace.MarkServiceSuccess(span)

    return nil
}
```

## Common Patterns

### Pattern 1: Validation Heavy Operations

```go
func (s *UserService) CreateComplexUser(ctx context.Context, req *dto.CreateUserRequest) error {
    ctx, span := trace.StartServiceSpan(ctx, "UserService.CreateComplexUser")
    defer span.End()

    // Validation
    errors := make([]string, 0)

    if err := s.validateEmail(ctx, req.Email); err != nil {
        errors = append(errors, err.Error())
    }

    if err := s.validatePhone(ctx, req.Phone); err != nil {
        errors = append(errors, err.Error())
    }

    if err := s.validateAge(ctx, req.Age); err != nil {
        errors = append(errors, err.Error())
    }

    // Add validation result to span
    validationPassed := len(errors) == 0
    trace.AddServiceAttributes(span,
        trace.ValidationAttributes(validationPassed, len(errors))...)

    if !validationPassed {
        err := fmt.Errorf("validation failed: %v", errors)
        trace.RecordServiceError(span, err)
        return err
    }

    // Continue with creation...
    trace.MarkServiceSuccess(span)
    return nil
}
```

### Pattern 2: Multi-Step Operations

```go
func (s *QuizService) PublishQuiz(ctx context.Context, quizID string) error {
    // Use WithServiceSpan for cleaner code
    return trace.WithServiceSpan(ctx, "QuizService.PublishQuiz",
        func(ctx context.Context, span trace.Span) error {
            trace.AddServiceAttribute(span, "quiz_id", quizID)

            // Step 1: Validate quiz
            trace.RecordServiceEvent(span, "validation.started")
            if err := s.validateQuiz(ctx, quizID); err != nil {
                return err
            }

            // Step 2: Update status
            trace.RecordServiceEvent(span, "status.update.started")
            if err := s.updateQuizStatus(ctx, quizID, "published"); err != nil {
                return err
            }

            // Step 3: Notify users
            trace.RecordServiceEvent(span, "notification.started")
            if err := s.notifyUsers(ctx, quizID); err != nil {
                return err
            }

            return nil
        },
        trace.StringAttr("quiz_id", quizID),
    )
}
```

### Pattern 3: Performance-Critical Sections

```go
func (s *AnalyticsService) ProcessLargeDataset(ctx context.Context, datasetID string) error {
    ctx, span := trace.StartServiceSpan(ctx, "AnalyticsService.ProcessLargeDataset")
    defer span.End()

    // Only do expensive operations if span is recording
    if trace.IsSpanRecording(ctx) {
        // Add expensive debug information
        dataInfo, _ := s.getDatasetInfo(ctx, datasetID)
        trace.AddServiceAttribute(span, "dataset_size", dataInfo.Size)
        trace.AddServiceAttribute(span, "dataset_type", dataInfo.Type)
    }

    // Main processing logic
    result, err := s.processData(ctx, datasetID)
    if err != nil {
        trace.RecordServiceError(span, err)
        return err
    }

    trace.AddServiceAttribute(span, "records_processed", result.Count)
    trace.MarkServiceSuccess(span)

    return nil
}
```

## Best Practices

### 1. **Span Naming Convention**
Use the format `ServiceName.MethodName`:
```go
trace.StartServiceSpan(ctx, "UserService.CreateUser")
trace.StartServiceSpan(ctx, "QuizService.GenerateQuiz")
trace.StartServiceSpan(ctx, "ProfileService.UpdateProfile")
```

### 2. **Always Use defer for span.End()**
```go
ctx, span := trace.StartServiceSpan(ctx, "Service.Method")
defer span.End()  // Always defer immediately
```

### 3. **Record Both Success and Failure**
```go
if err != nil {
    trace.RecordServiceError(span, err)
    return err
}
trace.MarkServiceSuccess(span)  // Don't forget success!
```

### 4. **Use Semantic Attribute Names**
Use the predefined attribute constants:
```go
// Good
trace.StringAttr(trace.AttrUserUID, userID)
trace.IntAttr(trace.AttrQuestionCount, count)

// Avoid
trace.StringAttr("uid", userID)  // Not semantic
trace.IntAttr("count", count)    // Too generic
```

### 5. **Don't Over-Instrument**
Only add spans for:
- Important business operations
- Operations that might fail
- Operations with significant performance impact
- Operations you want to debug

Don't add spans for:
- Trivial getters/setters
- Simple utility functions
- Every single database query (already instrumented)

### 6. **Use Events for Milestones, Attributes for Data**
```go
// Use events for "things that happened"
trace.RecordServiceEvent(span, "validation.completed")
trace.RecordServiceEvent(span, "email.sent")

// Use attributes for "data about the operation"
trace.AddServiceAttribute(span, "user_count", 100)
trace.AddServiceAttribute(span, "quiz_type", "practice")
```

## When to Use Manual Tracing

Use manual service tracing when:
- ✅ You have complex multi-step business logic
- ✅ You need to debug performance bottlenecks
- ✅ You want to track specific business operations
- ✅ You need detailed visibility into error paths

Don't use manual tracing when:
- ❌ HTTP requests (already traced by middleware)
- ❌ Database queries (already traced by wrapper)
- ❌ AI provider calls (already traced by wrapper)
- ❌ Simple CRUD operations without complex logic
