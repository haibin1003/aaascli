package api

// abilityOrderPayload 能力订购请求体
type abilityOrderPayload struct {
	CapacityID   string `json:"capacityId"`
	OrderType    string `json:"orderType"`
	CapacityName string `json:"capacityName"`
}

// appAuthAbilityPayload 应用授权能力请求体
type appAuthAbilityPayload struct {
	AppID              string              `json:"appId"`
	OrderedGoodList    []string            `json:"orderedGoodList"`
	NewOrderedGoodList []string            `json:"newOrderedGoodList"`
	CrmOrderList       []string            `json:"crmOrderList"`
	Status             string              `json:"status"`
	AuthType           string              `json:"authType"`
	BomcID             string              `json:"bomcId"`
	LimitAndQuotaData  []limitAndQuotaItem `json:"limitAndQuotaData"`
	AppNames           string              `json:"appNames"`
	GoodsNames         string              `json:"goodsNames"`
}

// serviceOrderPayload 服务订购请求体
type serviceOrderPayload struct {
	AppID             string              `json:"appId"`
	OrderedAppList    []string            `json:"orderedAppList"`
	NewOrderedAppList []string            `json:"newOrderedAppList"`
	APIName           string              `json:"apiName"`
	DomainID          string              `json:"domainId"`
	MaxVersion        string              `json:"maxVersion"`
	AuthType          string              `json:"authType"`
	InterfaceID       string              `json:"interfaceId"`
	AppName           string              `json:"appName"`
	BomcID            string              `json:"bomcId"`
	LimitAndQuotaData []limitAndQuotaItem `json:"limitAndQuotaData"`
	AppNames          string              `json:"appNames"`
	GoodsNames        string              `json:"goodsNames"`
}

// limitAndQuotaItem 流控配额项
type limitAndQuotaItem struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	AppID          string `json:"appId"`
	AppName        string `json:"appName"`
	Type           string `json:"type"`
	QuotaLimit     string `json:"quotaLimit"`
	LimitCount     string `json:"limitCount"`
	PolicyPeriod   string `json:"policyPeriod"`
	PolicyTimeUnit string `json:"policyTimeUnit"`
}
