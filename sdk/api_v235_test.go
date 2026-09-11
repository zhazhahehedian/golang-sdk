package sdk

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/fishpioffical/golang-sdk/config"
	"github.com/fishpioffical/golang-sdk/types"
)

func newTestSDK(t *testing.T, handler http.HandlerFunc, options ...Option) *FishPiSDK {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return NewSDK(config.NewMemoryConfigProvider(&config.Config{
		BaseUrl: server.URL, ApiKey: "local-test-key", UserAgent: "sdk-contract-test",
	}), options...)
}

func writeJSON(w http.ResponseWriter, body string) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = io.WriteString(w, body)
}

func checkJSONFields(t *testing.T, r *http.Request, want map[string]any) {
	t.Helper()
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		t.Errorf("Content-Type = %q", r.Header.Get("Content-Type"))
	}
	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		t.Error(err)
		return
	}
	for key, value := range want {
		actual, ok := body[key]
		if !ok || !reflect.DeepEqual(actual, value) {
			t.Errorf("body[%s] = %#v, want %#v", key, actual, value)
		}
	}
}

func TestV235APIContracts(t *testing.T) {
	falseValue, zero, longArticle := false, 0, 6
	tests := []struct {
		name, method, path, response string
		query                        url.Values
		body                         map[string]any
		call                         func(*FishPiSDK) (bool, error)
	}{
		{name: "draft list", method: "GET", path: "/api/article-drafts",
			response: `{"code":0,"data":{"drafts":[{"oId":"d1","articleDraftTitle":"标题","articleDraftType":6,"articleDraftUpdatedTime":1770000000000}]}}`,
			call: func(s *FishPiSDK) (bool, error) {
				r, e := s.GetArticleDrafts()
				return e == nil && r.Data != nil && len(r.Data.Drafts) == 1 && r.Data.Drafts[0].ArticleDraftType == 6 && r.Data.Drafts[0].ArticleDraftUpdatedTime == 1770000000000, e
			}},
		{name: "draft create", method: "POST", path: "/api/article-drafts",
			body:     map[string]any{"apiKey": "local-test-key", "articleTitle": "新草稿", "articleContent": "", "articleType": float64(6), "articleCommentable": false, "articleRewardPoint": float64(0), "articleShowInList": float64(0), "columnId": "c1", "chapterNo": "第1章"},
			response: `{"code":0,"data":{"draft":{"oId":"d2","articleDraftTitle":"新草稿"}}}`,
			call: func(s *FishPiSDK) (bool, error) {
				r, e := s.PostArticleDraft(&types.PostArticleDraftRequest{ArticleTitle: "新草稿", ArticleType: &longArticle, ArticleCommentable: &falseValue, ArticleRewardPoint: &zero, ArticleShowInList: &zero, ColumnID: "c1", ChapterNo: "第1章"})
				return e == nil && r.Data != nil && r.Data.Draft != nil && r.Data.Draft.OId == "d2", e
			}},
		{name: "draft update", method: "POST", path: "/api/article-drafts",
			body:     map[string]any{"apiKey": "local-test-key", "articleDraftId": "d1", "articleContent": "更新正文", "articleRewardContent": ""},
			response: `{"code":0,"data":{"draft":{"oId":"d1"}}}`,
			call: func(s *FishPiSDK) (bool, error) {
				r, e := s.PostArticleDraft(&types.PostArticleDraftRequest{ArticleDraftID: "d1", ArticleContent: "更新正文"})
				return e == nil && r.Data != nil && r.Data.Draft != nil && r.Data.Draft.OId == "d1", e
			}},
		{name: "draft detail", method: "GET", path: "/api/article-drafts/d1",
			response: `{"code":0,"data":{"draft":{"oId":"d1","articleDraftContent":"正文","articleDraftThoughtContent":"思绪","articleDraftRewardContent":"打赏","articleDraftChapterNo":"1"}}}`,
			call: func(s *FishPiSDK) (bool, error) {
				r, e := s.GetArticleDraft("d1")
				return e == nil && r.Data != nil && r.Data.Draft != nil && r.Data.Draft.ArticleDraftContent == "正文" && r.Data.Draft.ArticleDraftRewardContent == "打赏" && r.Data.Draft.ArticleDraftThoughtContent == "思绪", e
			}},
		{name: "draft delete", method: "DELETE", path: "/api/article-drafts/d1",
			response: `{"code":0,"data":{"id":"d1"}}`,
			call: func(s *FishPiSDK) (bool, error) {
				r, e := s.DeleteArticleDraft("d1")
				return e == nil && r.Data != nil && r.Data.ID == "d1", e
			}},
		{name: "repeater list", method: "GET", path: "/api/repeater/items", query: url.Values{"repeaterContentType": {"joke"}},
			response: `{"code":0,"data":{"items":[{"oId":"r1","repeaterContent":"笑话","repeaterContentLiked":true,"repeaterContentCreatedTime":1770000000000}]}}`,
			call: func(s *FishPiSDK) (bool, error) {
				r, e := s.GetRepeaterItems("joke")
				return e == nil && r.Data != nil && len(r.Data.Items) == 1 && r.Data.Items[0].RepeaterContentLiked && r.Data.Items[0].RepeaterContentCreatedTime == 1770000000000, e
			}},
		{name: "repeater next", method: "GET", path: "/api/repeater/next", query: url.Values{"repeaterContentType": {"fish"}, "excludeId": {"r1"}},
			response: `{"code":0,"data":{"item":{"oId":"r2","repeaterContentSource":"seed"}}}`,
			call: func(s *FishPiSDK) (bool, error) {
				r, e := s.GetRepeaterNext("fish", "r1")
				return e == nil && r.Data != nil && r.Data.Item != nil && r.Data.Item.OId == "r2", e
			}},
		{name: "repeater post", method: "POST", path: "/api/repeater",
			body:     map[string]any{"apiKey": "local-test-key", "repeaterContentType": "kfc", "repeaterContent": "今天星期四"},
			response: `{"code":0,"data":{"item":{"oId":"r3","repeaterContent":"今天星期四"}}}`,
			call: func(s *FishPiSDK) (bool, error) {
				r, e := s.PostRepeater("kfc", "今天星期四")
				return e == nil && r.Data != nil && r.Data.Item != nil && r.Data.Item.RepeaterContent == "今天星期四", e
			}},
		{name: "repeater unlike", method: "POST", path: "/api/repeater/r1/like", body: map[string]any{"apiKey": "local-test-key"},
			response: `{"code":0,"data":{"liked":false,"repeaterContentLikeCount":10}}`,
			call: func(s *FishPiSDK) (bool, error) {
				r, e := s.PostRepeaterLike("r1")
				return e == nil && r.Data != nil && !r.Data.Liked && r.Data.RepeaterContentLikeCount == 10, e
			}},
		{name: "column cover", method: "POST", path: "/api/columns/c1/cover", body: map[string]any{"apiKey": "local-test-key", "columnCoverURL": "https://file.fishpi.cn/demo.jpg"},
			response: `{"code":0,"data":{"column":{"oId":"c1","columnCoverURL":"https://file.fishpi.cn/demo.jpg","columnHasCover":true}}}`,
			call: func(s *FishPiSDK) (bool, error) {
				r, e := s.PostColumnCover("c1", "https://file.fishpi.cn/demo.jpg")
				return e == nil && r.Data != nil && r.Data.Column != nil && r.Data.Column.ColumnHasCover, e
			}},
		{name: "clear column cover", method: "POST", path: "/api/columns/c1/cover", body: map[string]any{"apiKey": "local-test-key", "columnCoverURL": ""},
			response: `{"code":0,"data":{"column":{"oId":"c1","columnCoverURL":"","columnHasCover":false}}}`,
			call: func(s *FishPiSDK) (bool, error) {
				r, e := s.PostColumnCover("c1", "")
				return e == nil && r.Data != nil && r.Data.Column != nil && !r.Data.Column.ColumnHasCover && r.Data.Column.ColumnCoverURL == "", e
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := 0
			s := newTestSDK(t, func(w http.ResponseWriter, r *http.Request) {
				calls++
				if r.Method != tt.method || r.URL.Path != tt.path {
					t.Errorf("request = %s %s, want %s %s", r.Method, r.URL.Path, tt.method, tt.path)
				}
				wantQuery := url.Values{"apiKey": {"local-test-key"}}
				for k, v := range tt.query {
					wantQuery[k] = v
				}
				if !reflect.DeepEqual(r.URL.Query(), wantQuery) {
					t.Errorf("query = %v, want %v", r.URL.Query(), wantQuery)
				}
				if r.Header.Get("User-Agent") != "sdk-contract-test" || r.Header.Get("Authorization") != "" {
					t.Error("incorrect auth or UA")
				}
				if tt.body != nil {
					checkJSONFields(t, r, tt.body)
				}
				writeJSON(w, tt.response)
			})
			ok, err := tt.call(s)
			if err != nil || !ok {
				t.Fatalf("response contract failed: %v", err)
			}
			if calls != 1 {
				t.Errorf("sent %d requests", calls)
			}
		})
	}
}

