// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v4.25.1
// source: api/grpc/device/device.proto

package device

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type StatusRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id       string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Uuid     string `protobuf:"bytes,2,opt,name=uuid,proto3" json:"uuid,omitempty"`
	Mac      string `protobuf:"bytes,3,opt,name=mac,proto3" json:"mac,omitempty"`
	ApiToken string `protobuf:"bytes,4,opt,name=api_token,json=apiToken,proto3" json:"api_token,omitempty"`
}

func (x *StatusRequest) Reset() {
	*x = StatusRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_grpc_device_device_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StatusRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StatusRequest) ProtoMessage() {}

func (x *StatusRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_grpc_device_device_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StatusRequest.ProtoReflect.Descriptor instead.
func (*StatusRequest) Descriptor() ([]byte, []int) {
	return file_api_grpc_device_device_proto_rawDescGZIP(), []int{0}
}

func (x *StatusRequest) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *StatusRequest) GetUuid() string {
	if x != nil {
		return x.Uuid
	}
	return ""
}

func (x *StatusRequest) GetMac() string {
	if x != nil {
		return x.Mac
	}
	return ""
}

func (x *StatusRequest) GetApiToken() string {
	if x != nil {
		return x.ApiToken
	}
	return ""
}

type StatusResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	On          bool  `protobuf:"varint,1,opt,name=on,proto3" json:"on,omitempty"`
	Temperature int32 `protobuf:"varint,2,opt,name=temperature,proto3" json:"temperature,omitempty"`
	Mode        int32 `protobuf:"varint,3,opt,name=mode,proto3" json:"mode,omitempty"`
	FanSpeed    int32 `protobuf:"varint,4,opt,name=fan_speed,json=fanSpeed,proto3" json:"fan_speed,omitempty"`
	CreatedAt   int64 `protobuf:"varint,5,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty"`
	ModifiedAt  int64 `protobuf:"varint,6,opt,name=modified_at,json=modifiedAt,proto3" json:"modified_at,omitempty"`
}

func (x *StatusResponse) Reset() {
	*x = StatusResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_grpc_device_device_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StatusResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StatusResponse) ProtoMessage() {}

func (x *StatusResponse) ProtoReflect() protoreflect.Message {
	mi := &file_api_grpc_device_device_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StatusResponse.ProtoReflect.Descriptor instead.
func (*StatusResponse) Descriptor() ([]byte, []int) {
	return file_api_grpc_device_device_proto_rawDescGZIP(), []int{1}
}

func (x *StatusResponse) GetOn() bool {
	if x != nil {
		return x.On
	}
	return false
}

func (x *StatusResponse) GetTemperature() int32 {
	if x != nil {
		return x.Temperature
	}
	return 0
}

func (x *StatusResponse) GetMode() int32 {
	if x != nil {
		return x.Mode
	}
	return 0
}

func (x *StatusResponse) GetFanSpeed() int32 {
	if x != nil {
		return x.FanSpeed
	}
	return 0
}

func (x *StatusResponse) GetCreatedAt() int64 {
	if x != nil {
		return x.CreatedAt
	}
	return 0
}

func (x *StatusResponse) GetModifiedAt() int64 {
	if x != nil {
		return x.ModifiedAt
	}
	return 0
}

type ValuesRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id          string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Uuid        string `protobuf:"bytes,2,opt,name=uuid,proto3" json:"uuid,omitempty"`
	Mac         string `protobuf:"bytes,3,opt,name=mac,proto3" json:"mac,omitempty"`
	ApiToken    string `protobuf:"bytes,4,opt,name=api_token,json=apiToken,proto3" json:"api_token,omitempty"`
	On          bool   `protobuf:"varint,5,opt,name=on,proto3" json:"on,omitempty"`
	Temperature int32  `protobuf:"varint,6,opt,name=temperature,proto3" json:"temperature,omitempty"`
	Mode        int32  `protobuf:"varint,7,opt,name=mode,proto3" json:"mode,omitempty"`
	FanSpeed    int32  `protobuf:"varint,8,opt,name=fan_speed,json=fanSpeed,proto3" json:"fan_speed,omitempty"`
}

func (x *ValuesRequest) Reset() {
	*x = ValuesRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_grpc_device_device_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ValuesRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ValuesRequest) ProtoMessage() {}

func (x *ValuesRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_grpc_device_device_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ValuesRequest.ProtoReflect.Descriptor instead.
func (*ValuesRequest) Descriptor() ([]byte, []int) {
	return file_api_grpc_device_device_proto_rawDescGZIP(), []int{2}
}

func (x *ValuesRequest) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *ValuesRequest) GetUuid() string {
	if x != nil {
		return x.Uuid
	}
	return ""
}

func (x *ValuesRequest) GetMac() string {
	if x != nil {
		return x.Mac
	}
	return ""
}

