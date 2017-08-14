package bitrix24

import (
	"testing"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/skratchdot/open-golang/open"
	"net/http"
	"fmt"
	"encoding/json"
	"log"
	"net/url"
	"time"
	. "bitrix24/interfaces"
)

const (
	URL_SERVER  = "http://localhost"
	PORT_SERVER = ":8081"

	DOMAIN             = "demoapplic.bitrix24.ru"
	APPLICATION_ID     = "local.598ca908469d91.28776402"
	APPLICATION_SECRET = "5Ujx5Od38TdQ4jSNwyc2EtZRCPhclvIpNRaYbKKJuc24ZoIxd7"

	ACCESS_TOKEN  = "90vucimcz2bn349ableoe9z0gchswes5"
	REFRESH_TOKEN = "4dd9pmb1x2iml0hd821qoz5bg39k8qek"
	MEMBER_ID     = "56de4be4aba5516795b585fbaf0798ea"

	APPLICATION_SCOPE = "crm, lists"
	REDIRECT_URL      = URL_SERVER + "/authorization/"
)

func TestBitrix24(t *testing.T) {
	bx24 := Bitrix24{}
	serverStringChannel := make(chan url.Values)
	srv := startServer(serverStringChannel)

	title := ""

	Convey("Check Bitrix24", t, func() {
		bx24.Init(
			DOMAIN,
			APPLICATION_SECRET,
			APPLICATION_ID,
			nil)

		Convey("Check initiation Bitrix24", func() {

			So(bx24.Domain(), ShouldEqual, DOMAIN)
			So(bx24.ApplicationSecret(), ShouldEqual, APPLICATION_SECRET)
			So(bx24.ApplicationId(), ShouldEqual, APPLICATION_ID)
		})

		setting := SettingsInterface{
			AccessToken:      ACCESS_TOKEN,
			RefreshToken:     REFRESH_TOKEN,
			MemberId:         MEMBER_ID,
			ApplicationScope: APPLICATION_SCOPE,
			RedirectUri:      REDIRECT_URL,
		}

		bx24.SetAttributes(setting)

		Convey("SetAttributes Bitrix24", func() {
			So(bx24.Domain(), ShouldEqual, DOMAIN)
			So(bx24.AccessToken(), ShouldEqual, ACCESS_TOKEN)
			So(bx24.RefreshToken(), ShouldEqual, REFRESH_TOKEN)
			So(bx24.MemberId(), ShouldEqual, MEMBER_ID)
			So(bx24.ApplicationScope(), ShouldResemble, APPLICATION_SCOPE)
			So(bx24.RedirectUri(), ShouldEqual, REDIRECT_URL)
		})

		params := &url.Values{}
		errs := bx24.generateParams(params, "domain", "applicationSecret",
			"applicationId", "accessToken", "applicationScope", "refreshToken",
			"memberId", "applicationScope", "redirectUri", "FALSE_PARAMS")

		Convey("generateParams Bitrix24", func() {
			So(len(errs), ShouldEqual, 1)
			So(params.Get("domain"), ShouldEqual, DOMAIN)
			So(params.Get("client_secret"), ShouldEqual, APPLICATION_SECRET)
			So(params.Get("client_id"), ShouldEqual, APPLICATION_ID)
			So(params.Get("access_token"), ShouldEqual, ACCESS_TOKEN)
			So(params.Get("refresh_token"), ShouldResemble, REFRESH_TOKEN)
			So(params.Get("member_id"), ShouldEqual, MEMBER_ID)
			So(params.Get("scope"), ShouldResemble, APPLICATION_SCOPE)
			So(params.Get("redirect_uri"), ShouldEqual, REDIRECT_URL)
		})

		Convey("Execute Bitrix24", func() {
			data := url.Values{
				"key1": {"value1"},
				"key2": {"value2"},
				"key3": {"value3"},
				"key4": {"value4"},
			}

			_, _, result, _ := bx24.execute(URL_SERVER+PORT_SERVER+"/simplePost/", data)

			jsData, _ := json.Marshal(data)

			So(string(jsData), ShouldEqual, result.String())
		})

		Convey("Check auth Bitrix24", func() {

			params := url.Values{
				"client_id":     {bx24.ApplicationId()},
				"state":         {time.Now().String()},
				"redirect_uri":  {bx24.RedirectUri()},
				"response_type": {"code"},
				"scope":         {bx24.ApplicationScope()},
			}

			urlAuthClient := PROTOCOL + bx24.Domain() + AUTH_URL + "?" + params.Encode()

			Convey("getUrl Bitrix24", func() {
				So(urlAuthClient, ShouldEqual, bx24.GetUrlClientAuth(&params))
			})

			title = "Check authorization Bitrix24"

			if testing.Short() {
				SkipConvey(title, func() {})
			} else {
				clientAuthTest := func(update bool) {
					open.Run(urlAuthClient)

					params := <-serverStringChannel

					params.Set("grant_type", "authorization_code")
					params.Set("client_id", bx24.ApplicationId())
					params.Set("client_secret", bx24.ApplicationSecret())
					params.Set("scope", bx24.ApplicationScope())

					urlAuthToken := PROTOCOL + bx24.Domain() + OAUTH_TOKEN + "?" + params.Encode()

					urlAccessTokenCkeck, data, _ := bx24.GetFirstAccessToken(&params, update)

					So(urlAccessTokenCkeck, ShouldEqual, urlAuthToken)

					setting = *GetSettingsByJson(data)

					if update {
						So(bx24.AccessToken(), ShouldEqual, setting.AccessToken)
						So(bx24.RefreshToken(), ShouldEqual, setting.RefreshToken)

						Print("\nAccessToken = " + setting.AccessToken + "\n" +
							"RefreshToken = " + setting.RefreshToken + "\n" +
							"MemberId = " + setting.MemberId + "\n")
					} else {
						So(bx24.AccessToken(), ShouldNotEqual, setting.AccessToken)
						So(bx24.RefreshToken(), ShouldNotEqual, setting.RefreshToken)
					}
				}
				Convey(title, func() {
					clientAuthTest(true)
					clientAuthTest(false)
				})
			}
		})
	})

	srv.Close()
}

func startServer(channel chan<- url.Values) *http.Server {
	srv := &http.Server{Addr: PORT_SERVER}
	http.HandleFunc("/simplePost/", func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "application/json")

		r.ParseMultipartForm(32 << 20)
		js, _ := json.Marshal(r.PostForm)
		fmt.Fprintf(w, "%s", string(js))
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		channel <- r.URL.Query()

		w.Header().Set("Content-Type", "application/json")

		fmt.Fprint(w, "ok")
	})

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			// cannot panic, because this probably is an intentional close
			log.Printf("Httpserver: ListenAndServe() error: %s", err)
		}
	}()

	return srv
}
