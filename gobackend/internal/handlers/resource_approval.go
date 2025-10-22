package handlers

import (
	"dongman/internal/models"
	"dongman/internal/config"
	"dongman/internal/utils"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// ApproveResource 审批资源 - 仅管理员可访问
func ApproveResource(c *gin.Context) {
	// 获取资源ID
	resourceID, errParse := strconv.Atoi(c.Param("id"))
	if errParse != nil {
		log.Printf("[ERROR] 无效的资源ID: %v", errParse)
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的资源ID"})
		return
	}

	// 解析请求
	var approval models.ResourceApproval
	if errBind := c.ShouldBindJSON(&approval); errBind != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	// 检查资源是否存在
	var resource models.Resource
	errGet := models.DB.Get(&resource, `SELECT * FROM resources WHERE id = ?`, resourceID)
	if errGet != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "资源未找到"})
		return
	}

	// 如果是补充内容审批
	if resource.Supplement != nil && resource.IsSupplementApproval == false {
		log.Printf("当前是补充资源审批")
		approveResourceSupplement(c, resourceID, resource, approval)
		resource.Status = models.ResourceStatus(strings.ToUpper(string(approval.Status)))
		resource.UpdatedAt = time.Now()
		return
	}

	// 更新资源状态
	resource.Status = models.ResourceStatus(strings.ToUpper(string(approval.Status)))
	resource.UpdatedAt = time.Now()

	log.Printf("当前是初始资源审批 Received approval: %+v", approval)

	// 如果资源被批准，处理图片移动
	// 保存所有新路径
	newImagePaths := make([]string, 0, len(approval.ApprovedImages))
	if strings.ToLower(string(approval.Status)) == strings.ToLower(string(models.ResourceStatusApproved)){
		// 检查是否没有批准任何图片和链接
		if len(approval.ApprovedImages) == 0 && len(approval.ApprovedLinks) == 0 {
			log.Printf("[INFO] 资源ID: %d 被批准但没有批准任何图片和链接，将直接删除该资源", resourceID)
			
			// 删除资源
			_, errDelete := models.DB.Exec(`DELETE FROM resources WHERE id = ?`, resourceID)
			if errDelete != nil {
				log.Printf("[ERROR] 删除资源失败: %v", errDelete)
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("删除资源失败: %v", errDelete)})
				return
			}
			
			// 返回成功消息
			c.JSON(http.StatusOK, gin.H{
				"message": "资源已删除，因为没有批准任何图片和链接",
				"deleted": true,
				"resource_id": resourceID,
			})
			return
		}
		
		// 移动已批准的图片
		if len(approval.ApprovedImages) > 0 {
			log.Printf("[DEBUG] 开始移动已批准的图片，资源ID: %d, 图片数量: %d", resource.ID, len(approval.ApprovedImages))
			log.Printf("[DEBUG] 原始图片路径: %v", approval.ApprovedImages)

			// 获取assets目录路径
			assetsDir := config.GetAssetsDir()
			log.Printf("[DEBUG] Assets目录路径: %s", assetsDir)
			
			// 创建目标目录
			imgsDir := filepath.Join(assetsDir, "imgs", fmt.Sprintf("%d", resourceID))
			log.Printf("[DEBUG] 创建目标目录: %s", imgsDir)
			if errMkdir := os.MkdirAll(imgsDir, 0755); errMkdir != nil {
				log.Printf("[ERROR] 创建目录失败: %v", errMkdir)
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("创建图片目录失败: %v", errMkdir)})
				return
			}
			
			// 手动执行每个图片的移动操作
			for _, imgPath := range approval.ApprovedImages {
				if imgPath == "" {
					continue
				}
				
				// 检查是否为TMDB外部图片链接
				if strings.HasPrefix(imgPath, "@https://image.tmdb.org/") {
					// 对于TMDB外部图片链接，直接保存原始链接，无需移动
					log.Printf("[DEBUG] 检测到TMDB外部图片链接: %s，直接保存原始链接", imgPath)
					newImagePaths = append(newImagePaths, imgPath)
					continue
				}
				
				// 检查是否为TMDB外部图片链接（无@前缀）
				if strings.HasPrefix(imgPath, "https://image.tmdb.org/") {
					// 对于TMDB外部图片链接，直接保存原始链接，无需移动
					log.Printf("[DEBUG] 检测到TMDB外部图片链接: %s，直接保存原始链接", imgPath)
					newImagePaths = append(newImagePaths, imgPath)
					continue
				}
				
				// 提取文件名
				filename := filepath.Base(imgPath)
				log.Printf("[DEBUG] 处理图片: %s, 文件名: %s", imgPath, filename)
				
				// 源文件路径
				sourcePath := filepath.Join(assetsDir, imgPath[7:]) // 去掉前面的"/assets"
				
				// 目标文件路径
				destPath := filepath.Join(imgsDir, filename)
				
				log.Printf("[DEBUG] 移动图片: %s -> %s", sourcePath, destPath)
				
				// 检查源文件是否存在
				if _, errStat := os.Stat(sourcePath); os.IsNotExist(errStat) {
					log.Printf("[ERROR] 源文件不存在: %s", sourcePath)
					continue
				} else {
					log.Printf("[DEBUG] 源文件存在: %s", sourcePath)
				}
				
				// 确保目标目录存在
				errDir := os.MkdirAll(filepath.Dir(destPath), 0755)
				if errDir != nil {
					log.Printf("[ERROR] 创建目标目录失败: %v", errDir)
					continue
				}
				
				// 移动文件（复制后删除）
				// 1. 复制文件
				sourceFile, errOpen := os.Open(sourcePath)
				if errOpen != nil {
					log.Printf("[ERROR] 打开源文件失败: %v", errOpen)
					continue
				}
				defer sourceFile.Close()

				// 创建目标文件
				destFile, errCreate := os.Create(destPath)
				if errCreate != nil {
					log.Printf("[ERROR] 创建目标文件失败: %v", errCreate)
					continue
				}
				defer destFile.Close()

				// 复制内容
				_, errCopy := io.Copy(destFile, sourceFile)
				if errCopy != nil {
					log.Printf("[ERROR] 复制文件内容失败: %v", errCopy)
					continue
				}

				// 关闭文件以确保所有内容都已写入
				errSource := sourceFile.Close()
				if errSource != nil {
					log.Printf("[ERROR] 关闭源文件失败: %v", errSource)
				}

				errDest := destFile.Close()
				if errDest != nil {
					log.Printf("[ERROR] 关闭目标文件失败: %v", errDest)
				}

				// 验证目标文件已创建
				if _, errStat := os.Stat(destPath); os.IsNotExist(errStat) {
					log.Printf("[ERROR] 复制后目标文件不存在: %s", destPath)
					continue
				} else {
					log.Printf("[DEBUG] 成功创建目标文件: %s", destPath)
				}
				
				// 2. 删除原文件
				if errRemove := os.Remove(sourcePath); errRemove != nil {
					log.Printf("[WARN] 删除源文件失败，将重试: %v", errRemove)
					time.Sleep(100 * time.Millisecond)
					if errRetry := os.Remove(sourcePath); errRetry != nil {
						log.Printf("[ERROR] 第二次删除源文件失败: %v", errRetry)
					} else {
						log.Printf("[DEBUG] 第二次尝试删除源文件成功")
					}
				} else {
					log.Printf("[DEBUG] 成功删除源文件: %s", sourcePath)
				}
				
				log.Printf("[INFO] 成功移动图片: %s -> %s", sourcePath, destPath)
				
				// 生成新路径并保存
				newPath := fmt.Sprintf("/assets/imgs/%d/%s", resourceID, filename)
				log.Printf("[DEBUG] 生成新路径: %s", newPath)
				newImagePaths = append(newImagePaths, newPath)
			}
			
			resource.Images = newImagePaths
			log.Printf("[INFO] 变为 %v", resource.Images)
			
			// 异步调用WebP转换工具处理批准的图片
			go func(paths []string) {
				log.Printf("[INFO] 开始异步转换批准的图片为WebP格式，图片数量: %d", len(paths))
				convertImagesToWebP(paths)
			}(newImagePaths)
		}

		// 处理海报图片
		if approval.PosterImage != "" {
			log.Printf("[DEBUG] 开始移动海报图片，资源ID: %d, 原路径: %s", resource.ID, approval.PosterImage)	
			
			// 检查是否为TMDB外部图片链接
			if strings.HasPrefix(approval.PosterImage, "@https://image.tmdb.org/") || 
			   strings.HasPrefix(approval.PosterImage, "https://image.tmdb.org/") {
				// 对于TMDB外部图片链接，直接保存原始链接，无需移动
				log.Printf("[DEBUG] 检测到TMDB外部海报图片链接: %s，直接保存原始链接", approval.PosterImage)
				posterPath := approval.PosterImage
				resource.PosterImage = &posterPath
			} else {
				// 提取文件名
				filename := filepath.Base(approval.PosterImage)
				log.Printf("[DEBUG] 海报文件名: %s", filename)
				// 生成新路径
				newPosterPath := fmt.Sprintf("/assets/imgs/%d/%s", resourceID, filename)
				log.Printf("[INFO] 更新资源海报图片路径，从 %v 变为 %s", resource.PosterImage, newPosterPath)
				resource.PosterImage = &newPosterPath
			}
			
			// 异步调用WebP转换工具处理海报图片
			if resource.PosterImage != nil {
				posterPaths := []string{*resource.PosterImage}
				go func(paths []string) {
					log.Printf("[INFO] 开始异步转换海报图片为WebP格式")
					convertImagesToWebP(paths)
				}(posterPaths)
			}
		}
	}

	// 创建审批记录
	approvalRecord := models.ApprovalRecord{
		ResourceID:      resourceID,
		Status:          resource.Status,
		FieldApprovals:  models.JsonMap{},
		FieldRejections: models.JsonMap{},
		ApprovedImages:  approval.ApprovedImages,
		RejectedImages:  approval.RejectedImages,
		PosterImage:     approval.PosterImage,
		Notes:           approval.Notes,
		ApprovedLinks:   models.JsonMap{},
		RejectedLinks:   models.JsonMap{},
		CreatedAt:       time.Now(),
	}
	// 处理批准的链接，将它们追加到原始资源的Links字段中
	if len(approval.ApprovedLinks) > 0 {
		log.Printf("[DEBUG] 处理批准的链接，资源ID: %d, 链接数量: %d", resourceID, len(approval.ApprovedLinks))
		
		// 如果原始资源的Links字段为空，则初始化
		if resource.Links == nil {
			resource.Links = models.JsonMap{}
		}
		
		// 先按category分组链接
		linksByCategory := make(map[string][]map[string]interface{})
		// 遍历批准的链接，按category分组
		for _, link := range approval.ApprovedLinks {
			// 使用category作为键，将链接添加到对应分组
			if category, ok := link["category"].(string); ok && category != "" {
				// 创建不包含category字段的新map
				linkData := make(map[string]interface{})
				for k, v := range link {
					if k != "category" {
						linkData[k] = v
					}
				}
				
				linksByCategory[category] = append(linksByCategory[category], linkData)
			} else {
				// 如果没有有效的category，使用"unknown"作为键
				linksByCategory["other"] = append(linksByCategory["other"], link)
			}
		}

		log.Printf("[DEBUG] 分组后的链接: %v", linksByCategory)
		// 赋值给 approvalRecord.ApprovedLinks
		jsonMap := make(map[string]interface{})
		for k, v := range linksByCategory {
			jsonMap[k] = v // []map[string]interface{} 可作为 interface{}
		}
		approvalRecord.ApprovedLinks = models.JsonMap(jsonMap)
	}


	// 更新记录
	log.Printf("[DEBUG] 开始更新数据库记录，资源ID: %d", resourceID)
	log.Printf("[DEBUG] 资源状态: %s", resource.Status)
	log.Printf("[DEBUG] 资源图片: %v", resource.Images)
	log.Printf("[DEBUG] 资源海报图片: %v", resource.PosterImage)
	
	// approval_records插入审批记录
	result, errInsert := models.DB.Exec(
		`INSERT INTO approval_records (
			resource_id, status, field_approvals, field_rejections,
			approved_images, rejected_images, poster_image, notes,
			approved_links, rejected_links, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		approvalRecord.ResourceID, approvalRecord.Status,
		approvalRecord.FieldApprovals, approvalRecord.FieldRejections,
		models.JsonList(newImagePaths), approvalRecord.RejectedImages,
		approvalRecord.PosterImage, approvalRecord.Notes,
		approvalRecord.ApprovedLinks, approvalRecord.RejectedLinks,
		approvalRecord.CreatedAt,
	)

	if errInsert != nil {
		log.Printf("创建审批记录失败: %v", errInsert)
		// 继续处理，不要因为审批记录创建失败而中断流程
	} else {
		id, _ := result.LastInsertId()
		log.Printf("已创建审批记录，ID: %d", id)
	}

		
	
	// resources 更新资源
	var errUpdate error
	_, errUpdate = models.DB.Exec(
		`UPDATE resources SET 
			status = ?, images = ?, poster_image = ?, 
			approval_history = ?, updated_at = ?
		WHERE id = ?`,
		resource.Status, resource.Images, resource.PosterImage,
		resource.ApprovalHistory, resource.UpdatedAt, resource.ID,
	)

	if errUpdate != nil {
		log.Printf("[ERROR] 更新资源失败: %v", errUpdate)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("更新资源失败: %v", errUpdate)})
		return
	}
	
	log.Printf("[INFO] 成功更新资源，ID: %d", resourceID)

	// 再次从数据库获取资源，确保返回最新数据
	errGet = models.DB.Get(&resource, `SELECT * FROM resources WHERE id = ?`, resourceID)
	if errGet != nil {
		log.Printf("警告：获取更新后的资源失败，但资源已更新: %v", errGet)
	}

	c.JSON(http.StatusOK, resource)
}

// approveResourceSupplement 处理资源补充内容的审批
func approveResourceSupplement(c *gin.Context, resourceID int, resource models.Resource, approval models.ResourceApproval) {
	log.Printf("处理资源补充内容审批，资源ID: %d", resourceID)

	// 检查补充内容是否存在且状态为待审批
	if resource.Supplement == nil {
		log.Printf("资源 %d 没有补充内容", resourceID)
		c.JSON(http.StatusBadRequest, gin.H{"error": "资源没有补充内容"})
		return
	}

	status, ok := resource.Supplement["status"]
	if !ok {
		log.Printf("资源 %d 的补充内容没有status字段", resourceID)
		c.JSON(http.StatusBadRequest, gin.H{"error": "补充内容缺少状态信息"})
		return
	}

	statusStr, ok := status.(string)
	if !ok {
		log.Printf("资源 %d 的补充内容status字段不是字符串", resourceID)
		c.JSON(http.StatusBadRequest, gin.H{"error": "补充内容状态格式错误"})
			return
		}
		
	if statusStr != string(models.ResourceStatusPending) {
		log.Printf("资源 %d 的补充内容状态不是待审批: %s", resourceID, statusStr)
		c.JSON(http.StatusBadRequest, gin.H{"error": "补充内容不是待审批状态"})
			return
		}
		
	// 创建补充内容审批记录
	approvalRecord := models.ApprovalRecord{
		ResourceID:           resourceID,
		Status:               models.ResourceStatus(strings.ToUpper(string(approval.Status))),
		FieldApprovals:       models.JsonMap{},
		FieldRejections:      models.JsonMap{},
		ApprovedImages:       approval.ApprovedImages,
		RejectedImages:       approval.RejectedImages,
		PosterImage:          approval.PosterImage,
		Notes:                approval.Notes,
		ApprovedLinks:        models.JsonMap{},
		RejectedLinks:        models.JsonMap{},
		IsSupplementApproval: true,
		CreatedAt:            time.Now(),
	}

	// 转换字段审批信息
	if approval.FieldApprovals != nil {
		for k, v := range approval.FieldApprovals {
			approvalRecord.FieldApprovals[k] = v
		}
	}

	if approval.FieldRejections != nil {
		for k, v := range approval.FieldRejections {
			approvalRecord.FieldRejections[k] = v
		}
	}

	// 处理批准的链接，将它们追加到原始资源的Links字段中
	if len(approval.ApprovedLinks) > 0 {
		log.Printf("[DEBUG] 处理批准的链接，资源ID: %d, 链接数量: %d", resourceID, len(approval.ApprovedLinks))
		
		// 如果原始资源的Links字段为空，则初始化
		if resource.Links == nil {
			resource.Links = models.JsonMap{}
		}
		
		// 先按category分组链接
		linksByCategory := make(map[string][]map[string]interface{})
		// 遍历批准的链接，按category分组
		for _, link := range approval.ApprovedLinks {
			// 使用category作为键，将链接添加到对应分组
			if category, ok := link["category"].(string); ok && category != "" {
				// 创建不包含category字段的新map
				linkData := make(map[string]interface{})
				for k, v := range link {
					if k != "category" {
						linkData[k] = v
					}
				}
				
				linksByCategory[category] = append(linksByCategory[category], linkData)
			} else {
				// 如果没有有效的category，使用"other"作为键
				linksByCategory["other"] = append(linksByCategory["other"], link)
			}
		}

		log.Printf("[DEBUG] 分组后的链接: %v", linksByCategory)
		// 赋值给 approvalRecord.ApprovedLinks
		jsonMap := make(map[string]interface{})
		for k, v := range linksByCategory {
			jsonMap[k] = v // []map[string]interface{} 可作为 interface{}
		}
		approvalRecord.ApprovedLinks = models.JsonMap(jsonMap)
	}

	if len(approval.RejectedLinks) > 0 {
		linksMap := make(map[string]interface{})
		for i, link := range approval.RejectedLinks {
			linksMap[fmt.Sprintf("link_%d", i)] = link
		}
		approvalRecord.RejectedLinks = linksMap
	}

	// 保存所有新路径
	newImagePaths := make([]string, 0, len(approval.ApprovedImages))
	
	// 如果资源被批准，处理图片移动
	if strings.ToLower(string(approval.Status)) == strings.ToLower(string(models.ResourceStatusApproved)) {
		// 移动已批准的图片
		if len(approval.ApprovedImages) > 0 {
			log.Printf("[DEBUG] 开始移动已批准的补充图片，资源ID: %d, 图片数量: %d", resource.ID, len(approval.ApprovedImages))
			log.Printf("[DEBUG] 原始图片路径: %v", approval.ApprovedImages)

			// 获取assets目录路径
			assetsDir := config.GetAssetsDir()
			log.Printf("[DEBUG] Assets目录路径: %s", assetsDir)
			
			// 创建目标目录
			imgsDir := filepath.Join(assetsDir, "imgs", fmt.Sprintf("%d", resourceID))
			log.Printf("[DEBUG] 创建目标目录: %s", imgsDir)
			if errMkdir := os.MkdirAll(imgsDir, 0755); errMkdir != nil {
				log.Printf("[ERROR] 创建目录失败: %v", errMkdir)
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("创建图片目录失败: %v", errMkdir)})
				return
			}
			
			// 手动执行每个图片的移动操作
			for _, imgPath := range approval.ApprovedImages {
				if imgPath == "" {
					continue
				}
				
				// 检查是否为TMDB外部图片链接
				if strings.HasPrefix(imgPath, "@https://image.tmdb.org/") {
					// 对于TMDB外部图片链接，直接保存原始链接，无需移动
					log.Printf("[DEBUG] 检测到TMDB外部图片链接: %s，直接保存原始链接", imgPath)
					newImagePaths = append(newImagePaths, imgPath)
					continue
				}
				
				// 检查是否为TMDB外部图片链接（无@前缀）
				if strings.HasPrefix(imgPath, "https://image.tmdb.org/") {
					// 对于TMDB外部图片链接，直接保存原始链接，无需移动
					log.Printf("[DEBUG] 检测到TMDB外部图片链接: %s，直接保存原始链接", imgPath)
					newImagePaths = append(newImagePaths, imgPath)
					continue
				}
				
				// 提取文件名
				filename := filepath.Base(imgPath)
				log.Printf("[DEBUG] 处理图片: %s, 文件名: %s", imgPath, filename)
				
				// 源文件路径
				sourcePath := filepath.Join(assetsDir, imgPath[7:]) // 去掉前面的"/assets"
				
				// 目标文件路径
				destPath := filepath.Join(imgsDir, filename)
				
				log.Printf("[DEBUG] 移动图片: %s -> %s", sourcePath, destPath)
				
				// 检查源文件是否存在
				if _, errStat := os.Stat(sourcePath); os.IsNotExist(errStat) {
					log.Printf("[ERROR] 源文件不存在: %s", sourcePath)
					continue
				} else {
					log.Printf("[DEBUG] 源文件存在: %s", sourcePath)
				}
				
				// 确保目标目录存在
				errDir := os.MkdirAll(filepath.Dir(destPath), 0755)
				if errDir != nil {
					log.Printf("[ERROR] 创建目标目录失败: %v", errDir)
					continue
				}
				
				// 移动文件（复制后删除）
				// 1. 复制文件
				sourceFile, errOpen := os.Open(sourcePath)
				if errOpen != nil {
					log.Printf("[ERROR] 打开源文件失败: %v", errOpen)
					continue
				}
				defer sourceFile.Close()

				// 创建目标文件
				destFile, errCreate := os.Create(destPath)
				if errCreate != nil {
					log.Printf("[ERROR] 创建目标文件失败: %v", errCreate)
					continue
				}
				defer destFile.Close()

				// 复制内容
				_, errCopy := io.Copy(destFile, sourceFile)
				if errCopy != nil {
					log.Printf("[ERROR] 复制文件内容失败: %v", errCopy)
					continue
				}

				// 关闭文件以确保所有内容都已写入
				errSource := sourceFile.Close()
				if errSource != nil {
					log.Printf("[ERROR] 关闭源文件失败: %v", errSource)
				}

				errDest := destFile.Close()
				if errDest != nil {
					log.Printf("[ERROR] 关闭目标文件失败: %v", errDest)
				}

				// 验证目标文件已创建
				if _, errStat := os.Stat(destPath); os.IsNotExist(errStat) {
					log.Printf("[ERROR] 复制后目标文件不存在: %s", destPath)
					continue
				} else {
					log.Printf("[DEBUG] 成功创建目标文件: %s", destPath)
				}
				
				// 2. 删除原文件
				if errRemove := os.Remove(sourcePath); errRemove != nil {
					log.Printf("[WARN] 删除源文件失败，将重试: %v", errRemove)
					time.Sleep(100 * time.Millisecond)
					if errRetry := os.Remove(sourcePath); errRetry != nil {
						log.Printf("[ERROR] 第二次删除源文件失败: %v", errRetry)
					} else {
						log.Printf("[DEBUG] 第二次尝试删除源文件成功")
					}
				} else {
					log.Printf("[DEBUG] 成功删除源文件: %s", sourcePath)
				}
				
				log.Printf("[INFO] 成功移动图片: %s -> %s", sourcePath, destPath)
				
				// 生成新路径并保存
				newPath := fmt.Sprintf("/assets/imgs/%d/%s", resourceID, filename)
				log.Printf("[DEBUG] 生成新路径: %s", newPath)
				newImagePaths = append(newImagePaths, newPath)
			}
			
			// 获取资源当前的图片
			currentImages := resource.Images
			if currentImages == nil {
				currentImages = []string{}
			}
			
			// 将批准的图片路径追加到resource.Images，而不是覆盖
			log.Printf("[INFO] 更新资源图片路径，从 %v", resource.Images)
			resource.Images = append(currentImages, newImagePaths...)
			log.Printf("[INFO] 变为 %v（追加而非覆盖）", resource.Images)
			
			// 异步调用WebP转换工具处理批准的图片
			go func(paths []string) {
				log.Printf("[INFO] 开始异步转换批准的补充图片为WebP格式，图片数量: %d", len(paths))
				convertImagesToWebP(paths)
			}(newImagePaths)
		}
		
		// 处理海报图片，如果补充内容中设置了新的海报图片
		if approval.PosterImage != "" {
			log.Printf("[DEBUG] 处理补充内容的海报图片，资源ID: %d, 原路径: %s", resource.ID, approval.PosterImage)
			
			// 检查是否为TMDB外部图片链接
			if strings.HasPrefix(approval.PosterImage, "@https://image.tmdb.org/") || 
			   strings.HasPrefix(approval.PosterImage, "https://image.tmdb.org/") {
				// 对于TMDB外部图片链接，直接保存原始链接，无需移动
				log.Printf("[DEBUG] 检测到TMDB外部海报图片链接: %s，直接保存原始链接", approval.PosterImage)
				posterPath := approval.PosterImage
				resource.PosterImage = &posterPath
			} else {
				// 提取文件名
				filename := filepath.Base(approval.PosterImage)
				log.Printf("[DEBUG] 海报文件名: %s", filename)
				
				// 更新资源的海报图片
				newPosterPath := fmt.Sprintf("/assets/imgs/%d/%s", resourceID, filename)
				log.Printf("[INFO] 更新资源海报图片，从 %v 变为 %s", resource.PosterImage, newPosterPath)
				resource.PosterImage = &newPosterPath
			}
			
			// 异步调用WebP转换工具处理海报图片
			if resource.PosterImage != nil {
				posterPaths := []string{*resource.PosterImage}
				go func(paths []string) {
					log.Printf("[INFO] 开始异步转换补充资源的海报图片为WebP格式")
					convertImagesToWebP(paths)
				}(posterPaths)
			}
		}
		
		// 处理批准的链接，将它们追加到原始资源的Links字段中
		if len(approval.ApprovedLinks) > 0 {
			log.Printf("[DEBUG] 处理批准的链接，资源ID: %d, 链接数量: %d", resourceID, len(approval.ApprovedLinks))
			
			// 如果原始资源的Links字段为空，则初始化
			if resource.Links == nil {
				resource.Links = models.JsonMap{}
			}
			
			// 先按category分组链接
			linksByCategory := make(map[string][]map[string]interface{})
			
			// 遍历批准的链接，按category分组
			for _, link := range approval.ApprovedLinks {
				// 使用category作为键，将链接添加到对应分组
				if category, ok := link["category"].(string); ok && category != "" {
					// 创建不包含category字段的新map
					linkData := make(map[string]interface{})
					for k, v := range link {
						if k != "category" {
							linkData[k] = v
						}
					}
					
					linksByCategory[category] = append(linksByCategory[category], linkData)
				} else {
					// 如果没有有效的category，使用"unknown"作为键
					linksByCategory["unknown"] = append(linksByCategory["unknown"], link)
				}
			}
			
			// 将分组后的链接添加到resource.Links中
			for category, links := range linksByCategory {
				log.Printf("[DEBUG] 添加链接组，键: %s, 数量: %d", category, len(links))
				
				// 检查是否已存在该category的链接
				if existingLinks, ok := resource.Links[category]; ok {
					// 已存在该category的链接，将新链接追加到现有数组
					if existingLinksArray, ok := existingLinks.([]interface{}); ok {
						// 已经是数组格式，追加新链接
						for _, link := range links {
							existingLinksArray = append(existingLinksArray, link)
						}
						resource.Links[category] = existingLinksArray
					} else {
						// 不是数组格式，转换为数组后追加
						newLinks := []interface{}{existingLinks}
						for _, link := range links {
							newLinks = append(newLinks, link)
						}
						resource.Links[category] = newLinks
					}
				} else {
					// 不存在该category的链接，直接添加
					interfaceLinks := make([]interface{}, len(links))
					for i, link := range links {
						interfaceLinks[i] = link
					}
					resource.Links[category] = interfaceLinks
				}
			}
			
			log.Printf("[INFO] 更新后的资源链接: %v", resource.Links)
		}
		
		// 更新数据库中的资源信息
		var errUpdate error
		_, errUpdate = models.DB.Exec(
			`UPDATE resources SET 
				images = ?, poster_image = ?, links = ?,
				updated_at = ?
			WHERE id = ?`,
			resource.Images, resource.PosterImage, resource.Links,
			time.Now(), resourceID,
		)
		
		if errUpdate != nil {
			log.Printf("[ERROR] 更新资源图片失败: %v", errUpdate)
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("更新资源图片失败: %v", errUpdate)})
			return
		}
		
		log.Printf("[INFO] 成功更新资源图片，ID: %d", resourceID)
	}

	// 更新补充内容状态
	resource.IsSupplementApproval = true
	resource.Supplement = nil // 清空补充内容
	resource.UpdatedAt = time.Now()

	// 更新资源
	var errUpdate error
	_, errUpdate = models.DB.Exec(
		`UPDATE resources SET is_supplement_approval = 'True', supplement = NULL, updated_at = ? WHERE id = ?`,
		resource.UpdatedAt, resourceID,
	)

	// 检查错误
	if errUpdate != nil {
		log.Printf("更新资源is_supplement_approval失败: %v", errUpdate)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("更新资源is_supplement_approval失败: %v", errUpdate)})
		return
	}

	log.Printf("资源ID: %d 的is_supplement_approval已成功更新为True，supplement已清空", resourceID)

	// 插入审批记录
	result, errInsert := models.DB.Exec(
		`INSERT INTO approval_records (
			resource_id, status, field_approvals, field_rejections,
			approved_images, rejected_images, poster_image, notes,
			approved_links, rejected_links, is_supplement_approval, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		approvalRecord.ResourceID, approvalRecord.Status,
		approvalRecord.FieldApprovals, approvalRecord.FieldRejections,
		models.JsonList(newImagePaths), approvalRecord.RejectedImages,
		approvalRecord.PosterImage, approvalRecord.Notes,
		approvalRecord.ApprovedLinks, approvalRecord.RejectedLinks,
		approvalRecord.IsSupplementApproval, approvalRecord.CreatedAt,
	)
	

	if errInsert != nil {
		log.Printf("创建补充内容审批记录失败: %v", errInsert)
		// 继续处理，不要因为审批记录创建失败而中断流程
	} else {
		id, _ := result.LastInsertId()
		log.Printf("已创建补充内容审批记录，ID: %d", id)
	}
	
	
	// 返回更新后的资源
	var updatedResource models.Resource
	errGet := models.DB.Get(&updatedResource, `SELECT * FROM resources WHERE id = ?`, resourceID)
	if errGet != nil {
		log.Printf("警告：获取更新后的资源失败，但资源已更新: %v", errGet)
		c.JSON(http.StatusOK, resource)
	} else {
		c.JSON(http.StatusOK, updatedResource)
	}
}

