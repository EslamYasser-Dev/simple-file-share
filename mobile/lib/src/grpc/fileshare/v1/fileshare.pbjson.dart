// This is a generated file - do not edit.
//
// Generated from fileshare/v1/fileshare.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports
// ignore_for_file: unused_import

import 'dart:convert' as $convert;
import 'dart:core' as $core;
import 'dart:typed_data' as $typed_data;

@$core.Deprecated('Use userDescriptor instead')
const User$json = {
  '1': 'User',
  '2': [
    {'1': 'username', '3': 1, '4': 1, '5': 9, '10': 'username'},
    {'1': 'is_admin', '3': 2, '4': 1, '5': 8, '10': 'isAdmin'},
    {'1': 'role', '3': 3, '4': 1, '5': 9, '10': 'role'},
    {'1': 'enabled', '3': 4, '4': 1, '5': 8, '10': 'enabled'},
    {'1': 'quota_bytes', '3': 5, '4': 1, '5': 3, '10': 'quotaBytes'},
    {'1': 'size', '3': 6, '4': 1, '5': 3, '10': 'size'},
    {'1': 'files', '3': 7, '4': 1, '5': 3, '10': 'files'},
    {'1': 'permissions', '3': 8, '4': 3, '5': 9, '10': 'permissions'},
    {'1': 'created_at', '3': 9, '4': 1, '5': 9, '10': 'createdAt'},
  ],
};

/// Descriptor for `User`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List userDescriptor = $convert.base64Decode(
    'CgRVc2VyEhoKCHVzZXJuYW1lGAEgASgJUgh1c2VybmFtZRIZCghpc19hZG1pbhgCIAEoCFIHaX'
    'NBZG1pbhISCgRyb2xlGAMgASgJUgRyb2xlEhgKB2VuYWJsZWQYBCABKAhSB2VuYWJsZWQSHwoL'
    'cXVvdGFfYnl0ZXMYBSABKANSCnF1b3RhQnl0ZXMSEgoEc2l6ZRgGIAEoA1IEc2l6ZRIUCgVmaW'
    'xlcxgHIAEoA1IFZmlsZXMSIAoLcGVybWlzc2lvbnMYCCADKAlSC3Blcm1pc3Npb25zEh0KCmNy'
    'ZWF0ZWRfYXQYCSABKAlSCWNyZWF0ZWRBdA==');

@$core.Deprecated('Use authInfoResponseDescriptor instead')
const AuthInfoResponse$json = {
  '1': 'AuthInfoResponse',
  '2': [
    {'1': 'signup_enabled', '3': 1, '4': 1, '5': 8, '10': 'signupEnabled'},
  ],
};

/// Descriptor for `AuthInfoResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List authInfoResponseDescriptor = $convert.base64Decode(
    'ChBBdXRoSW5mb1Jlc3BvbnNlEiUKDnNpZ251cF9lbmFibGVkGAEgASgIUg1zaWdudXBFbmFibG'
    'Vk');

@$core.Deprecated('Use registerRequestDescriptor instead')
const RegisterRequest$json = {
  '1': 'RegisterRequest',
  '2': [
    {'1': 'username', '3': 1, '4': 1, '5': 9, '10': 'username'},
    {'1': 'password', '3': 2, '4': 1, '5': 9, '10': 'password'},
  ],
};

/// Descriptor for `RegisterRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List registerRequestDescriptor = $convert.base64Decode(
    'Cg9SZWdpc3RlclJlcXVlc3QSGgoIdXNlcm5hbWUYASABKAlSCHVzZXJuYW1lEhoKCHBhc3N3b3'
    'JkGAIgASgJUghwYXNzd29yZA==');

@$core.Deprecated('Use authenticateRequestDescriptor instead')
const AuthenticateRequest$json = {
  '1': 'AuthenticateRequest',
  '2': [
    {'1': 'username', '3': 1, '4': 1, '5': 9, '10': 'username'},
    {'1': 'password', '3': 2, '4': 1, '5': 9, '10': 'password'},
  ],
};

