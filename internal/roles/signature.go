package roles

import (
	"crypto/subtle"
	"errors"
	"os"

	"github.com/alkuinvito/ai-assistant/pkg/hash"
	"github.com/google/uuid"
)

// ErrTampered is returned when a stored role assignment does not match its
// signature, indicating the rows were modified directly in the database.
var ErrTampered = errors.New("rbac data integrity verification failed")

// signingSecret returns the secret used to sign role assignments. It prefers
// RBAC_HASH_SECRET and falls back to the shared HMAC_KEY.
func signingSecret() string {
	if secret := os.Getenv("RBAC_HASH_SECRET"); secret != "" {
		return secret
	}

	return os.Getenv("HMAC_KEY")
}

func signPermission(permissionId uuid.UUID, name string) (string, error) {
	return hash.Sign(signingSecret(), "permission:"+permissionId.String()+":"+name)
}

func signRole(roleId uuid.UUID, name string) (string, error) {
	return hash.Sign(signingSecret(), "role:"+roleId.String()+":"+name)
}

func signRolePermission(roleId, permissionId uuid.UUID) (string, error) {
	return hash.Sign(signingSecret(), "role_permission:"+roleId.String()+":"+permissionId.String())
}

func signUserRole(userId, roleId uuid.UUID) (string, error) {
	return hash.Sign(signingSecret(), "user_role:"+userId.String()+":"+roleId.String())
}

func verifyPermission(record Permission) error {
	expected, err := signPermission(record.Id, record.Name)
	if err != nil {
		return err
	}

	if subtle.ConstantTimeCompare([]byte(record.Signature), []byte(expected)) != 1 {
		return ErrTampered
	}

	return nil
}

func verifyRole(record Role) error {
	expected, err := signRole(record.Id, record.Name)
	if err != nil {
		return err
	}

	if subtle.ConstantTimeCompare([]byte(record.Signature), []byte(expected)) != 1 {
		return ErrTampered
	}

	return nil
}

func verifyRolePermission(record RolePermission) error {
	expected, err := signRolePermission(record.RoleId, record.PermissionId)
	if err != nil {
		return err
	}

	if subtle.ConstantTimeCompare([]byte(record.Signature), []byte(expected)) != 1 {
		return ErrTampered
	}

	return nil
}

func verifyUserRole(record UserRole) error {
	expected, err := signUserRole(record.UserId, record.RoleId)
	if err != nil {
		return err
	}

	if subtle.ConstantTimeCompare([]byte(record.Signature), []byte(expected)) != 1 {
		return ErrTampered
	}

	return nil
}