// convertImagesToWebP 将批准的图片转换成WebP格式
func convertImagesToWebP(imagePaths []string) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[ERROR] WebP转换过程中发生严重错误: %v", r)
		}
	}()

	if len(imagePaths) == 0 {
		log.Printf("[INFO] 没有需要转换为WebP的图片")
		return
	}
	
	log.Printf("[INFO] 开始将 %d 张批准的图片转换为WebP格式", len(imagePaths))
	startTime := time.Now()
	
	// 过滤掉TMDB外部图片链接，这些不需要转换为WebP
	localImagePaths := make([]string, 0, len(imagePaths))
	
	for _, path := range imagePaths {
		// 跳过TMDB外部图片链接
		if strings.HasPrefix(path, "@https://image.tmdb.org/") || 
		   strings.HasPrefix(path, "https://image.tmdb.org/") {
			log.Printf("[INFO] 跳过TMDB外部图片链接: %s，不进行WebP转换", path)
			continue
		}
		
		// 将 /assets/... 转换为 ../assets/...
		if strings.HasPrefix(path, "/assets/") {
			adjustedPath := "../" + strings.TrimPrefix(path, "/")
			localImagePaths = append(localImagePaths, adjustedPath)
		} else {
			log.Printf("[WARN] 图片路径格式不符合预期: %s，跳过处理", path)
		}
	}
	
	if len(localImagePaths) == 0 {
		log.Printf("[WARN] 没有有效的图片路径可供转换")
		return
	}
	
	// 将路径转换为JSON字符串
	pathsJSON, errJSON := json.Marshal(localImagePaths)
	if errJSON != nil {
		log.Printf("[ERROR] 无法将图片路径转为JSON: %v", errJSON)
		return
	}
	
	log.Printf("[DEBUG] 准备调用WebP转换工具，处理以下图片: %s", string(pathsJSON))
	
	// 调用WebP转换工具
	resultPaths, errConvert := utils.ConvertMultipleImages(string(pathsJSON), true, false, 4)
	if errConvert != nil {
		log.Printf("[ERROR] 转换WebP过程中发生错误: %v", errConvert)
		return
	}
	
	elapsedTime := time.Since(startTime)
	log.Printf("[INFO] 成功将 %d 张图片转换为WebP格式，耗时: %v", len(resultPaths), elapsedTime)
}

