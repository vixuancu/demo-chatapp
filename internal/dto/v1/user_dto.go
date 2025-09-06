package v1Dto

import "chat-app/internal/db/sqlc"

// UserDTO để trả về thông tin người dùng mà không bao gồm mật khẩu, hoặc các thông tin nhạy cảm khác
type UserDTO struct {
	UUID      string `json:"uuid"`
	Name      string `json:"full_name"`
	Email     string `json:"email_address"`
	CreatedAt string `json:"created_at"`
} 


type CreateUserInput struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email,email_advanced"`
	Password string `json:"password" binding:"required,password_strong"`
}
func (input *CreateUserInput) MapCreateInputToModel() sqlc.CreateUserParams {
	return sqlc.CreateUserParams{
		UserEmail:    input.Email,
		UserPassword: input.Password,
		UserFullname: input.Name,
	}

}
func MapUserToDTO(user sqlc.User) UserDTO {
	return UserDTO{
		UUID:      user.UserUuid.String(),
		Name:      user.UserFullname,
		Email:     user.UserEmail,
		CreatedAt: user.UserCreatedAt.String(),
	}
}
