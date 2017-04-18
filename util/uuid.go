package util

import (
	"github.com/nu7hatch/gouuid"
)

func GenerateUUID() string {
	u, err := uuid.NewV4()
	if err != nil {
		panic(err)
	}
	return u.String()
}