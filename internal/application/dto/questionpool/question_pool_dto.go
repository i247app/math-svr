// Package questionpool holds the request / response shapes of the question
// bank's admin routes (/question-pools/*). The JSON shapes of options and
// correct_answer are the ones stored in ma_question_pools (migration 036).
package questionpool

import (
	"encoding/json"

	domain "math-ai.com/math-ai/internal/domain/questionpool"
	"math-ai.com/math-ai/internal/shared/pagination"
)

// OptionDto is one choice. ImageUrl is response-only: the presigned URL of
// ImageKey, filled by the module (nil when there is no image or presigning
// failed); a request's image_url is ignored.
type OptionDto struct {
	Id       string  `json:"id"`
	Text     string  `json:"text"`
	ImageKey *string `json:"image_key,omitempty"`
	ImageUrl *string `json:"image_url,omitempty"`
}

// CorrectAnswerDto is the key: option_ids for choice kinds, value for
// TRUE_FALSE, values (numbers as strings: "5", "0,5", "1/2") for
// NUMERIC_INPUT. Exactly one is set; which one is checked by the validator.
type CorrectAnswerDto struct {
	OptionIds []string `json:"option_ids,omitempty"`
	Value     *bool    `json:"value,omitempty"`
	Values    []string `json:"values,omitempty"`
}

// CreateQuestionItem is one question an admin sends. Pointer fields tell
// "absent" from "zero" for the validator. Source, status, verified_count
// and ids are set by the server; there is no field for them.
type CreateQuestionPoolItem struct {
	QuestionKind     string            `json:"question_kind"`
	QuestionText     string            `json:"question_text"`
	QuestionImageKey *string           `json:"question_image_key,omitempty"`
	Options          []OptionDto       `json:"options,omitempty"`
	CorrectAnswer    *CorrectAnswerDto `json:"correct_answer"`
	Explanation      *string           `json:"explanation,omitempty"`
	QuestionGrade    *int              `json:"question_grade"`
	QuestionTopic    *string           `json:"question_topic,omitempty"`
	QuestionLevel    *int              `json:"question_level,omitempty"`
	QuestionTags     []string          `json:"question_tags,omitempty"`
}

// CreateQuestionsReq creates every question or none.
type CreateQuestionPoolsReq struct {
	QuestionPools []CreateQuestionPoolItem `json:"question_pools"`
}

// CreateQuestionsRes lists the created questions in request order.
type CreateQuestionPoolsRes struct {
	QuestionPools []*QuestionPoolResponse `json:"question_pools"`
}

// ListQuestionsReq pages the bank, newest first (OFFSET: page, size).
// Every filter is optional and set filters are ANDed. QuestionTags matches
// a question holding at least one of them. IsVerified nil = both, true =
// verified_count > 0, false = 0.
type ListQuestionPoolsReq struct {
	QuestionGrade    *int     `json:"question_grade,omitempty"`
	QuestionTopic    *string  `json:"question_topic,omitempty"`
	QuestionTags     []string `json:"question_tags,omitempty"`
	QuestionKinds    []string `json:"question_kinds,omitempty"`
	IsVerified       *bool    `json:"is_verified,omitempty"`
	QuestionStatuses []string `json:"question_statuses,omitempty"`
	QuestionSource   *string  `json:"question_source,omitempty"`
	Page             int64    `json:"page,omitempty"`
	Size             int64    `json:"size,omitempty"`
}

type ListQuestionPoolsRes struct {
	QuestionPools []*QuestionPoolResponse `json:"question_pools"`
	Pagination    *pagination.Pagination  `json:"pagination"`
}

// QuestionResponse is the admin view of one question: the key and the
// explanation are included (these routes are admin-only). Image URLs are
// filled by the module after mapping.
type QuestionPoolResponse struct {
	QuestionPoolID   int64            `json:"question_pool_id"`
	QuestionKind     string           `json:"question_kind"`
	QuestionText     string           `json:"question_text"`
	QuestionImageKey *string          `json:"question_image_key"`
	QuestionImageUrl *string          `json:"question_image_url"`
	Options          []OptionDto      `json:"options"`
	CorrectAnswer    CorrectAnswerDto `json:"correct_answer"`
	Explanation      *string          `json:"explanation"`
	QuestionGrade    int              `json:"question_grade"`
	QuestionTopic    *string          `json:"question_topic"`
	QuestionLevel    *int             `json:"question_level"`
	QuestionTags     []string         `json:"question_tags"`
	QuestionSource   string           `json:"question_source"`
	QuestionStatus   *string          `json:"question_status"`
	VerifiedCount    int              `json:"verified_count"`
	Note             *string          `json:"note,omitempty"`
	CreateID         *int64           `json:"create_id"`
	CreateDt         string           `json:"create_dt"`
	ModifyID         *int64           `json:"modify_id"`
	ModifyDt         string           `json:"modify_dt"`
}

