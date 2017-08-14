package bitrix24

import (
	. "bitrix24/interfaces"
	"reflect"
	"github.com/fatih/structs"
	"github.com/parnurzeal/gorequest"
	"net/url"
	"time"
	"net/http"
	"github.com/antonholmquist/jason"
	"strings"
	"errors"
)

const (
	PROTOCOL     = "https://"
	OAUTH_SERVER = "oauth.bitrix.info"
	OAUTH_TOKEN  = "/oauth/token/"
	AUTH_URL     = "/oauth/authorize/"
)

var request = gorequest.New()
var realationNames = map[string]string{
	"domain":            "domain",
	"applicationSecret": "client_secret",
	"applicationId":     "client_id",

	"accessToken":  "access_token",
	"refreshToken": "refresh_token",
	"memberId":     "member_id",

	"applicationScope": "scope",
	"redirectUri":      "redirect_uri",
}

//Consist data for authorization
type Bitrix24 struct {
	isAccessParams bool //Specifies that all access settings are set

	domain            string // domain bitrix24 application
	applicationSecret string // secret code application [0-9A-z]{50,50} "client_secret"
	applicationId     string //application identity, (app|local).[0-9a-z]{14,14}.[0-9]{8,8} "client_id"

	accessToken  string //token for access, [0-9a-z]{32}
	refreshToken string //token for refresh token access
	memberId     string //the unique Bitrix24 portal ID

	applicationScope string //array permissions connection
	redirectUri      string //url for redirect after authorization

	//timeout before disconnect (trying to connect + receiving a response)
	//https://github.com/parnurzeal/gorequest/blob/develop/gorequest.go#L452
	timeout int

	log Logger

	request gorequest.SuperAgent
}

//Constructor Bitrix24
func (t *Bitrix24) Init(domain string, applicationSecret string, applicationId string, logger Logger) {

	if logger != nil {
		t.log = logger
	} else {
		//t.log := NullLogger{};
	}

	t.timeout = 30

	//t.request = *gorequest.New()

	t.SetDomain(domain)
	t.SetApplicationSecret(applicationSecret)
	t.SetApplicationId(applicationId)
}

//Set settings attributes
func (t *Bitrix24) SetAttributes(attributes SettingsInterface) {
	tReflect := reflect.ValueOf(&t)

	if tReflect.Kind() == reflect.Ptr {
		tReflect = tReflect.Elem()
	}

	settings := structs.Map(&attributes)

	for key, value := range settings {
		if value == nil || value == "" {
			continue
		}

		attribute := tReflect.MethodByName("Set" + key)

		if attribute.IsValid() {
			attribute.Call([]reflect.Value{reflect.ValueOf(value)})
		} else {
			panic(key + " not exitst in " + tReflect.Type().Name())
		}
	}

	t.CheckAccessParams()
}

//Url to request authorization from the user
func (t *Bitrix24) GetUrlClientAuth(params *url.Values) string {
	t.generateParams(params, "applicationId", "applicationScope")
	params.Set("response_type", "code")

	return t.GetUrlAuth("", params)
}

//Use with the received code after request by getUrlClientAuth
func (t *Bitrix24) GetFirstAccessToken(params *url.Values, update bool) (string, *jason.Object, []error) {
	if params.Get("code") == "" {
		panic("Get code, request token returned by the server (the token default lifetime is 30 sec).")
	}
	t.generateParams(params, "applicationId", "applicationScope", "")
	params.Set("grant_type", "authorization_code")

	urlRequest := t.GetUrlOAuthToken("", params)

	_, _, data, errs := t.execute(urlRequest, nil)

	if len(errs) > 0 {
		return urlRequest, nil, errs
	}

	if update {
		t.updateAccessParams(data)
	}

	return urlRequest, data, nil
}

func (t *Bitrix24) GetNewAccessToken(update bool) (string, *jason.Object, []error) {

}

/*func (t *Bitrix24) UpdateToken(update bool) (string, *jason.Object, []error) {
	params := url.Values{
		"": {""},
		"": {""},
		"": {""},
		"": {""},
	}

	url := t.GetUrlOAuthToken("", &params)

}*/

func (t *Bitrix24) updateAccessParams(data *jason.Object) {
	memberId, _ := data.GetString("member_id")
	accessToken, _ := data.GetString("access_token")
	refreshToken, _ := data.GetString("refresh_token")
	applicationScope, _ := data.GetString("scope")

	t.SetAttributes(SettingsInterface{
		MemberId:         memberId,
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		ApplicationScope: applicationScope,
	})
}

