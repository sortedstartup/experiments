package api

import (
	"context"
	proto "sortedstartup/authservice/proto"
	"sortedstartup/authservice/service"
)

type UserServiceAPI struct {
	proto.UnimplementedUserServiceServer
	userService   *service.UserService
	tenantService *service.TenantService
}

func NewUserServiceAPI(userService *service.UserService, tenantService *service.TenantService) *UserServiceAPI {
	return &UserServiceAPI{
		userService:   userService,
		tenantService: tenantService,
	}
}

type TenantServiceAPI struct {
	proto.UnimplementedTenantServiceServer
	tenantService *service.TenantService
}

func NewTenantServiceAPI(tenantService *service.TenantService) *TenantServiceAPI {
	return &TenantServiceAPI{
		tenantService: tenantService,
	}
}

func (a *UserServiceAPI) UsersList(ctx context.Context, req *proto.UsersListRequest) (*proto.UsersListResponse, error) {
	// Extract pagination parameters
	page := req.GetPageRequest().GetPage()
	pageSize := req.GetPageRequest().GetPageSize()

	// Set defaults if not provided
	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 10
	}

	// Check if filters are provided
	filters := req.GetFilters()
	if filters != nil && filters.GetTenantId() != "" {
		// Call tenant service to get users filtered by tenant
		tenantUsers, err := a.tenantService.GetTenantUsers(filters.GetTenantId(), page, pageSize)
		if err != nil {
			return nil, err
		}

		// Convert DAO tenant users to proto users
		protoUsers := make([]*proto.User, 0, len(tenantUsers))
		for _, tenantUser := range tenantUsers {
			protoUsers = append(protoUsers, &proto.User{
				Id:    tenantUser.UserId,
				Email: tenantUser.Email,
				Name:  tenantUser.Name,
			})
		}

		return &proto.UsersListResponse{
			Users: protoUsers,
			PageResponse: &proto.PageResponse{
				TotalCount: int64(len(tenantUsers)),
			},
		}, nil
	}

	// No filters - get all users
	users, err := a.userService.GetUsersList(page, pageSize)
	if err != nil {
		return nil, err
	}

	// Convert DAO users to proto users
	protoUsers := make([]*proto.User, 0, len(users))
	for _, user := range users {
		protoUsers = append(protoUsers, &proto.User{
			Id:    user.ID,
			Email: user.Email,
			Name:  user.Name,
		})
	}

	return &proto.UsersListResponse{
		Users: protoUsers,
		PageResponse: &proto.PageResponse{
			TotalCount: int64(len(users)),
		},
	}, nil
}

func (a *TenantServiceAPI) TenantsList(ctx context.Context, req *proto.TenantsListRequest) (*proto.TenantsListResponse, error) {
	// Extract pagination parameters
	page := req.GetPageRequest().GetPage()
	pageSize := req.GetPageRequest().GetPageSize()

	// Set defaults if not provided
	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 10
	}

	// Call service layer
	tenants, err := a.tenantService.GetTenantsList(page, pageSize)
	if err != nil {
		return nil, err
	}

	// Convert DAO tenants to proto tenants
	protoTenants := make([]*proto.Tenant, 0, len(tenants))
	for _, tenant := range tenants {
		protoTenants = append(protoTenants, &proto.Tenant{
			Id:          tenant.ID,
			Name:        tenant.Name,
			Description: tenant.Description,
		})
	}

	return &proto.TenantsListResponse{
		Tenants: protoTenants,
	}, nil
}

func (a *TenantServiceAPI) CreateTenant(ctx context.Context, req *proto.CreateTenantRequest) (*proto.CreateTenantResponse, error) {
	// Call service layer - use default tenant type 0 and empty createdBy for now
	tenantID, err := a.tenantService.CreateTenant(req.GetName(), req.GetDescription(), 0, "")
	if err != nil {
		return nil, err
	}

	return &proto.CreateTenantResponse{
		Id:          tenantID,
		Name:        req.GetName(),
		Description: req.GetDescription(),
	}, nil
}

func (a *TenantServiceAPI) GetTenant(ctx context.Context, req *proto.GetTenantRequest) (*proto.GetTenantResponse, error) {
	// This RPC is defined in proto but not in the TODO list
	// We'll implement it for completeness
	return nil, nil
}

func (a *TenantServiceAPI) AddUser(ctx context.Context, req *proto.AddUserRequest) (*proto.AddUserToTenant, error) {
	// Extract user identifier - either user_id or email from the oneof
	var userID, email string
	switch user := req.GetUser().(type) {
	case *proto.AddUserRequest_UserId:
		userID = user.UserId
	case *proto.AddUserRequest_Email:
		email = user.Email
	}

	// Call service layer
	tenantUser, err := a.tenantService.AddUserToTenant(req.GetTenantId(), userID, email, req.GetRoleId())
	if err != nil {
		return nil, err
	}

	// Convert DAO TenantUser to proto TenantUser
	return &proto.AddUserToTenant{
		TenantUser: &proto.TenantUser{
			User: &proto.User{
				Id:    tenantUser.UserId,
				Email: tenantUser.Email,
				Name:  tenantUser.Name,
			},
			Tenant: &proto.Tenant{
				Id: tenantUser.TenantId,
			},
			Role: &proto.Role{
				Id:   tenantUser.Role,
				Name: tenantUser.Role,
			},
		},
	}, nil
}

func (a *TenantServiceAPI) RemoveUser(ctx context.Context, req *proto.RemoveUserRequest) (*proto.RemoveUserResponse, error) {
	// Extract user identifier - either user_id or email from the oneof
	var userID, email string
	switch user := req.GetUser().(type) {
	case *proto.RemoveUserRequest_UserId:
		userID = user.UserId
	case *proto.RemoveUserRequest_Email:
		email = user.Email
	}

	// Call service layer
	message, err := a.tenantService.RemoveUserFromTenant(req.GetTenantId(), userID, email)
	if err != nil {
		return nil, err
	}

	return &proto.RemoveUserResponse{
		Message: message,
	}, nil
}
