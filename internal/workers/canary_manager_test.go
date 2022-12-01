//
// Copyright Strimzi authors.
// License: Apache License 2.0 (see the file LICENSE or http://apache.org/licenses/LICENSE-2.0.html).
//

//go:build unit_test

package workers

import (
	"github.com/stretchr/testify/assert"
	"github.com/strimzi/strimzi-canary/internal/config"
	"github.com/strimzi/strimzi-canary/internal/services"
	"net/http"
	"sync/atomic"
	"testing"
	"time"
)

type mockTopicService struct {
	refresh atomic.Bool
	leaders map[int32]int32
}

func newMockTopicService() *mockTopicService {
	return &mockTopicService{}
}

func (ts *mockTopicService) Reconcile() (services.TopicReconcileResult, error) {
	refreshed := ts.refresh.CompareAndSwap(true, false)
	return services.TopicReconcileResult{
		RefreshMetadata: refreshed,
	}, nil
}

func (ts *mockTopicService) flagReconcile() {
	ts.refresh.Store(true)
}

func (ts *mockTopicService) resetLeaders(leaders map[int32]int32) {
	ts.leaders = leaders

}

func (ts *mockTopicService) Close() {
}

type mockProducerService struct {
	refreshCounter uint32
}

func newMockProducerService() *mockProducerService {
	return &mockProducerService{}
}

func (m *mockProducerService) Send(_ map[int32][]int32) {
}

func (m *mockProducerService) Refresh() {
	atomic.AddUint32(&m.refreshCounter, 1)
}

func (m *mockProducerService) getRefreshCount() uint32 {
	return atomic.LoadUint32(&m.refreshCounter)
}

func (m mockProducerService) Close() {
}

type mockConsumerService struct {
	refreshCounter uint32
	leaders        map[int32]int32
}

func newMockConsumerService() *mockConsumerService {
	return &mockConsumerService{
		leaders: map[int32]int32{},
	}
}

func (m *mockConsumerService) Consume() {
}

func (m *mockConsumerService) Refresh() {
}

func (m *mockConsumerService) Close() {
}

func (m *mockConsumerService) Leaders() (map[int32]int32, error) {
	return m.leaders, nil
}

func (m *mockConsumerService) getRefreshCount() uint32 {
	return atomic.LoadUint32(&m.refreshCounter)
}

func (m *mockConsumerService) setLeaders(leaders map[int32]int32) {
	m.leaders = leaders
}

type mockConnectionService struct{}

func newMockConnectionService() *mockConnectionService {
	return &mockConnectionService{}
}

func (m mockConnectionService) Open() {
}

func (m mockConnectionService) Close() {
}

type mockStatusService struct{}

func (m mockStatusService) Open() {
}

func (m mockStatusService) Close() {
}

func (m mockStatusService) StatusHandler() http.Handler {
	return nil
}

func newMockStatusService() *mockStatusService {
	return &mockStatusService{}
}

func TestPublisherMetadataRefresh(t *testing.T) {
	cfg := &config.CanaryConfig{
		ReconcileInterval: 1,
	}
	topicService := newMockTopicService()
	producerService := newMockProducerService()
	consumerService := newMockConsumerService()
	connectionService := newMockConnectionService()
	statusService := newMockStatusService()

	manager := NewCanaryManager(cfg, topicService, producerService, consumerService, connectionService, statusService)
	manager.Start()
	defer manager.Stop()

	go func() {
		time.Sleep(time.Millisecond * 50)
		topicService.flagReconcile()
	}()

	assert.Eventually(t, func() bool {
		return producerService.getRefreshCount() == 1
	}, time.Second*1, time.Millisecond*50, "expecting producer service to be refreshed once")
}
