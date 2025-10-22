package handlers

import (
	"github.com/gin-gonic/gin"
	"dongman/internal/config"
)

// SetupRoutes 设置API路由
func SetupRoutes(router *gin.Engine) {
	// 加载TMDB配置
	LoadTMDBConfig()
	
	// API路由组
	api := router.Group("/api")
	
	// CORS代理路由 - 无需认证，支持所有HTTP方法
	api.Any("/proxy", ProxyHandler)
	
	// 直接添加一个不带/api前缀的代理路由，适用于Vite代理重写后的路径
	router.Any("/proxy", ProxyHandler)
	
	// 认证路由
	auth := api.Group("/auth")
	{
		auth.POST("/token", Login)
		auth.GET("/me", JWTAuthMiddleware(), GetCurrentUserInfo)
		auth.POST("/change-password", JWTAuthMiddleware(), UpdatePassword)
	}
	
	// 网站设置路由
	settings := api.Group("/settings")
	{
		// 获取设置 - 公开API
		settings.GET("/:key", GetSiteSettings)
		settings.GET("/", GetAllSiteSettings)
		
		// TMDB状态API - 公开API，只返回是否启用，不返回密钥
		settings.GET("/tmdb_status", GetTMDBStatus)
		
		// 更新设置 - 需要管理员权限
		settings.PUT("/:key", JWTAuthMiddleware(), AdminAuthMiddleware(), UpdateSiteSettings)
	}
	
	// 管理员路由
	admin := api.Group("/admin", JWTAuthMiddleware(), AdminAuthMiddleware())
	{
		// 网站图标上传
		admin.POST("/upload/favicon", UploadFavicon)
		
		// TMDB配置
		admin.GET("/tmdb/config", GetTMDBConfig)
		admin.PUT("/tmdb/config", UpdateTMDBConfig)

		// 用户管理API
		admin.GET("/users", GetUsers)
		admin.GET("/users/roles", GetUserRoles)
		admin.POST("/users", CreateUser)
		admin.PUT("/users/:id", UpdateUser)
		admin.DELETE("/users/:id", DeleteUser)
	}

	// 图像处理工具路由
	imgtools := api.Group("/imgtools")
	{
		// 图像超分辨率API - 公开接口
		imgtools.POST("/enhance", EnhanceImageHandler)
	}
	
	// TMDB API路由
	tmdb := api.Group("/tmdb")
	{
		tmdb.GET("/search", SearchTMDB)
		tmdb.GET("/search_id", SearchTmdbId)
		tmdb.POST("/create", CreateResourceFromTMDB)
		tmdb.GET("/check-exists", CheckResourceExists)
		
		// 添加新的季节和剧集API路由
		tmdb.GET("/seasons/:series_id", GetTMDBSeasons)
		tmdb.GET("/seasons/:series_id/:season_number", GetTMDBEpisodes)
		tmdb.GET("/seasons/:series_id/:season_number/:episode_number/images", GetTMDBEpisodeImages)
		tmdb.GET("/seasons/:series_id/:season_number/:episode_number/credits", GetTMDBEpisodeCredits)
		tmdb.GET("/episode/:series_id/:season_number/:episode_number", GetEpisodeInfo)
		
		// 批量获取剧集信息的API
		tmdb.POST("/episodes/batch", GetBatchEpisodeInfo)
		
		// 通过TMDB ID查找本地资源
		tmdb.GET("/resource/:tmdb_id", GetResourceByTMDBID)
		
		// 更新资源的TMDB ID
		tmdb.PUT("/update-resource-id/:id/:tmdb_id", UpdateResourceTmdbID)
		
		// 添加新的多类型搜索API
		tmdb.GET("/multi_search", MultiSearchTMDB)
		
		// 添加新的媒体详情API
		tmdb.GET("/details/:media_type/:media_id", GetMediaDetails)
	}
	
	// 资源路由 - 需要认证
	resources := api.Group("/resources")
	{
		// 公开API - 无需认证
		resources.GET("/public", GetPublicResources)
		resources.GET("/:id", GetResourceByID)
		resources.POST("/:id/like", LikeResource)
		resources.POST("/:id/unlike", UnlikeResource)
		resources.PUT("/:id/supplement", SupplementResource)
		resources.PUT("/:id/stickers", UpdateResourceStickers)
		resources.POST("/:id/update-tmdb", UpdateResourceTMDBInfo)

		resources.POST("/", CreateResource)

		// 图片上传API - 处理不同URL路径格式
		resources.POST("/upload-images", UploadImage)
		resources.POST("/upload-images/", UploadImage)  // 添加带斜杠版本的路由
		resources.GET("/upload-images", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "upload API ready"})
		})
		resources.GET("/upload-images/", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "upload API ready"})
		})

		// 需要认证的API
		authResources := resources.Group("/", JWTAuthMiddleware())
		{
			authResources.GET("/", GetResources)
			// authResources.POST("/", CreateResource)
		}
		
		// 仅管理员可访问的API
		adminResources := resources.Group("/", JWTAuthMiddleware(), AdminAuthMiddleware())
		{
			adminResources.GET("/pending", GetPendingResources)
			adminResources.GET("/pending-supplements", GetPendingSupplementResources)
			adminResources.GET("/:id/supplement", GetResourceSupplement)
			adminResources.PUT("/:id", UpdateResource)
			adminResources.PUT("/:id/approve", ApproveResource)
			adminResources.POST("/batch-reject", BatchRejectResources)
			adminResources.DELETE("/:id", DeleteResource)
			adminResources.DELETE("/:id/record", DeleteApprovalRecord)
			adminResources.DELETE("/batch-delete-records", DeleteApprovalRecords)
			adminResources.GET("/approval-records", GetApprovalRecords)
			adminResources.GET("/:id/approval-records", GetResourceApprovalRecords)
		}
	}
	
	// 文章管理路由
	posts := api.Group("/posts")
	{
		// 公开API - 无需认证
		posts.GET("/", GetAllPosts)
		posts.GET("/search", SearchPosts)
		posts.GET("/id/:id", GetPostByID)
		posts.GET("/slug/:slug", GetPostBySlug)
		
		// 仅管理员可访问的API
		adminPosts := posts.Group("/admin", JWTAuthMiddleware(), AdminAuthMiddleware())
		{
			adminPosts.POST("/", CreatePost)
			adminPosts.PUT("/:id", UpdatePost)
			adminPosts.DELETE("/:id", DeletePost)
			adminPosts.POST("/upload/image", UploadPostImage)
			adminPosts.POST("/upload/file", UploadPostFile)
		}
	}
	
	// 静态资源路由 - 用于访问上传的文件
	api.Static("/assets", config.AssetPath)
} 