func (t *Bitrix24) GetUrlOAuthToken(url string, params *url.Values) string {
	return t.GetUrl(t.domain+OAUTH_TOKEN+url, params)
}

func (t *Bitrix24) GetUrlOAuth(url string, params *url.Values) string {
	return t.GetUrl(OAUTH_SERVER+url, params)
}

func (t *Bitrix24) GetUrlAuth(url string, params *url.Values) string {
	return t.GetUrl(t.domain+AUTH_URL+url, params)
}

func (t *Bitrix24) GetUrl(url string, params *url.Values) string {
	urlParams := ""
	if params != nil {
		urlParams = params.Encode()
	}

	return PROTOCOL + url + "?" + urlParams
}

func (t *Bitrix24) generateParams(params *url.Values, listParams ...string) []error {
	errs := []error{}

	tReflect := reflect.ValueOf(&t)

	if tReflect.Kind() == reflect.Ptr {
		tReflect = tReflect.Elem()
	}

	for _, value := range listParams {
		if len(realationNames[value]) > 0 {
			methodName := strings.Title(value)
			params.Set(realationNames[value], tReflect.MethodByName(methodName).Call([]reflect.Value{})[0].String())
		} else {
			errs = append(errs, errors.New(value+" isn't exist"))
		}
	}

	return errs
}

func (t *Bitrix24) execute(url string,
	additionalParameters url.Values) (*http.Request, *http.Response, *jason.Object, []error) {
	request.Post(url).SendMap(additionalParameters).Timeout(30 * time.Second)
	request.TargetType = "form"

	resp, body, errs := request.End()

	req, _ := request.MakeRequest()

	if len(errs) > 0 {
		return req, resp, nil, errs
	}

	data, _ := jason.NewObjectFromBytes([]byte(body))

	//json.Unmarshal([]byte(body), &data)

	return req, resp, data, nil
}

/*
	#Block getter and setter function
*/

func (t *Bitrix24) CheckAccessParams() (bool, []error) {
	errs := []error{}

	if len(t.Domain()) == 0 {
		errs = append(errs, errors.New("Domain is empty"))
	}
	if len(t.ApplicationSecret()) == 0 {
		errs = append(errs, errors.New("ApplicationSecret is empty"))
	}
	if len(t.ApplicationId()) == 0 {
		errs = append(errs, errors.New("ApplicationId is empty"))
	}

	if len(t.AccessToken()) == 0 {
		errs = append(errs, errors.New("AccessToken is empty"))
	}
	if len(t.RefreshToken()) == 0 {
		errs = append(errs, errors.New("RefreshToken is empty"))
	}
	if len(t.MemberId()) == 0 {
		errs = append(errs, errors.New("MemberId is empty"))
	}

	if len(t.ApplicationScope()) == 0 {
		errs = append(errs, errors.New("ApplicationScope is empty"))
	}

	t.isAccessParams = len(errs) == 0

	return t.isAccessParams, errs
}

func (t *Bitrix24) IsAccessParams() bool {
	return t.isAccessParams
}

func (t *Bitrix24) Domain() string {
	return t.domain
}
func (t *Bitrix24) SetDomain(domain string) {
	t.domain = domain
}

func (t *Bitrix24) ApplicationSecret() string {
	return t.applicationSecret
}
func (t *Bitrix24) SetApplicationSecret(clientSecret string) {
	t.applicationSecret = clientSecret
}

func (t *Bitrix24) ApplicationId() string {
	return t.applicationId
}
func (t *Bitrix24) SetApplicationId(applicationId string) {
	t.applicationId = applicationId
}

func (t *Bitrix24) AccessToken() string {
	return t.accessToken
}
func (t *Bitrix24) SetAccessToken(accessToken string) {
	t.accessToken = accessToken
}

func (t *Bitrix24) RefreshToken() string {
	return t.refreshToken
}
func (t *Bitrix24) SetRefreshToken(refreshToken string) {
	t.refreshToken = refreshToken
}

func (t *Bitrix24) MemberId() string {
	return t.memberId
}
func (t *Bitrix24) SetMemberId(memberId string) {
	t.memberId = memberId
}

func (t *Bitrix24) ApplicationScope() string {
	return t.applicationScope
}
func (t *Bitrix24) SetApplicationScope(applicationScope string) {
	t.applicationScope = applicationScope
}

func (t *Bitrix24) RedirectUri() string {
	return t.redirectUri
}
func (t *Bitrix24) SetRedirectUri(redirectUri string) {
	t.redirectUri = redirectUri
}

func (t *Bitrix24) Timeout() int {
	return t.timeout
}
func (t *Bitrix24) SetTimeout(timeout int) {
	t.timeout = timeout
}