// DeleteApprovalRecord 删除审批记录 - 仅管理员可访问
func DeleteApprovalRecord(c *gin.Context) {
	// 获取路径参数
	recordID, errParse := strconv.Atoi(c.Param("id"))
	if errParse != nil {
		log.Printf("解析审批记录ID失败: %v, 参数: %s", errParse, c.Param("id"))
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的审批记录ID"})
		return
	}

	log.Printf("尝试删除审批记录，ID: %d", recordID)

	// 检查记录是否存在
	var record models.ApprovalRecord
	errGet := models.DB.Get(&record, `SELECT * FROM approval_records WHERE id = ?`, recordID)
	if errGet != nil {
		log.Printf("未找到ID为%d的审批记录: %v", recordID, errGet)
		c.JSON(http.StatusNotFound, gin.H{"error": "未找到审批记录"})
		return
	}

	log.Printf("找到ID为%d的审批记录，资源ID: %d", recordID, record.ResourceID)

	// 删除记录
	result, errDelete := models.DB.Exec(`DELETE FROM approval_records WHERE id = ?`, recordID)
	if errDelete != nil {
		log.Printf("删除ID为%d的审批记录失败: %v", recordID, errDelete)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除审批记录失败"})
		return
	}

	// 检查是否真的删除了记录
	affected, errAffected := result.RowsAffected()
	if errAffected != nil {
		log.Printf("获取影响行数失败: %v", errAffected)
	} else if affected == 0 {
		log.Printf("ID为%d的审批记录未被删除", recordID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除审批记录失败，没有记录被删除"})
		return
	}

	log.Printf("成功删除ID为%d的审批记录", recordID)
	c.Status(http.StatusNoContent)
}

// SupplementResource 为资源添加补充内容
func SupplementResource(c *gin.Context) {
	// 获取路径参数
	resourceID, errParse := strconv.Atoi(c.Param("id"))
	if errParse != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的资源ID"})
		return
	}

	// 解析请求
	var supplement models.SupplementCreate
	if errBind := c.ShouldBindJSON(&supplement); errBind != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	// 检查资源是否存在并且是已批准的
	var resource models.Resource
	errGet := models.DB.Get(&resource, `SELECT * FROM resources WHERE id = ?`, resourceID)
	if errGet != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "资源未找到"})
		return
	}

	// 检查是否已有待审批的补充内容
	if resource.Supplement != nil {
		// 尝试获取现有补充内容的状态
		if status, ok := resource.Supplement["status"]; ok {
			if statusStr, ok := status.(string); ok && statusStr == string(models.ResourceStatusPending) {
				// 有待审批的补充内容，需要合并而不是覆盖
				log.Printf("资源 %d 已有待审批的补充内容，将进行合并", resourceID)
				
				// 合并图片列表
				existingImages := []string{}
				if imgs, ok := resource.Supplement["images"]; ok {
					if imgList, ok := imgs.([]interface{}); ok {
						for _, img := range imgList {
							if imgStr, ok := img.(string); ok {
								existingImages = append(existingImages, imgStr)
							}
						}
					}
				}
				
				// 将新图片追加到现有图片列表中
				mergedImages := append(existingImages, supplement.Images...)
				
				// 处理链接 - 合并现有链接和新链接
				existingLinks := make(map[string][]interface{})
				if links, ok := resource.Supplement["links"]; ok {
					if linksMap, ok := links.(map[string]interface{}); ok {
						for category, categoryLinks := range linksMap {
							if catLinks, ok := categoryLinks.([]interface{}); ok {
								existingLinks[category] = catLinks
							}
						}
					}
				}
				
				// 将新链接合并到现有链接中
				mergedLinks := make(map[string]interface{})
				for category, links := range existingLinks {
					mergedLinks[category] = links
				}
				
				// 合并新提交的链接
				for category, categoryLinks := range supplement.Links {
					if existingCatLinks, ok := mergedLinks[category]; ok {
						// 已有该分类的链接，追加
						if existingArr, ok := existingCatLinks.([]interface{}); ok {
							// 根据categoryLinks的类型进行不同处理
							if newLinksArray, ok := categoryLinks.([]interface{}); ok {
								// 如果已经是[]interface{}类型，直接追加
								mergedLinks[category] = append(existingArr, newLinksArray...)
							} else if newLinksArray, ok := categoryLinks.([]map[string]interface{}); ok {
								// 如果是[]map[string]interface{}类型，转换后追加
								for _, link := range newLinksArray {
									existingArr = append(existingArr, link)
								}
								mergedLinks[category] = existingArr
							} else {
								// 单个链接对象，直接追加
								mergedLinks[category] = append(existingArr, categoryLinks)
							}
						}
					} else {
						// 没有该分类的链接，直接添加
						mergedLinks[category] = categoryLinks
					}
				}
				
				// 更新合并后的补充内容
				supplementData := models.JsonMap{
					"images":          mergedImages,
					"links":           mergedLinks,
					"status":          string(models.ResourceStatusPending),
					"submission_date": time.Now().Format(time.RFC3339),
				}
				
				// 更新资源，添加补充内容
				_, errUpdate := models.DB.Exec(
					`UPDATE resources SET supplement = ?, is_supplement_approval = ?, updated_at = ? WHERE id = ?`,
					supplementData, false, time.Now(), resourceID,
				)
				
				if errUpdate != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("添加补充内容失败: %v", errUpdate)})
					return
				}
				
				// 更新内存中的资源对象
				resource.Supplement = supplementData
				resource.UpdatedAt = time.Now()
				
				c.JSON(http.StatusOK, resource)
				return
			}
		}
	}

	// 没有待审批的补充内容，直接创建新的
	supplementData := models.JsonMap{
		"images":          supplement.Images,
		"links":           supplement.Links,
		"status":          string(models.ResourceStatusPending),
		"submission_date": time.Now().Format(time.RFC3339),
	}

	// 更新资源，添加补充内容
	_, errUpdate := models.DB.Exec(
		`UPDATE resources SET supplement = ?, is_supplement_approval = ?, updated_at = ? WHERE id = ?`,
		supplementData, false, time.Now(), resourceID,
	)

	if errUpdate != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("添加补充内容失败: %v", errUpdate)})
		return
	}

	// 更新内存中的资源对象
	resource.Supplement = supplementData
	resource.UpdatedAt = time.Now()

	c.JSON(http.StatusOK, resource)
}

