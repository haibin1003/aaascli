package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestApplyService_ListMyApplies(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assertEqual(t, "/openportalsrv/rest/portaluser/myApply/queryMyApplyList", r.URL.Path)
			body, _ := io.ReadAll(r.Body)
			assertEqual(t, "passStatus=true&pgnum=1&pgsize=20", string(body))
			w.Write([]byte(`{"code":"00000","data":{"pageNum":1,"pageSize":20,"total":2,"pages":1,"list":[{"id":"A1","appName":"App1","goodsName":"短信能力","authType":"capacity","authTypeName":"能力","status":"1","statusName":"已通过","applyTime":"2026-04-01","passStatus":"true"}]}}`))
		}))
		defer server.Close()

		svc := NewApplyService(newTestClient(server.URL))
		resp, err := svc.ListMyApplies(1, 20, true)
		assertNoError(t, err)
		assertEqual(t, 1, len(resp.Data.List))
		assertEqual(t, "A1", resp.Data.List[0].ID)
		assertEqual(t, "App1", resp.Data.List[0].AppName)
		assertEqual(t, "短信能力", resp.Data.List[0].GoodsName)
		assertEqual(t, "已通过", resp.Data.List[0].StatusName)
	})

	t.Run("default pagination", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			assertEqual(t, "passStatus=false&pgnum=1&pgsize=10", string(body))
			w.Write([]byte(`{"code":"00000","data":{"pageNum":1,"pageSize":10,"total":0,"pages":0,"list":[]}}`))
		}))
		defer server.Close()

		svc := NewApplyService(newTestClient(server.URL))
		resp, err := svc.ListMyApplies(0, 0, false)
		assertNoError(t, err)
		assertEqual(t, 0, len(resp.Data.List))
		assertEqual(t, 1, resp.Data.PageNum)
		assertEqual(t, 10, resp.Data.PageSize)
	})

	t.Run("api error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"code":"99999","msg":"查询失败"}`))
		}))
		defer server.Close()

		svc := NewApplyService(newTestClient(server.URL))
		_, err := svc.ListMyApplies(1, 10, true)
		assertError(t, err)
		assertContains(t, err.Error(), "API error [99999]")
	})
}
