// Code generated by MockGen. DO NOT EDIT.
// Source: internal/handlers/auth.go
//
// Generated by this command:
//
//	mockgen -source=internal/handlers/auth.go -destination=internal/handlers/mocks/auth.go -package=handlers_mocks
//

// Package handlers_mocks is a generated GoMock package.
package handlers_mocks

import (
	context "context"
	reflect "reflect"

	auth "github.com/OutOfStack/game-library-auth/internal/auth"
	database "github.com/OutOfStack/game-library-auth/internal/database"
	jwt "github.com/golang-jwt/jwt/v4"
	gomock "go.uber.org/mock/gomock"
	idtoken "google.golang.org/api/idtoken"
)

// MockUserRepo is a mock of UserRepo interface.
type MockUserRepo struct {
	ctrl     *gomock.Controller
	recorder *MockUserRepoMockRecorder
	isgomock struct{}
}

// MockUserRepoMockRecorder is the mock recorder for MockUserRepo.
type MockUserRepoMockRecorder struct {
	mock *MockUserRepo
}

// NewMockUserRepo creates a new mock instance.
func NewMockUserRepo(ctrl *gomock.Controller) *MockUserRepo {
	mock := &MockUserRepo{ctrl: ctrl}
	mock.recorder = &MockUserRepoMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockUserRepo) EXPECT() *MockUserRepoMockRecorder {
	return m.recorder
}

// CheckUserExists mocks base method.
func (m *MockUserRepo) CheckUserExists(ctx context.Context, name string, role database.Role) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CheckUserExists", ctx, name, role)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CheckUserExists indicates an expected call of CheckUserExists.
func (mr *MockUserRepoMockRecorder) CheckUserExists(ctx, name, role any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CheckUserExists", reflect.TypeOf((*MockUserRepo)(nil).CheckUserExists), ctx, name, role)
}

// CreateUser mocks base method.
func (m *MockUserRepo) CreateUser(ctx context.Context, user database.User) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateUser", ctx, user)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateUser indicates an expected call of CreateUser.
func (mr *MockUserRepoMockRecorder) CreateUser(ctx, user any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateUser", reflect.TypeOf((*MockUserRepo)(nil).CreateUser), ctx, user)
}

