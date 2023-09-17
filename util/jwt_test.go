package util

import (
	"fmt"
	"github.com/NJUPT-SAST/sast-link-backend/model"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestJWT(t *testing.T) {
	Convey("Test JWT Access Generate", t, func() {
		token, err := GenerateTokenWithExp("xunop@qq.com-login", model.LOGIN_TOKEN_SUB, 0, time.Minute*3)
		So(err, ShouldBeNil)
		fmt.Println("token:", token)
		So(token, ShouldNotBeEmpty)
		claims, err := ParseToken(token)
		So(err, ShouldBeNil)
		fmt.Println("claims:", claims)
		So(claims, ShouldNotBeEmpty)
		fmt.Println(claims.GetExpirationTime())
		username, _ := GetUsername(token)
		So(username, ShouldEqual, "xunop@qq.com")
	})
}
