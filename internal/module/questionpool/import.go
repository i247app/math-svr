package questionpool

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path"
	"slices"
	"strings"

	command "math-ai.com/math-ai/internal/application/command/questionpool"
	dto "math-ai.com/math-ai/internal/application/dto/questionpool"
	domain "math-ai.com/math-ai/internal/domain/questionpool"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	sctx "math-ai.com/math-ai/internal/shared/context"
)

// Limits of one imported file. 2,000 questions are about 3–4 MB; the file
// cap must stay under MetadataMiddleware's 16 MiB, beyond which the body's
// metadata (and so the session token) is not read.
const (
	MaxImportFileBytes = 10 << 20
	MaxImportQuestions = 2000
	MaxReportErrors    = 200
)

// Why a valid question of the file is not imported.
const (
	DuplicateInFile = "IN_FILE"
	DuplicateInPool = "IN_POOL"
)

var (
	ErrImportFileMissing  = errors.New("file part is required")
	ErrImportFileTooLarge = errors.New("request body is over the import file size limit")
)

// ImportFile is the multipart "file" part; a nil Reader means the request
// had none. Size is the part's size as the client declared it; the read is
// capped at MaxImportFileBytes as well.
type ImportFile struct {
	Reader   io.Reader
	Filename string
	Size     int64
}

// importAnalysis is a file checked against the rules and the pool: the
// report so far, and the questions to store with their file index.
type importAnalysis struct {
	report   *dto.ImportQuestionPoolsReport
	toImport []checkedItem
}

// importFormError maps a body that is not a usable multipart form. The
// admin gate still comes first, so an unauthorised caller learns nothing.
func (s *Service) importFormError(ctx context.Context, err error) error {
	if gateErr := requireAdminGate(ctx); gateErr != nil {
		return gateErr
	}
	if errors.Is(err, ErrImportFileTooLarge) {
		return errs.NewError(ctx, status.QUESTION_POOL_IMPORT_FILE_TOO_LARGE, nil, err)
	}
	return errs.NewError(ctx, status.QUESTION_POOL_IMPORT_FILE_MISSING, nil, fmt.Errorf("body is not multipart/form-data with a file part: %w", err))
}

// ValidateFile checks a question file the way ImportFile would and writes
// nothing. A file with invalid questions is a SUCCESS response: the report
// is the answer.
func (s *Service) ValidateFile(ctx context.Context, file ImportFile) (*dto.ImportQuestionPoolsRes, error) {
	if err := requireAdminGate(ctx); err != nil {
		return nil, err
	}
	a, err := s.analyzeImport(ctx, file)
	if err != nil {
		return nil, err
	}
	logSummary(ctx, "question_pool.import.validated", a.report)

	res := &dto.ImportQuestionPoolsRes{
		Report: a.report,
	}

	return res, nil
}

// ImportFile stores every valid, new question of the file in one
// transaction. Any invalid question refuses the whole file; duplicates are
// skipped and logged.
func (s *Service) ImportFile(ctx context.Context, file ImportFile) (*dto.ImportQuestionPoolsRes, error) {
	if err := requireAdminGate(ctx); err != nil {
		return nil, err
	}
	a, err := s.analyzeImport(ctx, file)
	if err != nil {
		return nil, err
	}
	r := a.report
	if r.InvalidCount > 0 {
		msgs := make([]string, 0, len(r.Errors))
		for _, e := range r.Errors {
			msgs = append(msgs, e.Message)
		}
		return nil, errs.NewError(ctx, status.QUESTION_POOL_IMPORT_INVALID, nil,
			fmt.Errorf("%d of %d questions are invalid: %s", r.InvalidCount, r.Total, strings.Join(msgs, "; ")))
	}

	questions := make([]*domain.QuestionPool, 0, len(a.toImport))
	for _, c := range a.toImport {
		questions = append(questions, c.q)
	}
	if len(questions) > 0 {
		res, err := s.importCmd.Handle(ctx, command.ImportQuestionPoolsCommand{QuestionPools: questions, ActorUID: adminActor(ctx)})
		if err != nil {
			return nil, err
		}
		for _, i := range res.Skipped {
			r.Duplicates = append(r.Duplicates, dto.ImportQuestionPoolDuplicate{Index: a.toImport[i].index, Reason: DuplicateInPool})
			r.DuplicateInPoolCount++
		}
		for _, i := range res.Imported {
			r.QuestionPoolIds = append(r.QuestionPoolIds, questions[i].QuestionPoolId())
		}
		sortDuplicates(r.Duplicates)
		r.ImportedCount = len(res.Imported)
		r.ImportableCount = r.ImportedCount
	}

	log := logger.From(ctx)
	for _, d := range r.Duplicates {
		dupOf := -1
		if d.DuplicateOf != nil {
			dupOf = *d.DuplicateOf
		}
		log.Infof("question_pool.import.duplicate_skipped file=%s index=%d reason=%s duplicate_of=%d", r.FileName, d.Index, d.Reason, dupOf)
	}
	logSummary(ctx, "question_pool.imported", r)

	res := &dto.ImportQuestionPoolsRes{
		Report: r,
	}

	return res, nil
}