func TestDraftOptionalValues(t *testing.T) {
	s := newTestSDK(t, func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		for _, key := range []string{"articleDraftId", "articleType", "articleCommentable", "articleRewardPoint", "articleAnonymous", "articleNotifyFollowers", "articleShowInList", "articleStatement", "articleQnAOfferPoint"} {
			if _, ok := body[key]; ok {
				t.Errorf("unset field %s should be omitted", key)
			}
		}
		writeJSON(w, `{"code":0,"data":{"draft":{"oId":"d1"}}}`)
	})
	if _, err := s.PostArticleDraft(&types.PostArticleDraftRequest{}); err != nil {
		t.Fatal(err)
	}
}

func TestNewAPIResponseFailures(t *testing.T) {
	for _, tt := range []struct {
		name, body, contentType string
		status                  int
		wantErr                 bool
		wantCode                int
	}{
		{"business error", `{"code":-1,"msg":"草稿不存在"}`, "application/json", 200, false, -1},
		{"HTTP error", `{"code":-1,"msg":"无权限"}`, "application/json", 403, true, -1},
		{"HTML error", "<html>unavailable</html>", "text/html", 503, true, 0},
		{"malformed JSON", "{", "application/json", 200, true, 0},
		{"empty body", "", "application/json", 200, true, 0},
		{"missing code", `{"data":{}}`, "application/json", 200, true, 0},
		{"wrong data type", `{"code":0,"data":[]}`, "application/json", 200, true, 0},
		{"unlabelled JSON", `{"code":0,"data":{"drafts":[]}}`, "text/plain", 200, false, 0},
	} {
		t.Run(tt.name, func(t *testing.T) {
			s := newTestSDK(t, func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", tt.contentType)
				w.WriteHeader(tt.status)
				_, _ = io.WriteString(w, tt.body)
			})
			r, err := s.GetArticleDrafts()
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v", err)
			}
			if !tt.wantErr && (r == nil || r.Code != tt.wantCode) {
				t.Fatalf("response = %#v", r)
			}
			if tt.name == "business error" && (r.Msg != "草稿不存在" || r.Data != nil) {
				t.Fatal("business error lost")
			}
		})
	}
}

