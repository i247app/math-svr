package status

// StatusCode is the application status the client reads from the response
// envelope's "mstatus" field. The generic codes mirror HTTP and live here;
// every other code is declared, with its EN / VN messages (Register), by the
// package that owns it, in that package's status.go. Codes are wire
// contract: never renumber or reuse one (retired numbers: registry.go).
//
// Blocks taken (100 per aggregate; rules/conventions.md §4 is the full table):
//
//	domain/<agg>   40xx user · 41xx profile · 42xx auth (domain/login) · 43xx program ·
//	               44xx grade · 45xx semester · 46xx device · 47xx otp · 49xx school ·
//	               50xx notification · 121xx seq · 122xx–125xx classroom ·
//	               126xx–127xx exercise · 134xx banner · 135xx chat · 136xx presence ·
//	               137xx exam · 139xx role · 143xx prompt · 144xx questionpool ·
//	               146xx llmroute
//	module/<x>     120xx job/task · 130xx server · 131xx home · 133xx socket ·
//	               138xx pow · 141xx session
//	provider/<k>   80xx storage · 90xx sms · 100xx llm
//	shared         140xx pagination (shared/pagination) · 142xx rate limit
//	               (bootstrap/middleware) · 60xx messaging, 70xx cache (legacy.go)
//	reserved       145xx QUESTION_* (ma_questions) · next free block: 14700
//
// Register panics on a number taken twice, so a collision stops the server
// at boot.
type StatusCode int

// Generic codes. FAIL (= BAD_REQUEST) and ACCEPTED (= NO_DATA) are aliases
// of a registered number, not codes of their own.
const (
	SUCCESS               StatusCode = 200
	CREATED               StatusCode = 201
	ACCEPTED              StatusCode = 202
	NO_DATA               StatusCode = 202
	NO_CONTENT            StatusCode = 204
	BAD_REQUEST           StatusCode = 400
	FAIL                  StatusCode = 400
	UNAUTHORIZED          StatusCode = 401
	FORBIDDEN             StatusCode = 403
	NOT_FOUND             StatusCode = 404
	CONFLICT              StatusCode = 409
	INTERNAL_SERVER_ERROR StatusCode = 500
)
