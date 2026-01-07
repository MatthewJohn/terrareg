package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	analyticscmd "github.com/matthewjohn/terrareg/terrareg-go/internal/application/command/analytics"
)

// MockAnalyticsRepository is a mock for AnalyticsRepository
type MockAnalyticsRepository struct {
	mock.Mock
}

// RecordDownload mocks the RecordDownload method
func (m *MockAnalyticsRepository) RecordDownload(ctx context.Context, event analyticscmd.AnalyticsEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

// RecordProviderDownload mocks the RecordProviderDownload method
func (m *MockAnalyticsRepository) RecordProviderDownload(ctx context.Context, event analyticscmd.ProviderDownloadEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

// GetDownloadStats mocks the GetDownloadStats method
func (m *MockAnalyticsRepository) GetDownloadStats(ctx context.Context, namespace, module, provider string) (*analyticscmd.DownloadStats, error) {
	args := m.Called(ctx, namespace, module, provider)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*analyticscmd.DownloadStats), args.Error(1)
}

// GetDownloadsByVersionID mocks the GetDownloadsByVersionID method
func (m *MockAnalyticsRepository) GetDownloadsByVersionID(ctx context.Context, moduleVersionID int) (int, error) {
	args := m.Called(ctx, moduleVersionID)
	return args.Int(0), args.Error(1)
}

// GetMostRecentlyPublished mocks the GetMostRecentlyPublished method
func (m *MockAnalyticsRepository) GetMostRecentlyPublished(ctx context.Context) (*analyticscmd.ModuleVersionInfo, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*analyticscmd.ModuleVersionInfo), args.Error(1)
}

// GetMostDownloadedThisWeek mocks the GetMostDownloadedThisWeek method
func (m *MockAnalyticsRepository) GetMostDownloadedThisWeek(ctx context.Context) (*analyticscmd.ModuleProviderInfo, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*analyticscmd.ModuleProviderInfo), args.Error(1)
}

// GetModuleProviderID mocks the GetModuleProviderID method
func (m *MockAnalyticsRepository) GetModuleProviderID(ctx context.Context, namespace, module, provider string) (int, error) {
	args := m.Called(ctx, namespace, module, provider)
	return args.Int(0), args.Error(1)
}

// GetLatestTokenVersions mocks the GetLatestTokenVersions method
func (m *MockAnalyticsRepository) GetLatestTokenVersions(ctx context.Context, moduleProviderID int) (map[string]analyticscmd.TokenVersionInfo, error) {
	args := m.Called(ctx, moduleProviderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]analyticscmd.TokenVersionInfo), args.Error(1)
}