func TestNewAPIInputValidation(t *testing.T) {
	s := newTestSDK(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("invalid input sent an HTTP request")
		w.WriteHeader(500)
	})
	tests := []func() error{
		func() error { _, e := s.PostArticleDraft(nil); return e },
		func() error { _, e := s.GetArticleDraft(""); return e },
		func() error { _, e := s.DeleteArticleDraft(" "); return e },
		func() error { _, e := s.PostRepeaterLike(""); return e },
		func() error { _, e := s.PostColumnCover("", ""); return e },
		func() error { _, e := s.GetOAuthProfile(""); return e },
		func() error { _, e := s.PostOpenIdToken(" "); return e },
		func() error { _, e := s.GetOAuthPoints("token", 1, 201); return e },
		func() error { _, e := s.GetOAuthArticles("token", 1, 101); return e },
		func() error { _, e := s.GetOAuthArticles("token", -1, 1); return e },
		func() error { _, e := s.GetOAuthPoints("token", 1, -1); return e },
	}
	for i, call := range tests {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			if call() == nil {
				t.Error("expected error")
			}
		})
	}
}

func TestOptionalRepeaterFiltersAndEscapedID(t *testing.T) {
	s := newTestSDK(t, func(w http.ResponseWriter, r *http.Request) {
		if len(r.URL.Query()) != 1 {
			t.Error("empty filters should be omitted")
		}
		switch r.URL.EscapedPath() {
		case "/api/repeater/items":
			writeJSON(w, `{"code":0,"data":{"items":[]}}`)
		case "/api/repeater/next":
			writeJSON(w, `{"code":0,"data":{"item":null}}`)
		case "/api/article-drafts/a%2Fb%3Fc":
			writeJSON(w, `{"code":0,"data":{"draft":{"oId":"a/b?c"}}}`)
		default:
			t.Errorf("unexpected URL: %s", r.URL)
			w.WriteHeader(404)
		}
	})
	if _, err := s.GetRepeaterItems(""); err != nil {
		t.Fatal(err)
	}
	if r, err := s.GetRepeaterNext("", ""); err != nil || r.Data.Item != nil {
		t.Fatalf("null item: %v", err)
	}
	if _, err := s.GetArticleDraft("a/b?c"); err != nil {
		t.Fatal(err)
	}
}
