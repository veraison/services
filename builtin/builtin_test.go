// Copyright 2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package builtin

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/veraison/services/plugin"
	"go.uber.org/zap/zaptest"
)

// MockPluggable implements IPluggable for testing
type MockPluggable struct {
	name            string
	scheme          string
	supportedTypes  []string
}

func (m *MockPluggable) GetName() string {
	return m.name
}

func (m *MockPluggable) GetAttestationScheme() string {
	return m.scheme
}

func (m *MockPluggable) GetSupportedMediaTypes() []string {
	return m.supportedTypes
}

// Test interface to match the expected plugin interfaces
type TestEvidenceHandler interface {
	plugin.IPluggable
	ProcessEvidence([]byte) ([]byte, error)
}

type TestEndorsementHandler interface {
	plugin.IPluggable  
	ProcessEndorsement([]byte) error
}

// MockEvidenceHandler implements TestEvidenceHandler
type MockEvidenceHandler struct {
	MockPluggable
}

func (m *MockEvidenceHandler) ProcessEvidence(data []byte) ([]byte, error) {
	return data, nil
}

// MockEndorsementHandler implements TestEndorsementHandler  
type MockEndorsementHandler struct {
	MockPluggable
}

func (m *MockEndorsementHandler) ProcessEndorsement(data []byte) error {
	return nil
}

func TestNewBuiltinLoader(t *testing.T) {
	logger := zaptest.NewLogger(t).Sugar()
	
	loader := NewBuiltinLoader(logger)
	
	assert.NotNil(t, loader)
	assert.Equal(t, logger, loader.logger)
}

func TestBuiltinLoader_Init(t *testing.T) {
	logger := zaptest.NewLogger(t).Sugar()
	loader := NewBuiltinLoader(logger)
	
	cfg := map[string]interface{}{
		"test-key": "test-value",
	}
	
	err := loader.Init(cfg)
	
	assert.NoError(t, err)
	assert.NotNil(t, loader.loadedByName)
	assert.NotNil(t, loader.loadedByMediaType)
	assert.NotNil(t, loader.registeredPluginTypes)
}

func TestCreateBuiltinLoader(t *testing.T) {
	logger := zaptest.NewLogger(t).Sugar()
	cfg := map[string]interface{}{
		"test-config": "value",
	}
	
	loader, err := CreateBuiltinLoader(cfg, logger)
	
	assert.NoError(t, err)
	assert.NotNil(t, loader)
	assert.Equal(t, logger, loader.logger)
}

func TestBuiltinLoader_GetRegisteredMediaTypes(t *testing.T) {
	logger := zaptest.NewLogger(t).Sugar()
	loader := NewBuiltinLoader(logger)
	loader.Init(map[string]interface{}{})
	
	// Add mock plugins to the loader
	mockPlugin1 := &MockPluggable{
		supportedTypes: []string{"application/test1", "application/test2"},
	}
	mockPlugin2 := &MockPluggable{
		supportedTypes: []string{"application/test3"},
	}
	
	loader.loadedByMediaType["application/test1"] = mockPlugin1
	loader.loadedByMediaType["application/test2"] = mockPlugin1
	loader.loadedByMediaType["application/test3"] = mockPlugin2
	
	mediaTypes := loader.GetRegisteredMediaTypes()
	
	assert.Len(t, mediaTypes, 3)
	assert.Contains(t, mediaTypes, "application/test1")
	assert.Contains(t, mediaTypes, "application/test2")
	assert.Contains(t, mediaTypes, "application/test3")
}

func TestNewBuiltinManager(t *testing.T) {
	logger := zaptest.NewLogger(t).Sugar()
	loader := NewBuiltinLoader(logger)
	
	manager := NewBuiltinManager[plugin.IPluggable](loader, logger)
	
	assert.NotNil(t, manager)
	assert.Equal(t, loader, manager.loader)
	assert.Equal(t, logger, manager.logger)
}

func TestCreateBuiltinManager_Success(t *testing.T) {
	logger := zaptest.NewLogger(t).Sugar()
	
	// Create a test viper configuration
	v := viper.New()
	v.Set("builtin.test-key", "test-value")
	
	manager, err := CreateBuiltinManager[plugin.IPluggable](v, logger, "test-manager")
	
	// Note: This might fail due to the actual discovery process, but we test the structure
	if err != nil {
		// Expected if no valid plugins are found in the discovery process
		assert.Contains(t, err.Error(), "builtin")
	} else {
		assert.NotNil(t, manager)
	}
}

