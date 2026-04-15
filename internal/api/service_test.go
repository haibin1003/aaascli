package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestServiceService_ListAll(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assertEqual(t, "/openportalsrv/rest/portalmain/capacityMgr/qryApiCatalogListData", r.URL.Path)
			body, _ := io.ReadAll(r.Body)
			assertEqual(t, "parentId=APISHOWROOT&catalogType=APISHOW", string(body))
			w.Write([]byte(`{"code":"00000","data":{"cataLogList":[{"catalogId":"C1","catalogName":"短信","smallCatalogList":[{"catalogId":"C2","catalogName":"云信","apiList":[{"apiID":"API1","name":"短信发送","requestUrl":"/sms/send"}]}]}]}}`))
		}))
		defer server.Close()

		svc := NewServiceService(newTestClient(server.URL))
		resp, err := svc.ListAll()
		assertNoError(t, err)
		assertEqual(t, 1, len(resp.Data.CataLogList))
		assertEqual(t, "短信", resp.Data.CataLogList[0].CatalogName)
	})

	t.Run("api error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"code":"99999","msg":"系统错误"}`))
		}))
		defer server.Close()

		svc := NewServiceService(newTestClient(server.URL))
		_, err := svc.ListAll()
		assertError(t, err)
		assertContains(t, err.Error(), "API error [99999]")
	})
}

func TestServiceService_GetDetail(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assertEqual(t, "/openportalsrv/rest/portalmain/capacityMgr/queryServiceInfo", r.URL.Path)
			body, _ := io.ReadAll(r.Body)
			assertTrue(t, strings.Contains(string(body), "serviceId=SE123"))
			w.Write([]byte(`{"code":"00000","data":{"serviceInfo":{"apiId":"SE123","name":"定位服务","requestTypeText":"POST","requestUrl":"/loc/query","owner":"张三"}}}`))
		}))
		defer server.Close()

		svc := NewServiceService(newTestClient(server.URL))
		detail, err := svc.GetDetail("SE123")
		assertNoError(t, err)
		assertEqual(t, "SE123", detail.APIID)
		assertEqual(t, "定位服务", detail.Name)
		assertEqual(t, "POST", detail.RequestTypeText)
		assertEqual(t, "张三", detail.Owner)
	})

	t.Run("api error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"code":"99999","msg":"找不到服务"}`))
		}))
		defer server.Close()

		svc := NewServiceService(newTestClient(server.URL))
		_, err := svc.GetDetail("BAD")
		assertError(t, err)
		assertContains(t, err.Error(), "API error [99999]")
	})
}

func TestServiceService_OrderService(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assertEqual(t, "/openportalsrv/rest/portaluser/myOrder/doAppOrder", r.URL.Path)
			assertEqual(t, "application/json", r.Header.Get("Content-Type"))
			body, _ := io.ReadAll(r.Body)
			bodyStr := string(body)
			assertContains(t, bodyStr, `"appId":"SE001"`)
			assertContains(t, bodyStr, `"newOrderedAppList":["APP001"]`)
			assertContains(t, bodyStr, `"apiName":"短信API"`)
			assertContains(t, bodyStr, `"authType":"api"`)
			assertContains(t, bodyStr, `"bomcId":"BOMC123"`)
			w.Write([]byte(`{"code":"00000","data":{"orderId":"ORD456"}}`))
		}))
		defer server.Close()

		svc := NewServiceService(newTestClient(server.URL))
		resp, err := svc.OrderService(&OrderServiceRequest{
			ServiceID:      "SE001",
			AppID:          "APP001",
			AppName:        "测试应用",
			APIName:        "短信API",
			DomainID:       "D1",
			MaxVersion:     "v1",
			AuthType:       "api",
			InterfaceID:    "IF001",
			BomcID:         "BOMC123",
			QuotaLimit:     "100",
			LimitCount:     "100",
			PolicyPeriod:   "1",
			PolicyTimeUnit: "second",
			GoodsNames:     "短信服务",
		})
		assertNoError(t, err)
		assertEqual(t, "ORD456", resp.Data.OrderID)
	})

	t.Run("api error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"code":"99999","msg":"订购失败"}`))
		}))
		defer server.Close()

		svc := NewServiceService(newTestClient(server.URL))
		_, err := svc.OrderService(&OrderServiceRequest{
			ServiceID: "SE001",
			AppID:     "APP001",
		})
		assertError(t, err)
		assertContains(t, err.Error(), "API error [99999]")
	})
}