/// Descriptor for `AuthenticateRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List authenticateRequestDescriptor = $convert.base64Decode(
    'ChNBdXRoZW50aWNhdGVSZXF1ZXN0EhoKCHVzZXJuYW1lGAEgASgJUgh1c2VybmFtZRIaCghwYX'
    'Nzd29yZBgCIAEoCVIIcGFzc3dvcmQ=');

@$core.Deprecated('Use meRequestDescriptor instead')
const MeRequest$json = {
  '1': 'MeRequest',
};

/// Descriptor for `MeRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List meRequestDescriptor =
    $convert.base64Decode('CglNZVJlcXVlc3Q=');

@$core.Deprecated('Use getAuthInfoRequestDescriptor instead')
const GetAuthInfoRequest$json = {
  '1': 'GetAuthInfoRequest',
};

/// Descriptor for `GetAuthInfoRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getAuthInfoRequestDescriptor =
    $convert.base64Decode('ChJHZXRBdXRoSW5mb1JlcXVlc3Q=');

@$core.Deprecated('Use listUsersRequestDescriptor instead')
const ListUsersRequest$json = {
  '1': 'ListUsersRequest',
};

/// Descriptor for `ListUsersRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listUsersRequestDescriptor =
    $convert.base64Decode('ChBMaXN0VXNlcnNSZXF1ZXN0');

@$core.Deprecated('Use userStatsDescriptor instead')
const UserStats$json = {
  '1': 'UserStats',
  '2': [
    {'1': 'username', '3': 1, '4': 1, '5': 9, '10': 'username'},
    {'1': 'file_count', '3': 2, '4': 1, '5': 3, '10': 'fileCount'},
    {'1': 'total_size', '3': 3, '4': 1, '5': 3, '10': 'totalSize'},
    {'1': 'is_admin', '3': 4, '4': 1, '5': 8, '10': 'isAdmin'},
    {'1': 'role', '3': 5, '4': 1, '5': 9, '10': 'role'},
    {'1': 'enabled', '3': 6, '4': 1, '5': 8, '10': 'enabled'},
    {'1': 'quota_bytes', '3': 7, '4': 1, '5': 3, '10': 'quotaBytes'},
  ],
};

/// Descriptor for `UserStats`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List userStatsDescriptor = $convert.base64Decode(
    'CglVc2VyU3RhdHMSGgoIdXNlcm5hbWUYASABKAlSCHVzZXJuYW1lEh0KCmZpbGVfY291bnQYAi'
    'ABKANSCWZpbGVDb3VudBIdCgp0b3RhbF9zaXplGAMgASgDUgl0b3RhbFNpemUSGQoIaXNfYWRt'
    'aW4YBCABKAhSB2lzQWRtaW4SEgoEcm9sZRgFIAEoCVIEcm9sZRIYCgdlbmFibGVkGAYgASgIUg'
    'dlbmFibGVkEh8KC3F1b3RhX2J5dGVzGAcgASgDUgpxdW90YUJ5dGVz');

@$core.Deprecated('Use listUsersResponseDescriptor instead')
const ListUsersResponse$json = {
  '1': 'ListUsersResponse',
  '2': [
    {
      '1': 'users',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.fileshare.v1.UserStats',
      '10': 'users'
    },
  ],
};

/// Descriptor for `ListUsersResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listUsersResponseDescriptor = $convert.base64Decode(
    'ChFMaXN0VXNlcnNSZXNwb25zZRItCgV1c2VycxgBIAMoCzIXLmZpbGVzaGFyZS52MS5Vc2VyU3'
    'RhdHNSBXVzZXJz');

@$core.Deprecated('Use loginRequestDescriptor instead')
const LoginRequest$json = {
  '1': 'LoginRequest',
  '2': [
    {'1': 'username', '3': 1, '4': 1, '5': 9, '10': 'username'},
    {'1': 'password', '3': 2, '4': 1, '5': 9, '10': 'password'},
  ],
};

