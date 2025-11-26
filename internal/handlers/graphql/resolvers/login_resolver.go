package resolvers

import (
	"github.com/graphql-go/graphql"
	"math-ai.com/math-ai/internal/applications/dto"
	di "math-ai.com/math-ai/internal/core/di/services"
	"math-ai.com/math-ai/internal/session"
	"math-ai.com/math-ai/internal/shared/constant/status"
)

type LoginResolver struct {
	loginService di.ILoginService
}

func NewLoginResolver(loginService di.ILoginService) *LoginResolver {
	return &LoginResolver{
		loginService: loginService,
	}
}

// Login resolves user login
func (r *LoginResolver) Login(params graphql.ResolveParams) (interface{}, error) {
	input, ok := params.Args["input"].(map[string]interface{})
	if !ok {
		return nil, nil
	}

	req := &dto.LoginRequest{
		LoginName:   input["login_name"].(string),
		RawPassword: input["password"].(string),
	}

	if deviceUUID, ok := input["device_uuid"].(string); ok {
		req.DeviceUUID = deviceUUID
	}
	if deviceName, ok := input["device_name"].(string); ok {
		req.DeviceName = deviceName
	}

	// Create a new session for login
	sess := session.NewSession()

	statusCode, loginResponse, err := r.loginService.Login(params.Context, sess, req)
	if err != nil || statusCode != status.SUCCESS {
		return nil, err
	}

	return loginResponse, nil
}
