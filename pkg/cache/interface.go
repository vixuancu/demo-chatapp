
package cache

import "time"

type RedisCacheService interface {
	Set(key string, value any, expiration time.Duration) error // hàm lưu trữ dữ liệu vào cache
	Get(key string, dest any) error // hàm lấy dữ liệu từ cache
	Clear(pattern string) error // hàm xóa dữ liệu khỏi cache theo pattern
	Exists(key string) (bool, error) // hàm kiểm tra xem khóa có tồn tại trong cache hay không
}
