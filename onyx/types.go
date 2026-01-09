package onyx

import (
	"github.com/OnyxDevTools/onyx-database-go/core"
	"github.com/OnyxDevTools/onyx-database-go/onyxclient"
)

type Client interface {
	core.Client
	Typed() onyxclient.Client
}

type (
	Config         = core.Config
	Query          = core.Query
	Condition      = core.Condition
	Sort           = core.Sort
	QueryResults   = core.QueryResults
	PageResult     = core.PageResult
	Iterator       = core.Iterator
	CascadeSpec    = core.CascadeSpec
	CascadeBuilder = core.CascadeBuilder
	CascadeClient  = core.CascadeClient
	Schema         = core.Schema
	Table          = core.Table
	Field          = core.Field
	Resolver       = core.Resolver
	Document       = core.Document
	DocumentClient = core.DocumentClient
	Secret         = core.Secret
	SecretClient   = core.SecretClient
	Error          = core.Error
	ListResult     = core.ListResult
)
