package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/user/lc/internal/api"
	"github.com/user/lc/internal/common"
)

var libWorkspaceKey string

var libCmd = &cobra.Command{
	Use:   "lib",
	Short: "管理文档库",
	Long:  `管理文档库，包括查询列表、创建文档库、删除文档库、创建文件夹等功能。`,
}

var libListCmd = &cobra.Command{
	Use:   "list",
	Short: "查询文档库列表",
	Long: `查询当前研发空间下的文档库列表。

示例:
  lc lib list --workspace-key XXJSxiaobaice`,
	Run: func(cmd *cobra.Command, args []string) {
		common.HandleAutoDetectWithExit(func() error {
			return tryAutoDetectForLibList(cmd)
		}, "-w, --workspace-key")
		common.Execute(listLibraries, common.ExecuteOptions{
			DebugMode:   debugMode,
			Insecure:    insecureSkipVerify,
			Logger:      &logger,
			PrettyPrint: prettyMode,
		CommandName: "lib list",
		})
	},
}

var libCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "创建文档库",
	Long: `创建一个新的文档库。

示例:
  lc lib create "我的文档库" --workspace-key XXJSxiaobaice`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		common.HandleAutoDetectWithExit(func() error {
			return tryAutoDetectForLibCreate(cmd)
		}, "-w, --workspace-key")
		common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
			return createLibrary(ctx, args[0])
		}, common.ExecuteOptions{
			DebugMode:   debugMode,
			Insecure:    insecureSkipVerify,
			Logger:      &logger,
			PrettyPrint: prettyMode,
		CommandName: "lib create",
		})
	},
}

var libDeleteCmd = &cobra.Command{
	Use:   "delete [external-lib-id]",
	Short: "删除文档库",
	Long: `删除指定 externalLibId 的文档库。

externalLibId 可以从列表查询结果中获取 (externalLibId 字段)。

示例:
  lc lib delete 1709832

警告:
  删除文档库将永久删除其中的所有文件和文件夹，请谨慎操作。`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
			return deleteLibrary(ctx, args[0])
		}, common.ExecuteOptions{
			DebugMode:   debugMode,
			Insecure:    insecureSkipVerify,
			Logger:      &logger,
			PrettyPrint: prettyMode,
		CommandName: "lib delete",
		})
	},
}

var fileCmd = &cobra.Command{
	Use:   "file",
	Short: "文件管理",
	Long:  `管理文档库中的文件。`,
}

var fileDeleteCmd = &cobra.Command{
	Use:   "delete [obj-id]",
	Short: "删除文件或文件夹",
	Long: `删除文档库中的文件或文件夹。

参数:
  obj-id: 要删除的文件或文件夹 ID（必填）
  --folder-id: 文件所在的文件夹 ID（必填，用于权限验证）

示例:
  # 删除文件
  lc lib file delete 3648333 --folder-id <folder-id>

  # 删除文件夹
  lc lib file delete <folder-id> --folder-id <parent-folder-id>`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
			return deleteFileOrFolder(ctx, args[0])
		}, common.ExecuteOptions{
			DebugMode:   debugMode,
			Insecure:    insecureSkipVerify,
			Logger:      &logger,
			PrettyPrint: prettyMode,
		CommandName: "lib file delete",
		})
	},
}

var folderCmd = &cobra.Command{
	Use:   "folder",
	Short: "文件夹管理",
	Long:  `管理文档库中的文件夹。`,
}

var uploadCmd = &cobra.Command{
	Use:   "upload [file-path]",
	Short: "上传文件到文档库",
	Long: `上传文件到指定的文件夹。

参数:
  folder-id: 目标文件夹ID（必填）
  file-path: 要上传的本地文件路径（必填）

示例:
  # 上传文件到指定文件夹
  lc lib upload /path/to/file.pdf --folder-id <folder-id>

  # 上传并指定文件名
  lc lib upload /path/to/file.pdf --folder-id <folder-id> --name "新文件名.pdf"`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
			return uploadFile(ctx, args[0])
		}, common.ExecuteOptions{
			DebugMode:   debugMode,
			Insecure:    insecureSkipVerify,
			Logger:      &logger,
			PrettyPrint: prettyMode,
		CommandName: "lib upload",
		})
	},
}

var folderCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "在文档库中创建文件夹",
	Long: `在指定文档库下创建文件夹。

prtId (父文件夹ID):
  - 根目录使用文档库的 externalLibId
  - 子文件夹使用父文件夹的 ID

示例:
  # 在文档库根目录创建文件夹
  lc lib folder create "新建文件夹" --prt-id 1709832

  # 在子目录创建文件夹
  lc lib folder create "子文件夹" --prt-id 123456`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
			return createFolder(ctx, args[0])
		}, common.ExecuteOptions{
			DebugMode:   debugMode,
			Insecure:    insecureSkipVerify,
			Logger:      &logger,
			PrettyPrint: prettyMode,
		CommandName: "lib folder create",
		})
	},
}

var folderTreeCmd = &cobra.Command{
	Use:   "tree",
	Short: "查询文件夹列表（仅文件夹）",
	Long: `查询指定文件夹下的子文件夹列表。

注意: 此命令仅返回子文件夹，不返回文件。如需查看文件请使用 "lc lib folder list"。

参数:
  --prt-id: 父文件夹ID（必填）
    - 查询文档库根目录内容使用 externalLibId
    - 查询子文件夹内容使用文件夹的 folderId

示例:
  # 查询文档库根目录的子文件夹
  lc lib folder tree --prt-id 1709832

  # 查询子文件夹的子文件夹
  lc lib folder tree --prt-id 1712345`,
	Run: func(cmd *cobra.Command, args []string) {
		common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
			return treeList(ctx, folderPrtID)
		}, common.ExecuteOptions{
			DebugMode: debugMode,
			Insecure:  insecureSkipVerify,
			Logger:    &logger,
			PrettyPrint: prettyMode,
		CommandName: "lib folder tree",
		})
	},
}

var folderListCmd = &cobra.Command{
	Use:   "list",
	Short: "查询文件和文件夹列表（分页）",
	Long: `查询指定文件夹下的文件和文件夹列表，支持分页。

参数:
  --prt-id: 父文件夹ID（必填）
    - 查询文档库根目录使用 externalLibId
    - 查询子文件夹使用 folderId
  --page: 页码（可选，默认 1）
  --size: 每页数量（可选，默认 10）

示例:
  # 查询文档库根目录内容（第一页，10条）
  lc lib folder list --prt-id 1709832

  # 查询第2页，每页20条
  lc lib folder list --prt-id 1709832 --page 2 --size 20`,
	Run: func(cmd *cobra.Command, args []string) {
		common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
			return pageList(ctx, folderPrtID, listPageNo, listPageSize)
		}, common.ExecuteOptions{
			DebugMode: debugMode,
			Insecure:  insecureSkipVerify,
			Logger:    &logger,
			PrettyPrint: prettyMode,
		CommandName: "lib folder list",
		})
	},
}

var folderPrtID int64
var listPageNo int
var listPageSize int
var uploadFolderID int64
var uploadFileName string
var deleteObjID int64
var deleteFolderID int64

