package questionpool

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"unicode/utf8"

	dto "math-ai.com/math-ai/internal/application/dto/questionpool"
	domain "math-ai.com/math-ai/internal/domain/questionpool"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// Limits of one /question-pools/create call and of the stored text.
const (
	MaxQuestionPoolsPerCreate = 100
	MaxTextRunes              = 2000
	MaxOptionTextRunes        = 500
	MaxExplanationRunes       = 5000
	MaxTopicRunes             = 255
	MaxTags                   = 20
	MaxTagRunes               = 64 // what ix_question_tags (CHAR(64) ARRAY) accepts
	MaxImageKeyLen            = 1000
	MaxListSize               = 100 // one page; each question may carry 11 presigned URLs

	// ImageKeyPrefix is the S3 folder question images live under.
	ImageKeyPrefix = "question-images/"
)

// issue is one broken rule of one question, at a path like
// question_pools[3].options[1].id.
type issue struct {
	code status.StatusCode
	msg  string
}

// ValidateCreateQuestions checks every question and turns the request into
// bank rows ready to insert (no id yet): text trimmed, blank topic dropped,
// tags trimmed and deduplicated, server-owned fields set (source ADMIN,
// the source's initial status — APPROVED —, verified_count 0) and
// question_hash computed.
//
// It is all or nothing: every question is checked, and any broken rule
// refuses the whole request with ONE error — the code of the first issue,
// and every issue (by path) in debug, so the admin can fix all at once.
func ValidateCreateQuestionPools(ctx context.Context, req *dto.CreateQuestionPoolsReq) ([]*domain.QuestionPool, error) {
	if len(req.QuestionPools) == 0 {
		return nil, errs.NewError(ctx, status.QUESTION_POOL_EMPTY_LIST, nil, errors.New("question_pools is required"))
	}
	if len(req.QuestionPools) > MaxQuestionPoolsPerCreate {
		return nil, errs.NewError(ctx, status.QUESTION_POOL_TOO_MANY, nil,
			fmt.Errorf("at most %d questions per request, got %d", MaxQuestionPoolsPerCreate, len(req.QuestionPools)))
	}

	checked, found := checkItems(req.QuestionPools, func(i int) string { return fmt.Sprintf("question_pools[%d]", i) })
	issues := make([]issue, 0, len(found))
	for _, f := range found {
		issues = append(issues, f.issue)
	}
	out := make([]*domain.QuestionPool, 0, len(checked))
	firstByHash := map[string]int{}
	for _, c := range checked {
		if first, ok := firstByHash[c.q.QuestionHash()]; ok {
			issues = append(issues, issue{status.QUESTION_POOL_DUPLICATE, fmt.Sprintf("question_pools[%d] repeats question_pools[%d]", c.index, first)})
			continue
		}
		firstByHash[c.q.QuestionHash()] = c.index
		out = append(out, c.q)
	}
	if len(issues) > 0 {
		return nil, issuesError(ctx, issues)
	}
	return out, nil
}

// checkedItem is an item that passed every rule, with its place in the input.
type checkedItem struct {
	index int
	q     *domain.QuestionPool
}

// itemIssue is a broken rule of the item at index; msg already carries the
// item's path.
type itemIssue struct {
	index int
	issue
}

// checkItems runs every create rule over items, in order: the ones that
// pass, and every issue of the others with paths pathOf(i) + "." + field.
// Duplicates are the caller's call (create refuses them, import skips them).
func checkItems(items []dto.CreateQuestionPoolItem, pathOf func(int) string) ([]checkedItem, []itemIssue) {
	var (
		checked = make([]checkedItem, 0, len(items))
		issues  []itemIssue
	)
	for i, item := range items {
		q, found := buildQuestionPool(item)
		for _, is := range found {
			issues = append(issues, itemIssue{i, issue{is.code, pathOf(i) + "." + is.msg}})
		}
		if len(found) == 0 {
			checked = append(checked, checkedItem{i, q})
		}
	}
	return checked, issues
}

// issuesError is ONE refusal for every issue: the code of the first, every
// message (by path) in debug.
func issuesError(ctx context.Context, issues []issue) error {
	msgs := make([]string, 0, len(issues))
	for _, is := range issues {
		msgs = append(msgs, is.msg)
	}
	return errs.NewError(ctx, issues[0].code, nil, errors.New(strings.Join(msgs, "; ")))
}

// ValidateUpdateQuestion checks what needs no stored row.
func ValidateUpdateQuestionPool(ctx context.Context, req *dto.UpdateQuestionPoolReq) error {
	if req.QuestionPoolID <= 0 {
		return errs.NewError(ctx, status.QUESTION_POOL_MISSING_ID, nil, errors.New("question_pool_id is required"))
	}
	if req.QuestionKind == nil && req.QuestionText == nil && !req.QuestionImageKey.Set && !req.Options.Set &&
		req.CorrectAnswer == nil && !req.Explanation.Set && req.QuestionGrade == nil && !req.QuestionTopic.Set &&
		!req.QuestionLevel.Set && !req.QuestionSubLevel.Set && req.QuestionTags == nil {
		return errs.NewError(ctx, status.QUESTION_POOL_NOTHING_TO_UPDATE, nil, errors.New("no field to change"))
	}
	return nil
}

