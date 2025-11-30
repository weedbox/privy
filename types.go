package privy

import "time"

// Action represents an action that can be performed on a resource
type Action struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	Key         string    `gorm:"uniqueIndex:idx_resource_action;not null" json:"key"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	ResourceID  uint      `gorm:"uniqueIndex:idx_resource_action;not null" json:"resource_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// DefineAction is a helper function to create an Action
func DefineAction(key, name, description string) Action {
	return Action{
		Key:         key,
		Name:        name,
		Description: description,
	}
}

// Resource represents a resource in the system
type Resource struct {
	ID           uint       `gorm:"primarykey" json:"id"`
	Key          string     `gorm:"uniqueIndex:idx_parent_key;not null" json:"key"`
	Name         string     `json:"name"`
	Description  string     `json:"description"`
	ParentID     *uint      `gorm:"uniqueIndex:idx_parent_key;index" json:"parent_id"`
	Actions      []Action   `gorm:"foreignKey:ResourceID;constraint:OnDelete:CASCADE" json:"actions"`
	SubResources []Resource `gorm:"foreignKey:ParentID;constraint:OnDelete:CASCADE" json:"sub_resources"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// ResourceConfig is used to configure a resource during creation
type ResourceConfig struct {
	Key          string
	Name         string
	Description  string
	Actions      []Action
	SubResources []Resource
}

// Role represents a role in the system
type Role struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	Key         string    `gorm:"uniqueIndex;not null" json:"key"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Permissions []string  `gorm:"serializer:json" json:"permissions"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// RoleConfig is used to configure a role during creation
type RoleConfig struct {
	Name        string
	Description string
	Permissions []string
}
