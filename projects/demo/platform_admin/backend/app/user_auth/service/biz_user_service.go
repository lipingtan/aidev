package service

import (
	"log"
	"strconv"
	"crypto/rand"
	"errors"
	"math/big"

	"go-admin/app/user_auth/dto"
	"go-admin/app/user_auth/model"
	"go-admin/app/user_auth/repository"
	"go-admin/common/auth/middleware"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	// ErrUserNotFound 用户不存在
	ErrUserNotFound = errors.New("用户不存在")
	// ErrDuplicatePhone 手机号已存在
	ErrDuplicatePhone = errors.New("手机号已存在")
	// ErrDuplicateEmail 邮箱已存在
	ErrDuplicateEmail = errors.New("邮箱已存在")
	// ErrVersionConflict 版本冲突（乐观锁）
	ErrVersionConflict = errors.New("数据已被修改，请刷新后重试")
	// ErrInvalidStatus 无效状态值
	ErrInvalidStatus = errors.New("无效状态值，仅允许 0 或 1")
)

// BizUserService C端用户业务逻辑层
type BizUserService struct {
	repo repository.BizUserRepository
	db   *gorm.DB
}

// NewBizUserService 创建 BizUserService 实例
func NewBizUserService(db *gorm.DB, repo repository.BizUserRepository) *BizUserService {
	return &BizUserService{
		repo: repo,
		db:   db,
	}
}

// CreateBizUser 创建C端用户
// - 密码可选，有值则 bcrypt 哈希
// - ID 使用雪花算法（Model BeforeCreate 钩子处理）
// - Phone 和 Email 至少填一个
func (s *BizUserService) CreateBizUser(req *dto.CreateBizUserRequest) (*model.BizUser, error) {
	// 验证手机号和邮箱至少填一个
	if req.Phone == "" && req.Email == "" {
		return nil, errors.New("手机号和邮箱至少填写一个")
	}
	
	user := &model.BizUser{
		TenantID: req.TenantID,
		Nickname: req.Nickname,
		Status:   1,
	}
	
	// Phone 非空时设置
	if req.Phone != "" {
		user.Phone = &req.Phone
	}
	
	// Email 非空时设置
	if req.Email != "" {
		user.Email = &req.Email
	}

	// 密码可选，有值则 bcrypt 哈希
	if req.Password != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		user.Password = string(hashed)
	}

	// Token 过期时间（可选）
	user.AccessTokenExpireMinutes = req.AccessTokenExpireMinutes
	user.RefreshTokenExpireMinutes = req.RefreshTokenExpireMinutes

	if err := s.repo.Create(s.db, user); err != nil {
		// 唯一约束冲突判断
		if isDuplicateError(err) {
			errStr := err.Error()
			if contains(errStr, "email") || contains(errStr, "uk_tenant_email") {
				return nil, ErrDuplicateEmail
			}
			return nil, ErrDuplicatePhone
		}
		return nil, err
	}

	return user, nil
}

