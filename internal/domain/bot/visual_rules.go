package bot

// The visual contract — which question types exist, when icons are
// allowed, which emoji to use — is shared by every generation prompt in
// the product. It lives in its own file rather than inside one aggregate's
// template file so that retiring an aggregate cannot take the rules with
// it, and so a change to the icon policy is made in exactly one place.
//
// The icon DECISION itself is not here: it belongs to the grade band (see
// gradeProfile.iconLine*), and this block only tells the model to obey
// whatever that band decided.

const visualQuestionRulesVN = `
VISUAL QUESTION RULES:
- Mỗi câu có "question_type": ARITHMETIC (câu chữ thuần, mặc định) hoặc COUNT (đếm bằng icon). KHÔNG tạo bất kỳ question_type nào khác.
- VIỆC dùng COUNT/icon hay không và với tần suất nào do CHÍNH SÁCH icon trong GRADE PROFILE quyết định HOÀN TOÀN. Nếu chính sách đó là TẮT (hoặc không có GRADE PROFILE), dùng ARITHMETIC cho mọi câu và TUYỆT ĐỐI không phát sinh emoji hay token [icon:...]. Chỉ dùng COUNT khi chính sách cho phép, và trong đúng tần suất nó nêu.
- COUNT (chỉ khi chính sách theo lớp cho phép): "question_name" hiển thị các vật để đếm hoặc cộng bằng emoji kèm toán tử (+, "?"), ví dụ "🏓 🏓 🏓 + 🏓 🏓 🏓 = ?". Đáp án là số; "topic" là "phép đếm".

ICONS (chỉ dùng cho COUNT; TUYỆT ĐỐI không dùng trong ARITHMETIC):
- Emoji: với vật đếm được, dùng emoji phổ thông, thân thiện trẻ em, chèn trực tiếp và cách nhau bởi dấu cách, ví dụ 🏓 🍎 ⭐ 🐟 🎈 🚗 🌸 🍓 ⚽ 🐶. Mỗi câu chỉ dùng MỘT loại emoji.
`
