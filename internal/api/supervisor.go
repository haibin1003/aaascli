package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"regexp"
	"strconv"
	"time"
)

// SupervisorService 监管模块API服务
// 所有supervisor下的子命令都需要先通过Authenticate()获取认证，然后才能访问其他API
type SupervisorService struct {
	BaseService
	// authToken 通过认证获取的Authentication token，用于后续请求
	authToken string
}

// SupervisorBaseURL 监管平台基础URL
const SupervisorBaseURL = "http://4c.hq.cmcc"

// MOSSAuthURL MOSS认证跳转URL
const MOSSAuthURL = "http://4c.hq.cmcc/moss/web/auth/v1/user/oauth/authorize/jump-url"

// SupervisorAuthURL 监管平台认证URL
const SupervisorAuthURL = "http://4c.hq.cmcc/supervision/system-api/business-accept/supervision/authorize"

// NewSupervisorService 创建新的监管平台服务实例
// 注意：使用服务前必须先调用Authenticate()进行认证
func NewSupervisorService(cookie string, client HTTPClient) *SupervisorService {
	headers := map[string]string{
		"Accept":             "application/json, text/plain, */*",
		"Accept-Language":    "zh-CN,zh-TW;q=0.9,zh;q=0.8,en-US;q=0.7,en;q=0.6",
		"Cache-Control":      "no-cache",
		"Connection":         "keep-alive",
		"Content-Type":       "application/json",
		"Origin":             SupervisorBaseURL,
		"Pragma":             "no-cache",
		"User-Agent":         "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.6478.251 Safari/537.36 UOS Professional",
	}

	if cookie != "" {
		headers["Cookie"] = cookie
	}

	return &SupervisorService{
		BaseService: BaseService{
			BaseURL:    SupervisorBaseURL,
			Headers:    headers,
			HTTPClient: client,
		},
	}
}

// JumpURLResponse MOSS认证跳转URL响应
type JumpURLResponse struct {
	Head struct {
		RequestID  string `json:"requestId"`
		RespStatus string `json:"respStatus"`
		RespCode   string `json:"respCode"`
		RespDesc   string `json:"respDesc"`
	} `json:"head"`
	Data string `json:"data"` // 跳转URL，包含code参数
}

// AuthorizeResponse 监管平台认证响应
type AuthorizeResponse struct {
	Result   interface{} `json:"result"`
	Total    int         `json:"total"`
	RespCode string      `json:"respCode"`
	RespDesc string      `json:"respDesc"`
}

// Authenticate 执行完整的认证流程
// 步骤1: 请求MOSS认证跳转URL获取code
// 步骤2: 使用code请求监管平台认证接口获取Authentication token
// 返回获取到的authToken和可能的错误
//
// 所有supervisor子命令在访问API前都必须调用此方法：
//   svc := api.NewSupervisorService(cookie, client)
//   authToken, err := svc.Authenticate()
//   if err != nil { ... }
//   // 然后使用authToken访问其他API
func (s *SupervisorService) Authenticate() (string, error) {
	// 步骤1: 获取跳转URL和code
	code, err := s.getAuthCode()
	if err != nil {
		return "", fmt.Errorf("获取认证code失败: %w", err)
	}

	// 步骤2: 使用code换取Authentication token
	authToken, err := s.exchangeCodeForToken(code)
	if err != nil {
		return "", fmt.Errorf("换取认证token失败: %w", err)
	}

	s.authToken = authToken
	return authToken, nil
}

