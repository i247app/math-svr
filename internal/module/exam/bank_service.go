package exam

import (
	"cmp"
	"context"
	"fmt"
	"slices"

	command "math-ai.com/math-ai/internal/application/command/exam"
	"math-ai.com/math-ai/internal/application/dto/aiquestion"
	dto "math-ai.com/math-ai/internal/application/dto/exam"
	profileDomain "math-ai.com/math-ai/internal/domain/profile"
	"math-ai.com/math-ai/internal/domain/prompt/input"
	"math-ai.com/math-ai/internal/domain/questionpool"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/utils"
)

// bankChallengeSlots are the positions of an exam's harder questions — the
// probes of an ASSESSMENT (one grade up) and the challenge questions of a
// GRADE paper (one level up). The stored GRADE prompt puts them at Q3 and
// Q6 too; the rest of the paper is the child's own grade / level.
var bankChallengeSlots = []int{3, 6}

// Texts of the two options a TRUE_FALSE bank question is served with.
const (
	trueOptionText  = "Đúng"
	falseOptionText = "Sai"
)

// generateFromBank hands out an ASSESSMENT / GRADE paper drawn from the
// question bank instead of asking the model (QUESTION_POOL_ENABLED). The
// drawn questions are stored as a new ma_exam_pools row of source POOL,
// with no cache tag, so everything after this — the sitting, its shuffle,
// grading, the answer log — is the path an LLM paper takes.
func (s *Service) generateFromBank(ctx context.Context, req *dto.GenerateExamReq, profile *profileDomain.Profile,
	cmd command.GenerateExamCommand, level *int) (*dto.GenerateExamRes, error) {

	avoid, err := s.recentStems(ctx, profile.ProfileId(), cmd.Grade)
	if err != nil {
		return nil, err
	}

	//TODO: may be turn on avoid question later
	questions, err := s.pickFromBank(ctx, cmd.ExamType, cmd.Grade, level, []string{})
	if err != nil {
		return nil, err
	}
	raw, err := marshalQuestions(ctx, questions)
	if err != nil {
		return nil, err
	}

	cmd.NewContent = &command.NewExamPoolContent{
		NumQues:       len(questions),
		Level:         level,
		Semester:      utils.ToStringPtr(req.Semester),
		Program:       utils.ToStringPtr(req.Program),
		Title:         utils.ToStringPtr(bankTitle(cmd.ExamType, cmd.Grade, level)),
		QuestionsJSON: raw,
		Source:        enum.ExamSourcePool,
	}
	logger.From(ctx).Infof("exam.bank.drawn profile=%d type=%s grade=%d level=%s questions=%d avoided=%d",
		profile.ProfileId(), cmd.ExamType, cmd.Grade, levelLabel(level), len(questions), len(avoid))
	return s.handOut(ctx, cmd, questions, profile)
}

// pickFromBank draws a whole paper, numbered 1..N, or fails with
// EXAM_QUESTION_BANK_INSUFFICIENT naming the group that came up short.
func (s *Service) pickFromBank(ctx context.Context, examType enum.ExamType, grade int, level *int, avoid []string) ([]aiquestion.Question, error) {
	paper := make([]aiquestion.Question, DefaultNumQuestions)
	var chosen []int64

	for _, g := range bankGroups(examType, grade, level) {
		picked, err := s.drawGroup(ctx, g, chosen, avoid)
		if err != nil {
			return nil, err
		}
		// Easiest first, so a paper that is served in stored order still rises.
		slices.SortStableFunc(picked, func(a, b *questionpool.QuestionPool) int {
			return cmp.Compare(difficulty(a), difficulty(b))
		})
		for i, q := range picked {
			slot := g.slots[i]
			converted, err := toExamQuestion(q, slot)
			if err != nil {
				return nil, errs.NewError(ctx, status.EXAM_GENERATION_FAILED, nil, err)
			}
			paper[slot-1] = converted
			chosen = append(chosen, q.QuestionPoolId())
		}
	}
	return paper, nil
}

