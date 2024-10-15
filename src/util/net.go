package util

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/NJUPT-SAST/sast-link-backend/log"
	"go.uber.org/zap"
)

// FIXME(aimisaka): Need to remove the log package reference, util package should not depend on log package.
func PostWithHeader(url string, header map[string]string, body any) (*http.Response, error) {
	jsonData, err := json.Marshal(body)
	if err != nil {
		log.Error("Failed to marshal json data", zap.Error(err))
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Error("Failed to create http request", zap.Error(err))
		return nil, err
	}
	for k, v := range header {
		req.Header.Set(k, v)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Error("Failed to send http request", zap.Error(err))
		return nil, err
	}

	return res, nil
}

func GetWithHeader(url string, header map[string]string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Error("Failed to create http request", zap.Error(err))
		return nil, err
	}
	for k, v := range header {
		req.Header.Set(k, v)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Error("Failed to send http request", zap.Error(err))
		return nil, err
	}

	return res, nil
}
