package privy

import "strings"

// CheckPermission checks if a given permission satisfies the required permission.
// It supports hierarchical permission checking:
// - Exact match: "user.create" == "user.create"
// - Group match: "user" includes "user.create"
// - Hierarchical match: "infrastructure" includes "infrastructure.vm.start"
//
// Examples:
//   CheckPermission("user.create", "user.create")              // true (exact match)
//   CheckPermission("user.create", "user")                     // true (group match)
//   CheckPermission("infrastructure.vm", "infrastructure.vm.start") // false (required is more specific than given)
//   CheckPermission("infrastructure", "infrastructure.vm.stop") // true (hierarchical match)
//   CheckPermission("user.delete", "user.update")              // false (different permissions)
func CheckPermission(requiredPermission, givenPermission string) bool {
	// Exact match
	if requiredPermission == givenPermission {
		return true
	}

	// Check if given permission is a parent/group of the required permission
	// e.g., given "user" should match required "user.create"
	// e.g., given "infrastructure" should match required "infrastructure.vm.start"
	if strings.HasPrefix(requiredPermission, givenPermission+".") {
		return true
	}

	// Check if required permission is a parent/group of the given permission
	// e.g., given "infrastructure.vm.start" should match required "infrastructure.vm"
	// e.g., given "infrastructure.vm.start" should match required "infrastructure"
	if strings.HasPrefix(givenPermission, requiredPermission+".") {
		return true
	}

	return false
}

// CheckPermissions checks if any of the given permissions satisfies the required permission
func CheckPermissions(requiredPermission string, givenPermissions []string) bool {
	for _, given := range givenPermissions {
		if CheckPermission(requiredPermission, given) {
			return true
		}
	}
	return false
}

// CheckRolePermission checks if a role has the required permission
func (m *Manager) CheckRolePermission(roleKey, requiredPermission string) (bool, error) {
	role, err := m.storage.GetRole(roleKey)
	if err != nil {
		return false, err
	}

	return CheckPermissions(requiredPermission, role.Permissions), nil
}

// CheckRolesPermission checks if any of the given roles has the required permission
func (m *Manager) CheckRolesPermission(roleKeys []string, requiredPermission string) (bool, error) {
	for _, roleKey := range roleKeys {
		hasPermission, err := m.CheckRolePermission(roleKey, requiredPermission)
		if err != nil {
			// Skip roles that don't exist
			if err == ErrRoleNotFound {
				continue
			}
			return false, err
		}
		if hasPermission {
			return true, nil
		}
	}
	return false, nil
}
