package api

import (
	"fmt"
)

// BugService handles defect/bug operations
type BugService struct {
	BaseService
}

// BugCreateRequest represents the request to create a bug/defect
type BugCreateRequest struct {
	HandlerId           string   `json:"handlerId"`
	DefectLevel         string   `json:"defectLevel"`
	Priority            string   `json:"priority"`
	DefectFrom          int      `json:"defectFrom"`
	ProjectId           string   `json:"projectId"`
	VersionId           string   `json:"versionId"`
	FixedVersion        string   `json:"fixedVersion"`
	ReleaseVersion      string   `json:"releaseVersion"`
	ActualPublishTime   string   `json:"actualPublishTime"`
	IterationId         string   `json:"iterationId"`
	DemandId            []string `json:"demandId"`
	PlanStartTime       string   `json:"planStartTime"`
	PlanEndTime         string   `json:"planEndTime"`
	AttributionId       string   `json:"attributionId"`
	DefectAmbient       string   `json:"defectAmbient"`
	DefectType          string   `json:"defectType"`
	ProductOrPlatformId string   `json:"productOrPlatformId"`
	SystemId            string   `json:"systemId"`
	PlanId              []string `json:"planId"`
	CaseId              []string `json:"caseId"`
	LabelList           []string `json:"labelList"`
	ColumnJson          string   `json:"columnJson"`
	DefectName          string   `json:"defectName"`
	Remark              string   `json:"remark"`
	FileIds             []string `json:"fileIds"`
	RelCaseId           string   `json:"relCaseId"`
	RelPlanId           string   `json:"relPlanId"`
	TemplateId          string   `json:"templateId"`
	OriginalDefectId    string   `json:"originalDefectId"`
}

// BugCreateResponse represents the response from creating a bug
type BugCreateResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data bool   `json:"data"`
}

// NewBugService creates a new BugService
func NewBugService(baseURL string, headers map[string]string, client HTTPClient) *BugService {
	return &BugService{
		BaseService: NewBaseService(baseURL, headers, client),
	}
}

