package exam

import (
	"context"
	"time"

	"math-ai.com/math-ai/internal/adapter/storage"
	dto "math-ai.com/math-ai/internal/application/dto/exam"
	"math-ai.com/math-ai/internal/infrastructure/logger"
)

// imageURLTTL is how long a presigned question / option image URL stays
// valid — the same hour the question bank's own responses use. A sitting
// reopened later gets fresh URLs, because they are made per response.
const imageURLTTL = time.Hour

// urlSigner is the part of the storage adapter this module needs.
type urlSigner interface {
	CreatePresignedUrl(ctx context.Context, req *storage.CreatePresignedUrlRequest) (string, error)
}

// Images are stored as S3 keys inside questions_json (question bank sets
// only). A response carries the key and a presigned URL beside it; the URL
// is made here, when the response is built, and never stored. With no
// storage configured, or when signing fails, the URL is left empty (the
// request does not fail for it).

// imageURL returns a short-lived URL for key, "" when there is no key.
func (s *Service) imageURL(ctx context.Context, key string) string {
	if key == "" {
		return ""
	}
	if s.signer == nil {
		logger.From(ctx).Warnf("exam.image.no_storage key=%s", key)
		return ""
	}
	url, err := s.signer.CreatePresignedUrl(ctx, &storage.CreatePresignedUrlRequest{Key: key, Expiration: imageURLTTL})
	if err != nil {
		logger.From(ctx).Warnf("exam.image.presign_failed key=%s err=%v", key, err)
		return ""
	}
	return url
}

// signAnswers returns options with their image URLs filled. The input is
// not modified: its slice may be shared with a parsed question set.
func (s *Service) signAnswers(ctx context.Context, in []dto.ExamOption) []dto.ExamOption {
	imaged := false
	for _, a := range in {
		if a.ImageKey != "" {
			imaged = true
			break
		}
	}
	if !imaged {
		return in
	}
	out := make([]dto.ExamOption, len(in))
	for i, a := range in {
		a.ImageURL = s.imageURL(ctx, a.ImageKey)
		out[i] = a
	}
	return out
}

// signQuestions fills the image URLs of a served question set, in place.
func (s *Service) signQuestions(ctx context.Context, qs []dto.ExamQuestion) {
	for i := range qs {
		qs[i].QuestionImageURL = s.imageURL(ctx, qs[i].QuestionImageKey)
		qs[i].Options = s.signAnswers(ctx, qs[i].Options)
	}
}

func (s *Service) signExam(ctx context.Context, e *dto.ExamResponse) {
	if e != nil {
		s.signQuestions(ctx, e.Questions)
	}
}

// signDetails fills the image URLs of a review's rows, in place.
func (s *Service) signDetails(ctx context.Context, details []dto.ExamAnswerDetail) {
	for i := range details {
		details[i].QuestionImageURL = s.imageURL(ctx, details[i].QuestionImageKey)
		details[i].Options = s.signAnswers(ctx, details[i].Options)
	}
}

func (s *Service) signPool(ctx context.Context, res *dto.ExamPoolRes) *dto.ExamPoolRes {
	if res != nil && res.ExamPool != nil {
		s.signQuestions(ctx, res.ExamPool.Questions)
	}
	return res
}