// GetResourceSupplement 获取资源的补充内容 - 仅管理员可访问
func GetResourceSupplement(c *gin.Context) {
	// 获取路径参数
	resourceID, errParse := strconv.Atoi(c.Param("id"))
	if errParse != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的资源ID"})
		return
	}

	// 查询资源
	var resource models.Resource
	errGet := models.DB.Get(&resource, `SELECT * FROM resources WHERE id = ?`, resourceID)
	if errGet != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "资源未找到"})
		return
	}

	if resource.Supplement == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "资源没有补充内容"})
		return
	}

	c.JSON(http.StatusOK, resource.Supplement)
}

// GetPendingSupplementResources 获取待审批补充内容的资源列表 - 仅管理员可访问
func GetPendingSupplementResources(c *gin.Context) {
	// 解析查询参数
	skip, _ := strconv.Atoi(c.DefaultQuery("skip", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))

	// 查询所有包含补充内容的资源
	var resources []models.Resource
	errSelect := models.DB.Select(&resources, 
		`SELECT * FROM resources WHERE supplement IS NOT NULL LIMIT ? OFFSET ?`,
		limit, skip)
	if errSelect != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询待审批补充内容资源失败"})
		return
	}
	
	// 确保resources不为null
	if resources == nil {
		resources = []models.Resource{}
	}

	// 筛选待审批的补充内容资源
	pendingSupplements := []models.Resource{}
	for _, resource := range resources {
		if resource.Supplement == nil {
			continue
		}

		status, ok := resource.Supplement["status"]
		if !ok {
			continue
		}

		if statusStr, ok := status.(string); ok && statusStr == string(models.ResourceStatusPending) {
			resource.HasPendingSupplement = true
			pendingSupplements = append(pendingSupplements, resource)
		}
	}

	// 即使没有待审批补充内容也返回空数组
	c.JSON(http.StatusOK, pendingSupplements)
}

