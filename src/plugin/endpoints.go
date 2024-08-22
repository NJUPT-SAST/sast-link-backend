package plugin

import (
	"golang.org/x/oauth2"
)

// GitHub is the endpoint for Github.
var GitHub = oauth2.Endpoint{
	AuthURL:  "https://github.com/login/oauth/authorize",
	TokenURL: "https://github.com/login/oauth/access_token",
}

// Lark is the endpoint for Feishu
var Lark = oauth2.Endpoint{
	AuthURL:  "https://open.feishu.cn/open-apis/authen/v1/authorize",
	TokenURL: "https://open.feishu.cn/open-apis/authen/v1/oidc/access_token",
}

// QQ is the endpoint for QQ
var QQ = oauth2.Endpoint{
	AuthURL:  "https://graph.qq.com/oauth2.0/authorize",
	TokenURL: "https://graph.qq.com/oauth2.0/token",
}
