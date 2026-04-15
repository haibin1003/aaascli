package api

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func newTestClient(serverURL string) *Client {
	c := NewClient("test-token", true)
	c.BaseURL = serverURL
	return c
}

func newTestClientWithCrypto(serverURL string) *Client {
	c := newTestClient(serverURL)
	pair, _ := GenerateRSAKeyPair()
	c.Crypto.LocalKeyPair = pair
	c.Crypto.PlatformPubKey = pair.PublicKey
	c.Crypto.PubKeyExpireAt = time.Now().Add(time.Hour)
	return c
}

func TestAbilityService_ListAll(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assertEqual(t, "/openportalsrv/rest/portalmain/gwCapacityMgr/qryGwCapacityCatalogList", r.URL.Path)
			w.Write([]byte(`{"code":"00000","data":{"capacityCatalogList":[{"catalogName":"大数据","childList":[{"catalogName":"分析","capacityList":[{"capacityId":"A1","capacityName":"客流分析","capacityUniCode":"CODE1","capacityProviderName":"移动"}]}]}]}}`))
		}))
		defer server.Close()

		svc := NewAbilityService(newTestClient(server.URL))
		list, err := svc.ListAll()
		assertNoError(t, err)
		assertEqual(t, 1, len(list))
		assertEqual(t, "A1", list[0].ID)
		assertEqual(t, "分析", list[0].CatalogName)
	})

	t.Run("api error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"code":"99999","msg":"系统繁忙"}`))
		}))
		defer server.Close()

		svc := NewAbilityService(newTestClient(server.URL))
		_, err := svc.ListAll()
		assertError(t, err)
		assertContains(t, err.Error(), "API error [99999]")
	})

	t.Run("empty catalog", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"code":"00000","data":{"capacityCatalogList":[]}}`))
		}))
		defer server.Close()

		svc := NewAbilityService(newTestClient(server.URL))
		list, err := svc.ListAll()
		assertNoError(t, err)
		assertEqual(t, 0, len(list))
	})
}

func TestAbilityService_List(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assertEqual(t, "/openportalsrv/rest/portalmain/productMgr/initProductList", r.URL.Path)
			assertEqual(t, http.MethodPost, r.Method)
			// body is encrypted, skip detailed assertion
			w.Write([]byte(`{"code":"00000","data":{"productActionList":[{"capacityId":"P1","capacityName":"产品1"}]}}`))
		}))
		defer server.Close()

		svc := NewAbilityService(newTestClientWithCrypto(server.URL))
		resp, err := svc.List(1, 10)
		assertNoError(t, err)
		assertEqual(t, 1, len(resp.Data.ProductActionList))
		assertEqual(t, "P1", resp.Data.ProductActionList[0].ID)
	})

	t.Run("api error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"code":"50000","msg":"内部错误"}`))
		}))
		defer server.Close()

		svc := NewAbilityService(newTestClientWithCrypto(server.URL))
		_, err := svc.List(1, 10)
		assertError(t, err)
		assertContains(t, err.Error(), "API error [50000]")
	})
}

func TestAbilityService_Search(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"code":"00000","data":{"capacityCatalogList":[{"catalogName":"短信","childList":[{"catalogName":"触达","capacityList":[{"capacityId":"S1","capacityName":"短信通知","capacityUniCode":"SMS001","capacityProviderName":"移动"},{"capacityId":"S2","capacityName":"语音通知","capacityUniCode":"VOICE001","capacityProviderName":"移动"}]}]}]}}`))
	}))
	defer server.Close()

	t.Run("match by name", func(t *testing.T) {
		svc := NewAbilityService(newTestClient(server.URL))
		list, err := svc.Search("短信")
		assertNoError(t, err)
		assertEqual(t, 1, len(list))
		assertEqual(t, "S1", list[0].ID)
	})

	t.Run("match by code", func(t *testing.T) {
		svc := NewAbilityService(newTestClient(server.URL))
		list, err := svc.Search("VOICE")
		assertNoError(t, err)
		assertEqual(t, 1, len(list))
		assertEqual(t, "S2", list[0].ID)
	})

	t.Run("no match", func(t *testing.T) {
		svc := NewAbilityService(newTestClient(server.URL))
		list, err := svc.Search("不存在")
		assertNoError(t, err)
		assertEqual(t, 0, len(list))
	})

	t.Run("empty keyword", func(t *testing.T) {
		svc := NewAbilityService(newTestClient(server.URL))
		list, err := svc.Search("")
		assertNoError(t, err)
		assertEqual(t, 0, len(list))
	})
}

func TestAbilityService_GetDetail(t *testing.T) {
	t.Run("from capacityDefineBean", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assertEqual(t, "/openportalsrv/rest/portalmain/capacityMgr/initCapacityInfo", r.URL.Path)
			body, _ := io.ReadAll(r.Body)
			assertTrue(t, strings.Contains(string(body), "AID001"))
			w.Write([]byte(`{"code":"00000","data":{"capacityDefineBean":{"capacityId":"AID001","capacityName":"定义名称","capacityProviderName":"移动"},"product":{"capacityId":"AID001","capacityName":"产品名称"}}}`))
		}))
		defer server.Close()

		svc := NewAbilityService(newTestClient(server.URL))
		detail, err := svc.GetDetail("AID001")
		assertNoError(t, err)
		assertEqual(t, "AID001", detail.ID)
		assertEqual(t, "定义名称", detail.Name)
		assertEqual(t, "移动", detail.Provider)
	})

	t.Run("fallback to product", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"code":"00000","data":{"capacityDefineBean":{"capacityId":"","capacityName":""},"product":{"capacityId":"AID002","capacityName":"产品名称"}}}`))
		}))
		defer server.Close()

		svc := NewAbilityService(newTestClient(server.URL))
		detail, err := svc.GetDetail("AID002")
		assertNoError(t, err)
		assertEqual(t, "AID002", detail.ID)
		assertEqual(t, "产品名称", detail.Name)
	})

	t.Run("api error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"code":"99999","msg":"找不到"}`))
		}))
		defer server.Close()

		svc := NewAbilityService(newTestClient(server.URL))
		_, err := svc.GetDetail("BAD")
		assertError(t, err)
	})
}

