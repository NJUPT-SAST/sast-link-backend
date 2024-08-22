package service

import (
	"bytes"
	"context"
	"fmt"
	"mime/multipart"
	"net/http"
	"regexp"
	"strconv"

	"github.com/NJUPT-SAST/sast-link-backend/http/response"
	"github.com/NJUPT-SAST/sast-link-backend/log"
	"github.com/NJUPT-SAST/sast-link-backend/plugin/storage/cos"
	"github.com/NJUPT-SAST/sast-link-backend/store"
)

type ProfileService struct {
	*BaseService
}

func NewProfileService(store *BaseService) *ProfileService {
	return &ProfileService{store}
}

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

func (s *ProfileService) ChangeProfile(profile *store.Profile, uid string) error {
	// Check org_id
	if profile.OrgId > 26 || profile.OrgId < -1 {
		log.Errorf("OrgId illegal: %d", profile.OrgId)
		return response.OrgIdError
	}

	// Check hide
	if matchErr := checkHideLegal(profile.Hide); matchErr != nil {
		return matchErr
	}

	// Verify if profile exist by uid(student ID)
	resProfile, err := s.Store.SelectProfileByUid(uid)
	if err != nil {
		log.Errorf("CheckProfile error: %s", err.Error())
		return err
	}
	if resProfile == nil {
		log.Errorf("Profile [%s] Not Exist\n", uid)
		return response.ProfileNotExist
	}

	// Update profile
	if err := s.Store.UpdateProfile(resProfile, profile); err != nil {
		log.Errorf("UpdateProfile error: %s", err.Error())
		return err
	}
	return nil
}

func (s *ProfileService) GetProfileInfo(uid string) (*store.Profile, error) {
	// Verify if profile exist by uid(student ID)
	resProfile, err := s.Store.SelectProfileByUid(uid)
	if err != nil {
		log.Errorf("CheckProfile error: %s", err.Error())
		return nil, err
	}
	if resProfile == nil {
		log.Errorf("Profile [%s] Not Exist\n", uid)
		return nil, response.ProfileNotExist
	}

	hideFiled := resProfile.Hide
	// Check hide
	if matchErr := checkHideLegal(hideFiled); matchErr != nil {
		log.Errorf("Hide field illegal")
		return nil, matchErr
	}
	// Hide filed
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

func (s *ProfileService) GetProfileOrg(OrgId int) (string, string, error) {
	// Check org_id
	if OrgId > 26 {
		log.Errorf("OrgId illegal: %d", OrgId)
		return "", "", response.OrgIdError
	} else if OrgId == -1 || OrgId == 0 {
		return "", "", nil
	} else {
		// Get dep and org
		if dep, org, err := s.Store.GetDepAndOrgByOrgId(OrgId); err != nil {
			log.Errorf("GetDepAndOrgByOrgId error: %s", err.Error())
			return "", "", err
		} else {
			return dep, org, nil
		}
	}
}

func (s *ProfileService) UploadAvatar(avatar *multipart.FileHeader, uid string, ctx context.Context) (string, error) {
	// Construct fileName
	userInfo, userInfoErr := s.Store.UserInfo(uid)
	if userInfoErr != nil {
		log.Errorf("User not exist,ErrMsg:", userInfoErr)
		return "", userInfoErr
	}

	fileName := strconv.Itoa(int(userInfo.ID))
	// Get file stream
	fd, fileIOErr := avatar.Open()
	if fileIOErr != nil {
		log.Errorf("Open file fail,ErrMsg:", fileIOErr)
		return "", fileIOErr
	}
	defer fd.Close()

	systemSetting, err := s.Store.GetSystemSetting(ctx, store.StorageSettingType)
	if err != nil {
		log.Errorf("Get system setting error: %s", err.Error())
		return "", err
	}
	storageSetting, err := systemSetting.GetStorageSetting()
	if err != nil {
		log.Errorf("Get storage setting error: %s", err.Error())
		return "", err
	}

	client := cos.NewClient(storageSetting)
	uploadKey := fmt.Sprintf("avatar/%s.jpg", fileName)
	avatarURL, err := client.UploadObject(ctx, uploadKey, fd)
	if err != nil {
		log.Errorf("Upload file %s fail: %s", uploadKey, err.Error())
		client.DeleteObject(ctx, uploadKey)
		return "", err
	}

	// Write to database, file url refer:tencent cos bucket file
	if err := s.Store.UpdateAvatar(avatarURL, userInfo.ID); err != nil {
		// If update avatar to database fail, delete file in cos
		client.DeleteObject(ctx, uploadKey)

		log.Errorf("Update avatar to database fail: %s, %s", avatarURL, err.Error())
		return "", err
	}

	return avatarURL, nil
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
			log.Errorf("Hide field match fail,ErrMsg:", matchErr)
			return matchErr
		} else if matched == false || i > len(hideFiledPattern) {
			log.Errorf("Hide field illegal")
			return response.CheckHideIllegal
		}
	}
	return nil
}

