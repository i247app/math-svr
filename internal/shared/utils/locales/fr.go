package locales

import (
	"fmt"

	"math-ai.com/math-ai/internal/shared/constant/status"
)

var (
	FR LanguageType = "fr"
)

func GetMessageFRFromStatus(statusCode status.Code) string {
	args := GetArgsByStatatus(statusCode)

	switch statusCode {

	case status.OK:
		return "Succès."
	case status.CREATED:
		return "Ressource créée avec succès."
	case status.FAIL:
		return "La requête a échoué."
	case status.UNAUTHORIZED:
		return "Accès non autorisé."
	case status.NOT_FOUND:
		return "Ressource non trouvée."
	case status.INTERNAL:
		return "Erreur interne du serveur."

	// User status messages
	case status.USER_MISSING_ID:
		return "L'ID utilisateur est manquant."
	case status.USER_INVALID_PARAMS:
		return "Paramètres utilisateur invalides."
	case status.USER_INVALID_ID:
		return "ID utilisateur invalide."
	case status.USER_NOT_FOUND:
		return "Utilisateur non trouvé."
	case status.USER_MISSING_NAME:
		return "Le nom de l'utilisateur est manquant."
	case status.USER_MISSING_EMAIL:
		return "L'email de l'utilisateur est manquant."
	case status.USER_INVALID_EMAIL:
		return "Format d'email utilisateur invalide."
	case status.USER_EMAIL_ALREADY_EXISTS:
		return "L'email existe déjà."
	case status.USER_MISSING_PHONE:
		return "Le numéro de téléphone de l'utilisateur est manquant."
	case status.USER_PHONE_ALREADY_EXISTS:
		return "Le numéro de téléphone existe déjà."
	case status.USER_INVALID_PHONE:
		return "Numéro de téléphone utilisateur invalide."
	case status.USER_INVALID_ROLE:
		return fmt.Sprintf("Rôle invalide. Les rôles valides sont : %v", args)
	case status.USER_INVALID_STATUS:
		return "Statut utilisateur invalide."

	// Device status messages
	case status.DEVICE_INVALID_PARAMS:
		return "Paramètres de l'appareil invalides."
	case status.DEVICE_MISSING_UUID:
		return "L'UUID de l'appareil est manquant."

	// OTP status messages
	case status.OTP_MISSING_PURPOSE:
		return "Le but de l'OTP est manquant."
	case status.OTP_INVALID_PURPOSE:
		return "But OTP invalide."
	case status.OTP_MISSING_IDENTIFIER:
		return "L'identifiant OTP est manquant."
	case status.OTP_MISSING_CODE:
		return "Le code OTP est manquant."
	case status.OTP_INVALID_CODE:
		return "Code OTP invalide."
	case status.OTP_STILL_ACTIVE:
		return "L'OTP est toujours actif."
	case status.OTP_EXCEED_MAX_SEND:
		return "Nombre maximum d'envois d'OTP dépassé."
	case status.OTP_EXCEED_MAX_VERIFY:
		return "Nombre maximum de vérifications d'OTP dépassé."
	case status.OTP_EXPIRED:
		return "L'OTP a expiré."
	case status.OTP_NOT_ALLOWED:
		return "Demande OTP non autorisée."
	case status.OTP_BLOCK_DEVICE:
		return "L'appareil est bloqué en raison de violations OTP."
	case status.OTP_BLOCK_DEVICE_PHONE:
		return "Le téléphone de l'appareil est bloqué en raison de violations OTP."
	case status.OTP_BLOCK_DEVICE_EMAIL:
		return "L'email de l'appareil est bloqué en raison de violations OTP."

	// Grade status messages
	case status.GRADE_INVALID_PARAMS:
		return "Paramètres de niveau invalides."
	case status.GRADE_MISSING_ID:
		return "L'ID du niveau est manquant."
	case status.GRADE_MISSING_LABEL:
		return "Le nom du niveau est manquant."
	case status.GRADE_NOT_FOUND:
		return "Niveau non trouvé."
	case status.GRADE_ALREADY_EXISTS:
		return "Le niveau existe déjà."
	case status.GRADE_CANNOT_DELETE:
		return "Le niveau ne peut pas être supprimé."

	// Level status messages
	case status.LEVEL_INVALID_PARAMS:
		return "Paramètres de niveau invalides."
	case status.LEVEL_MISSING_LABEL:
		return "Le nom du niveau est manquant."
	case status.LEVEL_NOT_FOUND:
		return "Niveau non trouvé."
	case status.LEVEL_ALREADY_EXISTS:
		return "Le niveau existe déjà."
	case status.LEVEL_CANNOT_DELETE:
		return "Le niveau ne peut pas être supprimé."

	// Profile status messages
	case status.PROFILE_INVALID_PARAMS:
		return "Paramètres de profil invalides."
	case status.PROFILE_MISSING_UID:
		return "L'ID utilisateur du profil est manquant."
	case status.PROFILE_MISSING_GRADE:
		return "Le niveau du profil est manquant."
	case status.PROFILE_MISSING_TERM:
		return "Le niveau du profil est manquant."
	case status.PROFILE_NOT_FOUND:
		return "Profil non trouvé."
	case status.PROFILE_ALREADY_EXISTS:
		return "Le profil existe déjà."
	case status.PROFILE_CANNOT_DELETE:
		return "Le profil ne peut pas être supprimé."
	case status.PROFILE_INVALID_GRADE:
		return "Niveau de profil invalide."
	case status.PROFILE_INVALID_LEVEL:
		return "Niveau de profil invalide."

	// Term status messages
	case status.TERM_INVALID_PARAMS:
		return "Paramètres de semestre invalides."
	case status.TERM_MISSING_NAME:
		return "Le nom du semestre est manquant."
	case status.TERM_NOT_FOUND:
		return "Semestre non trouvé."
	case status.TERM_ALREADY_EXISTS:
		return "Le semestre existe déjà."
	case status.TERM_CANNOT_DELETE:
		return "Le semestre ne peut pas être supprimé."

	default:
		return "Unknown"
	}
}