func DomainToResponse(q *domain.QuestionPool) *QuestionPoolResponse {
	if q == nil {
		return nil
	}
	var options []OptionDto
	if q.Options() != nil {
		options = make([]OptionDto, 0, len(q.Options()))
		for _, o := range q.Options() {
			options = append(options, OptionDto{Id: o.Id, Text: o.Text, ImageKey: o.ImageKey})
		}
	}
	tags := q.QuestionTags()
	if tags == nil {
		tags = []string{}
	}
	a := q.CorrectAnswer()
	return &QuestionPoolResponse{
		QuestionPoolID:   q.QuestionPoolId(),
		QuestionKind:     string(q.QuestionKind()),
		QuestionText:     q.QuestionText(),
		QuestionImageKey: q.QuestionImageKey(),
		Options:          options,
		CorrectAnswer:    CorrectAnswerDto{OptionIds: a.OptionIds, Value: a.Value, Values: a.Values},
		Explanation:      q.Explanation(),
		QuestionGrade:    q.QuestionGrade(),
		QuestionTopic:    q.QuestionTopic(),
		QuestionLevel:    q.QuestionLevel(),
		QuestionTags:     tags,
		QuestionSource:   string(q.QuestionSource()),
		QuestionStatus:   q.QuestionStatus(),
		VerifiedCount:    q.VerifiedCount(),
		Note:             q.Note(),
		CreateID:         q.CreateId(),
		CreateDt:         q.CreateDt().String(),
		ModifyID:         q.ModifyId(),
		ModifyDt:         q.ModifyDt().String(),
	}
}

func DomainListToResponse(qs []*domain.QuestionPool) []*QuestionPoolResponse {
	out := make([]*QuestionPoolResponse, 0, len(qs))
	for _, q := range qs {
		out = append(out, DomainToResponse(q))
	}
	return out
}

// UploadImageRes is what /question-pools/images/upload returns: the key to put
// in question_image_key / options[].image_key, and a presigned URL to
// preview it (null when presigning failed).
type UploadImageRes struct {
	ImageKey string  `json:"image_key"`
	ImageUrl *string `json:"image_url"`
}

// Opt is a request field that may be absent, null, or a value. Absent
// (Set false) keeps the stored value; null (Set true, Value nil) clears it.
// A plain pointer cannot tell the two apart.
type Opt[T any] struct {
	Set   bool
	Value *T
}

func (o *Opt[T]) UnmarshalJSON(raw []byte) error {
	o.Set = true
	if string(raw) == "null" {
		o.Value = nil
		return nil
	}
	var v T
	if err := json.Unmarshal(raw, &v); err != nil {
		return err
	}
	o.Value = &v
	return nil
}

// UpdateQuestionReq changes one question. Every field but question_pool_id is
// optional: an absent field keeps its stored value. The nullable ones (Opt)
// are cleared with null; options is cleared with null or [] (TRUE_FALSE).
// The merged question must pass every create rule. Source, status and
// verified_count are not editable here.
type UpdateQuestionPoolReq struct {
	QuestionPoolID   int64             `json:"question_pool_id"`
	QuestionKind     *string           `json:"question_kind"`
	QuestionText     *string           `json:"question_text"`
	QuestionImageKey Opt[string]       `json:"question_image_key"`
	Options          Opt[[]OptionDto]  `json:"options"`
	CorrectAnswer    *CorrectAnswerDto `json:"correct_answer"`
	Explanation      Opt[string]       `json:"explanation"`
	QuestionGrade    *int              `json:"question_grade"`
	QuestionTopic    Opt[string]       `json:"question_topic"`
	QuestionLevel    Opt[int]          `json:"question_level"`
	QuestionTags     *[]string         `json:"question_tags"`
}

// MarkQuestionStatusReq sets one question's review state.
type MarkQuestionPoolStatusReq struct {
	QuestionPoolID int64  `json:"question_pool_id"`
	QuestionStatus string `json:"question_status"`
}

// VerifyQuestionReq records one more admin check of a question.
type VerifyQuestionPoolReq struct {
	QuestionPoolID int64 `json:"question_pool_id"`
}

// DeleteQuestionReq names the question to soft- or force-delete.
type DeleteQuestionPoolReq struct {
	QuestionPoolID int64 `json:"question_pool_id"`
}

// DeleteQuestionRes is empty: the question is gone.
type DeleteQuestionPoolRes struct{}

// QuestionRes wraps one question.
type QuestionPoolRes struct {
	QuestionPool *QuestionPoolResponse `json:"question_pool"`
}
