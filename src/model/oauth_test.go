package model

import (
	"database/sql"
	"log"
	"testing"

	"gorm.io/datatypes"
)

func TestInsertOAuth2Info(t *testing.T) {
	db, err := sql.Open("postgres", "your-database-connection-string")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	info := `{"login":"user","id":123,"node_id":"node_id","avatar_url":"url","gravatar_id":"","url":"url","html_url":"url","followers_url":"url","following_url":"url","gists_url":"url","starred_url":"url","subscriptions_url":"https://api.github.com/users/user/subscriptions","organizations_url":"https://api.github.com/users/user/orgs","repos_url":"https://api.github.com/users/user/repos","events_url":"https://api.github.com/users/user/events{/privacy}","received_events_url":"https://api.github.com/users/user/received_events","type":"User","site_admin":false,"name":"max","company":"@Xiaomi @sast @GreenPomelo","blog":"blog_url","location":"Nanjing","email":null,"hireable":null,"bio":"From Nanjing University of Posts and Telecommunications, member of @SAST. Working on 👀 visualization & 🌐 web dev\\r\\n","twitter_username":null,"public_repos":48,"public_gists":3,"followers":65,"following":148,"created_at":"2020-02-07T09:21:13Z","updated_at":"2024-07-19T12:26:27Z"}`
	testData := OAuth2Info{
		Client:  "github",
		Info:    datatypes.JSON(info),
		OauthID: "60775796",
		UserID:  "b21041313",
	}

	UpsetOauthInfo(testData)
}