// GetBizUserByTenantPhone 根据租户ID和手机号获取C端用户
func (s *BizUserService) GetBizUserByTenantPhone(tenantID int64, phone string) (*model.BizUser, error) {
	user, err := s.repo.FindByTenantPhone(s.db, tenantID, phone)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

// GetBizUserByTenantEmail 根据租户ID和邮箱获取C端用户
func (s *BizUserService) GetBizUserByTenantEmail(tenantID int64, email string) (*model.BizUser, error) {
	user, err := s.repo.FindByTenantEmail(s.db, tenantID, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

// GetBizUserByTenantAccount 根据租户ID和账号（手机号或邮箱）获取C端用户
func (s *BizUserService) GetBizUserByTenantAccount(tenantID int64, account string) (*model.BizUser, error) {
	user, err := s.repo.FindByTenantAccount(s.db, tenantID, account)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

// CreateBizUserByAccount 按账号（手机号或邮箱）创建C端用户
func (s *BizUserService) CreateBizUserByAccount(req *dto.CreateBizUserByAccountRequest) (*model.BizUser, error) {
	user := &model.BizUser{
		TenantID: req.TenantID,
		Nickname: req.Nickname,
		Status:   1,
	}
	
	// 设置手机号（可选）
	if req.Phone != "" {
		user.Phone = &req.Phone
	}
	
	// 设置邮箱（可选）
	if req.Email != "" {
		user.Email = &req.Email
	}

	if err := s.repo.Create(s.db, user); err != nil {
		// 唯一约束冲突判断
		if isDuplicateError(err) {
			// 根据错误信息判断是手机号还是邮箱冲突
			errStr := err.Error()
			if contains(errStr, "email") || contains(errStr, "uk_tenant_email") {
				return nil, ErrDuplicateEmail
			}
			return nil, ErrDuplicatePhone
		}
		return nil, err
	}

	return user, nil
}

// contains 检查字符串是否包含子串（不区分大小写）
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstr(s, substr)))
}

func findSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// GetBizUser 根据 ID 获取C端用户
func (s *BizUserService) GetBizUser(id int64) (*model.BizUser, error) {
	user, err := s.repo.FindByID(s.db, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

// ListBizUsers 分页列表（按 tenant_id 隔离）
func (s *BizUserService) ListBizUsers(tenantID int64, page, pageSize int, phone string) ([]model.BizUser, int64, error) {
	return s.repo.ListByTenant(s.db, tenantID, page, pageSize, phone)
}

// BizUserWithStatus FR-21: 列表项 = 用户 + 最新状态日志详情（禁用/解禁原因、时间、操作人）
type BizUserWithStatus struct {
	model.BizUser
	DisabledReason     string `json:"disabled_reason"`       // 最近禁用的管理端内部原因（logs 最新）
	DisabledReasonUser string `json:"disabled_reason_user"`  // 最近禁用的用户展示原因（logs 最新）
	EnabledAt          string `json:"enabled_at"`            // 最近解禁时间（logs 最新）
	EnabledBy          string `json:"enabled_by"`            // 最近解禁操作人（logs 最新）
}

// ListBizUsersWithStatus 分页列表 + FR-21 状态日志详情（单一事实源 = logs 最新一条）
func (s *BizUserService) ListBizUsersWithStatus(tenantID int64, page, pageSize int, phone string) ([]BizUserWithStatus, int64, error) {
	list, total, err := s.repo.ListByTenant(s.db, tenantID, page, pageSize, phone)
	if err != nil {
		return nil, 0, err
	}

	items := make([]BizUserWithStatus, 0, len(list))
	if len(list) == 0 {
		return items, total, nil
	}

	ids := make([]int64, 0, len(list))
	for i := range list {
		ids = append(ids, list[i].ID)
	}
	var logs []model.BizUserStatusLog
	if err := s.db.Where("user_id IN ?", ids).Order("created_at ASC").Find(&logs).Error; err == nil {
		latest := map[int64]model.BizUserStatusLog{}
		for _, lg := range logs {
			latest[lg.UserID] = lg // ASC 遍历，后写覆盖 = 最新
		}
		for i := range list {
			it := BizUserWithStatus{BizUser: list[i]}
			if lg, ok := latest[list[i].ID]; ok {
				if lg.Action == 0 {
					it.DisabledReason = lg.AdminReason
					it.DisabledReasonUser = lg.UserReason
				} else {
					it.EnabledAt = lg.CreatedAt.Format("2006-01-02 15:04:05")
					it.EnabledBy = strconv.FormatInt(lg.OperatorID, 10)
				}
			}
			items = append(items, it)
		}
	} else {
		// 日志查询失败降级：仅返回用户基础数据
		for i := range list {
			items = append(items, BizUserWithStatus{BizUser: list[i]})
		}
	}
	return items, total, nil
}

// UpdateBizUser 更新C端用户（乐观锁）
func (s *BizUserService) UpdateBizUser(id int64, req *dto.UpdateBizUserRequest) (*model.BizUser, error) {
	user, err := s.repo.FindByID(s.db, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	// 乐观锁校验
	if user.Version != req.Version {
		return nil, ErrVersionConflict
	}

	// 更新字段
	if req.Phone != "" {
		user.Phone = &req.Phone
	}
	if req.Email != "" {
		user.Email = &req.Email
	}
	if req.Nickname != "" {
		user.Nickname = req.Nickname
	}
	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}
	// Token 过期时间（可选，传 null 表示清除自定义值）
	user.AccessTokenExpireMinutes = req.AccessTokenExpireMinutes
	user.RefreshTokenExpireMinutes = req.RefreshTokenExpireMinutes

	if err := s.repo.Update(s.db, user); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrVersionConflict
		}
		return nil, err
	}

	return user, nil
}

// DeleteBizUser 软删除C端用户
func (s *BizUserService) DeleteBizUser(id int64) error {
	err := s.repo.Delete(s.db, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}
	return nil
}

// ResetPassword 重置密码
// 生成 8 位随机字符串（含大小写字母+数字），bcrypt 存储，明文返回
func (s *BizUserService) ResetPassword(id int64) (string, error) {
	user, err := s.repo.FindByID(s.db, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrUserNotFound
		}
		return "", err
	}

	// 生成 8 位随机密码
	plainPassword, err := generateRandomPassword(8)
	if err != nil {
		return "", err
	}

	// bcrypt 哈希存储
	hashed, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	user.Password = string(hashed)
	if err := s.repo.Update(s.db, user); err != nil {
		return "", err
	}

	return plainPassword, nil
}

