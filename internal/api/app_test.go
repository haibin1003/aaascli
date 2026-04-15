package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAppService_List(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assertEqual(t, "/openportalsrv/rest/portalmain/capacityMgr/qryAuthorizedList", r.URL.Path)
			w.Write([]byte(`{"code":"00000","data":{"authorizedList":[{"appName":"App1","abilityName":"短信能力","authStatus":"1","authStatusName":"已授权"}]}}`))
		}))
		defer server.Close()

		svc := NewAppService(newTestClient(server.URL))
		resp, err := svc.List()
		assertNoError(t, err)
		assertEqual(t, 1, len(resp.Data.AuthorizedList))
		assertEqual(t, "App1", resp.Data.AuthorizedList[0].AppName)
		assertEqual(t, "已授权", resp.Data.AuthorizedList[0].AuthStatusName)
	})

	t.Run("api error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"code":"99999","msg":"查询失败"}`))
		}))
		defer server.Close()

		svc := NewAppService(newTestClient(server.URL))
		_, err := svc.List()
		assertError(t, err)
		assertContains(t, err.Error(), "API error [99999]")
	})
}

func TestAppService_ListMyApps(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assertEqual(t, "/openportalsrv/rest/portaluser/appManager/qryMyAppList", r.URL.Path)
			body, _ := io.ReadAll(r.Body)
			assertEqual(t, "pgnum=1&pgsize=20&appName=测试", string(body))
			w.Write([]byte(`{"code":"00000","data":{"appList":{"pageNum":1,"pageSize":20,"total":1,"pages":1,"list":[{"appId":"A1","appName":"测试应用","status":"1"}]}}}`))
		}))
		defer server.Close()

		svc := NewAppService(newTestClient(server.URL))
		list, err := svc.ListMyApps(1, 20, "测试")
		assertNoError(t, err)
		assertEqual(t, 1, len(list))
		assertEqual(t, "A1", list[0].AppID)
		assertEqual(t, "测试应用", list[0].AppName)
	})

	t.Run("default pagination", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			assertEqual(t, "pgnum=1&pgsize=10&appName=", string(body))
			w.Write([]byte(`{"code":"00000","data":{"appList":{"list":[]}}}`))
		}))
		defer server.Close()

		svc := NewAppService(newTestClient(server.URL))
		list, err := svc.ListMyApps(0, 0, "")
		assertNoError(t, err)
		assertEqual(t, 0, len(list))
	})

	t.Run("api error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"code":"99999","msg":"失败"}`))
		}))
		defer server.Close()

		svc := NewAppService(newTestClient(server.URL))
		_, err := svc.ListMyApps(1, 10, "")
		assertError(t, err)
	})
}

func TestAppService_AuthAbility(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assertEqual(t, "/openportalsrv/rest/portaluser/appManager/doAppOeder", r.URL.Path)
			body, _ := io.ReadAll(r.Body)
			bodyStr := string(body)
			assertContains(t, bodyStr, `"appId":"APP001"`)
			assertContains(t, bodyStr, `"newOrderedGoodList":["ABI001"]`)
			assertContains(t, bodyStr, `"authType":"capacity"`)
			assertContains(t, bodyStr, `"status":"AppStatusOnline"`)
			assertContains(t, bodyStr, `"bomcId":"BOMC456"`)
			w.Write([]byte(`{"code":"00000","data":{"orderId":"ORD789"}}`))
		}))
		defer server.Close()

		svc := NewAppService(newTestClient(server.URL))
		resp, err := svc.AuthAbility(&AuthAbilityRequest{
			AppID:          "APP001",
			AbilityID:      "ABI001",
			AppName:        "测试应用",
			AuthType:       "capacity",
			Status:         "AppStatusOnline",
			BomcID:         "BOMC456",
			QuotaLimit:     "500",
			LimitCount:     "500",
			PolicyPeriod:   "1",
			PolicyTimeUnit: "second",
			GoodsNames:     "短信能力",
		})
		assertNoError(t, err)
		assertEqual(t, "ORD789", resp.Data.OrderID)
	})

	t.Run("api error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"code":"99999","msg":"授权失败"}`))
		}))
		defer server.Close()

		svc := NewAppService(newTestClient(server.URL))
		_, err := svc.AuthAbility(&AuthAbilityRequest{AppID: "BAD"})
		assertError(t, err)
		assertContains(t, err.Error(), "API error [99999]")
	})
}
