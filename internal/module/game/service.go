// Package game serves the game: rounds of bare-calculation questions drawn
// from the question bank. It is independent of module/exam — a round never
// counts toward a child's assessed ability — and makes no model call.
package game

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"slices"

	command "math-ai.com/math-ai/internal/application/command/game"
	"math-ai.com/math-ai/internal/application/dto"
	query "math-ai.com/math-ai/internal/application/query/game"
	"math-ai.com/math-ai/internal/application/transaction"
	domain "math-ai.com/math-ai/internal/domain/game"
	gradeDomain "math-ai.com/math-ai/internal/domain/grade"
	profileDomain "math-ai.com/math-ai/internal/domain/profile"
	input "math-ai.com/math-ai/internal/domain/prompt/build"
	"math-ai.com/math-ai/internal/domain/questionpool"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/infrastructure/metadata"
	"math-ai.com/math-ai/internal/provider/storage"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/errs"
	"math-ai.com/math-ai/internal/shared/utils"
)

// gameBankTags are the tags every question of a round must carry,
// kindergarten included.
var gameBankTags = []string{questionpool.TagOnlyNumber}

type Service struct {
	startCmd  *command.StartGameCommandHandler
	submitCmd *command.SubmitGameCommandHandler
	getQuery  *query.GetGameSessionQueryHandler
	listQuery *query.ListGameSessionsQueryHandler

	// questionPoolRepo is read-only here: the bank is managed through
	// module/questionpool.
	questionPoolRepo questionpool.IRepository
	profileRepo      profileDomain.IRepository
	gradeRepo        gradeDomain.IRepository
	signer           urlSigner // nil without storage
}

func NewService(
	uow transaction.UnitOfWork,
	gameRepo domain.IRepository,
	questionPoolRepo questionpool.IRepository,
	profileRepo profileDomain.IRepository,
	gradeRepo gradeDomain.IRepository,
	storageProvider *storage.Registry,
) *Service {
	s := &Service{
		startCmd:         command.NewStartGameCommandHandler(uow),
		submitCmd:        command.NewSubmitGameCommandHandler(uow),
		getQuery:         query.NewGetGameSessionQueryHandler(gameRepo),
		listQuery:        query.NewListGameSessionsQueryHandler(gameRepo),
		questionPoolRepo: questionPoolRepo,
		profileRepo:      profileRepo,
		gradeRepo:        gradeRepo,
	}
	if storageProvider != nil { // a nil *Registry in the interface would not be nil
		s.signer = storageProvider
	}
	return s
}

// StartGame opens a new round, always: an unfinished one is left behind.
func (s *Service) StartGame(ctx context.Context, req *dto.StartGameReq) (*dto.GameSessionRes, error) {
	if err := ValidateStartGame(ctx, req); err != nil {
		return nil, err
	}
	profile, err := s.loadPlayer(ctx, req.UID, req.ProfileID)
	if err != nil {
		return nil, err
	}
	grade, err := s.resolveGrade(ctx, req.Grade, profile)
	if err != nil {
		return nil, err
	}
	questionsJSON, err := s.drawQuestions(ctx, grade, req.Level, *req.NumQuestions)
	if err != nil {
		return nil, err
	}

	created, err := s.startCmd.Handle(ctx, command.StartGameCommand{
		UID:           profile.UID(),
		ProfileID:     profile.ProfileId(),
		Grade:         grade,
		Level:         req.Level,
		NumQuestions:  *req.NumQuestions,
		QuestionsJSON: questionsJSON,
	})
	if err != nil {
		return nil, err
	}
	logger.From(ctx).Infof("game.started gsess_id=%d profile=%d grade=%d level=%s questions=%d",
		created.GsessId(), profile.ProfileId(), grade, levelLabel(req.Level), *req.NumQuestions)
	return s.render(ctx, created), nil
}

// SubmitGame grades a round and returns it with the results.
func (s *Service) SubmitGame(ctx context.Context, req *dto.SubmitGameReq) (*dto.GameSessionRes, error) {
	if err := ValidateGameIDs(ctx, req.ProfileID, req.GsessID); err != nil {
		return nil, err
	}
	profile, err := s.loadPlayer(ctx, req.UID, req.ProfileID)
	if err != nil {
		return nil, err
	}
	submitted, err := s.submitCmd.Handle(ctx, command.SubmitGameCommand{
		GsessID:   req.GsessID,
		UID:       profile.UID(),
		ProfileID: profile.ProfileId(),
		Answers:   req.Answers,
	})
	if err != nil {
		return nil, err
	}
	logger.From(ctx).Infof("game.submitted gsess_id=%d profile=%d correct=%d total=%d skipped=%d pct=%d",
		submitted.GsessId(), profile.ProfileId(), utils.DerefInt(submitted.ResCorrectNumber()),
		utils.DerefInt(submitted.ResTotalQuestions()), utils.DerefInt(submitted.ResSkippedNumber()),
		utils.DerefInt(submitted.ResScorePercentage()))
	return s.render(ctx, submitted), nil
}

func (s *Service) ListGames(ctx context.Context, req *dto.ListGamesReq) (*dto.ListGamesRes, error) {
	if err := ValidateListGames(ctx, req); err != nil {
		return nil, err
	}
	profile, err := s.loadPlayer(ctx, req.UID, req.ProfileID)
	if err != nil {
		return nil, err
	}
	rows, err := s.listQuery.Handle(ctx, query.ListGameSessionsQuery{
		UID:       profile.UID(),
		ProfileID: profile.ProfileId(),
		Paging:    req.Request,
	})
	if err != nil {
		return nil, err
	}
	return &dto.ListGamesRes{
		GameSessions: dto.GameSessionsToResponse(rows.GameSessions, requestLanguage(ctx)),
		Pagination:   rows.Pagination,
		Cursor:       rows.Cursor,
	}, nil
}

