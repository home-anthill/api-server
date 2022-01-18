// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.17.3
// source: api/device/device.proto

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

	Id           string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Uuid         string `protobuf:"bytes,2,opt,name=uuid,proto3" json:"uuid,omitempty"`
	Mac          string `protobuf:"bytes,3,opt,name=mac,proto3" json:"mac,omitempty"`
	ProfileToken string `protobuf:"bytes,4,opt,name=profile_token,json=profileToken,proto3" json:"profile_token,omitempty"`
}

func (x *StatusRequest) Reset() {
	*x = StatusRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_device_device_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StatusRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StatusRequest) ProtoMessage() {}

func (x *StatusRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_device_device_proto_msgTypes[0]
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
	return file_api_device_device_proto_rawDescGZIP(), []int{0}
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

func (x *StatusRequest) GetProfileToken() string {
	if x != nil {
		return x.ProfileToken
	}
	return ""
}

type OnOffValueRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id           string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Uuid         string `protobuf:"bytes,2,opt,name=uuid,proto3" json:"uuid,omitempty"`
	Mac          string `protobuf:"bytes,3,opt,name=mac,proto3" json:"mac,omitempty"`
	ProfileToken string `protobuf:"bytes,4,opt,name=profile_token,json=profileToken,proto3" json:"profile_token,omitempty"`
	On           bool   `protobuf:"varint,5,opt,name=on,proto3" json:"on,omitempty"`
}

func (x *OnOffValueRequest) Reset() {
	*x = OnOffValueRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_device_device_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *OnOffValueRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*OnOffValueRequest) ProtoMessage() {}

func (x *OnOffValueRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_device_device_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use OnOffValueRequest.ProtoReflect.Descriptor instead.
func (*OnOffValueRequest) Descriptor() ([]byte, []int) {
	return file_api_device_device_proto_rawDescGZIP(), []int{1}
}

func (x *OnOffValueRequest) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *OnOffValueRequest) GetUuid() string {
	if x != nil {
		return x.Uuid
	}
	return ""
}

func (x *OnOffValueRequest) GetMac() string {
	if x != nil {
		return x.Mac
	}
	return ""
}

func (x *OnOffValueRequest) GetProfileToken() string {
	if x != nil {
		return x.ProfileToken
	}
	return ""
}

func (x *OnOffValueRequest) GetOn() bool {
	if x != nil {
		return x.On
	}
	return false
}

type TemperatureValueRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id           string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Uuid         string `protobuf:"bytes,2,opt,name=uuid,proto3" json:"uuid,omitempty"`
	Mac          string `protobuf:"bytes,3,opt,name=mac,proto3" json:"mac,omitempty"`
	ProfileToken string `protobuf:"bytes,4,opt,name=profile_token,json=profileToken,proto3" json:"profile_token,omitempty"`
	Temperature  int32  `protobuf:"varint,5,opt,name=temperature,proto3" json:"temperature,omitempty"`
}

func (x *TemperatureValueRequest) Reset() {
	*x = TemperatureValueRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_device_device_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TemperatureValueRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TemperatureValueRequest) ProtoMessage() {}

func (x *TemperatureValueRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_device_device_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TemperatureValueRequest.ProtoReflect.Descriptor instead.
func (*TemperatureValueRequest) Descriptor() ([]byte, []int) {
	return file_api_device_device_proto_rawDescGZIP(), []int{2}
}

func (x *TemperatureValueRequest) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *TemperatureValueRequest) GetUuid() string {
	if x != nil {
		return x.Uuid
	}
	return ""
}

func (x *TemperatureValueRequest) GetMac() string {
	if x != nil {
		return x.Mac
	}
	return ""
}

func (x *TemperatureValueRequest) GetProfileToken() string {
	if x != nil {
		return x.ProfileToken
	}
	return ""
}

func (x *TemperatureValueRequest) GetTemperature() int32 {
	if x != nil {
		return x.Temperature
	}
	return 0
}

type ModeValueRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id           string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Uuid         string `protobuf:"bytes,2,opt,name=uuid,proto3" json:"uuid,omitempty"`
	Mac          string `protobuf:"bytes,3,opt,name=mac,proto3" json:"mac,omitempty"`
	ProfileToken string `protobuf:"bytes,4,opt,name=profile_token,json=profileToken,proto3" json:"profile_token,omitempty"`
	Mode         int32  `protobuf:"varint,5,opt,name=mode,proto3" json:"mode,omitempty"`
}

func (x *ModeValueRequest) Reset() {
	*x = ModeValueRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_device_device_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ModeValueRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ModeValueRequest) ProtoMessage() {}

func (x *ModeValueRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_device_device_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ModeValueRequest.ProtoReflect.Descriptor instead.
func (*ModeValueRequest) Descriptor() ([]byte, []int) {
	return file_api_device_device_proto_rawDescGZIP(), []int{3}
}

func (x *ModeValueRequest) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *ModeValueRequest) GetUuid() string {
	if x != nil {
		return x.Uuid
	}
	return ""
}

func (x *ModeValueRequest) GetMac() string {
	if x != nil {
		return x.Mac
	}
	return ""
}

func (x *ModeValueRequest) GetProfileToken() string {
	if x != nil {
		return x.ProfileToken
	}
	return ""
}

func (x *ModeValueRequest) GetMode() int32 {
	if x != nil {
		return x.Mode
	}
	return 0
}

type FanModeValueRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id           string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Uuid         string `protobuf:"bytes,2,opt,name=uuid,proto3" json:"uuid,omitempty"`
	Mac          string `protobuf:"bytes,3,opt,name=mac,proto3" json:"mac,omitempty"`
	ProfileToken string `protobuf:"bytes,4,opt,name=profile_token,json=profileToken,proto3" json:"profile_token,omitempty"`
	FanMode      int32  `protobuf:"varint,5,opt,name=fan_mode,json=fanMode,proto3" json:"fan_mode,omitempty"`
}

func (x *FanModeValueRequest) Reset() {
	*x = FanModeValueRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_device_device_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FanModeValueRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FanModeValueRequest) ProtoMessage() {}

func (x *FanModeValueRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_device_device_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FanModeValueRequest.ProtoReflect.Descriptor instead.
func (*FanModeValueRequest) Descriptor() ([]byte, []int) {
	return file_api_device_device_proto_rawDescGZIP(), []int{4}
}

func (x *FanModeValueRequest) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *FanModeValueRequest) GetUuid() string {
	if x != nil {
		return x.Uuid
	}
	return ""
}

func (x *FanModeValueRequest) GetMac() string {
	if x != nil {
		return x.Mac
	}
	return ""
}

func (x *FanModeValueRequest) GetProfileToken() string {
	if x != nil {
		return x.ProfileToken
	}
	return ""
}

func (x *FanModeValueRequest) GetFanMode() int32 {
	if x != nil {
		return x.FanMode
	}
	return 0
}

type FanSpeedValueRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id           string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Uuid         string `protobuf:"bytes,2,opt,name=uuid,proto3" json:"uuid,omitempty"`
	Mac          string `protobuf:"bytes,3,opt,name=mac,proto3" json:"mac,omitempty"`
	ProfileToken string `protobuf:"bytes,4,opt,name=profile_token,json=profileToken,proto3" json:"profile_token,omitempty"`
	FanSpeed     int32  `protobuf:"varint,5,opt,name=fan_speed,json=fanSpeed,proto3" json:"fan_speed,omitempty"`
}

func (x *FanSpeedValueRequest) Reset() {
	*x = FanSpeedValueRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_device_device_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FanSpeedValueRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FanSpeedValueRequest) ProtoMessage() {}

