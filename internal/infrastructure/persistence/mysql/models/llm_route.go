package models

type LLMRouteModel struct {
	Audit

	LLMRouteId     int64
	RouteCode      string
	LLMProvider    *string
	LLMModel       *string
	Description    *string
	LLMRouteStatus *string
}