// ValidateMarkQuestionStatus needs a question_pool_id and a known review state.
func ValidateMarkQuestionPoolStatus(ctx context.Context, req *dto.MarkQuestionPoolStatusReq) error {
	if req.QuestionPoolID <= 0 {
		return errs.NewError(ctx, status.QUESTION_POOL_MISSING_ID, nil, errors.New("question_pool_id is required"))
	}
	if !slices.Contains(domain.Statuses, req.QuestionStatus) {
		return errs.NewError(ctx, status.QUESTION_POOL_INVALID_STATUS, nil,
			fmt.Errorf("question_status %q is not one of %v", req.QuestionStatus, domain.Statuses))
	}
	return nil
}

// MergeUpdate lays the request over the stored question and checks the
// result with every create rule (so a kind change must bring options and a
// key that fit it). Source, status and verified_count stay as stored; the
// hash is recomputed. Issue paths start with "question_pool.".
func MergeUpdate(ctx context.Context, stored *domain.QuestionPool, req *dto.UpdateQuestionPoolReq) (*domain.QuestionPool, error) {
	item := itemOf(stored)
	if req.QuestionKind != nil {
		item.QuestionKind = *req.QuestionKind
	}
	if req.QuestionText != nil {
		item.QuestionText = *req.QuestionText
	}
	if req.QuestionImageKey.Set {
		item.QuestionImageKey = req.QuestionImageKey.Value
	}
	if req.Options.Set {
		item.Options = nil
		if req.Options.Value != nil {
			item.Options = *req.Options.Value
		}
	}
	if req.CorrectAnswer != nil {
		item.CorrectAnswer = req.CorrectAnswer
	}
	if req.Explanation.Set {
		item.Explanation = req.Explanation.Value
	}
	if req.QuestionGrade != nil {
		item.QuestionGrade = req.QuestionGrade
	}
	if req.QuestionTopic.Set {
		item.QuestionTopic = req.QuestionTopic.Value
	}
	if req.QuestionLevel.Set {
		item.QuestionLevel = req.QuestionLevel.Value
	}
	if req.QuestionSubLevel.Set {
		item.QuestionSubLevel = req.QuestionSubLevel.Value
	}
	if req.QuestionTags != nil {
		item.QuestionTags = *req.QuestionTags
	}

	q, found := buildQuestionPool(item)
	if len(found) > 0 {
		for i := range found {
			found[i].msg = "question_pool." + found[i].msg
		}
		return nil, issuesError(ctx, found)
	}
	q.SetQuestionPoolId(stored.QuestionPoolId())
	q.SetQuestionSource(stored.QuestionSource())
	q.SetQuestionStatus(stored.QuestionStatus())
	q.SetVerifiedCount(stored.VerifiedCount())
	return q, nil
}

// itemOf is the stored question in request form, the base an update is
// laid over.
func itemOf(q *domain.QuestionPool) dto.CreateQuestionPoolItem {
	var options []dto.OptionDto
	for _, o := range q.Options() {
		options = append(options, dto.OptionDto{Id: o.Id, Text: o.Text, ImageKey: o.ImageKey})
	}
	a := q.CorrectAnswer()
	grade := q.QuestionGrade()
	return dto.CreateQuestionPoolItem{
		QuestionKind:     string(q.QuestionKind()),
		QuestionText:     q.QuestionText(),
		QuestionImageKey: q.QuestionImageKey(),
		Options:          options,
		CorrectAnswer:    &dto.CorrectAnswerDto{OptionIds: a.OptionIds, Value: a.Value, Values: a.Values},
		Explanation:      q.Explanation(),
		QuestionGrade:    &grade,
		QuestionTopic:    q.QuestionTopic(),
		QuestionLevel:    q.QuestionLevel(),
		QuestionSubLevel: q.QuestionSubLevel(),
		QuestionTags:     q.QuestionTags(),
	}
}