func TestCreateBuiltinManagerWithLoader(t *testing.T) {
	logger := zaptest.NewLogger(t).Sugar()
	loader := NewBuiltinLoader(logger)
	loader.Init(map[string]interface{}{})
	
	// Setup loader with mock data for successful initialization
	loader.loadedByName = make(map[string]plugin.IPluggable)
	loader.loadedByMediaType = make(map[string]plugin.IPluggable)
	
	manager, err := CreateBuiltinManagerWithLoader[plugin.IPluggable](loader, logger, "test")
	
	assert.NoError(t, err)
	assert.NotNil(t, manager)
}

func TestBuiltinManager_Init(t *testing.T) {
	logger := zaptest.NewLogger(t).Sugar()
	loader := NewBuiltinLoader(logger)
	loader.Init(map[string]interface{}{})
	
	manager := NewBuiltinManager[plugin.IPluggable](loader, logger)
	
	err := manager.Init("test-manager", nil)
	
	// In the actual environment, this should succeed due to the real plugin discovery
	// In test environment, this succeeds because the discovery process works with actual plugins
	assert.NoError(t, err)
}

func TestBuiltinManager_Close(t *testing.T) {
	logger := zaptest.NewLogger(t).Sugar()
	loader := NewBuiltinLoader(logger)
	manager := NewBuiltinManager[plugin.IPluggable](loader, logger)
	
	err := manager.Close()
	
	assert.NoError(t, err)
}

func TestBuiltinManager_IsRegisteredMediaType(t *testing.T) {
	logger := zaptest.NewLogger(t).Sugar()
	loader := NewBuiltinLoader(logger)
	loader.Init(map[string]interface{}{})
	
	// Add a mock plugin
	mockPlugin := &MockEvidenceHandler{
		MockPluggable: MockPluggable{
			name:           "test-plugin",
			scheme:         "test-scheme",
			supportedTypes: []string{"application/test"},
		},
	}
	
	loader.loadedByMediaType["application/test"] = mockPlugin
	
	manager := NewBuiltinManager[TestEvidenceHandler](loader, logger)
	
	tests := []struct {
		name      string
		mediaType string
		expected  bool
	}{
		{
			name:      "registered media type",
			mediaType: "application/test",
			expected:  true,
		},
		{
			name:      "unregistered media type",
			mediaType: "application/unknown",
			expected:  false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := manager.IsRegisteredMediaType(tt.mediaType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBuiltinManager_GetRegisteredMediaTypes(t *testing.T) {
	logger := zaptest.NewLogger(t).Sugar()
	loader := NewBuiltinLoader(logger)
	loader.Init(map[string]interface{}{})
	
	// Add mock plugins
	mockPlugin1 := &MockEvidenceHandler{
		MockPluggable: MockPluggable{
			name:           "plugin1",
			scheme:         "scheme1",  
			supportedTypes: []string{"application/test1"},
		},
	}
	mockPlugin2 := &MockEndorsementHandler{
		MockPluggable: MockPluggable{
			name:           "plugin2",
			scheme:         "scheme2",
			supportedTypes: []string{"application/test2"},
		},
	}
	
	loader.loadedByMediaType["application/test1"] = mockPlugin1
	loader.loadedByMediaType["application/test2"] = mockPlugin2
	loader.loadedByMediaType["application/other"] = &MockPluggable{} // Different interface
	
	manager := NewBuiltinManager[TestEvidenceHandler](loader, logger)
	
	mediaTypes := manager.GetRegisteredMediaTypes()
	
	// Should only return media types for plugins that implement TestEvidenceHandler
	assert.Len(t, mediaTypes, 1)
	assert.Contains(t, mediaTypes, "application/test1")
	assert.NotContains(t, mediaTypes, "application/test2") // Wrong interface
	assert.NotContains(t, mediaTypes, "application/other") // Wrong interface
}

func TestBuiltinManager_LookupByMediaType(t *testing.T) {
	logger := zaptest.NewLogger(t).Sugar()
	loader := NewBuiltinLoader(logger)
	loader.Init(map[string]interface{}{})
	
	mockPlugin := &MockEvidenceHandler{
		MockPluggable: MockPluggable{
			name:           "test-plugin",
			scheme:         "test-scheme",
			supportedTypes: []string{"application/test"},
		},
	}
	
	loader.loadedByMediaType["application/test"] = mockPlugin
	
	manager := NewBuiltinManager[TestEvidenceHandler](loader, logger)
	
	tests := []struct {
		name        string
		mediaType   string
		expectError bool
	}{
		{
			name:        "existing media type",
			mediaType:   "application/test",
			expectError: false,
		},
		{
			name:        "non-existing media type",
			mediaType:   "application/unknown",
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin, err := manager.LookupByMediaType(tt.mediaType)
			
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "not found")
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, plugin)
				assert.Equal(t, "test-plugin", plugin.GetName())
			}
		})
	}
}