func (s *Service) GetGame(ctx context.Context, req *dto.GetGameReq) (*dto.GameSessionRes, error) {
	if err := ValidateGameIDs(ctx, req.ProfileID, req.GsessID); err != nil {
		return nil, err
	}
	profile, err := s.loadPlayer(ctx, req.UID, req.ProfileID)
	if err != nil {
		return nil, err
	}
	session, err := s.getQuery.Handle(ctx, query.GetGameSessionQuery{
		GsessID:   req.GsessID,
		UID:       profile.UID(),
		ProfileID: profile.ProfileId(),
	})
	if err != nil {
		return nil, err
	}
	return s.render(ctx, session), nil
}

// loadPlayer resolves the acting child: the profile must belong to the
// session's user and must not be a guest's. The guest check is needed on top
// of authMiddleware so that a guest whose session was made secure — by
// /auth/resume-session before 2026-11-24, for up to 14 days — cannot play.
func (s *Service) loadPlayer(ctx context.Context, uid *int64, profileID int64) (*profileDomain.Profile, error) {
	if uid == nil || *uid == 0 {
		return nil, errs.NewError(ctx, status.UNAUTHORIZED, nil, ErrUIDNotFoundFromSession)
	}
	p, err := s.profileRepo.FindByProfileId(ctx, profileID)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if p == nil {
		return nil, errs.NewError(ctx, domain.GameProfileNotFound, nil, ErrProfileNotFound)
	}
	if p.UID() != *uid {
		return nil, errs.NewError(ctx, domain.GameProfileNotOwned, nil, ErrProfileNotOwned)
	}
	if enum.IdentityCodeType(utils.DerefString(p.IdentityCode())).IsGuest() {
		return nil, errs.NewError(ctx, domain.GameGuestNotAllowed, nil, ErrGuestNotAllowed)
	}
	return p, nil
}

// resolveGrade is the stated grade, else the class the profile attends (the
// band number is read from the grade's label), else kindergarten.
func (s *Service) resolveGrade(ctx context.Context, stated *int, p *profileDomain.Profile) (int, error) {
	if stated != nil {
		return *stated, nil
	}
	if p.GradeId() == nil {
		return enum.ExamGradeMin, nil
	}
	grades, err := s.gradeRepo.ListGradesByIds(ctx, []int64{*p.GradeId()})
	if err != nil {
		return 0, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if len(grades) > 0 {
		if n, ok := input.ResolveGradeNumber(grades[0].Label()); ok {
			return n, nil
		}
		logger.From(ctx).Warnf("game.grade_label_unresolved label=%q profile=%d", grades[0].Label(), p.ProfileId())
	}
	return enum.ExamGradeMin, nil
}

// drawQuestions draws n questions from the bank at random, easiest first,
// shuffles each one's options (a true/false pair keeps its order) and returns
// the set as it is stored. Fewer than n is refused: never a shorter round.
func (s *Service) drawQuestions(ctx context.Context, grade int, level *int, n int) (string, error) {
	filter := questionpool.PickFilter{Grade: grade, Tags: gameBankTags}
	if level != nil {
		filter.Levels = []int{*level}
	}
	picked, err := s.questionPoolRepo.PickForExam(ctx, filter, n)
	if err != nil {
		return "", errs.NewError(ctx, status.FAIL, nil, err)
	}
	if len(picked) < n {
		logger.From(ctx).Warnf("game.bank.insufficient grade=%d level=%s needed=%d found=%d", grade, levelLabel(level), n, len(picked))
		return "", errs.NewError(ctx, domain.GameQuestionBankInsufficient,
			map[string]any{"grade": grade, "level": levelLabel(level), "needed": n, "found": len(picked)},
			fmt.Errorf("game: the bank holds %d approved %v questions at grade %d level %s, the round needs %d",
				len(picked), gameBankTags, grade, levelLabel(level), n))
	}
	slices.SortStableFunc(picked, func(a, b *questionpool.QuestionPool) int {
		return cmp.Compare(dto.BankDifficulty(a), dto.BankDifficulty(b))
	})

	questions := make([]dto.Question, 0, n)
	for i, p := range picked {
		q, err := dto.QuestionFromBank(p, i+1)
		if err != nil {
			return "", errs.NewError(ctx, domain.GameGenerationFailed, nil, err)
		}
		if q.EffectiveKind() != dto.KindTrueFalse {
			rand.Shuffle(len(q.Answers), func(a, b int) { q.Answers[a], q.Answers[b] = q.Answers[b], q.Answers[a] })
		}
		questions = append(questions, q)
	}
	wire := dto.FromSharedQuestions(questions)
	if err := dto.CheckSet(wire); err != nil {
		return "", errs.NewError(ctx, domain.GameGenerationFailed, nil, err)
	}
	raw, err := json.Marshal(wire)
	if err != nil {
		return "", errs.NewError(ctx, domain.GameGenerationFailed, nil, err)
	}
	return string(raw), nil
}

func (s *Service) render(ctx context.Context, session *domain.GameSession) *dto.GameSessionRes {
	res := dto.GameSessionToRes(session, requestLanguage(ctx))
	s.signImages(ctx, res)
	return res
}

// requestLanguage is the short code of metadata.language, "" when the client
// sent none: questions are then served in their own language.
func requestLanguage(ctx context.Context) enum.LanguageType {
	return enum.LanguageType(metadata.GetLanguage(ctx)).Base()
}

func levelLabel(level *int) string {
	if level == nil {
		return "any"
	}
	return fmt.Sprint(*level)
}
