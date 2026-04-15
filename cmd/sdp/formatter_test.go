package cmd

import (
	"testing"

	"github.com/haibin1003/aaascli/internal/api"
)

func assertEqual(t *testing.T, expected, actual interface{}) {
	t.Helper()
	if expected != actual {
		t.Errorf("expected %v, got %v", expected, actual)
	}
}

func TestFormatAbilitySlice(t *testing.T) {
	items := []api.Ability{
		{ID: "A1", Name: "短信", Code: "SMS", Provider: "移动", CatalogName: "通信"},
		{ID: "A2", Name: "定位", Code: "LOC", Provider: "移动", CatalogName: "位置"},
	}
	result := formatAbilitySlice(items, 1, 10)
	assertEqual(t, 2, len(result["items"].([]map[string]interface{})))
	assertEqual(t, 2, result["pagination"].(map[string]interface{})["total"])
	assertEqual(t, 1, result["pagination"].(map[string]interface{})["page"])
}

func TestFormatAbilitySlicePagination(t *testing.T) {
	items := []api.Ability{
		{ID: "A1"}, {ID: "A2"}, {ID: "A3"},
	}
	result := formatAbilitySlice(items, 2, 1)
	pageItems := result["items"].([]map[string]interface{})
	assertEqual(t, 1, len(pageItems))
	assertEqual(t, "A2", pageItems[0]["id"])
}

func TestFormatAbilitySliceNil(t *testing.T) {
	result := formatAbilitySlice(nil, 1, 10)
	assertEqual(t, 0, len(result["items"].([]map[string]interface{})))
}

func TestFormatAbilityDetail(t *testing.T) {
	detail := &api.AbilityDetail{
		ID:         "A1",
		Name:       "短信能力",
		Code:       "SMS",
		DetailDesc: "详细描述",
		Provider:   "移动",
		TypeName:   "通信",
		CallType:   "REST",
		UserID:     "U1",
	}
	result := formatAbilityDetail(detail)
	assertEqual(t, "A1", result["id"])
	assertEqual(t, "详细描述", result["detailDesc"])
	assertEqual(t, "REST", result["callType"])
}

func TestFormatAbilityServiceSlice(t *testing.T) {
	items := []api.AbilityServiceItem{
		{ID: "S1", Name: "发送短信", Code: "SEND", ServiceType: "inner"},
	}
	result := formatAbilityServiceSlice(items, 1, 10)
	assertEqual(t, 1, len(result["items"].([]map[string]interface{})))
	assertEqual(t, "inner", result["items"].([]map[string]interface{})[0]["serviceType"])
}

func TestFormatAbilityProductList(t *testing.T) {
	resp := &api.AbilityListResponse{}
	resp.Data.ProductActionList = []api.Ability{
		{ID: "P1", Name: "产品1"},
	}
	result := formatAbilityProductList(resp)
	assertEqual(t, 1, len(result["items"].([]map[string]interface{})))
	assertEqual(t, 1, result["total"])
}

func TestFormatServiceDetail(t *testing.T) {
	detail := &api.ServiceDetail{
		APIID:           "SE1",
		Name:            "定位",
		APIVersion:      "v1",
		RequestTypeText: "POST",
		RequestURL:      "/loc",
		Owner:           "张三",
	}
	result := formatServiceDetail(detail)
	assertEqual(t, "SE1", result["id"])
	assertEqual(t, "POST", result["requestTypeText"])
	assertEqual(t, "张三", result["owner"])
}

func TestFlattenCatalogNodes(t *testing.T) {
	nodes := []api.CatalogNode{
		{
			CatalogID:    "C1",
			CatalogName:  "根目录",
			CatalogType:  "API",
			CatalogLevel: "1",
			IsLeaf:       "false",
			SmallCatalogList: []api.CatalogNode{
				{
					CatalogID:   "C2",
					CatalogName: "子目录",
					IsLeaf:      "true",
					APIList: []api.APIService{
						{APIID: "API1", Name: "接口1", InterfaceID: "IF1", RequestType: "POST", RequestURL: "/api1"},
					},
				},
			},
		},
	}
	result := flattenCatalogNodes(nodes, 0)
	assertEqual(t, 3, len(result))
	assertEqual(t, "catalog", result[0]["type"])
	assertEqual(t, 0, result[0]["depth"])
	assertEqual(t, "leaf-catalog", result[1]["type"])
	assertEqual(t, 1, result[1]["depth"])
	assertEqual(t, "api", result[2]["type"])
	assertEqual(t, 2, result[2]["depth"])
	assertEqual(t, "IF1", result[2]["code"])
}

func TestServiceContainsCI(t *testing.T) {
	assertEqual(t, true, serviceContainsCI("Hello World", "World"))
	assertEqual(t, false, serviceContainsCI("Hello", "World"))
	assertEqual(t, false, serviceContainsCI("Hi", "Hello"))
}

func TestFormatServiceSlice(t *testing.T) {
	items := []map[string]interface{}{
		{"id": "S1", "name": "服务1"},
		{"id": "S2", "name": "服务2"},
	}
	result := formatServiceSlice(items, 1, 1)
	assertEqual(t, 1, len(result["items"].([]map[string]interface{})))
	assertEqual(t, 2, result["pagination"].(map[string]interface{})["total"])
	assertEqual(t, 2, result["pagination"].(map[string]interface{})["pages"])
}

func TestFormatApplyList(t *testing.T) {
	resp := &api.ApplyListResponse{}
	resp.Data.PageNum = 1
	resp.Data.PageSize = 10
	resp.Data.Total = 1
	resp.Data.Pages = 1
	resp.Data.List = []api.ApplyItem{
		{ID: "AP1", AppName: "App1", GoodsName: "能力1", StatusName: "已通过"},
	}
	result := formatApplyList(resp)
	assertEqual(t, 1, len(result["items"].([]map[string]interface{})))
	assertEqual(t, 1, result["pagination"].(map[string]interface{})["total"])
}

func TestFormatMyAppList(t *testing.T) {
	items := []api.MyApp{
		{AppID: "A1", AppName: "App1", Status: "1", MaxQuotaNum: 100},
	}
	result := formatMyAppList(items, 1, 10)
	assertEqual(t, 1, len(result["items"].([]map[string]interface{})))
	assertEqual(t, 100, result["items"].([]map[string]interface{})[0]["maxQuotaNum"])
}

func TestFormatAppAuthList(t *testing.T) {
	resp := &api.AppAuthListResponse{}
	resp.Data.AuthorizedList = []api.AppAuth{
		{AppName: "App1", AbilityName: "短信", AuthStatusName: "已授权"},
	}
	result := formatAppAuthList(resp)
	assertEqual(t, 1, len(result["items"].([]map[string]interface{})))
	assertEqual(t, 1, result["total"])
}