func (x *FanSpeedValueRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_device_device_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FanSpeedValueRequest.ProtoReflect.Descriptor instead.
func (*FanSpeedValueRequest) Descriptor() ([]byte, []int) {
	return file_api_device_device_proto_rawDescGZIP(), []int{5}
}

func (x *FanSpeedValueRequest) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *FanSpeedValueRequest) GetUuid() string {
	if x != nil {
		return x.Uuid
	}
	return ""
}

func (x *FanSpeedValueRequest) GetMac() string {
	if x != nil {
		return x.Mac
	}
	return ""
}

func (x *FanSpeedValueRequest) GetProfileToken() string {
	if x != nil {
		return x.ProfileToken
	}
	return ""
}

func (x *FanSpeedValueRequest) GetFanSpeed() int32 {
	if x != nil {
		return x.FanSpeed
	}
	return 0
}

type StatusResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	On          bool  `protobuf:"varint,1,opt,name=on,proto3" json:"on,omitempty"`
	Temperature int32 `protobuf:"varint,2,opt,name=temperature,proto3" json:"temperature,omitempty"`
	Mode        int32 `protobuf:"varint,3,opt,name=mode,proto3" json:"mode,omitempty"`
	FanMode     int32 `protobuf:"varint,4,opt,name=fan_mode,json=fanMode,proto3" json:"fan_mode,omitempty"`
	FanSpeed    int32 `protobuf:"varint,5,opt,name=fan_speed,json=fanSpeed,proto3" json:"fan_speed,omitempty"`
}

func (x *StatusResponse) Reset() {
	*x = StatusResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_device_device_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StatusResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StatusResponse) ProtoMessage() {}

func (x *StatusResponse) ProtoReflect() protoreflect.Message {
	mi := &file_api_device_device_proto_msgTypes[6]
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
	return file_api_device_device_proto_rawDescGZIP(), []int{6}
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

func (x *StatusResponse) GetFanMode() int32 {
	if x != nil {
		return x.FanMode
	}
	return 0
}

func (x *StatusResponse) GetFanSpeed() int32 {
	if x != nil {
		return x.FanSpeed
	}
	return 0
}

type OnOffValueResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Status  string `protobuf:"bytes,1,opt,name=status,proto3" json:"status,omitempty"`
	Message string `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
}

func (x *OnOffValueResponse) Reset() {
	*x = OnOffValueResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_device_device_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *OnOffValueResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*OnOffValueResponse) ProtoMessage() {}

func (x *OnOffValueResponse) ProtoReflect() protoreflect.Message {
	mi := &file_api_device_device_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use OnOffValueResponse.ProtoReflect.Descriptor instead.
func (*OnOffValueResponse) Descriptor() ([]byte, []int) {
	return file_api_device_device_proto_rawDescGZIP(), []int{7}
}

func (x *OnOffValueResponse) GetStatus() string {
	if x != nil {
		return x.Status
	}
	return ""
}

func (x *OnOffValueResponse) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

type TemperatureValueResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Status  string `protobuf:"bytes,1,opt,name=status,proto3" json:"status,omitempty"`
	Message string `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
}

func (x *TemperatureValueResponse) Reset() {
	*x = TemperatureValueResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_device_device_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TemperatureValueResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TemperatureValueResponse) ProtoMessage() {}

func (x *TemperatureValueResponse) ProtoReflect() protoreflect.Message {
	mi := &file_api_device_device_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TemperatureValueResponse.ProtoReflect.Descriptor instead.
func (*TemperatureValueResponse) Descriptor() ([]byte, []int) {
	return file_api_device_device_proto_rawDescGZIP(), []int{8}
}

func (x *TemperatureValueResponse) GetStatus() string {
	if x != nil {
		return x.Status
	}
	return ""
}

func (x *TemperatureValueResponse) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

type ModeValueResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Status  string `protobuf:"bytes,1,opt,name=status,proto3" json:"status,omitempty"`
	Message string `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
}

func (x *ModeValueResponse) Reset() {
	*x = ModeValueResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_device_device_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ModeValueResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ModeValueResponse) ProtoMessage() {}

func (x *ModeValueResponse) ProtoReflect() protoreflect.Message {
	mi := &file_api_device_device_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ModeValueResponse.ProtoReflect.Descriptor instead.
func (*ModeValueResponse) Descriptor() ([]byte, []int) {
	return file_api_device_device_proto_rawDescGZIP(), []int{9}
}

func (x *ModeValueResponse) GetStatus() string {
	if x != nil {
		return x.Status
	}
	return ""
}

func (x *ModeValueResponse) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

type FanModeValueResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Status  string `protobuf:"bytes,1,opt,name=status,proto3" json:"status,omitempty"`
	Message string `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
}

