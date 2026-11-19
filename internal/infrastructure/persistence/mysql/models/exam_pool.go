package models

type ExamPoolModel struct {
	Audit

	ExamId        int64
	ReqExamType   string
	ReqGrade      int
	ReqLevel      *int
	ReqNumQues    int
	ReqSemester   *string
	ReqProgram    *string
	ReqExtras     *string
	ExamSource    string
	Title         *string
	ShortText     *string
	QuestionsJson string
	VerifiedCount int
	ExamStatus    *string
}
