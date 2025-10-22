<template>
  <div class="admin-container">
    <div class="admin-hero">
      <h1 class="hero-title">管理控制台</h1>
      <p class="hero-subtitle">管理资源、审批内容、维护系统</p>
    </div>
    
    <div v-if="loading" class="loading-container">
      <div class="loader"></div>
      <p>正在加载数据，请稍候...</p>
    </div>
    
    <div v-else-if="error" class="error-message">
      <i class="bi bi-exclamation-triangle-fill"></i>
      {{ error }}
    </div>
    
    <div v-else class="admin-content">
      <!-- 用户管理卡片 -->
      <div class="admin-card" v-if="isSuperAdmin">
        <div class="card-header">
          <h4><i class="bi bi-people-fill"></i> 用户管理</h4>
          <div class="header-actions">
            <button 
              type="button" 
              class="btn-custom btn-primary btn-sm" 
              @click="showAddUserDialog"
              v-if="showUserManagement"
            >
              <i class="bi bi-person-plus"></i> 
              <span class="btn-text">添加用户</span>
            </button>
            
            <button 
              type="button" 
              class="btn-custom btn-outline toggle-btn" 
              @click="showUserManagement = !showUserManagement"
            >
              <i :class="showUserManagement ? 'bi bi-chevron-up' : 'bi bi-chevron-down'"></i>
              <span class="btn-text">{{ showUserManagement ? '收起' : '展开' }}</span>
            </button>
          </div>
        </div>
        <div class="card-body" v-if="showUserManagement">
          <div v-if="loadingUsers" class="loading-inline">
            <div class="spinner small-spinner"></div>
            <span>加载用户列表...</span>
          </div>
          <div v-else-if="users.length === 0" class="empty-state">
            <i class="bi bi-people"></i>
            <p>暂无用户数据</p>
          </div>
          <div v-else class="table-container">
            <table class="custom-table">
              <thead>
                <tr>
                  <th>ID</th>
                  <th>用户名</th>
                  <th>管理员</th>
                  <th>创建时间</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="user in users" :key="user.id">
                  <td><span class="id-badge">#{{ user.id }}</span></td>
                  <td>{{ user.username }}</td>
                  <td>
                    <span 
                      class="status-badge" 
                      :class="{ 'status-approved': user.is_admin, 'status-rejected': !user.is_admin }"
                    >
                      {{ user.is_admin ? '是' : '否' }}
                    </span>
                  </td>
                  <td>{{ formatDate(user.created_at) }}</td>
                  <td class="actions-cell">
                    <button class="btn-custom btn-outline btn-sm" @click="showEditUserDialog(user)">
                      <i class="bi bi-pencil"></i> 
                      <span class="btn-text">编辑</span>
                    </button>
                    <button 
                      class="btn-custom btn-accent btn-sm" 
                      @click="confirmDeleteUser(user)"
                      :disabled="user.id === currentUser?.id"
                    >
                      <i class="bi bi-trash"></i> 
                      <span class="btn-text">删除</span>
                    </button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
      
      <!-- 添加/编辑用户对话框 -->
      <div v-if="userDialogVisible" class="custom-modal" @click.self="userDialogVisible = false">
        <div class="modal-dialog">
          <div class="modal-content">
            <div class="modal-header">
              <h5 class="modal-title">
                <i class="bi bi-person"></i> {{ isEditing ? '编辑用户' : '添加用户' }}
              </h5>
              <button type="button" class="close-btn" @click="userDialogVisible = false">
                <i class="bi bi-x-lg"></i>
              </button>
            </div>
            <div class="modal-body">
              <form @submit.prevent="submitUserForm">
                <div class="form-group">
                  <label class="form-label">用户名</label>
                  <div class="input-group">
                    <div class="input-prefix">
                      <i class="bi bi-person-fill"></i>
                    </div>
                    <input 
                      type="text" 
                      class="custom-input" 
                      v-model="userForm.username" 
                      placeholder=""
                      required
                      minlength="3"
                      maxlength="20"
                    >
                  </div>
                  <div class="form-text">用户名长度应为3-20个字符</div>
                </div>
                
                <div class="form-group" v-if="!isEditing || changePassword">
                  <label class="form-label">密码</label>
                  <div class="input-group">
                    <div class="input-prefix">
                      <i class="bi bi-lock-fill"></i>
                    </div>
                    <input 
                      type="password" 
                      class="custom-input" 
                      v-model="userForm.password" 
                      placeholder=""
                      required
                      minlength="6"
                    >
                  </div>
                  <div class="form-text">密码长度不能少于6个字符</div>
                </div>
                
                <div class="form-group" v-if="isEditing">
                  <label class="form-label">修改密码</label>
                  <div class="switch-toggle-wrapper">
                    <div class="toggle-switch">
                      <input 
                        type="checkbox" 
                        id="change-password" 
                        v-model="userChangePassword"
                      >
                      <label for="change-password" class="switch-label"></label>
                    </div>
                    <span class="switch-text">{{ userChangePassword ? '是' : '否' }}</span>
                  </div>
                </div>
                
                <div class="form-group">
                  <label class="form-label">管理员权限</label>
                  <div class="checkbox-wrapper horizontal-display">
                    <input 
                      type="checkbox" 
                      id="is-admin" 
                      class="custom-checkbox"
                      v-model="userForm.is_admin"
                    >
                    <label for="is-admin"></label>
                    <span class="checkbox-text">{{ userForm.is_admin ? '是' : '否' }}</span>
                  </div>
                </div>
              </form>
            </div>
            <div class="modal-footer">
              <button type="button" class="btn-custom btn-outline" @click="userDialogVisible = false">取消</button>
              <button 
                type="button" 
                class="btn-custom btn-primary" 
                @click="submitUserForm" 
                :disabled="submitting"
              >
                <div v-if="submitting" class="spinner small-spinner"></div>
                <span>{{ submitting ? '提交中...' : '确认' }}</span>
              </button>
            </div>
          </div>
        </div>
      </div>
      
      <!-- 删除用户确认对话框 -->
      <div v-if="deleteUserDialogVisible" class="custom-modal" @click.self="cancelDeleteUser">
        <div class="modal-dialog small-dialog">
          <div class="modal-content">
            <div class="modal-header">
              <h5 class="modal-title"><i class="bi bi-trash"></i> 确认删除用户</h5>
              <button type="button" class="close-btn" @click="cancelDeleteUser">
                <i class="bi bi-x-lg"></i>
              </button>
            </div>
            <div class="modal-body">
              <p class="confirm-message">确定要删除用户 <strong>{{ userToDelete?.username }}</strong> 吗？此操作不可撤销。</p>
            </div>
            <div class="modal-footer">
              <button type="button" class="btn-custom btn-outline" @click="cancelDeleteUser">取消</button>
              <button 
                type="button" 
                class="btn-custom btn-accent" 
                @click="deleteUser" 
                :disabled="deleting"
              >
                <div v-if="deleting" class="spinner small-spinner"></div>
                <span>{{ deleting ? '删除中...' : '确认删除' }}</span>
              </button>
            </div>
          </div>
        </div>
      </div>
      
      <!-- 修改密码卡片 -->
      <div class="admin-card">
        <div class="card-header">
          <h4><i class="bi bi-shield-lock"></i> 修改密码</h4>
          <div class="header-actions">
            <button 
              type="button" 
              class="btn-custom btn-outline toggle-btn" 
              @click="showChangePassword = !showChangePassword"
            >
              <i :class="showChangePassword ? 'bi bi-chevron-up' : 'bi bi-chevron-down'"></i>
              <span class="btn-text">{{ showChangePassword ? '收起' : '展开' }}</span>
            </button>
          </div>
        </div>
        <div class="card-body" v-if="showChangePassword">
          <div v-if="passwordSuccess" class="success-message">
            <i class="bi bi-check-circle-fill"></i>
            密码修改成功
          </div>
          <div v-if="passwordError" class="error-message">
            <i class="bi bi-exclamation-triangle-fill"></i>
            {{ passwordError }}
          </div>
          
          <form @submit.prevent="changePassword">
            <div class="form-group">
              <label for="currentPassword" class="form-label">当前密码</label>
              <div class="input-group">
                <div class="input-prefix">
                  <i class="bi bi-key-fill"></i>
                </div>
              <input 
                type="password" 
                  class="custom-input" 
                id="currentPassword" 
                v-model="passwordForm.currentPassword" 
                required
                  placeholder="请输入当前密码"
              >
              </div>
            </div>
            
            <div class="form-group">
              <label for="newPassword" class="form-label">新密码</label>
              <div class="input-group">
                <div class="input-prefix">
                  <i class="bi bi-lock-fill"></i>
                </div>
              <input 
                type="password" 
                  class="custom-input" 
                id="newPassword" 
                v-model="passwordForm.newPassword" 
                required
                  placeholder="请输入新密码"
              >
              </div>
            </div>
            
            <div class="form-group">
              <label for="confirmPassword" class="form-label">确认新密码</label>
              <div class="input-group">
                <div class="input-prefix">
                  <i class="bi bi-lock-fill"></i>
                </div>
              <input 
                type="password" 
                  class="custom-input" 
                id="confirmPassword" 
                v-model="passwordForm.confirmPassword" 
                required
                  placeholder="请再次输入新密码"
              >
              </div>
            </div>
            
            <div class="form-actions">
              <button 
                type="submit" 
                class="btn-custom btn-primary" 
                :disabled="passwordLoading"
              >
                <div v-if="passwordLoading" class="spinner"></div>
                <i class="bi bi-key"></i>
                <span class="btn-text">{{ passwordLoading ? '提交中...' : '修改密码' }}</span>
              </button>
            </div>
          </form>
        </div>
      </div>
      
      <!-- 网站设置卡片 -->
      <div class="admin-card" v-if="isSuperAdmin">
        <div class="card-header">
          <h4><i class="bi bi-gear-fill"></i> 网站设置</h4>
          <div class="header-actions">
            <button 
              type="button" 
              class="btn-custom btn-outline toggle-btn" 
              @click="showSiteSettings = !showSiteSettings"
            >
              <i :class="showSiteSettings ? 'bi bi-chevron-up' : 'bi bi-chevron-down'"></i>
              <span class="btn-text">{{ showSiteSettings ? '收起' : '展开' }}</span>
            </button>
          </div>
        </div>
        <div class="card-body" v-if="showSiteSettings">
          <div v-if="settingsSuccess" class="success-message">
            <i class="bi bi-check-circle-fill"></i>
            网站设置更新成功
          </div>
          <div v-if="settingsError" class="error-message">
            <i class="bi bi-exclamation-triangle-fill"></i>
            {{ settingsError }}
          </div>
          
          <!-- 标签页导航 -->
          <div class="settings-tabs">
            <div 
              class="settings-tab" 
              :class="{ 'active': activeSettingsTab === 'basic' }"
              @click="activeSettingsTab = 'basic'"
            >
              <i class="bi bi-info-circle"></i>
              <span>基本信息</span>
            </div>
            <div 
              class="settings-tab" 
              :class="{ 'active': activeSettingsTab === 'meta' }"
              @click="activeSettingsTab = 'meta'"
            >
              <i class="bi bi-card-text"></i>
              <span>页面元信息</span>
            </div>
            <div 
              class="settings-tab" 
              :class="{ 'active': activeSettingsTab === 'footer' }"
              @click="activeSettingsTab = 'footer'"
            >
              <i class="bi bi-layout-text-window"></i>
              <span>页脚设置</span>
            </div>
            <div 
              class="settings-tab" 
              :class="{ 'active': activeSettingsTab === 'datasources' }"
              @click="activeSettingsTab = 'datasources'"
            >
              <i class="bi bi-database-gear"></i>
              <span>采集解析源</span>
            </div>
            <div 
              class="settings-tab" 
              :class="{ 'active': activeSettingsTab === 'tmdb' }"
              @click="activeSettingsTab = 'tmdb'"
            >
              <i class="bi bi-film"></i>
              <span>TMDB配置</span>
            </div>
            <div 
              class="settings-tab" 
              :class="{ 'active': activeSettingsTab === 'about' }"
              @click="activeSettingsTab = 'about'"
            >
              <i class="bi bi-file-person"></i>
              <span>关于页面</span>
            </div>
          </div>
          
          <!-- 标签页内容区域 -->
          <div class="settings-tab-content">
            <!-- 基本信息设置标签页 -->
            <div class="settings-section" v-show="activeSettingsTab === 'basic'">
            <h5 class="section-title">基本信息设置</h5>
            
            <!-- 网站图标 (favicon) -->
            <div class="form-group">
              <label class="form-label">网站图标 (favicon)</label>
              <div class="favicon-uploader">
                <!-- 文件输入框（隐藏） -->
                <input 
                  type="file" 
                  id="faviconUpload" 
                  ref="faviconUploadRef" 
                  class="hidden-upload" 
                  accept="image/x-icon,image/png,image/jpeg,image/svg+xml"
                  @change="handleFaviconUpload"
                />
                
                <!-- 上传区域 -->
                <div 
                  class="favicon-upload-area"
                  :class="{'has-preview': siteFaviconPreview || (footerSettings.favicon && !siteFaviconFile)}"
                  @click="triggerFileInput"
                >
                  <!-- 预览图片 -->
                  <img 
                    v-if="siteFaviconPreview || (footerSettings.favicon && !siteFaviconFile)" 
                    :src="siteFaviconPreview || (footerSettings.favicon ? `${footerSettings.favicon}?t=${Date.now()}` : '')" 
                    alt="网站图标预览"
                    class="favicon-preview-img"
                  />
                  
                  <!-- 空状态提示 -->
                  <div v-else class="favicon-empty-state">
                    <i class="bi bi-upload"></i>
                    <span>点击上传图标</span>
                  </div>
                  
                  <!-- 上传按钮 - 中央覆盖层 -->
                  <div class="favicon-actions">
                    <button 
                      type="button" 
                      class="favicon-action-btn upload-btn" 
                      @click.stop="triggerFileInput" 
                      title="上传新图标"
                    >
                      <i class="bi bi-arrow-up-circle"></i>
                    </button>
                  </div>
                  </div>
                
                <div class="form-text">
                  支持.ico、.png、.jpg和.svg格式，推荐尺寸为32x32或64x64像素。上传后的图标路径为：/assets/public/favicon.ico
                </div>
              </div>
            </div>
            
            <!-- 网站标题 -->
            <div class="form-group">
              <label class="form-label">网站标题</label>
              <div class="input-group">
                <div class="input-prefix">
                  <i class="bi bi-type-h1"></i>
                </div>
                <input 
                  type="text" 
                  class="custom-input" 
                  v-model="footerSettings.title" 
                  placeholder="网站标题"
                >
              </div>
              <div class="form-text">显示在浏览器标签页和搜索结果中的网站标题</div>
            </div>
            
            <!-- Logo文本 -->
            <div class="form-group">
              <label class="form-label">Logo文本</label>
              <div class="input-group">
                <div class="input-prefix">
                  <i class="bi bi-card-heading"></i>
                </div>
                <input 
                  type="text" 
                  class="custom-input" 
                  v-model="footerSettings.logoText" 
                  placeholder="左上角Logo旁的文本"
                >
              </div>
              <div class="form-text">网站左上角Logo旁边显示的文本</div>
            </div>
            
            <!-- 网站描述 -->
            <div class="form-group">
              <label class="form-label">网站描述</label>
              <div class="input-group">
                <div class="input-prefix">
                  <i class="bi bi-file-text"></i>
                </div>
                <textarea 
                  class="custom-input" 
                  v-model="footerSettings.description" 
                  placeholder="网站描述"
                  rows="3"
                ></textarea>
              </div>
              <div class="form-text">用于SEO的网站描述，显示在搜索引擎结果中</div>
            </div>
            
            <!-- 关键词 -->
            <div class="form-group">
              <label class="form-label">关键词</label>
              <div class="input-group">
                <div class="input-prefix">
                  <i class="bi bi-tags"></i>
                </div>
                <input 
                  type="text" 
                  class="custom-input" 
                  v-model="footerSettings.keywords" 
                  placeholder="关键词（用逗号分隔）"
                >
              </div>
              <div class="form-text">用于SEO的关键词，使用逗号分隔</div>
            </div>
          </div>
          
          <!-- 页面元信息设置标签页 -->
          <div class="settings-section" v-show="activeSettingsTab === 'meta'">
            <h5 class="section-title">页面元信息设置</h5>
            <div class="section-description">
              设置各个页面的标题、描述和关键词。这些信息将显示在浏览器标签页和搜索引擎结果中。
              如果不设置，将使用默认值。
            </div>
            
            <!-- 页面选择器 -->
            <div class="form-group">
              <label class="form-label">选择页面</label>
              <div class="input-group">
                <div class="input-prefix">
                  <i class="bi bi-file-earmark-text"></i>
                </div>
                <select 
                  class="custom-input" 
                  v-model="selectedPage"
                  @change="selectPage"
                >
                  <option value="home">首页</option>
                  <option value="resource_detail">资源详情页</option>
                  <option value="submit_resource">提交资源页</option>
                  <option value="login">登录页</option>
                  <option value="admin">管理后台页</option>
                  <option value="resource_review">资源审核页</option>
                  <option value="about">关于我们页</option>
                  <option value="streams">流媒体内容页</option>
                </select>
              </div>
            </div>
            
            <!-- 页面标题 -->
            <div class="form-group">
              <label class="form-label">页面标题</label>
              <div class="input-group">
                <div class="input-prefix">
                  <i class="bi bi-type-h2"></i>
                </div>
                <input 
                  type="text" 
                  class="custom-input" 
                  v-model="currentPageMeta.title" 
                  :placeholder="getDefaultPageMeta(selectedPage).title"
                >
              </div>
              <div class="form-text">
                在浏览器标签页显示的标题。留空则使用默认值：{{ getDefaultPageMeta(selectedPage).title }}
              </div>
            </div>
            
            <!-- 页面描述 -->
            <div class="form-group">
              <label class="form-label">页面描述</label>
              <div class="input-group">
                <div class="input-prefix">
                  <i class="bi bi-card-text"></i>
                </div>
                <textarea 
                  class="custom-input" 
                  v-model="currentPageMeta.description" 
                  :placeholder="getDefaultPageMeta(selectedPage).description"
                  rows="3"
                ></textarea>
              </div>
              <div class="form-text">用于SEO的页面描述</div>
            </div>
            
            <!-- 页面关键词 -->
            <div class="form-group">
              <label class="form-label">页面关键词</label>
              <div class="input-group">
                <div class="input-prefix">
                  <i class="bi bi-tag"></i>
                </div>
                <input 
                  type="text" 
                  class="custom-input" 
                  v-model="currentPageMeta.keywords" 
                  :placeholder="getDefaultPageMeta(selectedPage).keywords"
                >
              </div>
              <div class="form-text">用于SEO的页面关键词，使用逗号分隔</div>
            </div>
          </div>
          
          <!-- 页脚设置标签页 -->
          <div class="settings-section" v-show="activeSettingsTab === 'footer'">
            <h5 class="section-title">页脚设置</h5>
            
            <!-- 显示访问统计 -->
            <div class="form-group">
              <label class="form-label">显示访问统计</label>
              <div class="checkbox-wrapper horizontal-display">
                <input id="show_visitor_count" class="custom-checkbox" type="checkbox" v-model="footerSettings.show_visitor_count">
                <label for="show_visitor_count"></label>
                <span class="checkbox-text">在网站底部显示访问人数统计</span>
              </div>
            </div>
            
            <!-- 版权信息 -->
            <div class="form-group">
              <label class="form-label">版权信息</label>
              <div class="input-group">
                <div class="input-prefix">
                  <i class="bi bi-c-circle"></i>
                </div>
                <input 
                  type="text" 
                  class="custom-input" 
                  v-model="footerSettings.copyright" 
                  placeholder="版权信息文本"
                >
              </div>
            </div>
            
            <!-- 页脚链接列表 -->
            <div class="form-group">
              <label class="form-label">页脚链接</label>
              <div class="scroll-container links-wrapper">
                <div class="links-container">
                  <!-- 标签头部，只显示一次 -->
                  <div class="link-header">
                    <div class="drag-handle-placeholder"></div>
                    <div class="link-field-header">显示文本</div>
                    <div class="link-field-header">URL</div>
                    <div class="link-field-header">图标</div>
                    <div class="link-field-header actions-header"></div>
                  </div>
                  
                  <!-- 链接列表，支持拖拽排序 -->
                  <draggable
                    v-model="footerSettings.links"
                    item-key="id"
                    handle=".drag-handle"
                    ghost-class="ghost-item"
                    @end="onDragEnd"
                  >
                    <template #item="{ element, index }">
                      <div :key="index" class="link-item">
                        <div class="drag-handle">
                          <i class="bi bi-grip-vertical"></i>
                        </div>
                        <div class="link-fields">
                          <!-- 链接文本 -->
                          <div class="link-field" data-label="显示文本">
                            <input type="text" v-model="element.text" class="custom-input" placeholder="文本">
                          </div>
                          
                          <!-- 链接URL -->
                          <div class="link-field" data-label="URL">
                            <input type="text" v-model="element.url" class="custom-input" placeholder="链接地址">
                          </div>
                          
                          <!-- 图标 (可选) -->
                          <div class="link-field icon-field">
                            <div class="icon-selector-button" @click="openFooterIconSelector(index)" :title="element.icon ? '更改图标' : '选择图标'">
                              <i v-if="element.icon" :class="element.icon"></i>
                              <span v-else class="no-icon">无</span>
                              <button v-if="element.icon" type="button" class="clear-icon-btn" @click.stop="clearIcon(index)" title="清除图标">
                                <i class="bi bi-x"></i>
                              </button>
                            </div>
                          </div>
                          
                          <!-- 提示文本 -->
                          <div class="link-field" data-label="提示文本">
                            <input type="text" v-model="element.title" class="custom-input" placeholder="鼠标悬停提示">
                          </div>
                        </div>
                        
                        <!-- 删除按钮 -->
                        <button 
                          type="button" 
                          class="remove-link-btn" 
                          @click="removeLink(index)"
                          title="删除此链接"
                        >
                          <i class="bi bi-trash"></i>
                        </button>
                      </div>
                    </template>
                  </draggable>
                </div>
              </div>
            </div>
            
            <!-- 添加新链接按钮 -->
            <div class="add-link-wrapper">
              <button 
                type="button" 
                class="btn-custom btn-outline btn-sm add-link-btn" 
                @click="addNewLink"
              >
                <i class="bi bi-plus-circle"></i>
                <span class="btn-text">添加链接</span>
              </button>
            </div>
          </div>
          
          <!-- 关于页面设置标签页 -->
          <div class="settings-section" v-show="activeSettingsTab === 'about'">
            <h5 class="section-title">关于页面设置</h5>
            
            <!-- 本站介绍设置 -->
            <div class="form-group">
              <h6 class="subsection-title"><i class="bi bi-info-circle"></i> 本站介绍</h6>
              
              <!-- 标题 -->
              <div class="form-group">
                <label class="form-label">标题</label>
                <div class="input-group">
                  <div class="input-prefix">
                    <i class="bi bi-type-h1"></i>
                  </div>
                  <input 
                    type="text" 
                    class="custom-input" 
                    v-model="aboutPageConfig.siteIntro.title" 
                    placeholder="本站介绍"
                  >
                </div>
              </div>
              
              <!-- 描述 -->
              <div class="form-group">
                <label class="form-label">描述</label>
                <div class="input-group">
                  <div class="input-prefix">
                    <i class="bi bi-card-text"></i>
                  </div>
                  <textarea 
                    class="custom-input" 
                    v-model="aboutPageConfig.siteIntro.description" 
                    rows="3"
                    placeholder="输入本站介绍描述..."
                  ></textarea>
                </div>
              </div>
              
              <!-- 图标 -->
              <div class="form-group">
                <label class="form-label">图标</label>
                <div class="icon-selector-button" 
                  @click="openIntroIconSelector" 
                  :title="aboutPageConfig.siteIntro.icon ? '更改图标' : '选择图标'"
                >
                  <i v-if="aboutPageConfig.siteIntro.icon" :class="`bi bi-${aboutPageConfig.siteIntro.icon}`"></i>
                  <span v-else class="no-icon">无</span>
                  <button v-if="aboutPageConfig.siteIntro.icon" 
                    type="button" 
                    class="clear-icon-btn" 
                    @click.stop="aboutPageConfig.siteIntro.icon = ''" 
                    title="清除图标"
                  >
                    <i class="bi bi-x"></i>
                  </button>
                </div>
              </div>
            </div>
            
            <!-- 特性项目设置 -->
            <div class="form-group">
              <h6 class="subsection-title"><i class="bi bi-grid-3x3-gap"></i> 特性项目</h6>
              
              <div class="scroll-container links-wrapper">
                <div class="links-container">
                  <!-- 标签头部 -->
                  <div class="link-header">
                    <div class="drag-handle-placeholder"></div>
                    <div class="link-field-header">标题</div>
                    <div class="link-field-header">描述</div>
                    <div class="link-field-header">图标</div>
                    <div class="link-field-header actions-header"></div>
                  </div>
                  
                  <!-- 特性项目列表，支持拖拽排序 -->
                  <draggable
                    v-model="aboutPageConfig.featureItems"
                    item-key="id"
                    handle=".drag-handle"
                    ghost-class="ghost-item"
                    @end="onDragEnd"
                  >
                    <template #item="{ element, index }">
                      <div :key="element.id" class="link-item">
                        <div class="drag-handle">
                          <i class="bi bi-grip-vertical"></i>
                        </div>
                        <div class="link-fields">
                          <!-- 标题 -->
                          <div class="link-field" data-label="标题">
                            <input type="text" v-model="element.title" class="custom-input" placeholder="特性标题">
                          </div>
                          
                          <!-- 描述 -->
                          <div class="link-field description-field" data-label="描述">
                            <textarea v-model="element.description" class="custom-input" rows="3" placeholder="特性描述"></textarea>
                            <div class="form-text">支持HTML标签如<code>&lt;strong&gt;</code>, <code>&lt;br&gt;</code></div>
                          </div>
                          
                          <!-- 图标 -->
                          <div class="link-field icon-field">
                            <div class="icon-selector-button" 
                              @click="openFeatureIconSelector(index)" 
                              :title="element.icon ? '更改图标' : '选择图标'"
                            >
                              <i v-if="element.icon" :class="`bi bi-${element.icon}`"></i>
                              <span v-else class="no-icon">无</span>
                              <button v-if="element.icon" 
                                type="button" 
                                class="clear-icon-btn" 
                                @click.stop="clearFeatureIcon(index)" 
                                title="清除图标"
                              >
                                <i class="bi bi-x"></i>
                              </button>
                            </div>
                          </div>
                        </div>
                        
                        <!-- 删除按钮 -->
                        <button 
                          type="button" 
                          class="remove-link-btn" 
                          @click="removeFeatureItem(index)"
                          title="删除此特性项"
                        >
                          <i class="bi bi-trash"></i>
                        </button>
                      </div>
                    </template>
                  </draggable>
                </div>
              </div>
              
              <!-- 添加新特性按钮 -->
              <div class="add-link-wrapper">
                <button 
                  type="button" 
                  class="btn-custom btn-outline btn-sm add-link-btn" 
                  @click="addNewFeatureItem"
                >
                  <i class="bi bi-plus-circle"></i>
                  <span class="btn-text">添加特性项</span>
                </button>
              </div>
            </div>
            
            <!-- 免责声明设置 -->
            <div class="form-group">
              <h6 class="subsection-title"><i class="bi bi-shield-exclamation"></i> 免责声明</h6>
              
              <!-- 启用开关 -->
              <div class="form-group">
                <label class="form-label">显示免责声明</label>
                <div class="switch-toggle-wrapper">
                  <div class="toggle-switch">
                    <input 
                      type="checkbox" 
                      id="disclaimer-enabled" 
                      v-model="aboutPageConfig.disclaimer.enabled"
                    >
                    <label for="disclaimer-enabled" class="switch-label"></label>
                  </div>
                  <span class="switch-text">{{ aboutPageConfig.disclaimer.enabled ? '已启用' : '已禁用' }}</span>
                </div>
              </div>
              
              <!-- 标题 -->
              <div class="form-group" v-if="aboutPageConfig.disclaimer.enabled">
                <label class="form-label">标题</label>
                <div class="input-group">
                  <div class="input-prefix">
                    <i class="bi bi-type-h1"></i>
                  </div>
                  <input 
                    type="text" 
                    class="custom-input" 
                    v-model="aboutPageConfig.disclaimer.title" 
                    placeholder="免责声明"
                  >
                </div>
              </div>
              
              <!-- 内容 -->
              <div class="form-group" v-if="aboutPageConfig.disclaimer.enabled">
                <label class="form-label">内容 <small class="text-muted">(支持HTML)</small></label>
                <div class="input-group">
                  <div class="input-prefix">
                    <i class="bi bi-code-square"></i>
                  </div>
                  <textarea 
                    class="custom-input code-editor" 
                    v-model="aboutPageConfig.disclaimer.content" 
                    rows="8"
                    placeholder="输入免责声明内容，支持HTML标签..."
                  ></textarea>
                </div>
                <div class="form-text mt-1">
                  <button 
                    type="button" 
                    class="btn-custom btn-sm btn-text template-btn" 
                    @click="loadDisclaimerTemplate"
                  >
                    <i class="bi bi-file-earmark-text"></i> 加载默认模板
                  </button>
                </div>
              </div>
              
              <!-- 图标 -->
              <div class="form-group" v-if="aboutPageConfig.disclaimer.enabled">
                <label class="form-label">图标</label>
                <div class="icon-selector-button" 
                  @click="openDisclaimerIconSelector" 
                  :title="aboutPageConfig.disclaimer.icon ? '更改图标' : '选择图标'"
                >
                  <i v-if="aboutPageConfig.disclaimer.icon" :class="`bi bi-${aboutPageConfig.disclaimer.icon}`"></i>
                  <span v-else class="no-icon">无</span>
                  <button v-if="aboutPageConfig.disclaimer.icon" 
                    type="button" 
                    class="clear-icon-btn" 
                    @click.stop="aboutPageConfig.disclaimer.icon = ''" 
                    title="清除图标"
                  >
                    <i class="bi bi-x"></i>
                  </button>
                </div>
              </div>
            </div>
            
            <!-- 联系我们设置 -->
            <div class="form-group">
              <h6 class="subsection-title"><i class="bi bi-chat-text"></i> 联系我们</h6>
              
              <!-- 标题 -->
              <div class="form-group">
                <label class="form-label">标题</label>
                <div class="input-group">
                  <div class="input-prefix">
                    <i class="bi bi-type-h1"></i>
                  </div>
                  <input 
                    type="text" 
                    class="custom-input" 
                    v-model="aboutPageConfig.contactSection.title" 
                    placeholder="联系我们"
                  >
                </div>
              </div>
              
              <!-- 描述 -->
              <div class="form-group">
                <label class="form-label">描述</label>
                <div class="input-group">
                  <div class="input-prefix">
                    <i class="bi bi-card-text"></i>
                  </div>
                  <textarea 
                    class="custom-input" 
                    v-model="aboutPageConfig.contactSection.description" 
                    rows="3"
                    placeholder="输入联系我们描述..."
                  ></textarea>
                </div>
              </div>
              
              <!-- 图标 -->
              <div class="form-group">
                <label class="form-label">图标</label>
                <div class="icon-selector-button" 
                  @click="openContactSectionIconSelector" 
                  :title="aboutPageConfig.contactSection.icon ? '更改图标' : '选择图标'"
                >
                  <i v-if="aboutPageConfig.contactSection.icon" :class="`bi bi-${aboutPageConfig.contactSection.icon}`"></i>
                  <span v-else class="no-icon">无</span>
                  <button v-if="aboutPageConfig.contactSection.icon" 
                    type="button" 
                    class="clear-icon-btn" 
                    @click.stop="aboutPageConfig.contactSection.icon = ''" 
                    title="清除图标"
                  >
                    <i class="bi bi-x"></i>
                  </button>
                </div>
              </div>
              
              <!-- 联系方式列表 -->
              <div class="form-group contact-items-wrapper">
                <label class="form-label">联系方式列表</label>
              
                <div class="scroll-container links-wrapper contact-items-container">
                <div class="links-container">
                    <!-- 联系方式列表 -->
                  <draggable
                    v-model="aboutPageConfig.contactItems"
                    item-key="id"
                    handle=".drag-handle"
                    ghost-class="ghost-item"
                    @end="onDragEnd"
                  >
                    <template #item="{ element, index }">
                        <div :key="element.id" class="link-item contact-item">
                        <div class="drag-handle">
                          <i class="bi bi-grip-vertical"></i>
                        </div>
                        <div class="link-fields">
                          <!-- 联系方式文本 -->
                          <div class="link-field" data-label="联系方式">
                              <input type="text" v-model="element.text" class="custom-input" placeholder="联系方式内容">
                          </div>
                          
                          <!-- 图标 -->
                          <div class="link-field icon-field">
                            <div class="icon-selector-button" 
                              @click="openContactItemIconSelector(index)" 
                              :title="element.icon ? '更改图标' : '选择图标'"
                            >
                              <i v-if="element.icon" :class="`bi bi-${element.icon}`"></i>
                              <span v-else class="no-icon">无</span>
                              <button v-if="element.icon" 
                                type="button" 
                                class="clear-icon-btn" 
                                @click.stop="clearContactItemIcon(index)" 
                                title="清除图标"
                              >
                                <i class="bi bi-x"></i>
                              </button>
                            </div>
                          </div>

                        </div>
                        <!-- 删除按钮 -->
                        <button 
                          type="button" 
                          class="remove-link-btn" 
                          @click="removeContactItem(index)"
                          title="删除此联系方式"
                        >
                          <i class="bi bi-trash"></i>
                        </button>
                      </div>
                    </template>
                  </draggable>
                </div>
              </div>
              
              <!-- 添加新联系方式按钮 -->
              <div class="add-link-wrapper">
                <button 
                  type="button" 
                  class="btn-custom btn-outline btn-sm add-link-btn" 
                  @click="addNewContactItem"
                >
                  <i class="bi bi-plus-circle"></i>
                  <span class="btn-text">添加联系方式</span>
                </button>
                </div>
              </div>
              </div>
            </div>
          </div>
          
          <!-- 采集解析源设置标签页 -->
          <div class="settings-section" v-show="activeSettingsTab === 'datasources'">
            <h5 class="section-title">采集解析源设置</h5>
            <div class="section-description">
              配置资源采集和解析的数据源，系统将根据这些配置自动获取和更新资源信息。每个数据源需要设置三个基本属性：
              <ul class="description-list">
                <li><strong>名称</strong>：数据源的识别名称，便于管理</li>
                <li><strong>基础URL</strong>：API接口地址或网页链接，系统将从此地址获取数据</li>
                <li><strong>使用XML</strong>：是否以XML格式解析返回数据，不勾选则使用JSON格式</li>
              </ul>
              <div class="description-note">注意：配置更改后需点击底部"保存设置"按钮使其生效。数据源优先级按列表顺序，可拖动排序调整。</div>
            </div>
            
            <!-- 数据源链接列表 -->
            <div class="form-group">
              <label class="form-label">数据源列表</label>
              <div class="scroll-container links-wrapper">
                <div class="links-container">
                  <!-- 标签头部，只显示一次 -->
                  <div class="link-header datasource-header">
                    <div class="drag-handle-placeholder"></div>
                    <div class="datasource-header-fields">
                      <div class="link-field-header name-field">名称</div>
                      <div class="link-field-header url-field">基础URL</div>
                      <div class="link-field-header xml-field">使用XML</div>
                    </div>
                    <div class="link-field-header actions-header">操作</div>
                  </div>
                  
                  <!-- 数据源列表，支持拖拽排序 -->
                  <draggable
                    v-model="dataSources"
                    item-key="id"
                    handle=".drag-handle"
                    ghost-class="ghost-item"
                    @end="onDragEnd"
                  >
                    <template #item="{ element, index }">
                      <div :key="index" class="link-item datasource-item">
                        <div class="drag-handle">
                          <i class="bi bi-grip-vertical"></i>
                        </div>
                        <div class="link-fields datasource-fields">
                          <!-- 数据源名称 -->
                          <div class="link-field name-field" data-label="名称">
                            <input type="text" v-model="element.name" class="custom-input" placeholder="数据源名称">
                          </div>
                          
                          <!-- 基础URL -->
                          <div class="link-field url-field" data-label="基础URL">
                            <input type="text" v-model="element.baseUrl" class="custom-input" placeholder="https://example.com/api">
                          </div>
                          
                          <!-- 使用XML -->
                          <div class="link-field xml-field checkbox-field" data-label="XML">
                            <div class="checkbox-wrapper checkbox-center">
                              <input :id="`use-xml-${index}`" class="custom-checkbox" type="checkbox" v-model="element.useXml">
                              <label :for="`use-xml-${index}`"></label>
                            </div>
                          </div>
                        </div>
                        
                        <!-- 删除按钮 -->
                        <button 
                          type="button" 
                          class="remove-link-btn actions-field" 
                          @click="removeDataSource(index)"
                          title="删除此数据源"
                        >
                          <i class="bi bi-trash"></i>
                        </button>
                      </div>
                    </template>
                  </draggable>
                </div>
              </div>
            </div>
            
            <!-- 添加新数据源按钮 -->
            <div class="add-link-wrapper">
              <button 
                type="button" 
                class="btn-custom btn-outline btn-sm add-link-btn" 
                @click="addNewDataSource"
              >
                <i class="bi bi-plus-circle"></i>
                <span class="btn-text">添加数据源</span>
              </button>
            </div>
            
            <!-- 外部数据源管理 -->
            <div class="external-datasource-section mt-4">
              <h5 class="section-title">外部数据源</h5>
              <div class="section-description">
                添加自定义JavaScript文件作为数据源，适用于需要特殊处理的非标准API接口。
                <div class="description-note">注意：外部数据源可能存在安全风险，请确保只添加来自可信来源的脚本。</div>
              </div>
              
              <!-- 添加外部数据源 -->
              <div class="form-group">
                <label class="form-label">添加外部数据源</label>
                <div class="input-group">
                  <input 
                    type="text" 
                    v-model="externalSourceUrl" 
                    class="custom-input" 
                    placeholder="https://example.com/datasource.js"
                  >
                  <button 
                    @click="addExternalDataSource" 
                    class="btn-custom btn-primary" 
                    :disabled="!externalSourceUrl || externalSourceLoading"
                  >
                    <div v-if="externalSourceLoading" class="spinner small-spinner"></div>
                    <i v-else class="bi bi-plus-circle"></i>
                    <span class="btn-text">{{ externalSourceLoading ? '加载中...' : '添加' }}</span>
                  </button>
                </div>
                <div v-if="externalSourceError" class="error-message mt-2">
                  <i class="bi bi-exclamation-triangle-fill"></i>
                  {{ externalSourceError }}
                </div>
              </div>
              
              <!-- 外部数据源列表 -->
              <div v-if="externalDataSources.length > 0" class="form-group">
                <label class="form-label">已加载的外部数据源</label>
                <div class="scroll-container links-wrapper">
                  <div class="links-container">
                    <div v-for="source in externalDataSources" :key="source.id" class="external-source-item">
                      <div class="source-info">
                        <div class="source-name">{{ source.name }}</div>
                        <div class="source-url">{{ source.url }}</div>
                      </div>
                      <button 
                        @click="removeExternalDataSource(source.id)" 
                        class="btn-custom btn-danger btn-sm"
                      >
                        <i class="bi bi-trash"></i>
                        <span class="btn-text">删除</span>
                      </button>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
          
          <!-- TMDB配置标签页 -->
          <div class="settings-section" v-show="activeSettingsTab === 'tmdb'">
            <h5 class="section-title">TMDB API 配置</h5>
            
            <div v-if="tmdbSuccess" class="success-message">
              <i class="bi bi-check-circle-fill"></i>
              TMDB配置更新成功
            </div>
            <div v-if="tmdbError" class="error-message">
              <i class="bi bi-exclamation-triangle-fill"></i>
              {{ tmdbError }}
            </div>
            
            <!-- API密钥配置 -->
            <div class="form-group">
              <label class="form-label">TMDB API 密钥</label>
              <div class="input-group">
                <div class="input-prefix">
                  <i class="bi bi-key-fill"></i>
                </div>
                <input 
                  type="text" 
                  class="custom-input" 
                  v-model="tmdbSettings.apiKey" 
                  placeholder="输入TMDB API密钥"
                >
              </div>
              <div class="form-text">
                用于访问TMDB API的密钥，可从<a href="https://www.themoviedb.org/settings/api" target="_blank">TMDB官网</a>获取。留空将使用环境变量中的密钥或系统默认密钥。
              </div>
            </div>
            
            <!-- 启用TMDB功能开关 -->
            <div class="form-group">
              <label class="form-label d-flex align-items-center">
                <span class="me-2">启用TMDB功能</span>
                <div class="toggle-switch">
                  <input 
                    type="checkbox" 
                    id="tmdb-enabled" 
                    v-model="tmdbSettings.enabled"
                  >
                  <label for="tmdb-enabled"></label>
                </div>
              </label>
              <div class="form-text">
                开启后，导航栏将显示TMDB搜索按钮，允许用户使用TMDB搜索和导入资源。关闭后，将隐藏此功能。
              </div>
            </div>
            
            <!-- TMDB保存按钮 -->
            <div class="form-actions">
              <button 
                type="button" 
                class="btn-custom btn-primary" 
                @click="saveTMDBSettings"
                :disabled="tmdbLoading"
              >
                <div v-if="tmdbLoading" class="spinner"></div>
                <i class="bi bi-save"></i>
                <span class="btn-text">{{ tmdbLoading ? '保存中...' : '保存配置' }}</span>
              </button>
            </div>
        </div>
        
        <!-- 网站设置info保存按钮 -->
          <div class="form-actions" v-if="activeSettingsTab !== 'tmdb'">
          <!-- 成功提示在按钮上方 -->
            <div v-if="settingsSuccess" class="settings-success-message">
              <i class="bi bi-check-circle-fill"></i> 设置保存成功！
            </div>
            
            <button 
              type="button" 
              class="btn-custom btn-primary" 
              @click="saveAllSettings"
              :disabled="settingsLoading"
            >
              <div v-if="settingsLoading" class="spinner"></div>
              <i class="bi bi-save"></i>
            <span class="btn-text">保存设置</span>
            </button>
          </div>
        </div>
      </div>
      
      <!-- 图标选择器模态框 -->
      <div v-if="showIconSelector" class="custom-modal" @click.self="closeIconSelector">
        <div class="modal-dialog">
          <div class="modal-content">
            <div class="modal-header">
              <h5 class="modal-title"><i class="bi bi-images"></i> 选择图标</h5>
              <button type="button" class="close-btn" @click="closeIconSelector">
                <i class="bi bi-x-lg"></i>
              </button>
            </div>
            <div class="modal-body">
              <div class="search-box mb-3">
                <div class="input-group">
                  <div class="input-prefix">
                    <i class="bi bi-search"></i>
                  </div>
                  <input 
                    type="text" 
                    class="custom-input" 
                    v-model="iconSearch" 
                    placeholder="搜索图标..."
                    @input="filterIcons"
                  >
                </div>
              </div>
              
              <div class="icon-grid">
                <div 
                  v-for="icon in filteredIcons" 
                  :key="icon" 
                  class="icon-item"
                  :class="{ 'selected': currentIcon === icon }"
                  @click="selectIcon(icon)"
                >
                  <i :class="`bi bi-${icon}`"></i>
                  <span class="icon-name">{{ icon }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
      
      <!-- 待审批资源卡片 -->
      <div class="admin-card">
        <div class="card-header">
          <div class="header-left">
            <h4>
              <i class="bi bi-hourglass-split"></i> 待审批资源
              <div v-if="pendingResources.length > 0" class="badge-count badge-inline">{{ pendingResources.length }}</div>
            </h4>
          </div>
          <!-- 批量操作按钮 -->
          <div v-if="pendingResources.length > 0" class="header-actions">
            <div v-if="selectedPendingResources.length > 0" class="batch-actions">
              <span class="selected-count">已选择 {{ selectedPendingResources.length }} 项</span>
              <button 
                type="button"
                class="btn-custom btn-accent btn-sm"
                @click="batchRejectResources"
                :disabled="batchRejectLoading"
              >
                <div v-if="batchRejectLoading" class="spinner small-spinner"></div>
                <i v-else class="bi bi-x-circle"></i>
                <span class="btn-text">{{ batchRejectLoading ? '处理中...' : '批量拒绝' }}</span>
              </button>
            </div>
          </div>
        </div>
        <div class="card-body">
          <div v-if="loadingPending" class="loading-inline">
            <div class="spinner small-spinner"></div>
            <span>加载待审批资源...</span>
            </div>
          <div v-else-if="pendingResources.length === 0" class="empty-state">
            <i class="bi bi-inbox"></i>
            <p>没有待审批的资源</p>
          </div>
          <div v-else class="table-container">
            <table class="custom-table">
                <thead>
                  <tr>
                  <th class="checkbox-column">
                    <input 
                      type="checkbox"
                      :checked="isAllPendingSelected"
                      @change="toggleAllPending"
                      :indeterminate="isIndeterminate"
                    />
                  </th>
                  <th>ID</th>
                  <th>标题</th>
                  <th>类型</th>
                  <th>审批类型</th>
                  <th>图片</th>
                  <th>提交时间</th>
                  <th>操作</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="resource in pendingResources" :key="resource.id">
                  <td class="checkbox-column">
                    <input 
                      type="checkbox"
                      :value="resource.id"
                      v-model="selectedPendingResources"
                    />
                  </td>
                  <td><span class="id-badge">#{{ resource.id }}</span></td>
                    <td>{{ resource.title || resource.title_en }}</td>
                  <td><span class="type-badge">{{ resource.resource_type }}</span></td>
                  <td>
                    <span 
                      class="badge-supplement" 
                      v-if="resource.has_pending_supplement || resource.supplement?.status === 'pending'"
                    >
                      补充审批
                    </span>
                    <span class="badge-initial" v-else>初始审批</span>
                  </td>
                    <td>
                      <button 
                      class="btn-custom btn-outline btn-sm view-images-btn" 
                        @click="previewImages(resource)"
                      >
                      <i class="bi bi-images"></i> {{ resource.supplement?.images?.length || resource.images?.length || 0 }}
                      </button>
                    </td>
                    <td>{{ formatDate(resource.created_at) }}</td>
                  <td class="actions-cell">
                      <router-link 
                        :to="`/admin/resource-review/${resource.id}`" 
                      class="btn-custom btn-primary btn-sm"
                      >
                      <i class="bi bi-clipboard-check"></i> 
                      <span class="btn-text">开始审核</span>
                      </router-link>
                    </td>
                  </tr>
                </tbody>
              </table>
          </div>
        </div>
      </div>
      
      <!-- 已审批资源卡片 - 审批记录 -->
      <div class="admin-card">
        <div class="card-header">
          <h4><i class="bi bi-clipboard-check"></i> 审批记录</h4>
          <div class="header-actions">
            <button 
              type="button" 
              class="btn-custom btn-accent btn-sm" 
              @click="confirmBatchDelete"
              :disabled="selectedResources.length === 0"
              v-if="showApprovalRecords"
            >
              <i class="bi bi-trash"></i> 
              <span class="btn-text">批量删除</span> 
              <span v-if="selectedResources.length > 0" class="badge-count">{{ selectedResources.length }}</span>
            </button>
            
            <button 
              type="button" 
              class="btn-custom btn-outline toggle-btn" 
              @click="showApprovalRecords = !showApprovalRecords"
            >
              <i :class="showApprovalRecords ? 'bi bi-chevron-up' : 'bi bi-chevron-down'"></i>
              <span class="btn-text">{{ showApprovalRecords ? '收起' : '展开' }}</span>
            </button>
          </div>
        </div>
        <div class="card-body" v-if="showApprovalRecords">
          <div class="table-container">
            <table class="custom-table">
              <thead>
                <tr>
                  <th>
                    <div class="checkbox-wrapper">
                      <input id="select-all" class="custom-checkbox" type="checkbox" v-model="selectAll" @change="toggleAllSelection">
                      <label for="select-all"></label>
                    </div>
                  </th>
                  <th>记录ID</th>
                  <th>资源标题</th>
                  <th>资源类型</th>
                  <th>类型</th>
                  <th>审批结果</th>
                  <th>审批时间</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="resource in resources" :key="resource.id">
                  <td>
                    <div class="checkbox-wrapper">
                      <input :id="`resource-${resource.id}`" class="custom-checkbox" type="checkbox" :value="resource.id" v-model="selectedResources">
                      <label :for="`resource-${resource.id}`"></label>
                    </div>
                  </td>
                  <td><span class="id-badge">#{{ resource.id }}</span></td>
                  <td>{{ resource.title || resource.title_en }}</td>
                  <td><span class="type-badge">{{ resource.resource_type }}</span></td>
                  <td>
                    <span 
                      class="badge-supplement" 
                      v-if="resource.is_supplement_approval"
                    >
                      补充审批
                    </span>
                    <span class="badge-initial" v-else>初始审批</span>
                  </td>
                  <td>
                    <span 
                      class="status-badge" 
                      :class="{
                        'status-approved': resource.status === 'approved' || resource.status === 'APPROVED',
                        'status-rejected': resource.status === 'rejected' || resource.status === 'REJECTED'
                      }"
                    >
                      {{ resource.status === 'approved' || resource.status === 'APPROVED' ? '已处理' : '已拒绝' }}
                    </span>
                  </td>
                  <td>{{ formatDate(resource.updated_at || resource.created_at) }}</td>
                  <td class="actions-cell">
                    <button class="btn-custom btn-outline btn-sm" @click="showApprovalDetails(resource)">
                      <i class="bi bi-info-circle"></i> 
                      <span class="btn-text">详情</span>
                    </button>
                    <button class="btn-custom btn-accent btn-sm" @click="confirmDelete(resource)">
                      <i class="bi bi-trash"></i> 
                      <span class="btn-text">删除</span>
                    </button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
      
      <!-- 审批详情弹窗 -->
      <div v-if="showApprovalModal" class="custom-modal" @click.self="closeApprovalDetails">
        <div class="modal-dialog">
          <div class="modal-content">
            <div class="modal-header">
              <h5 class="modal-title">
                {{ selectedResource?.is_supplement_approval ? '补充内容审批详情' : '资源审批详情' }}
              </h5>
              <button type="button" class="close-btn" @click="closeApprovalDetails">
                <i class="bi bi-x-lg"></i>
              </button>
            </div>
            <div class="modal-body">
              <div v-if="selectedResource">
                <!-- 基本信息卡片 -->
                <div class="detail-section">
                  <h6 class="detail-title"><i class="bi bi-info-circle"></i> 基本信息</h6>
                  <div class="detail-content">
                    <div class="detail-item">
                      <span class="detail-label">记录ID:</span>
                      <span class="detail-value id-badge">#{{ selectedResource.id }}</span>
                    </div>
                    <div class="detail-item">
                      <span class="detail-label">资源ID:</span>
                      <span class="detail-value id-badge">#{{ selectedResource.resource_id }}</span>
                    </div>
                    <div class="detail-item">
                      <span class="detail-label">标题:</span>
                      <span class="detail-value">{{ selectedResource.title || selectedResource.title_en }}</span>
                    </div>
                    <div class="detail-item">
                      <span class="detail-label">类型:</span>
                      <span class="detail-value type-badge">{{ selectedResource.resource_type }}</span>
                    </div>
                    <div class="detail-item">
                      <span class="detail-label">审批类型:</span>
                      <span 
                        class="detail-value badge-supplement" 
                        v-if="selectedResource.is_supplement_approval"
                      >
                        补充审批
                      </span>
                      <span class="detail-value badge-initial" v-else>初始审批</span>
                    </div>
                    <div class="detail-item">
                      <span class="detail-label">状态:</span>
                      <span 
                        class="detail-value status-badge" 
                        :class="{
                          'status-approved': selectedResource.status === 'approved' || selectedResource.status === 'APPROVED',
                          'status-rejected': selectedResource.status === 'rejected' || selectedResource.status === 'REJECTED'
                        }"
                      >
                        {{ selectedResource.status === 'approved' || selectedResource.status === 'APPROVED' ? '已通过' : '已拒绝' }}
                      </span>
                    </div>
                  </div>
                </div>
                
                <!-- 审批信息卡片 -->
                <div class="detail-section">
                  <h6 class="detail-title"><i class="bi bi-clock-history"></i> 审批信息</h6>
                  <div class="detail-content">
                    <div class="detail-item">
                      <span class="detail-label">审批时间:</span>
                      <span class="detail-value">{{ formatDate(selectedResource.created_at) }}</span>
                    </div>
                    <div class="detail-item" v-if="selectedResource.poster_image">
                      <span class="detail-label">海报:</span>
                      <span class="detail-value">{{ getImageFileName(selectedResource.poster_image) }}</span>
                    </div>
                    <div class="detail-item" v-if="selectedResource.resource_id">
                      <span class="detail-label">资源链接:</span>
                      <router-link :to="`/resource/${selectedResource.resource_id}`" class="resource-link">
                        查看资源 (#{{ selectedResource.resource_id }})
                      </router-link>
                    </div>
                  </div>
                </div>
                
                <div class="detail-section">
                  <h6 class="detail-title"><i class="bi bi-chat-left-text"></i> 审批备注</h6>
                  <div class="detail-note">
                    {{ selectedResource.notes || '无审批备注' }}
                  </div>
                </div>
                
                <div class="detail-section" v-if="selectedResource.approved_images && selectedResource.approved_images.length > 0">
                  <h6 class="detail-title">
                    <i class="bi bi-images"></i> 已批准图片 
                    <span class="image-count">({{ selectedResource.approved_images.length }})</span>
                  </h6>
                  <div class="images-grid">
                    <div 
                      v-for="(image, index) in selectedResource.approved_images" 
                      :key="index" 
                      class="image-preview-item"
                      @click="openLargeImage(image)"
                    >
                      <div class="image-card">
                        <img :src="image" :alt="`图片 ${index+1}`">
                        <div class="image-overlay">
                          <span v-if="image === selectedResource.poster_image" class="poster-badge">海报图片</span>
                        </div>
                      </div>
                    </div>
                    <div v-if="!selectedResource.approved_images || selectedResource.approved_images.length === 0" class="empty-images">
                      <i class="bi bi-image"></i>
                      <p>无图片</p>
                    </div>
                  </div>
                </div>
                
                <!-- 显示链接信息 -->
                <div class="detail-section" v-if="selectedResource.approved_links && Object.keys(selectedResource.approved_links).length > 0">
                  <h6 class="detail-title"><i class="bi bi-link-45deg"></i> 已批准链接</h6>
                  <div class="links-container">
                    <div 
                      v-for="(links, category) in selectedResource.approved_links" 
                      :key="category" 
                      class="link-category"
                    >
                      <div class="category-name">{{ getCategoryDisplayName(category) }}</div>
                      <ul class="links-list">
                        <li v-for="(link, index) in links" :key="index" class="link-item">
                          <i class="bi bi-link-45deg"></i>
                          <span class="link-url">{{ typeof link === 'string' ? link : link.url }}</span>
                          <span v-if="typeof link === 'object' && link.password" class="link-password">
                            密码: {{ link.password }}
                          </span>
                        </li>
                      </ul>
                    </div>
                  </div>
                </div>
                
                <div class="modal-actions">
                  <router-link 
                    :to="`/resource/${selectedResource.resource_id}`"
                    class="btn-custom btn-primary"
                    target="_blank"
                  >
                    <i class="bi bi-box-arrow-up-right"></i> <span class="btn-text">查看公开页面</span>
                  </router-link>
                </div>
              </div>
            </div>
            <div class="modal-footer">
              <button type="button" class="btn-custom btn-outline" @click="closeApprovalDetails">关闭</button>
            </div>
          </div>
        </div>
      </div>
      
      <!-- 删除记录确认弹窗 -->
      <div v-if="showDeleteModal" class="custom-modal" @click.self="cancelDelete">
        <div class="modal-dialog small-dialog">
          <div class="modal-content">
            <div class="modal-header">
              <h5 class="modal-title"><i class="bi bi-trash"></i> 确认删除审批记录</h5>
              <button type="button" class="close-btn" @click="cancelDelete">
                <i class="bi bi-x-lg"></i>
              </button>
            </div>
            <div class="modal-body">
              <p class="confirm-message">确定要删除资源 <strong>{{ resourceToDelete?.title || resourceToDelete?.title_en }}</strong> 的审批记录吗？</p>
              <div class="info-box">
                <i class="bi bi-info-circle"></i>
                <div>此操作只会删除审批记录，不会影响已发布的资源本身。</div>
              </div>
            </div>
            <div class="modal-footer">
              <button type="button" class="btn-custom btn-outline" @click="cancelDelete">取消</button>
              <button type="button" class="btn-custom btn-accent" @click="removeResourceFromList" :disabled="deleteLoading">
                <div v-if="deleteLoading" class="spinner small-spinner"></div>
                <span>{{ deleteLoading ? '删除中...' : '确认删除' }}</span>
              </button>
            </div>
          </div>
        </div>
      </div>
      
      <!-- 图片预览弹窗 -->
      <div v-if="showImagePreview" class="custom-modal" @click.self="closeImagePreview">
        <div class="modal-dialog large-dialog">
          <div class="modal-content">
            <div class="modal-header">
              <h5 class="modal-title"><i class="bi bi-images"></i> 图片预览</h5>
              <button type="button" class="close-btn" @click="closeImagePreview">
                <i class="bi bi-x-lg"></i>
              </button>
            </div>
            <div class="modal-body">
              <div class="preview-grid">
                <div 
                  v-for="(image, index) in previewImagesList" 
                  :key="index"
                  class="preview-item"
                  @click="openLargeImage(image)"
                >
                  <img :src="image" :alt="`预览图片 ${index+1}`">
                </div>
              </div>
            </div>
            <div class="modal-footer">
              <button type="button" class="btn-custom btn-outline" @click="closeImagePreview">关闭</button>
            </div>
          </div>
        </div>
      </div>
      
      <!-- 大图预览 - 修改为与ResourceDetail.vue相同的风格 -->
      <div v-if="largeImageUrl" class="custom-modal" @click.self="closeLargeImage">
        <div class="modal-image-container">
          <button type="button" class="image-close-btn bi bi-x-lg me-2" @click="closeLargeImage"></button>
          <img :src="largeImageUrl" class="preview-large-image" alt="图片预览">
        </div>
      </div>
      
      <!-- 批量删除确认弹窗 -->
      <div v-if="showBatchDeleteModal" class="custom-modal" @click.self="cancelBatchDelete">
        <div class="modal-dialog small-dialog">
          <div class="modal-content">
            <div class="modal-header">
              <h5 class="modal-title"><i class="bi bi-trash"></i> 确认批量删除审批记录</h5>
              <button type="button" class="close-btn" @click="cancelBatchDelete">
                <i class="bi bi-x-lg"></i>
              </button>
            </div>
            <div class="modal-body">
              <p class="confirm-message">确定要删除选中的 <strong>{{ selectedResources.length }}</strong> 条审批记录吗？</p>
              <div class="info-box">
                <i class="bi bi-info-circle"></i>
                <div>此操作只会删除审批记录，不会影响已发布的资源本身。</div>
              </div>
            </div>
            <div class="modal-footer">
              <button type="button" class="btn-custom btn-outline" @click="cancelBatchDelete">取消</button>
              <button type="button" class="btn-custom btn-accent" @click="batchDeleteResources" :disabled="batchDeleteLoading">
                <div v-if="batchDeleteLoading" class="spinner small-spinner"></div>
                <span>{{ batchDeleteLoading ? '删除中...' : '确认删除' }}</span>
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted,  watch, computed } from 'vue'
import axios from 'axios'
import { useRouter } from 'vue-router'
import { isAuthenticated, isAdmin, debugAuth } from '../utils/auth'
import draggable from 'vuedraggable'
import infoManager from '../utils/InfoManager'
import { getDataSourceManager } from '../utils/dataSourceManager'
import iconList from '../utils/icons.js'
import { ElMessage } from 'element-plus'

