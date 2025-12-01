# Privy

A flexible and powerful Role-Based Access Control (RBAC) library for Go with hierarchical resource management and GORM storage support.

## Features

- **Hierarchical Resources**: Support for nested resources (e.g., `article.comment.tag`)
- **Flexible Permissions**: Hierarchical permission checking with wildcard support
- **GORM Integration**: Built-in GORM storage implementation with SQLite support
- **Extensible Storage**: Storage interface allows custom implementations
- **Simple API**: Easy-to-use API for managing resources, actions, and roles
- **Well-tested**: Comprehensive test coverage with in-memory SQLite tests

## Installation

```bash
go get github.com/weedbox/privy
```

## Quick Start

### 1. Initialize Manager with GORM Storage

```go
import (
    "github.com/weedbox/privy"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
)

// Initialize GORM with SQLite
db, err := gorm.Open(sqlite.Open("rbac.db"), &gorm.Config{})
if err != nil {
    panic(err)
}

// Create RBAC manager with GORM storage
m := privy.CreateManager(
    privy.WithStorage(privy.NewGormStorage(db)),
)
```

### 2. Define Resources and Actions

```go
r, err := m.CreateResource(privy.ResourceConfig{
    Key:         "article",
    Name:        "Article",
    Description: "News article entity",
    Actions: []privy.Action{
        privy.DefineAction("read", "Read", "Read article content"),
        privy.DefineAction("create", "Create", "Create new article"),
        privy.DefineAction("update", "Update", "Edit existing article"),
        privy.DefineAction("delete", "Delete", "Delete article"),
        privy.DefineAction("publish", "Publish", "Publish article"),
    },
    SubResources: []privy.Resource{
        {
            Key:         "comment",
            Name:        "Comment",
            Description: "Article comments",
            Actions: []privy.Action{
                privy.DefineAction("read", "Read Comment", "Read comment content"),
                privy.DefineAction("create", "Create Comment", "Create a new comment"),
                privy.DefineAction("delete", "Delete Comment", "Delete comment"),
            },
        },
    },
})
```

### 3. Extend Existing Resources

```go
// Add actions to existing resource
err := m.AddActions("article", []privy.Action{
    privy.DefineAction("share", "Share", "Share article with others"),
    privy.DefineAction("like", "Like", "Like an article"),
})

// Add sub-resources to existing resource
err := m.CreateResources("article", []privy.Resource{
    {
        Key:         "tag",
        Name:        "Tag",
        Description: "Article tags",
        Actions: []privy.Action{
            privy.DefineAction("assign", "Assign Tag", "Assign tag to article"),
        },
    },
})
```

### 4. Create Roles and Assign Permissions

```go
// Create a role with initial permissions
role, err := m.CreateRole("editor", privy.RoleConfig{
    Name:        "Editor",
    Description: "Can edit and publish articles",
    Permissions: []string{
        "article.read",
        "article.create",
        "article.update",
        "article.publish",
        "article.comment.read",
        "article.comment.create",
    },
})

// Add more permissions to existing role
err = m.AssignPermissions("editor", []string{
    "article.delete",
    "article.comment.delete",
})

// Remove permissions from role
err = m.RemovePermissions("editor", []string{
    "article.delete",
})
```

### 5. Check Permissions

The permission system supports hierarchical matching:

```go
// Exact match
privy.CheckPermission("user.create", "user.create") // true

// Group match - given permission is a parent
privy.CheckPermission("user.create", "user") // true

// Hierarchical match - given permission is a child
privy.CheckPermission("infrastructure.vm", "infrastructure.vm.start") // true

// Hierarchical match - given permission is a parent
privy.CheckPermission("infrastructure", "infrastructure.vm.stop") // true

// No match - different permissions
privy.CheckPermission("user.delete", "user.update") // false

// No match - different groups
privy.CheckPermission("user.create", "infrastructure.vm.start") // false

// Check if a role has permission
hasPermission, err := m.CheckRolePermission("editor", "article.update")

// Check if any of the roles has permission
hasPermission, err := m.CheckRolesPermission([]string{"editor", "viewer"}, "article.read")
```

## API Reference

### Manager

#### Creating Resources

- `CreateResource(config ResourceConfig) (*Resource, error)` - Create a new resource with actions and sub-resources
- `AddActions(resourcePath string, actions []Action) error` - Add actions to an existing resource
- `CreateResources(parentPath string, subResources []Resource) error` - Create sub-resources under an existing resource
- `GetResource(path string) (*Resource, error)` - Get a resource by its path (e.g., "article.comment")
- `ListResources() ([]Resource, error)` - List all top-level resources
- `DeleteResource(path string) error` - Delete a resource by its path

#### Managing Roles

- `CreateRole(key string, config RoleConfig) (*Role, error)` - Create a new role
- `AssignPermissions(roleKey string, permissions []string) error` - Add permissions to a role
- `RemovePermissions(roleKey string, permissions []string) error` - Remove permissions from a role
- `GetRole(key string) (*Role, error)` - Get a role by its key
- `ListRoles() ([]Role, error)` - List all roles
- `DeleteRole(key string) error` - Delete a role

#### Checking Permissions

- `CheckRolePermission(roleKey, requiredPermission string) (bool, error)` - Check if a role has a permission
- `CheckRolesPermission(roleKeys []string, requiredPermission string) (bool, error)` - Check if any role has a permission

### Functions

- `CheckPermission(requiredPermission, givenPermission string) bool` - Check if a given permission satisfies the required permission
- `CheckPermissions(requiredPermission string, givenPermissions []string) bool` - Check if any given permission satisfies the required permission
- `DefineAction(key, name, description string) Action` - Helper to create an Action
- `BuildPermissionString(resourcePath, action string) string` - Build a permission string from resource path and action

## Storage Interface

Implement the `Storage` interface to use custom storage backends:

```go
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
```

## Examples

See the [examples/basic](examples/basic) directory for a complete working example.

To run the example:

```bash
cd examples/basic
go run main.go
```

## Testing

Run all tests:

```bash
go test -v ./...
```

Run specific test suites:

```bash
# Test GORM storage
go test -v -run TestGormStorage

# Test Manager
go test -v -run TestManager

# Test permission checking
go test -v -run TestCheckPermission
```

## License

See [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
