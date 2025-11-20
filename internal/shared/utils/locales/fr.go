package locales

import (
	"fmt"

	"math-ai.com/math-ai/internal/shared/constant/status"
)

var (
	FR LanguageType = "fr"
)

func GetMessageFRFromStatus(statusCode status.Code, args ...any) string {
	switch statusCode {
	case status.USER_INVALID_PARAMS:
		return "Paramètres invalides"
	case status.USER_INVALID_ID:
		return "ID utilisateur invalide"
	case status.USER_NOT_FOUND:
		return "Utilisateur non trouvé"
	case status.USER_MISSING_FIRST_NAME:
		return "Le prénom est requis"
	case status.USER_MISSING_LAST_NAME:
		return "Le nom de famille est requis"
	case status.USER_MISSING_EMAIL:
		return "L'email est requis"
	case status.USER_MISSING_PASSWORD:
		return "Le mot de passe est requis"
	case status.USER_INVALID_EMAIL:
		return "Format d'email invalide"
	case status.USER_EMAIL_ALREADY_EXISTS:
		return "L'email existe déjà"
	case status.USER_MISSING_PHONE:
		return "Le téléphone est requis"
	case status.USER_INVALID_PHONE:
		return "Format de téléphone invalide"
	case status.USER_PHONE_ALREADY_EXISTS:
		return "Le téléphone existe déjà"
	case status.USER_INVALID_ROLE:
		return fmt.Sprintf("Rôle invalide. Valid rôle are: %v", args)
	case status.USER_INVALID_STATUS:
		return fmt.Sprintf("Statut invalide. Valid statuts are: %v", args)
	case status.DEVICE_INVALID_PARAMS:
		return "Paramètres de l'appareil invalides"
	case status.DEVICE_MISSING_UUID:
		return "L'UUID de l'appareil est requis"
	case status.DEVICE_BLOCKED:
		return "L'appareil est bloqué"
	case status.DEVICE_MISSING_NAME:
		return "Le nom de l'appareil est requis"
	case status.LOGIN_MISSING_PARAMETERS:
		return "Paramètres requis manquants"
	case status.LOGIN_WRONG_CREDENTIALS:
		return "Identifiants de connexion incorrects"
	case status.BLOCK_MISSING_TYPE:
		return "Le type de bloc est requis"
	case status.BLOCK_INVALID_TYPE:
		return fmt.Sprintf("Type de bloc invalide. Statuts valides sont : %v", args)
	case status.BLOCK_MISSING_VALUE:
		return "La valeur du bloc est requise"
	case status.OTP_MISSING_PURPOSE:
		return "Le but de l'OTP est requis"
	case status.OTP_INVALID_PURPOSE:
		return "But OTP invalide"
	case status.OTP_MISSING_IDENTIFIER:
		return "L'identifiant est requis"
	case status.OTP_MISSING_CODE:
		return "Le code OTP est requis"
	case status.OTP_INVALID_CODE:
		return "Code OTP invalide"
	case status.OTP_STILL_ACTIVE:
		return fmt.Sprintf("L'OTP est toujours actif, veuillez réessayer après %d secondes", args...)
	case status.OTP_EXCEED_MAX_SEND:
		return "Nombre maximum d'envois d'OTP dépassé"
	case status.OTP_EXCEED_MAX_VERIFY:
		return fmt.Sprintf("Nombre maximum de tentatives de vérification OTP dépassé, veuillez attendre %d secondes pour redemander un OTP", args...)
	case status.OTP_EXPIRED:
		return "L'OTP a expiré"
	case status.OTP_NOT_ALLOWED:
		return "Action OTP non autorisée"
	case status.OTP_BLOCK_DEVICE:
		return fmt.Sprintf("Pour des raisons de sécurité, cet appareil a été bloqué pendant %d minutes", args...)
	case status.OTP_BLOCK_DEVICE_PHONE:
		return fmt.Sprintf("Pour des raisons de sécurité, cet appareil et ce numéro de téléphone ont été bloqués pendant %d minutes", args...)
	case status.OTP_BLOCK_DEVICE_EMAIL:
		return fmt.Sprintf("Pour des raisons de sécurité, cet appareil et cet email ont été bloqués pendant %d minutes", args...)
	case status.SUCCESS:
		return "Succès"
	default:
		return "Unknown"
	}
}