/// Descriptor for `LoginRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List loginRequestDescriptor = $convert.base64Decode(
    'CgxMb2dpblJlcXVlc3QSGgoIdXNlcm5hbWUYASABKAlSCHVzZXJuYW1lEhoKCHBhc3N3b3JkGA'
    'IgASgJUghwYXNzd29yZA==');

@$core.Deprecated('Use loginResponseDescriptor instead')
const LoginResponse$json = {
  '1': 'LoginResponse',
  '2': [
    {'1': 'access_token', '3': 1, '4': 1, '5': 9, '10': 'accessToken'},
    {'1': 'token_type', '3': 2, '4': 1, '5': 9, '10': 'tokenType'},
    {'1': 'expires_in', '3': 3, '4': 1, '5': 3, '10': 'expiresIn'},
  ],
};

/// Descriptor for `LoginResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List loginResponseDescriptor = $convert.base64Decode(
    'Cg1Mb2dpblJlc3BvbnNlEiEKDGFjY2Vzc190b2tlbhgBIAEoCVILYWNjZXNzVG9rZW4SHQoKdG'
    '9rZW5fdHlwZRgCIAEoCVIJdG9rZW5UeXBlEh0KCmV4cGlyZXNfaW4YAyABKANSCWV4cGlyZXNJ'
    'bg==');

@$core.Deprecated('Use logoutRequestDescriptor instead')
const LogoutRequest$json = {
  '1': 'LogoutRequest',
};

/// Descriptor for `LogoutRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List logoutRequestDescriptor =
    $convert.base64Decode('Cg1Mb2dvdXRSZXF1ZXN0');

@$core.Deprecated('Use logoutResponseDescriptor instead')
const LogoutResponse$json = {
  '1': 'LogoutResponse',
  '2': [
    {'1': 'status', '3': 1, '4': 1, '5': 9, '10': 'status'},
  ],
};

/// Descriptor for `LogoutResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List logoutResponseDescriptor = $convert
    .base64Decode('Cg5Mb2dvdXRSZXNwb25zZRIWCgZzdGF0dXMYASABKAlSBnN0YXR1cw==');

@$core.Deprecated('Use createShareRequestDescriptor instead')
const CreateShareRequest$json = {
  '1': 'CreateShareRequest',
  '2': [
    {'1': 'path', '3': 1, '4': 1, '5': 9, '10': 'path'},
    {
      '1': 'expires_in_seconds',
      '3': 2,
      '4': 1,
      '5': 3,
      '10': 'expiresInSeconds'
    },
  ],
};

/// Descriptor for `CreateShareRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createShareRequestDescriptor = $convert.base64Decode(
    'ChJDcmVhdGVTaGFyZVJlcXVlc3QSEgoEcGF0aBgBIAEoCVIEcGF0aBIsChJleHBpcmVzX2luX3'
    'NlY29uZHMYAiABKANSEGV4cGlyZXNJblNlY29uZHM=');

@$core.Deprecated('Use shareDescriptor instead')
const Share$json = {
  '1': 'Share',
  '2': [
    {'1': 'token', '3': 1, '4': 1, '5': 9, '10': 'token'},
    {'1': 'path', '3': 2, '4': 1, '5': 9, '10': 'path'},
    {'1': 'name', '3': 3, '4': 1, '5': 9, '10': 'name'},
    {'1': 'owner', '3': 4, '4': 1, '5': 9, '10': 'owner'},
    {'1': 'created_at', '3': 5, '4': 1, '5': 9, '10': 'createdAt'},
    {'1': 'expires_at', '3': 6, '4': 1, '5': 9, '10': 'expiresAt'},
  ],
};

