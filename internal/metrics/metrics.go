package metrics

import (
	"expvar"
	"math"
	"sync"
	"time"
)

type Metrics struct {
	mu                         sync.Mutex
	AvgReqPerSecIn60SecWin     *expvar.Int
	TotalRequestReceived       *expvar.Int
	TotalResponsesSent         *expvar.Int
	TotalProcessingTimeMicro   *expvar.Int
	AvgProcessingTimeInMs      *expvar.Float
	ReqPerSec                  *expvar.Map
	TotalResponsesSentByStatus *expvar.Map
	tracker                    map[int64]int32
}

func NewMetrics() *Metrics {
	return &Metrics{
		AvgReqPerSecIn60SecWin:     expvar.NewInt("avg_req_per_sec_in_60_sec_win"),
		TotalRequestReceived:       expvar.NewInt("total_requests_received"),
		TotalResponsesSent:         expvar.NewInt("total_responses_sent"),
		TotalProcessingTimeMicro:   expvar.NewInt("total_processing_time_μs"),
		AvgProcessingTimeInMs:      expvar.NewFloat("avg_processing_time_ms"),
		ReqPerSec:                  expvar.NewMap("req_per_sec"),
		TotalResponsesSentByStatus: expvar.NewMap("total_responses_sent_by_status"),
		tracker:                    make(map[int64]int32),
	}
}

func (m *Metrics) CalculateAvgReqPerSecIn60SecWin() {
	now := time.Now().Unix()
	m.mu.Lock() // because maybe called thousands of times per sec
	defer m.mu.Unlock()
	m.tracker[now]++ // Incrementing the request count for timestamp t
	// Cleanup and sum
	cutOff := now - 60
	var total int32
	for t, n := range m.tracker {
		if t < cutOff {
			delete(m.tracker, t)
		} else {
			total += n
		}
	}
	m.AvgReqPerSecIn60SecWin.Set(int64(total / 60)) // I know the precision loss
}

func (m *Metrics) CalculateReqPerSec() {
	t := time.Now().Unix() - 5
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ReqPerSec.Init()
	for i := 1; i <= 5; i++ {
		m.ReqPerSec.Add(time.Unix(t, 0).Format(time.TimeOnly), int64(m.tracker[t]))
		t++
	}
}

func (m *Metrics) CalculateAvgProcessingTime() {
	if m.TotalRequestReceived.Value() == 0 {
		m.AvgProcessingTimeInMs.Set(0)
		return
	}
	avgMicro := float64(m.TotalProcessingTimeMicro.Value()) / float64(m.TotalRequestReceived.Value())
	avgMs := avgMicro / 1000
	roundedAvg := math.Round(avgMs*1000) / 1000
	m.AvgProcessingTimeInMs.Set(roundedAvg)
}
