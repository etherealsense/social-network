package env

import (
	"log"
	"os"
	"strconv"
)

func GetString(key string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		log.Panicf("env var %s is not set", key)
	}

	return val
}

func GetInt(key string) int {
	val, ok := os.LookupEnv(key)
	if !ok {
		log.Panicf("env var %s is not set", key)
	}

	valAsInt, err := strconv.Atoi(val)
	if err != nil {
		log.Panicf("env var %s is not a valid integer: %v", key, err)
	}

	return valAsInt
}
