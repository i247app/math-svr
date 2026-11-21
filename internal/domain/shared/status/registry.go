package status

import "fmt"

// Codes are declared, with their EN / VN messages, by the package that owns
// them — `const` + Register in that package's init(), in its status.go — so
// referencing a code links its package and runs its registration. See
// code.go for the block table.

// registered holds the EN / VN messages of every registered code.
var registered = map[StatusCode][2]StatusMessage{}

// Register adds the EN / VN messages of a code. Call it from the init() of
// the package that declares the code. It panics on a number already
// registered, a retired number, or an empty message, so a collision or a
// missing translation stops the server at boot.
func Register(code StatusCode, en, vn StatusMessage) {
	if _, taken := registered[code]; taken {
		panic(fmt.Sprintf("status: code %d registered twice", code))
	}

	if en.IsEmpty() || vn.IsEmpty() {
		panic(fmt.Sprintf("status: code %d needs both an EN and a VN message", code))
	}
	registered[code] = [2]StatusMessage{en, vn}
}

// Generic codes. FAIL (= BAD_REQUEST) and ACCEPTED (= NO_DATA) are aliases
// of a registered number, not codes of their own.
func init() {
	Register(SUCCESS, "Success", "Thành công")
	Register(CREATED, "Created", "Đã tạo")
	Register(NO_DATA, "No Resource found", "Không tìm thấy dữ liệu")
	Register(NO_CONTENT, "No Content", "Không có nội dung")
	Register(BAD_REQUEST, "Bad Request", "Yêu cầu không hợp lệ")
	Register(UNAUTHORIZED, "Unauthorized", "Không có quyền truy cập")
	Register(FORBIDDEN, "Forbidden", "Bị từ chối")
	Register(NOT_FOUND, "Not Found", "Không tìm thấy tài nguyên")
	Register(CONFLICT, "Conflict", "Xung đột")
	Register(INTERNAL_SERVER_ERROR, "Internal Server Error", "Lỗi máy chủ nội bộ")
}