// Create creates a new bug/defect
func (b *BugService) Create(req *BugCreateRequest) (*BugCreateResponse, error) {
	path := "/defect/defectInfo/insertDefectInfo"

	resp, err := b.Post(path, req)
	if err != nil {
		return nil, err
	}

	var result BugCreateResponse
	if err := b.ParseJSON(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetBugList gets the list of bugs for a project (deprecated, use ListBugs instead)
func (b *BugService) GetBugList(projectId string, pageNo, pageSize int) (*BugListResponse, error) {
	path := fmt.Sprintf("/defect/defectInfo/getDefectInfoPage?pageNo=%d&pageSize=%d&projectId=%s", pageNo, pageSize, projectId)

	resp, err := b.Get(path)
	if err != nil {
		return nil, err
	}

	var result BugListResponse
	if err := b.ParseJSON(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// BugListRequest represents the request for listing bugs
type BugListRequest struct {
	Page              int                    `json:"page"`
	Limit             int                    `json:"limit"`
	ColumnJsonQuery   map[string]interface{} `json:"columnJsonQuery,omitempty"`
	StatusIds         []string               `json:"statusIds,omitempty"`
	HandlerIds        []string               `json:"handlerIds,omitempty"`
	SystemIds         []string               `json:"systemIds,omitempty"`
	CreateDate        *DateRange             `json:"createDate,omitempty"`
	StatusTime        *DateRange             `json:"statusTime,omitempty"`
	ActualPublishTime *DateRange             `json:"actualPublishTime,omitempty"`
	PlanTime          *DateRange             `json:"planTime,omitempty"`
}

// DateRange represents a date range filter
type DateRange struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

// ListBugs lists bugs with filters (uses pageList API)
func (b *BugService) ListBugs(req *BugListRequest) (*BugPageListResponse, error) {
	path := "/defect/defectInfo/pageList"

	resp, err := b.Post(path, req)
	if err != nil {
		return nil, err
	}

	var result BugPageListResponse
	if err := b.ParseJSON(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// BugPageListResponse represents the response for pageList API
type BugPageListResponse struct {
	Code          int       `json:"code"`
	Msg           string    `json:"msg"`
	TotalRecCount int       `json:"totalRecCount"`
	Data          []BugItem `json:"data"`
}

// BugItem represents a single bug in the list
type BugItem struct {
	ID               string `json:"id"`
	DefectCode       string `json:"defectCode"`
	DefectName       string `json:"defectName"`
	DefectStatus     string `json:"defectStatus"`
	Priority         int    `json:"priority"`
	DefectLevel      int    `json:"defectLevel"`
	HandlerId        string `json:"handlerId"`
	HandlerName      string `json:"handlerName"`
	ProjectId        string `json:"projectId"`
	ProjectName      string `json:"projectName"`
	Creator          string `json:"creator"`
	CreatorName      string `json:"creatorName"`
	CreateDate       string `json:"createDate"`
	DefectStatusName struct {
		Name string `json:"name"`
	} `json:"defectStatusName"`
}

// BugListResponse represents the response for listing bugs
type BugListResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Records []struct {
			DefectId    string `json:"defectId"`
			DefectName  string `json:"defectName"`
			DefectLevel string `json:"defectLevel"`
			Priority    string `json:"priority"`
			Status      string `json:"status"`
			HandlerName string `json:"handlerName"`
			CreateTime  string `json:"createTime"`
		} `json:"records"`
		Total int `json:"total"`
	} `json:"data"`
}

// BugDetailResponse represents the response for getting bug detail
type BugDetailResponse struct {
	Code int       `json:"code"`
	Msg  string    `json:"msg"`
	Data BugDetail `json:"data"`
}

// BugDetail represents detailed bug information
type BugDetail struct {
	ID                    string `json:"id"`
	DefectCode            string `json:"defectCode"`
	DefectName            string `json:"defectName"`
	Remark                string `json:"remark"`
	ProjectId             string `json:"projectId"`
	ProjectName           string `json:"projectName"`
	Priority              int    `json:"priority"`
	DefectLevel           int    `json:"defectLevel"`
	DefectLevelDes        string `json:"defectLevelDes"`
	PlanStartTime         string `json:"planStartTime"`
	PlanEndTime           string `json:"planEndTime"`
	HandlerId             string `json:"handlerId"`
	HandlerName           string `json:"handlerName"`
	DefectStatus          string `json:"defectStatus"`
	RealStartTime         string `json:"realStartTime"`
	RealEndTime           string `json:"realEndTime"`
	FixedVersion          string `json:"fixedVersion"`
	ReleaseVersion        string `json:"releaseVersion"`
	ActualPublishTime     string `json:"actualPublishTime"`
	DiscoveryPhase        string `json:"discoveryPhase"`
	DiscoveryPhaseDesc    string `json:"discoveryPhaseDesc"`
	DefectFrom            int    `json:"defectFrom"`
	DefectFromDesc        string `json:"defectFromDesc"`
	DefectAmbient         string `json:"defectAmbient"`
	DefectAmbientDes      string `json:"defectAmbientDes"`
	DefectType            int    `json:"defectType"`
	DefectTypeDes         string `json:"defectTypeDes"`
	ProductOrPlatformId   string `json:"productOrPlatformId"`
	ProductOrPlatformName string `json:"productOrPlatformName"`
	SystemId              string `json:"systemId"`
	SystemName            string `json:"systemName"`
	IterationId           string `json:"iterationId"`
	IterationName         string `json:"iterationName"`
	WorkspaceId           string `json:"workspaceId"`
	CloseTime             string `json:"closeTime"`
	Creator               string `json:"creator"`
	CreatorName           string `json:"creatorName"`
	CreateDate            string `json:"createDate"`
	Updater               string `json:"updater"`
	UpdaterName           string `json:"updaterName"`
	UpdateDate            string `json:"updateDate"`
	TemplateId            string `json:"templateId"`
	DefectStatusName      struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Desc    string `json:"desc"`
		Color   string `json:"color"`
		StageId string `json:"stageId"`
	} `json:"defectStatusName"`
}

// GetBugDetailRequest represents the request for getting bug detail
type GetBugDetailRequest struct {
	ID string `json:"id"`
}

// GetBugDetail gets detailed information of a bug by ID
func (b *BugService) GetBugDetail(id string) (*BugDetailResponse, error) {
	path := "/defect/defectInfo/getDefectDetailById"
	req := &GetBugDetailRequest{ID: id}

	resp, err := b.Post(path, req)
	if err != nil {
		return nil, err
	}

	var result BugDetailResponse
	if err := b.ParseJSON(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// BugStatusRequest represents the request for getting bug statuses
type BugStatusRequest struct {
	SceneId int `json:"sceneId"`
}

// BugStatusResponse represents the response for getting bug statuses
type BugStatusResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data []BugStatus `json:"data"`
}

// BugStatus represents a single bug status
type BugStatus struct {
	StatusId          string `json:"statusId"`
	StatusName        string `json:"statusName"`
	IsStarStatus      string `json:"isStarStatus"`
	OptionalStatusIds string `json:"optionalStatusIds"`
	StageId           string `json:"stageId"`
	StageDesc         string `json:"stageDesc"`
	StageColor        string `json:"stageColor"`
	DelFlag           string `json:"delFlag"`
}

// GetBugStatuses gets the list of available bug statuses
// TODO: sceneId meaning needs to be confirmed, currently using 6 as default
func (b *BugService) GetBugStatuses(sceneId int) (*BugStatusResponse, error) {
	path := "/enterprise/sceneStatus/optionList"
	req := &BugStatusRequest{SceneId: sceneId}

	resp, err := b.Post(path, req)
	if err != nil {
		return nil, err
	}

	var result BugStatusResponse
	if err := b.ParseJSON(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// BugUpdateStatusRequest represents the request for updating bug status
type BugUpdateStatusRequest struct {
	ID           string `json:"id"`
	DefectStatus string `json:"defectStatus"`
}

// BugUpdateStatusResponse represents the response for updating bug status
type BugUpdateStatusResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data string `json:"data"`
}

// UpdateBugStatus updates the status of a bug
func (b *BugService) UpdateBugStatus(id, defectStatus string) (*BugUpdateStatusResponse, error) {
	path := "/defect/defectInfo/updateStatusDefectId"
	req := &BugUpdateStatusRequest{
		ID:           id,
		DefectStatus: defectStatus,
	}

	resp, err := b.Post(path, req)
	if err != nil {
		return nil, err
	}

	var result BugUpdateStatusResponse
	if err := b.ParseJSON(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// BugDeleteRequest represents the request for deleting bugs
// The API expects an array of bug IDs
// Example: ["2030502335635046402"]
type BugDeleteRequest []string

// BugDeleteResponse represents the response from deleting bugs
type BugDeleteResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data bool   `json:"data"`
}

// DeleteBugs deletes bugs by their IDs
// The API expects a POST request with an array of bug IDs
// Returns true in data field if deletion is successful
func (b *BugService) DeleteBugs(bugIDs []string) (*BugDeleteResponse, error) {
	path := "/defect/defectInfo/removeDefectInfo"

	resp, err := b.Post(path, BugDeleteRequest(bugIDs))
	if err != nil {
		return nil, err
	}

	var result BugDeleteResponse
	if err := b.ParseJSON(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
