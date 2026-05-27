package service

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"syncspace/backend/internal/auth"
	"syncspace/backend/internal/models"
	"syncspace/backend/internal/store"
)

func setupTestService(t *testing.T) *Service {
	t.Helper()
	db := filepath.Join(t.TempDir(), "test.db")
	st, err := store.Open(db)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	return New(st, t.TempDir())
}

func registerUser(t *testing.T, svc *Service, email, password, name, role string) models.User {
	t.Helper()
	u, err := svc.Register(context.Background(), models.RegisterRequest{
		Email:    email,
		Password: password,
		Name:     name,
		Role:     role,
	})
	if err != nil {
		t.Fatalf("registerUser(%s): %v", email, err)
	}
	return u
}

func activateUser(t *testing.T, svc *Service, adminID, userID int64) {
	t.Helper()
	if err := svc.ActivateUser(context.Background(), adminID, userID); err != nil {
		t.Fatalf("activateUser(%d): %v", userID, err)
	}
}

// ==================== Register Tests ====================

func TestRegisterSuccess(t *testing.T) {
	svc := setupTestService(t)
	ctx := context.Background()

	u, err := svc.Register(ctx, models.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
		Role:     "user",
	})
	if err != nil {
		t.Fatal(err)
	}
	if u.Email != "test@example.com" {
		t.Fatalf("expected email test@example.com, got %s", u.Email)
	}
	if u.Name != "Test User" {
		t.Fatalf("expected name Test User, got %s", u.Name)
	}
	if u.Role != "user" {
		t.Fatalf("expected role user, got %s", u.Role)
	}
	if u.Status != "pending" {
		t.Fatalf("expected status pending, got %s", u.Status)
	}
}

func TestRegisterCreatorRole(t *testing.T) {
	svc := setupTestService(t)
	u, err := svc.Register(context.Background(), models.RegisterRequest{
		Email:    "creator@example.com",
		Password: "password123",
		Name:     "Creator",
		Role:     "creator",
	})
	if err != nil {
		t.Fatal(err)
	}
	if u.Role != "creator" {
		t.Fatalf("expected role creator, got %s", u.Role)
	}
}

func TestRegisterEmailNormalization(t *testing.T) {
	svc := setupTestService(t)
	u, err := svc.Register(context.Background(), models.RegisterRequest{
		Email:    "  TEST@Example.COM  ",
		Password: "password123",
		Name:     "Test",
		Role:     "user",
	})
	if err != nil {
		t.Fatal(err)
	}
	if u.Email != "test@example.com" {
		t.Fatalf("expected normalized email test@example.com, got %s", u.Email)
	}
}

func TestRegisterNameTrimming(t *testing.T) {
	svc := setupTestService(t)
	u, err := svc.Register(context.Background(), models.RegisterRequest{
		Email:    "trim@example.com",
		Password: "password123",
		Name:     "  Trimmed Name  ",
		Role:     "user",
	})
	if err != nil {
		t.Fatal(err)
	}
	if u.Name != "Trimmed Name" {
		t.Fatalf("expected trimmed name 'Trimmed Name', got '%s'", u.Name)
	}
}

func TestRegisterEmptyEmail(t *testing.T) {
	svc := setupTestService(t)
	_, err := svc.Register(context.Background(), models.RegisterRequest{
		Email:    "",
		Password: "password123",
		Name:     "Test",
		Role:     "user",
	})
	if err == nil {
		t.Fatal("expected error for empty email")
	}
}

func TestRegisterEmptyPassword(t *testing.T) {
	svc := setupTestService(t)
	_, err := svc.Register(context.Background(), models.RegisterRequest{
		Email:    "test@example.com",
		Password: "",
		Name:     "Test",
		Role:     "user",
	})
	if err == nil {
		t.Fatal("expected error for empty password")
	}
}

func TestRegisterShortPassword(t *testing.T) {
	svc := setupTestService(t)
	_, err := svc.Register(context.Background(), models.RegisterRequest{
		Email:    "test@example.com",
		Password: "short",
		Name:     "Test",
		Role:     "user",
	})
	if err == nil {
		t.Fatal("expected error for short password")
	}
}