// buildQuestion checks one item and maps it; issues carry paths relative to
// the item.
func buildQuestionPool(item dto.CreateQuestionPoolItem) (*domain.QuestionPool, []issue) {
	var issues []issue
	add := func(code status.StatusCode, format string, args ...any) {
		issues = append(issues, issue{code, fmt.Sprintf(format, args...)})
	}

	kind := domain.Kind(strings.TrimSpace(item.QuestionKind))
	text := strings.TrimSpace(item.QuestionText)
	switch {
	case text == "":
		add(status.QUESTION_POOL_INVALID, "question_text: required")
	case utf8.RuneCountInString(text) > MaxTextRunes:
		add(status.QUESTION_POOL_INVALID, "question_text: longer than %d characters", MaxTextRunes)
	}
	if msg := checkImageKey(item.QuestionImageKey); msg != "" {
		add(status.QUESTION_POOL_INVALID_IMAGE_KEY, "question_image_key: %s", msg)
	}

	// An empty list is "no options": a TRUE_FALSE sent with "options": []
	// is not refused for it.
	var options []domain.Option
	if len(item.Options) > 0 {
		options = make([]domain.Option, 0, len(item.Options))
		for j, o := range item.Options {
			optText := strings.TrimSpace(o.Text)
			if utf8.RuneCountInString(optText) > MaxOptionTextRunes {
				add(status.QUESTION_POOL_INVALID_OPTIONS, "options[%d].text: longer than %d characters", j, MaxOptionTextRunes)
			}
			if msg := checkImageKey(o.ImageKey); msg != "" {
				add(status.QUESTION_POOL_INVALID_IMAGE_KEY, "options[%d].image_key: %s", j, msg)
			}
			options = append(options, domain.Option{Id: strings.TrimSpace(o.Id), Text: optText, ImageKey: o.ImageKey})
		}
	}

	var answer domain.CorrectAnswer
	if item.CorrectAnswer == nil {
		add(status.QUESTION_POOL_INVALID_ANSWER, "correct_answer: required")
	} else {
		answer = domain.CorrectAnswer{OptionIds: item.CorrectAnswer.OptionIds, Value: item.CorrectAnswer.Value, Values: item.CorrectAnswer.Values}
		if kind == domain.KindNumericInput {
			// Stored canonical ("0,50" → "0.5"); a value that does not parse
			// is reported by CheckShape below.
			if values, err := domain.NormalizeNumbers(answer.Values); err == nil {
				answer.Values = values
			}
		}
		if err := domain.CheckShape(kind, options, answer); err != nil {
			add(shapeCode(err), "%s", err.Error())
		}
	}

	switch g := item.QuestionGrade; {
	case g == nil:
		add(status.QUESTION_POOL_INVALID, "question_grade: required")
	case *g < domain.GradeMin || *g > domain.GradeMax:
		add(status.QUESTION_POOL_INVALID, "question_grade: %d is outside %d..%d", *g, domain.GradeMin, domain.GradeMax)
	}
	if l := item.QuestionLevel; l != nil && (*l < domain.LevelMin || *l > domain.LevelMax) {
		add(status.QUESTION_POOL_INVALID, "question_level: %d is outside %d..%d", *l, domain.LevelMin, domain.LevelMax)
	}
	if l := item.QuestionSubLevel; l != nil && (*l < domain.LevelMin || *l > domain.LevelMax) {
		add(status.QUESTION_POOL_INVALID, "question_sub_level: %d is outside %d..%d", *l, domain.LevelMin, domain.LevelMax)
	}
	explanation := trimmedOrNil(item.Explanation)
	if explanation != nil && utf8.RuneCountInString(*explanation) > MaxExplanationRunes {
		add(status.QUESTION_POOL_INVALID, "explanation: longer than %d characters", MaxExplanationRunes)
	}
	topic := trimmedOrNil(item.QuestionTopic)
	if topic != nil && utf8.RuneCountInString(*topic) > MaxTopicRunes {
		add(status.QUESTION_POOL_INVALID, "question_topic: longer than %d characters", MaxTopicRunes)
	}
	tags, msg := cleanTags(item.QuestionTags)
	if msg != "" {
		add(status.QUESTION_POOL_INVALID, "question_tags: %s", msg)
	}
	if len(issues) > 0 {
		return nil, issues
	}

	q := domain.NewQuestionPool()
	q.SetQuestionKind(kind)
	q.SetQuestionText(text)
	q.SetQuestionImageKey(item.QuestionImageKey)
	q.SetOptions(options)
	q.SetCorrectAnswer(answer)
	q.SetExplanation(explanation)
	q.SetQuestionGrade(*item.QuestionGrade)
	q.SetQuestionTopic(topic)
	q.SetQuestionLevel(item.QuestionLevel)
	q.SetQuestionSubLevel(item.QuestionSubLevel)
	q.SetQuestionTags(tags)
	q.SetQuestionSource(domain.SourceAdmin)
	initial := domain.SourceAdmin.InitialStatus()
	q.SetQuestionStatus(&initial)
	q.SetQuestionHash(domain.Hash(q))
	return q, nil
}

func shapeCode(err error) status.StatusCode {
	switch {
	case errors.Is(err, domain.ErrInvalidKind):
		return status.QUESTION_POOL_INVALID_KIND
	case errors.Is(err, domain.ErrInvalidOptions):
		return status.QUESTION_POOL_INVALID_OPTIONS
	default:
		return status.QUESTION_POOL_INVALID_ANSWER
	}
}

