package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestWatcherUsesExistingFile(t *testing.T) {

	configFile := createTempConfigFile(t)

	config := createMutableConfig(99, false)

	writeConfigFile(t, configFile, config)

	wgApply := &sync.WaitGroup{}
	wgApply.Add(1)
	var applied *MutableCanaryConfig
	defaulted := 0

	applyFunc := func(config *MutableCanaryConfig) {
		applied = config
		wgApply.Done()
	}
	defaultFunc := func() *MutableCanaryConfig {
		defaulted++
		return nil
	}

	watcher, err := NewMutableConfigWatcher(createCanaryConfig(configFile), applyFunc, defaultFunc)
	if err != nil {
		t.Errorf("failed to create watcher: %v", err)
	}
	defer watcher.Close()

	waitTimeout(t, wgApply, time.Second)
	if !reflect.DeepEqual(*config, *applied) {
		t.Errorf("unexpected config applied: expected %v actual %v", config, applied)
	}

	expectedDefaulted := 0
	if defaulted != expectedDefaulted {
		t.Errorf("unexpected number of calls to default: expected %d actual %d", expectedDefaulted, defaulted)
	}
}

func TestWatcherSessConfigFileDelete(t *testing.T) {
	configFile := createTempConfigFile(t)

	defaultConfig := createMutableConfig(88, true)
	writeConfigFile(t, configFile, createMutableConfig(99, false))

	once := &sync.Once{}
	wgApply := &sync.WaitGroup{}
	wgApply.Add(2)
	var applied *MutableCanaryConfig

	applyFunc := func(config *MutableCanaryConfig) {
		once.Do(func() {
			err := os.Remove(configFile)
			if err != nil {
				t.Errorf("failed to delete config: %v", err)
			}
		})
		applied = config
		wgApply.Done()
	}
	defaultFunc := func() *MutableCanaryConfig {
		return defaultConfig
	}

	watcher, err := NewMutableConfigWatcher(createCanaryConfig(configFile), applyFunc, defaultFunc)
	if err != nil {
		t.Errorf("failed to create watcher: %v", err)
	}
	defer watcher.Close()

	waitTimeout(t, wgApply, time.Second)
	if !reflect.DeepEqual(*defaultConfig, *applied) {
		t.Errorf("unexpected config applied: expected %v actual %v", defaultConfig, applied)
	}
}

func TestWatcherSeesConfigFileChange(t *testing.T) {
	configFile := createTempConfigFile(t)

	updatedConfig := createMutableConfig(88, true)
	writeConfigFile(t, configFile, createMutableConfig(99, false))

	once := &sync.Once{}
	wgApply := &sync.WaitGroup{}
	wgApply.Add(2)
	var applied *MutableCanaryConfig

	applyFunc := func(config *MutableCanaryConfig) {
		once.Do(func() {
			writeConfigFile(t, configFile, updatedConfig)
		})
		applied = config
		wgApply.Done()
	}
	defaultFunc := func() *MutableCanaryConfig {
		return NewMutableCanaryConfig()
	}

	watcher, err := NewMutableConfigWatcher(createCanaryConfig(configFile), applyFunc, defaultFunc)
	if err != nil {
		t.Errorf("failed to create watcher: %v", err)
	}
	defer watcher.Close()

	waitTimeout(t, wgApply, time.Second)
	if !reflect.DeepEqual(*updatedConfig, *applied) {
		t.Errorf("unexpected config applied: expected %v actual %v", updatedConfig, applied)
	}
}

func createTempConfigFile(t *testing.T) string {
	configFile := filepath.Join(t.TempDir(), "test_*.json")
	return configFile
}

func writeConfigFile(t *testing.T, configFile string, config *MutableCanaryConfig) {
	marshal, err := json.Marshal(config)
	if err != nil {
		t.Errorf("failed to marshal config : %v", err)
	}
	err = os.WriteFile(configFile, marshal, 0644)
	if err != nil {
		t.Errorf("failed to write config : %v", err)
	}
}

func waitTimeout(t *testing.T, wg *sync.WaitGroup, timeout time.Duration) {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()
	select {
	case <-c:
		return // completed normally
	case <-time.After(timeout):
		t.Errorf("Unexpected timeout waiting for condition")
	}
}

func createCanaryConfig(configFile string) *CanaryConfig {
	canaryConfig := &CanaryConfig{
		MutableConfigFile:            configFile,
		MutableConfigWatcherInterval: 50,
	}
	return canaryConfig
}

func createMutableConfig(logLevel int, saramaLogEnabled bool) *MutableCanaryConfig {
	config := NewMutableCanaryConfig()
	config.VerbosityLogLevel = &logLevel
	config.SaramaLogEnabled = &saramaLogEnabled
	return config
}
