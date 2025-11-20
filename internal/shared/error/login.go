package err_svc

import "errors"

var (
	ErrMissingParameters  = errors.New("missing required parameters")
	ErrMissingDeviceUUID  = errors.New("missing device UUID")
	ErrInvalidCredentials = errors.New("invalid login credentials")
)
