package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"regexp"
	"strconv"

	"github.com/NJUPT-SAST/sast-link-backend/config"
	"github.com/NJUPT-SAST/sast-link-backend/plugin/storage/cos"
	"github.com/NJUPT-SAST/sast-link-backend/store"
	"github.com/NJUPT-SAST/sast-link-backend/util"
	"go.uber.org/zap"

	"github.com/pkg/errors"
)

type ProfileService struct {
	*BaseService
}

func NewProfileService(base *BaseService) *ProfileService {
	return &ProfileService{base}
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
	if profile.OrgID > 26 || profile.OrgID < -1 {
		s.log.Error("Orgenization id invalid", zap.Int("org_id", profile.OrgID))
		return errors.New("Orgenization id invalid")
	}

	// Check hide
	if matchErr := checkHideLegal(profile.Hide); matchErr != nil {
		return matchErr
	}

	// Verify if profile exist by uid(student ID)
	resProfile, err := s.Store.SelectProfileByUID(uid)
	if err != nil {
		s.log.Error("Failed to get profile", zap.String("uid", uid), zap.Error(err))
		return err
	}
	if resProfile == nil {
		s.log.Error("Profile not exist", zap.String("uid", uid))
		return errors.New("Profile not exist")
	}

	// Update profile
	if err := s.Store.UpdateProfile(resProfile, profile); err != nil {
		s.log.Error("Failed to update profile", zap.String("uid", uid), zap.Error(err))
		return err
	}
	return nil
}