func init() {
	rootCmd.AddCommand(libCmd)
	libCmd.AddCommand(libListCmd)
	libCmd.AddCommand(libCreateCmd)
	libCmd.AddCommand(libDeleteCmd)
	libCmd.AddCommand(folderCmd)
	libCmd.AddCommand(uploadCmd)
	libCmd.AddCommand(fileCmd)
	folderCmd.AddCommand(folderCreateCmd)
	folderCmd.AddCommand(folderTreeCmd)
	folderCmd.AddCommand(folderListCmd)
	fileCmd.AddCommand(fileDeleteCmd)

	// List command flags
	libListCmd.Flags().StringVarP(&libWorkspaceKey, "workspace-key", "w", "", common.GetFlagDesc("workspace-key")+"（可选，支持自动探测）")

	// Create command flags
	libCreateCmd.Flags().StringVarP(&libWorkspaceKey, "workspace-key", "w", "", common.GetFlagDesc("workspace-key")+"（可选，支持自动探测）")

	// Delete command flags (no additional flags needed)

	// Folder create command flags
	folderCreateCmd.Flags().Int64Var(&folderPrtID, "prt-id", 0, common.GetFlagDesc("prt-id")+"（必填，根目录使用 externalLibId）")
	folderCreateCmd.MarkFlagRequired("prt-id")

	// Folder tree command flags
	folderTreeCmd.Flags().Int64Var(&folderPrtID, "prt-id", 0, common.GetFlagDesc("prt-id"))
	folderTreeCmd.MarkFlagRequired("prt-id")

	// Folder list command flags
	folderListCmd.Flags().Int64Var(&folderPrtID, "prt-id", 0, common.GetFlagDesc("prt-id"))
	folderListCmd.MarkFlagRequired("prt-id")
	folderListCmd.Flags().IntVar(&listPageNo, "page", 1, common.GetFlagDesc("page")+"（可选，默认 1）")
	folderListCmd.Flags().IntVar(&listPageSize, "size", 10, common.GetFlagDesc("size")+"（可选，默认 10）")

	// Upload command flags
	uploadCmd.Flags().Int64Var(&uploadFolderID, "folder-id", 0, common.GetFlagDesc("folder-id"))
	uploadCmd.Flags().StringVar(&uploadFileName, "name", "", common.GetFlagDesc("upload-name"))
	uploadCmd.MarkFlagRequired("folder-id")

	// Delete file/folder command flags
	fileDeleteCmd.Flags().Int64Var(&deleteFolderID, "folder-id", 0, common.GetFlagDesc("folder-id")+"（用于权限验证）")
	fileDeleteCmd.MarkFlagRequired("folder-id")
}

func listLibraries(ctx *common.CommandContext) (interface{}, error) {
	// Get config for lib service
	baseURL := ctx.Config.API.BaseDocURL
	if baseURL == "" {
		baseURL = "https://rdcloud.4c.hq.cmcc"
	}

	headers := ctx.GetHeaders(libWorkspaceKey)
	libService := api.NewDocService(baseURL, headers, ctx.Client, ctx.Config)

	resp, err := libService.ListLibraries(libWorkspaceKey)
	if err != nil {
		return nil, fmt.Errorf("查询文档库列表失败: %w", err)
	}

	var libraries []map[string]interface{}
	for _, lib := range resp.Data.List {
		libraries = append(libraries, map[string]interface{}{
			"libId":         lib.LibID,
			"externalLibId": lib.ExternalLibID,
			"libName":       lib.LibName,
			"libType":       lib.LibType,
			"createTime":    lib.CreateTime,
			"ownerIds":      lib.OwnerIds,
		})
	}

	return map[string]interface{}{
		"success":             true,
		"totalNumber":         resp.Data.TotalNumber,
		"deputyAccountNumber": resp.Data.DeputyAccountNumber,
		"libraries":           libraries,
	}, nil
}

// library 命令的自动探测字段配置
var libAutoDetectBase = []common.AutoDetectField{
	{FlagName: "workspace-key", TargetVar: &libWorkspaceKey, ContextKey: "WorkspaceKey"},
}

// tryAutoDetectForLibList 尝试为 lib list 命令自动探测参数
func tryAutoDetectForLibList(cmd *cobra.Command) error {
	_, err := common.ApplyAutoDetect(cmd, libAutoDetectBase)
	return err
}

// tryAutoDetectForLibCreate 尝试为 lib create 命令自动探测参数
func tryAutoDetectForLibCreate(cmd *cobra.Command) error {
	_, err := common.ApplyAutoDetect(cmd, libAutoDetectBase)
	return err
}

func createLibrary(ctx *common.CommandContext, name string) (interface{}, error) {
	baseURL := ctx.Config.API.BaseDocURL
	if baseURL == "" {
		baseURL = "https://rdcloud.4c.hq.cmcc"
	}

	headers := ctx.GetHeaders(libWorkspaceKey)
	libService := api.NewDocService(baseURL, headers, ctx.Client, ctx.Config)

	resp, err := libService.CreateLibrary(name, libWorkspaceKey)
	if err != nil {
		return nil, fmt.Errorf("创建文档库失败: %w", err)
	}

	return map[string]interface{}{
		"success": true,
		"libId":   resp.Data.LibID,
		"message": "文档库创建成功",
	}, nil
}

