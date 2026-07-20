package service

import (
	"go-admin/common/auth/config"
	"go-admin/common/auth/errors"
	"go-admin/common/auth/model"
	"go-admin/common/auth/repository"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserService 用户业务逻辑层
type UserService struct {
	db       *gorm.DB
	cfg      *config.Config
	userRepo repository.UserRepository
	logger   OperationLogger
}

// NewUserService 创建 UserService 实例
func NewUserService(db *gorm.DB, cfg *config.Config, userRepo repository.UserRepository) *UserService {
	return &UserService{
		db:       db,
		cfg:      cfg,
		userRepo: userRepo,
		logger:   &noopLogger{},
	}
}

// SetLogger 设置操作日志记录器
func (s *UserService) SetLogger(logger OperationLogger) {
	s.logger = logger
}

// CreateUserRequest 创建用户请求参数
type CreateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

// UpdateUserRequest 更新用户请求参数
type UpdateUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Status   *int   `json:"status"`
	Version  int    `json:"version" binding:"required"` // 乐观锁版本号
}

// AssociateTenantRequest 关联用户到租户请求参数（支持批量）
type AssociateTenantRequest struct {
	TenantID  int64            `json:"tenant_id,string"`
	TenantIDs StringInt64Slice `json:"tenant_ids"`
}

// CreateUser 创建全局用户，密码 bcrypt 加密
func (s *UserService) CreateUser(req *CreateUserRequest) (*model.User, error) {
	// 检查 username 唯一性
	existing, err := s.userRepo.FindByUsername(s.db, req.Username)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.NewAuthError(errors.ErrDuplicateEntity, "用户名已存在")
	}

	// 密码 bcrypt 加密
	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Username: req.Username,
		Password: string(hashedPwd),
		Email:    req.Email,
		Phone:    req.Phone,
		Nickname: req.Nickname,
		Avatar:   req.Avatar,
		Status:   1,
		Version:  1,
	}

	if err := s.userRepo.Create(s.db, user); err != nil {
		return nil, err
	}

	s.logger.Log(0, "create_user", "user", user.ID, "创建用户: "+user.Username)
	return user, nil
}

// UpdateUser 更新用户信息（乐观锁）
func (s *UserService) UpdateUser(id int64, req *UpdateUserRequest) (*model.User, error) {
	user, err := s.userRepo.FindByID(s.db, id)
	if err != nil {
		return nil, err
	}

	// 乐观锁版本校验
	if user.Version != req.Version {
		return nil, errors.NewAuthError(errors.ErrDuplicateEntity, "数据已被其他操作修改，请刷新后重试")
	}

	// 更新字段
	if req.Username != "" {
		user.Username = req.Username
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}
	if req.Nickname != "" {
		user.Nickname = req.Nickname
	}
	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}
	if req.Status != nil {
		user.Status = *req.Status
	}

	if err := s.userRepo.Update(s.db, user); err != nil {
		return nil, err
	}

	s.logger.Log(0, "update_user", "user", user.ID, "更新用户: "+user.Username)
	return user, nil
}

// DeleteUser 软删除用户（保护最后一个 SUPER_ADMIN）
func (s *UserService) DeleteUser(id int64) error {
	// 检查是否是最后一个 SUPER_ADMIN
	if s.isLastSuperAdmin(id) {
		return errors.NewAuthError(errors.ErrProtectedEntity, "无法删除最后一个超级管理员")
	}
	if err := s.userRepo.SoftDelete(s.db, id); err != nil {
		return err
	}
	s.logger.Log(0, "delete_user", "user", id, "")
	return nil
}

// isLastSuperAdmin 检查该用户是否是最后一个 SUPER_ADMIN
func (s *UserService) isLastSuperAdmin(userID int64) bool {
	// 查该用户是否有 SUPER_ADMIN 角色
	var count int64
	s.db.Table("admin_user_role ur").
		Joins("JOIN admin_role r ON r.id = ur.role_id").
		Where("ur.user_id = ? AND r.role_type = 'SUPER_ADMIN'", userID).
		Count(&count)
	if count == 0 {
		return false
	}
	// 查是否还有其他 SUPER_ADMIN 用户
	var otherCount int64
	s.db.Table("admin_user_role ur").
		Joins("JOIN admin_role r ON r.id = ur.role_id").
		Where("ur.user_id != ? AND r.role_type = 'SUPER_ADMIN'", userID).
		Count(&otherCount)
	return otherCount == 0
}

