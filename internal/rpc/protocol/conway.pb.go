

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        v3.21.11
// source: internal/rpc/protocol/conway.proto

package protocol

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// 辅助类型
type BoundaryType int32

const (
	BoundaryType_PERIODIC   BoundaryType = 0
	BoundaryType_FIXED      BoundaryType = 1
	BoundaryType_REFLECTIVE BoundaryType = 2
)

// Enum value maps for BoundaryType.
var (
	BoundaryType_name = map[int32]string{
		0: "PERIODIC",
		1: "FIXED",
		2: "REFLECTIVE",
	}
	BoundaryType_value = map[string]int32{
		"PERIODIC":   0,
		"FIXED":      1,
		"REFLECTIVE": 2,
	}
)

func (x BoundaryType) Enum() *BoundaryType {
	p := new(BoundaryType)
	*p = x
	return p
}

func (x BoundaryType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (BoundaryType) Descriptor() protoreflect.EnumDescriptor {
	return file_internal_rpc_protocol_conway_proto_enumTypes[0].Descriptor()
}

func (BoundaryType) Type() protoreflect.EnumType {
	return &file_internal_rpc_protocol_conway_proto_enumTypes[0]
}

func (x BoundaryType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use BoundaryType.Descriptor instead.
func (BoundaryType) EnumDescriptor() ([]byte, []int) {
	return file_internal_rpc_protocol_conway_proto_rawDescGZIP(), []int{0}
}

type RegistrationResponse_Status int32

const (
	RegistrationResponse_SUCCESS RegistrationResponse_Status = 0
	RegistrationResponse_FAILED  RegistrationResponse_Status = 1
)

// Enum value maps for RegistrationResponse_Status.
var (
	RegistrationResponse_Status_name = map[int32]string{
		0: "SUCCESS",
		1: "FAILED",
	}
	RegistrationResponse_Status_value = map[string]int32{
		"SUCCESS": 0,
		"FAILED":  1,
	}
)

func (x RegistrationResponse_Status) Enum() *RegistrationResponse_Status {
	p := new(RegistrationResponse_Status)
	*p = x
	return p
}

func (x RegistrationResponse_Status) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (RegistrationResponse_Status) Descriptor() protoreflect.EnumDescriptor {
	return file_internal_rpc_protocol_conway_proto_enumTypes[1].Descriptor()
}

func (RegistrationResponse_Status) Type() protoreflect.EnumType {
	return &file_internal_rpc_protocol_conway_proto_enumTypes[1]
}

func (x RegistrationResponse_Status) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use RegistrationResponse_Status.Descriptor instead.
func (RegistrationResponse_Status) EnumDescriptor() ([]byte, []int) {
	return file_internal_rpc_protocol_conway_proto_rawDescGZIP(), []int{5, 0}
}

type TaskStatusResponse_TaskStatus int32

const (
	TaskStatusResponse_UNKNOWN    TaskStatusResponse_TaskStatus = 0
	TaskStatusResponse_PENDING    TaskStatusResponse_TaskStatus = 1
	TaskStatusResponse_PROCESSING TaskStatusResponse_TaskStatus = 2
	TaskStatusResponse_COMPLETED  TaskStatusResponse_TaskStatus = 3
	TaskStatusResponse_FAILED     TaskStatusResponse_TaskStatus = 4 // 补充FAILED状态
)

// Enum value maps for TaskStatusResponse_TaskStatus.
var (
	TaskStatusResponse_TaskStatus_name = map[int32]string{
		0: "UNKNOWN",
		1: "PENDING",
		2: "PROCESSING",
		3: "COMPLETED",
		4: "FAILED",
	}
	TaskStatusResponse_TaskStatus_value = map[string]int32{
		"UNKNOWN":    0,
		"PENDING":    1,
		"PROCESSING": 2,
		"COMPLETED":  3,
		"FAILED":     4,
	}
)

func (x TaskStatusResponse_TaskStatus) Enum() *TaskStatusResponse_TaskStatus {
	p := new(TaskStatusResponse_TaskStatus)
	*p = x
	return p
}

func (x TaskStatusResponse_TaskStatus) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (TaskStatusResponse_TaskStatus) Descriptor() protoreflect.EnumDescriptor {
	return file_internal_rpc_protocol_conway_proto_enumTypes[2].Descriptor()
}

func (TaskStatusResponse_TaskStatus) Type() protoreflect.EnumType {
	return &file_internal_rpc_protocol_conway_proto_enumTypes[2]
}

func (x TaskStatusResponse_TaskStatus) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use TaskStatusResponse_TaskStatus.Descriptor instead.
func (TaskStatusResponse_TaskStatus) EnumDescriptor() ([]byte, []int) {
	return file_internal_rpc_protocol_conway_proto_rawDescGZIP(), []int{9, 0}
}

// 任务定义
type ComputeTask struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	TaskId        string                 `protobuf:"bytes,1,opt,name=task_id,json=taskId,proto3" json:"task_id,omitempty"`
	GridData      []byte                 `protobuf:"bytes,2,opt,name=grid_data,json=gridData,proto3" json:"grid_data,omitempty"`              // 网格数据
	Generations   int32                  `protobuf:"varint,3,opt,name=generations,proto3" json:"generations,omitempty"`                       // 演化代数
	Boundary      BoundaryType           `protobuf:"varint,4,opt,name=boundary,proto3,enum=golem.rpc.BoundaryType" json:"boundary,omitempty"` // 边界类型
	Timestamp     int64                  `protobuf:"varint,5,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ComputeTask) Reset() {
	*x = ComputeTask{}
	mi := &file_internal_rpc_protocol_conway_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ComputeTask) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ComputeTask) ProtoMessage() {}

func (x *ComputeTask) ProtoReflect() protoreflect.Message {
	mi := &file_internal_rpc_protocol_conway_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ComputeTask.ProtoReflect.Descriptor instead.
func (*ComputeTask) Descriptor() ([]byte, []int) {
	return file_internal_rpc_protocol_conway_proto_rawDescGZIP(), []int{0}
}

func (x *ComputeTask) GetTaskId() string {
	if x != nil {
		return x.TaskId
	}
	return ""
}

func (x *ComputeTask) GetGridData() []byte {
	if x != nil {
		return x.GridData
	}
	return nil
}

func (x *ComputeTask) GetGenerations() int32 {
	if x != nil {
		return x.Generations
	}
	return 0
}

func (x *ComputeTask) GetBoundary() BoundaryType {
	if x != nil {
		return x.Boundary
	}
	return BoundaryType_PERIODIC
}

func (x *ComputeTask) GetTimestamp() int64 {
	if x != nil {
		return x.Timestamp
	}
	return 0
}

// 计算结果
type ComputeResult struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	TaskId        string                 `protobuf:"bytes,1,opt,name=task_id,json=taskId,proto3" json:"task_id,omitempty"`             //任务ID
	ResultData    []byte                 `protobuf:"bytes,2,opt,name=result_data,json=resultData,proto3" json:"result_data,omitempty"` // 结果数据
	ExecTime      float64                `protobuf:"fixed64,3,opt,name=exec_time,json=execTime,proto3" json:"exec_time,omitempty"`     // 执行时间(秒)
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ComputeResult) Reset() {
	*x = ComputeResult{}
	mi := &file_internal_rpc_protocol_conway_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ComputeResult) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ComputeResult) ProtoMessage() {}

func (x *ComputeResult) ProtoReflect() protoreflect.Message {
	mi := &file_internal_rpc_protocol_conway_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ComputeResult.ProtoReflect.Descriptor instead.
func (*ComputeResult) Descriptor() ([]byte, []int) {
	return file_internal_rpc_protocol_conway_proto_rawDescGZIP(), []int{1}
}

func (x *ComputeResult) GetTaskId() string {
	if x != nil {
		return x.TaskId
	}
	return ""
}

func (x *ComputeResult) GetResultData() []byte {
	if x != nil {
		return x.ResultData
	}
	return nil
}

func (x *ComputeResult) GetExecTime() float64 {
	if x != nil {
		return x.ExecTime
	}
	return 0
}

// 节点注册
type WorkerRegistration struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Address       string                 `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`    // 工作节点地址
	Capacity      int32                  `protobuf:"varint,2,opt,name=capacity,proto3" json:"capacity,omitempty"` // 最大并行任务数
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *WorkerRegistration) Reset() {
	*x = WorkerRegistration{}
	mi := &file_internal_rpc_protocol_conway_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *WorkerRegistration) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*WorkerRegistration) ProtoMessage() {}

func (x *WorkerRegistration) ProtoReflect() protoreflect.Message {
	mi := &file_internal_rpc_protocol_conway_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use WorkerRegistration.ProtoReflect.Descriptor instead.
func (*WorkerRegistration) Descriptor() ([]byte, []int) {
	return file_internal_rpc_protocol_conway_proto_rawDescGZIP(), []int{2}
}

func (x *WorkerRegistration) GetAddress() string {
	if x != nil {
		return x.Address
	}
	return ""
}

func (x *WorkerRegistration) GetCapacity() int32 {
	if x != nil {
		return x.Capacity
	}
	return 0
}

// 新增消息类型
type ComputeRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	GridData      []byte                 `protobuf:"bytes,1,opt,name=grid_data,json=gridData,proto3" json:"grid_data,omitempty"`
	Generations   int32                  `protobuf:"varint,2,opt,name=generations,proto3" json:"generations,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ComputeRequest) Reset() {
	*x = ComputeRequest{}
	mi := &file_internal_rpc_protocol_conway_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ComputeRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ComputeRequest) ProtoMessage() {}

func (x *ComputeRequest) ProtoReflect() protoreflect.Message {
	mi := &file_internal_rpc_protocol_conway_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ComputeRequest.ProtoReflect.Descriptor instead.
func (*ComputeRequest) Descriptor() ([]byte, []int) {
	return file_internal_rpc_protocol_conway_proto_rawDescGZIP(), []int{3}
}

func (x *ComputeRequest) GetGridData() []byte {
	if x != nil {
		return x.GridData
	}
	return nil
}

func (x *ComputeRequest) GetGenerations() int32 {
	if x != nil {
		return x.Generations
	}
	return 0
}

type TaskResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	TaskId        string                 `protobuf:"bytes,1,opt,name=task_id,json=taskId,proto3" json:"task_id,omitempty"`
	Timestamp     int64                  `protobuf:"varint,2,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *TaskResponse) Reset() {
	*x = TaskResponse{}
	mi := &file_internal_rpc_protocol_conway_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *TaskResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TaskResponse) ProtoMessage() {}

func (x *TaskResponse) ProtoReflect() protoreflect.Message {
	mi := &file_internal_rpc_protocol_conway_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TaskResponse.ProtoReflect.Descriptor instead.
func (*TaskResponse) Descriptor() ([]byte, []int) {
	return file_internal_rpc_protocol_conway_proto_rawDescGZIP(), []int{4}
}

func (x *TaskResponse) GetTaskId() string {
	if x != nil {
		return x.TaskId
	}
	return ""
}

func (x *TaskResponse) GetTimestamp() int64 {
	if x != nil {
		return x.Timestamp
	}
	return 0
}

type RegistrationResponse struct {
	state             protoimpl.MessageState      `protogen:"open.v1"`
	Status            RegistrationResponse_Status `protobuf:"varint,1,opt,name=status,proto3,enum=golem.rpc.RegistrationResponse_Status" json:"status,omitempty"`
	AssignedId        string                      `protobuf:"bytes,2,opt,name=assigned_id,json=assignedId,proto3" json:"assigned_id,omitempty"` // 确保proto定义使用下划线命名
	HeartbeatInterval int32                       `protobuf:"varint,3,opt,name=heartbeat_interval,json=heartbeatInterval,proto3" json:"heartbeat_interval,omitempty"`
	unknownFields     protoimpl.UnknownFields
	sizeCache         protoimpl.SizeCache
}

func (x *RegistrationResponse) Reset() {
	*x = RegistrationResponse{}
	mi := &file_internal_rpc_protocol_conway_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *RegistrationResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RegistrationResponse) ProtoMessage() {}

func (x *RegistrationResponse) ProtoReflect() protoreflect.Message {
	mi := &file_internal_rpc_protocol_conway_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RegistrationResponse.ProtoReflect.Descriptor instead.
func (*RegistrationResponse) Descriptor() ([]byte, []int) {
	return file_internal_rpc_protocol_conway_proto_rawDescGZIP(), []int{5}
}

func (x *RegistrationResponse) GetStatus() RegistrationResponse_Status {
	if x != nil {
		return x.Status
	}
	return RegistrationResponse_SUCCESS
}

func (x *RegistrationResponse) GetAssignedId() string {
	if x != nil {
		return x.AssignedId
	}
	return ""
}

func (x *RegistrationResponse) GetHeartbeatInterval() int32 {
	if x != nil {
		return x.HeartbeatInterval
	}
	return 0
}

type RegistrationRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Address       string                 `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *RegistrationRequest) Reset() {
	*x = RegistrationRequest{}
	mi := &file_internal_rpc_protocol_conway_proto_msgTypes[6]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *RegistrationRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RegistrationRequest) ProtoMessage() {}

func (x *RegistrationRequest) ProtoReflect() protoreflect.Message {
	mi := &file_internal_rpc_protocol_conway_proto_msgTypes[6]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RegistrationRequest.ProtoReflect.Descriptor instead.
func (*RegistrationRequest) Descriptor() ([]byte, []int) {
	return file_internal_rpc_protocol_conway_proto_rawDescGZIP(), []int{6}
}

func (x *RegistrationRequest) GetAddress() string {
	if x != nil {
		return x.Address
	}
	return ""
}

// 新增流式结果请求协议
type StreamRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	ClientId      string                 `protobuf:"bytes,1,opt,name=client_id,json=clientId,proto3" json:"client_id,omitempty"` // 修正字段名称为client_id
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *StreamRequest) Reset() {
	*x = StreamRequest{}
	mi := &file_internal_rpc_protocol_conway_proto_msgTypes[7]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *StreamRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StreamRequest) ProtoMessage() {}

func (x *StreamRequest) ProtoReflect() protoreflect.Message {
	mi := &file_internal_rpc_protocol_conway_proto_msgTypes[7]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StreamRequest.ProtoReflect.Descriptor instead.
func (*StreamRequest) Descriptor() ([]byte, []int) {
	return file_internal_rpc_protocol_conway_proto_rawDescGZIP(), []int{7}
}

func (x *StreamRequest) GetClientId() string {
	if x != nil {
		return x.ClientId
	}
	return ""
}

// 任务状态查询协议
type TaskStatusRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	TaskId        string                 `protobuf:"bytes,1,opt,name=task_id,json=taskId,proto3" json:"task_id,omitempty"` // 修正字段名称为task_id
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *TaskStatusRequest) Reset() {
	*x = TaskStatusRequest{}
	mi := &file_internal_rpc_protocol_conway_proto_msgTypes[8]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *TaskStatusRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TaskStatusRequest) ProtoMessage() {}

func (x *TaskStatusRequest) ProtoReflect() protoreflect.Message {
	mi := &file_internal_rpc_protocol_conway_proto_msgTypes[8]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TaskStatusRequest.ProtoReflect.Descriptor instead.
func (*TaskStatusRequest) Descriptor() ([]byte, []int) {
	return file_internal_rpc_protocol_conway_proto_rawDescGZIP(), []int{8}
}

func (x *TaskStatusRequest) GetTaskId() string {
	if x != nil {
		return x.TaskId
	}
	return ""
}

type TaskStatusResponse struct {
	state         protoimpl.MessageState        `protogen:"open.v1"`
	Status        TaskStatusResponse_TaskStatus `protobuf:"varint,1,opt,name=status,proto3,enum=golem.rpc.TaskStatusResponse_TaskStatus" json:"status,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *TaskStatusResponse) Reset() {
	*x = TaskStatusResponse{}
	mi := &file_internal_rpc_protocol_conway_proto_msgTypes[9]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *TaskStatusResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TaskStatusResponse) ProtoMessage() {}

func (x *TaskStatusResponse) ProtoReflect() protoreflect.Message {
	mi := &file_internal_rpc_protocol_conway_proto_msgTypes[9]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TaskStatusResponse.ProtoReflect.Descriptor instead.
func (*TaskStatusResponse) Descriptor() ([]byte, []int) {
	return file_internal_rpc_protocol_conway_proto_rawDescGZIP(), []int{9}
}

func (x *TaskStatusResponse) GetStatus() TaskStatusResponse_TaskStatus {
	if x != nil {
		return x.Status
	}
	return TaskStatusResponse_UNKNOWN
}

// 节点同步协议
type SyncNodesRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Nodes         []*WorkerRegistration  `protobuf:"bytes,1,rep,name=nodes,proto3" json:"nodes,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *SyncNodesRequest) Reset() {
	*x = SyncNodesRequest{}
	mi := &file_internal_rpc_protocol_conway_proto_msgTypes[10]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SyncNodesRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SyncNodesRequest) ProtoMessage() {}

func (x *SyncNodesRequest) ProtoReflect() protoreflect.Message {
	mi := &file_internal_rpc_protocol_conway_proto_msgTypes[10]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SyncNodesRequest.ProtoReflect.Descriptor instead.
func (*SyncNodesRequest) Descriptor() ([]byte, []int) {
	return file_internal_rpc_protocol_conway_proto_rawDescGZIP(), []int{10}
}

func (x *SyncNodesRequest) GetNodes() []*WorkerRegistration {
	if x != nil {
		return x.Nodes
	}
	return nil
}

type SyncNodesResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Success       bool                   `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *SyncNodesResponse) Reset() {
	*x = SyncNodesResponse{}
	mi := &file_internal_rpc_protocol_conway_proto_msgTypes[11]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SyncNodesResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SyncNodesResponse) ProtoMessage() {}

func (x *SyncNodesResponse) ProtoReflect() protoreflect.Message {
	mi := &file_internal_rpc_protocol_conway_proto_msgTypes[11]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SyncNodesResponse.ProtoReflect.Descriptor instead.
func (*SyncNodesResponse) Descriptor() ([]byte, []int) {
	return file_internal_rpc_protocol_conway_proto_rawDescGZIP(), []int{11}
}

func (x *SyncNodesResponse) GetSuccess() bool {
	if x != nil {
		return x.Success
	}
	return false
}

var File_internal_rpc_protocol_conway_proto protoreflect.FileDescriptor

const file_internal_rpc_protocol_conway_proto_rawDesc = "" +
	"\n" +
	"\"internal/rpc/protocol/conway.proto\x12\tgolem.rpc\"\xb8\x01\n" +
	"\vComputeTask\x12\x17\n" +
	"\atask_id\x18\x01 \x01(\tR\x06taskId\x12\x1b\n" +
	"\tgrid_data\x18\x02 \x01(\fR\bgridData\x12 \n" +
	"\vgenerations\x18\x03 \x01(\x05R\vgenerations\x123\n" +
	"\bboundary\x18\x04 \x01(\x0e2\x17.golem.rpc.BoundaryTypeR\bboundary\x12\x1c\n" +
	"\ttimestamp\x18\x05 \x01(\x03R\ttimestamp\"f\n" +
	"\rComputeResult\x12\x17\n" +
	"\atask_id\x18\x01 \x01(\tR\x06taskId\x12\x1f\n" +
	"\vresult_data\x18\x02 \x01(\fR\n" +
	"resultData\x12\x1b\n" +
	"\texec_time\x18\x03 \x01(\x01R\bexecTime\"J\n" +
	"\x12WorkerRegistration\x12\x18\n" +
	"\aaddress\x18\x01 \x01(\tR\aaddress\x12\x1a\n" +
	"\bcapacity\x18\x02 \x01(\x05R\bcapacity\"O\n" +
	"\x0eComputeRequest\x12\x1b\n" +
	"\tgrid_data\x18\x01 \x01(\fR\bgridData\x12 \n" +
	"\vgenerations\x18\x02 \x01(\x05R\vgenerations\"E\n" +
	"\fTaskResponse\x12\x17\n" +
	"\atask_id\x18\x01 \x01(\tR\x06taskId\x12\x1c\n" +
	"\ttimestamp\x18\x02 \x01(\x03R\ttimestamp\"\xc9\x01\n" +
	"\x14RegistrationResponse\x12>\n" +
	"\x06status\x18\x01 \x01(\x0e2&.golem.rpc.RegistrationResponse.StatusR\x06status\x12\x1f\n" +
	"\vassigned_id\x18\x02 \x01(\tR\n" +
	"assignedId\x12-\n" +
	"\x12heartbeat_interval\x18\x03 \x01(\x05R\x11heartbeatInterval\"!\n" +
	"\x06Status\x12\v\n" +
	"\aSUCCESS\x10\x00\x12\n" +
	"\n" +
	"\x06FAILED\x10\x01\"/\n" +
	"\x13RegistrationRequest\x12\x18\n" +
	"\aaddress\x18\x01 \x01(\tR\aaddress\",\n" +
	"\rStreamRequest\x12\x1b\n" +
	"\tclient_id\x18\x01 \x01(\tR\bclientId\",\n" +
	"\x11TaskStatusRequest\x12\x17\n" +
	"\atask_id\x18\x01 \x01(\tR\x06taskId\"\xa9\x01\n" +
	"\x12TaskStatusResponse\x12@\n" +
	"\x06status\x18\x01 \x01(\x0e2(.golem.rpc.TaskStatusResponse.TaskStatusR\x06status\"Q\n" +
	"\n" +
	"TaskStatus\x12\v\n" +
	"\aUNKNOWN\x10\x00\x12\v\n" +
	"\aPENDING\x10\x01\x12\x0e\n" +
	"\n" +
	"PROCESSING\x10\x02\x12\r\n" +
	"\tCOMPLETED\x10\x03\x12\n" +
	"\n" +
	"\x06FAILED\x10\x04\"G\n" +
	"\x10SyncNodesRequest\x123\n" +
	"\x05nodes\x18\x01 \x03(\v2\x1d.golem.rpc.WorkerRegistrationR\x05nodes\"-\n" +
	"\x11SyncNodesResponse\x12\x18\n" +
	"\asuccess\x18\x01 \x01(\bR\asuccess*7\n" +
	"\fBoundaryType\x12\f\n" +
	"\bPERIODIC\x10\x00\x12\t\n" +
	"\x05FIXED\x10\x01\x12\x0e\n" +
	"\n" +
	"REFLECTIVE\x10\x022\xb5\x02\n" +
	"\rBrokerService\x12=\n" +
	"\n" +
	"SubmitTask\x12\x16.golem.rpc.ComputeTask\x1a\x17.golem.rpc.TaskResponse\x12P\n" +
	"\x0eRegisterWorker\x12\x1d.golem.rpc.WorkerRegistration\x1a\x1f.golem.rpc.RegistrationResponse\x12E\n" +
	"\rStreamResults\x12\x18.golem.rpc.StreamRequest\x1a\x18.golem.rpc.ComputeResult0\x01\x12L\n" +
	"\rGetTaskStatus\x12\x1c.golem.rpc.TaskStatusRequest\x1a\x1d.golem.rpc.TaskStatusResponse2\x97\x01\n" +
	"\rConwayService\x12?\n" +
	"\vComputeSync\x12\x16.golem.rpc.ComputeTask\x1a\x18.golem.rpc.ComputeResult\x12E\n" +
	"\rComputeStream\x12\x16.golem.rpc.ComputeTask\x1a\x18.golem.rpc.ComputeResult(\x010\x01B!Z\x1fGolemCore/internal/rpc/protocolb\x06proto3"

var (
	file_internal_rpc_protocol_conway_proto_rawDescOnce sync.Once
	file_internal_rpc_protocol_conway_proto_rawDescData []byte
)

func file_internal_rpc_protocol_conway_proto_rawDescGZIP() []byte {
	file_internal_rpc_protocol_conway_proto_rawDescOnce.Do(func() {
		file_internal_rpc_protocol_conway_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_internal_rpc_protocol_conway_proto_rawDesc), len(file_internal_rpc_protocol_conway_proto_rawDesc)))
	})
	return file_internal_rpc_protocol_conway_proto_rawDescData
}

var file_internal_rpc_protocol_conway_proto_enumTypes = make([]protoimpl.EnumInfo, 3)
var file_internal_rpc_protocol_conway_proto_msgTypes = make([]protoimpl.MessageInfo, 12)
var file_internal_rpc_protocol_conway_proto_goTypes = []any{
	(BoundaryType)(0),                  // 0: golem.rpc.BoundaryType
	(RegistrationResponse_Status)(0),   // 1: golem.rpc.RegistrationResponse.Status
	(TaskStatusResponse_TaskStatus)(0), // 2: golem.rpc.TaskStatusResponse.TaskStatus
	(*ComputeTask)(nil),                // 3: golem.rpc.ComputeTask
	(*ComputeResult)(nil),              // 4: golem.rpc.ComputeResult
	(*WorkerRegistration)(nil),         // 5: golem.rpc.WorkerRegistration
	(*ComputeRequest)(nil),             // 6: golem.rpc.ComputeRequest
	(*TaskResponse)(nil),               // 7: golem.rpc.TaskResponse
	(*RegistrationResponse)(nil),       // 8: golem.rpc.RegistrationResponse
	(*RegistrationRequest)(nil),        // 9: golem.rpc.RegistrationRequest
	(*StreamRequest)(nil),              // 10: golem.rpc.StreamRequest
	(*TaskStatusRequest)(nil),          // 11: golem.rpc.TaskStatusRequest
	(*TaskStatusResponse)(nil),         // 12: golem.rpc.TaskStatusResponse
	(*SyncNodesRequest)(nil),           // 13: golem.rpc.SyncNodesRequest
	(*SyncNodesResponse)(nil),          // 14: golem.rpc.SyncNodesResponse
}
var file_internal_rpc_protocol_conway_proto_depIdxs = []int32{
	0,  // 0: golem.rpc.ComputeTask.boundary:type_name -> golem.rpc.BoundaryType
	1,  // 1: golem.rpc.RegistrationResponse.status:type_name -> golem.rpc.RegistrationResponse.Status
	2,  // 2: golem.rpc.TaskStatusResponse.status:type_name -> golem.rpc.TaskStatusResponse.TaskStatus
	5,  // 3: golem.rpc.SyncNodesRequest.nodes:type_name -> golem.rpc.WorkerRegistration
	3,  // 4: golem.rpc.BrokerService.SubmitTask:input_type -> golem.rpc.ComputeTask
	5,  // 5: golem.rpc.BrokerService.RegisterWorker:input_type -> golem.rpc.WorkerRegistration
	10, // 6: golem.rpc.BrokerService.StreamResults:input_type -> golem.rpc.StreamRequest
	11, // 7: golem.rpc.BrokerService.GetTaskStatus:input_type -> golem.rpc.TaskStatusRequest
	3,  // 8: golem.rpc.ConwayService.ComputeSync:input_type -> golem.rpc.ComputeTask
	3,  // 9: golem.rpc.ConwayService.ComputeStream:input_type -> golem.rpc.ComputeTask
	7,  // 10: golem.rpc.BrokerService.SubmitTask:output_type -> golem.rpc.TaskResponse
	8,  // 11: golem.rpc.BrokerService.RegisterWorker:output_type -> golem.rpc.RegistrationResponse
	4,  // 12: golem.rpc.BrokerService.StreamResults:output_type -> golem.rpc.ComputeResult
	12, // 13: golem.rpc.BrokerService.GetTaskStatus:output_type -> golem.rpc.TaskStatusResponse
	4,  // 14: golem.rpc.ConwayService.ComputeSync:output_type -> golem.rpc.ComputeResult
	4,  // 15: golem.rpc.ConwayService.ComputeStream:output_type -> golem.rpc.ComputeResult
	10, // [10:16] is the sub-list for method output_type
	4,  // [4:10] is the sub-list for method input_type
	4,  // [4:4] is the sub-list for extension type_name
	4,  // [4:4] is the sub-list for extension extendee
	0,  // [0:4] is the sub-list for field type_name
}

func init() { file_internal_rpc_protocol_conway_proto_init() }
func file_internal_rpc_protocol_conway_proto_init() {
	if File_internal_rpc_protocol_conway_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_internal_rpc_protocol_conway_proto_rawDesc), len(file_internal_rpc_protocol_conway_proto_rawDesc)),
			NumEnums:      3,
			NumMessages:   12,
			NumExtensions: 0,
			NumServices:   2,
		},
		GoTypes:           file_internal_rpc_protocol_conway_proto_goTypes,
		DependencyIndexes: file_internal_rpc_protocol_conway_proto_depIdxs,
		EnumInfos:         file_internal_rpc_protocol_conway_proto_enumTypes,
		MessageInfos:      file_internal_rpc_protocol_conway_proto_msgTypes,
	}.Build()
	File_internal_rpc_protocol_conway_proto = out.File
	file_internal_rpc_protocol_conway_proto_goTypes = nil
	file_internal_rpc_protocol_conway_proto_depIdxs = nil
}
