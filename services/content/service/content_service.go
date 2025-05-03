package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"

	"backend/pkg/config"
	"backend/pkg/models"
	"backend/pkg/storage"
	"backend/services/content/repository"
	"image"
)

// ContentServiceImpl 内容服务实现
type ContentServiceImpl struct {
	repo        *repository.Repository
	redisClient *redis.Client
	logger      zerolog.Logger
}

// NewContentService 创建内容服务实例
func NewContentService(repo *repository.Repository, redisClient *redis.Client, logger zerolog.Logger) *ContentServiceImpl {
	return &ContentServiceImpl{
		repo:        repo,
		redisClient: redisClient,
		logger:      logger,
	}
}

// --------- 帖子相关处理程序 ---------

// ListContentHandler 处理列出内容请求
func (s *ContentServiceImpl) ListContentHandler(w http.ResponseWriter, r *http.Request) {
	// 获取分页参数
	offset, limit := getPaginationParams(r)

	// 获取查询参数
	categoryID := r.URL.Query().Get("category_id")
	tagID := r.URL.Query().Get("tag_id")

	ctx := r.Context()
	var posts []*models.Post
	var total int64
	var err error

	if categoryID != "" {
		// 获取特定分类的帖子
		posts, total, err = s.repo.GetPostsByCategory(ctx, categoryID, offset, limit)
	} else if tagID != "" {
		// 获取特定标签的帖子
		posts, total, err = s.repo.GetPostsByTag(ctx, tagID, offset, limit)
	} else {
		// 获取所有帖子
		posts, total, err = s.repo.ListPosts(ctx, offset, limit)
	}

	if err != nil {
		s.logger.Error().Err(err).Msg("获取帖子列表失败")
		http.Error(w, "获取帖子列表失败", http.StatusInternalServerError)
		return
	}

	// 构建响应
	response := map[string]interface{}{
		"posts": posts,
		"total": total,
		"pagination": map[string]int{
			"offset": offset,
			"limit":  limit,
		},
	}

	// 返回JSON响应
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		s.logger.Error().Err(err).Msg("编码响应失败")
		http.Error(w, "内部服务器错误", http.StatusInternalServerError)
		return
	}
}

// CreatePostHandler 处理创建内容请求
func (s *ContentServiceImpl) CreatePostHandler(w http.ResponseWriter, r *http.Request) {
	// 解析请求体
	var postInput struct {
		Type        models.PostType   `json:"type"`
		Content     string            `json:"content"`
		MediaFiles  models.MediaFiles `json:"media_files,omitempty"`
		Tags        []string          `json:"tags,omitempty"`
		Categories  []string          `json:"categories,omitempty"`
		PollOptions []string          `json:"poll_options,omitempty"`
		PollEndsAt  *time.Time        `json:"poll_ends_at,omitempty"`
		LinkURL     string            `json:"link_url,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&postInput); err != nil {
		s.logger.Error().Err(err).Msg("解析请求体失败")
		http.Error(w, "无效的请求数据", http.StatusBadRequest)
		return
	}

	// 从上下文中获取用户ID (假设中间件已经设置)
	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		s.logger.Error().Msg("未找到用户ID")
		http.Error(w, "未授权", http.StatusUnauthorized)
		return
	}

	// 创建帖子对象
	post := &models.Post{
		UserID:        userID,
		Type:          postInput.Type,
		Title:         "", // 在这个应用中可能不需要标题
		Content:       postInput.Content,
		Status:        models.PostStatusPublished,
		MediaFiles:    postInput.MediaFiles,
		LinkURL:       postInput.LinkURL,
		PublishedAt:   timePtr(time.Now()),
		PrivacyLevel:  "public", // 默认为公开
		AllowComments: true,     // 默认允许评论
	}

	// 处理投票选项（如果有）
	if postInput.Type == models.PostTypePoll && len(postInput.PollOptions) > 0 {
		pollOptions := models.PollOptions{}
		for _, optionText := range postInput.PollOptions {
			pollOptions = append(pollOptions, models.PollOption{
				ID:    generateUUID(),
				Text:  optionText,
				Count: 0,
			})
		}
		post.PollOptions = pollOptions
		post.PollEndsAt = postInput.PollEndsAt
	}

	// 创建帖子
	ctx := r.Context()
	if err := s.repo.CreatePost(ctx, post); err != nil {
		s.logger.Error().Err(err).Msg("创建帖子失败")
		http.Error(w, "创建帖子失败", http.StatusInternalServerError)
		return
	}

	// 处理标签
	if len(postInput.Tags) > 0 {
		for _, tagID := range postInput.Tags {
			if err := s.repo.AddTagToPost(ctx, post.ID, tagID); err != nil {
				s.logger.Warn().Err(err).Str("post_id", post.ID).Str("tag_id", tagID).Msg("添加标签失败")
			}
		}
	}

	// 处理分类
	if len(postInput.Categories) > 0 {
		for _, categoryID := range postInput.Categories {
			if err := s.repo.AddCategoryToPost(ctx, post.ID, categoryID); err != nil {
				s.logger.Warn().Err(err).Str("post_id", post.ID).Str("category_id", categoryID).Msg("添加分类失败")
			}
		}
	}

	// 记录审计日志
	auditLog := &models.PostAuditLog{
		PostID:     post.ID,
		UserID:     userID,
		ActionType: "create",
		NewValue:   post.Content,
	}
	if err := s.repo.LogPostActivity(ctx, auditLog); err != nil {
		s.logger.Warn().Err(err).Msg("记录审计日志失败")
	}

	// 发布成功后，增加用户的帖子计数
	if err := s.repo.IncrementUserPostCount(ctx, userID); err != nil {
		s.logger.Warn().Err(err).Str("user_id", userID).Msg("增加用户帖子计数失败")
	}

	// 返回创建的帖子
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(post); err != nil {
		s.logger.Error().Err(err).Msg("编码响应失败")
		http.Error(w, "内部服务器错误", http.StatusInternalServerError)
		return
	}
}

// GetContentHandler 处理获取内容请求
func (s *ContentServiceImpl) GetContentHandler(w http.ResponseWriter, r *http.Request) {
	// 获取URL参数
	vars := mux.Vars(r)
	postID := vars["id"]
	if postID == "" {
		http.Error(w, "缺少帖子ID", http.StatusBadRequest)
		return
	}

	// 获取帖子
	ctx := r.Context()
	post, err := s.repo.GetPostByID(ctx, postID)
	if err != nil {
		s.logger.Error().Err(err).Str("post_id", postID).Msg("获取帖子失败")
		http.Error(w, "帖子不存在", http.StatusNotFound)
		return
	}

	// 增加查看次数
	go func() {
		// 在后台异步增加查看次数
		bgCtx := context.Background()
		if err := s.repo.IncrementViewCount(bgCtx, postID); err != nil {
			s.logger.Warn().Err(err).Str("post_id", postID).Msg("增加查看次数失败")
		}
	}()

	// 获取帖子统计数据
	stats, err := s.repo.GetPostStats(ctx, postID)
	if err != nil {
		s.logger.Warn().Err(err).Str("post_id", postID).Msg("获取帖子统计数据失败")
	}

	// 获取标签和分类
	tags, err := s.repo.GetPostTags(ctx, postID)
	if err != nil {
		s.logger.Warn().Err(err).Str("post_id", postID).Msg("获取帖子标签失败")
	}

	categories, err := s.repo.GetPostCategories(ctx, postID)
	if err != nil {
		s.logger.Warn().Err(err).Str("post_id", postID).Msg("获取帖子分类失败")
	}

	// 构造响应
	response := map[string]interface{}{
		"post":       post,
		"stats":      stats,
		"tags":       tags,
		"categories": categories,
	}

	// 返回JSON响应
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		s.logger.Error().Err(err).Msg("编码响应失败")
		http.Error(w, "内部服务器错误", http.StatusInternalServerError)
		return
	}
}

// UpdateContentHandler 处理更新内容请求
func (s *ContentServiceImpl) UpdateContentHandler(w http.ResponseWriter, r *http.Request) {
	// 获取URL参数
	vars := mux.Vars(r)
	postID := vars["id"]
	if postID == "" {
		http.Error(w, "缺少帖子ID", http.StatusBadRequest)
		return
	}

	// 从上下文中获取用户ID (假设中间件已经设置)
	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		s.logger.Error().Msg("未找到用户ID")
		http.Error(w, "未授权", http.StatusUnauthorized)
		return
	}

	// 获取当前帖子
	ctx := r.Context()
	post, err := s.repo.GetPostByID(ctx, postID)
	if err != nil {
		s.logger.Error().Err(err).Str("post_id", postID).Msg("获取帖子失败")
		http.Error(w, "帖子不存在", http.StatusNotFound)
		return
	}

	// 检查权限
	if post.UserID != userID {
		// 权限校验可能还需要检查用户角色
		s.logger.Warn().Str("user_id", userID).Str("post_id", postID).Msg("未授权的更新请求")
		http.Error(w, "未授权", http.StatusForbidden)
		return
	}

	// 解析请求体
	var updateData struct {
		Title       *string            `json:"title,omitempty"`
		Content     *string            `json:"content,omitempty"`
		Status      *models.PostStatus `json:"status,omitempty"`
		MediaFiles  *models.MediaFiles `json:"media_files,omitempty"`
		Tags        []string           `json:"tags,omitempty"`
		Categories  []string           `json:"categories,omitempty"`
		PollOptions []string           `json:"poll_options,omitempty"`
		PollEndsAt  *time.Time         `json:"poll_ends_at,omitempty"`
		LinkURL     *string            `json:"link_url,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		s.logger.Error().Err(err).Msg("解析请求体失败")
		http.Error(w, "无效的请求数据", http.StatusBadRequest)
		return
	}

	// 保存旧值用于审计
	oldValue := post.Content

	// 更新帖子字段
	if updateData.Title != nil {
		post.Title = *updateData.Title
	}
	if updateData.Content != nil {
		post.Content = *updateData.Content
	}
	if updateData.Status != nil {
		post.Status = *updateData.Status
	}
	if updateData.MediaFiles != nil {
		post.MediaFiles = *updateData.MediaFiles
	}
	if updateData.LinkURL != nil {
		post.LinkURL = *updateData.LinkURL
	}

	// 更新帖子
	if err := s.repo.UpdatePost(ctx, post); err != nil {
		s.logger.Error().Err(err).Str("post_id", postID).Msg("更新帖子失败")
		http.Error(w, "更新帖子失败", http.StatusInternalServerError)
		return
	}

	// 处理投票选项（如果有）
	if post.Type == models.PostTypePoll && len(updateData.PollOptions) > 0 {
		pollOptions := models.PollOptions{}
		for _, optionText := range updateData.PollOptions {
			pollOptions = append(pollOptions, models.PollOption{
				ID:    generateUUID(),
				Text:  optionText,
				Count: 0,
			})
		}
		if err := s.repo.UpdatePostPoll(ctx, postID, pollOptions, *updateData.PollEndsAt); err != nil {
			s.logger.Warn().Err(err).Str("post_id", postID).Msg("更新投票选项失败")
		}
	}

	// 处理标签和分类更新
	if len(updateData.Tags) > 0 || len(updateData.Categories) > 0 {
		// 这里应该实现标签和分类的更新逻辑
		// 需要考虑移除旧标签/分类，添加新标签/分类
	}

	// 记录审计日志
	auditLog := &models.PostAuditLog{
		PostID:     post.ID,
		UserID:     userID,
		ActionType: "update",
		OldValue:   oldValue,
		NewValue:   post.Content,
	}
	if err := s.repo.LogPostActivity(ctx, auditLog); err != nil {
		s.logger.Warn().Err(err).Msg("记录审计日志失败")
	}

	// 返回更新后的帖子
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(post); err != nil {
		s.logger.Error().Err(err).Msg("编码响应失败")
		http.Error(w, "内部服务器错误", http.StatusInternalServerError)
		return
	}
}

