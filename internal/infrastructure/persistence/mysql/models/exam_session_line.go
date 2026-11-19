package models

type ExamSessionLineModel struct {
	Audit

	EsessLnId       int64
	ElinkId         int64
	EsessId         int64
	ExamId          int64
	ReqExamType     string
	QuestionNumber  int
	QuestionKind    *string
	QuestionText    *string
	QuestionTopic   *string
	QuestionTags    string // JSON ["addition", …], never NULL
	Explanation     *string
	QuestionGrade   *int
	QuestionLevel   *int
	RightOptionIDs  *string
	RightAnswerText *string
	ChildOptionIDs  string
	ChildAnswerText *string
	IsCorrect       bool
	EsessLnStatus   *string
}