func deleteLibrary(ctx *common.CommandContext, externalLibIDStr string) (interface{}, error) {
	externalLibID, err := strconv.ParseInt(externalLibIDStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("externalLibId 必须是数字: %w", err)
	}

	baseURL := ctx.Config.API.BaseDocURL
	if baseURL == "" {
		baseURL = "https://rdcloud.4c.hq.cmcc"
	}

	headers := ctx.GetHeaders(libWorkspaceKey)
	libService := api.NewDocService(baseURL, headers, ctx.Client, ctx.Config)

	resp, err := libService.DeleteLibrary(externalLibID)
	if err != nil {
		return nil, fmt.Errorf("删除文档库失败: %w", err)
	}

	return map[string]interface{}{
		"success": true,
		"code":    resp.Code,
		"message": "文档库删除成功",
	}, nil
}

func createFolder(ctx *common.CommandContext, name string) (interface{}, error) {
	if folderPrtID == 0 {
		return nil, fmt.Errorf("必须指定 --prt-id 参数")
	}

	baseURL := ctx.Config.API.BaseDocURL
	if baseURL == "" {
		baseURL = "https://rdcloud.4c.hq.cmcc"
	}

	headers := ctx.GetHeaders(libWorkspaceKey)
	libService := api.NewDocService(baseURL, headers, ctx.Client, ctx.Config)

	resp, err := libService.CreateFolder(folderPrtID, name)
	if err != nil {
		return nil, fmt.Errorf("创建文件夹失败: %w", err)
	}

	return map[string]interface{}{
		"success":    true,
		"folderId":   resp.Data.ID,
		"prtId":      resp.Data.PrtID,
		"name":       resp.Data.Name,
		"createTime": resp.Data.CreateTime,
		"message":    "文件夹创建成功",
	}, nil
}

func treeList(ctx *common.CommandContext, prtID int64) (interface{}, error) {
	if prtID == 0 {
		return nil, fmt.Errorf("必须指定 --prt-id 参数")
	}

	baseURL := ctx.Config.API.BaseDocURL
	if baseURL == "" {
		baseURL = "https://rdcloud.4c.hq.cmcc"
	}

	headers := ctx.GetHeaders(libWorkspaceKey)
	libService := api.NewDocService(baseURL, headers, ctx.Client, ctx.Config)

	resp, err := libService.TreeList(prtID)
	if err != nil {
		return nil, fmt.Errorf("查询文件夹列表失败: %w", err)
	}

	// Convert TreeItem list to simplified format
	items := make([]map[string]interface{}, 0, len(resp.List))
	for _, item := range resp.List {
		items = append(items, map[string]interface{}{
			"id":         item.ID,
			"name":       item.Text,
			"type":       item.ObjType, // 32=文件夹, 33=文件
			"hasChild":   item.HasChild,
			"permission": item.Permission,
		})
	}

	return map[string]interface{}{
		"success": true,
		"prtId":   prtID,
		"count":   len(items),
		"items":   items,
	}, nil
}

func pageList(ctx *common.CommandContext, prtID int64, pageNo, pageSize int) (interface{}, error) {
	if prtID == 0 {
		return nil, fmt.Errorf("必须指定 --prt-id 参数")
	}

	baseURL := ctx.Config.API.BaseDocURL
	if baseURL == "" {
		baseURL = "https://rdcloud.4c.hq.cmcc"
	}

	headers := ctx.GetHeaders(libWorkspaceKey)
	libService := api.NewDocService(baseURL, headers, ctx.Client, ctx.Config)

	resp, err := libService.PageList(prtID, pageNo, pageSize)
	if err != nil {
		return nil, fmt.Errorf("查询文件列表失败: %w", err)
	}

	// Convert PageListItem list to simplified format
	items := make([]map[string]interface{}, 0, len(resp.List))
	for _, item := range resp.List {
		typeStr := "文件夹"
		if item.ObjType == 33 {
			typeStr = "文件"
		}
		items = append(items, map[string]interface{}{
			"id":          item.ID,
			"name":        item.Name,
			"type":        item.ObjType, // 32=文件夹, 33=文件
			"typeStr":     typeStr,
			"size":        item.SizeStr,
			"owner":       item.OwnerName,
			"updatedBy":   item.UpdatedByName,
			"updatedDt":   item.UpdatedDt,
			"canView":     item.CanView,
			"canDownload": item.CanDownload,
			"canEdit":     item.CanEdit,
		})
	}

	return map[string]interface{}{
		"success":     true,
		"prtId":       prtID,
		"prtName":     resp.Data.PrtName,
		"pageNo":      resp.Data.PageNo,
		"pageSize":    resp.Data.PageSize,
		"totalNumber": resp.Data.TotalNumber,
		"totalPage":   resp.Data.TotalPage,
		"count":       len(items),
		"items":       items,
	}, nil
}

