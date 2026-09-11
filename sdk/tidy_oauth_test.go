package sdk

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/fishpioffical/golang-sdk/config"
)

func TestOpenIDScopedURL(t *testing.T) {
	s := NewSDKWithAPIKey("unused")
	legacy := s.GetOpenIdUrl("https://example.com", "https://example.com/callback?state=a&b=中文")
	if s.GetOpenIdUrlWithScopes("https://example.com", "https://example.com/callback?state=a&b=中文") != legacy {
		t.Fatal("no-scope URL changed")
	}
	u, err := url.Parse(s.GetOpenIdUrlWithScopes("https://example.com", "https://example.com/callback?state=a&b=中文", "profile.read", "points.read"))
	if err != nil {
		t.Fatal(err)
	}
	q := u.Query()
	if q.Get("fishpi.scope") != "profile.read points.read" || q.Get("openid.return_to") != "https://example.com/callback?state=a&b=中文" || q.Get("openid.mode") != "checkid_setup" {
		t.Fatalf("bad URL parameters: %v", q)
	}
	if q.Has("apiKey") {
		t.Error("API key in authorization URL")
	}
}

func TestOpenIDVerifyTokensAndLegacy(t *testing.T) {
	query := map[string]string{"openid.mode": "id_res", "openid.identity": "https://fishpi.cn/openid/id/1659430635383", "openid.sig": "signature", "state": "do-not-forward"}
	before := map[string]string{}
	for k, v := range query {
		before[k] = v
	}
	s := newTestSDK(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/openid/verify" || r.URL.RawQuery != "" {
			t.Errorf("bad verify request: %s %s", r.Method, r.URL)
		}
		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		if body["openid.mode"] != "check_authentication" || body["openid.sig"] != "signature" {
			t.Error("verification parameters missing")
		}
		if _, ok := body["state"]; ok {
			t.Error("non-OpenID parameter forwarded")
		}
		w.Header().Set("Content-Type", "text/plain")
		_, _ = io.WriteString(w, "ns:http://specs.openid.net/auth/2.0\r\nis_valid:true\r\nscope:profile.read points.read\r\ntoken_type:Bearer\r\naccess_token:access:token\r\nexpires_in:604800\r\nrefresh_token:refresh:token\r\nrefresh_expires_in:31104000\r\n")
	})
	r, err := s.PostOpenIdVerifyWithTokens(query)
	if err != nil {
		t.Fatal(err)
	}
	if r.OpenID != "1659430635383" || r.AccessToken != "access:token" || r.RefreshToken != "refresh:token" || r.ExpiresIn != 604800 || r.RefreshExpiresIn != 31104000 || r.Scope != "profile.read points.read" {
		t.Fatal("verification response fields lost")
	}
	legacy, err := s.PostOpenIdVerify(query)
	if err != nil || legacy == nil || *legacy != r.OpenID {
		t.Fatalf("legacy verification failed: %v", err)
	}
	if !reflect.DeepEqual(query, before) {
		t.Error("caller map was changed")
	}
}

func TestOpenIDRejectsInvalidResponses(t *testing.T) {
	for _, body := range []string{
		"is_valid:false\naccess_token:SECRET",
		"<html>is_valid:true</html>",
		"is_valid:true\nis_valid:false",
		"is_valid:true\nexpires_in:SECRET",
		"is_valid:true\nrefresh_expires_in:-1",
	} {
		t.Run(strings.Split(body, "\n")[0], func(t *testing.T) {
			s := newTestSDK(t, func(w http.ResponseWriter, r *http.Request) { _, _ = io.WriteString(w, body) })
			r, err := s.PostOpenIdVerifyWithTokens(map[string]string{"openid.identity": "https://fishpi.cn/openid/id/123"})
			if err == nil || r != nil {
				t.Fatal("invalid verification accepted")
			}
			if strings.Contains(err.Error(), "SECRET") {
				t.Error("raw response leaked into error")
			}
		})
	}
	s := newTestSDK(t, func(w http.ResponseWriter, r *http.Request) { _, _ = io.WriteString(w, "is_valid:true\n") })
	if _, err := s.PostOpenIdVerify(nil); err == nil {
		t.Error("missing identity accepted")
	}
	if id, err := s.PostOpenIdVerify(map[string]string{"openid.identity": "https://fishpi.cn/openid/id/123"}); err != nil || *id != "123" {
		t.Fatal("legacy tokenless response rejected")
	}
}

