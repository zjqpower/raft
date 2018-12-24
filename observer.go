package raft

import (
	"sync/atomic"
)

// Observation is sent along the given channel to observers when an event occurs.
// 当事件发生时，观察结果通过给定的通道被发送给观察者
type Observation struct {
	// Raft holds the Raft instance generating the observation.
	Raft *Raft
	// Data holds observation-specific data. Possible types are
	// *RequestVoteRequest and RaftState.
	Data interface{}
}

// LeaderObservation is used for the data when leadership changes.
type LeaderObservation struct {
	leader ServerAddress
}

// nextObserverId is used to provide a unique ID for each observer to aid in
// deregistration.
var nextObserverID uint64

// FilterFn is a function that can be registered in order to filter observations.
// The function reports whether the observation should be included - if
// it returns false, the observation will be filtered out.
/* FilterFn是为了过滤观察结果而注册的函数。这个函数确定观察结果是否应该被包含 - 如果他返回
   false，则观察结果将被过滤掉。*/
type FilterFn func(o *Observation) bool

// Observer describes what to do with a given observation.
type Observer struct {
	// numObserved and numDropped are performance counters for this observer.
	// 64 bit types must be 64 bit aligned to use with atomic operations on
	// 32 bit platforms, so keep them at the top of the struct.
	numObserved uint64
	numDropped  uint64

	// channel receives observations.
	// 接收观察结果的通道
	channel chan Observation

	// blocking, if true, will cause Raft to block when sending an observation
	// to this observer. This should generally be set to false.
	// 如果为true，在发送观察结果给这个观察者是会导致raft阻塞。通常应该被设置为false。
	blocking bool

	// filter will be called to determine if an observation should be sent to
	// the channel.
	// 调用filter方法来决定观察结果是否需要发送给这个通道
	filter FilterFn

	// id is the ID of this observer in the Raft map.
	// 观察者唯一标识
	id uint64
}

// NewObserver creates a new observer that can be registered
// to make observations on a Raft instance. Observations
// will be sent on the given channel if they satisfy the
// given filter.
//
// If blocking is true, the observer will block when it can't
// send on the channel, otherwise it may discard events.
func NewObserver(channel chan Observation, blocking bool, filter FilterFn) *Observer {
	return &Observer{
		channel:  channel,
		blocking: blocking,
		filter:   filter,
		id:       atomic.AddUint64(&nextObserverID, 1),
	}
}

// GetNumObserved returns the number of observations.
func (or *Observer) GetNumObserved() uint64 {
	return atomic.LoadUint64(&or.numObserved)
}

// GetNumDropped returns the number of dropped observations due to blocking.
func (or *Observer) GetNumDropped() uint64 {
	return atomic.LoadUint64(&or.numDropped)
}

// RegisterObserver registers a new observer.
func (r *Raft) RegisterObserver(or *Observer) {
	r.observersLock.Lock()
	defer r.observersLock.Unlock()
	r.observers[or.id] = or
}

// DeregisterObserver deregisters an observer.
func (r *Raft) DeregisterObserver(or *Observer) {
	r.observersLock.Lock()
	defer r.observersLock.Unlock()
	delete(r.observers, or.id)
}

// observe sends an observation to every observer.
// 将观察结果发送给每个观察者
func (r *Raft) observe(o interface{}) {
	// In general observers should not block. But in any case this isn't
	// disastrous as we only hold a read lock, which merely prevents
	// registration / deregistration of observers.
	r.observersLock.RLock()
	defer r.observersLock.RUnlock()
	for _, or := range r.observers {
		// It's wasteful to do this in the loop, but for the common case
		// where there are no observers we won't create any objects.
		ob := Observation{Raft: r, Data: o}
		// filter返回false，过滤掉观察结果
		if or.filter != nil && !or.filter(&ob) {
			continue
		}
		// 通道为nil
		if or.channel == nil {
			continue
		}
		if or.blocking {
			or.channel <- ob
			atomic.AddUint64(&or.numObserved, 1)
		} else {
			select {
			case or.channel <- ob:
				atomic.AddUint64(&or.numObserved, 1)
			default:
				atomic.AddUint64(&or.numDropped, 1)
			}
		}
	}
}
