package api

import (
	"fmt"
)

// CIBuild represents a CI build task from the API
type CIBuild struct {
	ID                      string                 `json:"id"`
	TaskName                string                 `json:"taskName"`
	BuildStatus             string                 `json:"buildStatus"`
	BuildNumber             string                 `json:"buildNumber"`
	BuildType               string                 `json:"buildType"`
	BuildCodeType           string                 `json:"buildCodeType"`
	FileID                  string                 `json:"fileId"`
	LastBuildID             string                 `json:"lastBuildId"`
	LastBuildTime           string                 `json:"lastBuildTime"`
	HasDisable              bool                   `json:"hasDisable"`
	SpaceID                 string                 `json:"spaceId"`
	TenantID                string                 `json:"tenantId"`
	Deleted                 bool                   `json:"deleted"`
	CreateTime              string                 `json:"createTime"`
	CreateUid               string                 `json:"createUid"`
	UpdateTime              string                 `json:"updateTime"`
	UpdateUid               string                 `json:"updateUid"`
	DeleteTime              *string                `json:"deleteTime"`
	DeleteUid               *string                `json:"deleteUid"`
	BuildScript             interface{}            `json:"buildScript"`
	CodeConfig              interface{}            `json:"codeConfig"`
	StepConfigs             interface{}            `json:"stepConfigs"`
	Parameters              interface{}            `json:"parameters"`
	Caches                  interface{}            `json:"caches"`
	NotificationSubscriptions interface{}          `json:"notificationSubscriptions"`
	AccurateNotification    interface{}            `json:"accurateNotification"`
	AdvancedConfig          interface{}            `json:"advancedConfig"`
	BuildGroup              interface{}            `json:"buildGroup"`
	PermissionType          string                 `json:"permissionType"`
	SystemCode              *string                `json:"systemCode"`
	AssignedNode            int                    `json:"assignedNode"`
	Overtime                int                    `json:"overtime"`
	CacheMode               string                 `json:"cacheMode"`
	CacheGroupID            *string                `json:"cacheGroupId"`
	BuildCodeSnapshot       *BuildCodeSnapshot     `json:"buildCodeSnapshot"`
	BuildTriggerConfig      interface{}            `json:"buildTriggerConfig"`
	BuildSource             string                 `json:"buildSource"`
	ConcurrentStrategy      int                    `json:"concurrentStrategy"`
	MaxConcurrentCount      int                    `json:"maxConcurrentCount"`
	Projects                interface{}            `json:"projects"`
	TaskSign                interface{}            `json:"taskSign"`
	ProjectCodes            interface{}            `json:"projectCodes"`
	LatestVersion           interface{}            `json:"latestVersion"`
	PipelineName            interface{}            `json:"pipelineName"`
	PipelineID              interface{}            `json:"pipelineId"`
	CreateUserName          string                 `json:"createUserName"`
	IsFavorited             bool                   `json:"isFavorited"`
	ResourceAllocation      string                 `json:"resourceAllocation"`
}

// BuildCodeSnapshot represents the code snapshot information for a build
type BuildCodeSnapshot struct {
	ID              int                    `json:"id"`
	TaskID          string                 `json:"taskId"`
	BuildSnapshotID string                 `json:"buildSnapshotId"`
	VcsID           string                 `json:"vcsId"`
	VcsName         string                 `json:"vcsName"`
	VcsRepository   string                 `json:"vcsRepository"`
	VcsBranch       string                 `json:"vcsBranch"`
	VcsCloneType    string                 `json:"vcsCloneType"`
	VcsSubmodule    interface{}            `json:"vcsSubmodule"`
	CommitID        string                 `json:"commitId"`
	CommitMessage   string                 `json:"commitMessage"`
	ChangeMessage   interface{}            `json:"changeMessage"`
	CommitMsg       *CommitMessage         `json:"commitMsg"`
	SpaceID         interface{}            `json:"spaceId"`
	TenantID        string                 `json:"tenantId"`
	Deleted         bool                   `json:"deleted"`
	CreateTime      string                 `json:"createTime"`
	CreateUid       interface{}            `json:"createUid"`
	UpdateTime      string                 `json:"updateTime"`
	UpdateUid       interface{}            `json:"updateUid"`
	DeleteTime      interface{}            `json:"deleteTime"`
	DeleteUid       interface{}            `json:"deleteUid"`
}

// CommitMessage represents the parsed commit message
type CommitMessage struct {
	ProjectName  string `json:"projectName"`
	ProjectPath  string `json:"projectPath"`
	ID           string `json:"id"`
	ShortID      string `json:"shortId"`
	CreatedDate  string `json:"createdDate"`
	AuthorName   string `json:"authorName"`
	Message      string `json:"message"`
	AuthoredDate string `json:"authoredDate"`
	Title        string `json:"title"`
}

