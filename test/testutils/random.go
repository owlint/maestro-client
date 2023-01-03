package testutils

import "github.com/google/uuid"

func RandomStr() string {
	return uuid.NewString()
}