/// Descriptor for `Share`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List shareDescriptor = $convert.base64Decode(
    'CgVTaGFyZRIUCgV0b2tlbhgBIAEoCVIFdG9rZW4SEgoEcGF0aBgCIAEoCVIEcGF0aBISCgRuYW'
    '1lGAMgASgJUgRuYW1lEhQKBW93bmVyGAQgASgJUgVvd25lchIdCgpjcmVhdGVkX2F0GAUgASgJ'
    'UgljcmVhdGVkQXQSHQoKZXhwaXJlc19hdBgGIAEoCVIJZXhwaXJlc0F0');

@$core.Deprecated('Use listSharesRequestDescriptor instead')
const ListSharesRequest$json = {
  '1': 'ListSharesRequest',
};

/// Descriptor for `ListSharesRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listSharesRequestDescriptor =
    $convert.base64Decode('ChFMaXN0U2hhcmVzUmVxdWVzdA==');

@$core.Deprecated('Use listSharesResponseDescriptor instead')
const ListSharesResponse$json = {
  '1': 'ListSharesResponse',
  '2': [
    {
      '1': 'shares',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.fileshare.v1.Share',
      '10': 'shares'
    },
  ],
};

/// Descriptor for `ListSharesResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listSharesResponseDescriptor = $convert.base64Decode(
    'ChJMaXN0U2hhcmVzUmVzcG9uc2USKwoGc2hhcmVzGAEgAygLMhMuZmlsZXNoYXJlLnYxLlNoYX'
    'JlUgZzaGFyZXM=');

@$core.Deprecated('Use revokeShareRequestDescriptor instead')
const RevokeShareRequest$json = {
  '1': 'RevokeShareRequest',
  '2': [
    {'1': 'token', '3': 1, '4': 1, '5': 9, '10': 'token'},
  ],
};

/// Descriptor for `RevokeShareRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List revokeShareRequestDescriptor = $convert
    .base64Decode('ChJSZXZva2VTaGFyZVJlcXVlc3QSFAoFdG9rZW4YASABKAlSBXRva2Vu');

@$core.Deprecated('Use revokeShareResponseDescriptor instead')
const RevokeShareResponse$json = {
  '1': 'RevokeShareResponse',
  '2': [
    {'1': 'message', '3': 1, '4': 1, '5': 9, '10': 'message'},
  ],
};

/// Descriptor for `RevokeShareResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List revokeShareResponseDescriptor =
    $convert.base64Decode(
        'ChNSZXZva2VTaGFyZVJlc3BvbnNlEhgKB21lc3NhZ2UYASABKAlSB21lc3NhZ2U=');

@$core.Deprecated('Use subscribeRequestDescriptor instead')
const SubscribeRequest$json = {
  '1': 'SubscribeRequest',
};

/// Descriptor for `SubscribeRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List subscribeRequestDescriptor =
    $convert.base64Decode('ChBTdWJzY3JpYmVSZXF1ZXN0');

@$core.Deprecated('Use eventMessageDescriptor instead')
const EventMessage$json = {
  '1': 'EventMessage',
  '2': [
    {'1': 'type', '3': 1, '4': 1, '5': 9, '10': 'type'},
    {'1': 'path', '3': 2, '4': 1, '5': 9, '10': 'path'},
    {'1': 'user', '3': 3, '4': 1, '5': 9, '10': 'user'},
    {'1': 'at', '3': 4, '4': 1, '5': 9, '10': 'at'},
    {'1': 'bytes', '3': 5, '4': 1, '5': 3, '10': 'bytes'},
  ],
};

/// Descriptor for `EventMessage`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List eventMessageDescriptor = $convert.base64Decode(
    'CgxFdmVudE1lc3NhZ2USEgoEdHlwZRgBIAEoCVIEdHlwZRISCgRwYXRoGAIgASgJUgRwYXRoEh'
    'IKBHVzZXIYAyABKAlSBHVzZXISDgoCYXQYBCABKAlSAmF0EhQKBWJ5dGVzGAUgASgDUgVieXRl'
    'cw==');

