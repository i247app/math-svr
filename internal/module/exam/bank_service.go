package exam

import (
	"cmp"
	"context"
	"fmt"
	"slices"

	command "math-ai.com/math-ai/internal/application/command/exam"
	"math-ai.com/math-ai/internal/application/dto"

	domain "math-ai.com/math-ai/internal/domain/exam"
	profileDomain "math-ai.com/math-ai/internal/domain/profile"
	input "math-ai.com/math-ai/internal/domain/prompt/build"
	"math-ai.com/math-ai/internal/domain/questionpool"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/errs"
	"math-ai.com/math-ai/internal/shared/utils"
)

// bankChallengeSlots are the positions of an exam's harder questions — the
// probes of an ASSESSMENT (one grade up) and the challenge questions of a
// GRADE paper (one level up). The stored GRADE prompt puts them at Q3 and
// Q6 too; the rest of the paper is the child's own grade / level.
var bankChallengeSlots = []int{3, 6}

// assessmentBankTags are the tags every question of an ASSESSMENT drawn
// from the bank must carry, all of them, probes included.
var assessmentBankTags = []string{questionpool.TagOnlyNumber}

// kindergartenGrade (Mẫu giáo) questions are exempt from
// assessmentBankTags. The rule follows the QUESTION's grade, so the grade-1
// probes of a kindergarten paper still carry the tags.
const kindergartenGrade = enum.ExamGradeMin

// assessmentTags returns the tags an ASSESSMENT draws questions of
// questionGrade with.
func assessmentTags(questionGrade int) []string {
	if questionGrade == kindergartenGrade {
		return nil
	}
	return assessmentBankTags
}

// isBankInsufficient reports whether err says the bank cannot fill a paper
// (pickFromBank): nothing was stored, so the caller may ask the model instead.
func isBankInsufficient(err error) bool {
	me, ok := errs.IsMathError(err)
	return ok && me.GetStatusCode() == domain.ExamQuestionBankInsufficient
}

// generateFromBank hands out an ASSESSMENT / GRADE paper drawn from the
// question bank instead of asking the model (QUESTION_POOL_ENABLED). The
// drawn questions are stored as a new ma_exam_pools row of source POOL,
// with no cache tag, so everything after this — the sitting, its shuffle,
// grading, the answer log — is the path an LLM paper takes.
func (s *Service) generateFromBank(
	ctx context.Context,
	req *dto.GenerateExamReq,
	profile *profileDomain.Profile,
	cmd command.GenerateExamCommand,
	level *int,
) (*dto.GenerateExamRes, error) {

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
func (s *Service) pickFromBank(ctx context.Context, examType enum.ExamType, grade int, level *int, avoid []string) ([]dto.Question, error) {
	paper := make([]dto.Question, DefaultNumQuestions)
	var chosen []int64

	bankGroups := bankGroups(examType, grade, level)

	for _, g := range bankGroups {
		picked, err := s.drawGroup(ctx, g, chosen, avoid)
		if err != nil {
			return nil, err
		}
		// Easiest first, so a paper that is served in stored order still rises.
		slices.SortStableFunc(picked, func(a, b *questionpool.QuestionPool) int {
			return cmp.Compare(dto.BankDifficulty(a), dto.BankDifficulty(b))
		})
		for i, q := range picked {
			slot := g.slots[i]
			converted, err := dto.QuestionFromBank(q, slot)
			if err != nil {
				return nil, errs.NewError(ctx, domain.ExamGenerationFailed, nil, err)
			}
			paper[slot-1] = converted
			chosen = append(chosen, q.QuestionPoolId())
		}
	}
	return paper, nil
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
//   - ASSESSMENT: one group per position, at that position's allowed levels
//     (input.AllowedLevels); the challenge slots come from grade + 1, the
//     rest from the grade; every question must carry every
//     assessmentBankTags tag, except a kindergarten (grade 0) question — the
//     probes of a grade-0 paper (grade 1) still carry them.
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
				filter: questionpool.PickFilter{Grade: grade, Levels: []int{*level}},
				slots:  all,
			},
			{
				name:   fmt.Sprintf("grade %d level %d (challenge)", grade, up),
				filter: questionpool.PickFilter{Grade: grade, Levels: []int{up}},
				slots:  challenge,
			},
		}
	case examType == enum.ExamTypeAssessment:
		// One group per position: each has its own allowed levels
		// (input.AllowedLevels). Every question, probes included, is a bare
		// calculation — except a kindergarten question (assessmentTags).
		// ponytail: one ORDER BY RAND() query per slot (10 per paper), each
		// over one grade's approved rows at those levels; one windowed query
		// (ROW_NUMBER() OVER (PARTITION BY question_level)) if it shows.
		gradeUp := input.ProbeGrade(grade)
		allowed := input.AllowedLevels(examType, DefaultNumQuestions, nil)
		groups := make([]bankGroup, 0, DefaultNumQuestions)
		for slot := 1; slot <= DefaultNumQuestions; slot++ {
			g, probe := grade, ""
			if slices.Contains(bankChallengeSlots, slot) {
				g, probe = gradeUp, " (probe)"
			}
			tags, levels := assessmentTags(g), allowed[slot-1]
			groups = append(groups, bankGroup{
				name:   fmt.Sprintf("Q%d grade %d%s levels %v %v", slot, g, probe, levels, tags),
				filter: questionpool.PickFilter{Grade: g, Levels: levels, Tags: tags},
				slots:  []int{slot},
			})
		}
		return groups
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
		return nil, errs.NewError(ctx, domain.ExamQuestionBankInsufficient,
			map[string]any{"group": g.name, "needed": need, "found": len(got)},
			fmt.Errorf("exam: the question bank holds %d approved questions for %s, the paper needs %d", len(got), g.name, need))
	}
	return got, nil
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
