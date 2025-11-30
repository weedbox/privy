package privy

// Storage defines the interface for persisting and retrieving RBAC data
type Storage interface {
	// Resource operations
	CreateResource(resource *Resource) error
	GetResource(key string, parentID *uint) (*Resource, error)
	GetResourceByID(id uint) (*Resource, error)
	ListResources(parentID *uint) ([]Resource, error)
	UpdateResource(resource *Resource) error
	DeleteResource(id uint) error

	// Action operations
	CreateActions(resourceID uint, actions []Action) error
	GetAction(resourceID uint, key string) (*Action, error)
	ListActions(resourceID uint) ([]Action, error)
	DeleteAction(id uint) error

	// Role operations
	CreateRole(role *Role) error
	GetRole(key string) (*Role, error)
	GetRoleByID(id uint) (*Role, error)
	ListRoles() ([]Role, error)
	UpdateRole(role *Role) error
	DeleteRole(id uint) error

	// Initialize creates necessary tables/schemas
	Initialize() error
}
