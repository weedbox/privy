package privy

import (
	"errors"

	"gorm.io/gorm"
)

var (
	ErrResourceNotFound = errors.New("resource not found")
	ErrActionNotFound   = errors.New("action not found")
	ErrRoleNotFound     = errors.New("role not found")
	ErrDuplicateKey     = errors.New("duplicate key")
)

// GormStorage implements Storage interface using GORM
type GormStorage struct {
	db *gorm.DB
}

// NewGormStorage creates a new GormStorage instance
func NewGormStorage(db *gorm.DB) *GormStorage {
	return &GormStorage{db: db}
}

// Initialize creates necessary tables
func (s *GormStorage) Initialize() error {
	return s.db.AutoMigrate(&Resource{}, &Action{}, &Role{})
}

// Resource operations

func (s *GormStorage) CreateResource(resource *Resource) error {
	return s.db.Create(resource).Error
}

func (s *GormStorage) GetResource(key string, parentID *uint) (*Resource, error) {
	var resource Resource
	query := s.db.Where("key = ?", key)

	if parentID == nil {
		query = query.Where("parent_id IS NULL")
	} else {
		query = query.Where("parent_id = ?", *parentID)
	}

	err := query.Preload("Actions").Preload("SubResources").First(&resource).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrResourceNotFound
		}
		return nil, err
	}

	return &resource, nil
}

func (s *GormStorage) GetResourceByID(id uint) (*Resource, error) {
	var resource Resource
	err := s.db.Preload("Actions").Preload("SubResources").First(&resource, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrResourceNotFound
		}
		return nil, err
	}

	return &resource, nil
}

func (s *GormStorage) ListResources(parentID *uint) ([]Resource, error) {
	var resources []Resource
	query := s.db.Preload("Actions").Preload("SubResources")

	if parentID == nil {
		query = query.Where("parent_id IS NULL")
	} else {
		query = query.Where("parent_id = ?", *parentID)
	}

	err := query.Find(&resources).Error
	if err != nil {
		return nil, err
	}

	return resources, nil
}

func (s *GormStorage) UpdateResource(resource *Resource) error {
	return s.db.Save(resource).Error
}

func (s *GormStorage) DeleteResource(id uint) error {
	return s.db.Delete(&Resource{}, id).Error
}

// Action operations

func (s *GormStorage) CreateActions(resourceID uint, actions []Action) error {
	for i := range actions {
		actions[i].ResourceID = resourceID
	}
	return s.db.Create(&actions).Error
}

func (s *GormStorage) GetAction(resourceID uint, key string) (*Action, error) {
	var action Action
	err := s.db.Where("resource_id = ? AND key = ?", resourceID, key).First(&action).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrActionNotFound
		}
		return nil, err
	}

	return &action, nil
}

func (s *GormStorage) ListActions(resourceID uint) ([]Action, error) {
	var actions []Action
	err := s.db.Where("resource_id = ?", resourceID).Find(&actions).Error
	if err != nil {
		return nil, err
	}

	return actions, nil
}

func (s *GormStorage) DeleteAction(id uint) error {
	return s.db.Delete(&Action{}, id).Error
}

// Role operations

func (s *GormStorage) CreateRole(role *Role) error {
	return s.db.Create(role).Error
}

func (s *GormStorage) GetRole(key string) (*Role, error) {
	var role Role
	err := s.db.Where("key = ?", key).First(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRoleNotFound
		}
		return nil, err
	}

	return &role, nil
}

func (s *GormStorage) GetRoleByID(id uint) (*Role, error) {
	var role Role
	err := s.db.First(&role, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRoleNotFound
		}
		return nil, err
	}

	return &role, nil
}

func (s *GormStorage) ListRoles() ([]Role, error) {
	var roles []Role
	err := s.db.Find(&roles).Error
	if err != nil {
		return nil, err
	}

	return roles, nil
}

func (s *GormStorage) UpdateRole(role *Role) error {
	return s.db.Save(role).Error
}

func (s *GormStorage) DeleteRole(id uint) error {
	return s.db.Delete(&Role{}, id).Error
}