// DeleteContentHandler 处理删除内容请求
func (s *ContentServiceImpl) DeleteContentHandler(w http.ResponseWriter, r *http.Request) {
	// 获取URL参数
	vars := mux.Vars(r)
	postID := vars["id"]
	if postID == "" {
		http.Error(w, "缺少帖子ID", http.StatusBadRequest)
		return
	}

	// 从上下文中获取用户ID (假设中间件已经设置)
	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		s.logger.Error().Msg("未找到用户ID")
		http.Error(w, "未授权", http.StatusUnauthorized)
		return
	}

	// 获取当前帖子
	ctx := r.Context()
	post, err := s.repo.GetPostByID(ctx, postID)
	if err != nil {
		s.logger.Error().Err(err).Str("post_id", postID).Msg("获取帖子失败")
		http.Error(w, "帖子不存在", http.StatusNotFound)
		return
	}

	// 检查权限
	if post.UserID != userID {
		// 权限校验可能还需要检查用户角色
		s.logger.Warn().Str("user_id", userID).Str("post_id", postID).Msg("未授权的删除请求")
		http.Error(w, "未授权", http.StatusForbidden)
		return
	}

	// 删除帖子
	if err := s.repo.DeletePost(ctx, postID); err != nil {
		s.logger.Error().Err(err).Str("post_id", postID).Msg("删除帖子失败")
		http.Error(w, "删除帖子失败", http.StatusInternalServerError)
		return
	}

	// 记录审计日志
	auditLog := &models.PostAuditLog{
		PostID:     post.ID,
		UserID:     userID,
		ActionType: "delete",
		OldValue:   post.Content,
	}
	if err := s.repo.LogPostActivity(ctx, auditLog); err != nil {
		s.logger.Warn().Err(err).Msg("记录审计日志失败")
	}

	// 返回成功响应
	w.WriteHeader(http.StatusNoContent)
}

// --------- 分类相关处理程序 ---------

// ListCategoriesHandler 处理列出分类请求
func (s *ContentServiceImpl) ListCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	// 获取分页参数
	offset, limit := getPaginationParams(r)

	// 获取分类列表
	ctx := r.Context()
	var categories []*models.Category
	var total int64
	var err error

	// 检查是否请求层次结构
	if r.URL.Query().Get("hierarchy") == "true" {
		categories, err = s.repo.GetCategoryHierarchy(ctx)
		if err != nil {
			s.logger.Error().Err(err).Msg("获取分类层次结构失败")
			http.Error(w, "获取分类层次结构失败", http.StatusInternalServerError)
			return
		}
		total = int64(len(categories))
	} else {
		categories, total, err = s.repo.ListCategories(ctx, offset, limit)
		if err != nil {
			s.logger.Error().Err(err).Msg("获取分类列表失败")
			http.Error(w, "获取分类列表失败", http.StatusInternalServerError)
			return
		}
	}

	// 构建响应
	response := map[string]interface{}{
		"categories": categories,
		"total":      total,
		"pagination": map[string]int{
			"offset": offset,
			"limit":  limit,
		},
	}

	// 返回JSON响应
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		s.logger.Error().Err(err).Msg("编码响应失败")
		http.Error(w, "内部服务器错误", http.StatusInternalServerError)
		return
	}
}

// --------- 标签相关处理程序 ---------

