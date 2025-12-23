package api

import (
	"context"
	"log/slog"
	proto "sortedstartup/authservice/proto"
	"sortedstartup/authservice/service"

	auth "sortedstartup/common/auth"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserServiceAPI struct {
	proto.UnimplementedUserServiceServer
	userService *service.UserService
}

func NewUserServiceAPI(userService *service.UserService) *UserServiceAPI {
	return &UserServiceAPI{
		userService: userService,
	}
}

func (a *UserServiceAPI) UsersList(ctx context.Context, req *proto.UsersListRequest) (*proto.UsersListResponse, error) {
	isAdmin, err := auth.IsUserAdmin(ctx)
	if err != nil {
		slog.Error("paymentservice:api:GetDashboardData", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to check if user is admin: %v", err)
	}
	if !isAdmin {
		return nil, status.Errorf(codes.PermissionDenied, "user is not admin")
	}
	users, err := a.userService.GetUsersList(req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}
	protoUsers := make([]*proto.User, len(users))
	for i, user := range users {
		protoUsers[i] = &proto.User{
			Id:    user.ID,
			Email: user.Email,
			Name:  user.Name,
			Roles: user.Roles,
		}
	}
	return &proto.UsersListResponse{
		Users: protoUsers,
	}, nil
}
