package config

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/golang/glog"
	"io/ioutil"
	"os"
	"time"
)

type  MutableConfigWatcher struct {
	ticker *time.Ticker
	exists bool
	hash string
}

func NewMutableConfigWatcher(canaryConfig *CanaryConfig, applyFunc func(config *MutableCanaryConfig), defaultFunc func() (*MutableCanaryConfig)) (*MutableConfigWatcher, error) {
	mutableConfigWatcher := &MutableConfigWatcher{}

	if canaryConfig.MutableConfigFile != "" && canaryConfig.MutableConfigWatcherInterval > 0 {
		glog.Infof("Starting mutable config watcher for file %s with period %d ms", canaryConfig.MutableConfigFile, canaryConfig.MutableConfigWatcherInterval)

		// Apply any existing config from the file-system
		target, hsh, err := readAndHash(canaryConfig.MutableConfigFile)
		if err == nil && target != nil {
			mutableConfigWatcher.hash = hsh
			mutableConfigWatcher.exists = true
			applyFunc(target)
		}

		mutableConfigWatcher.ticker = time.NewTicker(canaryConfig.MutableConfigWatcherInterval * time.Millisecond)
		quit := make(chan struct{})
		go func() {
			for {
				select {
				case <- mutableConfigWatcher.ticker.C:
					if _, err := os.Stat(canaryConfig.MutableConfigFile); err == nil {
						mutableConfigWatcher.exists = true
						target, hsh, err := readAndHash(canaryConfig.MutableConfigFile)
						if err != nil || target == nil {
							glog.Warningf("failed to read and hash %s : %v (ignored)", canaryConfig.MutableConfigFile, err)
							continue
						}
						if hsh == mutableConfigWatcher.hash {
							continue
						}
						mutableConfigWatcher.hash = hsh
						applyFunc(target)
					} else if errors.Is(err, os.ErrNotExist) && mutableConfigWatcher.exists {
						mutableConfigWatcher.exists = false
						applyFunc(defaultFunc())
					}
				case <- quit:
					mutableConfigWatcher.ticker.Stop()
					return
				}
			}
		}()
	}

	return mutableConfigWatcher, nil
}

func (c MutableConfigWatcher) Close()  {
	if c.ticker != nil {
		c.ticker.Stop()
		c.ticker = nil
	}
}

func readAndHash(filename string) (target *MutableCanaryConfig, h string, err error) {
	byteValue, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}
	target = NewMutableCanaryConfig()
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