@$core.Deprecated('Use fileInfoDescriptor instead')
const FileInfo$json = {
  '1': 'FileInfo',
  '2': [
    {'1': 'name', '3': 1, '4': 1, '5': 9, '10': 'name'},
    {'1': 'path', '3': 2, '4': 1, '5': 9, '10': 'path'},
    {'1': 'size', '3': 3, '4': 1, '5': 3, '10': 'size'},
    {'1': 'is_dir', '3': 4, '4': 1, '5': 8, '10': 'isDir'},
    {
      '1': 'modified',
      '3': 5,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'modified'
    },
  ],
};

/// Descriptor for `FileInfo`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List fileInfoDescriptor = $convert.base64Decode(
    'CghGaWxlSW5mbxISCgRuYW1lGAEgASgJUgRuYW1lEhIKBHBhdGgYAiABKAlSBHBhdGgSEgoEc2'
    'l6ZRgDIAEoA1IEc2l6ZRIVCgZpc19kaXIYBCABKAhSBWlzRGlyEjYKCG1vZGlmaWVkGAUgASgL'
    'MhouZ29vZ2xlLnByb3RvYnVmLlRpbWVzdGFtcFIIbW9kaWZpZWQ=');

@$core.Deprecated('Use listFilesRequestDescriptor instead')
const ListFilesRequest$json = {
  '1': 'ListFilesRequest',
  '2': [
    {'1': 'path', '3': 1, '4': 1, '5': 9, '10': 'path'},
  ],
};

/// Descriptor for `ListFilesRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listFilesRequestDescriptor = $convert
    .base64Decode('ChBMaXN0RmlsZXNSZXF1ZXN0EhIKBHBhdGgYASABKAlSBHBhdGg=');

@$core.Deprecated('Use listFilesResponseDescriptor instead')
const ListFilesResponse$json = {
  '1': 'ListFilesResponse',
  '2': [
    {
      '1': 'files',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.fileshare.v1.FileInfo',
      '10': 'files'
    },
  ],
};

/// Descriptor for `ListFilesResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listFilesResponseDescriptor = $convert.base64Decode(
    'ChFMaXN0RmlsZXNSZXNwb25zZRIsCgVmaWxlcxgBIAMoCzIWLmZpbGVzaGFyZS52MS5GaWxlSW'
    '5mb1IFZmlsZXM=');

@$core.Deprecated('Use getFileInfoRequestDescriptor instead')
const GetFileInfoRequest$json = {
  '1': 'GetFileInfoRequest',
  '2': [
    {'1': 'path', '3': 1, '4': 1, '5': 9, '10': 'path'},
  ],
};

/// Descriptor for `GetFileInfoRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getFileInfoRequestDescriptor = $convert
    .base64Decode('ChJHZXRGaWxlSW5mb1JlcXVlc3QSEgoEcGF0aBgBIAEoCVIEcGF0aA==');

@$core.Deprecated('Use searchFilesRequestDescriptor instead')
const SearchFilesRequest$json = {
  '1': 'SearchFilesRequest',
  '2': [
    {'1': 'query', '3': 1, '4': 1, '5': 9, '10': 'query'},
    {'1': 'limit', '3': 2, '4': 1, '5': 5, '10': 'limit'},
  ],
};

/// Descriptor for `SearchFilesRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List searchFilesRequestDescriptor = $convert.base64Decode(
    'ChJTZWFyY2hGaWxlc1JlcXVlc3QSFAoFcXVlcnkYASABKAlSBXF1ZXJ5EhQKBWxpbWl0GAIgAS'
    'gFUgVsaW1pdA==');

@$core.Deprecated('Use searchFilesResponseDescriptor instead')
const SearchFilesResponse$json = {
  '1': 'SearchFilesResponse',
  '2': [
    {
      '1': 'files',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.fileshare.v1.FileInfo',
      '10': 'files'
    },
  ],
};

