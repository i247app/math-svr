# Service Layer Tracing Helpers

This package provides easy-to-use helper functions for adding custom OpenTelemetry tracing to your service layer code.

## 📚 Quick Links

- [Examples](./EXAMPLES.md) - Comprehensive usage examples
- [Main Documentation](../../../OBSERVABILITY.md) - Full observability guide

## 🚀 Quick Start

### Basic Service Method Tracing

```go
import (
    "context"
    "math-ai.com/math-ai/internal/shared/telemetry/trace"
)

func (s *MyService) DoSomething(ctx context.Context, userID string) error {
    // Create a span
    ctx, span := trace.StartServiceSpan(ctx, "MyService.DoSomething",
        trace.StringAttr("user_id", userID),
    )
    defer span.End()

    // Your business logic
    result, err := s.performOperation(ctx, userID)
    if err != nil {
        trace.RecordServiceError(span, err)
        return err
    }

    // Mark success
    trace.MarkServiceSuccess(span)
    return nil
}
```

## 📦 Available Functions

### Span Management

| Function | Description |
|----------|-------------|
| `StartServiceSpan(ctx, name, attrs...)` | Start a new service span |
| `StartAsyncOperation(ctx, name, attrs...)` | Start span for async operation |
| `WithServiceSpan(ctx, name, fn, attrs...)` | Wrap function with automatic span |
| `GetSpanFromContext(ctx)` | Get current span |
| `IsSpanRecording(ctx)` | Check if span is recording |

### Span Attributes

| Function | Description |
|----------|-------------|
| `AddServiceAttribute(span, key, value)` | Add single attribute |
| `AddServiceAttributes(span, attrs...)` | Add multiple attributes |
| `StringAttr(key, value)` | Create string attribute |
| `IntAttr(key, value)` | Create int attribute |
| `Float64Attr(key, value)` | Create float64 attribute |
| `BoolAttr(key, value)` | Create bool attribute |

### Pre-built Attribute Builders

| Function | Description |
|----------|-------------|
| `UserAttributes(uid, email, platform)` | User-related attributes |
| `ProfileAttributes(id, grade, semester)` | Profile attributes |
| `QuizAttributes(id, type, difficulty, topic, count)` | Quiz attributes |
| `AssessmentAttributes(id, total, correct, percentage)` | Assessment attributes |
| `OperationAttributes(type, name)` | Operation attributes |
| `ResourceAttributes(id, type)` | Resource attributes |
| `ValidationAttributes(passed, errorCount)` | Validation attributes |
| `BatchProcessingAttributes(size, processed, failed)` | Batch attributes |

### Status Management

| Function | Description |
|----------|-------------|
| `MarkServiceSuccess(span)` | Mark span as successful |
| `RecordServiceError(span, err)` | Record error in span |
| `RecordServiceErrorWithMessage(span, err, msg)` | Record error with custom message |
| `RecordServiceEvent(span, name, attrs...)` | Record discrete event |

## 🎯 Common Use Cases

### 1. Simple Service Method

```go
func (s *UserService) GetUserProfile(ctx context.Context, userID string) (*Profile, error) {
    ctx, span := trace.StartServiceSpan(ctx, "UserService.GetUserProfile",
        trace.StringAttr(trace.AttrUserID, userID),
    )
    defer span.End()

    profile, err := s.repo.FindByID(ctx, userID)
    if err != nil {
        trace.RecordServiceError(span, err)
        return nil, err
    }

    trace.MarkServiceSuccess(span)
    return profile, nil
}
```

### 2. Multi-Step Operation

```go
func (s *QuizService) PublishQuiz(ctx context.Context, quizID string) error {
    ctx, span := trace.StartServiceSpan(ctx, "QuizService.PublishQuiz")
    defer span.End()

    // Step 1
    trace.RecordServiceEvent(span, "validation.started")
    if err := s.validateQuiz(ctx, quizID); err != nil {
        trace.RecordServiceError(span, err)
        return err
    }
    trace.RecordServiceEvent(span, "validation.completed")

    // Step 2
    trace.RecordServiceEvent(span, "publish.started")
    if err := s.publishToDatabase(ctx, quizID); err != nil {
        trace.RecordServiceError(span, err)
        return err
    }
    trace.RecordServiceEvent(span, "publish.completed")

    trace.MarkServiceSuccess(span)
    return nil
}
```

### 3. With Pre-built Attributes

```go
func (s *AssessmentService) SubmitAssessment(ctx context.Context, req *SubmitRequest) error {
    attrs := trace.AssessmentAttributes(
        req.AssessmentID,
        req.TotalQuestions,
        req.CorrectAnswers,
        req.ScorePercentage,
    )

    ctx, span := trace.StartServiceSpan(ctx, "AssessmentService.SubmitAssessment", attrs...)
    defer span.End()

    // Process assessment...

    trace.MarkServiceSuccess(span)
    return nil
}
```

### 4. Async Operations

```go
func (s *EmailService) SendBulkEmails(ctx context.Context, emails []string) error {
    ctx, span := trace.StartServiceSpan(ctx, "EmailService.SendBulkEmails",
        trace.IntAttr("email_count", len(emails)),
    )
    defer span.End()

    for _, email := range emails {
        go func(addr string) {
            asyncCtx, asyncSpan := trace.StartAsyncOperation(ctx, "EmailService.SendEmail",
                trace.StringAttr("email", addr),
            )
            defer asyncSpan.End()

            err := s.sendEmail(asyncCtx, addr)
            if err != nil {
                trace.RecordServiceError(asyncSpan, err)
            } else {
                trace.MarkServiceSuccess(asyncSpan)
            }
        }(email)
    }

    trace.MarkServiceSuccess(span)
    return nil
}
```

