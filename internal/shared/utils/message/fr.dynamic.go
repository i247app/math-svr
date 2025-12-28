package message

import "math-ai.com/math-ai/internal/shared/constant/status"

// GetMessageTemplateFR returns French message templates with placeholders for dynamic arguments
// If a template exists for the status code, it returns the template; otherwise returns empty string
// Use this when you have dynamic arguments to interpolate
func GetMessageTemplateFR(statusCode status.Code) string {
	switch statusCode {
	// User templates with dynamic arguments
	case status.USER_EMAIL_ALREADY_EXISTS:
		return "L'email {email} existe déjà."
	case status.USER_PHONE_ALREADY_EXISTS:
		return "Le numéro de téléphone {phone} existe déjà."
	case status.USER_NOT_FOUND:
		return "Utilisateur avec ID {user_id} introuvable."

	// Grade templates
	case status.GRADE_ALREADY_EXISTS:
		return "La classe '{label}' existe déjà."
	case status.GRADE_NOT_FOUND:
		return "Classe '{label}' introuvable."

	// Semester templates
	case status.SEMESTER_ALREADY_EXISTS:
		return "Le semestre '{name}' existe déjà."
	case status.SEMESTER_NOT_FOUND:
		return "Semestre '{name}' introuvable."

	// Level templates
	case status.LEVEL_ALREADY_EXISTS:
		return "Le niveau '{label}' existe déjà."
	case status.LEVEL_NOT_FOUND:
		return "Niveau '{label}' introuvable."

	// Contact templates
	case status.CONTACT_NAME_TOO_LONG:
		return "Le nom du contact est trop long (max {max_length} caractères)."
	case status.CONTACT_MESSAGE_TOO_LONG:
		return "Le message du contact est trop long (max {max_length} caractères)."
	case status.CONTACT_EMAIL_TOO_LONG:
		return "L'email du contact est trop long (max {max_length} caractères)."

	default:
		return "" // No template available, use static message
	}
}
