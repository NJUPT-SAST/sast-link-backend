package service

import (
	"bytes"
	"fmt"
	"github.com/NJUPT-SAST/sast-link-backend/model"
	"github.com/NJUPT-SAST/sast-link-backend/model/result"
	"github.com/NJUPT-SAST/sast-link-backend/util"
	"github.com/gin-gonic/gin"
	"mime/multipart"
	"net/http"
	"regexp"
	"strconv"
)

var cos = util.T_cos

const reviewFailMsg = `{
        "msg_type": "post",
        "content": {
                "post": {
                        "zh_cn": {
                                "title": "腾讯云COS头像审核失败",
                                "content": [
                                        [{
														"tag": "a",
														"text": "图片",
														"href": "%s"
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
														"text": "。错误信息：%s" 
												}
                                        ]
                                ]
                        }
                }
        }
}`
const picSensitiveMsg = `{
        "msg_type": "post",
        "content": {
                "post": {
                        "zh_cn": {
                                "title": "头像审核通知",
                                "content": [
                                        [
												{
														"tag": "a",
														"text": "疑似敏感文件",
														"href": "%s"
												},
												{
                                                        "tag": "text",
                                                        "text": "，建议人工复审/更改审核库: "
                                                },
                                                {
                                                        "tag": "a",
                                                        "text": "请查看",
                                                        "href": "https://console.cloud.tencent.com/cos/bucket"
                                                }
                                        ]
                                ]
                        }
                }
        }
}`

func ChangeProfile(profile *model.Profile, uid string) error {
	// check org_id
	if profile.OrgId > 26 || profile.OrgId < -1 {
		serviceLogger.Infof("org_id input Err")
		return result.OrgIdError
	}

	// check hide
	if matchErr := checkHideLegal(profile.Hide); matchErr != nil {
		serviceLogger.Infof("hide field illegal")
		return matchErr
	}

	// verify if profile exist by uid(student ID)
	resProfile, err := model.SelectProfileByUid(uid)
	if err != nil {
		serviceLogger.Errorln("CheckProfileByUid Err,ErrMsg:", err)
		return err
	}
	if resProfile == nil {
		serviceLogger.Infof("profile don`t exist")
		return result.ProfileNotExist
	}

	// update profile
	if err := model.UpdateProfile(resProfile, profile); err != nil {
		serviceLogger.Errorln("UpdateProfile Err,ErrMsg:", err)
		return err
	}
	return nil
}
func GetProfileInfo(uid string) (*model.Profile, error) {
	// verify if profile exist by uid(student ID)
	resProfile, err := model.SelectProfileByUid(uid)
	if err != nil {
		serviceLogger.Errorln("CheckProfileByUid Err,ErrMsg:", err)
		return nil, err
	}
	if resProfile == nil {
		serviceLogger.Infof("profile don`t exist")
		return nil, result.ProfileNotExist
	}

	hideFiled := resProfile.Hide
	// check hide
	if matchErr := checkHideLegal(hideFiled); matchErr != nil {
		serviceLogger.Infof("hide field illegal")
		return nil, matchErr
	}
	// hide filed
	for i := range hideFiled {
		switch hideFiled[i] {
		case "bio":
			resProfile.Bio = nil
		case "link":
			resProfile.Link = nil
		case "badge":
			resProfile.Badge = nil
		}
	}
	return resProfile, nil

}
func GetProfileOrg(OrgId int) (string, string, error) {
	// check org_id
	if OrgId > 26 {
		serviceLogger.Errorln("org_id input Err,ErrMsg:")
		return "", "", result.OrgIdError
	} else if OrgId == -1 || OrgId == 0 {
		return "", "", nil
	} else {
		//get dep and org
		if dep, org, err := model.GetDepAndOrgByOrgId(OrgId); err != nil {
			serviceLogger.Errorln("GetDepAndOrgByOrgId Err,ErrMsg:", err)
			return "", "", err
		} else {
			return dep, org, nil
		}
	}
}