func TestOAuthResourceContracts(t *testing.T) {
	tests := []struct {
		path, response string
		query          url.Values
		call           func(*FishPiSDK) (bool, error)
	}{
		{path: "/openid/user/profile", response: `{"code":0,"data":{"userId":"123","userName":"tester"}}`, call: func(s *FishPiSDK) (bool, error) {
			r, e := s.GetOAuthProfile("access-token")
			return e == nil && r.Data != nil && r.Data.UserID == "123" && r.Data.UserName == "tester", e
		}},
		{path: "/openid/user/detail", response: `{"code":0,"data":{"oId":"123","userNo":"1","userAppRole":"0","sysMetal":{"list":[]},"userPoint":183939,"userRole":"普通用户"}}`, call: func(s *FishPiSDK) (bool, error) {
			r, e := s.GetOAuthUserDetail("access-token")
			return e == nil && r.Data != nil && r.Data.UserNo.String() == "1" && r.Data.UserAppRole.String() == "0" && string(r.Data.SysMetal) == `{"list":[]}`, e
		}},
		{path: "/openid/user/membership", response: `{"code":0,"data":{"active":true,"state":1,"lvCode":"VIP1","expiresAt":1760000000000,"configJson":"{}"}}`, call: func(s *FishPiSDK) (bool, error) {
			r, e := s.GetOAuthMembership("access-token")
			return e == nil && r.Data != nil && r.Data.Active && r.Data.ExpiresAt == 1760000000000, e
		}},
		{path: "/openid/user/points", query: url.Values{"p": {"2"}, "size": {"200"}}, response: `{"code":0,"data":{"userId":"123","userPoint":99,"records":[{"sum":20,"time":1760000000000,"sourceAppName":"应用","sourceScene":"point_issue","operation":"+","balance":99}],"pagination":{"paginationCurrentPageNum":2,"paginationPageSize":200,"paginationRecordCount":201,"paginationPageCount":2,"paginationPageNums":[1,2]}}}`, call: func(s *FishPiSDK) (bool, error) {
			r, e := s.GetOAuthPoints("access-token", 2, 200)
			return e == nil && r.Data != nil && len(r.Data.Records) == 1 && r.Data.Records[0].SourceAppName == "应用" && r.Data.Pagination.PaginationCurrentPageNum == 2 && r.Data.Pagination.PaginationPageSize == 200, e
		}},
		{path: "/openid/user/articles", query: url.Values{"p": {"1"}, "size": {"100"}}, response: `{"code":0,"data":{"userId":"123","articles":[{"oId":"a1","articleTitle":"标题","articleType":6,"articleCreateTime":1760000000000}],"pagination":{"paginationPageCount":1}}}`, call: func(s *FishPiSDK) (bool, error) {
			r, e := s.GetOAuthArticles("access-token", 1, 100)
			return e == nil && r.Data != nil && len(r.Data.Articles) == 1 && r.Data.Articles[0].ArticleCreateTime == 1760000000000 && r.Data.Articles[0].ArticleType == 6, e
		}},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			s := newTestSDK(t, func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "GET" || r.URL.Path != tt.path {
					t.Errorf("bad request: %s %s", r.Method, r.URL)
				}
				if r.Header.Get("Authorization") != "Bearer access-token" || r.Header.Get("User-Agent") != "sdk-contract-test" {
					t.Error("missing Bearer or UA")
				}
				q := tt.query
				if q == nil {
					q = url.Values{}
				}
				if !reflect.DeepEqual(r.URL.Query(), q) {
					t.Errorf("query = %v, want %v", r.URL.Query(), q)
				}
				writeJSON(w, tt.response)
			})
			ok, err := tt.call(s)
			if !ok || err != nil {
				t.Fatalf("response contract: %v", err)
			}
		})
	}
}

func TestOAuthRefreshAndLoggingIsolation(t *testing.T) {
	logDir := t.TempDir()
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, nil))
	s := newTestSDK(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "" || r.Header.Get("Authorization") != "" || r.Header.Get("Cookie") != "" {
			t.Error("unrelated credentials sent to token endpoint")
		}
		if r.Method != "POST" || r.URL.Path != "/openid/token" {
			t.Error("incorrect refresh endpoint")
		}
		checkJSONFields(t, r, map[string]any{"grant_type": "refresh_token", "refresh_token": "old-secret"})
		writeJSON(w, `{"code":0,"data":{"token_type":"Bearer","access_token":"new-access-secret","refresh_token":"new-refresh-secret","expires_in":604800,"refresh_expires_in":31104000,"scope":"profile.read","unknown":"unlogged-secret"}}`)
	}, WithLogDir(logDir), WithCustomUnmarshaler(logger))
	s.client.SetCommonHeader("Authorization", "Bearer other-secret").SetCommonHeader("Cookie", "other-cookie")
	r, err := s.PostOpenIdToken("old-secret")
	if err != nil {
		t.Fatal(err)
	}
	if r.Data == nil || r.Data.RefreshToken != "new-refresh-secret" || r.Data.ExpiresIn != 604800 {
		t.Fatal("rotated tokens lost")
	}
	files, err := os.ReadDir(logDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 0 || logs.Len() != 0 {
		t.Fatal("OAuth request or response was logged")
	}
	if s.GetAPIKey() != "local-test-key" {
		t.Error("refresh changed API key configuration")
	}
}