// getAuthCode 请求MOSS认证接口获取跳转URL并提取code
func (s *SupervisorService) getAuthCode() (string, error) {
	requestBody := map[string]interface{}{
		"head": map[string]interface{}{
			"requestId": generateSupervisorRequestID(),
			"sourceid":  "7923641850",
		},
		"data": map[string]interface{}{
			"redirectUri":  SupervisorBaseURL + "/supervision/#/plugin-auth",
			"responseType": "code",
		},
	}

	req := &Request{
		URL:    MOSSAuthURL,
		Method: "POST",
		Headers: map[string]string{
			"Accept":       "*/*",
			"Content-Type": "application/json;charset=UTF-8",
			"Origin":       SupervisorBaseURL,
			"Referer":      SupervisorBaseURL + "/moss/micrologin/",
			"Cookie":       s.Headers["Cookie"],
		},
		Body: requestBody,
	}

	resp, err := s.HTTPClient.Send(req)
	if err != nil {
		return "", fmt.Errorf("请求认证跳转URL失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	var jumpResp JumpURLResponse
	if err := json.Unmarshal(body, &jumpResp); err != nil {
		return "", fmt.Errorf("解析响应失败: %w, body: %s", err, string(body))
	}

	if jumpResp.Head.RespCode != "00" {
		return "", fmt.Errorf("认证失败: %s", jumpResp.Head.RespDesc)
	}

	// 从跳转URL中提取code
	code, err := extractCodeFromURL(jumpResp.Data)
	if err != nil {
		return "", fmt.Errorf("提取code失败: %w", err)
	}

	return code, nil
}

// exchangeCodeForToken 使用code换取Authentication token
func (s *SupervisorService) exchangeCodeForToken(code string) (string, error) {
	redirectURL := url.QueryEscape(SupervisorBaseURL + "/supervision/#/plugin-auth")

	requestBody := map[string]interface{}{
		"tenantId":    "01",
		"code":        code,
		"mocked":      false,
		"redirectUrl": redirectURL,
	}

	req := &Request{
		URL:    SupervisorAuthURL,
		Method: "POST",
		Headers: map[string]string{
			"Accept":       "application/json, text/plain, */*",
			"Content-Type": "application/json",
			"Origin":       SupervisorBaseURL,
			"Referer":      SupervisorBaseURL + "/supervision/",
			"Cookie":       s.Headers["Cookie"],
		},
		Body: requestBody,
	}

	resp, err := s.HTTPClient.Send(req)
	if err != nil {
		return "", fmt.Errorf("请求认证接口失败: %w", err)
	}
	defer resp.Body.Close()

	// 从响应头中获取Set-Cookie中的Authentication值
	cookies := resp.Header.Values("Set-Cookie")
	for _, cookie := range cookies {
		authToken := extractAuthFromCookie(cookie)
		if authToken != "" {
			return authToken, nil
		}
	}

	return "", fmt.Errorf("未在响应中找到Authentication cookie")
}

// GetAuthHeaders 获取包含认证信息的请求头
// 必须在Authenticate()成功后调用
func (s *SupervisorService) GetAuthHeaders() (map[string]string, error) {
	if s.authToken == "" {
		return nil, fmt.Errorf("未获取认证token，请先调用Authenticate()")
	}

	headers := make(map[string]string)
	for k, v := range s.Headers {
		headers[k] = v
	}
	headers["Authentication"] = s.authToken
	headers["Referer"] = SupervisorBaseURL + "/supervision/"

	return headers, nil
}

// extractCodeFromURL 从跳转URL中提取code参数
func extractCodeFromURL(jumpURL string) (string, error) {
	// URL格式: http://4c.hq.cmcc/supervision/#/plugin-auth?code=xxx
	parsedURL, err := url.Parse(jumpURL)
	if err != nil {
		return "", err
	}

	// 解析fragment中的query参数
	fragment := parsedURL.Fragment
	if fragment == "" {
		return "", fmt.Errorf("URL中没有fragment")
	}

	// fragment格式: /plugin-auth?code=xxx
	re := regexp.MustCompile(`code=([^&]+)`)
	matches := re.FindStringSubmatch(fragment)
	if len(matches) < 2 {
		return "", fmt.Errorf("未在URL中找到code参数")
	}

	return matches[1], nil
}

// extractAuthFromCookie 从Set-Cookie头中提取Authentication值
func extractAuthFromCookie(cookieHeader string) string {
	// Cookie格式: Authentication=xxx; Max-Age=86400; ...
	re := regexp.MustCompile(`Authentication=([^;]+)`)
	matches := re.FindStringSubmatch(cookieHeader)
	if len(matches) < 2 {
		return ""
	}
	return matches[1]
}

// generateSupervisorRequestID 生成监管平台请求ID
// 使用与doc.go中generateRequestID相同的逻辑
func generateSupervisorRequestID() string {
	return strconv.FormatInt(generateRequestID(), 10) + generateRandomString(20)
}

// generateRandomString 生成指定长度的随机字符串
func generateRandomString(length int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(result)
}

// MainOverviewResponse 主面板概览响应
type MainOverviewResponse struct {
	Result struct {
		Stat []StatItem `json:"stat"`
	} `json:"result"`
	Total    int    `json:"total"`
	RespCode string `json:"respCode"`
	RespDesc string `json:"respDesc"`
}

// StatItem 统计项
type StatItem struct {
	Label  string      `json:"label"`  // 统计项名称，如"组内待认领工单"、"我的待办"
	Values []StatValue `json:"values"` // 具体统计值列表
}

// StatValue 统计值
type StatValue struct {
	Label string `json:"label"` // 类型名称，如"问题咨询"、"普通投诉"
	Value string `json:"value"` // 数量值
}

// GetMainOverview 获取主面板概览数据（待办数量等）
// 注意：调用前必须先执行Authenticate()获取认证
func (s *SupervisorService) GetMainOverview() (*MainOverviewResponse, error) {
	headers, err := s.GetAuthHeaders()
	if err != nil {
		return nil, err
	}

	req := &Request{
		URL:     SupervisorBaseURL + "/supervision/system-api/business-accept/mainBoardController/getMainOverview",
		Method:  "POST",
		Headers: headers,
		Body:    map[string]interface{}{},
	}

	resp, err := s.HTTPClient.Send(req)
	if err != nil {
		return nil, fmt.Errorf("请求主面板概览失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var overviewResp MainOverviewResponse
	if err := json.Unmarshal(body, &overviewResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w, body: %s", err, string(body))
	}

	if overviewResp.RespCode != "0000" {
		return nil, fmt.Errorf("获取主面板概览失败: %s", overviewResp.RespDesc)
	}

	return &overviewResp, nil
}

// GroupListResponse 业务分组列表响应
type GroupListResponse struct {
	Result struct {
		Total   int     `json:"total"`
		Size    int     `json:"size"`
		Current int     `json:"current"`
		Records []Group `json:"records"`
	} `json:"result"`
	Total    int    `json:"total"`
	RespCode string `json:"respCode"`
	RespDesc string `json:"respDesc"`
}

// Group 业务分组信息
type Group struct {
	ID             int    `json:"id"`
	Code           string `json:"code"`
	Name           string `json:"name"`
	DefaultFlag    string `json:"defaultFlag"`
	Label          string `json:"label"`
	DeleteFlag     int    `json:"deleteFlag"`
	TenantId       string `json:"tenantId"`
	TenantName     string `json:"tenantName"`
	CreateStaffId  string `json:"createStaffId"`
	CreateStaffName string `json:"createStaffName"`
	ModifyStaffId  string `json:"modifyStaffId"`
	ModifyStaffName string `json:"modifyStaffName"`
	CreateTime     string `json:"createTime"`
	ModifyTime     string `json:"modifyTime"`
	Description    string `json:"description,omitempty"`
	Remark         string `json:"remark,omitempty"`
}

// GetAllGroups 获取全量业务分组列表
// 注意：调用前必须先执行Authenticate()获取认证
func (s *SupervisorService) GetAllGroups() (*GroupListResponse, error) {
	headers, err := s.GetAuthHeaders()
	if err != nil {
		return nil, err
	}

	// 请求全量数据，使用最大页大小
	requestBody := map[string]interface{}{
		"current":  1,
		"pageSize": 10000,
		"tenantId": "01",
	}

	req := &Request{
		URL:     SupervisorBaseURL + "/supervision/system-api/business-accept/group/page",
		Method:  "POST",
		Headers: headers,
		Body:    requestBody,
	}

	resp, err := s.HTTPClient.Send(req)
	if err != nil {
		return nil, fmt.Errorf("请求业务分组列表失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var groupResp GroupListResponse
	if err := json.Unmarshal(body, &groupResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w, body: %s", err, string(body))
	}

	if groupResp.RespCode != "0000" {
		return nil, fmt.Errorf("获取业务分组列表失败: %s", groupResp.RespDesc)
	}

	return &groupResp, nil
}

// GetGroupList 获取业务分组列表（支持分页和筛选）
// 注意：调用前必须先执行Authenticate()获取认证
func (s *SupervisorService) GetGroupList(current, pageSize int, code, name, description string) (*GroupListResponse, error) {
	headers, err := s.GetAuthHeaders()
	if err != nil {
		return nil, err
	}

	// 构建请求参数
	requestBody := map[string]interface{}{
		"current":     current,
		"pageSize":    pageSize,
		"code":        code,
		"name":        name,
		"description": description,
	}

	req := &Request{
		URL:     SupervisorBaseURL + "/supervision/system-api/business-accept/group/page",
		Method:  "POST",
		Headers: headers,
		Body:    requestBody,
	}

	resp, err := s.HTTPClient.Send(req)
	if err != nil {
		return nil, fmt.Errorf("请求业务分组列表失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var groupResp GroupListResponse
	if err := json.Unmarshal(body, &groupResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w, body: %s", err, string(body))
	}

	if groupResp.RespCode != "0000" {
		return nil, fmt.Errorf("获取业务分组列表失败: %s", groupResp.RespDesc)
	}

	return &groupResp, nil
}

// GroupMemberByGroupCode 工作组成员信息（通过groupCode查询）
type GroupMemberByGroupCode struct {
	ID         int    `json:"id"`
	NickName   string `json:"nickName"`
	Phone      string `json:"phone"`
	LeaderFlag string `json:"leaderFlag"`
}

// GroupMemberByGroupCodeResponse 工作组成员列表响应（通过groupCode查询）
type GroupMemberByGroupCodeResponse struct {
	Result   []GroupMemberByGroupCode `json:"result"`
	Total    int                      `json:"total"`
	RespCode string                   `json:"respCode"`
	RespDesc string                   `json:"respDesc"`
}

// GetGroupMembersByCode 获取工作组成员列表（通过工作组编码）
// 注意：调用前必须先执行Authenticate()获取认证
func (s *SupervisorService) GetGroupMembersByCode(groupCode string) (*GroupMemberByGroupCodeResponse, error) {
	headers, err := s.GetAuthHeaders()
	if err != nil {
		return nil, err
	}

	// 构建请求参数
	requestBody := map[string]interface{}{
		"groupCode": groupCode,
	}

	req := &Request{
		URL:     SupervisorBaseURL + "/supervision/system-api/business-accept/group/listUserListByGroupCode",
		Method:  "POST",
		Headers: headers,
		Body:    requestBody,
	}

	resp, err := s.HTTPClient.Send(req)
	if err != nil {
		return nil, fmt.Errorf("请求工作组成员列表失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var memberResp GroupMemberByGroupCodeResponse
	if err := json.Unmarshal(body, &memberResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w, body: %s", err, string(body))
	}

	if memberResp.RespCode != "0000" {
		return nil, fmt.Errorf("获取工作组成员列表失败: %s", memberResp.RespDesc)
	}

	return &memberResp, nil
}

// GetGroupMembers 获取工作组成员列表（保留原函数，但使用groupCode）
// 注意：调用前必须先执行Authenticate()获取认证
func (s *SupervisorService) GetGroupMembers(deptId string) (*GroupMemberByGroupCodeResponse, error) {
	// 将deptId作为groupCode使用
	return s.GetGroupMembersByCode(deptId)
}

// WorkOrderListResponse 工单列表响应
type WorkOrderListResponse struct {
	Result struct {
		Total   int        `json:"total"`
		Size    int        `json:"size"`
		Current int        `json:"current"`
		Records []WorkOrder `json:"records"`
	} `json:"result"`
	Total    int    `json:"total"`
	RespCode string `json:"respCode"`
	RespDesc string `json:"respDesc"`
}

// WorkOrder 工单信息
type WorkOrder struct {
	ID                   int    `json:"id"`
	Code                 string `json:"code"`
	ComplaintStaffName   string `json:"complaintStaffName"`
	ProvinceName         string `json:"provinceName"`
	ProvinceCode         string `json:"provinceCode"`
	ComplaintTypeName    string `json:"complaintTypeName"`
	ComplaintTypeCode    string `json:"complaintTypeCode"`
	ComplaintTypeProperty string `json:"complaintTypeProperty"`
	CreateTime           string `json:"createTime"`
	ModifyTime           string `json:"modifyTime"`
	AcceptTime           string `json:"acceptTime"`
	HandleTime           string `json:"handleTime,omitempty"`
	Status               string `json:"status"`
	MenuName             string `json:"menuName"`
	ModuleName           string `json:"moduleName"`
	ModuleCode           string `json:"moduleCode"`
	WorkorderContent     string `json:"workorderContent"`
	PlatformName         string `json:"platformName"`
	PlatformCode         string `json:"platformCode"`
	AcceptGroupName      string `json:"acceptGroupName"`
	AcceptGroupCode      string `json:"acceptGroupCode"`
	AcceptStaffId        string `json:"acceptStaffId,omitempty"`
	AcceptStaffName      string `json:"acceptStaffName,omitempty"`
	HandleStaffId        string `json:"handleStaffId,omitempty"`
	HandleStaffName      string `json:"handleStaffName,omitempty"`
	HistoryHandleStaffId string `json:"historyHandleStaffId,omitempty"`
	HandleDescription    string `json:"handleDescription,omitempty"`
	SatisfactionValue   string `json:"satisfactionValue"`
	SatisfactionResultId string `json:"satisfactionResultId,omitempty"`
	MoreTimeInfo         string `json:"moreTimeInfo"`
	MoreTimeStatus       string `json:"moreTimeStatus"`
	MenuUrl              string `json:"menuUrl"`
	ChildModuleName      string `json:"childModuleName"`
	ChildModuleCode      string `json:"childModuleCode"`
	MenuCode             string `json:"menuCode"`
	InitTenantId         string `json:"initTenantId"`
	InitTenantName       string `json:"initTenantName"`
	TenantId             string `json:"tenantId"`
	TenantName           string `json:"tenantName"`
	LastTenantId         string `json:"lastTenantId,omitempty"`
	LastTenantName       string `json:"lastTenantName,omitempty"`
	ComplaintStaffMobile string `json:"complaintStaffMobile"`
	ComplaintStaffEmail  string `json:"complaintStaffEmail,omitempty"`
	SatStatus            string `json:"satStatus"`
	TransferTime         string `json:"transferTime,omitempty"`
	TransferTenantStaffIds string `json:"transferTenantStaffIds,omitempty"`
	CreateStaffId        string `json:"createStaffId"`
	CreateStaffName      string `json:"createStaffName"`
	KcbStatus            string `json:"kcbStatus"`
	ErrorMsg             string `json:"errorMsg,omitempty"`
}

// GetWorkOrderList 获取工单列表（支持分页和筛选）
// 注意：调用前必须先执行Authenticate()获取认证
func (s *SupervisorService) GetWorkOrderList(current, pageSize int, params map[string]string) (*WorkOrderListResponse, error) {
	headers, err := s.GetAuthHeaders()
	if err != nil {
		return nil, err
	}

	// 默认请求参数
	requestBody := map[string]interface{}{
		"current":             current,
		"pageSize":            pageSize,
		"code":                "",
		"moduleCode":          "",
		"menuCode":            "",
		"tenantId":            "01",
		"type":                "",
		"status":              "",
		"complaintTypeCode":   "",
		"acceptGroupCode":     "",
		"historyHandleStaffName": "",
		"complaintStaffName":   "",
		"complaintStaffMobile": "",
		"crossTenantStatus":   "",
		"nowHandleStaffName":  "",
		"queryType":           "LOCAL_ALL",
	}

	// 覆盖用户指定的筛选参数
	for k, v := range params {
		requestBody[k] = v
	}

	req := &Request{
		URL:     SupervisorBaseURL + "/supervision/system-api/business-accept/workorderValue/page",
		Method:  "POST",
		Headers: headers,
		Body:    requestBody,
	}

	resp, err := s.HTTPClient.Send(req)
	if err != nil {
		return nil, fmt.Errorf("请求工单列表失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var workorderResp WorkOrderListResponse
	if err := json.Unmarshal(body, &workorderResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w, body: %s", err, string(body))
	}

	if workorderResp.RespCode != "0000" {
		return nil, fmt.Errorf("获取工单列表失败: %s", workorderResp.RespDesc)
	}

	return &workorderResp, nil
}

// GroupMember 用户信息（用于工作组成员）
type GroupMember struct {
	ID              int    `json:"id"`
	LoginName       string `json:"loginName"`
	NickName        string `json:"nickName"`
	UserType        string `json:"userType"`
	Email           string `json:"email"`
	Phone           string `json:"phone"`
	Sex             string `json:"sex"`
	ValidFlag       string `json:"validFlag"`
	DeptId          string `json:"deptId"`
	DeptName        string `json:"deptName"`
	LockFlag        string `json:"lockFlag"`
	DeputyAccountNumber string `json:"deputyAccountNumber,omitempty"`
}

// GroupMemberListResponse 工作组成员列表响应
type GroupMemberListResponse struct {
	Result struct {
		Total   int           `json:"total"`
		Size    int           `json:"size"`
		Current int           `json:"current"`
		Records []GroupMember `json:"records"`
	} `json:"result"`
	Total    int    `json:"total"`
	RespCode string `json:"respCode"`
	RespDesc string `json:"respDesc"`
}

// CheckinUserInfo 签到用户信息
type CheckinUserInfo struct {
	UserID       int    `json:"userId"`
	LoginName   string `json:"loginName"`
	NickName    string `json:"nickName"`
	CheckInTime string `json:"checkInTime"`
	GroupInfo   string `json:"groupInfo"`
}

// CheckinResponse 签到列表响应
type CheckinResponse struct {
	Result []CheckinUserInfo `json:"result"`
	Total   int    `json:"total"`
	RespCode string `json:"respCode"`
	RespDesc string `json:"respDesc"`
}

// GetCheckinList 获取签到用户列表
// 注意：调用前必须先执行Authenticate()认证
func (s *SupervisorService) GetCheckinList() (*CheckinResponse, error) {
	headers, err := s.GetAuthHeaders()
	if err != nil {
		return nil, err
	}

	req := &Request{
		URL:     SupervisorBaseURL + "/supervision/system-api/business-accept/user/getCheckInfoList",
		Method:  "POST",
		Headers: headers,
		Body:    map[string]interface{}{"groupCode": ""},
	}

	resp, err := s.HTTPClient.Send(req)
	if err != nil {
		return nil, fmt.Errorf("请求签到列表失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var checkinResp CheckinResponse
	if err := json.Unmarshal(body, &checkinResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w, body: %s", err, string(body))
	}

	if checkinResp.RespCode != "0000" {
		return nil, fmt.Errorf("获取签到列表失败: %s", checkinResp.RespDesc)
	}

	return &checkinResp, nil
}
