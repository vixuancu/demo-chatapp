package v1Dto

import "time"

type RoomWithLastMessage struct {
	RoomID           int64     `json:"room_id"`
	RoomCode         string    `json:"room_code"`
	RoomName         *string   `json:"room_name"`
	RoomIsDirectChat bool      `json:"room_is_direct_chat"`
	RoomCreatedBy    string    `json:"room_created_by"`
	RoomCreatedAt    time.Time `json:"room_created_at"`
	RoomUpdatedAt    time.Time `json:"room_updated_at"`

	// Last message info
	LastMessage *LastMessageInfo `json:"last_message,omitempty"`
}

type LastMessageInfo struct {
	MessageID  *int64     `json:"message_id,omitempty"`
	Content    *string    `json:"content,omitempty"`
	SenderName *string    `json:"sender_name,omitempty"`
	SenderUUID *string    `json:"sender_uuid,omitempty"`
	CreatedAt  *time.Time `json:"created_at,omitempty"`
	IsOwn      bool       `json:"is_own"` // Tin nhắn của chính user này
}

type MessageWithUser struct {
	MessageID        int64     `json:"message_id"`
	RoomID           int64     `json:"room_id"`
	UserUUID         string    `json:"user_uuid"`
	UserFullname     string    `json:"user_fullname"`
	UserEmail        string    `json:"user_email"`
	Content          string    `json:"content"`
	MessageCreatedAt time.Time `json:"created_at"`
	IsOwn            bool      `json:"is_own"` // Tin nhắn của chính user này
}
