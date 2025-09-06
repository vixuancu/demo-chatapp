package utils

import "strings"

/* chuyên xử lý chuỗi kiểu string
1. Kiểm tra & tìm kiếm
strings.Contains(s, substr) → kiểm tra chuỗi con.
strings.HasPrefix(s, prefix) → bắt đầu bằng.
strings.HasSuffix(s, suffix) → kết thúc bằng.
strings.Index(s, substr) → vị trí đầu tiên.
strings.LastIndex(s, substr) → vị trí cuối.
2. Biến đổi chuỗi
strings.ToUpper(s) → viết hoa toàn bộ.
strings.ToLower(s) → viết thường toàn bộ.
strings.Title(s) (deprecated) → viết hoa chữ cái đầu mỗi từ.
strings.TrimSpace(s) → bỏ khoảng trắng đầu & cuối.
strings.ReplaceAll(s, old, new) → thay thế toàn bộ.
3. Chia & ghép
strings.Split(s, sep) → chia thành slice.
strings.Join(slice, sep) → nối slice thành string.
strings.Fields(s) → tách theo khoảng trắng.
4. Đếm & lặp
strings.Count(s, substr) → đếm số lần xuất hiện.
strings.Repeat(s, n) → lặp lại chuỗi.
5. So sánh
strings.EqualFold(s1, s2) → so sánh không phân biệt hoa thường.
*/
func NormalizeString(text string) string {
	return strings.ToLower(strings.TrimSpace(text)) // Chuyển đổi thành chữ thường và loại bỏ khoảng trắng ở đầu và cuối
}

func CapitalizeFirst(text string) string {
	if len(text) == 0 {
		return text
	}
	//text[:1]: Lấy ký tự đầu tiên của chuỗi text (từ vị trí 0 đến 1, không bao gồm 1). Kết quả là một chuỗi con chứa ký tự đầu tiên.
	//text[1:]: Lấy phần còn lại của chuỗi, từ vị trí 1 đến hết.
	return strings.ToLower(text[:1])+text[1:] 
}