func UploadAvatar(avatar *multipart.FileHeader, uid string, ctx *gin.Context) (string, error) {
	//construct fileName
	userInfo, userInfoErr := model.UserInfo(uid)
	if userInfoErr != nil {
		serviceLogger.Errorln("user not exist,ErrMsg:", userInfoErr)
		return "", userInfoErr
	}
	fileName := strconv.Itoa(int(userInfo.ID))

	//get file stream
	fd, fileIOErr := avatar.Open()
	if fileIOErr != nil {
		serviceLogger.Errorln("get file stream err,ErrMsg:", fileIOErr)
		return "", fileIOErr
	}
	defer fd.Close()

	//upload to cos
	uploadKey := "avatar/" + fileName + ".jpg"
	if _, cosUpErr := cos.Object.Put(ctx, uploadKey, fd, nil); cosUpErr != nil {
		serviceLogger.Errorln("upload avatar to cos fail,ErrMsg:", cosUpErr)
		return "", cosUpErr
	}

	//write to database, file url refer:tencent cos bucket file
	if dBUpErr := model.UpdateAvatar("https://sast-link-1309205610.cos.ap-shanghai.myqcloud.com/"+uploadKey, userInfo.ID); dBUpErr != nil {
		//del cos file
		if _, cosDelErr := cos.Object.Delete(ctx, uploadKey); cosDelErr != nil {
			serviceLogger.Errorln("upload avatar to cos fail,ErrMsg:", cosDelErr)
			return "", cosDelErr
		}

		serviceLogger.Errorln("write file url to database Err,ErrMsg:", dBUpErr)
		return "", dBUpErr
	}

	return "https://sast-link-1309205610.cos.ap-shanghai.myqcloud.com/" + uploadKey, nil
}

func checkHideLegal(hide []string) error {
	// declare allow hide field: bio,link,badge
	var hideFiledPattern = []string{
		"bio",
		"link",
		"badge",
	}

	//matching declared Filed, and if hide > declared Filed, match fail
	for i := range hide {
		matched, matchErr := regexp.MatchString(hideFiledPattern[0]+"|"+hideFiledPattern[1]+"|"+hideFiledPattern[2], hide[i])

		if matchErr != nil {
			serviceLogger.Errorln("match hide field Err,ErrMsg:", matchErr)
			return matchErr
		} else if matched == false || i > len(hideFiledPattern) {
			serviceLogger.Infof("hide field illegal")
			return result.CheckHideIllegal
		}
	}
	return nil
}

func SentMsgToBot(checkRes *model.CheckRes) error {
	var message []byte
	if checkRes.Code != 0 {
		message = []byte(fmt.Sprintf(reviewFailMsg, checkRes.Data.Url, checkRes.Message))
	} else if checkRes.Data.Result == 2 {
		message = []byte(fmt.Sprintf(picSensitiveMsg, checkRes.Data.Url))
	} else {
		return nil
	}

	// set request
	url := "https://open.feishu.cn/open-apis/bot/v2/hook/a569c93d-ef19-49ef-ba30-0b2ca73e4aa5"
	req, reqErr := http.NewRequest("POST", url, bytes.NewBuffer(message))
	req.Header.Set("Content-Type", "application/json")

	// do request
	client := &http.Client{}
	resp, reqErr := client.Do(req)
	if reqErr != nil {
		serviceLogger.Errorln("Sent message to group bot fail,ErrMsg:", reqErr)
		return reqErr
	}
	defer resp.Body.Close()
	return nil
}

func DealWithFrozenImage(ctx *gin.Context, checkRes *model.CheckRes) error {
	compileRegex := regexp.MustCompile("[0-9]+")
	matchArr := compileRegex.FindAllString(checkRes.Data.Url, -1)
	if matchArr == nil {
		return result.PicURLErr
	}
	userId := matchArr[1]
	source := "avatar/" + userId + ".jpg"
	//mv image
	sourceURL := fmt.Sprintf("sast-link-1309205610.cos.ap-shanghai.myqcloud.com/%s", source)
	dest := "ban/" + userId + ".jpg"
	_, _, err := cos.Object.Copy(ctx, dest, sourceURL, nil)
	if err == nil {
		_, err := cos.Object.Delete(ctx, source, nil)
		if err != nil {
			serviceLogger.Errorln("del file fail,ErrMsg:", err)
			return err
		}
	} else {
		serviceLogger.Errorln("cp file fail,ErrMsg:", err)
	}

	//replace image to Err image
	avatarURL := "https://sast-link-1309205610.cos.ap-shanghai.myqcloud.com/ErrorImage.jpg"
	parseUint, _ := strconv.Atoi(userId)
	if upErr := model.UpdateAvatar(avatarURL, uint(parseUint)); upErr != nil {
		serviceLogger.Errorln("updateAvatar fail,ErrMsg:", upErr)
		return upErr
	}
	return nil
}

// GetBindList get bind list by uid
func GetBindList(uid string) ([]string, error) {
	binds, err := model.GetOauthBindStatusByUID(uid)
	if err != nil {
		serviceLogger.Errorln("GetBindList Err,ErrMsg:", err)
		return nil, err
	}
	return binds, nil
}