### 5. Using WithServiceSpan Helper

```go
func (s *UserService) ActivateUser(ctx context.Context, userID string) error {
    return trace.WithServiceSpan(ctx, "UserService.ActivateUser",
        func(ctx context.Context, span trace.Span) error {
            // Automatically handles span creation and cleanup
            // Automatically records errors
            // Automatically marks success

            trace.AddServiceAttribute(span, "user_id", userID)

            if err := s.validateUser(ctx, userID); err != nil {
                return err  // Error automatically recorded
            }

            if err := s.updateStatus(ctx, userID, "active"); err != nil {
                return err
            }

            return nil  // Success automatically marked
        },
        trace.StringAttr("user_id", userID),
    )
}
```

## 🏷️ Semantic Attribute Keys

Use predefined attribute constants for consistency:

```go
// User attributes
trace.AttrUserID
trace.AttrUserUID
trace.AttrUserEmail
trace.AttrUserPlatform

// Profile attributes
trace.AttrProfileID
trace.AttrGradeLevel
trace.AttrSemester

// Quiz attributes
trace.AttrQuizID
trace.AttrQuizType
trace.AttrQuizDifficulty
trace.AttrQuizTopic
trace.AttrQuestionCount

// Assessment attributes
trace.AttrAssessmentID
trace.AttrScore
trace.AttrTotalQuestions
trace.AttrCorrectAnswers
trace.AttrScorePercentage

// Operation attributes
trace.AttrOperationType
trace.AttrOperationName
trace.AttrResourceID
trace.AttrResourceType

// Validation attributes
trace.AttrValidationPassed
trace.AttrValidationErrors

// Batch processing attributes
trace.AttrItemCount
trace.AttrProcessedCount
trace.AttrFailedCount
trace.AttrBatchSize
```

## 📋 Best Practices

### ✅ DO

- **Use descriptive span names**: `"UserService.CreateUser"` not `"createUser"`
- **Always defer span.End()**: Immediately after creating the span
- **Record both success and errors**: Use `MarkServiceSuccess()` and `RecordServiceError()`
- **Use semantic attribute keys**: `trace.AttrUserID` not `"user_id"`
- **Add events for milestones**: `RecordServiceEvent(span, "validation.completed")`
- **Use pre-built attribute builders**: `UserAttributes()`, `QuizAttributes()`, etc.

### ❌ DON'T

- **Don't forget to end spans**: Always use `defer span.End()`
- **Don't over-instrument**: Only trace important business operations
- **Don't trace already-instrumented code**: HTTP/DB/AI calls are auto-traced
- **Don't use generic attribute names**: Be specific and semantic
- **Don't add spans to getters/setters**: Only meaningful operations
- **Don't create spans when not recording**: Check with `IsSpanRecording()` for expensive ops

## 🔍 When to Use Manual Tracing

### Use When:
- ✅ Complex multi-step business logic
- ✅ Operations that might fail in interesting ways
- ✅ Performance-critical code you want to monitor
- ✅ Business operations you want to track in Jaeger

### Don't Use When:
- ❌ HTTP requests (auto-traced by middleware)
- ❌ Database queries (auto-traced by wrapper)
- ❌ AI provider calls (auto-traced by wrapper)
- ❌ Simple getters/setters
- ❌ Trivial utility functions

## 📊 Viewing Traces

After adding instrumentation:

1. **Start observability stack**: `cd docker/observability && ./start.sh`
2. **Run your application**: `make run`
3. **Open Jaeger**: http://localhost:16686
4. **Select service**: "math-srv"
5. **Find your traces**: Look for operations like "UserService.CreateUser"

You'll see:
```
HTTP POST /users (500ms)
├─ UserService.CreateUser (450ms)
│  ├─ validation.started (event)
│  ├─ db.query: FindByEmail (50ms)
│  ├─ validation.completed (event)
│  ├─ db.exec: INSERT user (100ms)
│  └─ Attributes: user_email, user_platform
└─ HTTP Response 200
```

## 📖 More Examples

See [EXAMPLES.md](./EXAMPLES.md) for comprehensive examples including:
- Error handling patterns
- Validation-heavy operations
- Batch processing
- Async operations
- Performance-critical sections
- And more!

## 🆘 Troubleshooting

### Spans not appearing in Jaeger

1. **Check if tracing is enabled**: `OTEL_ENABLE_TRACING=true` in `.env`
2. **Check sampling rate**: `OTEL_TRACE_SAMPLE_RATE=1.0` for all traces
3. **Verify Jaeger is running**: `docker ps | grep jaeger`
4. **Check you called span.End()**: Use `defer span.End()`

### Too many spans / performance issues

1. **Reduce sampling rate**: `OTEL_TRACE_SAMPLE_RATE=0.1` (10%)
2. **Remove unnecessary instrumentation**: Only trace important operations
3. **Use `IsSpanRecording()`**: Skip expensive attribute collection when not recording

### Attributes not showing in Jaeger

1. **Check attribute types**: Use correct type functions (`StringAttr`, `IntAttr`, etc.)
2. **Verify attribute names**: Use semantic constants (`trace.AttrUserID`)
3. **Call before span.End()**: Attributes must be added before span ends

## 🔗 Related Documentation

- [Full Observability Guide](../../../OBSERVABILITY.md)
- [Usage Examples](./EXAMPLES.md)
- [OpenTelemetry Documentation](https://opentelemetry.io/docs/instrumentation/go/manual/)