// ListTagsHandler 处理列出标签请求
func (s *ContentServiceImpl) ListTagsHandler(w http.ResponseWriter, r *http.Request) {
	// 获取分页参数
	offset, limit := getPaginationParams(r)

	// 获取标签列表
	ctx := r.Context()
	tags, total, err := s.repo.ListTags(ctx, offset, limit)
	if err != nil {
		s.logger.Error().Err(err).Msg("获取标签列表失败")
		http.Error(w, "获取标签列表失败", http.StatusInternalServerError)
		return
	}

	// 构建响应
	response := map[string]interface{}{
		"tags":  tags,
		"total": total,
		"pagination": map[string]int{
			"offset": offset,
			"limit":  limit,
		},
	}

	// 返回JSON响应
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		s.logger.Error().Err(err).Msg("编码响应失败")
		http.Error(w, "内部服务器错误", http.StatusInternalServerError)
		return
	}
}

// SearchContentHandler 处理搜索内容请求
func (s *ContentServiceImpl) SearchContentHandler(w http.ResponseWriter, r *http.Request) {
	// 获取查询参数
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "查询参数不能为空", http.StatusBadRequest)
		return
	}

	// 获取分页参数
	offset, limit := getPaginationParams(r)

	// 搜索帖子
	ctx := r.Context()
	posts, total, err := s.repo.SearchPosts(ctx, query, offset, limit)
	if err != nil {
		s.logger.Error().Err(err).Str("query", query).Msg("搜索帖子失败")
		http.Error(w, "搜索帖子失败", http.StatusInternalServerError)
		return
	}

	// 构建响应
	response := map[string]interface{}{
		"posts": posts,
		"total": total,
		"pagination": map[string]int{
			"offset": offset,
			"limit":  limit,
		},
		"query": query,
	}

	// 返回JSON响应
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		s.logger.Error().Err(err).Msg("编码响应失败")
		http.Error(w, "内部服务器错误", http.StatusInternalServerError)
		return
	}
}

// GetTrendingContentHandler 处理获取热门内容请求
func (s *ContentServiceImpl) GetTrendingContentHandler(w http.ResponseWriter, r *http.Request) {
	// 获取时间段参数
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "day" // 默认时间段
	}

	// 获取限制参数
	limitStr := r.URL.Query().Get("limit")
	limit := 10 // 默认限制
	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	// 获取热门帖子
	ctx := r.Context()
	posts, err := s.repo.GetTopPosts(ctx, period, limit)
	if err != nil {
		s.logger.Error().Err(err).Str("period", period).Msg("获取热门帖子失败")
		http.Error(w, "获取热门帖子失败", http.StatusInternalServerError)
		return
	}

	// 构建响应
	response := map[string]interface{}{
		"posts":  posts,
		"period": period,
		"limit":  limit,
	}

	// 返回JSON响应
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		s.logger.Error().Err(err).Msg("编码响应失败")
		http.Error(w, "内部服务器错误", http.StatusInternalServerError)
		return
	}
}