func logSummary(ctx context.Context, event string, r *dto.ImportQuestionPoolsReport) {
	logger.From(ctx).Infof("%s file=%s total=%d valid=%d invalid=%d dup_in_file=%d dup_in_pool=%d imported=%d admin_uid=%v via=%s",
		event, r.FileName, r.Total, r.ValidCount, r.InvalidCount, r.DuplicateInFileCount, r.DuplicateInPoolCount,
		r.ImportedCount, sctx.UID(ctx), sctx.GetAdminVia(ctx))
}

// analyzeImport reads the file, checks every question with the create
// rules, then sets duplicates aside: a repeat inside the file keeps its
// first occurrence, and content a live row holds is skipped. File-level
// problems (missing, too large, not a JSON array) are errors; question
// problems go in the report.
func (s *Service) analyzeImport(ctx context.Context, file ImportFile) (*importAnalysis, error) {
	raw, err := readImportFile(file)
	if err != nil {
		return nil, err.toMathError(ctx)
	}
	items, positions, total, found, ferr := parseImportItems(raw)
	if ferr != nil {
		return nil, ferr.toMathError(ctx)
	}

	r := &dto.ImportQuestionPoolsReport{
		FileName:        path.Base(file.Filename),
		Total:           total,
		Errors:          []dto.ImportQuestionPoolError{},
		Duplicates:      []dto.ImportQuestionPoolDuplicate{},
		QuestionPoolIds: []int64{},
	}
	switch {
	case r.Total == 0:
		return nil, errs.NewError(ctx, status.QUESTION_POOL_EMPTY_LIST, nil, errors.New("the file holds no question"))
	case r.Total > MaxImportQuestions:
		return nil, errs.NewError(ctx, status.QUESTION_POOL_TOO_MANY, nil,
			fmt.Errorf("at most %d questions per file, got %d", MaxImportQuestions, r.Total))
	}

	checked, ruleIssues := checkItems(items, func(i int) string { return fmt.Sprintf("[%d]", positions[i]) })
	for _, is := range ruleIssues {
		is.index = positions[is.index]
		found = append(found, is)
	}
	invalid := map[int]bool{}
	for _, is := range found {
		invalid[is.index] = true
	}
	sortIssues(found)
	for _, is := range found {
		if len(r.Errors) == MaxReportErrors {
			r.ErrorsTruncated = true
			break
		}
		r.Errors = append(r.Errors, dto.ImportQuestionPoolError{Index: is.index, Code: int(is.code), Message: is.msg})
	}
	r.InvalidCount = len(invalid)
	r.ValidCount = len(checked)
	for i := range checked {
		checked[i].index = positions[checked[i].index]
	}

	unique := make([]checkedItem, 0, len(checked))
	firstByHash := map[string]int{}
	for _, c := range checked {
		if first, ok := firstByHash[c.q.QuestionHash()]; ok {
			r.Duplicates = append(r.Duplicates, dto.ImportQuestionPoolDuplicate{Index: c.index, Reason: DuplicateInFile, DuplicateOf: &first})
			r.DuplicateInFileCount++
			continue
		}
		firstByHash[c.q.QuestionHash()] = c.index
		unique = append(unique, c)
	}

	hashes := make([]string, 0, len(unique))
	for _, c := range unique {
		hashes = append(hashes, c.q.QuestionHash())
	}
	held, err2 := s.repo.FindLiveHashes(ctx, hashes)
	if err2 != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err2)
	}
	isHeld := make(map[string]bool, len(held))
	for _, h := range held {
		isHeld[h] = true
	}
	toImport := make([]checkedItem, 0, len(unique))
	for _, c := range unique {
		if isHeld[c.q.QuestionHash()] {
			r.Duplicates = append(r.Duplicates, dto.ImportQuestionPoolDuplicate{Index: c.index, Reason: DuplicateInPool})
			r.DuplicateInPoolCount++
			continue
		}
		toImport = append(toImport, c)
	}
	sortDuplicates(r.Duplicates)
	r.ImportableCount = len(toImport)

	res := &importAnalysis{
		report:   r,
		toImport: toImport,
	}

	return res, nil
}