func (x *FanModeValueResponse) Reset() {
	*x = FanModeValueResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_device_device_proto_msgTypes[10]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FanModeValueResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FanModeValueResponse) ProtoMessage() {}

func (x *FanModeValueResponse) ProtoReflect() protoreflect.Message {
	mi := &file_api_device_device_proto_msgTypes[10]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FanModeValueResponse.ProtoReflect.Descriptor instead.
func (*FanModeValueResponse) Descriptor() ([]byte, []int) {
	return file_api_device_device_proto_rawDescGZIP(), []int{10}
}

func (x *FanModeValueResponse) GetStatus() string {
	if x != nil {
		return x.Status
	}
	return ""
}

func (x *FanModeValueResponse) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

type FanSpeedValueResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Status  string `protobuf:"bytes,1,opt,name=status,proto3" json:"status,omitempty"`
	Message string `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
}

func (x *FanSpeedValueResponse) Reset() {
	*x = FanSpeedValueResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_device_device_proto_msgTypes[11]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FanSpeedValueResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FanSpeedValueResponse) ProtoMessage() {}

func (x *FanSpeedValueResponse) ProtoReflect() protoreflect.Message {
	mi := &file_api_device_device_proto_msgTypes[11]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FanSpeedValueResponse.ProtoReflect.Descriptor instead.
func (*FanSpeedValueResponse) Descriptor() ([]byte, []int) {
	return file_api_device_device_proto_rawDescGZIP(), []int{11}
}

func (x *FanSpeedValueResponse) GetStatus() string {
	if x != nil {
		return x.Status
	}
	return ""
}

func (x *FanSpeedValueResponse) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

var File_api_device_device_proto protoreflect.FileDescriptor

var file_api_device_device_proto_rawDesc = []byte{
	0x0a, 0x17, 0x61, 0x70, 0x69, 0x2f, 0x64, 0x65, 0x76, 0x69, 0x63, 0x65, 0x2f, 0x64, 0x65, 0x76,
	0x69, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x06, 0x64, 0x65, 0x76, 0x69, 0x63,
	0x65, 0x22, 0x6a, 0x0a, 0x0d, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02,
	0x69, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x75, 0x75, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x04, 0x75, 0x75, 0x69, 0x64, 0x12, 0x10, 0x0a, 0x03, 0x6d, 0x61, 0x63, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x03, 0x6d, 0x61, 0x63, 0x12, 0x23, 0x0a, 0x0d, 0x70, 0x72, 0x6f, 0x66,
	0x69, 0x6c, 0x65, 0x5f, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x0c, 0x70, 0x72, 0x6f, 0x66, 0x69, 0x6c, 0x65, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x22, 0x7e, 0x0a,
	0x11, 0x4f, 0x6e, 0x4f, 0x66, 0x66, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02,
	0x69, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x75, 0x75, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x04, 0x75, 0x75, 0x69, 0x64, 0x12, 0x10, 0x0a, 0x03, 0x6d, 0x61, 0x63, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x03, 0x6d, 0x61, 0x63, 0x12, 0x23, 0x0a, 0x0d, 0x70, 0x72, 0x6f, 0x66,
	0x69, 0x6c, 0x65, 0x5f, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x0c, 0x70, 0x72, 0x6f, 0x66, 0x69, 0x6c, 0x65, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x12, 0x0e, 0x0a,
	0x02, 0x6f, 0x6e, 0x18, 0x05, 0x20, 0x01, 0x28, 0x08, 0x52, 0x02, 0x6f, 0x6e, 0x22, 0x96, 0x01,
	0x0a, 0x17, 0x54, 0x65, 0x6d, 0x70, 0x65, 0x72, 0x61, 0x74, 0x75, 0x72, 0x65, 0x56, 0x61, 0x6c,
	0x75, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x75, 0x75, 0x69,
	0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x75, 0x75, 0x69, 0x64, 0x12, 0x10, 0x0a,
	0x03, 0x6d, 0x61, 0x63, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6d, 0x61, 0x63, 0x12,
	0x23, 0x0a, 0x0d, 0x70, 0x72, 0x6f, 0x66, 0x69, 0x6c, 0x65, 0x5f, 0x74, 0x6f, 0x6b, 0x65, 0x6e,
	0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x70, 0x72, 0x6f, 0x66, 0x69, 0x6c, 0x65, 0x54,
	0x6f, 0x6b, 0x65, 0x6e, 0x12, 0x20, 0x0a, 0x0b, 0x74, 0x65, 0x6d, 0x70, 0x65, 0x72, 0x61, 0x74,
	0x75, 0x72, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0b, 0x74, 0x65, 0x6d, 0x70, 0x65,
	0x72, 0x61, 0x74, 0x75, 0x72, 0x65, 0x22, 0x81, 0x01, 0x0a, 0x10, 0x4d, 0x6f, 0x64, 0x65, 0x56,
	0x61, 0x6c, 0x75, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x0e, 0x0a, 0x02, 0x69,
	0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x75,
	0x75, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x75, 0x75, 0x69, 0x64, 0x12,
	0x10, 0x0a, 0x03, 0x6d, 0x61, 0x63, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6d, 0x61,
	0x63, 0x12, 0x23, 0x0a, 0x0d, 0x70, 0x72, 0x6f, 0x66, 0x69, 0x6c, 0x65, 0x5f, 0x74, 0x6f, 0x6b,
	0x65, 0x6e, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x70, 0x72, 0x6f, 0x66, 0x69, 0x6c,
	0x65, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x12, 0x12, 0x0a, 0x04, 0x6d, 0x6f, 0x64, 0x65, 0x18, 0x05,
	0x20, 0x01, 0x28, 0x05, 0x52, 0x04, 0x6d, 0x6f, 0x64, 0x65, 0x22, 0x8b, 0x01, 0x0a, 0x13, 0x46,
	0x61, 0x6e, 0x4d, 0x6f, 0x64, 0x65, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02,
	0x69, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x75, 0x75, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x04, 0x75, 0x75, 0x69, 0x64, 0x12, 0x10, 0x0a, 0x03, 0x6d, 0x61, 0x63, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x03, 0x6d, 0x61, 0x63, 0x12, 0x23, 0x0a, 0x0d, 0x70, 0x72, 0x6f, 0x66,
	0x69, 0x6c, 0x65, 0x5f, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x0c, 0x70, 0x72, 0x6f, 0x66, 0x69, 0x6c, 0x65, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x12, 0x19, 0x0a,
	0x08, 0x66, 0x61, 0x6e, 0x5f, 0x6d, 0x6f, 0x64, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x05, 0x52,
	0x07, 0x66, 0x61, 0x6e, 0x4d, 0x6f, 0x64, 0x65, 0x22, 0x8e, 0x01, 0x0a, 0x14, 0x46, 0x61, 0x6e,
	0x53, 0x70, 0x65, 0x65, 0x64, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69,
	0x64, 0x12, 0x12, 0x0a, 0x04, 0x75, 0x75, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x04, 0x75, 0x75, 0x69, 0x64, 0x12, 0x10, 0x0a, 0x03, 0x6d, 0x61, 0x63, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x03, 0x6d, 0x61, 0x63, 0x12, 0x23, 0x0a, 0x0d, 0x70, 0x72, 0x6f, 0x66, 0x69,
	0x6c, 0x65, 0x5f, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c,
	0x70, 0x72, 0x6f, 0x66, 0x69, 0x6c, 0x65, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x12, 0x1b, 0x0a, 0x09,
	0x66, 0x61, 0x6e, 0x5f, 0x73, 0x70, 0x65, 0x65, 0x64, 0x18, 0x05, 0x20, 0x01, 0x28, 0x05, 0x52,
	0x08, 0x66, 0x61, 0x6e, 0x53, 0x70, 0x65, 0x65, 0x64, 0x22, 0x8e, 0x01, 0x0a, 0x0e, 0x53, 0x74,
	0x61, 0x74, 0x75, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x0e, 0x0a, 0x02,
	0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x02, 0x6f, 0x6e, 0x12, 0x20, 0x0a, 0x0b,
	0x74, 0x65, 0x6d, 0x70, 0x65, 0x72, 0x61, 0x74, 0x75, 0x72, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x05, 0x52, 0x0b, 0x74, 0x65, 0x6d, 0x70, 0x65, 0x72, 0x61, 0x74, 0x75, 0x72, 0x65, 0x12, 0x12,
	0x0a, 0x04, 0x6d, 0x6f, 0x64, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x05, 0x52, 0x04, 0x6d, 0x6f,
	0x64, 0x65, 0x12, 0x19, 0x0a, 0x08, 0x66, 0x61, 0x6e, 0x5f, 0x6d, 0x6f, 0x64, 0x65, 0x18, 0x04,
	0x20, 0x01, 0x28, 0x05, 0x52, 0x07, 0x66, 0x61, 0x6e, 0x4d, 0x6f, 0x64, 0x65, 0x12, 0x1b, 0x0a,
	0x09, 0x66, 0x61, 0x6e, 0x5f, 0x73, 0x70, 0x65, 0x65, 0x64, 0x18, 0x05, 0x20, 0x01, 0x28, 0x05,
	0x52, 0x08, 0x66, 0x61, 0x6e, 0x53, 0x70, 0x65, 0x65, 0x64, 0x22, 0x46, 0x0a, 0x12, 0x4f, 0x6e,
	0x4f, 0x66, 0x66, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x12, 0x16, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x18, 0x0a, 0x07, 0x6d, 0x65, 0x73, 0x73,
	0x61, 0x67, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61,
	0x67, 0x65, 0x22, 0x4c, 0x0a, 0x18, 0x54, 0x65, 0x6d, 0x70, 0x65, 0x72, 0x61, 0x74, 0x75, 0x72,
	0x65, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x16,
	0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06,
	0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x18, 0x0a, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67,
	0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65,
	0x22, 0x45, 0x0a, 0x11, 0x4d, 0x6f, 0x64, 0x65, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x18, 0x0a,
	0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07,
	0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x22, 0x48, 0x0a, 0x14, 0x46, 0x61, 0x6e, 0x4d, 0x6f,
	0x64, 0x65, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12,
	0x16, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x18, 0x0a, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61,
	0x67, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67,
	0x65, 0x22, 0x49, 0x0a, 0x15, 0x46, 0x61, 0x6e, 0x53, 0x70, 0x65, 0x65, 0x64, 0x56, 0x61, 0x6c,
	0x75, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x74,
	0x61, 0x74, 0x75, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74,
	0x75, 0x73, 0x12, 0x18, 0x0a, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x32, 0xbd, 0x03, 0x0a,
	0x06, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65, 0x12, 0x3c, 0x0a, 0x09, 0x47, 0x65, 0x74, 0x53, 0x74,
	0x61, 0x74, 0x75, 0x73, 0x12, 0x15, 0x2e, 0x64, 0x65, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x53, 0x74,
	0x61, 0x74, 0x75, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x16, 0x2e, 0x64, 0x65,
	0x76, 0x69, 0x63, 0x65, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x43, 0x0a, 0x08, 0x53, 0x65, 0x74, 0x4f, 0x6e, 0x4f, 0x66,
	0x66, 0x12, 0x19, 0x2e, 0x64, 0x65, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x4f, 0x6e, 0x4f, 0x66, 0x66,
	0x56, 0x61, 0x6c, 0x75, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1a, 0x2e, 0x64,
	0x65, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x4f, 0x6e, 0x4f, 0x66, 0x66, 0x56, 0x61, 0x6c, 0x75, 0x65,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x55, 0x0a, 0x0e, 0x53, 0x65,
	0x74, 0x54, 0x65, 0x6d, 0x70, 0x65, 0x72, 0x61, 0x74, 0x75, 0x72, 0x65, 0x12, 0x1f, 0x2e, 0x64,
	0x65, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x54, 0x65, 0x6d, 0x70, 0x65, 0x72, 0x61, 0x74, 0x75, 0x72,
	0x65, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x20, 0x2e,
	0x64, 0x65, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x54, 0x65, 0x6d, 0x70, 0x65, 0x72, 0x61, 0x74, 0x75,
	0x72, 0x65, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22,
	0x00, 0x12, 0x40, 0x0a, 0x07, 0x53, 0x65, 0x74, 0x4d, 0x6f, 0x64, 0x65, 0x12, 0x18, 0x2e, 0x64,
	0x65, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x4d, 0x6f, 0x64, 0x65, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x19, 0x2e, 0x64, 0x65, 0x76, 0x69, 0x63, 0x65, 0x2e,
	0x4d, 0x6f, 0x64, 0x65, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x22, 0x00, 0x12, 0x49, 0x0a, 0x0a, 0x53, 0x65, 0x74, 0x46, 0x61, 0x6e, 0x4d, 0x6f, 0x64,
	0x65, 0x12, 0x1b, 0x2e, 0x64, 0x65, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x46, 0x61, 0x6e, 0x4d, 0x6f,
	0x64, 0x65, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1c,
	0x2e, 0x64, 0x65, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x46, 0x61, 0x6e, 0x4d, 0x6f, 0x64, 0x65, 0x56,
	0x61, 0x6c, 0x75, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x4c,
	0x0a, 0x0b, 0x53, 0x65, 0x74, 0x46, 0x61, 0x6e, 0x53, 0x70, 0x65, 0x65, 0x64, 0x12, 0x1c, 0x2e,
	0x64, 0x65, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x46, 0x61, 0x6e, 0x53, 0x70, 0x65, 0x65, 0x64, 0x56,
	0x61, 0x6c, 0x75, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1d, 0x2e, 0x64, 0x65,
	0x76, 0x69, 0x63, 0x65, 0x2e, 0x46, 0x61, 0x6e, 0x53, 0x70, 0x65, 0x65, 0x64, 0x56, 0x61, 0x6c,
	0x75, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x42, 0x34, 0x5a, 0x32,
	0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x4b, 0x73, 0x38, 0x39, 0x2f,
	0x61, 0x69, 0x72, 0x2d, 0x63, 0x6f, 0x6e, 0x64, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x65, 0x72, 0x2f,
	0x61, 0x70, 0x69, 0x2d, 0x64, 0x65, 0x76, 0x69, 0x63, 0x65, 0x73, 0x2f, 0x64, 0x65, 0x76, 0x69,
	0x63, 0x65, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_api_device_device_proto_rawDescOnce sync.Once
	file_api_device_device_proto_rawDescData = file_api_device_device_proto_rawDesc
)

func file_api_device_device_proto_rawDescGZIP() []byte {
	file_api_device_device_proto_rawDescOnce.Do(func() {
		file_api_device_device_proto_rawDescData = protoimpl.X.CompressGZIP(file_api_device_device_proto_rawDescData)
	})
	return file_api_device_device_proto_rawDescData
}

var file_api_device_device_proto_msgTypes = make([]protoimpl.MessageInfo, 12)
var file_api_device_device_proto_goTypes = []interface{}{
	(*StatusRequest)(nil),            // 0: device.StatusRequest
	(*OnOffValueRequest)(nil),        // 1: device.OnOffValueRequest
	(*TemperatureValueRequest)(nil),  // 2: device.TemperatureValueRequest
	(*ModeValueRequest)(nil),         // 3: device.ModeValueRequest
	(*FanModeValueRequest)(nil),      // 4: device.FanModeValueRequest
	(*FanSpeedValueRequest)(nil),     // 5: device.FanSpeedValueRequest
	(*StatusResponse)(nil),           // 6: device.StatusResponse
	(*OnOffValueResponse)(nil),       // 7: device.OnOffValueResponse
	(*TemperatureValueResponse)(nil), // 8: device.TemperatureValueResponse
	(*ModeValueResponse)(nil),        // 9: device.ModeValueResponse
	(*FanModeValueResponse)(nil),     // 10: device.FanModeValueResponse
	(*FanSpeedValueResponse)(nil),    // 11: device.FanSpeedValueResponse
}
var file_api_device_device_proto_depIdxs = []int32{
	0,  // 0: device.Device.GetStatus:input_type -> device.StatusRequest
	1,  // 1: device.Device.SetOnOff:input_type -> device.OnOffValueRequest
	2,  // 2: device.Device.SetTemperature:input_type -> device.TemperatureValueRequest
	3,  // 3: device.Device.SetMode:input_type -> device.ModeValueRequest
	4,  // 4: device.Device.SetFanMode:input_type -> device.FanModeValueRequest
	5,  // 5: device.Device.SetFanSpeed:input_type -> device.FanSpeedValueRequest
	6,  // 6: device.Device.GetStatus:output_type -> device.StatusResponse
	7,  // 7: device.Device.SetOnOff:output_type -> device.OnOffValueResponse
	8,  // 8: device.Device.SetTemperature:output_type -> device.TemperatureValueResponse
	9,  // 9: device.Device.SetMode:output_type -> device.ModeValueResponse
	10, // 10: device.Device.SetFanMode:output_type -> device.FanModeValueResponse
	11, // 11: device.Device.SetFanSpeed:output_type -> device.FanSpeedValueResponse
	6,  // [6:12] is the sub-list for method output_type
	0,  // [0:6] is the sub-list for method input_type
	0,  // [0:0] is the sub-list for extension type_name
	0,  // [0:0] is the sub-list for extension extendee
	0,  // [0:0] is the sub-list for field type_name
}

func init() { file_api_device_device_proto_init() }
func file_api_device_device_proto_init() {
	if File_api_device_device_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_api_device_device_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
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
		file_api_device_device_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*OnOffValueRequest); i {
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
		file_api_device_device_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TemperatureValueRequest); i {
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
		file_api_device_device_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ModeValueRequest); i {
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
		file_api_device_device_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FanModeValueRequest); i {
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
		file_api_device_device_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FanSpeedValueRequest); i {
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
		file_api_device_device_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
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
		file_api_device_device_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*OnOffValueResponse); i {
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
		file_api_device_device_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TemperatureValueResponse); i {
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
		file_api_device_device_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ModeValueResponse); i {
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
		file_api_device_device_proto_msgTypes[10].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FanModeValueResponse); i {
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
		file_api_device_device_proto_msgTypes[11].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FanSpeedValueResponse); i {
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
			RawDescriptor: file_api_device_device_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   12,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_api_device_device_proto_goTypes,
		DependencyIndexes: file_api_device_device_proto_depIdxs,
		MessageInfos:      file_api_device_device_proto_msgTypes,
	}.Build()
	File_api_device_device_proto = out.File
	file_api_device_device_proto_rawDesc = nil
	file_api_device_device_proto_goTypes = nil
	file_api_device_device_proto_depIdxs = nil
}