// UploadContentHandler 处理上传内容请求
func (s *ContentServiceImpl) UploadContentHandler(w http.ResponseWriter, r *http.Request) {
	// 设置响应头
	w.Header().Set("Content-Type", "application/json")

	// 从上下文中获取用户ID
	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		s.logger.Error().Msg("未找到用户ID")
		http.Error(w, "未授权", http.StatusUnauthorized)
		return
	}

	// 记录请求信息用于调试
	contentLength := r.ContentLength
	contentType := r.Header.Get("Content-Type")
	authorization := r.Header.Get("Authorization")
	s.logger.Info().
		Int64("content_length", contentLength).
		Str("content_type", contentType).
		Str("user_id", userID).
		Str("method", r.Method).
		Str("remote_addr", r.RemoteAddr).
		Str("auth_header_exists", fmt.Sprintf("%t", authorization != "")).
		Msg("收到文件上传请求")

	// 检查Content-Type
	if !strings.HasPrefix(contentType, "multipart/form-data") {
		s.logger.Error().Str("content_type", contentType).Msg("错误的Content-Type，应为multipart/form-data")
		http.Error(w, "Content-Type必须为multipart/form-data", http.StatusBadRequest)
		return
	}

	// 尝试使用标准方法解析表单，设置较大的内存限制以处理大文件
	maxMemory := int64(32 << 20) // 32MB的内存缓冲区
	if err := r.ParseMultipartForm(maxMemory); err != nil {
		s.logger.Error().Err(err).Msg("解析multipart表单失败")
		http.Error(w, "解析上传表单失败: "+err.Error(), http.StatusBadRequest)
		return
	}

	// 直接从已解析的表单中获取文件
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		s.logger.Error().Err(err).Msg("获取上传文件失败")
		http.Error(w, "获取上传文件失败: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	s.logger.Info().
		Str("filename", fileHeader.Filename).
		Int64("size", fileHeader.Size).
		Str("content_type", fileHeader.Header.Get("Content-Type")).
		Msg("文件获取成功")

	// 检查文件大小
	if fileHeader.Size > 15<<20 { // 15MB限制
		s.logger.Warn().Int64("size", fileHeader.Size).Msg("文件太大")
		http.Error(w, "文件大小超过限制", http.StatusBadRequest)
		return
	}

	// 创建临时文件的副本，防止直接使用multipart.File导致的问题
	tempFile, err := os.CreateTemp("", "upload-*")
	if err != nil {
		s.logger.Error().Err(err).Msg("创建临时文件失败")
		http.Error(w, "服务器无法处理上传", http.StatusInternalServerError)
		return
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// 将上传的文件内容复制到临时文件
	fileSize, err := io.Copy(tempFile, file)
	if err != nil {
		s.logger.Error().Err(err).Msg("复制上传文件内容失败")
		http.Error(w, "处理上传文件失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 确保文件大小一致
	if fileSize != fileHeader.Size {
		s.logger.Warn().
			Int64("expected_size", fileHeader.Size).
			Int64("actual_size", fileSize).
			Msg("文件大小不一致")
	}

	// 重置文件指针
	if _, err := tempFile.Seek(0, io.SeekStart); err != nil {
		s.logger.Error().Err(err).Msg("重置临时文件位置失败")
		http.Error(w, "处理上传失败", http.StatusInternalServerError)
		return
	}

	// 验证是否为图片或视频
	fileContentType := fileHeader.Header.Get("Content-Type")
	if fileContentType == "" || fileContentType == "application/octet-stream" {
		// 如果MIME类型未知，尝试自动检测
		buffer := make([]byte, 512)
		_, err = tempFile.Read(buffer)
		if err != nil && err != io.EOF {
			s.logger.Error().Err(err).Msg("读取文件头失败")
			http.Error(w, "无法检测文件类型", http.StatusBadRequest)
			return
		}
		fileContentType = http.DetectContentType(buffer)

		// 重置文件指针
		if _, err := tempFile.Seek(0, io.SeekStart); err != nil {
			s.logger.Error().Err(err).Msg("重置临时文件位置失败")
			http.Error(w, "处理上传失败", http.StatusInternalServerError)
			return
		}
	}

	// 验证文件类型
	isValidType := strings.HasPrefix(fileContentType, "image/") || strings.HasPrefix(fileContentType, "video/")
	if !isValidType {
		s.logger.Warn().Str("content_type", fileContentType).Msg("不支持的文件类型")
		http.Error(w, "只支持图片或视频文件", http.StatusBadRequest)
		return
	}

	// 创建MinIO客户端
	cfg := config.GetConfig()
	minioClient, err := storage.NewMinioClient(storage.MinioConfig{
		Endpoint:       cfg.MinioEndpoint,
		AccessKey:      cfg.MinioAccessKey,
		SecretKey:      cfg.MinioSecretKey,
		BucketName:     "content-media",
		UseSSL:         cfg.MinioUseSSL,
		PublicEndpoint: cfg.MinioPublicEndpoint,
	})
	if err != nil {
		s.logger.Error().Err(err).Msg("创建MinIO客户端失败")
		http.Error(w, "存储服务连接失败", http.StatusInternalServerError)
		return
	}

	// 确定文件类型
	fileType := "image"
	if strings.HasPrefix(fileContentType, "video/") {
		fileType = "video"
	}

	// 上传文件
	s.logger.Info().Str("file_type", fileType).Msg("开始上传文件到MinIO")
	fileURL, objectName, err := minioClient.UploadFileWithObjectName(r.Context(), userID, fileType, fileSize, tempFile, fileHeader.Filename)
	if err != nil {
		s.logger.Error().Err(err).Msg("上传文件到MinIO失败")
		http.Error(w, "文件上传失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	s.logger.Info().
		Str("file_url", fileURL).
		Str("object_name", objectName).
		Msg("文件上传成功")

	// 准备媒体文件元数据
	// 对于图片，尝试获取宽度和高度
	width, height := 0, 0
	if strings.HasPrefix(fileContentType, "image/") {
		// 重置文件指针
		if _, err := tempFile.Seek(0, io.SeekStart); err == nil {
			if config, _, err := image.DecodeConfig(tempFile); err == nil {
				width = config.Width
				height = config.Height
				s.logger.Info().
					Int("width", width).
					Int("height", height).
					Msg("已获取图片尺寸")
			} else {
				s.logger.Warn().
					Err(err).
					Msg("无法获取图片尺寸")
			}
		}
	}

	// 保存到数据库（如有需要）
	media := &models.ContentMedia{
		UserID:      userID,
		MediaType:   fileType,
		URL:         fileURL,
		FileName:    fileHeader.Filename,
		FileSize:    fileHeader.Size,
		ContentType: fileContentType,
		ObjectName:  objectName,
		Width:       width,
		Height:      height,
		Duration:    0,                          // 视频时长，需要特殊处理
		Description: r.FormValue("description"), // 尝试从表单获取描述
	}

	// 检查是否存在重复的媒体文件
	duplicateMedia, err := s.repo.FindDuplicateMedia(r.Context(), userID, fileType, fileContentType, fileHeader.Size, objectName)
	if err != nil {
		s.logger.Warn().
			Err(err).
			Msg("检查重复媒体文件失败")
	}

	if duplicateMedia != nil {
		s.logger.Warn().
			Str("duplicate_id", duplicateMedia.ID).
			Str("duplicate_url", duplicateMedia.URL).
			Msg("发现重复的媒体文件，使用现有记录")

		// 使用已存在的媒体记录
		media = duplicateMedia
	} else {
		// 保存媒体记录到数据库
		err = s.repo.SaveContentMedia(r.Context(), media)
		if err != nil {
			s.logger.Error().Err(err).Msg("保存媒体记录失败")
			http.Error(w, "保存媒体记录失败: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// 返回结果
	response := map[string]interface{}{
		"success":     true,
		"url":         fileURL,
		"type":        fileType,
		"size":        fileHeader.Size,
		"width":       width,
		"height":      height,
		"media_id":    media.ID,
		"object_name": objectName,
	}

	s.logger.Info().Str("response", fmt.Sprintf("%+v", response)).Msg("准备返回响应")

	if err := json.NewEncoder(w).Encode(response); err != nil {
		s.logger.Error().Err(err).Msg("编码响应失败")
		http.Error(w, "内部服务器错误", http.StatusInternalServerError)
		return
	}

	s.logger.Info().Msg("文件上传处理完成")
}

// VotePollHandler 处理对帖子进行投票的请求
func (s *ContentServiceImpl) VotePollHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 获取用户ID
	userID := getUserIDFromContext(ctx)
	if userID == "" {
		http.Error(w, "未授权", http.StatusUnauthorized)
		return
	}

	// 获取帖子ID
	postID := mux.Vars(r)["postID"]
	if postID == "" {
		http.Error(w, "缺少帖子ID", http.StatusBadRequest)
		return
	}

	// 解析请求体
	var req struct {
		OptionID string `json:"option_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "无效的请求体", http.StatusBadRequest)
		return
	}

	// 检查选项ID是否为空
	if req.OptionID == "" {
		http.Error(w, "选项ID不能为空", http.StatusBadRequest)
		return
	}

	// 检查用户是否已经投票
	voted, _, err := s.repo.HasVoted(ctx, userID, postID)
	if err != nil {
		http.Error(w, "检查投票状态失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if voted {
		http.Error(w, "您已经在此投票中投过票", http.StatusBadRequest)
		return
	}

	// 创建投票记录
	vote := &models.PollVote{
		ID:        generateUUID(),
		UserID:    userID,
		PostID:    postID,
		OptionID:  req.OptionID,
		CreatedAt: time.Now(),
	}

	// 提交投票
	if err := s.repo.VoteInPoll(ctx, vote); err != nil {
		http.Error(w, "投票失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 获取更新后的投票结果
	options, total, err := s.repo.GetPollVotes(ctx, postID)
	if err != nil {
		http.Error(w, "获取投票结果失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 返回投票结果
	resp := struct {
		Success  bool                `json:"success"`
		Options  *models.PollOptions `json:"options"`
		Total    int64               `json:"total"`
		YourVote string              `json:"your_vote"`
	}{
		Success:  true,
		Options:  options,
		Total:    total,
		YourVote: req.OptionID,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "编码响应失败", http.StatusInternalServerError)
		return
	}
}

// GetPollResultsHandler 处理获取投票结果的请求
func (s *ContentServiceImpl) GetPollResultsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 获取帖子ID
	postID := mux.Vars(r)["postID"]
	if postID == "" {
		http.Error(w, "缺少帖子ID", http.StatusBadRequest)
		return
	}

	// 获取用户ID（如果已登录）
	userID := getUserIDFromContext(ctx)

	// 获取投票结果
	options, total, err := s.repo.GetPollVotes(ctx, postID)
	if err != nil {
		http.Error(w, "获取投票结果失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 检查用户是否已投票
	var voted bool
	var userVoteOptionID string
	if userID != "" {
		voted, userVoteOptionID, err = s.repo.HasVoted(ctx, userID, postID)
		if err != nil {
			http.Error(w, "检查投票状态失败: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// 构建响应
	resp := struct {
		Options  *models.PollOptions `json:"options"`
		Total    int64               `json:"total"`
		HasVoted bool                `json:"has_voted"`
		YourVote string              `json:"your_vote,omitempty"`
	}{
		Options:  options,
		Total:    total,
		HasVoted: voted,
	}

	if voted {
		resp.YourVote = userVoteOptionID
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "编码响应失败", http.StatusInternalServerError)
		return
	}
}

// SavePostHandler 处理保存帖子的请求
func (s *ContentServiceImpl) SavePostHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 获取用户ID
	userID := getUserIDFromContext(ctx)
	if userID == "" {
		http.Error(w, "未授权", http.StatusUnauthorized)
		return
	}

	// 获取帖子ID
	postID := mux.Vars(r)["postID"]
	if postID == "" {
		http.Error(w, "缺少帖子ID", http.StatusBadRequest)
		return
	}

	// 解析请求体
	var req struct {
		Collection string `json:"collection,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// 如果请求体为空或解析错误，使用默认收藏夹
		req.Collection = "default"
	}

	// 检查帖子是否已被保存
	saved, err := s.repo.IsSaved(ctx, userID, postID)
	if err != nil {
		http.Error(w, "检查保存状态失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if saved {
		http.Error(w, "帖子已被保存", http.StatusBadRequest)
		return
	}

	// 创建保存记录
	savedPost := &models.SavedPost{
		ID:         generateUUID(),
		UserID:     userID,
		PostID:     postID,
		Collection: req.Collection,
		CreatedAt:  time.Now(),
	}

	// 保存帖子
	if err := s.repo.SavePost(ctx, savedPost); err != nil {
		http.Error(w, "保存帖子失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 返回成功响应
	resp := struct {
		Success    bool   `json:"success"`
		Collection string `json:"collection"`
	}{
		Success:    true,
		Collection: req.Collection,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "编码响应失败", http.StatusInternalServerError)
		return
	}
}

// UnsavePostHandler 处理取消保存帖子的请求
func (s *ContentServiceImpl) UnsavePostHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 获取用户ID
	userID := getUserIDFromContext(ctx)
	if userID == "" {
		http.Error(w, "未授权", http.StatusUnauthorized)
		return
	}

	// 获取帖子ID
	postID := mux.Vars(r)["postID"]
	if postID == "" {
		http.Error(w, "缺少帖子ID", http.StatusBadRequest)
		return
	}

	// 检查帖子是否已被保存
	saved, err := s.repo.IsSaved(ctx, userID, postID)
	if err != nil {
		http.Error(w, "检查保存状态失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if !saved {
		http.Error(w, "帖子未被保存", http.StatusBadRequest)
		return
	}

	// 取消保存帖子
	if err := s.repo.UnsavePost(ctx, userID, postID); err != nil {
		http.Error(w, "取消保存帖子失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 返回成功响应
	resp := struct {
		Success bool `json:"success"`
	}{
		Success: true,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "编码响应失败", http.StatusInternalServerError)
		return
	}
}

// GetSavedPostsHandler 处理获取已保存帖子的请求
func (s *ContentServiceImpl) GetSavedPostsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 获取用户ID
	userID := getUserIDFromContext(ctx)
	if userID == "" {
		http.Error(w, "未授权", http.StatusUnauthorized)
		return
	}

	// 获取分页参数
	offset, limit := getPaginationParams(r)

	// 获取收藏夹参数
	collection := r.URL.Query().Get("collection")

	// 获取保存的帖子
	posts, total, err := s.repo.GetSavedPosts(ctx, userID, collection, offset, limit)
	if err != nil {
		http.Error(w, "获取保存的帖子失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 构建响应
	resp := struct {
		Posts []*models.Post `json:"posts"`
		Total int64          `json:"total"`
		Page  int            `json:"page"`
		Size  int            `json:"size"`
	}{
		Posts: posts,
		Total: total,
		Page:  offset/limit + 1,
		Size:  limit,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "编码响应失败", http.StatusInternalServerError)
		return
	}
}

// GetTopPostsHandler 处理获取热门帖子的请求
func (s *ContentServiceImpl) GetTopPostsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 获取时间周期参数（默认为周）
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "week"
	}

	// 验证时间周期参数
	validPeriods := map[string]bool{
		"day":   true,
		"week":  true,
		"month": true,
		"year":  true,
		"all":   true,
	}

	if !validPeriods[period] {
		http.Error(w, "无效的时间周期参数: "+period, http.StatusBadRequest)
		return
	}

	// 获取限制参数（默认为10）
	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		var err error
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit <= 0 || limit > 50 {
			http.Error(w, "无效的限制参数", http.StatusBadRequest)
			return
		}
	}

	// 获取热门帖子
	posts, err := s.repo.GetTopPosts(ctx, period, limit)
	if err != nil {
		http.Error(w, "获取热门帖子失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 构建响应
	resp := struct {
		Posts  []*models.Post `json:"posts"`
		Period string         `json:"period"`
		Limit  int            `json:"limit"`
	}{
		Posts:  posts,
		Period: period,
		Limit:  limit,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "编码响应失败", http.StatusInternalServerError)
		return
	}
}

// GetPostStatsHandler 处理获取帖子统计数据的请求
func (s *ContentServiceImpl) GetPostStatsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 获取帖子ID
	postID := mux.Vars(r)["postID"]
	if postID == "" {
		http.Error(w, "缺少帖子ID", http.StatusBadRequest)
		return
	}

	// 获取帖子统计数据
	stats, err := s.repo.GetPostStats(ctx, postID)
	if err != nil {
		http.Error(w, "获取帖子统计数据失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 返回统计数据
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		http.Error(w, "编码响应失败", http.StatusInternalServerError)
		return
	}
}

// ReportPostHandler 处理举报帖子的请求
func (s *ContentServiceImpl) ReportPostHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 获取用户ID
	userID := getUserIDFromContext(ctx)
	if userID == "" {
		http.Error(w, "未授权", http.StatusUnauthorized)
		return
	}

	// 获取帖子ID
	postID := mux.Vars(r)["postID"]
	if postID == "" {
		http.Error(w, "缺少帖子ID", http.StatusBadRequest)
		return
	}

	// 解析请求体
	var req struct {
		Reason      string `json:"reason"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "无效的请求体", http.StatusBadRequest)
		return
	}

	if req.Reason == "" {
		http.Error(w, "举报原因不能为空", http.StatusBadRequest)
		return
	}

	// 创建举报记录
	report := &models.PostReport{
		ID:          generateUUID(),
		PostID:      postID,
		ReporterID:  userID,
		ReasonCode:  req.Reason,
		Description: req.Description,
		Status:      "pending",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 保存举报
	if err := s.repo.CreatePostReport(ctx, report); err != nil {
		http.Error(w, "创建举报失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 返回成功响应
	resp := struct {
		Success  bool   `json:"success"`
		ReportID string `json:"report_id"`
	}{
		Success:  true,
		ReportID: report.ID,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "编码响应失败", http.StatusInternalServerError)
		return
	}
}

// GetUserFeedHandler 处理获取用户Feed的请求
func (s *ContentServiceImpl) GetUserFeedHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 获取用户ID
	userID := getUserIDFromContext(ctx)
	if userID == "" {
		http.Error(w, "未授权", http.StatusUnauthorized)
		return
	}

	s.logger.Info().Str("user_id", userID).Msg("处理用户Feed请求")

	// 获取分页参数
	offset, limit := getPaginationParams(r)

	// 获取用户的Feed内容，这里简单实现为获取最新内容
	// 实际上应该根据推荐算法来获取
	posts, total, err := s.repo.ListPosts(ctx, offset, limit)
	if err != nil {
		s.logger.Error().Err(err).Str("user_id", userID).Msg("获取Feed内容失败")
		http.Error(w, "获取Feed内容失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 确保每个帖子都包含User信息
	for _, post := range posts {
		if post.User.ID == "" {
			// 如果没有加载用户信息，这里可以记录一个警告日志
			s.logger.Warn().Str("post_id", post.ID).Str("user_id", post.UserID).Msg("帖子缺少用户信息")
		}
	}

	// 构建响应
	resp := struct {
		Posts []*models.Post `json:"posts"`
		Total int64          `json:"total"`
		Page  int            `json:"page"`
		Size  int            `json:"size"`
	}{
		Posts: posts,
		Total: total,
		Page:  offset/limit + 1,
		Size:  limit,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.logger.Error().Err(err).Msg("编码响应失败")
		http.Error(w, "内部服务器错误", http.StatusInternalServerError)
		return
	}
}

// GetPostByPermalinkHandler 通过用户名和永久链接ID获取帖子
func (s *ContentServiceImpl) GetPostByPermalinkHandler(w http.ResponseWriter, r *http.Request) {
	// 从URL路径中获取用户名和永久链接ID
	vars := mux.Vars(r)
	username := vars["username"]
	permalinkID := vars["permalink_id"]

	// 打印所有请求参数和路径
	s.logger.Info().
		Str("username", username).
		Str("permalink_id", permalinkID).
		Str("path", r.URL.Path).
		Str("method", r.Method).
		Str("query", r.URL.RawQuery).
		Str("referer", r.Header.Get("Referer")).
		Str("auth", fmt.Sprintf("%t", r.Header.Get("Authorization") != "")).
		Str("handler", "GetPostByPermalinkHandler").
		Str("vars", fmt.Sprintf("%+v", vars)).
		Msg("通过永久链接获取帖子 - 详细请求信息")

	if username == "" || permalinkID == "" {
		s.logger.Error().
			Str("username", username).
			Str("permalink_id", permalinkID).
			Str("path", r.URL.Path).
			Str("vars", fmt.Sprintf("%+v", vars)).
			Msg("无效的请求参数 - 缺少用户名或永久链接ID")
		http.Error(w, "必须提供用户名和永久链接ID", http.StatusBadRequest)
		return
	}

	// 通过永久链接查找帖子
	s.logger.Info().
		Str("permalink_id", permalinkID).
		Msg("开始查询帖子")

	post, err := s.repo.GetPostByPermalink(r.Context(), permalinkID)
	if err != nil {
		s.logger.Error().
			Err(err).
			Str("permalink_id", permalinkID).
			Str("path", r.URL.Path).
			Str("method", r.Method).
			Msg("获取帖子失败")

		// 判断错误类型
		if strings.Contains(err.Error(), "not found") {
			s.logger.Error().
				Err(err).
				Str("permalink_id", permalinkID).
				Str("username", username).
				Str("database_query", "SELECT * FROM flick_posts WHERE permalink_id = ?").
				Msg("数据库中未找到对应ID的帖子")
			http.Error(w, fmt.Sprintf("帖子不存在: %s", err.Error()), http.StatusNotFound)
		} else {
			s.logger.Error().
				Err(err).
				Str("permalink_id", permalinkID).
				Str("database_error", err.Error()).
				Msg("数据库查询出错")
			http.Error(w, fmt.Sprintf("获取帖子失败: %s", err.Error()), http.StatusInternalServerError)
		}
		return
	}

	// 成功获取到帖子后，打印详细信息
	s.logger.Info().
		Str("post_id", post.ID).
		Str("permalink_id", post.PermalinkID).
		Str("user_id", post.UserID).
		Str("username", post.User.Username).
		Str("display_name", post.User.DisplayName).
		Str("avatar_url", post.User.AvatarURL).
		Int("media_files_count", len(post.MediaFiles)).
		Str("content_preview", post.Content[:min(20, len(post.Content))]).
		Msg("成功查询到帖子 - 详细信息")

	// 验证帖子是否属于请求的用户名
	if post.User.Username != username {
		// 如果用户名不匹配但找到了帖子，可能只是用户信息没有正确加载
		// 先尝试获取正确的用户信息
		user, userErr := s.repo.GetUserByID(r.Context(), post.UserID)
		if userErr != nil {
			s.logger.Warn().
				Err(userErr).
				Str("user_id", post.UserID).
				Msg("获取用户信息失败")
			// 用户信息获取失败，不作为致命错误处理
		} else {
			post.User = *user
			// 重新检查用户名是否匹配
			if post.User.Username != username {
				s.logger.Warn().
					Str("permalink_id", permalinkID).
					Str("requested_username", username).
					Str("actual_username", post.User.Username).
					Msg("请求的用户名与帖子所属用户不匹配")
				http.Error(w, fmt.Sprintf("帖子不存在或不属于用户 %s", username), http.StatusNotFound)
				return
			}
		}
	}

	// 检查帖子是否有User字段，如果没有则加载
	if post.User.ID == "" {
		user, err := s.repo.GetUserByID(r.Context(), post.UserID)
		if err != nil {
			s.logger.Warn().
				Err(err).
				Str("user_id", post.UserID).
				Msg("获取用户信息失败")
		} else {
			post.User = *user
		}
	}

	// 获取当前登录用户（如果有）以更新点赞/收藏状态
	currentUserID := ""
	if userIDVal := r.Context().Value("user_id"); userIDVal != nil {
		if userID, ok := userIDVal.(string); ok {
			currentUserID = userID
		}
	}

	// 如果有登录用户，检查该用户是否已点赞/收藏帖子
	if currentUserID != "" {
		// 检查是否点赞
		isLiked, err := s.repo.IsPostLikedByUser(r.Context(), post.ID, currentUserID)
		if err != nil {
			s.logger.Warn().
				Err(err).
				Str("post_id", post.ID).
				Str("user_id", currentUserID).
				Msg("检查点赞状态失败")
		} else {
			post.IsLiked = &isLiked
		}

		// 检查是否收藏
		isSaved, err := s.repo.IsSaved(r.Context(), currentUserID, post.ID)
		if err != nil {
			s.logger.Warn().
				Err(err).
				Str("post_id", post.ID).
				Str("user_id", currentUserID).
				Msg("检查收藏状态失败")
		} else {
			post.IsSaved = &isSaved
		}
	}

	// 增加帖子查看次数
	go func() {
		if err := s.repo.IncrementViewCount(context.Background(), post.ID); err != nil {
			s.logger.Warn().Err(err).Str("post_id", post.ID).Msg("增加查看次数失败")
		}
	}()

	// 返回帖子信息
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(post); err != nil {
		s.logger.Error().Err(err).Msg("编码帖子失败")
		http.Error(w, "内部服务器错误", http.StatusInternalServerError)
		return
	}
}

// GetPostMediaByIndexHandler 获取帖子中特定索引的媒体文件
func (s *ContentServiceImpl) GetPostMediaByIndexHandler(w http.ResponseWriter, r *http.Request) {
	// 从URL路径中获取用户名、永久链接ID和索引
	vars := mux.Vars(r)
	username := vars["username"]
	permalinkID := vars["permalink_id"]
	indexStr := vars["index"]

	// 解析索引
	index, err := strconv.Atoi(indexStr)
	if err != nil || index < 1 {
		s.logger.Error().Err(err).Str("index", indexStr).Msg("无效的媒体索引")
		http.Error(w, "无效的媒体索引", http.StatusBadRequest)
		return
	}

	// 获取帖子
	post, err := s.repo.GetPostByPermalink(r.Context(), permalinkID)
	if err != nil {
		s.logger.Error().Err(err).Str("permalink_id", permalinkID).Msg("获取帖子失败")
		http.Error(w, "帖子不存在", http.StatusNotFound)
		return
	}

	// 验证帖子是否属于请求的用户名
	if post.User.Username != username {
		s.logger.Warn().Str("permalink_id", permalinkID).Str("username", username).Msg("请求的用户名与帖子所属用户不匹配")
		http.Error(w, "帖子不存在", http.StatusNotFound)
		return
	}

	// 检查索引是否超出媒体文件范围
	if index > len(post.MediaFiles) {
		s.logger.Error().Int("index", index).Int("media_count", len(post.MediaFiles)).Msg("媒体索引超出范围")
		http.Error(w, "媒体索引超出范围", http.StatusBadRequest)
		return
	}

	// 获取指定索引的媒体文件（索引从1开始，数组从0开始）
	mediaFile := post.MediaFiles[index-1]

	// 返回媒体文件信息
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(mediaFile); err != nil {
		s.logger.Error().Err(err).Msg("编码媒体文件失败")
		http.Error(w, "内部服务器错误", http.StatusInternalServerError)
		return
	}
}

// UpdatePostByPermalinkHandler 通过用户名和永久链接ID更新帖子
func (s *ContentServiceImpl) UpdatePostByPermalinkHandler(w http.ResponseWriter, r *http.Request) {
	// 从URL路径中获取用户名和永久链接ID
	vars := mux.Vars(r)
	username := vars["username"]
	permalinkID := vars["permalink_id"]

	// 获取当前用户ID
	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		s.logger.Error().Msg("未找到用户ID")
		http.Error(w, "未授权", http.StatusUnauthorized)
		return
	}

	// 获取帖子
	post, err := s.repo.GetPostByPermalink(r.Context(), permalinkID)
	if err != nil {
		s.logger.Error().Err(err).Str("permalink_id", permalinkID).Msg("获取帖子失败")
		http.Error(w, "帖子不存在", http.StatusNotFound)
		return
	}

	// 验证帖子是否属于请求的用户名
	if post.User.Username != username {
		s.logger.Warn().Str("permalink_id", permalinkID).Str("username", username).Msg("请求的用户名与帖子所属用户不匹配")
		http.Error(w, "帖子不存在", http.StatusNotFound)
		return
	}

	// 验证当前用户是否有权限更新帖子
	if post.UserID != userID {
		s.logger.Warn().Str("post_user_id", post.UserID).Str("current_user_id", userID).Msg("无权更新帖子")
		http.Error(w, "无权更新此帖子", http.StatusForbidden)
		return
	}

	// 解析请求体
	var updateInput struct {
		Content    *string            `json:"content,omitempty"`
		MediaFiles *models.MediaFiles `json:"media_files,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&updateInput); err != nil {
		s.logger.Error().Err(err).Msg("解析请求体失败")
		http.Error(w, "无效的请求数据", http.StatusBadRequest)
		return
	}

	// 更新帖子内容
	if updateInput.Content != nil {
		post.Content = *updateInput.Content
	}

	// 更新媒体文件
	if updateInput.MediaFiles != nil {
		post.MediaFiles = *updateInput.MediaFiles
	}

	// 保存更新
	if err := s.repo.UpdatePost(r.Context(), post); err != nil {
		s.logger.Error().Err(err).Str("post_id", post.ID).Msg("更新帖子失败")
		http.Error(w, "更新帖子失败", http.StatusInternalServerError)
		return
	}

	// 记录审计日志
	auditLog := &models.PostAuditLog{
		PostID:     post.ID,
		UserID:     userID,
		ActionType: "update",
		NewValue:   post.Content,
	}
	if err := s.repo.LogPostActivity(r.Context(), auditLog); err != nil {
		s.logger.Warn().Err(err).Msg("记录审计日志失败")
	}

	// 返回更新后的帖子
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(post); err != nil {
		s.logger.Error().Err(err).Msg("编码帖子失败")
		http.Error(w, "内部服务器错误", http.StatusInternalServerError)
	}
}

// DeletePostByPermalinkHandler 通过用户名和永久链接ID删除帖子
func (s *ContentServiceImpl) DeletePostByPermalinkHandler(w http.ResponseWriter, r *http.Request) {
	// 从URL路径中获取用户名和永久链接ID
	vars := mux.Vars(r)
	username := vars["username"]
	permalinkID := vars["permalink_id"]

	// 获取当前用户ID
	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		s.logger.Error().Msg("未找到用户ID")
		http.Error(w, "未授权", http.StatusUnauthorized)
		return
	}

	// 获取帖子
	post, err := s.repo.GetPostByPermalink(r.Context(), permalinkID)
	if err != nil {
		s.logger.Error().Err(err).Str("permalink_id", permalinkID).Msg("获取帖子失败")
		http.Error(w, "帖子不存在", http.StatusNotFound)
		return
	}

	// 验证帖子是否属于请求的用户名
	if post.User.Username != username {
		s.logger.Warn().Str("permalink_id", permalinkID).Str("username", username).Msg("请求的用户名与帖子所属用户不匹配")
		http.Error(w, "帖子不存在", http.StatusNotFound)
		return
	}

	// 验证当前用户是否有权限删除帖子
	if post.UserID != userID {
		s.logger.Warn().Str("post_user_id", post.UserID).Str("current_user_id", userID).Msg("无权删除帖子")
		http.Error(w, "无权删除此帖子", http.StatusForbidden)
		return
	}

	// 删除帖子（软删除）
	if err := s.repo.DeletePost(r.Context(), post.ID); err != nil {
		s.logger.Error().Err(err).Str("post_id", post.ID).Msg("删除帖子失败")
		http.Error(w, "删除帖子失败", http.StatusInternalServerError)
		return
	}

	// 记录审计日志
	auditLog := &models.PostAuditLog{
		PostID:     post.ID,
		UserID:     userID,
		ActionType: "delete",
	}
	if err := s.repo.LogPostActivity(r.Context(), auditLog); err != nil {
		s.logger.Warn().Err(err).Msg("记录审计日志失败")
	}

	// 返回成功响应
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "帖子已删除"})
}

// getUserIDFromContext 从上下文中获取用户ID
func getUserIDFromContext(ctx context.Context) string {
	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		return ""
	}
	return userID
}

// getPaginationParams 从请求中获取分页参数
func getPaginationParams(r *http.Request) (int, int) {
	pageStr := r.URL.Query().Get("page")
	sizeStr := r.URL.Query().Get("size")

	page := 1
	size := 10

	if pageStr != "" {
		parsedPage, err := strconv.Atoi(pageStr)
		if err == nil && parsedPage > 0 {
			page = parsedPage
		}
	}

	if sizeStr != "" {
		parsedSize, err := strconv.Atoi(sizeStr)
		if err == nil && parsedSize > 0 && parsedSize <= 50 {
			size = parsedSize
		}
	}

	offset := (page - 1) * size
	return offset, size
}

// generateUUID 生成UUID
func generateUUID() string {
	// 在实际应用中，应该使用uuid包生成真正的UUID
	// 这里简化处理，返回一个占位符
	return "temp-uuid"
}

// timePtr 返回时间指针
func timePtr(t time.Time) *time.Time {
	return &t
}

// GetPostCommentsHandler 处理获取帖子评论的请求
func (s *ContentServiceImpl) GetPostCommentsHandler(w http.ResponseWriter, r *http.Request) {
	// 设置内容类型
	w.Header().Set("Content-Type", "application/json")

	// 从URL获取用户名和permalinkID
	vars := mux.Vars(r)
	username := vars["username"]
	permalinkID := vars["permalink_id"]

	if username == "" || permalinkID == "" {
		http.Error(w, "缺少必要的参数", http.StatusBadRequest)
		return
	}

	// 获取分页参数
	offset, limit := getPaginationParams(r)

	// 先通过permalinkID获取帖子ID
	ctx := r.Context()
	post, err := s.repo.GetPostByPermalink(ctx, permalinkID)
	if err != nil {
		s.logger.Error().Err(err).Str("permalink_id", permalinkID).Msg("获取帖子失败")
		http.Error(w, "帖子不存在", http.StatusNotFound)
		return
	}

	// 获取当前用户ID (如果已登录)
	userID := getUserIDFromContext(ctx)

	// 构建并发送响应
	response := map[string]interface{}{
		"message":      "获取评论功能尚未实现",
		"status":       "pending",
		"post_id":      post.ID,
		"permalink_id": permalinkID,
		"username":     username,
		"user_id":      userID,
		"pagination": map[string]int{
			"offset": offset,
			"limit":  limit,
		},
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		s.logger.Error().Err(err).Msg("编码响应失败")
		http.Error(w, "内部服务器错误", http.StatusInternalServerError)
	}
}

// AddCommentToPostHandler 处理添加评论到帖子的请求
func (s *ContentServiceImpl) AddCommentToPostHandler(w http.ResponseWriter, r *http.Request) {
	// 设置内容类型
	w.Header().Set("Content-Type", "application/json")

	// 从URL获取用户名和permalinkID
	vars := mux.Vars(r)
	username := vars["username"]
	permalinkID := vars["permalink_id"]

	if username == "" || permalinkID == "" {
		http.Error(w, "缺少必要的参数", http.StatusBadRequest)
		return
	}

	// 解析请求体
	var commentInput struct {
		Content string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&commentInput); err != nil {
		s.logger.Error().Err(err).Msg("解析评论内容失败")
		http.Error(w, "无效的请求数据", http.StatusBadRequest)
		return
	}

	// 从上下文中获取用户ID
	ctx := r.Context()
	userID := getUserIDFromContext(ctx)
	if userID == "" {
		s.logger.Error().Msg("未找到用户ID")
		http.Error(w, "未授权", http.StatusUnauthorized)
		return
	}

	// 先通过permalinkID获取帖子ID
	post, err := s.repo.GetPostByPermalink(ctx, permalinkID)
	if err != nil {
		s.logger.Error().Err(err).Str("permalink_id", permalinkID).Msg("获取帖子失败")
		http.Error(w, "帖子不存在", http.StatusNotFound)
		return
	}

	// 构建并发送响应
	response := map[string]interface{}{
		"message":      "添加评论功能尚未实现",
		"status":       "pending",
		"post_id":      post.ID,
		"permalink_id": permalinkID,
		"username":     username,
		"user_id":      userID,
		"content":      commentInput.Content,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		s.logger.Error().Err(err).Msg("编码响应失败")
		http.Error(w, "内部服务器错误", http.StatusInternalServerError)
	}
}

// LikePostHandler 处理对帖子点赞的请求
func (s *ContentServiceImpl) LikePostHandler(w http.ResponseWriter, r *http.Request) {
	// 设置内容类型
	w.Header().Set("Content-Type", "application/json")

	// 从URL获取用户名和permalinkID
	vars := mux.Vars(r)
	username := vars["username"]
	permalinkID := vars["permalink_id"]

	if username == "" || permalinkID == "" {
		http.Error(w, "缺少必要的参数", http.StatusBadRequest)
		return
	}

	// 从上下文中获取用户ID
	ctx := r.Context()
	userID := getUserIDFromContext(ctx)
	if userID == "" {
		s.logger.Error().Msg("未找到用户ID")
		http.Error(w, "未授权", http.StatusUnauthorized)
		return
	}

	// 先通过permalinkID获取帖子ID
	post, err := s.repo.GetPostByPermalink(ctx, permalinkID)
	if err != nil {
		s.logger.Error().Err(err).Str("permalink_id", permalinkID).Msg("获取帖子失败")
		http.Error(w, "帖子不存在", http.StatusNotFound)
		return
	}

	// 构建并发送响应
	response := map[string]interface{}{
		"message":      "点赞功能尚未实现",
		"status":       "pending",
		"post_id":      post.ID,
		"permalink_id": permalinkID,
		"username":     username,
		"user_id":      userID,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		s.logger.Error().Err(err).Msg("编码响应失败")
		http.Error(w, "内部服务器错误", http.StatusInternalServerError)
	}
}

// UnlikePostHandler 处理取消对帖子点赞的请求
func (s *ContentServiceImpl) UnlikePostHandler(w http.ResponseWriter, r *http.Request) {
	// 设置内容类型
	w.Header().Set("Content-Type", "application/json")

	// 从URL获取用户名和permalinkID
	vars := mux.Vars(r)
	username := vars["username"]
	permalinkID := vars["permalink_id"]

	if username == "" || permalinkID == "" {
		http.Error(w, "缺少必要的参数", http.StatusBadRequest)
		return
	}

	// 从上下文中获取用户ID
	ctx := r.Context()
	userID := getUserIDFromContext(ctx)
	if userID == "" {
		s.logger.Error().Msg("未找到用户ID")
		http.Error(w, "未授权", http.StatusUnauthorized)
		return
	}

	// 先通过permalinkID获取帖子ID
	post, err := s.repo.GetPostByPermalink(ctx, permalinkID)
	if err != nil {
		s.logger.Error().Err(err).Str("permalink_id", permalinkID).Msg("获取帖子失败")
		http.Error(w, "帖子不存在", http.StatusNotFound)
		return
	}

	// 构建并发送响应
	response := map[string]interface{}{
		"message":      "取消点赞功能尚未实现",
		"status":       "pending",
		"post_id":      post.ID,
		"permalink_id": permalinkID,
		"username":     username,
		"user_id":      userID,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		s.logger.Error().Err(err).Msg("编码响应失败")
		http.Error(w, "内部服务器错误", http.StatusInternalServerError)
	}
}

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// GetUserPostsByUsernameHandler 处理获取用户帖子列表的请求
func (s *ContentServiceImpl) GetUserPostsByUsernameHandler(w http.ResponseWriter, r *http.Request) {
	// 设置内容类型
	w.Header().Set("Content-Type", "application/json")

	// 从URL获取用户名
	vars := mux.Vars(r)
	username := vars["username"]

	if username == "" {
		s.logger.Error().Msg("缺少用户名参数")
		http.Error(w, "缺少必要的参数", http.StatusBadRequest)
		return
	}

	s.logger.Info().Str("username", username).Msg("获取用户帖子列表")

	// 获取分页参数
	offset, limit := getPaginationParams(r)

	ctx := r.Context()

	// 获取所有帖子
	posts, _, err := s.repo.ListPosts(ctx, 0, 1000) // 先获取较大数量的帖子
	if err != nil {
		s.logger.Error().Err(err).Msg("获取帖子列表失败")
		http.Error(w, "获取帖子失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 过滤出指定用户名的帖子
	var userPosts []*models.Post
	for _, post := range posts {
		if post.User.Username == username {
			userPosts = append(userPosts, post)
		}
	}

	// 重新计算总数
	filteredTotal := int64(len(userPosts))

	// 应用分页
	start := int(offset)
	end := int(offset + limit)
	if int64(start) > filteredTotal {
		start = int(filteredTotal)
	}
	if int64(end) > filteredTotal {
		end = int(filteredTotal)
	}

	var pagedPosts []*models.Post
	if start < end {
		pagedPosts = userPosts[start:end]
	}

	// 构建响应
	resp := struct {
		Posts []*models.Post `json:"posts"`
		Total int64          `json:"total"`
		Page  int            `json:"page"`
		Size  int            `json:"size"`
	}{
		Posts: pagedPosts,
		Total: filteredTotal,
		Page:  offset/limit + 1,
		Size:  limit,
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.logger.Error().Err(err).Msg("编码响应失败")
		http.Error(w, "内部服务器错误", http.StatusInternalServerError)
		return
	}
}

// GetUserPostsHandler 处理获取当前用户帖子列表的请求（通过用户ID）
func (s *ContentServiceImpl) GetUserPostsHandler(w http.ResponseWriter, r *http.Request) {
	// 设置内容类型
	w.Header().Set("Content-Type", "application/json")

	// 从URL参数获取用户ID，这通常是从上下文中设置的
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		// 尝试从上下文获取
		if ctx := r.Context(); ctx != nil {
			if id, ok := ctx.Value("user_id").(string); ok && id != "" {
				userID = id
			}
		}
	}

	if userID == "" {
		s.logger.Error().Msg("缺少用户ID参数")
		http.Error(w, "缺少必要的参数", http.StatusBadRequest)
		return
	}

	s.logger.Info().Str("user_id", userID).Msg("通过用户ID获取帖子列表")

	// 获取分页参数
	offset, limit := getPaginationParams(r)

	ctx := r.Context()

	// 直接从数据库获取该用户的帖子
	posts, total, err := s.repo.GetPostsByUser(ctx, userID, offset, limit)
	if err != nil {
		s.logger.Error().Err(err).Str("user_id", userID).Msg("获取用户帖子失败")
		http.Error(w, "获取帖子失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 构建响应
	resp := struct {
		Posts []*models.Post `json:"posts"`
		Total int64          `json:"total"`
		Page  int            `json:"page"`
		Size  int            `json:"size"`
	}{
		Posts: posts,
		Total: total,
		Page:  offset/limit + 1,
		Size:  limit,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.logger.Error().Err(err).Msg("编码响应失败")
		http.Error(w, "内部服务器错误", http.StatusInternalServerError)
		return
	}
}