func deleteFileOrFolder(ctx *common.CommandContext, objIDStr string) (interface{}, error) {
	objID, err := strconv.ParseInt(objIDStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("obj-id 必须是数字: %w", err)
	}

	if deleteFolderID == 0 {
		return nil, fmt.Errorf("必须指定 --folder-id 参数")
	}

	baseURL := ctx.Config.API.BaseDocURL
	if baseURL == "" {
		baseURL = "https://rdcloud.4c.hq.cmcc"
	}

	headers := ctx.GetHeaders(libWorkspaceKey)
	libService := api.NewDocService(baseURL, headers, ctx.Client, ctx.Config)

	resp, err := libService.DeleteFileOrFolder(objID, deleteFolderID)
	if err != nil {
		return nil, fmt.Errorf("删除失败: %w", err)
	}

	var deletedItems []map[string]interface{}
	for _, item := range resp.Data.SuccessList {
		deletedItems = append(deletedItems, map[string]interface{}{
			"id":       item.ID,
			"name":     item.Name,
			"type":     item.ObjType,
			"path":     item.PrtPath,
			"size":     item.SizeStr,
		})
	}

	return map[string]interface{}{
		"success":      true,
		"deletedCount": len(resp.Data.SuccessList),
		"failedCount":  len(resp.Data.FailureList),
		"items":        deletedItems,
		"message":      "删除成功",
	}, nil
}

func uploadFile(ctx *common.CommandContext, filePath string) (interface{}, error) {
	if uploadFolderID == 0 {
		return nil, fmt.Errorf("必须指定 --folder-id 参数")
	}

	// Get file name from path or use provided name
	name := uploadFileName
	if name == "" {
		// Extract filename from path
		for i := len(filePath) - 1; i >= 0; i-- {
			if filePath[i] == '/' || filePath[i] == '\\' {
				name = filePath[i+1:]
				break
			}
		}
		if name == "" {
			name = filePath
		}
	}

	baseURL := ctx.Config.API.BaseDocURL
	if baseURL == "" {
		baseURL = "https://rdcloud.4c.hq.cmcc"
	}

	headers := ctx.GetHeaders(libWorkspaceKey)
	libService := api.NewDocService(baseURL, headers, ctx.Client, ctx.Config)

	// Step 1: Pre-upload to get upload key
	preUploadResp, err := libService.PreUpload(uploadFolderID)
	if err != nil {
		return nil, fmt.Errorf("获取上传凭证失败: %w", err)
	}

	// Step 2: Upload file using upload URL
	uploadResp, err := libService.UploadFile(preUploadResp.Data.UploadURL, filePath, name)
	if err != nil {
		return nil, fmt.Errorf("上传文件失败: %w", err)
	}

	if len(uploadResp.List) == 0 {
		return nil, fmt.Errorf("上传返回空列表")
	}

	uploadedFile := uploadResp.List[0]

	return map[string]interface{}{
		"success":   true,
		"docId":     uploadedFile.DocID,
		"revId":     uploadedFile.RevID,
		"fileName":  uploadedFile.FileName,
		"size":      uploadedFile.SizeStr,
		"status":    uploadedFile.Status,
		"folderId":  uploadedFile.PrtID,
		"message":   "文件上传成功",
	}, nil
}
