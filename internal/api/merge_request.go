package api

import (
	"fmt"
)

type MergeRequestService struct {
	BaseService
}

type MergeRequestCreateRequest struct {
	SourceBranch           string `json:"sourceBranch"`
	TargetBranch           string `json:"targetBranch"`
	SourceProjectId        *int   `json:"sourceProjectId"`
	Title                  string `json:"title"`
	Description            string `json:"description"`
	RemoveSourceBranch     bool   `json:"removeSourceBranch"`
	PrimaryReviewerNum     int    `json:"primaryReviewerNum"`
	PrimaryReviewerIds     []int  `json:"primaryReviewerIds"`
	GeneralReviewerNum     int    `json:"generalReviewerNum"`
	GeneralReviewerIds     []int  `json:"generalReviewerIds"`
	PrimaryReviewerUserIds []int  `json:"primaryReviewerUserIds"`
	GeneralReviewerUserIds []int  `json:"generalReviewerUserIds"`
	WorkItems              []int  `json:"workItems"`
	OriginalReviewerIds    []int  `json:"originalReviewerIds"`
	MergeType              string `json:"mergeType"`
	StateEvent             string `json:"stateEvent"`
	PrAutoMergeEnabled     bool   `json:"prAutoMergeEnabled"`
}

type MergeRequestCreateResponse struct {
	Success bool             `json:"success"`
	Code    string           `json:"code"`
	Message string           `json:"message"`
	Data    MergeRequestData `json:"data"`
}

type MergeRequestData struct {
	Id                       int    `json:"id"`
	Iid                      int    `json:"iid"`
	ProjectId                int    `json:"projectId"`
	ProjectName              string `json:"projectName"`
	ProjectPath              string `json:"projectPath"`
	FullPath                 string `json:"fullPath"`
	Title                    string `json:"title"`
	Description              string `json:"description"`
	State                    string `json:"state"`
	CreatedAt                string `json:"createdAt"`
	UpdatedAt                string `json:"updatedAt"`
	TargetBranch             string `json:"targetBranch"`
	SourceBranch             string `json:"sourceBranch"`
	MergeStatus              string `json:"mergeStatus"`
	Sha                      string `json:"sha"`
	WebUrl                   string `json:"webUrl"`
	ShouldRemoveSourceBranch bool   `json:"shouldRemoveSourceBranch"`
	Author                   Author `json:"author"`
}

type Author struct {
	Id       int    `json:"id"`
	Name     string `json:"name"`
	UserName string `json:"userName"`
}

func NewMergeRequestService(baseURL string, headers map[string]string, client HTTPClient) *MergeRequestService {
	return &MergeRequestService{
		BaseService: NewBaseService(baseURL, headers, client),
	}
}

