package service

import (
	"fmt"
	"log/slog"
	"sortedstartup/authservice/dao"
)

type TenantService struct {
	tenantDAO dao.TenantDAO
	userDAO   dao.UserDAO
}

func NewTenantService(tenantDAO dao.TenantDAO, userDAO dao.UserDAO) *TenantService {
	slog.Debug("authservice:service:NewTenantService")
	return &TenantService{
		tenantDAO: tenantDAO,
		userDAO:   userDAO,
	}
}

func (t *TenantService) GetTenantsList(page int64, pageSize int64) ([]*dao.Tenant, error) {
	slog.Info("authservice:service:GetTenantsList", "page", page, "pageSize", pageSize)
	tenants, err := t.tenantDAO.GetTenantsList(page, pageSize)
	if err != nil {
		slog.Error("authservice:service:GetTenantsList", "error", err)
		return nil, err
	}
	return tenants, nil
}

func (t *TenantService) GetTenantUsers(tenantID string, page int64, pageSize int64) ([]*dao.TenantUser, error) {
	slog.Info("authservice:service:GetTenantUsers", "tenantID", tenantID, "page", page, "pageSize", pageSize)
	users, err := t.tenantDAO.GetTenantUsers(tenantID, page, pageSize)
	if err != nil {
		slog.Error("authservice:service:GetTenantUsers", "error", err)
		return nil, err
	}
	return users, nil
}

func (t *TenantService) CreateTenant(name string, description string, tenantType int64, createdBy string) (string, error) {
	slog.Info("authservice:service:CreateTenant", "name", name, "description", description, "tenantType", tenantType, "createdBy", createdBy)
	tenantID, err := t.tenantDAO.CreateTenant(name, description, tenantType, createdBy)
	if err != nil {
		slog.Error("authservice:service:CreateTenant", "error", err)
		return "", err
	}
	return tenantID, nil
}

func (t *TenantService) AddUserToTenant(tenantID string, userID string, email string, roleID string) (*dao.TenantUser, error) {
	slog.Info("authservice:service:AddUserToTenant", "tenantID", tenantID, "userID", userID, "email", email, "roleID", roleID)

	// Determine user ID - either provided directly or look up by email
	finalUserID := userID
	if finalUserID == "" && email != "" {
		// Look up user by email
		var err error
		finalUserID, err = t.userDAO.GetUserIDByEmail(email)
		if err != nil {
			slog.Error("authservice:service:AddUserToTenant", "step", "lookup user by email", "error", err)
			return nil, fmt.Errorf("failed to find user by email: %w", err)
		}
		if finalUserID == "" {
			return nil, fmt.Errorf("user not found with email: %s", email)
		}
	}

	if finalUserID == "" {
		return nil, fmt.Errorf("either user_id or email must be provided")
	}

	// Add user to tenant with role
	_, err := t.tenantDAO.AddUserToTenant(tenantID, finalUserID, roleID)
	if err != nil {
		slog.Error("authservice:service:AddUserToTenant", "error", err)
		return nil, err
	}

	// Get the tenant users to return the created relationship
	// Using page 1 and large page size to get all users
	tenantUsers, err := t.tenantDAO.GetTenantUsers(tenantID, 1, 1000)
	if err != nil {
		slog.Error("authservice:service:AddUserToTenant", "step", "get tenant users", "error", err)
		return nil, err
	}

	// Find the specific user we just added
	for _, tu := range tenantUsers {
		if tu.UserId == finalUserID {
			return tu, nil
		}
	}

	return nil, fmt.Errorf("failed to retrieve added user")
}

func (t *TenantService) RemoveUserFromTenant(tenantID string, userID string, email string) (string, error) {
	slog.Info("authservice:service:RemoveUserFromTenant", "tenantID", tenantID, "userID", userID, "email", email)

	// Determine user ID - either provided directly or look up by email
	finalUserID := userID
	if finalUserID == "" && email != "" {
		// Look up user by email
		var err error
		finalUserID, err = t.userDAO.GetUserIDByEmail(email)
		if err != nil {
			slog.Error("authservice:service:RemoveUserFromTenant", "step", "lookup user by email", "error", err)
			return "", fmt.Errorf("failed to find user by email: %w", err)
		}
		if finalUserID == "" {
			return "", fmt.Errorf("user not found with email: %s", email)
		}
	}

	if finalUserID == "" {
		return "", fmt.Errorf("either user_id or email must be provided")
	}

	// Remove user from tenant
	message, err := t.tenantDAO.RemoveUserFromTenant(tenantID, finalUserID)
	if err != nil {
		slog.Error("authservice:service:RemoveUserFromTenant", "error", err)
		return "", err
	}

	return message, nil
}
