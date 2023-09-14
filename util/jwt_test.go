package util

import (
	"fmt"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestJWT(t *testing.T) {
	Convey("Test JWT Access Generate", t, func() {
		token, err := GenerateTokenWithExp("xunop@qq.com-login", time.Minute*3)
		So(err, ShouldBeNil)
		fmt.Println("token:", token)
		So(token, ShouldNotBeEmpty)
		claims, err := ParseToken(token)
		So(err, ShouldBeNil)
		fmt.Println("claims:", claims)
		So(claims, ShouldNotBeEmpty)
		fmt.Println(claims.GetExpirationTime())
		username, _ := GetUsername(token, "login")
		So(username, ShouldEqual, "xunop@qq.com")
	})
}