// DeleteApprovalRecords 批量删除审批记录 - 仅管理员可访问
func DeleteApprovalRecords(c *gin.Context) {
	// 解析请求体中的审批记录ID列表
	var request struct {
		IDs []int `json:"ids" binding:"required"`
	}

	if errBind := c.ShouldBindJSON(&request); errBind != nil {
		log.Printf("解析请求体失败: %v", errBind)
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	if len(request.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID列表为空"})
		return
	}

	log.Printf("批量删除审批记录，ID数量: %d, IDs: %v", len(request.IDs), request.IDs)

	// 批量删除记录
	var successCount int
	var failedIDs []int

	for _, id := range request.IDs {
		// 检查记录是否存在
		var count int
		errCount := models.DB.Get(&count, `SELECT COUNT(*) FROM approval_records WHERE id = ?`, id)
		if errCount != nil || count == 0 {
			log.Printf("未找到ID为%d的审批记录", id)
			failedIDs = append(failedIDs, id)
			continue
		}

		// 删除记录
		result, errDelete := models.DB.Exec(`DELETE FROM approval_records WHERE id = ?`, id)
		if errDelete != nil {
			log.Printf("删除ID为%d的审批记录失败: %v", id, errDelete)
			failedIDs = append(failedIDs, id)
			continue
		}

		affected, errAffected := result.RowsAffected()
		if errAffected != nil || affected == 0 {
			log.Printf("ID为%d的审批记录未被删除", id)
			failedIDs = append(failedIDs, id)
			continue
		}

		successCount++
	}

	log.Printf("批量删除完成，成功: %d, 失败: %d", successCount, len(failedIDs))
	c.JSON(http.StatusOK, gin.H{
		"success_count": successCount,
		"failed_ids":    failedIDs,
	})
}

// GetApprovalRecords 获取所有审批记录 - 仅管理员可访问
func GetApprovalRecords(c *gin.Context) {
	// 解析查询参数
	skip, _ := strconv.Atoi(c.DefaultQuery("skip", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))

	log.Printf("获取审批记录: skip=%d, limit=%d", skip, limit)

	// 获取审批记录总数
	var count int
	errCount := models.DB.Get(&count, "SELECT COUNT(*) FROM approval_records")
	if errCount != nil {
		log.Printf("获取审批记录总数失败: %v", errCount)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取审批记录总数失败"})
		return
	}

	// 如果没有记录，返回空数组
	if count == 0 {
		log.Printf("没有审批记录")
		c.JSON(http.StatusOK, gin.H{
			"records": []interface{}{},
			"total":   0,
		})
		return
	}

	// 查询审批记录
	query := `
		SELECT ar.*, r.title, r.title_en, r.resource_type, r.status as resource_status
		FROM approval_records ar
		LEFT JOIN resources r ON ar.resource_id = r.id
		ORDER BY ar.created_at DESC
		LIMIT ? OFFSET ?
	`
	
	rows, errQuery := models.DB.Queryx(query, limit, skip)
	if errQuery != nil {
		log.Printf("查询审批记录失败: %v", errQuery)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询审批记录失败"})
		return
	}
	defer rows.Close()

	// 处理结果
	type ApprovalRecordResponse struct {
		models.ApprovalRecord
		Title          string             `db:"title" json:"title"`
		TitleEn        string             `db:"title_en" json:"title_en"`
		ResourceType   string             `db:"resource_type" json:"resource_type"`
		ResourceStatus models.ResourceStatus `db:"resource_status" json:"resource_status"`
	}

	records := []ApprovalRecordResponse{}
	for rows.Next() {
		var record ApprovalRecordResponse
		if errScan := rows.StructScan(&record); errScan != nil {
			log.Printf("扫描审批记录失败: %v", errScan)
			continue
		}
		records = append(records, record)
	}

	if errRows := rows.Err(); errRows != nil {
		log.Printf("遍历审批记录结果集失败: %v", errRows)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "处理审批记录失败"})
		return
	}

	log.Printf("成功获取 %d 条审批记录", len(records))
	c.JSON(http.StatusOK, gin.H{
		"records": records,
		"total":   count,
	})
}

