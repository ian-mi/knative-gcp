// Code generated by protoc-gen-go. DO NOT EDIT.
// source: google/logging/type/http_request.proto

package ltype

import (
	fmt "fmt"
	math "math"

	proto "github.com/golang/protobuf/proto"
	duration "github.com/golang/protobuf/ptypes/duration"
	_ "google.golang.org/genproto/googleapis/api/annotations"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

// A common proto for logging HTTP requests. Only contains semantics
// defined by the HTTP specification. Product-specific logging
// information MUST be defined in a separate message.
type HttpRequest struct {
	// The request method. Examples: `"GET"`, `"HEAD"`, `"PUT"`, `"POST"`.
	RequestMethod string `protobuf:"bytes,1,opt,name=request_method,json=requestMethod,proto3" json:"request_method,omitempty"`
	// The scheme (http, https), the host name, the path and the query
	// portion of the URL that was requested.
	// Example: `"http://example.com/some/info?color=red"`.
	RequestUrl string `protobuf:"bytes,2,opt,name=request_url,json=requestUrl,proto3" json:"request_url,omitempty"`
	// The size of the HTTP request message in bytes, including the request
	// headers and the request body.
	RequestSize int64 `protobuf:"varint,3,opt,name=request_size,json=requestSize,proto3" json:"request_size,omitempty"`
	// The response code indicating the status of response.
	// Examples: 200, 404.
	Status int32 `protobuf:"varint,4,opt,name=status,proto3" json:"status,omitempty"`
	// The size of the HTTP response message sent back to the client, in bytes,
	// including the response headers and the response body.
	ResponseSize int64 `protobuf:"varint,5,opt,name=response_size,json=responseSize,proto3" json:"response_size,omitempty"`
	// The user agent sent by the client. Example:
	// `"Mozilla/4.0 (compatible; MSIE 6.0; Windows 98; Q312461; .NET
	// CLR 1.0.3705)"`.
	UserAgent string `protobuf:"bytes,6,opt,name=user_agent,json=userAgent,proto3" json:"user_agent,omitempty"`
	// The IP address (IPv4 or IPv6) of the client that issued the HTTP
	// request. Examples: `"192.168.1.1"`, `"FE80::0202:B3FF:FE1E:8329"`.
	RemoteIp string `protobuf:"bytes,7,opt,name=remote_ip,json=remoteIp,proto3" json:"remote_ip,omitempty"`
	// The IP address (IPv4 or IPv6) of the origin server that the request was
	// sent to.
	ServerIp string `protobuf:"bytes,13,opt,name=server_ip,json=serverIp,proto3" json:"server_ip,omitempty"`
	// The referer URL of the request, as defined in
	// [HTTP/1.1 Header Field
	// Definitions](http://www.w3.org/Protocols/rfc2616/rfc2616-sec14.html).
	Referer string `protobuf:"bytes,8,opt,name=referer,proto3" json:"referer,omitempty"`
	// The request processing latency on the server, from the time the request was
	// received until the response was sent.
	Latency *duration.Duration `protobuf:"bytes,14,opt,name=latency,proto3" json:"latency,omitempty"`
	// Whether or not a cache lookup was attempted.
	CacheLookup bool `protobuf:"varint,11,opt,name=cache_lookup,json=cacheLookup,proto3" json:"cache_lookup,omitempty"`
	// Whether or not an entity was served from cache
	// (with or without validation).
	CacheHit bool `protobuf:"varint,9,opt,name=cache_hit,json=cacheHit,proto3" json:"cache_hit,omitempty"`
	// Whether or not the response was validated with the origin server before
	// being served from cache. This field is only meaningful if `cache_hit` is
	// True.
	CacheValidatedWithOriginServer bool `protobuf:"varint,10,opt,name=cache_validated_with_origin_server,json=cacheValidatedWithOriginServer,proto3" json:"cache_validated_with_origin_server,omitempty"`
	// The number of HTTP response bytes inserted into cache. Set only when a
	// cache fill was attempted.
	CacheFillBytes int64 `protobuf:"varint,12,opt,name=cache_fill_bytes,json=cacheFillBytes,proto3" json:"cache_fill_bytes,omitempty"`
	// Protocol used for the request. Examples: "HTTP/1.1", "HTTP/2", "websocket"
	Protocol             string   `protobuf:"bytes,15,opt,name=protocol,proto3" json:"protocol,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *HttpRequest) Reset()         { *m = HttpRequest{} }
func (m *HttpRequest) String() string { return proto.CompactTextString(m) }
func (*HttpRequest) ProtoMessage()    {}
func (*HttpRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_ea45f4ec7ed7b641, []int{0}
}

func (m *HttpRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_HttpRequest.Unmarshal(m, b)
}
func (m *HttpRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_HttpRequest.Marshal(b, m, deterministic)
}
func (m *HttpRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_HttpRequest.Merge(m, src)
}
func (m *HttpRequest) XXX_Size() int {
	return xxx_messageInfo_HttpRequest.Size(m)
}
func (m *HttpRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_HttpRequest.DiscardUnknown(m)
}

var xxx_messageInfo_HttpRequest proto.InternalMessageInfo

func (m *HttpRequest) GetRequestMethod() string {
	if m != nil {
		return m.RequestMethod
	}
	return ""
}

func (m *HttpRequest) GetRequestUrl() string {
	if m != nil {
		return m.RequestUrl
	}
	return ""
}

func (m *HttpRequest) GetRequestSize() int64 {
	if m != nil {
		return m.RequestSize
	}
	return 0
}

func (m *HttpRequest) GetStatus() int32 {
	if m != nil {
		return m.Status
	}
	return 0
}

func (m *HttpRequest) GetResponseSize() int64 {
	if m != nil {
		return m.ResponseSize
	}
	return 0
}

func (m *HttpRequest) GetUserAgent() string {
	if m != nil {
		return m.UserAgent
	}
	return ""
}

func (m *HttpRequest) GetRemoteIp() string {
	if m != nil {
		return m.RemoteIp
	}
	return ""
}

func (m *HttpRequest) GetServerIp() string {
	if m != nil {
		return m.ServerIp
	}
	return ""
}

func (m *HttpRequest) GetReferer() string {
	if m != nil {
		return m.Referer
	}
	return ""
}

func (m *HttpRequest) GetLatency() *duration.Duration {
	if m != nil {
		return m.Latency
	}
	return nil
}

func (m *HttpRequest) GetCacheLookup() bool {
	if m != nil {
		return m.CacheLookup
	}
	return false
}

func (m *HttpRequest) GetCacheHit() bool {
	if m != nil {
		return m.CacheHit
	}
	return false
}

func (m *HttpRequest) GetCacheValidatedWithOriginServer() bool {
	if m != nil {
		return m.CacheValidatedWithOriginServer
	}
	return false
}

func (m *HttpRequest) GetCacheFillBytes() int64 {
	if m != nil {
		return m.CacheFillBytes
	}
	return 0
}

func (m *HttpRequest) GetProtocol() string {
	if m != nil {
		return m.Protocol
	}
	return ""
}

func init() {
	proto.RegisterType((*HttpRequest)(nil), "google.logging.type.HttpRequest")
}

func init() {
	proto.RegisterFile("google/logging/type/http_request.proto", fileDescriptor_ea45f4ec7ed7b641)
}

var fileDescriptor_ea45f4ec7ed7b641 = []byte{
	// 511 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x6c, 0x92, 0x5b, 0x6b, 0x14, 0x31,
	0x18, 0x86, 0x99, 0x1e, 0xf6, 0x90, 0x3d, 0x58, 0x22, 0x68, 0xba, 0x6a, 0x5d, 0x2b, 0xca, 0x5c,
	0xcd, 0x80, 0xbd, 0x11, 0xbc, 0x72, 0x15, 0x6d, 0xa5, 0x62, 0x99, 0x7a, 0x00, 0x59, 0x18, 0x66,
	0x77, 0xbf, 0x9d, 0x09, 0x66, 0x27, 0x31, 0xc9, 0x54, 0xb6, 0x7f, 0xc6, 0x7b, 0x6f, 0xfc, 0x1f,
	0xfe, 0x2a, 0xc9, 0x97, 0x0c, 0x28, 0xf4, 0x66, 0x21, 0xef, 0xf3, 0xbc, 0x49, 0xf6, 0x9b, 0x90,
	0xa7, 0xa5, 0x94, 0xa5, 0x80, 0x54, 0xc8, 0xb2, 0xe4, 0x75, 0x99, 0xda, 0xad, 0x82, 0xb4, 0xb2,
	0x56, 0xe5, 0x1a, 0xbe, 0x37, 0x60, 0x6c, 0xa2, 0xb4, 0xb4, 0x92, 0xde, 0xf6, 0x5e, 0x12, 0xbc,
	0xc4, 0x79, 0x93, 0xa3, 0x50, 0x46, 0x65, 0xd1, 0xac, 0xd3, 0x55, 0xa3, 0x0b, 0xcb, 0x65, 0xed,
	0x4b, 0x93, 0xfb, 0x81, 0x17, 0x8a, 0xa7, 0x45, 0x5d, 0x4b, 0x8b, 0xd0, 0x78, 0x7a, 0xfc, 0x7b,
	0x8f, 0x0c, 0x4e, 0xad, 0x55, 0x99, 0x3f, 0x88, 0x3e, 0x21, 0xe3, 0x70, 0x66, 0xbe, 0x01, 0x5b,
	0xc9, 0x15, 0x8b, 0xa6, 0x51, 0xdc, 0xcf, 0x46, 0x21, 0x7d, 0x8f, 0x21, 0x7d, 0x48, 0x06, 0xad,
	0xd6, 0x68, 0xc1, 0x76, 0xd0, 0x21, 0x21, 0xfa, 0xa4, 0x05, 0x7d, 0x44, 0x86, 0xad, 0x60, 0xf8,
	0x35, 0xb0, 0xdd, 0x69, 0x14, 0xef, 0x66, 0x6d, 0xe9, 0x92, 0x5f, 0x03, 0xbd, 0x43, 0x3a, 0xc6,
	0x16, 0xb6, 0x31, 0x6c, 0x6f, 0x1a, 0xc5, 0xfb, 0x59, 0x58, 0xd1, 0xc7, 0x64, 0xa4, 0xc1, 0x28,
	0x59, 0x1b, 0xf0, 0xdd, 0x7d, 0xec, 0x0e, 0xdb, 0x10, 0xcb, 0x0f, 0x08, 0x69, 0x0c, 0xe8, 0xbc,
	0x28, 0xa1, 0xb6, 0xac, 0x83, 0xe7, 0xf7, 0x5d, 0xf2, 0xd2, 0x05, 0xf4, 0x1e, 0xe9, 0x6b, 0xd8,
	0x48, 0x0b, 0x39, 0x57, 0xac, 0x8b, 0xb4, 0xe7, 0x83, 0x33, 0xe5, 0xa0, 0x01, 0x7d, 0x05, 0xda,
	0xc1, 0x91, 0x87, 0x3e, 0x38, 0x53, 0x94, 0x91, 0xae, 0x86, 0x35, 0x68, 0xd0, 0xac, 0x87, 0xa8,
	0x5d, 0xd2, 0x13, 0xd2, 0x15, 0x85, 0x85, 0x7a, 0xb9, 0x65, 0xe3, 0x69, 0x14, 0x0f, 0x9e, 0x1d,
	0x26, 0xe1, 0x7b, 0xb4, 0xa3, 0x4f, 0x5e, 0x87, 0xd1, 0x67, 0xad, 0xe9, 0xe6, 0xb0, 0x2c, 0x96,
	0x15, 0xe4, 0x42, 0xca, 0x6f, 0x8d, 0x62, 0x83, 0x69, 0x14, 0xf7, 0xb2, 0x01, 0x66, 0xe7, 0x18,
	0xb9, 0xeb, 0x78, 0xa5, 0xe2, 0x96, 0xf5, 0x91, 0xf7, 0x30, 0x38, 0xe5, 0x96, 0xbe, 0x23, 0xc7,
	0x1e, 0x5e, 0x15, 0x82, 0xaf, 0x0a, 0x0b, 0xab, 0xfc, 0x07, 0xb7, 0x55, 0x2e, 0x35, 0x2f, 0x79,
	0x9d, 0xfb, 0x6b, 0x33, 0x82, 0xad, 0x23, 0x34, 0x3f, 0xb7, 0xe2, 0x17, 0x6e, 0xab, 0x0f, 0xa8,
	0x5d, 0xa2, 0x45, 0x63, 0x72, 0xe0, 0xf7, 0x5a, 0x73, 0x21, 0xf2, 0xc5, 0xd6, 0x82, 0x61, 0x43,
	0x9c, 0xed, 0x18, 0xf3, 0x37, 0x5c, 0x88, 0x99, 0x4b, 0xe9, 0x84, 0xf4, 0xf0, 0x3f, 0x2d, 0xa5,
	0x60, 0xb7, 0xfc, 0x80, 0xda, 0xf5, 0xec, 0x67, 0x44, 0xee, 0x2e, 0xe5, 0x26, 0xb9, 0xe1, 0x2d,
	0xce, 0x0e, 0xfe, 0x79, 0x4a, 0x17, 0xae, 0x70, 0x11, 0x7d, 0x7d, 0x1e, 0xc4, 0x52, 0x8a, 0xa2,
	0x2e, 0x13, 0xa9, 0xcb, 0xb4, 0x84, 0x1a, 0xb7, 0x4b, 0x3d, 0x2a, 0x14, 0x37, 0xff, 0xbd, 0xfd,
	0x17, 0xc2, 0xfd, 0xfe, 0xda, 0x39, 0x7c, 0xeb, 0xab, 0xaf, 0x84, 0x6c, 0x56, 0xc9, 0x79, 0x38,
	0xe9, 0xe3, 0x56, 0xc1, 0x9f, 0x96, 0xcd, 0x91, 0xcd, 0x03, 0x9b, 0x3b, 0xb6, 0xe8, 0xe0, 0xe6,
	0x27, 0x7f, 0x03, 0x00, 0x00, 0xff, 0xff, 0x43, 0x25, 0x30, 0x14, 0x57, 0x03, 0x00, 0x00,
}
