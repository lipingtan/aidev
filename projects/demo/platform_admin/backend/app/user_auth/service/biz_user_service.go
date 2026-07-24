package service

import (
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
func (s *BizUserService) CreateBizUser(req *dto.CreateBizUserRequest) (*model.BizUser, error) {
	user := &model.BizUser{
		TenantID: req.TenantID,
		Phone:    req.Phone,
		Nickname: req.Nickname,
		Status:   1,
	}

	// 密码可选，有值则 bcrypt 哈希
	if req.Password != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		user.Password = string(hashed)
	}

	if err := s.repo.Create(s.db, user); err != nil {
		// 唯一约束冲突判断
		if isDuplicateError(err) {
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
		user.Phone = req.Phone
	}
	if req.Nickname != "" {
		user.Nickname = req.Nickname
	}
	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}

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
func (s *BizUserService) ToggleStatus(id int64, status int) error {
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
