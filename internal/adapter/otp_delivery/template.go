package otp_delivery

import (
	"fmt"
	"time"

	"math-ai.com/math-ai/internal/shared/enum"
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
				Subject:  "Your Math-AI login code",
				BodyText: fmt.Sprintf("Your Math-AI login code is %s. It expires in %d minutes.", code, minutes),
			}
		case enum.OtpTypeRegister:
			return renderedTemplate{
				Subject:  "Confirm your Math-AI account",
				BodyText: fmt.Sprintf("Your Math-AI verification code is %s. It expires in %d minutes.", code, minutes),
			}
		case enum.OtpTypeForgotPassword:
			return renderedTemplate{
				Subject:  "Math-AI password reset code",
				BodyText: fmt.Sprintf("Your password reset code is %s. It expires in %d minutes. If you didn't request this, ignore this message.", code, minutes),
			}
		case enum.OtpTypeChangePassword:
			return renderedTemplate{
				Subject:  "Confirm your password change",
				BodyText: fmt.Sprintf("Your confirmation code is %s. It expires in %d minutes.", code, minutes),
			}
		case enum.OtpTypeVerifyEmail:
			return renderedTemplate{
				Subject:  "Verify your email",
				BodyText: fmt.Sprintf("Your Math-AI email verification code is %s. It expires in %d minutes.", code, minutes),
			}
		case enum.OtpTypeVerifyPhone:
			return renderedTemplate{
				Subject:  "Verify your phone",
				BodyText: fmt.Sprintf("Your Math-AI phone verification code is %s. It expires in %d minutes.", code, minutes),
			}
		}
	}

	// Vietnamese (default)
	switch otpType {
	case enum.OtpTypeLogin2FA:
		return renderedTemplate{
			Subject:  "Mã đăng nhập Math-AI",
			BodyText: fmt.Sprintf("Mã đăng nhập Math-AI của bạn là %s. Mã có hiệu lực trong %d phút.", code, minutes),
		}
	case enum.OtpTypeRegister:
		return renderedTemplate{
			Subject:  "Xác nhận tài khoản Math-AI",
			BodyText: fmt.Sprintf("Mã xác minh Math-AI của bạn là %s. Mã có hiệu lực trong %d phút.", code, minutes),
		}
	case enum.OtpTypeForgotPassword:
		return renderedTemplate{
			Subject:  "Mã đặt lại mật khẩu Math-AI",
			BodyText: fmt.Sprintf("Mã đặt lại mật khẩu của bạn là %s. Mã có hiệu lực trong %d phút. Nếu không phải bạn yêu cầu, hãy bỏ qua tin nhắn này.", code, minutes),
		}
	case enum.OtpTypeChangePassword:
		return renderedTemplate{
			Subject:  "Xác nhận đổi mật khẩu",
			BodyText: fmt.Sprintf("Mã xác nhận của bạn là %s. Mã có hiệu lực trong %d phút.", code, minutes),
		}
	case enum.OtpTypeVerifyEmail:
		return renderedTemplate{
			Subject:  "Xác minh email",
			BodyText: fmt.Sprintf("Mã xác minh email Math-AI của bạn là %s. Mã có hiệu lực trong %d phút.", code, minutes),
		}
	case enum.OtpTypeVerifyPhone:
		return renderedTemplate{
			Subject:  "Xác minh số điện thoại",
			BodyText: fmt.Sprintf("Mã xác minh số điện thoại Math-AI của bạn là %s. Mã có hiệu lực trong %d phút.", code, minutes),
		}
	}

	return renderedTemplate{
		Subject:  "Math-AI",
		BodyText: fmt.Sprintf("Math-AI code: %s (expires in %d minutes)", code, minutes),
	}
}
