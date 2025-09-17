package v1Dto

import "chat-app/internal/db/sqlc"

// UserDTO để trả về thông tin người dùng mà không bao gồm mật khẩu, hoặc các thông tin nhạy cảm khác
type UserDTO struct {
	UUID      string `json:"uuid"`
	Name      string `json:"full_name"`
	Email     string `json:"email_address"`
	Role      string `json:"role"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

type CreateUserInput struct {
	Name     string `json:"user_fullname" binding:"required"`
	Email    string `json:"user_email" binding:"required,email,email_advanced"`
	Password string `json:"user_password" binding:"required,password_strong"`
}

func (input *CreateUserInput) MapCreateInputToModel() sqlc.CreateUserParams {
	return sqlc.CreateUserParams{
		UserEmail:    input.Email,
		UserPassword: input.Password,
		UserFullname: input.Name,
	}

}
func MapUserToDTO(user sqlc.User) *UserDTO {
	return &UserDTO{
		UUID:      user.UserUuid.String(),
		Name:      user.UserFullname,
		Email:     user.UserEmail,
		Role:      user.UserRole,
		UpdatedAt: user.UserUpdatedAt.String(),
		CreatedAt: user.UserCreatedAt.String(),
	}
}

func MapUsersToDTO(users []sqlc.User) []UserDTO {
	dtoUsers := make([]UserDTO,0, len(users))
	for _, user := range users {
		// không nên dùng index gán giá trị trực tiếp vào slice vì nếu length = 0 thì sẽ panic
		dtoUsers = append(dtoUsers, *MapUserToDTO(user)) 
	}
	return dtoUsers
}

