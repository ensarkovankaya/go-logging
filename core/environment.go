package core

import (
	"fmt"
	"os"
	"strconv"
)

func ParseBool(env string, dft bool, required bool) (bool, error) {
	if os.Getenv(env) == "" {
		if required {
			return false, fmt.Errorf("required environment variable not set: %s", env)
		}
		return dft, nil
	}
	return strconv.ParseBool(os.Getenv(env))
}

func ReadEnvironment() string {
	if os.Getenv("ENV") != "" {
		return os.Getenv("ENV")
	}
	return os.Getenv("ENVIRONMENT")
}