func TestOAuthErrorsAndNoRedirect(t *testing.T) {
	for _, tt := range []struct {
		body    string
		status  int
		wantErr bool
	}{
		{`{"code":-1,"msg":"无权限"}`, 200, false},
		{`{"code":-1,"msg":"无权限"}`, 401, true},
		{"<html>SECRET</html>", 503, true},
		{`{"data":{"access_token":"SECRET"}}`, 200, true},
		{`{"code":0,"data":"SECRET"}`, 200, true},
	} {
		t.Run(http.StatusText(tt.status), func(t *testing.T) {
			s := newTestSDK(t, func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.status)
				_, _ = io.WriteString(w, tt.body)
			})
			r, err := s.GetOAuthProfile("token")
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v", err)
			}
			if err != nil && strings.Contains(err.Error(), "SECRET") {
				t.Error("secret in error")
			}
			if !tt.wantErr && (r.Code != -1 || r.Msg != "无权限" || r.Data != nil) {
				t.Error("business error lost")
			}
		})
	}
	s := newTestSDK(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/openid/token" {
			t.Error("redirect followed")
		}
		w.Header().Set("Location", "/unexpected")
		w.WriteHeader(http.StatusTemporaryRedirect)
	})
	if _, err := s.PostOpenIdToken("secret"); err == nil {
		t.Error("redirect treated as success")
	}
}

func TestOAuthConcurrentCredentialIsolation(t *testing.T) {
	s := newTestSDK(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/openid/user/profile" {
			if r.URL.Query().Has("apiKey") || r.Header.Get("Authorization") != "Bearer user-token" {
				t.Error("OAuth credentials mixed")
			}
			writeJSON(w, `{"code":0,"data":{"userId":"123"}}`)
		} else {
			if r.Header.Get("Authorization") != "" || r.URL.Query().Get("apiKey") != "local-test-key" {
				t.Error("API key credentials mixed")
			}
			writeJSON(w, `{"code":0,"data":{"drafts":[]}}`)
		}
	})
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			if _, err := s.GetOAuthProfile("user-token"); err != nil {
				t.Error(err)
			}
		}()
		go func() {
			defer wg.Done()
			if _, err := s.GetArticleDrafts(); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
}

func TestOAuthUsesUpdatedBaseURLAndCustomUA(t *testing.T) {
	s := newTestSDK(t, func(w http.ResponseWriter, r *http.Request) { t.Error("old BaseURL used") })
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") != "custom-UA" || r.URL.RawQuery != "" {
			t.Error("UA or default pagination incorrect")
		}
		writeJSON(w, `{"code":0,"data":{"articles":[]}}`)
	}))
	defer server.Close()
	WithUserAgent("custom-UA")(s)
	if err := s.UpdateConfig(&config.Config{BaseUrl: server.URL, ApiKey: "new-key"}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.GetOAuthArticles("token", 0, 0); err != nil {
		t.Fatal(err)
	}
}

func TestOAuthDetailDocumentVariants(t *testing.T) {
	for _, body := range []string{
		`{"code":0,"data":{"userNo":1,"userAppRole":0,"sysMetal":"{\"list\":[]}"}}`,
		`{"code":0,"data":{"userNo":"1","userAppRole":"0","sysMetal":{"list":[]}}}`,
	} {
		s := newTestSDK(t, func(w http.ResponseWriter, r *http.Request) { writeJSON(w, body) })
		result, err := s.GetOAuthUserDetail("token")
		if err != nil {
			t.Fatal(err)
		}
		if result.Data == nil || result.Data.UserNo.String() != "1" || result.Data.UserAppRole.String() != "0" || !json.Valid(result.Data.SysMetal) {
			t.Fatal("documented detail representation was not preserved")
		}
	}
}

func TestOAuthTransportAndVerificationHTTPFailures(t *testing.T) {
	s := newTestSDK(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = io.WriteString(w, "is_valid:true\naccess_token:SECRET")
	})
	if result, err := s.PostOpenIdVerifyWithTokens(map[string]string{"openid.identity": "https://fishpi.cn/openid/id/123"}); err == nil || result != nil {
		t.Fatal("failed HTTP verification accepted")
	}
	s.oauthClient.SetTimeout(1)
	if result, err := s.PostOpenIdToken("SECRET"); err == nil || result != nil || strings.Contains(err.Error(), "SECRET") {
		t.Fatal("transport failure should return a sanitized error")
	}
}
