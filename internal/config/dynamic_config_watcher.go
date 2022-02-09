package config

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/golang/glog"
	"io/ioutil"
	"os"
	"sync"
	"time"
)

type  DynamicConfigWatcher struct {
	exists bool
	hash   string
	closer sync.Once
	quit   chan struct{}
}

func NewDynamicConfigWatcher(canaryConfig *CanaryConfig, applyFunc func(config *DynamicCanaryConfig), defaultFunc func() (*DynamicCanaryConfig)) (*DynamicConfigWatcher, error) {
	dynamicConfigWatcher := &DynamicConfigWatcher{
		quit: make(chan struct{}),
	}

	if canaryConfig.DynamicConfigFile != "" && canaryConfig.DynamicConfigWatcherInterval > 0 {
		glog.Infof("Starting dynamic config watcher for file %s with period %d ms", canaryConfig.DynamicConfigFile, canaryConfig.DynamicConfigWatcherInterval)

		// Apply any existing config from the file-system
		target, hsh, err := readAndHash(canaryConfig.DynamicConfigFile)
		if err == nil && target != nil {
			dynamicConfigWatcher.hash = hsh
			dynamicConfigWatcher.exists = true
			applyFunc(target)
		}

		go func() {
			ticker := time.NewTicker(canaryConfig.DynamicConfigWatcherInterval * time.Millisecond)
			for {
				select {
				case <- ticker.C:
					if _, err := os.Stat(canaryConfig.DynamicConfigFile); err == nil {
						dynamicConfigWatcher.exists = true
						target, hsh, err := readAndHash(canaryConfig.DynamicConfigFile)
						if err != nil || target == nil {
							glog.Warningf("failed to read and hash %s : %v (ignored)", canaryConfig.DynamicConfigFile, err)
							continue
						}
						if hsh == dynamicConfigWatcher.hash {
							continue
						}
						dynamicConfigWatcher.hash = hsh
						applyFunc(target)
					} else if errors.Is(err, os.ErrNotExist) && dynamicConfigWatcher.exists {
						dynamicConfigWatcher.exists = false
						applyFunc(defaultFunc())
					}
				case <- dynamicConfigWatcher.quit:
					ticker.Stop()
					return
				}
			}
		}()
	}

	return dynamicConfigWatcher, nil
}

func (c *DynamicConfigWatcher) Close()  {
	c.closer.Do(func() {
		close(c.quit)
	})
}

func readAndHash(filename string) (target *DynamicCanaryConfig, h string, err error) {
	byteValue, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}
	target = NewDynamicCanaryConfig()
	err = json.Unmarshal(byteValue, target)
	if err != nil {
		target = nil
		return
	}
	hasher := sha256.New()
	_, err = hasher.Write(byteValue)
	if err != nil {
		return
	}
	h =  hex.EncodeToString(hasher.Sum(nil))
	return
}
