package util

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"testing"
)

func TestGenerateCode(_ *testing.T) {
	for i := 0; i < 10; i++ {
		code := GenerateCode()
		fmt.Println(code)
	}
}

func TestConnectToCOS(t *testing.T) {
	sID := os.Getenv("SECRETID")
	sKey := os.Getenv("SECRETKEY")
	if sID == "" || sKey == "" {
		t.Error()
	} else {
		print(fmt.Sprintf("SECRETID=%s,SECRETKEY=%s\n", sID, sKey))
	}
}

func TestGenerateRandomString(t *testing.T) {
	randomString, err := GenerateRandomString(32)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(randomString)
}

func TestSentMsgToBot(_ *testing.T) {
	message := `{"msg_type":"text","content":{"text":"测试"}}`
	print(message)
	// convert msg to json
	// resJSON, _ := json.Marshal(message)
	resJSON := []byte(`{
        "msg_type": "post",
        "content": {
                "post": {
                        "zh_cn": {
                                "title": "腾讯云COS头像审核失败",
                                "content": [
                                        [{
														"tag": "a",
														"text": "图片",
														"href": "https://console.cloud.tencent.com/cos/bucket"
												},		
												{
                                                        "tag": "text",
                                                        "text": "审核失败,后端开发人员:"
                                                },
                                                {
                                                        "tag": "a",
                                                        "text": "请查看",
                                                        "href": "https://console.cloud.tencent.com/cos/bucket"
                                                },
												{		
														"tag": "text",
														"text": "。错误信息：test" 
												}
                                        ]
                                ]
                        }
                }
        }
}`)
	// set request
	url := "https://open.feishu.cn/open-apis/bot/v2/hook/a569c93d-ef19-49ef-ba30-0b2ca73e4aa5"
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(resJSON))
	req.Header.Set("Content-Type", "application/json")

	// do request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()
}
