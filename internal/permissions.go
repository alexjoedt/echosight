package echosight

import (
	"github.com/alexjoedt/echosight/internal/permissions"
	"github.com/google/uuid"
)

type UserHostPermissions struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	HostID     uuid.UUID
	Permission permissions.Permission
}

type UserPermissions struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	AppPermission AppPermission
}

type AppPermission string

const (
	PermissionCreateHosts    AppPermission = "create_hosts"
	PermissionCreateServices AppPermission = "create_services"
)
