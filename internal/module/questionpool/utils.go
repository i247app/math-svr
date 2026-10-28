package questionpool

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"path"
	"slices"
	"strings"

	dto "math-ai.com/math-ai/internal/application/dto/questionpool"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

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