func (m *MergeRequestService) Create(projectId int, req *MergeRequestCreateRequest) (*MergeRequestCreateResponse, error) {
	path := fmt.Sprintf("/projects/%d/merge-requests", projectId)

	resp, err := m.Post(path, req)
	if err != nil {
		return nil, err
	}

	var result MergeRequestCreateResponse
	if err := m.ParseJSON(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// MergeRequestReviewRequest represents the request for reviewing a merge request
type MergeRequestReviewRequest struct {
	ReviewType string `json:"reviewType"` // approve, reject
	Comment    string `json:"comment"`
}

// MergeRequestReviewResponse represents the response from reviewing a merge request
type MergeRequestReviewResponse struct {
	Success bool   `json:"success"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Id        int    `json:"id"`
		Iid       int    `json:"iid"`
		ProjectId int    `json:"projectId"`
		State     string `json:"state"`
	} `json:"data"`
}

// Review submits a review (approve/reject) for a merge request
func (m *MergeRequestService) Review(projectId int, mrId int, req *MergeRequestReviewRequest) (*MergeRequestReviewResponse, error) {
	path := fmt.Sprintf("/projects/%d/merge-requests/%d/review", projectId, mrId)

	resp, err := m.Post(path, req)
	if err != nil {
		return nil, err
	}

	var result MergeRequestReviewResponse
	if err := m.ParseJSON(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// MergeRequestMergeRequest represents the request for merging a merge request
type MergeRequestMergeRequest struct {
	ShouldRemoveSourceBranch bool   `json:"shouldRemoveSourceBranch"`
	MergeType                string `json:"mergeType"` // merge, squash, rebase
	Squash                   bool   `json:"squash"`
	MergeCommitMessage       string `json:"mergeCommitMessage"`
}

// MergeRequestMergeResponse represents the response from merging a merge request
type MergeRequestMergeResponse struct {
	Success bool   `json:"success"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Id          int    `json:"id"`
		Iid         int    `json:"iid"`
		ProjectId   int    `json:"projectId"`
		State       string `json:"state"`
		MergeStatus string `json:"mergeStatus"`
	} `json:"data"`
}

// Merge merges a merge request
func (m *MergeRequestService) Merge(projectId int, mrId int, req *MergeRequestMergeRequest) (*MergeRequestMergeResponse, error) {
	path := fmt.Sprintf("/projects/%d/merge-requests/%d/merge", projectId, mrId)

	resp, err := m.Put(path, req)
	if err != nil {
		return nil, err
	}

	var result MergeRequestMergeResponse
	if err := m.ParseJSON(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// MergeRequestCommentRequest represents the request for adding a comment to a merge request
type MergeRequestCommentRequest struct {
	Note         string `json:"note"`
	NoteableType string `json:"noteableType"` // PullRequestTheme
}

// MergeRequestCommentResponse represents the response from adding a comment
type MergeRequestCommentResponse struct {
	Success bool   `json:"success"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Id        int    `json:"id"`
		Note      string `json:"note"`
		CreatedAt string `json:"createdAt"`
		Author    struct {
			Id       int    `json:"id"`
			Name     string `json:"name"`
			UserName string `json:"userName"`
		} `json:"author"`
	} `json:"data"`
}

// Comment adds a comment to a merge request
func (m *MergeRequestService) Comment(projectId int, mrId int, req *MergeRequestCommentRequest) (*MergeRequestCommentResponse, error) {
	path := fmt.Sprintf("/projects/%d/merge-requests/%d/notes", projectId, mrId)

	resp, err := m.Post(path, req)
	if err != nil {
		return nil, err
	}

	var result MergeRequestCommentResponse
	if err := m.ParseJSON(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// MRComment represents a comment in a merge request
type MRComment struct {
	Id              int       `json:"id"`
	Note            string    `json:"note"`
	CreatedAt       string    `json:"createdAt"`
	UpdatedAt       string    `json:"updatedAt"`
	Author          MRAuthor  `json:"author"`
	Notes           []MRReply `json:"notes"`
	ResolvedState   string    `json:"resolvedState"`
	ResolvedEnabled bool      `json:"resolvedEnabled"`
}

type MRAuthor struct {
	Id       int    `json:"id"`
	Name     string `json:"name"`
	UserName string `json:"userName"`
	Email    string `json:"email"`
}

type MRReply struct {
	Id        int      `json:"id"`
	Note      string   `json:"note"`
	CreatedAt string   `json:"createdAt"`
	UpdatedAt string   `json:"updatedAt"`
	Author    MRAuthor `json:"author"`
}

// MergeRequestCommentsResponse represents the response for listing comments
type MergeRequestCommentsResponse struct {
	Success bool        `json:"success"`
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Data    []MRComment `json:"data"`
}

// GetComments retrieves all comments for a merge request
func (m *MergeRequestService) GetComments(projectId int, mrId int) (*MergeRequestCommentsResponse, error) {
	path := fmt.Sprintf("/projects/%d/merge-requests/%d/notes?prId=%d&noteType=1", projectId, mrId, mrId)

	resp, err := m.Get(path)
	if err != nil {
		return nil, err
	}

	var result MergeRequestCommentsResponse
	if err := m.ParseJSON(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// MergeRequestItem represents a merge request in the list
type MergeRequestItem struct {
	Iid             int      `json:"iid"`
	Title           string   `json:"title"`
	Description     string   `json:"description"`
	State           string   `json:"state"`
	CreatedAt       string   `json:"createdAt"`
	UpdatedAt       string   `json:"updatedAt"`
	TargetBranch    string   `json:"targetBranch"`
	SourceBranch    string   `json:"sourceBranch"`
	MergeStatus     string   `json:"mergeStatus"`
	UserNotesCount  string   `json:"userNotesCount"`
	Author          MRAuthor `json:"author"`
	SourceProjectId int      `json:"sourceProjectId"`
}

// MergeRequestListResponse represents the response for listing merge requests
type MergeRequestListResponse struct {
	Success   bool               `json:"success"`
	Code      string             `json:"code"`
	Message   string             `json:"message"`
	Data      []MergeRequestItem `json:"data"`
	PageNo    int                `json:"pageNo"`
	PageSize  int                `json:"pageSize"`
	Count     int                `json:"count"`
	PageCount int                `json:"pageCount"`
}

// MergeRequestListRequest represents the request for listing merge requests
type MergeRequestListRequest struct {
	PageNo   int
	PageSize int
	State    string // opened, merged, closed, all
}

// List retrieves merge requests for a project
func (m *MergeRequestService) List(projectId int, req *MergeRequestListRequest) (*MergeRequestListResponse, error) {
	state := req.State
	if state == "" {
		state = "all"
	}
	path := fmt.Sprintf("/projects/%d/merge-requests?pageNo=%d&pageSize=%d&state=%s&orderBy=updated_at&sort=desc",
		projectId, req.PageNo, req.PageSize, state)

	resp, err := m.Get(path)
	if err != nil {
		return nil, err
	}

	var result MergeRequestListResponse
	if err := m.ParseJSON(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// MergeRequestUpdateCommentStateRequest represents the request for updating comment state
type MergeRequestUpdateCommentStateRequest struct {
	ResolvedState string `json:"resolvedState"` // wontFix, closed, pending, fixed, active
}

// MergeRequestUpdateCommentStateResponse represents the response from updating comment state
type MergeRequestUpdateCommentStateResponse struct {
	Success bool        `json:"success"`
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// UpdateCommentState updates the state of a comment (note) in a merge request
// States: active (活动中), fixed (已解决), wontFix (无法修复), closed (已关闭), pending (正在挂起)
func (m *MergeRequestService) UpdateCommentState(projectId int, mrId int, noteId int, state string) (*MergeRequestUpdateCommentStateResponse, error) {
	path := fmt.Sprintf("/projects/%d/merge-requests/%d/notes/%d/state?resolvedState=%s", projectId, mrId, noteId, state)

	resp, err := m.Put(path, map[string]interface{}{})
	if err != nil {
		return nil, err
	}

	var result MergeRequestUpdateCommentStateResponse
	if err := m.ParseJSON(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
