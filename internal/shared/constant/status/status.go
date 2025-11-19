package status

type Code = int

const (
	UNKNOW      Code = 100
	SUCCESS     Code = 200
	BAD_REQUEST Code = 400
	FAIL        Code = 400
	NOT_FOUND   Code = 404
	INTERNAL    Code = 500
)
