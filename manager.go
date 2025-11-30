package privy

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrInvalidResourcePath = errors.New("invalid resource path")
	ErrResourceExists      = errors.New("resource already exists")
	ErrRoleExists          = errors.New("role already exists")
)

// Manager manages RBAC resources, actions, and roles
type Manager struct {
	storage Storage
}

// ManagerOption is a function that configures a Manager
type ManagerOption func(*Manager)

// WithStorage sets the storage for the manager
func WithStorage(storage Storage) ManagerOption {
	return func(m *Manager) {
		m.storage = storage
	}
}

// CreateManager creates a new Manager with the given options
func CreateManager(opts ...ManagerOption) *Manager {
	m := &Manager{}

	for _, opt := range opts {
		opt(m)
	}

	// Initialize storage if provided
	if m.storage != nil {
		m.storage.Initialize()
	}

	return m
}

// parseResourcePath parses a resource path like "article.comment.tag" into individual keys
func parseResourcePath(path string) []string {
	return strings.Split(path, ".")
}

// getResourceByPath gets a resource by its path (e.g., "article.comment")
func (m *Manager) getResourceByPath(path string) (*Resource, error) {
	keys := parseResourcePath(path)
	if len(keys) == 0 {
		return nil, ErrInvalidResourcePath
	}

	var parentID *uint
	var resource *Resource
	var err error

	for _, key := range keys {
		resource, err = m.storage.GetResource(key, parentID)
		if err != nil {
			return nil, err
		}
		parentID = &resource.ID
	}

	return resource, nil
}

// CreateResource creates a new resource with the given configuration
func (m *Manager) CreateResource(config ResourceConfig) (*Resource, error) {
	resource := &Resource{
		Key:         config.Key,
		Name:        config.Name,
		Description: config.Description,
	}

	// Check if resource already exists
	existing, err := m.storage.GetResource(config.Key, nil)
	if err == nil && existing != nil {
		return nil, ErrResourceExists
	}

	// Create the resource
	if err := m.storage.CreateResource(resource); err != nil {
		return nil, err
	}

	// Create actions
	if len(config.Actions) > 0 {
		if err := m.storage.CreateActions(resource.ID, config.Actions); err != nil {
			return nil, err
		}
	}

	// Create sub-resources
	for _, subConfig := range config.SubResources {
		subResource := &Resource{
			Key:         subConfig.Key,
			Name:        subConfig.Name,
			Description: subConfig.Description,
			ParentID:    &resource.ID,
		}

		if err := m.storage.CreateResource(subResource); err != nil {
			return nil, err
		}

		// Create actions for sub-resource
		if len(subConfig.Actions) > 0 {
			if err := m.storage.CreateActions(subResource.ID, subConfig.Actions); err != nil {
				return nil, err
			}
		}
	}

	// Reload resource with all relations
	return m.storage.GetResourceByID(resource.ID)
}

// AddActions adds actions to an existing resource
func (m *Manager) AddActions(resourcePath string, actions []Action) error {
	resource, err := m.getResourceByPath(resourcePath)
	if err != nil {
		return err
	}

	return m.storage.CreateActions(resource.ID, actions)
}

// CreateResources creates sub-resources under an existing resource
func (m *Manager) CreateResources(parentPath string, subResources []Resource) error {
	parent, err := m.getResourceByPath(parentPath)
	if err != nil {
		return err
	}

	for _, subConfig := range subResources {
		// Check if sub-resource already exists
		existing, err := m.storage.GetResource(subConfig.Key, &parent.ID)
		if err == nil && existing != nil {
			// Sub-resource exists, just add actions
			if len(subConfig.Actions) > 0 {
				if err := m.storage.CreateActions(existing.ID, subConfig.Actions); err != nil {
					return err
				}
			}
			continue
		}

		// Create new sub-resource
		subResource := &Resource{
			Key:         subConfig.Key,
			Name:        subConfig.Name,
			Description: subConfig.Description,
			ParentID:    &parent.ID,
		}

		if err := m.storage.CreateResource(subResource); err != nil {
			return err
		}

		// Create actions for sub-resource
		if len(subConfig.Actions) > 0 {
			if err := m.storage.CreateActions(subResource.ID, subConfig.Actions); err != nil {
				return err
			}
		}
	}

	return nil
}

// GetResource gets a resource by its path
func (m *Manager) GetResource(path string) (*Resource, error) {
	return m.getResourceByPath(path)
}

// ListResources lists all top-level resources
func (m *Manager) ListResources() ([]Resource, error) {
	return m.storage.ListResources(nil)
}

// CreateRole creates a new role with the given configuration
func (m *Manager) CreateRole(key string, config RoleConfig) (*Role, error) {
	// Check if role already exists
	existing, err := m.storage.GetRole(key)
	if err == nil && existing != nil {
		return nil, ErrRoleExists
	}

	role := &Role{
		Key:         key,
		Name:        config.Name,
		Description: config.Description,
		Permissions: config.Permissions,
	}

	if err := m.storage.CreateRole(role); err != nil {
		return nil, err
	}

	return role, nil
}

// AssignPermissions adds permissions to an existing role
func (m *Manager) AssignPermissions(roleKey string, permissions []string) error {
	role, err := m.storage.GetRole(roleKey)
	if err != nil {
		return err
	}

	// Add permissions (avoiding duplicates)
	permMap := make(map[string]bool)
	for _, p := range role.Permissions {
		permMap[p] = true
	}

	for _, p := range permissions {
		if !permMap[p] {
			role.Permissions = append(role.Permissions, p)
			permMap[p] = true
		}
	}

	return m.storage.UpdateRole(role)
}

// RemovePermissions removes permissions from an existing role
func (m *Manager) RemovePermissions(roleKey string, permissions []string) error {
	role, err := m.storage.GetRole(roleKey)
	if err != nil {
		return err
	}

	// Create a map for quick lookup
	toRemove := make(map[string]bool)
	for _, p := range permissions {
		toRemove[p] = true
	}

	// Filter out permissions to remove
	newPermissions := make([]string, 0)
	for _, p := range role.Permissions {
		if !toRemove[p] {
			newPermissions = append(newPermissions, p)
		}
	}

	role.Permissions = newPermissions
	return m.storage.UpdateRole(role)
}

// GetRole gets a role by its key
func (m *Manager) GetRole(key string) (*Role, error) {
	return m.storage.GetRole(key)
}

// ListRoles lists all roles
func (m *Manager) ListRoles() ([]Role, error) {
	return m.storage.ListRoles()
}

// DeleteRole deletes a role by its key
func (m *Manager) DeleteRole(key string) error {
	role, err := m.storage.GetRole(key)
	if err != nil {
		return err
	}

	return m.storage.DeleteRole(role.ID)
}

// DeleteResource deletes a resource by its path
func (m *Manager) DeleteResource(path string) error {
	resource, err := m.getResourceByPath(path)
	if err != nil {
		return err
	}

	return m.storage.DeleteResource(resource.ID)
}

// BuildPermissionString builds a permission string from resource path and action
func BuildPermissionString(resourcePath, action string) string {
	return fmt.Sprintf("%s.%s", resourcePath, action)
}
