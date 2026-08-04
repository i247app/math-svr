package otp_delivery

import (
	"fmt"
	"time"

	"math-ai.com/math-ai/internal/shared/enum"
)

const (
	AppName = "NumiNumi"
)

// renderedTemplate is what a Deliverer needs to construct a payload.
type renderedTemplate struct {
	Subject  string // email only
	BodyText string
}

// renderTemplate produces a localized message body for the given OTP type.
// Vietnamese is the project default — any unknown LanguageType falls back to
// VN.
//
// Templates intentionally only include the code, expiry, and a short purpose
// line. They never reference the user_id or any PII beyond what the client
// already submitted, so a leaked SMS log doesn't widen the breach.
func renderTemplate(otpType enum.OtpType, lang enum.LanguageType, code string, expiresAt time.Time) renderedTemplate {
	minutes := int(time.Until(expiresAt).Round(time.Minute).Minutes())
	if minutes < 1 {
		minutes = 1
	}

	if lang == enum.LanguageTypeEnglish {
		switch otpType {
		case enum.OtpTypeLogin2FA:
			return renderedTemplate{
				Subject:  fmt.Sprintf("%s Login 2FA", AppName),
				BodyText: fmt.Sprintf("Your %s login code is %s. It expires in %d minutes.", AppName, code, minutes),
			}
		case enum.OtpTypeRegister:
			return renderedTemplate{
				Subject:  fmt.Sprintf("Confirm your %s account", AppName),
				BodyText: fmt.Sprintf("Your %s verification code is %s. It expires in %d minutes.", AppName, code, minutes),
			}
		case enum.OtpTypeForgotPassword:
			return renderedTemplate{
				Subject:  fmt.Sprintf("%s password reset code", AppName),
				BodyText: fmt.Sprintf("Your password reset code is %s. It expires in %d minutes. If you didn't request this, ignore this message.", code, minutes),
			}
		case enum.OtpTypeChangePassword:
			return renderedTemplate{
				Subject:  fmt.Sprintf("Confirm your %s password change", AppName),
				BodyText: fmt.Sprintf("Your confirmation code is %s. It expires in %d minutes.", code, minutes),
			}
		case enum.OtpTypeVerifyEmail:
			return renderedTemplate{
				Subject:  fmt.Sprintf("Verify your %s email", AppName),
				BodyText: fmt.Sprintf("Your %s email verification code is %s. It expires in %d minutes.", AppName, code, minutes),
			}
		case enum.OtpTypeVerifyPhone:
			return renderedTemplate{
				Subject:  fmt.Sprintf("Verify your %s phone", AppName),
				BodyText: fmt.Sprintf("Your Math-AI phone verification code is %s. It expires in %d minutes.", code, minutes),
			}
		}
	}

	// Vietnamese (default)
	switch otpType {
	case enum.OtpTypeLogin2FA:
		return renderedTemplate{
			Subject:  fmt.Sprintf("Mã đăng nhập %s", AppName),
			BodyText: fmt.Sprintf("Mã đăng nhập %s của bạn là %s. Mã có hiệu lực trong %d phút.", AppName, code, minutes),
		}
	case enum.OtpTypeRegister:
		return renderedTemplate{
			Subject:  fmt.Sprintf("Xác nhận tài khoản %s", AppName),
			BodyText: fmt.Sprintf("Mã xác minh %s của bạn là %s. Mã có hiệu lực trong %d phút.", AppName, code, minutes),
		}
	case enum.OtpTypeForgotPassword:
		return renderedTemplate{
			Subject:  fmt.Sprintf("Mã đặt lại mật khẩu %s", AppName),
			BodyText: fmt.Sprintf("Mã đặt lại mật khẩu của bạn là %s. Mã có hiệu lực trong %d phút. Nếu không phải bạn yêu cầu, hãy bỏ qua tin nhắn này.", code, minutes),
		}
	case enum.OtpTypeChangePassword:
		return renderedTemplate{
			Subject:  fmt.Sprintf("Xác nhận đổi mật khẩu %s", AppName),
			BodyText: fmt.Sprintf("Mã xác nhận đổi mật khẩu %s của bạn là %s. Mã có hiệu lực trong %d phút.", AppName, code, minutes),
		}
	case enum.OtpTypeVerifyEmail:
		return renderedTemplate{
			Subject:  fmt.Sprintf("Xác minh email %s", AppName),
			BodyText: fmt.Sprintf("Mã xác minh email %s của bạn là %s. Mã có hiệu lực trong %d phút.", AppName, code, minutes),
		}
	case enum.OtpTypeVerifyPhone:
		return renderedTemplate{
			Subject:  fmt.Sprintf("Xác minh số điện thoại %s", AppName),
			BodyText: fmt.Sprintf("Mã xác minh số điện thoại %s của bạn là %s. Mã có hiệu lực trong %d phút.", AppName, code, minutes),
		}
	}

	return renderedTemplate{
		Subject:  "Math-AI",
		BodyText: fmt.Sprintf("Math-AI code: %s (expires in %d minutes)", code, minutes),
	}
}