/// Descriptor for `SearchFilesResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List searchFilesResponseDescriptor = $convert.base64Decode(
    'ChNTZWFyY2hGaWxlc1Jlc3BvbnNlEiwKBWZpbGVzGAEgAygLMhYuZmlsZXNoYXJlLnYxLkZpbG'
    'VJbmZvUgVmaWxlcw==');

@$core.Deprecated('Use createDirectoryRequestDescriptor instead')
const CreateDirectoryRequest$json = {
  '1': 'CreateDirectoryRequest',
  '2': [
    {'1': 'path', '3': 1, '4': 1, '5': 9, '10': 'path'},
  ],
};

/// Descriptor for `CreateDirectoryRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createDirectoryRequestDescriptor =
    $convert.base64Decode(
        'ChZDcmVhdGVEaXJlY3RvcnlSZXF1ZXN0EhIKBHBhdGgYASABKAlSBHBhdGg=');

@$core.Deprecated('Use createDirectoryResponseDescriptor instead')
const CreateDirectoryResponse$json = {
  '1': 'CreateDirectoryResponse',
  '2': [
    {'1': 'message', '3': 1, '4': 1, '5': 9, '10': 'message'},
  ],
};

/// Descriptor for `CreateDirectoryResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createDirectoryResponseDescriptor =
    $convert.base64Decode(
        'ChdDcmVhdGVEaXJlY3RvcnlSZXNwb25zZRIYCgdtZXNzYWdlGAEgASgJUgdtZXNzYWdl');

@$core.Deprecated('Use deletePathRequestDescriptor instead')
const DeletePathRequest$json = {
  '1': 'DeletePathRequest',
  '2': [
    {'1': 'path', '3': 1, '4': 1, '5': 9, '10': 'path'},
  ],
};

/// Descriptor for `DeletePathRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deletePathRequestDescriptor = $convert
    .base64Decode('ChFEZWxldGVQYXRoUmVxdWVzdBISCgRwYXRoGAEgASgJUgRwYXRo');

@$core.Deprecated('Use deletePathResponseDescriptor instead')
const DeletePathResponse$json = {
  '1': 'DeletePathResponse',
  '2': [
    {'1': 'message', '3': 1, '4': 1, '5': 9, '10': 'message'},
  ],
};

/// Descriptor for `DeletePathResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deletePathResponseDescriptor =
    $convert.base64Decode(
        'ChJEZWxldGVQYXRoUmVzcG9uc2USGAoHbWVzc2FnZRgBIAEoCVIHbWVzc2FnZQ==');

@$core.Deprecated('Use updateFileContentRequestDescriptor instead')
const UpdateFileContentRequest$json = {
  '1': 'UpdateFileContentRequest',
  '2': [
    {'1': 'path', '3': 1, '4': 1, '5': 9, '10': 'path'},
    {'1': 'content', '3': 2, '4': 1, '5': 9, '10': 'content'},
  ],
};

/// Descriptor for `UpdateFileContentRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateFileContentRequestDescriptor =
    $convert.base64Decode(
        'ChhVcGRhdGVGaWxlQ29udGVudFJlcXVlc3QSEgoEcGF0aBgBIAEoCVIEcGF0aBIYCgdjb250ZW'
        '50GAIgASgJUgdjb250ZW50');

@$core.Deprecated('Use updateFileContentResponseDescriptor instead')
const UpdateFileContentResponse$json = {
  '1': 'UpdateFileContentResponse',
  '2': [
    {'1': 'message', '3': 1, '4': 1, '5': 9, '10': 'message'},
    {'1': 'size', '3': 2, '4': 1, '5': 3, '10': 'size'},
  ],
};

/// Descriptor for `UpdateFileContentResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateFileContentResponseDescriptor =
    $convert.base64Decode(
        'ChlVcGRhdGVGaWxlQ29udGVudFJlc3BvbnNlEhgKB21lc3NhZ2UYASABKAlSB21lc3NhZ2USEg'
        'oEc2l6ZRgCIAEoA1IEc2l6ZQ==');