// drawGroup draws len(g.slots) questions, preferring ones whose stem the
// child has not met recently; only when the bank cannot fill the group that
// way does it accept a recent one.
func (s *Service) drawGroup(ctx context.Context, g bankGroup, chosen []int64, avoid []string) ([]*questionpool.QuestionPool, error) {
	need := len(g.slots)
	f := g.filter
	f.ExcludeIds, f.AvoidTexts = chosen, avoid

	got, err := s.questionPoolRepo.PickForExam(ctx, f, need)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if len(got) < need && len(avoid) > 0 {
		f.AvoidTexts = nil
		f.ExcludeIds = slices.Clone(chosen)
		for _, q := range got {
			f.ExcludeIds = append(f.ExcludeIds, q.QuestionPoolId())
		}
		more, err := s.questionPoolRepo.PickForExam(ctx, f, need-len(got))
		if err != nil {
			return nil, errs.NewError(ctx, status.FAIL, nil, err)
		}
		got = append(got, more...)
	}
	if len(got) < need {
		logger.From(ctx).Warnf("exam.bank.insufficient group=%q need=%d found=%d", g.name, need, len(got))
		return nil, errs.NewError(ctx, status.EXAM_QUESTION_BANK_INSUFFICIENT,
			map[string]any{"group": g.name, "needed": need, "found": len(got)},
			fmt.Errorf("exam: the question bank holds %d approved questions for %s, the paper needs %d", len(got), g.name, need))
	}
	return got, nil
}

// bankGroup is the part of a paper drawn under one filter, filling slots
// (1-based question positions) in ascending difficulty.
type bankGroup struct {
	name   string
	filter questionpool.PickFilter
	slots  []int
}

// bankGroups plans a paper of DefaultNumQuestions questions.
//
//   - ASSESSMENT: the challenge slots come from grade + 1, the rest from the grade.
//   - GRADE with a level: the challenge slots come from level + 1 (level 9 has
//     none above it and draws level 9 again), the rest from the level.
//   - GRADE without a level: the whole paper from the grade.
func bankGroups(examType enum.ExamType, grade int, level *int) []bankGroup {
	var all, challenge []int
	for slot := 1; slot <= DefaultNumQuestions; slot++ {
		if slices.Contains(bankChallengeSlots, slot) {
			challenge = append(challenge, slot)
		} else {
			all = append(all, slot)
		}
	}

	switch {
	case examType == enum.ExamTypeGrade && level != nil:
		up := min(*level+1, enum.ExamLevelMax)
		return []bankGroup{
			{
				name:   fmt.Sprintf("grade %d level %d", grade, *level),
				filter: questionpool.PickFilter{Grade: grade, Level: level},
				slots:  all,
			},
			{
				name:   fmt.Sprintf("grade %d level %d (challenge)", grade, up),
				filter: questionpool.PickFilter{Grade: grade, Level: &up},
				slots:  challenge,
			},
		}
	case examType == enum.ExamTypeAssessment:
		up := input.ProbeGrade(grade)
		return []bankGroup{
			{
				name:   fmt.Sprintf("grade %d", grade),
				filter: questionpool.PickFilter{Grade: grade},
				slots:  all,
			},
			{
				name:   fmt.Sprintf("grade %d (probe)", up),
				filter: questionpool.PickFilter{Grade: up},
				slots:  challenge,
			},
		}
	}
	every := slices.Concat(all, challenge)
	slices.Sort(every)

	return []bankGroup{
		{
			name:   fmt.Sprintf("grade %d", grade),
			filter: questionpool.PickFilter{Grade: grade},
			slots:  every,
		},
	}
}

// difficulty orders questions within a group: level, then sub-level. A
// question that is not rated counts as the middle of the scale.
func difficulty(q *questionpool.QuestionPool) int {
	const unrated = 4
	level, sub := unrated, unrated
	if l := q.QuestionLevel(); l != nil {
		level = *l
	}
	if l := q.QuestionSubLevel(); l != nil {
		sub = *l
	}
	return level*10 + sub
}

