// Package models contains domain structs shared across the examples.
package models

// User models a simple user record.
type User struct {
	ID       string  `json:"id"`
	Email    string  `json:"email"`
	Username string  `json:"username"`
	Region   string  `json:"region,omitempty"`
	IsActive bool    `json:"isActive"`
	Profile  Profile `json:"profile,omitempty"`
}

// Profile captures nested profile data.
type Profile struct {
	AvatarURL string `json:"avatarUrl"`
	Bio       string `json:"bio,omitempty"`
}

// Role represents a role that can be assigned to a user.
type Role struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// UserRole links users to roles.
type UserRole struct {
	ID     string `json:"id"`
	UserID string `json:"userId"`
	RoleID string `json:"roleId"`
}

// Post represents a blog-style post.
type Post struct {
	ID        string   `json:"id"`
	AuthorID  string   `json:"authorId"`
	Title     string   `json:"title"`
	Body      string   `json:"body"`
	Tags      []string `json:"tags,omitempty"`
	CreatedAt string   `json:"createdAt"`
}

// Comment is a comment on a post.
type Comment struct {
	ID        string `json:"id"`
	PostID    string `json:"postId"`
	AuthorID  string `json:"authorId"`
	Body      string `json:"body"`
	CreatedAt string `json:"createdAt"`
}

// Order models a purchase with an amount that can be aggregated.
type Order struct {
	ID        string  `json:"id"`
	Region    string  `json:"region"`
	Status    string  `json:"status"`
	Total     float64 `json:"total"`
	CreatedAt string  `json:"createdAt"`
}

// Event captures audit or change events that can be streamed.
type Event struct {
	ID        string         `json:"id"`
	Type      string         `json:"type"`
	Payload   map[string]any `json:"payload,omitempty"`
	CreatedAt string         `json:"createdAt"`
}

// Session tracks active user sessions.
type Session struct {
	ID        string `json:"id"`
	UserID    string `json:"userId"`
	ExpiresAt string `json:"expiresAt"`
}