func TestAbilityService_ListServices(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assertEqual(t, "/openportalsrv/rest/portalmain/capacityMgr/queryServiceMenuList", r.URL.Path)
			body, _ := io.ReadAll(r.Body)
			assertEqual(t, "capacityId=AID001", string(body))
			w.Write([]byte(`{"code":"00000","data":{"serviceMenus":[{"id":"SM1","name":"服务1","code":"SVC1"}]}}`))
		}))
		defer server.Close()

		svc := NewAbilityService(newTestClient(server.URL))
		list, err := svc.ListServices("AID001")
		assertNoError(t, err)
		assertEqual(t, 1, len(list))
		assertEqual(t, "SM1", list[0].ID)
	})

	t.Run("api error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"code":"99999","msg":"失败"}`))
		}))
		defer server.Close()

		svc := NewAbilityService(newTestClient(server.URL))
		_, err := svc.ListServices("BAD")
		assertError(t, err)
	})
}

func TestAbilityService_OrderAbility(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/openportalsrv/rest/portalmain/capacityMgr/initCapacityInfo":
				w.Write([]byte(`{"code":"00000","data":{"capacityDefineBean":{"capacityId":"AID001","capacityName":"短信能力"}}}`))
			case "/openportalsrv/rest/portaluser/capacityOrder/orderCapacity":
				body, _ := io.ReadAll(r.Body)
				assertContains(t, string(body), `"capacityId":"AID001"`)
				assertContains(t, string(body), `"orderType":"CAPACITY"`)
				assertContains(t, string(body), `"capacityName":"短信能力"`)
				w.Write([]byte(`{"code":"00000","data":{"orderId":"ORD123"}}`))
			default:
				t.Errorf("unexpected path: %s", r.URL.Path)
			}
		}))
		defer server.Close()

		svc := NewAbilityService(newTestClient(server.URL))
		resp, err := svc.OrderAbility(&OrderRequest{
			AbilityID: "AID001",
			AppID:     "APP001",
			Period:    "1年",
		})
		assertNoError(t, err)
		assertEqual(t, "ORD123", resp.Data.OrderID)
	})

	t.Run("detail api error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"code":"99999","msg":"找不到能力"}`))
		}))
		defer server.Close()

		svc := NewAbilityService(newTestClient(server.URL))
		_, err := svc.OrderAbility(&OrderRequest{AbilityID: "BAD"})
		assertError(t, err)
		assertContains(t, err.Error(), "获取能力详情失败")
	})

	t.Run("order api error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/openportalsrv/rest/portalmain/capacityMgr/initCapacityInfo":
				w.Write([]byte(`{"code":"00000","data":{"capacityDefineBean":{"capacityId":"AID001","capacityName":"短信能力"}}}`))
			case "/openportalsrv/rest/portaluser/capacityOrder/orderCapacity":
				w.Write([]byte(`{"code":"99999","msg":"订购失败"}`))
			}
		}))
		defer server.Close()

		svc := NewAbilityService(newTestClient(server.URL))
		_, err := svc.OrderAbility(&OrderRequest{AbilityID: "AID001"})
		assertError(t, err)
		assertContains(t, err.Error(), "API error [99999]")
	})
}

func TestContainsAny(t *testing.T) {
	tests := []struct {
		s        string
		substr   string
		expected bool
	}{
		{"Hello World", "Hello", true},
		{"Hello World", "World", true},
		{"Hello", "World", false},
		{"", "abc", false},
		{"abc", "", false},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_%s", tt.s, tt.substr), func(t *testing.T) {
			assertEqual(t, tt.expected, containsAny(tt.s, tt.substr))
		})
	}
}

func TestContainsCI(t *testing.T) {
	tests := []struct {
		a        string
		b        string
		expected bool
	}{
		{"abcdef", "cde", true},
		{"abcdef", "xyz", false},
		{"abc", "abcdef", false},
		{"", "a", false},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_%s", tt.a, tt.b), func(t *testing.T) {
			assertEqual(t, tt.expected, containsCI(tt.a, tt.b))
		})
	}
}