const router = useRouter()
const resources = ref([])
const pendingResources = ref([])
const selectedPendingResources = ref([])
const batchRejectLoading = ref(false)
const loading = ref(true)
const loadingPending = ref(true)
const loadingUsers = ref(true)
const error = ref(null)
const showDeleteModal = ref(false)
const resourceToDelete = ref(null)
const deleteLoading = ref(false)
const approvalLoading = ref(null)

// 图片预览相关
const showImagePreview = ref(false)
const previewImagesList = ref([])
const largeImageUrl = ref(null) // 用于大图预览

// 密码修改相关状态
const showChangePassword = ref(false)
const passwordLoading = ref(false)
const passwordSuccess = ref(false)
const passwordError = ref(null)
const passwordForm = reactive({
  currentPassword: '',
  newPassword: '',
  confirmPassword: ''
})

// 网站设置相关状态
const showSiteSettings = ref(false)
const settingsLoading = ref(false)
const settingsSuccess = ref(false)
const settingsError = ref(null)
const activeSettingsTab = ref('basic') // 添加标签页状态变量
const footerSettings = ref({
  links: [],
  copyright: '',
  show_visitor_count: true
})

// 审批记录显示状态
const showApprovalRecords = ref(false)

// 图标选择器相关
const showIconSelector = ref(false)
const currentLinkIndex = ref(-1)
const iconSearch = ref('')
const currentIcon = ref('')
const iconSelectorTarget = ref('') // 用于区分当前为哪个元素选择图标（链接、介绍、特性等）
const iconSelectorIndex = ref(-1) // 当前编辑的项目索引
const iconDisplay = ref({}) // 用于显示图标的简化名称
const filteredIcons = ref([...iconList])

