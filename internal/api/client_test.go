package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func assertEqual(t *testing.T, expected, actual interface{}) {
	t.Helper()
	if expected != actual {
		t.Errorf("expected %v, got %v", expected, actual)
	}
}

func assertTrue(t *testing.T, actual bool) {
	t.Helper()
	if !actual {
		t.Errorf("expected true, got false")
	}
}

func assertFalse(t *testing.T, actual bool) {
	t.Helper()
	if actual {
		t.Errorf("expected false, got true")
	}
}

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func assertError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func assertContains(t *testing.T, s, substr string) {
	t.Helper()
	if !strings.Contains(s, substr) {
		t.Errorf("expected %q to contain %q", s, substr)
	}
}

func TestNewClient(t *testing.T) {
	c := NewClient("test-cookie", true)
	if c.HTTPClient == nil {
		t.Fatal("HTTPClient is nil")
	}
	assertEqual(t, "test-cookie", c.Cookie)
	assertEqual(t, BaseURL, c.BaseURL)
	assertTrue(t, c.HTTPClient.Transport.(*http.Transport).TLSClientConfig.InsecureSkipVerify)
	if c.Crypto == nil {
		t.Fatal("Crypto is nil")
	}
}

func TestNewClientWithExtra(t *testing.T) {
	c := NewClientWithExtra("cookie-val", "verif-code", "svc-id", false)
	assertEqual(t, "cookie-val", c.Cookie)
	assertEqual(t, "verif-code", c.VerificationCode)
	assertEqual(t, "svc-id", c.ServiceID)
	assertEqual(t, BaseURL, c.BaseURL)
	assertFalse(t, c.HTTPClient.Transport.(*http.Transport).TLSClientConfig.InsecureSkipVerify)
}

func TestClient_GetFullCookie(t *testing.T) {
	tests := []struct {
		name     string
		cookie   string
		verif    string
		svcID    string
		expected string
	}{
		{
			name:     "all fields",
			cookie:   "abc123",
			verif:    "v456",
			svcID:    "s789",
			expected: "#openPortal#token#=abc123; openPortalVerificationCode=v456; openPortalServiceID=s789",
		},
		{
			name:     "cookie with prefix",
			cookie:   "#openPortal#token#=abc123",
			verif:    "v456",
			svcID:    "s789",
			expected: "#openPortal#token#=abc123; openPortalVerificationCode=v456; openPortalServiceID=s789",
		},
		{
			name:     "only cookie",
			cookie:   "abc123",
			expected: "#openPortal#token#=abc123",
		},
		{
			name:     "empty",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				Cookie:           tt.cookie,
				VerificationCode: tt.verif,
				ServiceID:        tt.svcID,
			}
			assertEqual(t, tt.expected, c.GetFullCookie())
		})
	}
}

func TestClient_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertEqual(t, "/test-path", r.URL.Path)
		assertEqual(t, http.MethodGet, r.Method)
		assertEqual(t, "application/json, text/plain, */*", r.Header.Get("Accept"))
		assertEqual(t, "https://service.sd.10086.cn/aaas/", r.Header.Get("Referer"))
		assertEqual(t, "token-value", r.Header.Get("token"))
		assertEqual(t, "#openPortal#token#=token-value", r.Header.Get("Cookie"))

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"code":"00000","data":"ok"}`))
	}))
	defer server.Close()

	c := NewClient("token-value", true)
	c.BaseURL = server.URL

	resp, err := c.Get("/test-path")
	assertNoError(t, err)
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	assertEqual(t, `{"code":"00000","data":"ok"}`, string(body))
}

func TestClient_Post(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertEqual(t, "/test-post", r.URL.Path)
		assertEqual(t, http.MethodPost, r.Method)
		assertEqual(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))

		body, _ := io.ReadAll(r.Body)
		assertEqual(t, `{"key":"value"}`, string(body))

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success":true}`))
	}))
	defer server.Close()

	c := NewClient("token-value", true)
	c.BaseURL = server.URL

	resp, err := c.Post("/test-post", map[string]string{"key": "value"})
	assertNoError(t, err)
	defer resp.Body.Close()

	assertEqual(t, http.StatusOK, resp.StatusCode)
}

func TestClient_PostMultipart(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertEqual(t, "/test-multipart", r.URL.Path)
		assertEqual(t, http.MethodPost, r.Method)
		assertEqual(t, "multipart/form-data; boundary=xxx", r.Header.Get("Content-Type"))

		body, _ := io.ReadAll(r.Body)
		assertEqual(t, "form-data", string(body))

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success":true}`))
	}))
	defer server.Close()

	c := NewClient("token-value", true)
	c.BaseURL = server.URL

	resp, err := c.PostMultipart("/test-multipart", "multipart/form-data; boundary=xxx", []byte("form-data"))
	assertNoError(t, err)
	defer resp.Body.Close()

	assertEqual(t, http.StatusOK, resp.StatusCode)
}

func TestParseJSON(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		resp := httptest.NewRecorder()
		resp.WriteHeader(http.StatusOK)
		resp.Write([]byte(`{"code":"00000","msg":"ok"}`))

		var result struct {
			Code string `json:"code"`
			Msg  string `json:"msg"`
		}
		err := ParseJSON(resp.Result(), &result)
		assertNoError(t, err)
		assertEqual(t, "00000", result.Code)
		assertEqual(t, "ok", result.Msg)
	})

	t.Run("http error", func(t *testing.T) {
		resp := httptest.NewRecorder()
		resp.WriteHeader(http.StatusInternalServerError)
		resp.Write([]byte(`server error`))

		var result map[string]interface{}
		err := ParseJSON(resp.Result(), &result)
		assertError(t, err)
		assertContains(t, err.Error(), "HTTP 500")
	})

	t.Run("gbk encoded body", func(t *testing.T) {
		// GBK 编码的 {"msg":"中文测试"}
		gbkBody := []byte{0x7b, 0x22, 0x6d, 0x73, 0x67, 0x22, 0x3a, 0x22, 0xd6, 0xd0, 0xce, 0xc4, 0xb2, 0xe2, 0xca, 0xd4, 0x22, 0x7d}

		resp := httptest.NewRecorder()
		resp.WriteHeader(http.StatusOK)
		resp.Write(gbkBody)

		var result struct {
			Msg string `json:"msg"`
		}
		err := ParseJSON(resp.Result(), &result)
		assertNoError(t, err)
		assertEqual(t, "中文测试", result.Msg)
	})
}

func TestResponse_IsSuccess(t *testing.T) {
	tests := []struct {
		name     string
		resp     Response
		expected bool
	}{
		{"code 200", Response{Code: 200}, true},
		{"success true", Response{Success: true}, true},
		{"both", Response{Code: 200, Success: true}, true},
		{"failure", Response{Code: 500, Success: false}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertEqual(t, tt.expected, tt.resp.IsSuccess())
		})
	}
}

func TestResponse_Error(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		r := Response{Code: 200, Message: "ok"}
		assertNoError(t, r.Error())
	})

	t.Run("failure", func(t *testing.T) {
		r := Response{Code: 500, Message: "internal error"}
		err := r.Error()
		assertError(t, err)
		assertContains(t, err.Error(), "API error [500]")
		assertContains(t, err.Error(), "internal error")
	})
}