// ForceLogout 强制登出（递增 token_version + 失效缓存）
func (s *BizUserService) ForceLogout(id int64) error {
	if err := s.repo.IncrementTokenVersion(s.db, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}
	middleware.InvalidateBizUserCache(id)
	return nil
}

// CheckBizUser 实现 middleware.BizUserChecker 接口
// 查询C端用户的状态和 token 版本号
func (s *BizUserService) CheckBizUser(userID int64) (status int, tokenVersion int, err error) {
	user, err := s.repo.FindByID(s.db, userID)
	if err != nil {
		return 0, 0, err
	}
	return user.Status, user.TokenVersion, nil
}

// ToggleStatus 切换启用/禁用状态
func (s *BizUserService) ToggleStatus(id int64, status int, adminReason, userReason string, operator int64) error {
	if status != 0 && status != 1 {
		return ErrInvalidStatus
	}

	user, err := s.repo.FindByID(s.db, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	user.Status = status
	if err := s.repo.Update(s.db, user); err != nil {
		return err
	}

	// FR-21：写操作历史（失败不影响主操作）
	logRow := model.BizUserStatusLog{
		TenantID:    user.TenantID,
		UserID:      user.ID,
		Action:      status, // 0=禁用 1=启用（与 action 语义一致）
		AdminReason: adminReason,
		UserReason:  userReason,
		OperatorID:  operator,
	}
	if err := s.db.Create(&logRow).Error; err != nil {
		// 历史写失败不影响主操作（记日志继续）
		log.Printf("[user_auth] ToggleStatus write status log failed: %v", err)
	}
	return nil
}

// generateRandomPassword 生成指定长度的随机密码（含大小写字母+数字）
func generateRandomPassword(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		result[i] = charset[idx.Int64()]
	}
	return string(result), nil
}

// isDuplicateError 判断是否为唯一约束冲突错误
func isDuplicateError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := err.Error()
	// MySQL: Duplicate entry
	// PostgreSQL: duplicate key value violates unique constraint
	// SQLite: UNIQUE constraint failed
	return containsStr(errMsg, "Duplicate entry") ||
		containsStr(errMsg, "duplicate key") ||
		containsStr(errMsg, "UNIQUE constraint failed")
}

// containsStr 字符串包含判断
func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstr(s, substr))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