func (s *ProfileService) GetProfileInfo(uid string) (*store.Profile, error) {
	// Verify if profile exist by uid(student ID)
	resProfile, err := s.Store.SelectProfileByUID(uid)
	if err != nil {
		s.log.Error("Failed to get profile", zap.String("uid", uid), zap.Error(err))
		return nil, err
	}
	if resProfile == nil {
		s.log.Error("Profile not exist", zap.String("uid", uid))
		return nil, errors.New("Profile not exist")
	}

	hideFiled := resProfile.Hide
	// Check hide
	if err := checkHideLegal(hideFiled); err != nil {
		s.log.Error("Failed to check hide field", zap.Error(err))
		return nil, err
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

func (s *ProfileService) GetProfileOrg(orgID int) (string, string, error) {
	// Check org_id
	if orgID > 26 {
		s.log.Error("Orgenization id is invalid", zap.Int("org_id", orgID))
		return "", "", errors.New("Orgenization id is invalid")
	} else if orgID == -1 || orgID == 0 {
		return "", "", nil
	}

	// Get dep and org
	dep, org, err := s.Store.GetDepAndOrgByOrgID(orgID)
	if err != nil {
		s.log.Error("Failed to get departmental and organizational information", zap.Error(err))
		return "", "", err
	}
	return dep, org, nil
}

func (s *ProfileService) UploadAvatar(ctx context.Context, avatar *multipart.FileHeader, uid string) (string, error) {
	// Construct fileName
	userInfo, err := s.Store.UserInfo(ctx, uid)
	if err != nil {
		s.log.Error("Failed to get user info", zap.String("uid", uid), zap.Error(err))
		return "", err
	}

	fileName := strconv.Itoa(int(userInfo.ID))
	// Get file stream
	fd, err := avatar.Open()
	if err != nil {
		s.log.Error("Failed to open file", zap.Error(err))
		return "", err
	}
	defer fd.Close()

	// Convert image to WebP
	fileBytes, err := io.ReadAll(fd)
	if err != nil {
		s.log.Error("Failed to read file", zap.Error(err))
		return "", err
	}
	webpBytes, err := util.ImageToWebp(fileBytes, 75)
	if err != nil {
		s.log.Error("Failed to convert image to WebP", zap.Error(err))
		return "", err
	}

	systemSetting, err := s.Store.GetSystemSetting(ctx, config.StorageSettingType.String())
	if err != nil {
		s.log.Error("Failed to get storage setting", zap.Error(err))
		return "", err
	}
	storageSetting := systemSetting.GetStorageSetting()
	if storageSetting == nil {
		s.log.Error("Failed to get storage setting", zap.Error(err))
		return "", errors.New("Get storage setting error")
	}

	client := cos.NewClient(storageSetting)
	uploadKey := fmt.Sprintf("avatar/%s.jpg", fileName)
	avatarURL, err := client.UploadObject(ctx, uploadKey, bytes.NewReader(webpBytes))
	if err != nil {
		s.log.Error("Failed to upload file", zap.String("key", uploadKey), zap.Error(err))
		// If upload file fail, delete file in cos
		if err := client.DeleteObject(ctx, uploadKey); err != nil {
			// If delete file in cos fail, log error and return
			s.log.Error("Failed to delete file", zap.String("key", uploadKey), zap.Error(err))
			return "", errors.Wrap(err, "failed to delete file when upload fail")
		}

		return "", err
	}

	// Write to database, file url refer:tencent cos bucket file
	if err := s.Store.UpdateAvatar(avatarURL, userInfo.ID); err != nil {
		// If update avatar to database fail, delete file in cos
		if err := client.DeleteObject(ctx, uploadKey); err != nil {
			s.log.Error("Failed to delete file", zap.String("key", uploadKey), zap.Error(err))
			return "", errors.Wrap(err, "failed to delete file when update avatar fail")
		}

		s.log.Error("Failed to update avatar to database", zap.String("avatar_url", avatarURL), zap.Error(err))
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

	// matching declared Filed, and if hide > declared Filed, match fail
	for i := range hide {
		matched, err := regexp.MatchString(hideFiledPattern[0]+"|"+hideFiledPattern[1]+"|"+hideFiledPattern[2], hide[i])

		if err != nil {
			return errors.Wrap(err, "failed to match hide field")
		} else if !matched || i > len(hideFiledPattern) {
			return errors.New("Hide field invalid")
		}
	}
	return nil
}

func (s *ProfileService) SentMsgToBot(ctx context.Context, checkRes *store.CheckRes) error {
	var message []byte
	if checkRes.Code != 0 {
		message = []byte(fmt.Sprintf(reviewFailMsg, checkRes.Data.URL, checkRes.Message))
	} else if checkRes.Data.Result == 2 {
		message = []byte(fmt.Sprintf(picSensitiveMsg, checkRes.Data.URL))
	} else {
		return nil
	}

	systemSetting, err := s.Store.GetSystemSetting(ctx, config.WebsiteSettingType.String())
	if err != nil {
		s.log.Error("Failed to get system setting", zap.Error(err))
		return errors.Wrap(err, "failed to get system setting")
	}
	webSetting := systemSetting.GetWebsiteSetting()
	if webSetting == nil {
		return errors.New("Get website setting error")
	}

	// Set request
	url := webSetting.ImageReviewWebhook
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(message))
	if err != nil {
		s.log.Error("Failed to send message to feishu bot", zap.Error(err))
		return errors.Wrap(err, "failed to send message to feishu bot")
	}
	req.Header.Set("Content-Type", "application/json")

	// Do request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		s.log.Error("Failed to send message to feishu bot", zap.Error(err))
		return errors.Wrap(err, "failed to send message to feishu bot")
	}
	defer resp.Body.Close()
	return nil
}

func (s *ProfileService) DealWithFrozenImage(ctx context.Context, checkRes *store.CheckRes) error {
	compileRegex := regexp.MustCompile("[0-9]+")
	matchArr := compileRegex.FindAllString(checkRes.Data.URL, -1)
	if matchArr == nil {
		return errors.New("Failed to match image url")
	}
	userID := matchArr[1]

	systemSetting, err := s.Store.ListSystemSetting(ctx)
	if err != nil {
		s.log.Error("Failed to get system setting", zap.Error(err))
		return errors.Wrap(err, "failed to get system setting")
	}
	orStoragesetting := systemSetting[config.StorageSettingType.String()]
	storageSetting := orStoragesetting.GetStorageSetting()
	if storageSetting == nil {
		return errors.New("Get storage setting error")
	}

	client := cos.NewClient(storageSetting)
	// Mv image
	source := "avatar/" + userID + ".jpg"
	dest := "ban/" + userID + ".jpg"
	if err := client.MoveObject(ctx, source, dest); err != nil {
		s.log.Error("Failed to move file to ban", zap.String("source", source), zap.String("dest", dest), zap.Error(err))
		return errors.Wrap(err, "failed to move file to ban")
	}
	if err := client.DeleteObject(ctx, source); err != nil {
		s.log.Error("Failed to delete file", zap.String("source", source), zap.Error(err))
		return errors.Wrap(err, "failed to delete file")
	}

	orSiteSetting := systemSetting[config.WebsiteSettingType.String()]
	siteSetting := orSiteSetting.GetWebsiteSetting()
	if siteSetting == nil {
		return errors.New("Get website setting error")
	}
	avatarURL := siteSetting.AvatarErrorURLImage
	parseUint, _ := strconv.Atoi(userID)
	if err := s.Store.UpdateAvatar(avatarURL, uint(parseUint)); err != nil {
		s.log.Error("Failed to update avatar", zap.String("avatar_url", avatarURL), zap.Error(err))
		return errors.Wrap(err, "failed to update avatar")
	}
	return nil
}

// GetBindList get bind list by uid
func (s *ProfileService) GetBindList(ctx context.Context, uid string) ([]string, error) {
	binds, err := s.Store.GetOauthBindStatusByUID(ctx, uid)
	if err != nil {
		s.log.Error("Failed to get bind list", zap.String("uid", uid), zap.Error(err))
		return nil, err
	}
	return binds, nil
}