func (s *ProfileService) SentMsgToBot(checkRes *store.CheckRes) error {
	var message []byte
	if checkRes.Code != 0 {
		message = []byte(fmt.Sprintf(reviewFailMsg, checkRes.Data.Url, checkRes.Message))
	} else if checkRes.Data.Result == 2 {
		message = []byte(fmt.Sprintf(picSensitiveMsg, checkRes.Data.Url))
	} else {
		return nil
	}

	// Set request
	url := "https://open.feishu.cn/open-apis/bot/v2/hook/a569c93d-ef19-49ef-ba30-0b2ca73e4aa5"
	req, reqErr := http.NewRequest("POST", url, bytes.NewBuffer(message))
	req.Header.Set("Content-Type", "application/json")

	// Do request
	client := &http.Client{}
	resp, reqErr := client.Do(req)
	if reqErr != nil {
		log.Errorf("Sent msg to feishu bot fail,ErrMsg:", reqErr)
		return reqErr
	}
	defer resp.Body.Close()
	return nil
}

func (s *ProfileService) DealWithFrozenImage(ctx context.Context, checkRes *store.CheckRes) error {
	compileRegex := regexp.MustCompile("[0-9]+")
	matchArr := compileRegex.FindAllString(checkRes.Data.Url, -1)
	if matchArr == nil {
		return response.PicURLErr
	}
	userId := matchArr[1]

	systemSetting, err := s.Store.ListSystemSetting(ctx)
	if err != nil {
		log.Errorf("Get system setting error: %s", err.Error())
		return err
	}
	storageSetting, err := systemSetting[store.StorageSettingType].(*store.SystemSetting).GetStorageSetting()
	if err != nil {
		log.Errorf("Get storage setting error: %s", err.Error())
		return err
	}

	client := cos.NewClient(storageSetting)
	// Mv image
	source := "avatar/" + userId + ".jpg"
	dest := "ban/" + userId + ".jpg"
	if err := client.MoveObject(ctx, source, dest); err != nil {
		log.Errorf("Move file %s to %s failed: %s", source, dest, err)
		return err
	}
	if err := client.DeleteObject(ctx, source); err != nil {
		log.Errorf("Delete file %s failed: %s", source, err.Error())
		return err
	}

	// Replace image to Err image
	// TODO: Get error image url from system setting

	siteSetting, err := systemSetting[store.WebsiteSettingType].(*store.SystemSetting).GetWebsiteSetting()
	if err != nil {
		log.Errorf("Get storage setting error: %s", err.Error())
		return err
	}
	avatarURL := siteSetting.AvatarErrorURLImage
	parseUint, _ := strconv.Atoi(userId)
	if upErr := s.Store.UpdateAvatar(avatarURL, uint(parseUint)); upErr != nil {
		log.Errorf("Update avatar failed: %s", upErr)
		return upErr
	}
	return nil
}
