package service

import (
	"github.com/NJUPT-SAST/sast-link-backend/model"
	"github.com/NJUPT-SAST/sast-link-backend/model/result"
	"github.com/NJUPT-SAST/sast-link-backend/util"
	"github.com/gin-gonic/gin"
	"mime/multipart"
	"regexp"
	"strconv"
)

var cos = util.T_cos

func ChangeProfile(profile *model.Profile, uid string) error {
	// check org_id
	if profile.OrgId > 26 {
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
		serviceLogger.Errorln("CheckProfileByUid Err", err)
		return err
	}
	if resProfile == nil {
		serviceLogger.Infof("profile don`t exist")
		return result.ProfileNotExist
	}

	// update profile
	if err := model.UpdateProfile(resProfile, profile); err != nil {
		serviceLogger.Errorln("UpdateProfile Err", err)
		return err
	}
	return nil
}
func GetProfileInfo(uid string) (*model.Profile, error) {
	// verify if profile exist by uid(student ID)
	resProfile, err := model.SelectProfileByUid(uid)
	if err != nil {
		serviceLogger.Errorln("CheckProfileByUid Err", err)
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
		serviceLogger.Errorln("org_id input Err")
		return "", "", result.OrgIdError
	}

	//get dep and org
	if dep, org, err := model.GetDepAndOrgByOrgId(OrgId); err != nil {
		serviceLogger.Errorln("GetDepAndOrgByOrgId Err", err)
		return "", "", err
	} else {
		return dep, org, nil
	}
}

func UploadAvatar(avatar *multipart.FileHeader, uid string, ctx *gin.Context) error {
	//construct fileName
	userInfo, userInfoErr := model.UserInfo(uid)
	if userInfoErr != nil {
		serviceLogger.Errorln("user not exist", userInfoErr)
		return userInfoErr
	}
	fileName := strconv.Itoa(int(userInfo.ID))

	//get file stream
	fd, fileIOErr := avatar.Open()
	if fileIOErr != nil {
		serviceLogger.Errorln("get file stream err", fileIOErr)
		return fileIOErr
	}
	defer fd.Close()

	//upload to cos
	uploadKey := "/avatar/" + fileName + ".jpg"
	if _, cosUpErr := cos.Object.Put(ctx, uploadKey, fd, nil); cosUpErr != nil {
		serviceLogger.Errorln("upload avatar to cos fail", cosUpErr)
		return cosUpErr
	}

	//write to database, file url refer:tencent cos bucket file
	if dBUpErr := model.UpdateAvatar("https://sast-link-1309205610.cos.ap-shanghai.myqcloud.com"+uploadKey, userInfo.ID); dBUpErr != nil {
		//del cos file
		if _, cosDelErr := cos.Object.Delete(ctx, uploadKey); cosDelErr != nil {
			serviceLogger.Errorln("upload avatar to cos fail", cosDelErr)
			return cosDelErr
		}

		serviceLogger.Errorln("write file url to database Err", dBUpErr)
		return dBUpErr
	}

	return nil
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
			serviceLogger.Errorln("match hide field Err", matchErr)
			return matchErr
		} else if matched == false || i > len(hideFiledPattern) {
			serviceLogger.Infof("hide field illegal")
			return result.CheckHideIllegal
		}
	}
	return nil
}