// DeleteUser mocks base method.
func (m *MockUserRepo) DeleteUser(ctx context.Context, userID string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteUser", ctx, userID)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteUser indicates an expected call of DeleteUser.
func (mr *MockUserRepoMockRecorder) DeleteUser(ctx, userID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteUser", reflect.TypeOf((*MockUserRepo)(nil).DeleteUser), ctx, userID)
}

// GetUserByID mocks base method.
func (m *MockUserRepo) GetUserByID(ctx context.Context, userID string) (database.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserByID", ctx, userID)
	ret0, _ := ret[0].(database.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUserByID indicates an expected call of GetUserByID.
func (mr *MockUserRepoMockRecorder) GetUserByID(ctx, userID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserByID", reflect.TypeOf((*MockUserRepo)(nil).GetUserByID), ctx, userID)
}

// GetUserByOAuth mocks base method.
func (m *MockUserRepo) GetUserByOAuth(ctx context.Context, provider, oauthID string) (database.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserByOAuth", ctx, provider, oauthID)
	ret0, _ := ret[0].(database.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUserByOAuth indicates an expected call of GetUserByOAuth.
func (mr *MockUserRepoMockRecorder) GetUserByOAuth(ctx, provider, oauthID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserByOAuth", reflect.TypeOf((*MockUserRepo)(nil).GetUserByOAuth), ctx, provider, oauthID)
}

// GetUserByUsername mocks base method.
func (m *MockUserRepo) GetUserByUsername(ctx context.Context, username string) (database.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserByUsername", ctx, username)
	ret0, _ := ret[0].(database.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUserByUsername indicates an expected call of GetUserByUsername.
func (mr *MockUserRepoMockRecorder) GetUserByUsername(ctx, username any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserByUsername", reflect.TypeOf((*MockUserRepo)(nil).GetUserByUsername), ctx, username)
}

// UpdateUser mocks base method.
func (m *MockUserRepo) UpdateUser(ctx context.Context, user database.User) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateUser", ctx, user)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateUser indicates an expected call of UpdateUser.
func (mr *MockUserRepoMockRecorder) UpdateUser(ctx, user any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateUser", reflect.TypeOf((*MockUserRepo)(nil).UpdateUser), ctx, user)
}

// MockAuth is a mock of Auth interface.
type MockAuth struct {
	ctrl     *gomock.Controller
	recorder *MockAuthMockRecorder
	isgomock struct{}
}

// MockAuthMockRecorder is the mock recorder for MockAuth.
type MockAuthMockRecorder struct {
	mock *MockAuth
}

// NewMockAuth creates a new mock instance.
func NewMockAuth(ctrl *gomock.Controller) *MockAuth {
	mock := &MockAuth{ctrl: ctrl}
	mock.recorder = &MockAuthMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockAuth) EXPECT() *MockAuthMockRecorder {
	return m.recorder
}

// CreateClaims mocks base method.
func (m *MockAuth) CreateClaims(user database.User) jwt.Claims {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateClaims", user)
	ret0, _ := ret[0].(jwt.Claims)
	return ret0
}

// CreateClaims indicates an expected call of CreateClaims.
func (mr *MockAuthMockRecorder) CreateClaims(user any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateClaims", reflect.TypeOf((*MockAuth)(nil).CreateClaims), user)
}

// GenerateToken mocks base method.
func (m *MockAuth) GenerateToken(claims jwt.Claims) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GenerateToken", claims)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GenerateToken indicates an expected call of GenerateToken.
func (mr *MockAuthMockRecorder) GenerateToken(claims any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GenerateToken", reflect.TypeOf((*MockAuth)(nil).GenerateToken), claims)
}

// ValidateToken mocks base method.
func (m *MockAuth) ValidateToken(tokenStr string) (auth.Claims, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ValidateToken", tokenStr)
	ret0, _ := ret[0].(auth.Claims)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ValidateToken indicates an expected call of ValidateToken.
func (mr *MockAuthMockRecorder) ValidateToken(tokenStr any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ValidateToken", reflect.TypeOf((*MockAuth)(nil).ValidateToken), tokenStr)
}

// MockGoogleTokenValidator is a mock of GoogleTokenValidator interface.
type MockGoogleTokenValidator struct {
	ctrl     *gomock.Controller
	recorder *MockGoogleTokenValidatorMockRecorder
	isgomock struct{}
}

// MockGoogleTokenValidatorMockRecorder is the mock recorder for MockGoogleTokenValidator.
type MockGoogleTokenValidatorMockRecorder struct {
	mock *MockGoogleTokenValidator
}

// NewMockGoogleTokenValidator creates a new mock instance.
func NewMockGoogleTokenValidator(ctrl *gomock.Controller) *MockGoogleTokenValidator {
	mock := &MockGoogleTokenValidator{ctrl: ctrl}
	mock.recorder = &MockGoogleTokenValidatorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockGoogleTokenValidator) EXPECT() *MockGoogleTokenValidatorMockRecorder {
	return m.recorder
}

// Validate mocks base method.
func (m *MockGoogleTokenValidator) Validate(ctx context.Context, idToken, audience string) (*idtoken.Payload, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Validate", ctx, idToken, audience)
	ret0, _ := ret[0].(*idtoken.Payload)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Validate indicates an expected call of Validate.
func (mr *MockGoogleTokenValidatorMockRecorder) Validate(ctx, idToken, audience any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Validate", reflect.TypeOf((*MockGoogleTokenValidator)(nil).Validate), ctx, idToken, audience)
}
