package otp_delivery

import "context"

// Deliverer is the contract every concrete OTP channel implements. The
// adapter routes to one of these based on identifier shape or an explicit
// channel choice from the caller.
type Deliverer interface {
	Name() ChannelName
	Deliver(ctx context.Context, msg Message) error
}
