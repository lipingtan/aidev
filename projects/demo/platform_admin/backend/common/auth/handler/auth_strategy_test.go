package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-admin/common/auth/strategy"
)

// TestLoginHandler_NoGrantType 不传 grant_type 时行为与之前完全一致
func TestLoginHandler_NoGrantType(t *testing.T) {
	db := setupAuthTestDB(t)
	r, _ := setupAuthRouter(db)

	user := createAuthUser(t, db, "admin", "password123")
	tenant := createAuthTenant(t, db, "corp_a", "公司A")
	createAuthUserTenant(t, db, user.ID, tenant.ID)

	body := map[string]interface{}{
		"username": "admin",
		"password": "password123",
	}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望状态码 200，实际: %d, body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Code int `json:"code"`
		Data struct {
			TokenType   string              `json:"token_type"`
			AccessToken string              `json:"access_token"`
			Tenants     []strategy.TenantInfo `json:"tenants"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.Code != 0 {
		t.Fatalf("期望 code=0，实际: %d", resp.Code)
	}
	if resp.Data.TokenType != "access" {
		t.Fatalf("单租户期望 token_type=access，实际: %s", resp.Data.TokenType)
	}
	if resp.Data.AccessToken == "" {
		t.Fatal("access_token 不应为空")
	}
}

// TestLoginHandler_GrantTypePassword 显式传 grant_type=password 等效于不传
func TestLoginHandler_GrantTypePassword(t *testing.T) {
	db := setupAuthTestDB(t)
	r, _ := setupAuthRouter(db)

	user := createAuthUser(t, db, "admin", "password123")
	tenant := createAuthTenant(t, db, "corp_a", "公司A")
	createAuthUserTenant(t, db, user.ID, tenant.ID)

	body := map[string]interface{}{
		"username":   "admin",
		"password":   "password123",
		"grant_type": "password",
	}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望状态码 200，实际: %d, body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Code int `json:"code"`
		Data struct {
			TokenType   string `json:"token_type"`
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.Code != 0 {
		t.Fatalf("期望 code=0，实际: %d", resp.Code)
	}
	if resp.Data.TokenType != "access" {
		t.Fatalf("单租户期望 token_type=access，实际: %s", resp.Data.TokenType)
	}
}

// TestLoginHandler_GrantTypeUnknown 未注册的 grant_type 返回 400
func TestLoginHandler_GrantTypeUnknown(t *testing.T) {
	db := setupAuthTestDB(t)
	r, _ := setupAuthRouter(db)

	createAuthUser(t, db, "admin", "password123")

	body := map[string]interface{}{
		"username":   "admin",
		"password":   "password123",
		"grant_type": "unknown",
	}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("期望状态码 400，实际: %d, body: %s", w.Code, w.Body.String())
	}

	var resp Response
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != 40000 {
		t.Fatalf("期望 code=40000，实际: %d", resp.Code)
	}
}

// TestPasswordStrategy_Authenticate_Success 密码策略认证成功
func TestPasswordStrategy_Authenticate_Success(t *testing.T) {
	db := setupAuthTestDB(t)
	r, _ := setupAuthRouter(db)

	user := createAuthUser(t, db, "admin", "password123")
	tenant := createAuthTenant(t, db, "corp_a", "公司A")
	createAuthUserTenant(t, db, user.ID, tenant.ID)

	// 通过 grant_type=password 走策略路由（虽然实际逻辑走 handlePasswordLogin，
	// 这里验证 grant_type=password 时响应与不传一致）
	body := map[string]interface{}{
		"username":   "admin",
		"password":   "password123",
		"grant_type": "password",
	}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望状态码 200，实际: %d, body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Code int `json:"code"`
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != 0 {
		t.Fatalf("期望 code=0，实际: %d", resp.Code)
	}
	if resp.Data.AccessToken == "" {
		t.Fatal("access_token 不应为空")
	}
}

// TestPasswordStrategy_Authenticate_WrongPassword 密码策略认证失败
func TestPasswordStrategy_Authenticate_WrongPassword(t *testing.T) {
	db := setupAuthTestDB(t)
	r, _ := setupAuthRouter(db)

	user := createAuthUser(t, db, "admin", "password123")
	tenant := createAuthTenant(t, db, "corp_a", "公司A")
	createAuthUserTenant(t, db, user.ID, tenant.ID)

	body := map[string]interface{}{
		"username":   "admin",
		"password":   "wrong_password",
		"grant_type": "password",
	}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("期望状态码 401，实际: %d, body: %s", w.Code, w.Body.String())
	}
}