// UserListResult 用户列表分页结果
type UserListResult struct {
	List     []model.User `json:"list"`
	Total    int64        `json:"total"`
	Page     int          `json:"page"`
	PageSize int          `json:"page_size"`
}

// ListUsers 通过 tenant_id 过滤查询用户列表（JOIN admin_user_tenant）
func (s *UserService) ListUsers(tenantID int64, page, pageSize int) (*UserListResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	params := repository.UserListParams{
		TenantID: tenantID,
		Page:     page,
		PageSize: pageSize,
	}

	list, total, err := s.userRepo.ListByTenantID(s.db, params)
	if err != nil {
		return nil, err
	}

	return &UserListResult{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// AssociateTenant 关联用户到租户
func (s *UserService) AssociateTenant(userID, tenantID int64) error {
	// 确认用户存在
	_, err := s.userRepo.FindByID(s.db, userID)
	if err != nil {
		return err
	}

	// 检查是否已关联
	existing, err := s.userRepo.FindUserTenant(s.db, userID, tenantID)
	if err != nil {
		return err
	}
	if existing != nil {
		return errors.NewAuthError(errors.ErrDuplicateEntity, "用户已关联该租户")
	}

	ut := &model.UserTenant{
		UserID:   userID,
		TenantID: tenantID,
	}
	if err := s.userRepo.CreateUserTenant(s.db, ut); err != nil {
		return err
	}

	s.logger.Log(0, "associate_tenant", "user_tenant", ut.ID, "关联用户到租户")
	return nil
}

// DissociateTenant 解除用户-租户关联 + 级联删除该租户下角色绑定
func (s *UserService) DissociateTenant(userID, tenantID int64) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 先级联删除该用户在该租户下的 admin_user_role 记录
		if err := s.userRepo.DeleteUserRolesByTenant(tx, userID, tenantID); err != nil {
			return err
		}

		// 再删除用户-租户关联
		if err := s.userRepo.DeleteUserTenant(tx, userID, tenantID); err != nil {
			return err
		}

		s.logger.Log(0, "dissociate_tenant", "user_tenant", userID, "解除用户-租户关联（含级联删除角色绑定）")
		return nil
	})
}

// UserTenantInfo 用户关联租户简要信息（前端消费格式）
type UserTenantInfo struct {
	ID   int64  `json:"id,string"`
	Name string `json:"name"`
}

// ListUserTenants 查询用户已关联的租户列表（返回租户详情）
func (s *UserService) ListUserTenants(userID int64) ([]UserTenantInfo, error) {
	// 确认用户存在
	_, err := s.userRepo.FindByID(s.db, userID)
	if err != nil {
		return nil, err
	}

	// 查询关联记录
	uts, err := s.userRepo.ListUserTenants(s.db, userID)
	if err != nil {
		return nil, err
	}

	if len(uts) == 0 {
		return []UserTenantInfo{}, nil
	}

	// 批量查询租户详情
	tenantIDs := make([]int64, 0, len(uts))
	for _, ut := range uts {
		tenantIDs = append(tenantIDs, ut.TenantID)
	}

	var tenants []model.Tenant
	if err := s.db.Where("id IN ?", tenantIDs).Find(&tenants).Error; err != nil {
		return nil, err
	}

	// 组装结果
	result := make([]UserTenantInfo, 0, len(tenants))
	for _, t := range tenants {
		result = append(result, UserTenantInfo{ID: t.ID, Name: t.Name})
	}
	return result, nil
}

// ForceOffline 强制下线用户（设置状态为禁用）
func (s *UserService) ForceOffline(userID int64) error {
	user, err := s.userRepo.FindByID(s.db, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.NewAuthError(errors.ErrEntityNotFound, "用户不存在")
	}

	disabled := 0
	_, err = s.UpdateUser(userID, &UpdateUserRequest{
		Status:  &disabled,
		Version: user.Version,
	})
	if err != nil {
		return err
	}

	s.logger.Log(0, "force_offline", "user", userID, "强制下线用户")
	return nil
}
