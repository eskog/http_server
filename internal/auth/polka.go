package auth

import (
	"errors"
	"net/http"
	"strings"
)

func GetAPIKey(headers http.Header) (string, error) {
	headerString := headers.Get("Authorization")
	apikey, ok := strings.CutPrefix(headerString, "ApiKey ")
	if !ok {
		return "", errors.New("unable to find apykey")
	}
	return apikey, nil
}