@$core.Deprecated('Use uploadMetadataDescriptor instead')
const UploadMetadata$json = {
  '1': 'UploadMetadata',
  '2': [
    {'1': 'path', '3': 1, '4': 1, '5': 9, '10': 'path'},
    {'1': 'filename', '3': 2, '4': 1, '5': 9, '10': 'filename'},
  ],
};

/// Descriptor for `UploadMetadata`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List uploadMetadataDescriptor = $convert.base64Decode(
    'Cg5VcGxvYWRNZXRhZGF0YRISCgRwYXRoGAEgASgJUgRwYXRoEhoKCGZpbGVuYW1lGAIgASgJUg'
    'hmaWxlbmFtZQ==');

@$core.Deprecated('Use uploadRequestDescriptor instead')
const UploadRequest$json = {
  '1': 'UploadRequest',
  '2': [
    {
      '1': 'metadata',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.fileshare.v1.UploadMetadata',
      '9': 0,
      '10': 'metadata'
    },
    {'1': 'chunk', '3': 2, '4': 1, '5': 12, '9': 0, '10': 'chunk'},
  ],
  '8': [
    {'1': 'payload'},
  ],
};

/// Descriptor for `UploadRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List uploadRequestDescriptor = $convert.base64Decode(
    'Cg1VcGxvYWRSZXF1ZXN0EjoKCG1ldGFkYXRhGAEgASgLMhwuZmlsZXNoYXJlLnYxLlVwbG9hZE'
    '1ldGFkYXRhSABSCG1ldGFkYXRhEhYKBWNodW5rGAIgASgMSABSBWNodW5rQgkKB3BheWxvYWQ=');

@$core.Deprecated('Use uploadResponseDescriptor instead')
const UploadResponse$json = {
  '1': 'UploadResponse',
  '2': [
    {'1': 'path', '3': 1, '4': 1, '5': 9, '10': 'path'},
    {'1': 'size', '3': 2, '4': 1, '5': 3, '10': 'size'},
  ],
};

/// Descriptor for `UploadResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List uploadResponseDescriptor = $convert.base64Decode(
    'Cg5VcGxvYWRSZXNwb25zZRISCgRwYXRoGAEgASgJUgRwYXRoEhIKBHNpemUYAiABKANSBHNpem'
    'U=');

@$core.Deprecated('Use downloadFileRequestDescriptor instead')
const DownloadFileRequest$json = {
  '1': 'DownloadFileRequest',
  '2': [
    {'1': 'path', '3': 1, '4': 1, '5': 9, '10': 'path'},
  ],
};

/// Descriptor for `DownloadFileRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List downloadFileRequestDescriptor = $convert
    .base64Decode('ChNEb3dubG9hZEZpbGVSZXF1ZXN0EhIKBHBhdGgYASABKAlSBHBhdGg=');

@$core.Deprecated('Use downloadChunkDescriptor instead')
const DownloadChunk$json = {
  '1': 'DownloadChunk',
  '2': [
    {'1': 'filename', '3': 1, '4': 1, '5': 9, '10': 'filename'},
    {'1': 'content_type', '3': 2, '4': 1, '5': 9, '10': 'contentType'},
    {'1': 'data', '3': 3, '4': 1, '5': 12, '10': 'data'},
  ],
};

/// Descriptor for `DownloadChunk`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List downloadChunkDescriptor = $convert.base64Decode(
    'Cg1Eb3dubG9hZENodW5rEhoKCGZpbGVuYW1lGAEgASgJUghmaWxlbmFtZRIhCgxjb250ZW50X3'
    'R5cGUYAiABKAlSC2NvbnRlbnRUeXBlEhIKBGRhdGEYAyABKAxSBGRhdGE=');