// GetResourceApprovalRecords 获取单个资源的审批记录 - 仅管理员可访问
func GetResourceApprovalRecords(c *gin.Context) {
	// 获取资源ID
	resourceID, errParse := strconv.Atoi(c.Param("id"))
	if errParse != nil {
		log.Printf("无效的资源ID: %v", errParse)
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的资源ID"})
		return
	}

	// 检查资源是否存在
	var resource models.Resource
	errGet := models.DB.Get(&resource, "SELECT * FROM resources WHERE id = ?", resourceID)
	if errGet != nil {
		log.Printf("资源未找到: %v", errGet)
		c.JSON(http.StatusNotFound, gin.H{"error": "资源未找到"})
		return
	}

	// 查询该资源的审批记录
	var records []models.ApprovalRecord
	errSelect := models.DB.Select(&records, "SELECT * FROM approval_records WHERE resource_id = ? ORDER BY created_at DESC", resourceID)
	if errSelect != nil {
		log.Printf("查询资源审批记录失败: %v", errSelect)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询资源审批记录失败"})
		return
	}

	log.Printf("成功获取资源ID=%d的%d条审批记录", resourceID, len(records))
	c.JSON(http.StatusOK, gin.H{
		"resource": resource,
		"records":  records,
	})
}

// BatchRejectResources 批量拒绝资源 - 仅管理员可访问
func BatchRejectResources(c *gin.Context) {
	// 解析请求体中的资源ID列表
	var request struct {
		ResourceIDs []int  `json:"resource_ids" binding:"required"`
		Notes       string `json:"notes"`
	}

	if errBind := c.ShouldBindJSON(&request); errBind != nil {
		log.Printf("解析请求体失败: %v", errBind)
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	if len(request.ResourceIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "资源ID列表为空"})
		return
	}

	log.Printf("批量拒绝资源，ID数量: %d, IDs: %v", len(request.ResourceIDs), request.ResourceIDs)

	// 批量处理资源
	var successCount int
	var failedIDs []int

	for _, resourceID := range request.ResourceIDs {
		// 检查资源是否存在且为待审批状态
		var resource models.Resource
		errGet := models.DB.Get(&resource, `SELECT * FROM resources WHERE id = ?`, resourceID)
		if errGet != nil {
			log.Printf("资源 %d 未找到: %v", resourceID, errGet)
			failedIDs = append(failedIDs, resourceID)
			continue
		}

		// 检查资源状态
		if resource.Status != models.ResourceStatusPending {
			log.Printf("资源 %d 不是待审批状态: %s", resourceID, resource.Status)
			failedIDs = append(failedIDs, resourceID)
			continue
		}

		// 更新资源状态为拒绝
		resource.Status = models.ResourceStatusRejected
		resource.UpdatedAt = time.Now()

		// 创建审批记录
		approvalRecord := models.ApprovalRecord{
			ResourceID:      resourceID,
			Status:          models.ResourceStatusRejected,
			FieldApprovals:  models.JsonMap{},
			FieldRejections: models.JsonMap{},
			ApprovedImages:  []string{},
			RejectedImages:  []string{},
			Notes:           request.Notes,
			ApprovedLinks:   models.JsonMap{},
			RejectedLinks:   models.JsonMap{},
			CreatedAt:       time.Now(),
		}

		// 开始事务
		tx := models.DB.MustBegin()

		// 更新资源状态
		_, errUpdate := tx.Exec(
			`UPDATE resources SET status = ?, updated_at = ? WHERE id = ?`,
			resource.Status, resource.UpdatedAt, resourceID,
		)

		if errUpdate != nil {
			log.Printf("更新资源 %d 状态失败: %v", resourceID, errUpdate)
			tx.Rollback()
			failedIDs = append(failedIDs, resourceID)
			continue
		}

		// 插入审批记录
		_, errInsert := tx.Exec(
			`INSERT INTO approval_records (
				resource_id, status, field_approvals, field_rejections,
				approved_images, rejected_images, poster_image, notes,
				approved_links, rejected_links, created_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			approvalRecord.ResourceID, approvalRecord.Status,
			approvalRecord.FieldApprovals, approvalRecord.FieldRejections,
			models.JsonList(approvalRecord.ApprovedImages), models.JsonList(approvalRecord.RejectedImages),
			"", approvalRecord.Notes,
			approvalRecord.ApprovedLinks, approvalRecord.RejectedLinks,
			approvalRecord.CreatedAt,
		)

		if errInsert != nil {
			log.Printf("创建资源 %d 的审批记录失败: %v", resourceID, errInsert)
			tx.Rollback()
			failedIDs = append(failedIDs, resourceID)
			continue
		}

		// 提交事务
		if errCommit := tx.Commit(); errCommit != nil {
			log.Printf("提交资源 %d 的事务失败: %v", resourceID, errCommit)
			failedIDs = append(failedIDs, resourceID)
			continue
		}

		successCount++
		log.Printf("成功拒绝资源 %d", resourceID)
	}

	log.Printf("批量拒绝完成，成功: %d, 失败: %d", successCount, len(failedIDs))
	c.JSON(http.StatusOK, gin.H{
		"success_count": successCount,
		"failed_ids":    failedIDs,
		"total":         len(request.ResourceIDs),
	})
} 