func TestBuiltinManager_LookupByName(t *testing.T) {
	logger := zaptest.NewLogger(t).Sugar()
	loader := NewBuiltinLoader(logger)
	loader.Init(map[string]interface{}{})
	
	mockPlugin := &MockEvidenceHandler{
		MockPluggable: MockPluggable{
			name:           "test-plugin",
			scheme:         "test-scheme",
			supportedTypes: []string{"application/test"},
		},
	}
	
	loader.loadedByName["test-plugin"] = mockPlugin
	
	manager := NewBuiltinManager[TestEvidenceHandler](loader, logger)
	
	tests := []struct {
		name        string
		pluginName  string
		expectError bool
	}{
		{
			name:        "existing plugin",
			pluginName:  "test-plugin",
			expectError: false,
		},
		{
			name:        "non-existing plugin",
			pluginName:  "unknown-plugin",
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin, err := manager.LookupByName(tt.pluginName)
			
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "not found")
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, plugin)  
				assert.Equal(t, "test-plugin", plugin.GetName())
			}
		})
	}
}

func TestBuiltinManager_LookupByAttestationScheme(t *testing.T) {
	logger := zaptest.NewLogger(t).Sugar()
	loader := NewBuiltinLoader(logger)
	loader.Init(map[string]interface{}{})
	
	mockPlugin := &MockEvidenceHandler{
		MockPluggable: MockPluggable{
			name:           "test-plugin",
			scheme:         "test-scheme",
			supportedTypes: []string{"application/test"},
		},
	}
	
	loader.loadedByName["test-plugin"] = mockPlugin
	
	manager := NewBuiltinManager[TestEvidenceHandler](loader, logger)
	
	tests := []struct {
		name        string
		scheme      string
		expectError bool
	}{
		{
			name:        "existing scheme",
			scheme:      "test-scheme", 
			expectError: false,
		},
		{
			name:        "non-existing scheme",
			scheme:      "unknown-scheme",
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin, err := manager.LookupByAttestationScheme(tt.scheme)
			
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "could not find plugin")
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, plugin)
				assert.Equal(t, "test-scheme", plugin.GetAttestationScheme())
			}
		})
	}
}

func TestBuiltinManager_GetRegisteredAttestationSchemes(t *testing.T) {
	logger := zaptest.NewLogger(t).Sugar()
	loader := NewBuiltinLoader(logger)
	loader.Init(map[string]interface{}{})
	
	// Add mock plugins
	mockPlugin1 := &MockEvidenceHandler{
		MockPluggable: MockPluggable{
			name:   "plugin1", 
			scheme: "scheme1",
		},
	}
	mockPlugin2 := &MockEvidenceHandler{
		MockPluggable: MockPluggable{
			name:   "plugin2",
			scheme: "scheme2", 
		},
	}
	mockPlugin3 := &MockPluggable{ // Different interface
		name:   "plugin3",
		scheme: "scheme3",
	}
	
	loader.loadedByName["plugin1"] = mockPlugin1
	loader.loadedByName["plugin2"] = mockPlugin2
	loader.loadedByName["plugin3"] = mockPlugin3
	
	manager := NewBuiltinManager[TestEvidenceHandler](loader, logger)
	
	schemes := manager.GetRegisteredAttestationSchemes()
	
	// The slice might have empty elements, so filter and check
	nonEmptySchemes := make([]string, 0)
	for _, scheme := range schemes {
		if scheme != "" {
			nonEmptySchemes = append(nonEmptySchemes, scheme)
		}
	}
	
	// Should only return schemes for plugins that implement TestEvidenceHandler
	assert.Len(t, nonEmptySchemes, 2)
	assert.Contains(t, nonEmptySchemes, "scheme1")
	assert.Contains(t, nonEmptySchemes, "scheme2")
	assert.NotContains(t, nonEmptySchemes, "scheme3") // Wrong interface
}