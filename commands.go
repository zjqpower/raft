package raft

// RPCHeader is a common sub-structure used to pass along protocol version and
// other information about the cluster. For older Raft implementations before
// versioning was added this will default to a zero-valued structure when read
// by newer Raft versions.
type RPCHeader struct {
	// ProtocolVersion is the version of the protocol the sender is
	// speaking.
	ProtocolVersion ProtocolVersion
}

// WithRPCHeader is an interface that exposes the RPC header.
type WithRPCHeader interface {
	GetRPCHeader() RPCHeader
}

// AppendEntriesRequest is the command used to append entries to the
// replicated log.
// 用于向副本日志追加条目
type AppendEntriesRequest struct {
	RPCHeader

	// Provide the current term and leader
	Term   uint64 //领导人的任期号
	Leader []byte //领导人的 id，为了其他服务器能重定向到客户端

	// Provide the previous entries for integrity checking
	PrevLogEntry uint64 // 最新日志之前的日志的索引值
	PrevLogTerm  uint64 // 最新日志之前的日志的领导人任期号

	// New entries to commit
	Entries []*Log // 将要存储的日志条目（表示 heartbeat 时为空，有时会为了效率发送超过一条）

	// Commit index on the leader
	LeaderCommitIndex uint64 // 领导人提交的日志条目索引值
}

// See WithRPCHeader.
func (r *AppendEntriesRequest) GetRPCHeader() RPCHeader {
	return r.RPCHeader
}

// AppendEntriesResponse is the response returned from an
// AppendEntriesRequest.
type AppendEntriesResponse struct {
	RPCHeader

	// Newer term if leader is out of date
	Term uint64

	// Last Log is a hint to help accelerate rebuilding slow nodes
	LastLog uint64

	// We may not succeed if we have a conflicting entry
	Success bool

	// There are scenarios where this request didn't succeed
	// but there's no need to wait/back-off the next attempt.
	// 有些情况下，这个请求没有成功，但也不需要进行等待或进行下一次尝试
	NoRetryBackoff bool
}

// See WithRPCHeader.
func (r *AppendEntriesResponse) GetRPCHeader() RPCHeader {
	return r.RPCHeader
}

// RequestVoteRequest is the command used by a candidate to ask a Raft peer
// for a vote in an election.
type RequestVoteRequest struct {
	RPCHeader

	// Provide the term and our id
	Term      uint64 // 候选人的任期号
	Candidate []byte // 请求投票的候选人ID

	// Used to ensure safety
	LastLogIndex uint64 // 候选人最新日志条目的索引值
	LastLogTerm  uint64 // 候选人最新日志条目对应的任期号
}

// See WithRPCHeader.
func (r *RequestVoteRequest) GetRPCHeader() RPCHeader {
	return r.RPCHeader
}

// RequestVoteResponse is the response returned from a RequestVoteRequest.
type RequestVoteResponse struct {
	RPCHeader

	// Newer term if leader is out of date.
	Term uint64 // 目前的任期号，用于候选人更新自己

	// Peers is deprecated, but required by servers that only understand
	// protocol version 0. This is not populated in protocol version 2
	// and later.
	Peers []byte // 协议0需要此字段，协议2及之后的版本不再需要此字段

	// Is the vote granted.
	Granted bool // 如果候选人收到投票为true
}

// See WithRPCHeader.
func (r *RequestVoteResponse) GetRPCHeader() RPCHeader {
	return r.RPCHeader
}

// InstallSnapshotRequest is the command sent to a Raft peer to bootstrap its
// log (and state machine) from a snapshot on another peer.
// 安装快照请求，领导人发给追随者使用
type InstallSnapshotRequest struct {
	RPCHeader
	SnapshotVersion SnapshotVersion // 快照版本，用于兼容旧服务

	Term   uint64 // 领导人的任期号
	Leader []byte // 领导人的标识

	// These are the last index/term included in the snapshot
	LastLogIndex uint64 // 快照中包含的最后日志条目的索引值
	LastLogTerm  uint64 // 快照中包含的最后日志条目的任期号

	// Peer Set in the snapshot. This is deprecated in favor of Configuration
	// but remains here in case we receive an InstallSnapshot from a leader
	// that's running old code.
	Peers []byte // 新的不需要，为了兼容旧版本

	// Cluster membership.
	Configuration []byte // 配置：集群成员关系
	// Log index where 'Configuration' entry was originally written.
	ConfigurationIndex uint64 // 配置索引号

	// Size of the snapshot
	Size int64 // 快照的大小
}

// See WithRPCHeader.
func (r *InstallSnapshotRequest) GetRPCHeader() RPCHeader {
	return r.RPCHeader
}

// InstallSnapshotResponse is the response returned from an
// InstallSnapshotRequest.
type InstallSnapshotResponse struct {
	RPCHeader

	Term    uint64 // 当前的任期号
	Success bool
}

// See WithRPCHeader.
func (r *InstallSnapshotResponse) GetRPCHeader() RPCHeader {
	return r.RPCHeader
}
