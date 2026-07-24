package service

import (
	"testing"

	"go-admin/app/user_auth/dto"
	"go-admin/app/user_auth/model"
	"go-admin/app/user_auth/repository"

	"golang.org/x/crypto/bcrypt"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// setupTestDB 创建 SQLite 内存数据库并自动迁移
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("打开 SQLite 内存数据库失败: %v", err)
	}
	if err := db.AutoMigrate(&model.BizUser{}); err != nil {
		t.Fatalf("数据库迁移失败: %v", err)
	}
	return db
}

// newTestBizUserService 创建测试用 BizUserService
func newTestBizUserService(t *testing.T) (*BizUserService, *gorm.DB) {
	t.Helper()
	db := setupTestDB(t)
	repo := repository.NewBizUserRepository()
	svc := NewBizUserService(db, repo)
	return svc, db
}

// TestCreateBizUser_Success 创建C端用户成功，DB 有记录
func TestCreateBizUser_Success(t *testing.T) {
	svc, db := newTestBizUserService(t)

	req := &dto.CreateBizUserRequest{
		TenantID: 1001,
		Phone:    "13800138000",
		Password: "abc123",
		Nickname: "测试用户",
	}

	user, err := svc.CreateBizUser(req)
	if err != nil {
		t.Fatalf("CreateBizUser 应成功，但返回错误: %v", err)
	}

	if user.ID == 0 {
		t.Error("用户 ID 不应为 0（应由雪花算法生成）")
	}
	if user.Phone != "13800138000" {
		t.Errorf("Phone 应为 13800138000，实际: %s", user.Phone)
	}
	if user.Status != 1 {
		t.Errorf("默认状态应为 1，实际: %d", user.Status)
	}

	// 验证密码已 bcrypt 哈希
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte("abc123")); err != nil {
		t.Error("密码应已正确 bcrypt 哈希")
	}

	// 验证 DB 中有记录
	var count int64
	db.Model(&model.BizUser{}).Where("phone = ?", "13800138000").Count(&count)
	if count != 1 {
		t.Errorf("DB 中应有 1 条记录，实际: %d", count)
	}
}

// TestCreateBizUser_DuplicatePhone 重复手机号创建返回错误
func TestCreateBizUser_DuplicatePhone(t *testing.T) {
	svc, _ := newTestBizUserService(t)

	req := &dto.CreateBizUserRequest{
		TenantID: 1001,
		Phone:    "13800138000",
	}

	// 第一次创建成功
	_, err := svc.CreateBizUser(req)
	if err != nil {
		t.Fatalf("第一次创建应成功: %v", err)
	}

	// 第二次创建应失败（唯一约束）
	_, err = svc.CreateBizUser(req)
	if err == nil {
		t.Fatal("重复手机号创建应返回错误")
	}
	if err != ErrDuplicatePhone {
		t.Errorf("应返回 ErrDuplicatePhone，实际: %v", err)
	}
}

// TestCreateBizUser_NoPassword 密码可选，无密码时创建成功
func TestCreateBizUser_NoPassword(t *testing.T) {
	svc, _ := newTestBizUserService(t)

	req := &dto.CreateBizUserRequest{
		TenantID: 1001,
		Phone:    "13900139000",
	}

	user, err := svc.CreateBizUser(req)
	if err != nil {
		t.Fatalf("无密码创建应成功: %v", err)
	}
	if user.Password != "" {
		t.Error("未提供密码时，Password 字段应为空")
	}
}

// TestResetPassword_Success 重置密码返回明文 + bcrypt.Compare 通过
func TestResetPassword_Success(t *testing.T) {
	svc, db := newTestBizUserService(t)

	// 先创建用户
	req := &dto.CreateBizUserRequest{
		TenantID: 1001,
		Phone:    "13800138000",
	}
	user, err := svc.CreateBizUser(req)
	if err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}

	// 重置密码
	plainPwd, err := svc.ResetPassword(user.ID)
	if err != nil {
		t.Fatalf("ResetPassword 应成功: %v", err)
	}

	// 明文密码应为 8 位
	if len(plainPwd) != 8 {
		t.Errorf("密码应为 8 位，实际: %d", len(plainPwd))
	}

	// 从 DB 读取最新密码哈希
	var dbUser model.BizUser
	db.Where("id = ?", user.ID).First(&dbUser)

	// bcrypt.Compare 应通过
	if err := bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(plainPwd)); err != nil {
		t.Error("ResetPassword 后 bcrypt.Compare 应通过")
	}
}

// TestForceLogout_IncrementTokenVersion 强制登出 token_version 递增
func TestForceLogout_IncrementTokenVersion(t *testing.T) {
	svc, db := newTestBizUserService(t)

	// 先创建用户
	req := &dto.CreateBizUserRequest{
		TenantID: 1001,
		Phone:    "13800138000",
	}
	user, err := svc.CreateBizUser(req)
	if err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}

	// 记录初始 token_version
	var beforeUser model.BizUser
	db.Where("id = ?", user.ID).First(&beforeUser)
	initialVersion := beforeUser.TokenVersion

	// 强制登出
	err = svc.ForceLogout(user.ID)
	if err != nil {
		t.Fatalf("ForceLogout 应成功: %v", err)
	}

	// 验证 token_version 递增
	var afterUser model.BizUser
	db.Where("id = ?", user.ID).First(&afterUser)
	if afterUser.TokenVersion != initialVersion+1 {
		t.Errorf("token_version 应从 %d 递增到 %d，实际: %d",
			initialVersion, initialVersion+1, afterUser.TokenVersion)
	}
}

// TestToggleStatus_DisableUser 切换用户状态为禁用
func TestToggleStatus_DisableUser(t *testing.T) {
	svc, db := newTestBizUserService(t)

	// 先创建用户（默认 status=1）
	req := &dto.CreateBizUserRequest{
		TenantID: 1001,
		Phone:    "13800138000",
	}
	user, err := svc.CreateBizUser(req)
	if err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}

	// 禁用
	err = svc.ToggleStatus(user.ID, 0)
	if err != nil {
		t.Fatalf("ToggleStatus(0) 应成功: %v", err)
	}

	// 验证 DB 中 status=0
	var dbUser model.BizUser
	db.Where("id = ?", user.ID).First(&dbUser)
	if dbUser.Status != 0 {
		t.Errorf("status 应为 0，实际: %d", dbUser.Status)
	}
}

// TestToggleStatus_InvalidStatus 无效状态值应返回错误
func TestToggleStatus_InvalidStatus(t *testing.T) {
	svc, _ := newTestBizUserService(t)

	req := &dto.CreateBizUserRequest{
		TenantID: 1001,
		Phone:    "13800138000",
	}
	user, err := svc.CreateBizUser(req)
	if err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}

	err = svc.ToggleStatus(user.ID, 2)
	if err != ErrInvalidStatus {
		t.Errorf("无效状态值应返回 ErrInvalidStatus，实际: %v", err)
	}
}
