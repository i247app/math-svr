package trace

import (
	"go.opentelemetry.io/otel/attribute"
)

// Common attribute keys for consistent naming across the application
const (
	// User-related attributes
	AttrUserID       = "user.id"
	AttrUserUID      = "user.uid"
	AttrUserEmail    = "user.email"
	AttrUserPlatform = "user.platform"

	// Profile-related attributes
	AttrProfileID    = "profile.id"
	AttrGradeLevel   = "profile.grade_level"
	AttrSemester     = "profile.semester"

	// Quiz-related attributes
	AttrQuizID         = "quiz.id"
	AttrQuizType       = "quiz.type"
	AttrQuizDifficulty = "quiz.difficulty"
	AttrQuizTopic      = "quiz.topic"
	AttrQuestionCount  = "quiz.question_count"

	// Assessment-related attributes
	AttrAssessmentID    = "assessment.id"
	AttrScore           = "assessment.score"
	AttrTotalQuestions  = "assessment.total_questions"
	AttrCorrectAnswers  = "assessment.correct_answers"
	AttrScorePercentage = "assessment.score_percentage"

	// Operation-related attributes
	AttrOperationType = "operation.type"
	AttrOperationName = "operation.name"
	AttrResourceID    = "resource.id"
	AttrResourceType  = "resource.type"

	// Validation-related attributes
	AttrValidationPassed = "validation.passed"
	AttrValidationErrors = "validation.errors"

	// Business logic attributes
	AttrItemCount      = "items.count"
	AttrProcessedCount = "items.processed"
	AttrFailedCount    = "items.failed"
	AttrBatchSize      = "batch.size"
)

// UserAttributes creates common user-related attributes
func UserAttributes(uid, email, platform string) []attribute.KeyValue {
	attrs := make([]attribute.KeyValue, 0, 3)
	if uid != "" {
		attrs = append(attrs, attribute.String(AttrUserUID, uid))
	}
	if email != "" {
		attrs = append(attrs, attribute.String(AttrUserEmail, email))
	}
	if platform != "" {
		attrs = append(attrs, attribute.String(AttrUserPlatform, platform))
	}
	return attrs
}

// ProfileAttributes creates common profile-related attributes
func ProfileAttributes(profileID, gradeLevel, semester string) []attribute.KeyValue {
	attrs := make([]attribute.KeyValue, 0, 3)
	if profileID != "" {
		attrs = append(attrs, attribute.String(AttrProfileID, profileID))
	}
	if gradeLevel != "" {
		attrs = append(attrs, attribute.String(AttrGradeLevel, gradeLevel))
	}
	if semester != "" {
		attrs = append(attrs, attribute.String(AttrSemester, semester))
	}
	return attrs
}

// QuizAttributes creates common quiz-related attributes
func QuizAttributes(quizID, quizType, difficulty, topic string, questionCount int) []attribute.KeyValue {
	attrs := make([]attribute.KeyValue, 0, 5)
	if quizID != "" {
		attrs = append(attrs, attribute.String(AttrQuizID, quizID))
	}
	if quizType != "" {
		attrs = append(attrs, attribute.String(AttrQuizType, quizType))
	}
	if difficulty != "" {
		attrs = append(attrs, attribute.String(AttrQuizDifficulty, difficulty))
	}
	if topic != "" {
		attrs = append(attrs, attribute.String(AttrQuizTopic, topic))
	}
	if questionCount > 0 {
		attrs = append(attrs, attribute.Int(AttrQuestionCount, questionCount))
	}
	return attrs
}

// AssessmentAttributes creates common assessment-related attributes
func AssessmentAttributes(assessmentID string, totalQuestions, correctAnswers int, scorePercentage float64) []attribute.KeyValue {
	attrs := make([]attribute.KeyValue, 0, 4)
	if assessmentID != "" {
		attrs = append(attrs, attribute.String(AttrAssessmentID, assessmentID))
	}
	if totalQuestions > 0 {
		attrs = append(attrs, attribute.Int(AttrTotalQuestions, totalQuestions))
	}
	if correctAnswers >= 0 {
		attrs = append(attrs, attribute.Int(AttrCorrectAnswers, correctAnswers))
	}
	if scorePercentage >= 0 {
		attrs = append(attrs, attribute.Float64(AttrScorePercentage, scorePercentage))
	}
	return attrs
}

// OperationAttributes creates common operation-related attributes
func OperationAttributes(operationType, operationName string) []attribute.KeyValue {
	attrs := make([]attribute.KeyValue, 0, 2)
	if operationType != "" {
		attrs = append(attrs, attribute.String(AttrOperationType, operationType))
	}
	if operationName != "" {
		attrs = append(attrs, attribute.String(AttrOperationName, operationName))
	}
	return attrs
}

// ResourceAttributes creates common resource-related attributes
func ResourceAttributes(resourceID, resourceType string) []attribute.KeyValue {
	attrs := make([]attribute.KeyValue, 0, 2)
	if resourceID != "" {
		attrs = append(attrs, attribute.String(AttrResourceID, resourceID))
	}
	if resourceType != "" {
		attrs = append(attrs, attribute.String(AttrResourceType, resourceType))
	}
	return attrs
}

// ValidationAttributes creates validation-related attributes
func ValidationAttributes(passed bool, errorCount int) []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.Bool(AttrValidationPassed, passed),
		attribute.Int(AttrValidationErrors, errorCount),
	}
}

// BatchProcessingAttributes creates batch processing attributes
func BatchProcessingAttributes(batchSize, processedCount, failedCount int) []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.Int(AttrBatchSize, batchSize),
		attribute.Int(AttrProcessedCount, processedCount),
		attribute.Int(AttrFailedCount, failedCount),
	}
}

// StringAttr is a convenience function for creating string attributes
func StringAttr(key, value string) attribute.KeyValue {
	return attribute.String(key, value)
}

// IntAttr is a convenience function for creating int attributes
func IntAttr(key string, value int) attribute.KeyValue {
	return attribute.Int(key, value)
}

// Int64Attr is a convenience function for creating int64 attributes
func Int64Attr(key string, value int64) attribute.KeyValue {
	return attribute.Int64(key, value)
}

// Float64Attr is a convenience function for creating float64 attributes
func Float64Attr(key string, value float64) attribute.KeyValue {
	return attribute.Float64(key, value)
}

// BoolAttr is a convenience function for creating bool attributes
func BoolAttr(key string, value bool) attribute.KeyValue {
	return attribute.Bool(key, value)
}