// fileError is a problem with the file as a whole, mapped to its code.
type fileError struct {
	code status.StatusCode
	err  error
}

func (e *fileError) toMathError(ctx context.Context) error {
	return errs.NewError(ctx, e.code, nil, e.err)
}

// readImportFile returns the file's bytes: present, named .json, at most
// MaxImportFileBytes.
func readImportFile(file ImportFile) ([]byte, *fileError) {
	if file.Reader == nil {
		return nil, &fileError{status.QUESTION_POOL_IMPORT_FILE_MISSING, ErrImportFileMissing}
	}
	if !strings.EqualFold(path.Ext(file.Filename), ".json") {
		return nil, &fileError{status.QUESTION_POOL_IMPORT_FILE_INVALID, fmt.Errorf("file %q is not a .json file", path.Base(file.Filename))}
	}
	if file.Size > MaxImportFileBytes {
		return nil, &fileError{status.QUESTION_POOL_IMPORT_FILE_TOO_LARGE, fmt.Errorf("file is %d bytes, at most %d", file.Size, MaxImportFileBytes)}
	}
	data, err := io.ReadAll(io.LimitReader(file.Reader, MaxImportFileBytes+1))
	if err != nil {
		return nil, &fileError{status.QUESTION_POOL_IMPORT_FILE_INVALID, fmt.Errorf("read file: %w", err)}
	}
	if len(data) > MaxImportFileBytes {
		return nil, &fileError{status.QUESTION_POOL_IMPORT_FILE_TOO_LARGE, fmt.Errorf("file is over %d bytes", MaxImportFileBytes)}
	}
	return data, nil
}

// parseImportItems decodes the file: a JSON array (else a file error),
// each element one question. An element that does not decode — not an
// object, a field of the wrong type, a field the request has no place for
// (question_source, id…) — is an issue of that index, not a file error, so
// one bad element does not hide the others. items[i] is the element at
// file index positions[i]; total counts every element, decoded or not.
func parseImportItems(raw []byte) (items []dto.CreateQuestionPoolItem, positions []int, total int, issues []itemIssue, ferr *fileError) {
	var elems []json.RawMessage
	if err := json.Unmarshal(raw, &elems); err != nil {
		return nil, nil, 0, nil, &fileError{status.QUESTION_POOL_IMPORT_FILE_INVALID,
			fmt.Errorf("the file must be a JSON array of questions: %w", err)}
	}
	items = make([]dto.CreateQuestionPoolItem, 0, len(elems))
	positions = make([]int, 0, len(elems))
	for i, e := range elems {
		var item dto.CreateQuestionPoolItem
		dec := json.NewDecoder(bytes.NewReader(e))
		dec.DisallowUnknownFields()
		if err := dec.Decode(&item); err != nil {
			issues = append(issues, itemIssue{i, issue{status.QUESTION_POOL_INVALID, fmt.Sprintf("[%d]: %v", i, err)}})
			continue
		}
		items = append(items, item)
		positions = append(positions, i)
	}
	return items, positions, len(elems), issues, nil
}

// sortIssues orders issues by file index, keeping each question's own order.
func sortIssues(issues []itemIssue) {
	slices.SortStableFunc(issues, func(a, b itemIssue) int { return a.index - b.index })
}

func sortDuplicates(d []dto.ImportQuestionPoolDuplicate) {
	slices.SortStableFunc(d, func(a, b dto.ImportQuestionPoolDuplicate) int { return a.Index - b.Index })
}
