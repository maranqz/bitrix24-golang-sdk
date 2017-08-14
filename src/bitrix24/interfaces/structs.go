package bitrix24

import (
	"github.com/antonholmquist/jason"
	"strings"
)

type Scope []string

func (t *Scope) String() string {
	return strings.Join(*t, ",")
}

type SettingsInterface struct {
	Domain            string // domain bitrix24 application
	ApplicationSecret string // secret code application [0-9A-z]{50} "client_secret"
	ApplicationId     string //application identity, (app|local).[0-9a-z]{14,14}.[0-9]{8} "client_id"

	AccessToken  string //token for access, [0-9a-z]{32}
	RefreshToken string //token for refresh token access, [0-9a-z]{32}
	MemberId     string //the unique Bitrix24 portal ID

	/*
	permissions connection
	calendar, crm, disk, department, entity, im, imbot, lists, log,
	mailservice, sonet_group, task, tasks_extended, telephony, call, user,
	imopenlines, placement
	*/
	ApplicationScope string
	RedirectUri      string //url for redirect after authorization

	timeout int //timeout before disconnect

	Log Logger
}

func GetSettingsByJson(data *jason.Object) *SettingsInterface {
	memberId, _ := data.GetString("member_id")
	accessToken, _ := data.GetString("access_token")
	refreshToken, _ := data.GetString("refresh_token")
	applicationScope, _ := data.GetString("scope")

	return &SettingsInterface{
		MemberId:         memberId,
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		ApplicationScope: applicationScope,
	}
}