func TestRegisterEmptyName(t *testing.T) {
	svc := setupTestService(t)
	_, err := svc.Register(context.Background(), models.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "",
		Role:     "user",
	})
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestRegisterInvalidRole(t *testing.T) {
	svc := setupTestService(t)
	_, err := svc.Register(context.Background(), models.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test",
		Role:     "superadmin",
	})
	if err == nil {
		t.Fatal("expected error for invalid role")
	}
}

func TestRegisterDuplicateEmail(t *testing.T) {
	svc := setupTestService(t)
	registerUser(t, svc, "dup@example.com", "password123", "First", "user")
	_, err := svc.Register(context.Background(), models.RegisterRequest{
		Email:    "dup@example.com",
		Password: "password456",
		Name:     "Second",
		Role:     "user",
	})
	if err == nil {
		t.Fatal("expected error for duplicate email")
	}
}

// ==================== Login Tests ====================

func TestLoginSuccess(t *testing.T) {
	svc := setupTestService(t)
	ctx := context.Background()

	admin, _ := svc.GetUserByID(ctx, 1)
	registerUser(t, svc, "login@example.com", "password123", "Login User", "user")
	user, _ := svc.store.GetUserByEmail(ctx, "login@example.com")
	activateUser(t, svc, admin.ID, user.ID)

	token, u, err := svc.Login(ctx, models.LoginRequest{
		Email:    "login@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatal(err)
	}
	if token.AccessToken == "" {
		t.Fatal("expected non-empty token")
	}
	if token.ExpiresIn != 86400 {
		t.Fatalf("expected expires_in 86400, got %d", token.ExpiresIn)
	}
	if u.Email != "login@example.com" {
		t.Fatalf("expected email login@example.com, got %s", u.Email)
	}
}

func TestLoginWrongPassword(t *testing.T) {
	svc := setupTestService(t)
	ctx := context.Background()

	admin, _ := svc.GetUserByID(ctx, 1)
	registerUser(t, svc, "wrong@example.com", "password123", "Wrong", "user")
	user, _ := svc.store.GetUserByEmail(ctx, "wrong@example.com")
	activateUser(t, svc, admin.ID, user.ID)

	_, _, err := svc.Login(ctx, models.LoginRequest{
		Email:    "wrong@example.com",
		Password: "wrongpassword",
	})
	if err == nil {
		t.Fatal("expected error for wrong password")
	}
	if !strings.Contains(err.Error(), "invalid email or password") {
		t.Fatalf("expected 'invalid email or password', got: %s", err.Error())
	}
}

func TestLoginNonexistentUser(t *testing.T) {
	svc := setupTestService(t)
	_, _, err := svc.Login(context.Background(), models.LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "password123",
	})
	if err == nil {
		t.Fatal("expected error for nonexistent user")
	}
	if !strings.Contains(err.Error(), "invalid email or password") {
		t.Fatalf("expected 'invalid email or password', got: %s", err.Error())
	}
}

func TestLoginPendingAccount(t *testing.T) {
	svc := setupTestService(t)
	registerUser(t, svc, "pending@example.com", "password123", "Pending", "user")

	_, _, err := svc.Login(context.Background(), models.LoginRequest{
		Email:    "pending@example.com",
		Password: "password123",
	})
	if err == nil {
		t.Fatal("expected error for pending account")
	}
	if !strings.Contains(err.Error(), "pending approval") {
		t.Fatalf("expected 'pending approval', got: %s", err.Error())
	}
}

func TestLoginSuspendedAccount(t *testing.T) {
	svc := setupTestService(t)
	ctx := context.Background()

	admin, _ := svc.GetUserByID(ctx, 1)
	registerUser(t, svc, "suspended@example.com", "password123", "Suspended", "user")
	user, _ := svc.store.GetUserByEmail(ctx, "suspended@example.com")
	activateUser(t, svc, admin.ID, user.ID)
	svc.SuspendUser(ctx, admin.ID, user.ID)

	_, _, err := svc.Login(ctx, models.LoginRequest{
		Email:    "suspended@example.com",
		Password: "password123",
	})
	if err == nil {
		t.Fatal("expected error for suspended account")
	}
	if !strings.Contains(err.Error(), "suspended") {
		t.Fatalf("expected 'suspended', got: %s", err.Error())
	}
}

