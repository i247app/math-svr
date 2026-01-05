package metrics

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var (
	// User metrics
	userSignupCounter metric.Int64Counter
	userLoginCounter  metric.Int64Counter

	// Quiz generation metrics
	quizGenerationCounter metric.Int64Counter
	quizGenerationErrors  metric.Int64Counter

	// Assessment metrics
	assessmentSubmissionCounter metric.Int64Counter
	assessmentScoreHistogram    metric.Float64Histogram

	// Profile metrics
	profileCreationCounter metric.Int64Counter
	profileUpdateCounter   metric.Int64Counter
)

func init() {
	var err error

	// User metrics
	userSignupCounter, err = meter.Int64Counter(
		"business.user.signup.count",
		metric.WithDescription("Total number of user signups"),
		metric.WithUnit("{user}"),
	)
	if err != nil {
		panic(err)
	}

	userLoginCounter, err = meter.Int64Counter(
		"business.user.login.count",
		metric.WithDescription("Total number of user logins"),
		metric.WithUnit("{login}"),
	)
	if err != nil {
		panic(err)
	}

	// Quiz generation metrics
	quizGenerationCounter, err = meter.Int64Counter(
		"business.quiz.generation.count",
		metric.WithDescription("Total number of quizzes generated"),
		metric.WithUnit("{quiz}"),
	)
	if err != nil {
		panic(err)
	}

	quizGenerationErrors, err = meter.Int64Counter(
		"business.quiz.generation.errors",
		metric.WithDescription("Total number of quiz generation errors"),
		metric.WithUnit("{error}"),
	)
	if err != nil {
		panic(err)
	}

	// Assessment metrics
	assessmentSubmissionCounter, err = meter.Int64Counter(
		"business.assessment.submission.count",
		metric.WithDescription("Total number of assessment submissions"),
		metric.WithUnit("{submission}"),
	)
	if err != nil {
		panic(err)
	}

	assessmentScoreHistogram, err = meter.Float64Histogram(
		"business.assessment.score",
		metric.WithDescription("Distribution of assessment scores"),
		metric.WithUnit("{score}"),
	)
	if err != nil {
		panic(err)
	}

	// Profile metrics
	profileCreationCounter, err = meter.Int64Counter(
		"business.profile.creation.count",
		metric.WithDescription("Total number of profiles created"),
		metric.WithUnit("{profile}"),
	)
	if err != nil {
		panic(err)
	}

	profileUpdateCounter, err = meter.Int64Counter(
		"business.profile.update.count",
		metric.WithDescription("Total number of profile updates"),
		metric.WithUnit("{update}"),
	)
	if err != nil {
		panic(err)
	}
}

// RecordUserSignup records a user signup event
func RecordUserSignup(platform string) {
	ctx := context.Background()

	attrs := []attribute.KeyValue{
		attribute.String("platform", platform),
	}

	userSignupCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
}

// RecordUserLogin records a user login event
func RecordUserLogin(platform, method string, success bool) {
	ctx := context.Background()

	status := "success"
	if !success {
		status = "failure"
	}

	attrs := []attribute.KeyValue{
		attribute.String("platform", platform),
		attribute.String("method", method),
		attribute.String("status", status),
	}

	userLoginCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
}

// RecordQuizGeneration records a quiz generation event
func RecordQuizGeneration(gradeLevel, topic, difficulty string, success bool) {
	ctx := context.Background()

	attrs := []attribute.KeyValue{
		attribute.String("grade_level", gradeLevel),
		attribute.String("topic", topic),
		attribute.String("difficulty", difficulty),
	}

	if success {
		quizGenerationCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
	} else {
		quizGenerationErrors.Add(ctx, 1, metric.WithAttributes(attrs...))
	}
}

// RecordAssessmentSubmission records an assessment submission
func RecordAssessmentSubmission(gradeLevel, quizType string, score float64, totalQuestions int) {
	ctx := context.Background()

	// Calculate percentage score
	percentageScore := (score / float64(totalQuestions)) * 100

	attrs := []attribute.KeyValue{
		attribute.String("grade_level", gradeLevel),
		attribute.String("quiz_type", quizType),
		attribute.Int("total_questions", totalQuestions),
	}

	assessmentSubmissionCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
	assessmentScoreHistogram.Record(ctx, percentageScore, metric.WithAttributes(attrs...))
}

// RecordProfileCreation records a profile creation event
func RecordProfileCreation(gradeLevel string) {
	ctx := context.Background()

	attrs := []attribute.KeyValue{
		attribute.String("grade_level", gradeLevel),
	}

	profileCreationCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
}

// RecordProfileUpdate records a profile update event
func RecordProfileUpdate(updateType string) {
	ctx := context.Background()

	attrs := []attribute.KeyValue{
		attribute.String("update_type", updateType),
	}

	profileUpdateCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
}
