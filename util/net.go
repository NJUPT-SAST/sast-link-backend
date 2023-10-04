package util

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/NJUPT-SAST/sast-link-backend/log"
)

func PostWithHeader(url string, header map[string]string, body any) (*http.Response, error) {
	jsonData, err := json.Marshal(body);
	if err != nil {
		log.Logger.Errorln("json.Marshal ::: ", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Logger.Errorln("http.NewRequest ::: ", err)
		return nil, err
	}
	for k, v := range header {
		req.Header.Set(k, v)
	}

	// DEBUG
	log.LogReq(req)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Logger.Errorln("http.DefaultClient.Do ::: ", err)
		return nil, err
	}

	// DEBUG
	log.LogRes(res)
	return res, nil
}
