package entity

import (
	"encoding/json"

	"github.com/gofrs/uuid"
	"github.com/golang-jwt/jwt/v5"
	log "github.com/sirupsen/logrus"
)

type JwtClaims struct {
	jwt.RegisteredClaims

	UserID uuid.UUID
}

type User struct {
	ID          uuid.UUID
	Login       string
	PasswordSHA []byte
}

type Counter struct {
	ID    uuid.UUID
	Name  string
	Value int64
}

type Gauge struct {
	ID    uuid.UUID
	Name  string
	Value float64
}

type Gophermart struct {
	ID    string   `json:"id"`              // Имя метрики
	MType string   `json:"type"`            // Параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // Значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // Значение метрики в случае передачи gauge
}

func (m Gophermart) String() string {
	str, err := json.Marshal(m)
	if err != nil {
		log.Errorf("unable to marshal Metric: %v", err)
	}

	return string(str)
}