// 审批详情相关
const showApprovalModal = ref(false)
const selectedResource = ref(null)

// 批量删除相关
const selectedResources = ref([])
const selectAll = ref(false)
const showBatchDeleteModal = ref(false)
const batchDeleteLoading = ref(false)

// 格式化日期
const formatDate = (dateString) => {
  if (!dateString) return '未知'
  const date = new Date(dateString)
  return new Intl.DateTimeFormat('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  }).format(date)
}

// 批量操作相关计算属性
const isAllPendingSelected = computed(() => {
  if (pendingResources.value.length === 0) return false
  return selectedPendingResources.value.length === pendingResources.value.length
})

const isIndeterminate = computed(() => {
  return selectedPendingResources.value.length > 0 && selectedPendingResources.value.length < pendingResources.value.length
})

// 全选/取消全选
const toggleAllPending = (event) => {
  if (event.target.checked) {
    selectedPendingResources.value = pendingResources.value.map(r => r.id)
  } else {
    selectedPendingResources.value = []
  }
}

// 批量拒绝资源
const batchRejectResources = async () => {
  if (selectedPendingResources.value.length === 0) {
    ElMessage.warning('请选择要拒绝的资源')
    return
  }

  const confirmMsg = `确定要拒绝选中的 ${selectedPendingResources.value.length} 个资源吗？`
  if (!confirm(confirmMsg)) {
    return
  }

  batchRejectLoading.value = true
  
  try {
    // 使用新的批量拒绝API
    const response = await axios.post('/api/resources/batch-reject', {
      resource_ids: selectedPendingResources.value,
      notes: '批量拒绝'
    })
    
    const data = response.data
    
    if (data.success_count > 0) {
      ElMessage.success(`成功拒绝 ${data.success_count} 个资源`)
      // 重新加载待审批资源列表
      await fetchPendingResources()
      // 清空选中项
      selectedPendingResources.value = []
    }

    if (data.failed_ids && data.failed_ids.length > 0) {
      ElMessage.error(`${data.failed_ids.length} 个资源拒绝失败`)
    }
  } catch (err) {
    console.error('批量拒绝失败:', err)
    ElMessage.error('批量拒绝失败，请稍后重试')
  } finally {
    batchRejectLoading.value = false
  }
}

// 获取所有已审批资源
const fetchResources = async () => {
  loading.value = true
  error.value = null
  
  try {
    const token = localStorage.getItem('accessToken')
    
    if (!token) {
      console.log('No token found in fetchResources')
      router.push('/login')
      return
    }
    
    // 使用新的审批记录API端点
    console.log('Fetching approval records with auth token')
    const response = await axios.get('/api/resources/approval-records')
    
    // 处理响应中的审批记录数据
    resources.value = response.data.records || []
    console.log(`Fetched ${resources.value.length} approval records`)
  } catch (err) {
    console.error('获取审批记录失败:', err)
    if (err.response && err.response.status === 401) {
      error.value = '认证失败，请重新登录'
      setTimeout(() => {
        router.push('/login')
      }, 2000)
    } else {
      error.value = '获取审批记录失败，请稍后重试'
    }
  } finally {
    loading.value = false
  }
}

// 加载网站设置
const loadSiteSettings = async (settingType) => {
  try {
    console.log(`加载${settingType}设置...`);
    
    // 这里我们实际上不需要单独加载各种设置类型
    // 因为loadFooterSettings已经加载了所有需要的数据
    // 这个函数保留为兼容性目的，实际不做额外操作
    
    return true;
  } catch (error) {
    console.error(`加载${settingType}设置失败:`, error);
    return false;
  }
};

// 获取待审批资源
const fetchPendingResources = async () => {
  loadingPending.value = true
  
  try {
    const token = localStorage.getItem('accessToken')
    
    if (!token) {
      console.log('No token found in fetchPendingResources')
      router.push('/login')
      return
    }
    
    console.log('Fetching pending resources with auth token')
    // 修复URL格式，移除尾部斜杠
    const response = await axios.get('/api/resources/pending')
    pendingResources.value = response.data
    console.log(`Fetched ${pendingResources.value.length} pending resources`)
  } catch (err) {
    console.error('获取待审批资源失败:', err)
    if (err.response && err.response.status === 401) {
      // 避免在这里显示错误，让主界面处理认证失败的情况
      console.log('Authentication failed when fetching pending resources')
      router.push('/login')
    }
  } finally {
    loadingPending.value = false
  }
}

// 预览图片
const previewImages = (resource) => {
  previewImagesList.value = resource.supplement?.images || resource.images || [];
  showImagePreview.value = true
}

// 关闭图片预览
const closeImagePreview = () => {
  showImagePreview.value = false
  previewImagesList.value = []
  largeImageUrl.value = null // 确保关闭大图预览
}

// 打开大图预览
const openLargeImage = (imageUrl) => {
  largeImageUrl.value = imageUrl
}

// 关闭大图预览
const closeLargeImage = () => {
  largeImageUrl.value = null
}

// 修改密码
const changePassword = async () => {
  // 验证两次密码是否一致
  if (passwordForm.newPassword !== passwordForm.confirmPassword) {
    passwordError.value = '两次输入的新密码不一致'
    return
  }
  
  passwordLoading.value = true
  passwordError.value = null
  passwordSuccess.value = false
  
  try {
    const token = localStorage.getItem('accessToken')
    
    if (!token) {
      router.push('/login')
      return
    }
    
    await axios.post('/api/auth/change-password', {
      current_password: passwordForm.currentPassword,
      new_password: passwordForm.newPassword
    })
    
    // 清空表单
    passwordForm.currentPassword = ''
    passwordForm.newPassword = ''
    passwordForm.confirmPassword = ''
    
    // 显示成功消息
    passwordSuccess.value = true
    
    // 3秒后自动隐藏成功消息
    setTimeout(() => {
      passwordSuccess.value = false
    }, 3000)
    
  } catch (err) {
    console.error('修改密码失败:', err)
    if (err.response && err.response.status === 400) {
      passwordError.value = '当前密码不正确'
    } else {
      passwordError.value = '修改密码失败，请稍后重试'
    }
  } finally {
    passwordLoading.value = false
  }
}

// 确认删除
const confirmDelete = (resource) => {
  resourceToDelete.value = resource
  showDeleteModal.value = true
}

// 取消删除
const cancelDelete = () => {
  showDeleteModal.value = false
  resourceToDelete.value = null
}

// 从列表中删除记录（实际从数据库中删除记录，但不影响资源）
const removeResourceFromList = async () => {
  if (!resourceToDelete.value) return
  
  deleteLoading.value = true
  
  try {
    const token = localStorage.getItem('accessToken')
    
    if (!token) {
      router.push('/login')
      return
    }
    
    // 使用新的审批记录删除API
    await axios.delete(`/api/resources/${resourceToDelete.value.id}/record`)
    
    // 从前端列表中移除
    resources.value = resources.value.filter(r => r.id !== resourceToDelete.value.id)
    showDeleteModal.value = false
    resourceToDelete.value = null
  } catch (err) {
    console.error('删除审批记录失败:', err)
    error.value = '删除审批记录失败，请稍后重试'
  } finally {
    deleteLoading.value = false
  }
}

// 显示审批详情
const showApprovalDetails = async (resource) => {
  try {
    // 获取资源的审批记录详情
    const response = await axios.get(`/api/resources/${resource.resource_id}/approval-records`)
    
    // 使用资源和审批记录
    selectedResource.value = {
      ...resource,
      records: response.data.records || []
    }
    
    showApprovalModal.value = true
    
    console.log('Approval records loaded:', selectedResource.value.records ? 
                selectedResource.value.records.length : 'none')
  } catch (err) {
    console.error('获取审批记录详情失败:', err)
    error.value = '获取审批记录详情失败，请稍后重试'
  }
}

// 关闭审批详情
const closeApprovalDetails = () => {
  showApprovalModal.value = false
  selectedResource.value = null
}


// 获取图片文件名
const getImageFileName = (imagePath) => {
  if (!imagePath) return ''
  
  try {
    const parts = imagePath.split('/')
    return parts[parts.length - 1]
  } catch (error) {
    return imagePath
  }
}

// 获取链接分类显示名称
const getCategoryDisplayName = (category) => {
  const categoryDisplayNames = {
    "magnet": "磁力链接",
    "ed2k": "电驴(ed2k)",
    "uc": "UC网盘",
    "mobile": "移动云盘",
    "tianyi": "天翼云盘",
    "quark": "夸克网盘",
    "115": "115网盘",
    "aliyun": "阿里云盘",
    "pikpak": "PikPak",
    "baidu": "百度网盘",
    "123": "123网盘",
    "online": "在线观看",
    "others": "其他链接"
  }
  
  return categoryDisplayNames[category] || category
}

// 切换全选
const toggleAllSelection = () => {
  if (selectAll.value) {
    selectedResources.value = resources.value.map(r => r.id);
  } else {
    selectedResources.value = [];
  }
}

// 打开批量删除确认弹窗
const confirmBatchDelete = () => {
  if (selectedResources.value.length === 0) {
    return;
  }
  showBatchDeleteModal.value = true;
}

// 取消批量删除
const cancelBatchDelete = () => {
  showBatchDeleteModal.value = false;
}

// 批量删除资源记录
const batchDeleteResources = async () => {
  if (selectedResources.value.length === 0) {
    return;
  }
  
  batchDeleteLoading.value = true;
  
  try {
    const token = localStorage.getItem('accessToken');
    
    if (!token) {
      router.push('/login');
      return;
    }
    
    // 使用批量删除审批记录API
    await axios.delete('/api/resources/batch-delete-records', {
      headers: {
        'Authorization': `Bearer ${token}`
      },
      data: {  // DELETE请求的请求体需要放在data字段中
        ids: selectedResources.value
      }
    });
    
    // 从前端列表中移除已删除的记录
    resources.value = resources.value.filter(r => !selectedResources.value.includes(r.id));
    
    // 重置选择状态
    selectedResources.value = [];
    selectAll.value = false;
    showBatchDeleteModal.value = false;
    
  } catch (err) {
    console.error('批量删除审批记录失败:', err);
    error.value = '批量删除审批记录失败，请稍后重试';
  } finally {
    batchDeleteLoading.value = false;
  }
}

// 加载页脚设置
const loadFooterSettings = async () => {
  try {
    // 使用InfoManager获取缓存的信息
    const infoManager = (await import('../utils/InfoManager')).default;
    footerSettings.value = await infoManager.getInfo();
    console.log('网站设置加载成功:', footerSettings.value);
    
    // 确保基本设置字段存在
    if (!footerSettings.value.title) footerSettings.value.title = '美漫资源共建';
    if (!footerSettings.value.logoText) footerSettings.value.logoText = '美漫资源共建';
    if (!footerSettings.value.description) footerSettings.value.description = '美漫共建平台是一个开源的美漫资源共享网站，用户可以自由提交动漫信息，像马赛克一样，由多方贡献拼凑成完整资源。';
    if (!footerSettings.value.keywords) footerSettings.value.keywords = '美漫, 动漫资源, 资源共享, 开源平台, 美漫共建';
    if (!footerSettings.value.routeMeta) footerSettings.value.routeMeta = {};
    
    // 加载数据源配置
    if (footerSettings.value.dataSources && Array.isArray(footerSettings.value.dataSources)) {
      dataSources.value = footerSettings.value.dataSources;
      console.log(`从数据库加载 ${dataSources.value.length} 个数据源配置`);
    } else {
      // 初始化为空数组
      dataSources.value = [];
      
      // 只有在没有现有数据源配置时，才加载默认数据源
      try {
        const dataSourcesConfigModule = await import('../utils/dataSourcesConfig.js');
        const dataSourcesConfig = dataSourcesConfigModule.default;
        if (dataSourcesConfig && Array.isArray(dataSourcesConfig)) {
          // 转换为数据源格式
          const defaultDataSources = dataSourcesConfig.map((source, index) => ({
            id: Date.now() + index + 1000,  // 使用一个较大的基数，避免ID冲突
            name: source.name || '',
            baseUrl: source.baseUrl || '',
            useXml: source.useXml || false
          }));
          
          // 设置为默认数据源
          dataSources.value = defaultDataSources;
          console.log(`从默认配置加载 ${dataSources.value.length} 个数据源`);
        }
      } catch (err) {
        console.log('未找到dataSourcesConfig.js或加载过程中出错:', err);
      }
    }
    
    // 不再无条件加载默认数据源
    
    // 如果存在favicon，显示预览
    if (footerSettings.value.favicon) {
      console.log('Found existing favicon:', footerSettings.value.favicon);
      // 这里不需要设置siteFaviconPreview，因为我们在模板中直接使用footerSettings.value.favicon
    }
    
    // 初始化About页面配置
    if (footerSettings.value.aboutPageConfig) {
      await initAboutPageConfig(footerSettings.value.aboutPageConfig);
    } else {
      // 如果不存在配置，初始化为默认值
      await initAboutPageConfig({});
    }
    
    // 初始化图标显示名称
    updateIconDisplay();
    
    // 初始化当前页面Meta信息
    selectPage();
  } catch (error) {
    console.error('获取网站设置失败:', error);
    // 使用默认设置
    footerSettings.value = {
      title: '美漫资源共建',
      logoText: '美漫资源共建',
      description: '美漫共建平台是一个开源的美漫资源共享网站，用户可以自由提交动漫信息，像马赛克一样，由多方贡献拼凑成完整资源。',
      keywords: '美漫, 动漫资源, 资源共享, 开源平台, 美漫共建',
      links: [
        { id: 1, text: "关于我们", url: "/about", type: "internal" },
        { id: 2, text: "Telegram", url: "https://t.me/xueximeng", icon: "bi bi-telegram", type: "external", title: "加入Telegram群组" },
        { id: 3, text: "GitHub", url: "https://github.com/fish2018/GoComicMosaic", icon: "bi bi-github", type: "external", title: "查看GitHub源码" },
        { id: 4, text: "在线点播", url: "/streams", type: "internal" },
        { id: 5, text: "漫迪小站", url: "https://mdsub.top/", type: "external" },
        { id: 6, text: "三次元成瘾者康复中心", url: "https://www.kangfuzhongx.in/", type: "external" },
      ],
      copyright: "© 2025 美漫资源共建. 保留所有权利",
      show_visitor_count: true
    };
    
    // 初始化默认的About页面配置
    await initAboutPageConfig({});
    
    // 初始化图标显示名称
    updateIconDisplay();
  }
};

// 统一保存所有设置
const saveAllSettings = async () => {
  settingsLoading.value = true;
  aboutSettingsLoading.value = true;
  settingsError.value = null;
  settingsSuccess.value = false;
  aboutSettingsSuccess.value = false;
  
  try {
    // 获取令牌并验证
    const token = localStorage.getItem('accessToken');
    if (!token) {
      settingsError.value = '您的登录已过期，请重新登录';
      console.error('保存设置失败: 未找到认证令牌');
      return;
    }
    
    // 获取当前的网站信息
    const siteInfo = await infoManager.getInfo();
    
    // 记录原始favicon值，确保在后续处理中不会丢失
    const originalFavicon = siteInfo.favicon;
    console.log('原始favicon值:', originalFavicon);
    
    // 处理Favicon上传
    if (siteFaviconFile.value) {
      try {
        const formData = new FormData();
        formData.append('favicon', siteFaviconFile.value);
        
        // 上传favicon
        const response = await axios.post('/api/admin/upload/favicon', formData, {
          headers: {
            'Content-Type': 'multipart/form-data',
            'Authorization': `Bearer ${token}`
          }
        });
        
        if (response.data && response.data.faviconPath) {
          // 更新设置中的favicon路径
          footerSettings.value.favicon = response.data.faviconPath;
          console.log('Favicon上传成功:', response.data.faviconPath);
        }
      } catch (err) {
        console.error('上传Favicon失败:', err);
        settingsError.value = '上传图标失败，但其他设置仍将保存';
      }
    } else {
      // 没有新上传的图标，保留原有的favicon
      // 注意：我们不再依赖siteFaviconPreview.value作为判断依据
      footerSettings.value.favicon = originalFavicon;
      console.log('保留原始favicon:', originalFavicon);
    }
    
    // 更新About页面配置
    siteInfo.aboutPageConfig = JSON.parse(JSON.stringify(aboutPageConfig));
    
    // 更新采集解析源配置
    siteInfo.dataSources = JSON.parse(JSON.stringify(dataSources.value));
    
    // 更新所有页脚设置和其他设置
    // 使用解构赋值而不是Object.assign以避免覆盖其他重要属性
    siteInfo.title = footerSettings.value.title;
    siteInfo.logoText = footerSettings.value.logoText;
    siteInfo.description = footerSettings.value.description;
    siteInfo.keywords = footerSettings.value.keywords;
    siteInfo.links = footerSettings.value.links;
    siteInfo.copyright = footerSettings.value.copyright;
    siteInfo.show_visitor_count = footerSettings.value.show_visitor_count;
    siteInfo.favicon = footerSettings.value.favicon; // 使用上面处理后的favicon值
    siteInfo.routeMeta = footerSettings.value.routeMeta;
    
    console.log('保存前的favicon值:', siteInfo.favicon);
    
    // 保存更新后的网站信息
    await infoManager.updateInfo(siteInfo);
    
    // 清理和显示成功消息
    siteFaviconFile.value = null;
    siteFaviconPreview.value = '';
    settingsSuccess.value = true;
    aboutSettingsSuccess.value = true;
    
    // 设置成功消息显示时间
    setTimeout(() => {
      settingsSuccess.value = false;
      aboutSettingsSuccess.value = false;
    }, 3000);
    
    // 设置一个短暂的延迟后刷新页面，确保用户能看到成功消息
    setTimeout(() => {
      console.log('设置保存成功，刷新页面...');
      window.location.reload();
    }, 1500);
    
  } catch (error) {
    console.error('保存设置失败:', error);
    
    if (error.response && error.response.status === 401) {
      settingsError.value = '认证失败，请刷新页面或重新登录';
    } else {
      settingsError.value = '保存设置失败，请稍后重试';
    }
  } finally {
    settingsLoading.value = false;
    aboutSettingsLoading.value = false;
  }
};

// 添加新链接
const addNewLink = () => {
  if (!footerSettings.value.links) {
    footerSettings.value.links = [];
  }
  
  const newIndex = footerSettings.value.links.length;
  const newId = Date.now() + Math.floor(Math.random() * 1000);
  
  // 为每个链接添加唯一ID，用于拖拽排序（注意：移除了type字段）
  footerSettings.value.links.push({
    id: newId,
    text: '',
    url: '',
    icon: '',
    title: ''
  });
  
  // 初始化图标显示名称
  iconDisplay.value[newIndex] = '';
};

// 移除链接
const removeLink = (index) => {
  if (footerSettings.value.links && index >= 0 && index < footerSettings.value.links.length) {
    footerSettings.value.links.splice(index, 1);
  }
};

// 获取图标的简化名称（不带bi bi-前缀）
const getIconName = (iconClass) => {
  if (!iconClass) return '';
  return iconClass.replace('bi bi-', '');
};

// 打开图标选择器
const openIconSelector = (index) => {
  // 当处理页脚链接图标时
  if (iconSelectorTarget.value === '' || iconSelectorTarget.value === 'footer') {
    currentLinkIndex.value = index;
    currentIcon.value = getIconName(footerSettings.value.links[index]?.icon || '');
  }
  // 其他类型的图标选择器（About页面设置）不需要在这里设置currentIcon，
  // 因为它们已经在各自的函数中设置了

  iconSearch.value = '';
  filteredIcons.value = [...iconList];
  showIconSelector.value = true;
  
  // 只有在处理页脚链接时才更新图标显示名称
  if (iconSelectorTarget.value === '' || iconSelectorTarget.value === 'footer') {
    updateIconDisplay();
  }
};

// 关闭图标选择器
const closeIconSelector = () => {
  showIconSelector.value = false;
  currentLinkIndex.value = -1;
};

// 选择图标
const selectIcon = (icon) => {
  // 页脚链接图标处理
  if (currentLinkIndex.value !== -1) {
    footerSettings.value.links[currentLinkIndex.value].icon = `bi bi-${icon}`;
    iconDisplay.value[currentLinkIndex.value] = icon; // 更新显示名称
    closeIconSelector();
    return;
  }
  
  // About页面图标处理
  if (iconSelectorTarget.value === 'intro') {
    aboutPageConfig.siteIntro.icon = icon;
  } else if (iconSelectorTarget.value === 'feature' && iconSelectorIndex.value >= 0) {
    aboutPageConfig.featureItems[iconSelectorIndex.value].icon = icon;
  } else if (iconSelectorTarget.value === 'contactSection') {
    aboutPageConfig.contactSection.icon = icon;
  } else if (iconSelectorTarget.value === 'contactItem' && iconSelectorIndex.value >= 0) {
    aboutPageConfig.contactItems[iconSelectorIndex.value].icon = icon;
  } else if (iconSelectorTarget.value === 'disclaimer') {
    aboutPageConfig.disclaimer.icon = icon;
  }
  
  closeIconSelector();
};

// 清除图标
const clearIcon = (index) => {
  if (index >= 0 && index < footerSettings.value.links.length) {
    footerSettings.value.links[index].icon = '';
    iconDisplay.value[index] = '';
  }
};

// 更新所有图标的显示名称
const updateIconDisplay = () => {
  if (!footerSettings.value || !footerSettings.value.links) {
    console.log('链接尚未初始化，跳过图标显示更新');
    return;
  }
  
  footerSettings.value.links.forEach((link, index) => {
    iconDisplay.value[index] = getIconName(link.icon);
  });
};

// 过滤图标
const filterIcons = () => {
  const search = iconSearch.value.toLowerCase();
  if (!search) {
    filteredIcons.value = [...iconList];
  } else {
    filteredIcons.value = iconList.filter(icon => 
      icon.toLowerCase().includes(search)
    );
  }
};

// 拖拽排序结束
const onDragEnd = () => {
  console.log('链接排序已更新');
};

onMounted(async () => {
  console.log('Admin component mounted')
  
  // 调试打印当前认证状态
  const authStatus = debugAuth()
  
  // 检查本地存储中的token和用户信息
  if (!authStatus.isAuthenticated) {
    console.error('No valid authentication found, redirecting to login')
    router.push('/login')
    return
  }
  
  if (!authStatus.isAdmin) {
    console.error('User is not an admin, redirecting to home')
    router.push('/')
    return
  }
  
  // 并行加载资源列表和页脚设置
  loading.value = true
  loadingPending.value = true
  loadingUsers.value = true
  error.value = null
  
  try {
    console.log('Starting to fetch resources and pending resources')
    const resourcesPromise = fetchResources()
    const pendingResourcesPromise = fetchPendingResources()
    const footerSettingsPromise = loadFooterSettings()
    const usersPromise = loadUsers()
    
    const results = await Promise.allSettled([resourcesPromise, pendingResourcesPromise, footerSettingsPromise, usersPromise])
    
    console.log('Fetch results:', results.map(r => r.status))
    
    // 基础数据加载完成后，立即加载外部数据源
    try {
      await loadExternalDataSources();
      console.log('外部数据源加载完成');
    } catch (error) {
      console.error('加载外部数据源失败:', error);
    }
    
    // 检查是否有任何请求失败
    const anyFailed = results.some(result => result.status === 'rejected')
    if (anyFailed) {
      console.warn('Some requests failed, check the error logs above')
      
      // 获取失败的原因
      const failures = results
        .filter(result => result.status === 'rejected')
        .map(result => result.reason?.response?.status || 'Unknown error')
      
      console.error('Request failures:', failures)
      
      // 如果有401错误，认证可能有问题
      if (failures.includes(401)) {
        error.value = '认证失败，请重新登录'
        setTimeout(() => {
          router.push('/login')
        }, 2000)
      } else {
        error.value = '部分数据加载失败，请刷新页面重试'
      }
    }
  } catch (e) {
    console.error('Error during Admin initialization:', e)
    error.value = '初始化失败，请稍后重试'
  } finally {
    loading.value = false
    loadingPending.value = false
    loadingUsers.value = false
  }
  
  // 加载各种设置
  await Promise.all([
    loadSiteSettings('info'),
    loadSiteSettings('footer'),
    loadSiteSettings('about'),
    loadTMDBConfig() // 加载TMDB配置
  ]);
})

// 页面Meta信息设置
const selectedPage = ref('home')
const currentPageMeta = reactive({
  title: '',
  description: '',
  keywords: ''
})

// 默认页面Meta信息
const defaultPageMetaInfo = {
  home: {
    title: '美漫资源共建 - 动漫爱好者共同贡献的资源平台',
    description: '美漫共建平台是一个开源的美漫资源共享网站，用户可以自由提交动漫信息，像马赛克一样，由多方贡献拼凑成完整资源。',
    keywords: '美漫, 动漫资源, 资源共享, 开源平台, 美漫共建'
  },
  resource_detail: {
    title: '资源详情 - 美漫资源共建平台',
    description: '查看详细的动漫资源信息，包括简介、图片、下载链接等。在这里您可以浏览由社区贡献的美漫资源详情。',
    keywords: '美漫资源, 动漫详情, 资源下载, 美漫共建'
  },
  submit_resource: {
    title: '提交资源 - 美漫资源共建平台',
    description: '在这里提交您收集的美漫资源，包括标题、简介、链接等信息，与社区共同构建完整的资源库。',
    keywords: '提交资源, 分享美漫, 资源贡献, 美漫共建'
  },
  login: {
    title: '用户登录 - 美漫资源共建平台',
    description: '登录美漫资源共建平台，管理您的资源贡献并参与社区建设。',
    keywords: '用户登录, 账号登录, 美漫共建'
  },
  admin: {
    title: '管理后台 - 美漫资源共建平台',
    description: '美漫资源共建平台管理后台，用于管理用户提交的资源和维护网站内容。',
    keywords: '管理后台, 资源审核, 美漫共建'
  },
  resource_review: {
    title: '资源审核 - 美漫资源共建平台',
    description: '审核用户提交的美漫资源，确保内容质量和合规性。',
    keywords: '资源审核, 内容审核, 美漫共建'
  },
  about: {
    title: '关于我们 - 美漫资源共建平台',
    description: '了解美漫资源共建平台的宗旨、团队和发展历程。我们致力于为动漫爱好者提供优质的资源共享环境。',
    keywords: '关于我们, 平台介绍, 团队介绍, 美漫共建'
  },
  streams: {
    title: '流媒体内容 - 美漫资源共建平台',
    description: '浏览和观看各种高质量的动漫流媒体内容，包括动画、电影和连续剧。',
    keywords: '流媒体内容, 动漫视频, 在线观看, 美漫共建'
  }
}

// 获取默认页面Meta信息
const getDefaultPageMeta = (page) => {
  return defaultPageMetaInfo[page] || defaultPageMetaInfo.home
}

// 选择页面时更新当前页面Meta信息
const selectPage = () => {
  // 如果页面之前没有自定义设置，则使用空值
  if (!footerSettings.value.routeMeta) {
    footerSettings.value.routeMeta = {}
  }
  
  const pageTitleKey = `${selectedPage.value}_title`
  const pageDescKey = `${selectedPage.value}_description`
  const pageKeywordsKey = `${selectedPage.value}_keywords`
  
  // 从已有设置中获取值，或者使用空字符串
  currentPageMeta.title = footerSettings.value.routeMeta[pageTitleKey] || ''
  currentPageMeta.description = footerSettings.value.routeMeta[pageDescKey] || ''
  currentPageMeta.keywords = footerSettings.value.routeMeta[pageKeywordsKey] || ''
}

// 监听currentPageMeta变化，更新footerSettings
const updateCurrentPageMeta = () => {
  if (!footerSettings.value.routeMeta) {
    footerSettings.value.routeMeta = {}
  }
  
  const pageTitleKey = `${selectedPage.value}_title`
  const pageDescKey = `${selectedPage.value}_description`
  const pageKeywordsKey = `${selectedPage.value}_keywords`
  
  // 只有当值不为空时才设置，否则删除属性以使用默认值
  if (currentPageMeta.title) {
    footerSettings.value.routeMeta[pageTitleKey] = currentPageMeta.title
  } else {
    delete footerSettings.value.routeMeta[pageTitleKey]
  }
  
  if (currentPageMeta.description) {
    footerSettings.value.routeMeta[pageDescKey] = currentPageMeta.description
  } else {
    delete footerSettings.value.routeMeta[pageDescKey]
  }
  
  if (currentPageMeta.keywords) {
    footerSettings.value.routeMeta[pageKeywordsKey] = currentPageMeta.keywords
  } else {
    delete footerSettings.value.routeMeta[pageKeywordsKey]
  }
}

// 监听currentPageMeta变化
watch(currentPageMeta, () => {
  updateCurrentPageMeta()
}, { deep: true })

// 页脚设置加载后初始化当前页面Meta信息
watch(footerSettings, () => {
  selectPage()
}, { immediate: true })

// Favicon上传功能变量
const siteFaviconPreview = ref('');
const siteFaviconFile = ref(null);
const faviconUploadRef = ref(null);

// 处理图标上传
const handleFaviconUpload = (event) => {
  const file = event.target.files[0];
  if (!file) return;
  
  // 检查文件扩展名，而不是MIME类型
  const fileName = file.name.toLowerCase();
  const validExtensions = ['.ico', '.png', '.jpg', '.jpeg', '.svg'];
  const isValid = validExtensions.some(ext => fileName.endsWith(ext));
  
  if (!isValid) {
    alert('请上传.ico、.png、.jpg或.svg格式的图标');
    return;
  }
  
  // 检查文件大小（限制为1MB）
  if (file.size > 1024 * 1024) {
    alert('图标文件大小不能超过1MB');
    return;
  }
  
  // 生成预览
  siteFaviconFile.value = file;
  const reader = new FileReader();
  reader.onload = (e) => {
    siteFaviconPreview.value = e.target.result;
  };
  reader.readAsDataURL(file);
};

// 触发文件输入控件
const triggerFileInput = () => {
  if (faviconUploadRef.value) {
    faviconUploadRef.value.click();
  }
};

// About页面设置
const aboutPageConfig = reactive({
  siteIntro: {
    title: '',
    description: '',
    icon: ''
  },
  featureItems: [],
  contactSection: {
    title: '',
    description: '',
    icon: ''
  },
  contactItems: [],
  disclaimer: {
    enabled: false,
    title: '',
    content: '',
    icon: ''
  }
});

const aboutSettingsLoading = ref(false);
const aboutSettingsSuccess = ref(false);

// 初始化About页面配置
const initAboutPageConfig = async (config) => {
  // 确保所有必要的嵌套对象都存在
  aboutPageConfig.siteIntro = config.siteIntro || {
    title: '本站介绍',
    description: '欢迎来到美漫资源共建平台，我们致力于为美漫爱好者提供一个便捷、高效、安全的资源分享平台。',
    icon: 'collection-fill'
  };
  aboutPageConfig.featureItems = config.featureItems || [];
  aboutPageConfig.contactSection = config.contactSection || {
    title: '联系我们',
    description: '如有任何问题、建议或合作意向，欢迎通过以下方式联系我们。我们非常重视每一位用户的反馈。',
    icon: 'chat-text-fill'
  };
  aboutPageConfig.contactItems = config.contactItems || [];
  aboutPageConfig.disclaimer = config.disclaimer || {
    enabled: false,
    title: '免责声明',
    content: '',
    icon: ''
  };
  
  // 确保featureItems和contactItems有唯一ID
  aboutPageConfig.featureItems.forEach((item, index) => {
    if (!item.id) item.id = Date.now() + index;
  });
  
  aboutPageConfig.contactItems.forEach((item, index) => {
    if (!item.id) item.id = Date.now() + index + 1000;
  });
};

// 添加新特性项
const addNewFeatureItem = () => {
  const newId = Date.now();
  aboutPageConfig.featureItems.push({
    id: newId,
    title: '',
    description: '',
    icon: ''
  });
};

// 删除特性项
const removeFeatureItem = (index) => {
  // aboutPageConfig.featureItems.splice(index, 1);
  if (confirm('删除会立刻执行，确定要删除这个特性项吗？')) {
    aboutPageConfig.featureItems.splice(index, 1);
  }
};

// 添加新联系方式
const addNewContactItem = () => {
  const newId = Date.now();
  aboutPageConfig.contactItems.push({
    id: newId,
    text: '',
    icon: ''
  });
};

// 删除联系方式
const removeContactItem = (index) => {
  // aboutPageConfig.contactItems.splice(index, 1);
  if (confirm('删除会立刻执行，确定要删除这个联系方式吗？')) {
    aboutPageConfig.contactItems.splice(index, 1);
  }
};

// 打开介绍部分图标选择器
const openIntroIconSelector = () => {
  iconSelectorTarget.value = 'intro';
  currentIcon.value = aboutPageConfig.siteIntro.icon;
  openIconSelector();
};

// 打开特性项图标选择器
const openFeatureIconSelector = (index) => {
  iconSelectorTarget.value = 'feature';
  iconSelectorIndex.value = index;
  currentIcon.value = aboutPageConfig.featureItems[index].icon;
  openIconSelector();
};

// 清除特性项图标
const clearFeatureIcon = (index) => {
  aboutPageConfig.featureItems[index].icon = '';
};

// 打开联系部分图标选择器
const openContactSectionIconSelector = () => {
  iconSelectorTarget.value = 'contactSection';
  currentIcon.value = aboutPageConfig.contactSection.icon;
  openIconSelector();
};

// 打开联系方式图标选择器
const openContactItemIconSelector = (index) => {
  iconSelectorTarget.value = 'contactItem';
  iconSelectorIndex.value = index;
  currentIcon.value = aboutPageConfig.contactItems[index].icon;
  openIconSelector();
};

// 清除联系方式图标
const clearContactItemIcon = (index) => {
  aboutPageConfig.contactItems[index].icon = '';
};

// 打开页脚链接图标选择器
const openFooterIconSelector = (index) => {
  iconSelectorTarget.value = 'footer';
  currentLinkIndex.value = index;
  currentIcon.value = getIconName(footerSettings.value.links[index]?.icon || '');
  openIconSelector(index);
}

// 数据源设置
const dataSources = ref([]);

// 添加新数据源
const addNewDataSource = () => {
  dataSources.value.push({
    id: Date.now() + Math.floor(Math.random() * 1000),
    name: '',
    baseUrl: '',
    useXml: false
  });
};

// 移除数据源
const removeDataSource = (index) => {
  dataSources.value.splice(index, 1);
  
  // 提示用户需要保存设置
  settingsSuccess.value = true;
  setTimeout(() => {
    settingsSuccess.value = false;
  }, 500);
  
  setTimeout(() => {
    settingsError.value = "请记得点击底部的'保存设置'按钮，否则删除操作不会永久生效！";
    setTimeout(() => {
      settingsError.value = null;
    }, 3000);
  }, 600);
};

// 外部数据源管理
const externalSourceUrl = ref('');
const externalSourceLoading = ref(false);
const externalSourceError = ref(null);
const externalDataSources = ref([]);

// 加载外部数据源列表
const loadExternalDataSources = async () => {
  try {
    const dataSourceManager = getDataSourceManager();
    
    // 确保dataSourceManager和dataSources存在
    if (!dataSourceManager || !dataSourceManager.dataSources) {
      console.log('数据源管理器尚未初始化或不包含数据源');
      externalDataSources.value = [];
      return;
    }
    
    externalDataSources.value = Object.values(dataSourceManager.dataSources)
      .filter(ds => ds && ds.isExternal)
      .map(ds => ({
        id: ds.id,
        name: ds.name,
        url: ds.sourceUrl
      }));
    
    console.log(`已加载 ${externalDataSources.value.length} 个外部数据源`);
  } catch (error) {
    console.error('加载外部数据源列表失败:', error);
    externalDataSources.value = [];
  }
};

// 添加外部数据源
const addExternalDataSource = async () => {
  if (!externalSourceUrl.value) {
    externalSourceError.value = '请输入外部数据源URL';
    return;
  }
  
  // 检查URL格式
  try {
    new URL(externalSourceUrl.value);
  } catch (e) {
    externalSourceError.value = 'URL格式不正确';
    return;
  }
  
  externalSourceLoading.value = true;
  externalSourceError.value = null;
  
  try {
    const dataSourceManager = getDataSourceManager();
    
    if (!dataSourceManager) {
      throw new Error('数据源管理器未初始化');
    }
    
    const dataSource = await dataSourceManager.loadExternalDataSource(externalSourceUrl.value);
    
    // 成功加载后清空输入框
    externalSourceUrl.value = '';
    
    // 刷新列表
    await loadExternalDataSources();
    
    // 显示成功消息
    settingsSuccess.value = true;
    setTimeout(() => {
      settingsSuccess.value = false;
    }, 3000);
  } catch (error) {
    console.error('加载外部数据源失败:', error);
    externalSourceError.value = `加载失败: ${error.message || '未知错误'}`;
  } finally {
    externalSourceLoading.value = false;
  }
};

// 移除外部数据源
const removeExternalDataSource = async (id) => {
  try {
    const dataSourceManager = getDataSourceManager();
    
    if (!dataSourceManager) {
      throw new Error('数据源管理器未初始化');
    }
    
    dataSourceManager.removeExternalDataSource(id);
    
    // 刷新列表
    await loadExternalDataSources();
    
    // 显示成功消息
    settingsSuccess.value = true;
    setTimeout(() => {
      settingsSuccess.value = false;
    }, 3000);
  } catch (error) {
    console.error('删除外部数据源失败:', error);
    externalSourceError.value = `删除失败: ${error.message || '未知错误'}`;
    settingsError.value = '删除外部数据源失败: ' + error.message;
    setTimeout(() => {
      settingsError.value = null;
    }, 3000);
  }
};

// 监听标签页切换，在切换到采集解析源标签页时刷新外部数据源列表
watch(activeSettingsTab, async (newTab) => {
  if (newTab === 'datasources') {
    // 切换到采集解析源标签页时自动刷新外部数据源列表
    try {
      await loadExternalDataSources();
      console.log('在切换到采集解析源标签页时刷新外部数据源列表');
    } catch (error) {
      console.error('刷新外部数据源列表失败:', error);
    }
  }
});

// TMDB配置
const tmdbSettings = reactive({
  apiKey: '',
  enabled: true
});

const tmdbLoading = ref(false);
const tmdbSuccess = ref(false);
const tmdbError = ref(null);

// 加载TMDB配置
const loadTMDBConfig = async () => {
  try {
    // 获取令牌并验证
    const token = localStorage.getItem('accessToken');
    if (!token) {
      console.error('加载TMDB配置失败: 未找到认证令牌');
      return;
    }
    
    // 请求TMDB配置
    const response = await axios.get('/api/admin/tmdb/config', {
      headers: {
        'Authorization': `Bearer ${token}`
      }
    });
    
    // 更新设置
    if (response.data && response.data.setting_value) {
      tmdbSettings.apiKey = response.data.setting_value.api_key || '';
      tmdbSettings.enabled = response.data.setting_value.enabled !== false; // 如果未设置，默认为true
    }
    
  } catch (error) {
    console.error('加载TMDB配置失败:', error);
  }
};

// 保存TMDB配置
const saveTMDBSettings = async () => {
  tmdbLoading.value = true;
  tmdbError.value = null;
  tmdbSuccess.value = false;
  
  try {
    // 获取令牌并验证
    const token = localStorage.getItem('accessToken');
    if (!token) {
      tmdbError.value = '您的登录已过期，请重新登录';
      console.error('保存TMDB配置失败: 未找到认证令牌');
      return;
    }
    
    // 更新TMDB配置
    tmdbSettings.apiKey = tmdbSettings.apiKey.trim();
    
    // 保存TMDB配置
    await axios.put('/api/admin/tmdb/config', {
      api_key: tmdbSettings.apiKey,
      enabled: tmdbSettings.enabled
    }, {
      headers: {
        'Authorization': `Bearer ${token}`
      }
    });
    
    // 清除TMDB状态缓存，确保前端状态立即更新
    const TmdbStatusService = (await import('../services/TmdbStatusService')).default;
    TmdbStatusService.clearCache();
    
    // 通知App.vue重新加载TMDB状态
    window.dispatchEvent(new Event('tmdb-config-updated'));
    
    // 显示成功消息
    tmdbSuccess.value = true;
    setTimeout(() => {
      tmdbSuccess.value = false;
    }, 3000);
    
  } catch (error) {
    console.error('保存TMDB配置失败:', error);
    
    if (error.response && error.response.status === 401) {
      tmdbError.value = '认证失败，请刷新页面或重新登录';
    } else {
      tmdbError.value = '保存TMDB配置失败，请稍后重试';
    }
  } finally {
    tmdbLoading.value = false;
  }
};

// 打开免责声明图标选择器
const openDisclaimerIconSelector = () => {
  iconSelectorTarget.value = 'disclaimer';
  currentIcon.value = aboutPageConfig.disclaimer.icon;
  openIconSelector();
};

// 加载默认免责声明模板
const loadDisclaimerTemplate = () => {
  // 确认对话框
  if (confirm('加载默认模板将覆盖当前内容，确认继续？')) {
    // 默认免责声明模板
    const defaultDisclaimerTemplate = `
<div class="disclaimer-section-content">
  <p><strong>内容来源：</strong>影视信息来自 TMDB，资源链接来自互联网公开搜索。</p>
  <p><strong>用户责任：</strong>用户需自行判断链接的安全性与合法性，遵守当地法律法规。</p>
  <p><strong>免责条款：</strong>本网站不对内容的准确性、合法性或可用性负责，不承担因使用本网站产生的任何损害责任。</p>
  <p><strong>版权声明：</strong>本网站不拥有任何资源版权，如有侵权请联系我们处理。</p>
  <p><strong>适用法律：</strong>本声明受您所在国家/地区适用法律管辖。</p>
</div>`;
    
    // 设置免责声明内容
    aboutPageConfig.disclaimer.content = defaultDisclaimerTemplate.trim();
    aboutPageConfig.disclaimer.enabled = true;
    aboutPageConfig.disclaimer.title = '免责声明';
    aboutPageConfig.disclaimer.icon = 'shield-exclamation';
    
    // 提示成功
    alert('已加载默认免责声明模板');
  }
};

// 用户管理相关状态
const showUserManagement = ref(false)
const users = ref([])
// loadingUsers 已在文件开头声明
const userDialogVisible = ref(false)
const isEditing = ref(false)
const userChangePassword = ref(false)
const submitting = ref(false)
const deleteUserDialogVisible = ref(false)
const userToDelete = ref(null)
const deleting = ref(false)
const currentUser = ref(null)

// 用户表单
const userForm = ref({
  id: null,
  username: '',
  password: '',
  is_admin: ''
})

// 加载用户列表
async function loadUsers() {
  loadingUsers.value = true
  try {
    // 获取用户列表
    const response = await axios.get('/api/admin/users')
    users.value = response.data
    console.log(`加载了 ${users.value.length} 个用户`)
    
    // 获取当前用户信息
    const userResponse = await axios.get('/api/auth/me')
    currentUser.value = userResponse.data
  } catch (error) {
    console.error('获取用户列表失败:', error)
    ElMessage.error('获取用户列表失败')
  } finally {
    loadingUsers.value = false
  }
}

// 显示添加用户对话框
function showAddUserDialog() {
  isEditing.value = false
  userChangePassword.value = true
  userForm.value = {
    id: null,
    username: '',
    password: '',
    is_admin: ''
  }
  userDialogVisible.value = true
}

// 显示编辑用户对话框
function showEditUserDialog(user) {
  isEditing.value = true
  userChangePassword.value = false
  userForm.value = {
    id: user.id,
    username: user.username,
    password: '',
    is_admin: user.is_admin
  }
  userDialogVisible.value = true
}

// 确认删除用户
function confirmDeleteUser(user) {
  userToDelete.value = user
  deleteUserDialogVisible.value = true
}

// 取消删除用户
function cancelDeleteUser() {
  deleteUserDialogVisible.value = false
  userToDelete.value = null
}

// 提交用户表单
async function submitUserForm() {
  // 表单验证
  if (!userForm.value.username || ((!isEditing.value || userChangePassword.value) && !userForm.value.password)) {
    ElMessage.error('请填写必填字段')
    return
  }
  
  if (userForm.value.username.length < 3 || userForm.value.username.length > 20) {
    ElMessage.error('用户名长度应为3-20个字符')
    return
  }
  
  if ((!isEditing.value || userChangePassword.value) && userForm.value.password.length < 6) {
    ElMessage.error('密码长度不能少于6个字符')
    return
  }

  // 如果是编辑模式且未选择修改密码，移除密码字段
  const userData = { ...userForm.value }
  if (isEditing.value && !userChangePassword.value) {
    delete userData.password
  }

  submitting.value = true
  try {
    if (isEditing.value) {
      // 编辑用户
      await axios.put(`/api/admin/users/${userData.id}`, userData)
      ElMessage.success('用户更新成功')
    } else {
      // 添加用户
      await axios.post('/api/admin/users', userData)
      ElMessage.success('用户添加成功')
    }
    userDialogVisible.value = false
    await loadUsers()
  } catch (error) {
    console.error('操作失败:', error)
    ElMessage.error(`${isEditing.value ? '更新' : '添加'}用户失败: ${error.response?.data?.error || error.message}`)
  } finally {
    submitting.value = false
  }
}

// 删除用户
async function deleteUser() {
  if (!userToDelete.value) return

  deleting.value = true
  try {
    await axios.delete(`/api/admin/users/${userToDelete.value.id}`)
    ElMessage.success('用户删除成功')
    deleteUserDialogVisible.value = false
    await loadUsers()
  } catch (error) {
    console.error('删除失败:', error)
    ElMessage.error(`删除用户失败: ${error.response?.data?.error || error.message}`)
  } finally {
    deleting.value = false
  }
}

// 计算属性：判断当前用户是否为超级管理员
const isSuperAdmin = computed(() => {
  return currentUser.value && currentUser.value.id === 1;
});

</script>

<style scoped src="@/styles/Admin.css"></style>