// checkImageKey returns why key is not a question image key, or "". Only
// the shape is checked: a key under ImageKeyPrefix, never a URL or a path
// that climbs out of it. Whether the object exists is not.
func checkImageKey(key *string) string {
	if key == nil {
		return ""
	}
	k := *key
	switch {
	case !strings.HasPrefix(k, ImageKeyPrefix) || len(k) == len(ImageKeyPrefix):
		return fmt.Sprintf("must be a key under %s", ImageKeyPrefix)
	case len(k) > MaxImageKeyLen:
		return fmt.Sprintf("longer than %d characters", MaxImageKeyLen)
	case strings.Contains(k, "://") || strings.Contains(k, "..") || strings.ContainsAny(k, " \t\r\n\\?#"):
		return "must be an S3 key, not a URL or path"
	}
	return ""
}

// cleanTags trims, drops blanks and repeats; it returns why the list is
// refused, or "".
func cleanTags(raw []string) ([]string, string) {
	tags := make([]string, 0, len(raw))
	for _, t := range raw {
		t = strings.TrimSpace(t)
		if t == "" || slices.Contains(tags, t) {
			continue
		}
		if utf8.RuneCountInString(t) > MaxTagRunes {
			return nil, fmt.Sprintf("%q is longer than %d characters", t, MaxTagRunes)
		}
		tags = append(tags, t)
	}
	if len(tags) > MaxTags {
		return nil, fmt.Sprintf("at most %d tags, got %d", MaxTags, len(tags))
	}
	return tags, ""
}

func trimmedOrNil(s *string) *string {
	if s == nil {
		return nil
	}
	t := strings.TrimSpace(*s)
	if t == "" {
		return nil
	}
	return &t
}

// ValidateListQuestions checks the filters and turns them into the repo
// filter. A size above MaxListSize is lowered to it; page and size are
// otherwise clamped by the repository.
func ValidateListQuestionPools(ctx context.Context, req *dto.ListQuestionPoolsReq) (domain.ListQuestionPoolsFilter, error) {
	var filter domain.ListQuestionPoolsFilter
	if g := req.QuestionGrade; g != nil {
		if *g < domain.GradeMin || *g > domain.GradeMax {
			return filter, errs.NewError(ctx, status.QUESTION_POOL_INVALID_FILTER, nil,
				fmt.Errorf("question_grade: %d is outside %d..%d", *g, domain.GradeMin, domain.GradeMax))
		}
		filter.Grade = g
	}
	if l := req.QuestionLevel; l != nil {
		if *l < domain.LevelMin || *l > domain.LevelMax {
			return filter, errs.NewError(ctx, status.QUESTION_POOL_INVALID_FILTER, nil,
				fmt.Errorf("question_level: %d is outside %d..%d", *l, domain.LevelMin, domain.LevelMax))
		}
		filter.Level = l
	}
	filter.Topic = trimmedOrNil(req.QuestionTopic)
	tags, msg := cleanTags(req.QuestionTags)
	if msg != "" {
		return filter, errs.NewError(ctx, status.QUESTION_POOL_INVALID_FILTER, nil, errors.New("question_tags: "+msg))
	}
	if len(tags) > 0 {
		filter.Tags = tags
	}
	for _, raw := range req.QuestionKinds {
		k := domain.Kind(strings.TrimSpace(raw))
		if !k.IsValid() {
			return filter, errs.NewError(ctx, status.QUESTION_POOL_INVALID_KIND, nil, fmt.Errorf("question_kinds: unknown kind %q", raw))
		}
		if !slices.Contains(filter.Kinds, k) {
			filter.Kinds = append(filter.Kinds, k)
		}
	}
	for _, raw := range req.QuestionStatuses {
		s := strings.TrimSpace(raw)
		if !slices.Contains(domain.Statuses, s) {
			return filter, errs.NewError(ctx, status.QUESTION_POOL_INVALID_FILTER, nil, fmt.Errorf("question_statuses: unknown status %q", raw))
		}
		if !slices.Contains(filter.Statuses, s) {
			filter.Statuses = append(filter.Statuses, s)
		}
	}
	if req.QuestionSource != nil {
		src := domain.Source(strings.TrimSpace(*req.QuestionSource))
		if src != domain.SourceAI && src != domain.SourceAdmin {
			return filter, errs.NewError(ctx, status.QUESTION_POOL_INVALID_FILTER, nil, fmt.Errorf("question_source: unknown source %q", *req.QuestionSource))
		}
		filter.Source = &src
	}
	filter.IsVerified = req.IsVerified
	if req.Size > MaxListSize {
		req.Size = MaxListSize
	}
	return filter, nil
}