// toExamQuestion converts a bank question into the exam's question at
// position number. Options keep their bank ids, in display order; the key
// follows from the kind (aiquestion.Question documents the four shapes).
// Image keys travel with the question and are presigned when a response is
// built (image_service.go).
func toExamQuestion(q *questionpool.QuestionPool, number int) (aiquestion.Question, error) {
	out := aiquestion.Question{
		QuestionNumber: number,
		// The bank carries no render hint; the client draws a bank question
		// from its kind.
		QuestionType:  aiquestion.TypeArithmetic,
		QuestionName:  q.QuestionText(),
		QuestionTopic: utils.DerefString(q.QuestionTopic()),
		Kind:          string(q.QuestionKind()),
	}
	grade := q.QuestionGrade()
	out.QuestionGrade = &grade
	out.ImageKey = utils.DerefString(q.QuestionImageKey())

	textOf := make(map[string]string, len(q.Options()))
	for _, o := range q.Options() {
		textOf[o.Id] = o.Text
		out.Answers = append(out.Answers, aiquestion.AnswerChoice{Label: o.Id, Content: o.Text, ImageKey: utils.DerefString(o.ImageKey)})
	}
	correct := q.CorrectAnswer()

	switch q.QuestionKind() {
	case questionpool.KindSingleChoice:
		if len(correct.OptionIds) != 1 {
			return out, fmt.Errorf("question %d: single-choice key does not name exactly one option", q.QuestionPoolId())
		}
		text, ok := textOf[correct.OptionIds[0]]
		if !ok {
			return out, fmt.Errorf("question %d: single-choice key names unknown option %q", q.QuestionPoolId(), correct.OptionIds[0])
		}
		out.RightAnswerLabel = correct.OptionIds[0]
		out.RightAnswerContent = text
	case questionpool.KindMultipleChoice:
		for _, id := range correct.OptionIds {
			if _, ok := textOf[id]; !ok {
				return out, fmt.Errorf("question %d: multiple-choice key names unknown option %q", q.QuestionPoolId(), id)
			}
			out.RightOptionIDs = append(out.RightOptionIDs, id)
		}
		slices.Sort(out.RightOptionIDs)
		if len(out.RightOptionIDs) == 0 {
			return out, fmt.Errorf("question %d: multiple-choice key is empty", q.QuestionPoolId())
		}
	case questionpool.KindTrueFalse:
		if correct.Value == nil {
			return out, fmt.Errorf("question %d: true/false key has no value", q.QuestionPoolId())
		}
		out.Answers = []aiquestion.AnswerChoice{
			{Label: aiquestion.TrueOptionID, Content: trueOptionText},
			{Label: aiquestion.FalseOptionID, Content: falseOptionText},
		}
		out.RightAnswerLabel = aiquestion.FalseOptionID
		if *correct.Value {
			out.RightAnswerLabel = aiquestion.TrueOptionID
		}
	case questionpool.KindNumericInput:
		if len(correct.Values) == 0 {
			return out, fmt.Errorf("question %d: numeric key has no values", q.QuestionPoolId())
		}
		out.RightValues = slices.Clone(correct.Values)
	default:
		return out, fmt.Errorf("question %d: unknown kind %q", q.QuestionPoolId(), q.QuestionKind())
	}
	return out, nil
}

// bankTitle is the title of a bank paper, which has no model-written one.
func bankTitle(examType enum.ExamType, grade int, level *int) string {
	name := fmt.Sprintf("Lớp %d", grade)
	if grade == enum.ExamGradeMin {
		name = "Mẫu giáo"
	}
	if examType == enum.ExamTypeGrade && level != nil {
		return fmt.Sprintf("%s – Level %d", name, *level)
	}
	return "Bài đánh giá " + name
}

func levelLabel(level *int) string {
	if level == nil {
		return "NA"
	}
	return fmt.Sprint(*level)
}
