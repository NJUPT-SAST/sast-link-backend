package util

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/NJUPT-SAST/sast-link-backend/log"
)

// FIXME(aimisaka): Need to remove the log package reference, util package should not depend on log package.
func PostWithHeader(url string, header map[string]string, body any) (*http.Response, error) {
	jsonData, err := json.Marshal(body)
	if err != nil {
		log.Log.Errorln("json.Marshal ::: ", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Log.Errorln("http.NewRequest ::: ", err)
		return nil, err
	}
	for k, v := range header {
		req.Header.Set(k, v)
	}

	// DEBUG
	log.LogReq(req)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Log.Errorln("http.DefaultClient.Do ::: ", err)
		return nil, err
	}

	// DEBUG
	log.LogRes(res)
	return res, nil
}

func GetWithHeader(url string, header map[string]string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Log.Errorln("http.NewRequest ::: ", err)
		return nil, err
	}
	for k, v := range header {
		req.Header.Set(k, v)
	}

	// DEBUG
	log.LogReq(req)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Log.Errorln("http.DefaultClient.Do ::: ", err)
		return nil, err
	}

	// DEBUG
	log.LogRes(res)
	return res, nil
}