func TestLoginEmptyCredentials(t *testing.T) {
	svc := setupTestService(t)
	_, _, err := svc.Login(context.Background(), models.LoginRequest{
		Email:    "",
		Password: "",
	})
	if err == nil {
		t.Fatal("expected error for empty credentials")
	}
}

func TestLoginCaseInsensitive(t *testing.T) {
	svc := setupTestService(t)
	ctx := context.Background()

	admin, _ := svc.GetUserByID(ctx, 1)
	registerUser(t, svc, "case@example.com", "password123", "Case", "user")
	user, _ := svc.store.GetUserByEmail(ctx, "case@example.com")
	activateUser(t, svc, admin.ID, user.ID)

	_, _, err := svc.Login(ctx, models.LoginRequest{
		Email:    "CASE@EXAMPLE.COM",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("login should be case-insensitive: %v", err)
	}
}

// ==================== Token Validation ====================

func TestTokenIsValid(t *testing.T) {
	svc := setupTestService(t)
	ctx := context.Background()

	admin, _ := svc.GetUserByID(ctx, 1)
	registerUser(t, svc, "token@example.com", "password123", "Token", "user")
	user, _ := svc.store.GetUserByEmail(ctx, "token@example.com")
	activateUser(t, svc, admin.ID, user.ID)

	token, _, err := svc.Login(ctx, models.LoginRequest{
		Email:    "token@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatal(err)
	}

	claims, err := auth.ValidateToken(token.AccessToken)
	if err != nil {
		t.Fatalf("token should be valid: %v", err)
	}
	if claims.Email != "token@example.com" {
		t.Fatalf("expected email token@example.com, got %s", claims.Email)
	}
	if claims.Role != "user" {
		t.Fatalf("expected role user, got %s", claims.Role)
	}
}

// ==================== GetUserByID ====================

func TestGetUserByID(t *testing.T) {
	svc := setupTestService(t)
	u := registerUser(t, svc, "getbyid@example.com", "password123", "GetByID", "user")

	got, err := svc.GetUserByID(context.Background(), u.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Email != "getbyid@example.com" {
		t.Fatalf("expected email getbyid@example.com, got %s", got.Email)
	}
}

func TestGetUserByIDNotFound(t *testing.T) {
	svc := setupTestService(t)
	_, err := svc.GetUserByID(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for non-existent user")
	}
}

// ==================== ListUsers ====================

func TestListUsers(t *testing.T) {
	svc := setupTestService(t)
	registerUser(t, svc, "list1@example.com", "password123", "List1", "user")
	registerUser(t, svc, "list2@example.com", "password123", "List2", "creator")

	users, err := svc.ListUsers(context.Background(), "", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(users) < 3 {
		t.Fatalf("expected at least 3 users (admin + 2 registered), got %d", len(users))
	}
}

func TestListUsersByRole(t *testing.T) {
	svc := setupTestService(t)
	registerUser(t, svc, "filterrole@example.com", "password123", "FilterRole", "creator")

	users, err := svc.ListUsers(context.Background(), "creator", "")
	if err != nil {
		t.Fatal(err)
	}
	for _, u := range users {
		if u.Role != "creator" {
			t.Fatalf("expected role creator, got %s", u.Role)
		}
	}
}

func TestListUsersByStatus(t *testing.T) {
	svc := setupTestService(t)
	registerUser(t, svc, "filterstatus@example.com", "password123", "FilterStatus", "user")

	users, err := svc.ListUsers(context.Background(), "", "pending")
	if err != nil {
		t.Fatal(err)
	}
	for _, u := range users {
		if u.Status != "pending" {
			t.Fatalf("expected status pending, got %s", u.Status)
		}
	}
}
