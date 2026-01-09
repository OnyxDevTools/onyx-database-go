package onyx

import "github.com/OnyxDevTools/onyx-database-go/onyxclient"

type TypedClient = onyxclient.Client

var Tables = onyxclient.Tables
var Resolvers = onyxclient.Resolvers

type AuditLog = onyxclient.AuditLog
type Permission = onyxclient.Permission
type Role = onyxclient.Role
type RolePermission = onyxclient.RolePermission
type User = onyxclient.User
type UserProfile = onyxclient.UserProfile
type UserRole = onyxclient.UserRole
