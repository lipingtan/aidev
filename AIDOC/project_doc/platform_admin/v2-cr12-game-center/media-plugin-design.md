# Media 插件设计文档

## 目标

独立素材管理插件，支持文件上传/管理/分类，对外暴露 MediaPicker 全局弹窗供其他插件集成。

## 架构

### 后端

**StorageProvider 抽象接口**（`app/media/service/storage.go`）：
```go
type StorageProvider interface {
    Upload(ctx, filename, contentType string, data io.Reader) (StorageFile, error)
    Delete(ctx, path string) error
    GetSignedURL(path string, expireSecs int64) (string, error)
    GetPublicURL(path string) string
}
```

**数据模型**：
- `MediaFile`：id, tenant_id, name, original_name, path, url, size, mime_type, category_id, uploader_id, created_at
- `MediaCategory`：id, tenant_id, name, parent_id（嵌套分类）, sort, created_at

**Provider 实现**：
- `TencentCOSProvider`：腾讯云 COS（优先实现）
- `AliyunOSSProvider`：预留 stub
- `LocalProvider`：本地存储（开发用）

**配置**（settings.yml）：
```yaml
media_storage:
  provider: tencent_cos  # tencent_cos / aliyun_oss / local
  tencent_cos:
    secret_id: ""
    secret_key: ""
    bucket: ""
    region: "ap-guangzhou"
    base_url: ""          # 自定义 CDN 域名（可选）
```

**API 路由**：
- 管理端：`/api/v1/admin/media/files` CRUD + 上传
- 管理端：`/api/v1/admin/media/categories` CRUD
- 管理端：`/api/v1/admin/media/config` 配置读写
- 公开：`/api/v1/user/public/media/signed-url` 获取签名链接

### 管理端 UI

**MediaManager.vue**（`/media/files`）：
- 左侧：分类树（可折叠，支持新建/重命名/删除）
- 右侧：文件卡片列表（图片直接预览，其他文件显示类型图标）
- 顶部工具栏：上传按钮、搜索、筛选（类型）
- 卡片操作：复制链接、删除、移动分类

**MediaPicker 全局弹窗**（`src/components/MediaPicker/`）：
- 通过事件总线调用：`eventBus.emit('media:pick', { accept, multiple, onSelect })`
- 弹窗内嵌入完整的文件浏览 + 上传功能
- 确认选择后通过回调返回文件列表

**其他插件使用方式**：
```typescript
import { openMediaPicker } from '@/components/MediaPicker'

openMediaPicker({
  accept: 'image/*',
  multiple: false,
  onSelect: (files) => { form.cover = files[0].url }
})
```

## 路径规划

| 路径 | 说明 |
|------|------|
| `/media/files` | 素材管理主页面 |
| `/media/config` | OSS 配置页 |

## 权限

- `media:files` — 素材列表/查看
- `media:files:upload` — 上传
- `media:files:delete` — 删除
- `media:categories` — 分类管理
- `media:config` — 配置管理