func (x *ValuesRequest) GetApiToken() string {
	if x != nil {
		return x.ApiToken
	}
	return ""
}

func (x *ValuesRequest) GetOn() bool {
	if x != nil {
		return x.On
	}
	return false
}

func (x *ValuesRequest) GetTemperature() int32 {
	if x != nil {
		return x.Temperature
	}
	return 0
}

func (x *ValuesRequest) GetMode() int32 {
	if x != nil {
		return x.Mode
	}
	return 0
}

func (x *ValuesRequest) GetFanSpeed() int32 {
	if x != nil {
		return x.FanSpeed
	}
	return 0
}

type ValuesResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Status  string `protobuf:"bytes,1,opt,name=status,proto3" json:"status,omitempty"`
	Message string `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
}

func (x *ValuesResponse) Reset() {
	*x = ValuesResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_grpc_device_device_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ValuesResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ValuesResponse) ProtoMessage() {}

func (x *ValuesResponse) ProtoReflect() protoreflect.Message {
	mi := &file_api_grpc_device_device_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ValuesResponse.ProtoReflect.Descriptor instead.
func (*ValuesResponse) Descriptor() ([]byte, []int) {
	return file_api_grpc_device_device_proto_rawDescGZIP(), []int{3}
}

func (x *ValuesResponse) GetStatus() string {
	if x != nil {
		return x.Status
	}
	return ""
}

func (x *ValuesResponse) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

var File_api_grpc_device_device_proto protoreflect.FileDescriptor

var file_api_grpc_device_device_proto_rawDesc = []byte{
	0x0a, 0x1c, 0x61, 0x70, 0x69, 0x2f, 0x67, 0x72, 0x70, 0x63, 0x2f, 0x64, 0x65, 0x76, 0x69, 0x63,
	0x65, 0x2f, 0x64, 0x65, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x06,
	0x64, 0x65, 0x76, 0x69, 0x63, 0x65, 0x22, 0x62, 0x0a, 0x0d, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x75, 0x75, 0x69, 0x64, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x75, 0x75, 0x69, 0x64, 0x12, 0x10, 0x0a, 0x03, 0x6d,
	0x61, 0x63, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6d, 0x61, 0x63, 0x12, 0x1b, 0x0a,
	0x09, 0x61, 0x70, 0x69, 0x5f, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x08, 0x61, 0x70, 0x69, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x22, 0xb3, 0x01, 0x0a, 0x0e, 0x53,
	0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x0e, 0x0a,
	0x02, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x02, 0x6f, 0x6e, 0x12, 0x20, 0x0a,
	0x0b, 0x74, 0x65, 0x6d, 0x70, 0x65, 0x72, 0x61, 0x74, 0x75, 0x72, 0x65, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x05, 0x52, 0x0b, 0x74, 0x65, 0x6d, 0x70, 0x65, 0x72, 0x61, 0x74, 0x75, 0x72, 0x65, 0x12,
	0x12, 0x0a, 0x04, 0x6d, 0x6f, 0x64, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x05, 0x52, 0x04, 0x6d,
	0x6f, 0x64, 0x65, 0x12, 0x1b, 0x0a, 0x09, 0x66, 0x61, 0x6e, 0x5f, 0x73, 0x70, 0x65, 0x65, 0x64,
	0x18, 0x04, 0x20, 0x01, 0x28, 0x05, 0x52, 0x08, 0x66, 0x61, 0x6e, 0x53, 0x70, 0x65, 0x65, 0x64,
	0x12, 0x1d, 0x0a, 0x0a, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x5f, 0x61, 0x74, 0x18, 0x05,
	0x20, 0x01, 0x28, 0x03, 0x52, 0x09, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x41, 0x74, 0x12,
	0x1f, 0x0a, 0x0b, 0x6d, 0x6f, 0x64, 0x69, 0x66, 0x69, 0x65, 0x64, 0x5f, 0x61, 0x74, 0x18, 0x06,
	0x20, 0x01, 0x28, 0x03, 0x52, 0x0a, 0x6d, 0x6f, 0x64, 0x69, 0x66, 0x69, 0x65, 0x64, 0x41, 0x74,
	0x22, 0xc5, 0x01, 0x0a, 0x0d, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02,
	0x69, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x75, 0x75, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x04, 0x75, 0x75, 0x69, 0x64, 0x12, 0x10, 0x0a, 0x03, 0x6d, 0x61, 0x63, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x03, 0x6d, 0x61, 0x63, 0x12, 0x1b, 0x0a, 0x09, 0x61, 0x70, 0x69, 0x5f,
	0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x61, 0x70, 0x69,
	0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x12, 0x0e, 0x0a, 0x02, 0x6f, 0x6e, 0x18, 0x05, 0x20, 0x01, 0x28,
	0x08, 0x52, 0x02, 0x6f, 0x6e, 0x12, 0x20, 0x0a, 0x0b, 0x74, 0x65, 0x6d, 0x70, 0x65, 0x72, 0x61,
	0x74, 0x75, 0x72, 0x65, 0x18, 0x06, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0b, 0x74, 0x65, 0x6d, 0x70,
	0x65, 0x72, 0x61, 0x74, 0x75, 0x72, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x6d, 0x6f, 0x64, 0x65, 0x18,
	0x07, 0x20, 0x01, 0x28, 0x05, 0x52, 0x04, 0x6d, 0x6f, 0x64, 0x65, 0x12, 0x1b, 0x0a, 0x09, 0x66,
	0x61, 0x6e, 0x5f, 0x73, 0x70, 0x65, 0x65, 0x64, 0x18, 0x08, 0x20, 0x01, 0x28, 0x05, 0x52, 0x08,
	0x66, 0x61, 0x6e, 0x53, 0x70, 0x65, 0x65, 0x64, 0x22, 0x42, 0x0a, 0x0e, 0x56, 0x61, 0x6c, 0x75,
	0x65, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x74,
	0x61, 0x74, 0x75, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74,
	0x75, 0x73, 0x12, 0x18, 0x0a, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x32, 0x84, 0x01, 0x0a,
	0x06, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65, 0x12, 0x3c, 0x0a, 0x09, 0x47, 0x65, 0x74, 0x53, 0x74,
	0x61, 0x74, 0x75, 0x73, 0x12, 0x15, 0x2e, 0x64, 0x65, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x53, 0x74,
	0x61, 0x74, 0x75, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x16, 0x2e, 0x64, 0x65,
	0x76, 0x69, 0x63, 0x65, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x3c, 0x0a, 0x09, 0x53, 0x65, 0x74, 0x56, 0x61, 0x6c, 0x75,
	0x65, 0x73, 0x12, 0x15, 0x2e, 0x64, 0x65, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x56, 0x61, 0x6c, 0x75,
	0x65, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x16, 0x2e, 0x64, 0x65, 0x76, 0x69,
	0x63, 0x65, 0x2e, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x22, 0x00, 0x42, 0x33, 0x5a, 0x31, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f,
	0x6d, 0x2f, 0x4b, 0x73, 0x38, 0x39, 0x2f, 0x61, 0x69, 0x72, 0x2d, 0x63, 0x6f, 0x6e, 0x64, 0x69,
	0x74, 0x69, 0x6f, 0x6e, 0x65, 0x72, 0x2f, 0x61, 0x70, 0x69, 0x2d, 0x73, 0x65, 0x72, 0x76, 0x65,
	0x72, 0x2f, 0x64, 0x65, 0x76, 0x69, 0x63, 0x65, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_api_grpc_device_device_proto_rawDescOnce sync.Once
	file_api_grpc_device_device_proto_rawDescData = file_api_grpc_device_device_proto_rawDesc
)

func file_api_grpc_device_device_proto_rawDescGZIP() []byte {
	file_api_grpc_device_device_proto_rawDescOnce.Do(func() {
		file_api_grpc_device_device_proto_rawDescData = protoimpl.X.CompressGZIP(file_api_grpc_device_device_proto_rawDescData)
	})
	return file_api_grpc_device_device_proto_rawDescData
}

var file_api_grpc_device_device_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_api_grpc_device_device_proto_goTypes = []interface{}{
	(*StatusRequest)(nil),  // 0: device.StatusRequest
	(*StatusResponse)(nil), // 1: device.StatusResponse
	(*ValuesRequest)(nil),  // 2: device.ValuesRequest
	(*ValuesResponse)(nil), // 3: device.ValuesResponse
}
var file_api_grpc_device_device_proto_depIdxs = []int32{
	0, // 0: device.Device.GetStatus:input_type -> device.StatusRequest
	2, // 1: device.Device.SetValues:input_type -> device.ValuesRequest
	1, // 2: device.Device.GetStatus:output_type -> device.StatusResponse
	3, // 3: device.Device.SetValues:output_type -> device.ValuesResponse
	2, // [2:4] is the sub-list for method output_type
	0, // [0:2] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_api_grpc_device_device_proto_init() }
func file_api_grpc_device_device_proto_init() {
	if File_api_grpc_device_device_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_api_grpc_device_device_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StatusRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_grpc_device_device_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StatusResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_grpc_device_device_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ValuesRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_grpc_device_device_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ValuesResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_api_grpc_device_device_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_api_grpc_device_device_proto_goTypes,
		DependencyIndexes: file_api_grpc_device_device_proto_depIdxs,
		MessageInfos:      file_api_grpc_device_device_proto_msgTypes,
	}.Build()
	File_api_grpc_device_device_proto = out.File
	file_api_grpc_device_device_proto_rawDesc = nil
	file_api_grpc_device_device_proto_goTypes = nil
	file_api_grpc_device_device_proto_depIdxs = nil
}
