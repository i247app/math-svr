package exam

import (
	"context"

	botAdapter "math-ai.com/math-ai/internal/adapter/bot"
	dto "math-ai.com/math-ai/internal/application/dto/exam"
	"math-ai.com/math-ai/internal/application/dto/question"
	domainBot "math-ai.com/math-ai/internal/domain/bot"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/utils"
)

// botClient is the exam module's typed surface over the generic bot
// adapter: prompt in, validated questions out.
//
// It has one method. Grading never reaches the model — the server scores
// every submission — so unlike the quiz module there is no GradeExam here
// and no grading prompt behind it.
//
// Temperature 0.2 keeps phrasing varied between calls without letting the
// schema drift; JSON mode is requested even though the system prompt
// mandates the shape anyway, because backends that honour it fail earlier
// and more cheaply.
type botClient struct {
	adapter *botAdapter.Adapter
}

func newBotClient(adapter *botAdapter.Adapter) *botClient {
	return &botClient{adapter: adapter}
}

type generateExamInput struct {
	ExamType     enum.ExamType
	Grade        int
	NumQuestions int
	Semester     string
	Program      string
	// Level is the clamped intensity for a GRADE review; nil otherwise.
	Level *int
	// Practice aims a PRACTICE round; nil for every other type.
	Practice *domainBot.PracticeBrief
}

type generateExamOutput struct {
	Title     string
	ShortText string
	Questions []question.Question
}

func (c *botClient) GenerateExam(ctx context.Context, in generateExamInput) (*generateExamOutput, error) {
	log := logger.From(ctx)
	if c.adapter == nil {
		return nil, errs.NewError(ctx, status.BOT_CONFIG_INVALID, nil, ErrBotAdapterNotConfigured)
	}

	promptIn := domainBot.ExamPromptInput{
		ExamType:     in.ExamType,
		Grade:        in.Grade,
		NumQuestions: in.NumQuestions,
		Semester:     in.Semester,
		Program:      in.Program,
		Level:        in.Level,
		Practice:     in.Practice,
	}

	system, user, err := domainBot.BuildExamPrompt(promptIn)
	if err != nil {
		return nil, errs.NewError(ctx, status.EXAM_GENERATION_FAILED, nil, err)
	}

	log.Infof("PROMPT GENERATE EXAM: system=%s user=%s", system, user)

	res, err := c.adapter.Chat(ctx, botAdapter.ChatRequest{
		Messages: []botAdapter.Message{
			{Role: botAdapter.RoleSystem, Content: system},
			{Role: botAdapter.RoleUser, Content: user},
		},
		Temperature: 0.2,
		TopP:        0.95,
		JSONMode:    true,
	})
	if err != nil {
		return nil, err
	}

	log.Infof("BOT RESPONSE: %s", res.Content)

	// Parsed with the exam vocabulary (right_answer_label, question_topic,
	// ...) and converted straight into the shared shape, which is what the
	// normalisers and the scorer operate on.
	gen, err := question.ParseGenerationOf[dto.ExamQuestion](res.Content)
	if err != nil {
		log.Warnf("exam.bot.generate_parse_failed err=%v", err)
		return nil, errs.NewError(ctx, status.EXAM_GENERATION_FAILED,
			map[string]any{"reason": err.Error()}, err)
	}
	if len(gen.Questions) == 0 {
		return nil, errs.NewError(ctx, status.EXAM_NO_QUESTIONS, nil, ErrModelReturnedNothing)
	}

	// Two normalisation passes, and they answer different questions.
	//
	// The first clamps render types and reports icon drift: cosmetic, never
	// fatal, since grading is label-based.
	questions, warnings := question.Normalize(dto.ToSharedQuestions(gen.Questions))
	for _, w := range warnings {
		log.Warnf("exam.normalize.%s value=%q q=%d", w.Kind, w.Value, w.QuestionNumber)
	}

	// The second stamps question_grade from the request. This one is not
	// cosmetic: question_grade feeds placement, so a value the model
	// invented could move a child up or down a year. The mismatches are
	// logged because a run of them means the prompt has stopped landing.
	questions, mismatches := NormalizeQuestionBands(questions, in.ExamType, in.Grade)
	for _, m := range mismatches {
		log.Warnf("exam.question_grade.mismatch q=%d model=%v applied=%d",
			m.QuestionNumber, utils.DerefInt(m.ModelGrade), m.Applied)
	}
	questions = StampQuestionLevel(questions, in.Level)

	return &generateExamOutput{
		Title:     gen.Title,
		ShortText: gen.ShortText,
		Questions: questions,
	}, nil
}
