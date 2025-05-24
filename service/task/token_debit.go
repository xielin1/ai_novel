package task

import (
	"encoding/json"
	"gin-template/model"
	"gin-template/service"
)

func CompensationTokenDebit(t model.CompensationTask) error {
	var payload struct {
		UserID            int64  `json:"user_id"`
		Amount            int64  `json:"amount"`
		TransactionUUID   string `json:"transaction_uuid"`
		TransactionType   string `json:"transaction_type"`
		Description       string `json:"description"`
		RelatedEntityType string `json:"related_entity_type"`
		RelatedEntityID   string `json:"related_entity_id"`
	}
	json.Unmarshal([]byte(t.Payload), &payload)

	_, err := service.GetTokenService().DebitToken(
		payload.UserID,
		payload.Amount,
		payload.TransactionUUID,
		payload.TransactionType,
		payload.Description,
		payload.RelatedEntityType,
		payload.RelatedEntityID,
	)
	return err
}