// CIBuildListResponse represents the API response for CI build list
type CIBuildListResponse struct {
	Success   bool      `json:"success"`
	Code      string    `json:"code"`
	Message   string    `json:"message"`
	Data      []CIBuild `json:"data"`
	PageNo    int       `json:"pageNo"`
	PageSize  int       `json:"pageSize"`
	Count     int       `json:"count"`
	PageCount int       `json:"pageCount"`
	StartRow  int       `json:"startRow"`
	EndRow    int       `json:"endRow"`
}

// CIService provides methods for CI build management
type CIService struct {
	BaseService
}

// NewCIService creates a new CIService
func NewCIService(baseURL string, headers map[string]string, client HTTPClient) *CIService {
	return &CIService{
		BaseService: NewBaseService(baseURL, headers, client),
	}
}

// ListBuilds retrieves the list of CI builds
// buildStatus: 2=执行中, 3=成功, 4=失败, 5=已停止
func (s *CIService) ListBuilds(spaceID string, buildStatus string, pageNo, pageSize int) (*CIBuildListResponse, error) {
	path := fmt.Sprintf("/cmdevops-ci/server/api/v1/build/pageList?pageNo=%d&pageSize=%d&buildTaskDTO.spaceId=%s",
		pageNo, pageSize, spaceID)

	if buildStatus != "" {
		path += fmt.Sprintf("&buildTaskDTO.buildStatus=%s", buildStatus)
	}

	resp, err := s.Get(path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result CIBuildListResponse
	if err := s.ParseJSON(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetBuildStatusText returns the human-readable status text
func GetBuildStatusText(status string) string {
	switch status {
	case "2":
		return "执行中"
	case "3":
		return "成功"
	case "4":
		return "失败"
	case "5":
		return "已停止"
	default:
		return "未知"
	}
}

// CIBuildHistory represents a single build history record
type CIBuildHistory struct {
	ID                 string             `json:"id"`
	TaskID             string             `json:"taskId"`
	BuildStatus        string             `json:"buildStatus"`
	BuildNumber        int                `json:"buildNumber"`
	Duration           int                `json:"duration"`
	StartTime          string             `json:"startTime"`
	EndTime            string             `json:"endTime"`
	SpaceID            string             `json:"spaceId"`
	TenantID           string             `json:"tenantId"`
	Source             string             `json:"source"`
	Deleted            bool               `json:"deleted"`
	CreateTime         string             `json:"createTime"`
	CreateUid          string             `json:"createUid"`
	UpdateTime         string             `json:"updateTime"`
	UpdateUid          *string            `json:"updateUid"`
	DeleteTime         *string            `json:"deleteTime"`
	DeleteUid          *string            `json:"deleteUid"`
	BuildCodeSnapshot  *BuildCodeSnapshot `json:"buildCodeSnapshot"`
	BuildStepSnapshots interface{}        `json:"buildStepSnapshots"`
	SystemCode         *string            `json:"systemCode"`
	AssignedNode       int                `json:"assignedNode"`
	FirstFailTime      *string            `json:"firstFailTime"`
	FixTime            *int               `json:"fixTime"`
	CreateUserName     string             `json:"createUserName"`
}

// CIBuildHistoryResponse represents the API response for build history
type CIBuildHistoryResponse struct {
	Success   bool             `json:"success"`
	Code      string           `json:"code"`
	Message   string           `json:"message"`
	Data      []CIBuildHistory `json:"data"`
	PageNo    int              `json:"pageNo"`
	PageSize  int              `json:"pageSize"`
	Count     int              `json:"count"`
	PageCount int              `json:"pageCount"`
	StartRow  int              `json:"startRow"`
	EndRow    int              `json:"endRow"`
}

// GetBuildHistory retrieves the build history for a specific task
func (s *CIService) GetBuildHistory(spaceID, taskID string, pageNo, pageSize int) (*CIBuildHistoryResponse, error) {
	path := fmt.Sprintf("/cmdevops-ci/server/api/v1/build/buildHistory?buildSnapshotDTO.spaceId=%s&buildSnapshotDTO.taskId=%s&pageNo=%d&pageSize=%d",
		spaceID, taskID, pageNo, pageSize)

	resp, err := s.Get(path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result CIBuildHistoryResponse
	if err := s.ParseJSON(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// FormatDuration formats duration in milliseconds to human readable string
func FormatDuration(durationMs int) string {
	if durationMs < 1000 {
		return fmt.Sprintf("%dms", durationMs)
	}
	seconds := durationMs / 1000
	if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	}
	minutes := seconds / 60
	secs := seconds % 60
	if minutes < 60 {
		return fmt.Sprintf("%dm%ds", minutes, secs)
	}
	hours := minutes / 60
	mins := minutes % 60
	return fmt.Sprintf("%dh%dm", hours, mins)
}

// CIBuildStep represents a single build step in the build detail
type CIBuildStep struct {
	ID             int         `json:"id"`
	TaskID         string      `json:"taskId"`
	BuildSnapshotID string     `json:"buildSnapshotId"`
	Name           string      `json:"name"`
	Type           string      `json:"type"`
	Serial         int         `json:"serial"`
	Config         interface{} `json:"config"`
	HasOpen        bool        `json:"hasOpen"`
	StepState      string      `json:"stepState"`
	StepResult     string      `json:"stepResult"`
	StartTime      *string     `json:"startTime"`
	EndTime        *string     `json:"endTime"`
	Duration       *int        `json:"duration"`
	StepLog        interface{} `json:"stepLog"`
	StepLogPath    *string     `json:"stepLogPath"`
	StepStatus     interface{} `json:"stepStatus"`
	SpaceID        string      `json:"spaceId"`
	TenantID       string      `json:"tenantId"`
	Deleted        bool        `json:"deleted"`
	CreateTime     string      `json:"createTime"`
	CreateUid      string      `json:"createUid"`
	UpdateTime     string      `json:"updateTime"`
	UpdateUid      interface{} `json:"updateUid"`
	DeleteTime     interface{} `json:"deleteTime"`
	DeleteUid      interface{} `json:"deleteUid"`
	Stage          interface{} `json:"stage"`
	StageNum       interface{} `json:"stageNum"`
	Log            interface{} `json:"log"`
}

// CIBuildDetail represents the detailed build information
type CIBuildDetail struct {
	ID                string             `json:"id"`
	TaskID            string             `json:"taskId"`
	BuildStatus       string             `json:"buildStatus"`
	BuildNumber       int                `json:"buildNumber"`
	Duration          int                `json:"duration"`
	StartTime         string             `json:"startTime"`
	EndTime           string             `json:"endTime"`
	SpaceID           string             `json:"spaceId"`
	TenantID          string             `json:"tenantId"`
	Source            string             `json:"source"`
	Deleted           bool               `json:"deleted"`
	CreateTime        string             `json:"createTime"`
	CreateUid         string             `json:"createUid"`
	UpdateTime        string             `json:"updateTime"`
	UpdateUid         interface{}        `json:"updateUid"`
	DeleteTime        interface{}        `json:"deleteTime"`
	DeleteUid         interface{}        `json:"deleteUid"`
	BuildCodeSnapshot *BuildCodeSnapshot `json:"buildCodeSnapshot"`
	BuildStepSnapshots []CIBuildStep     `json:"buildStepSnapshots"`
	SystemCode        interface{}        `json:"systemCode"`
	AssignedNode      int                `json:"assignedNode"`
	FirstFailTime     *string            `json:"firstFailTime"`
	FixTime           interface{}        `json:"fixTime"`
	CreateUserName    interface{}        `json:"createUserName"`
}

// CIBuildDetailResponse represents the API response for build detail
type CIBuildDetailResponse struct {
	Success bool          `json:"success"`
	Code    string        `json:"code"`
	Message string        `json:"message"`
	Data    CIBuildDetail `json:"data"`
}

// GetBuildDetail retrieves the detailed information for a specific build snapshot
func (s *CIService) GetBuildDetail(buildSnapshotID string) (*CIBuildDetailResponse, error) {
	path := fmt.Sprintf("/cmdevops-ci/server/api/v1/build/getBuildSnapshotDetail/%s", buildSnapshotID)

	resp, err := s.Get(path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result CIBuildDetailResponse
	if err := s.ParseJSON(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetStepResultText returns the human-readable step result text
func GetStepResultText(result string) string {
	switch result {
	case "SUCCESS":
		return "成功"
	case "FAILURE":
		return "失败"
	case "NOT_BUILT":
		return "未执行"
	case "ABORTED":
		return "已中止"
	default:
		return result
	}
}

// GetStepStateText returns the human-readable step state text
func GetStepStateText(state string) string {
	switch state {
	case "FINISHED":
		return "已完成"
	case "RUNNING":
		return "执行中"
	case "SKIPPED":
		return "已跳过"
	case "PENDING":
		return "等待中"
	default:
		return state
	}
}

// CIBuildLogResponse represents the API response for build log
type CIBuildLogResponse struct {
	Success bool   `json:"success"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

// GetBuildLog retrieves the build log for a specific build snapshot
func (s *CIService) GetBuildLog(buildSnapshotID string) (*CIBuildLogResponse, error) {
	path := fmt.Sprintf("/cmdevops-ci/server/api/v1/build/queryLog/%s", buildSnapshotID)

	resp, err := s.Get(path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result CIBuildLogResponse
	if err := s.ParseJSON(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
