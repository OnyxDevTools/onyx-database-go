package onyx

import "github.com/OnyxDevTools/onyx-database-go/contract"

type (
	Client                      = contract.Client
	Config                      = contract.Config
	Query                       = contract.Query
	Condition                   = contract.Condition
	Sort                        = contract.Sort
	QueryResults                = contract.QueryResults
	PageResult                  = contract.PageResult
	Iterator                    = contract.Iterator
	CascadeSpec                 = contract.CascadeSpec
	CascadeBuilder              = contract.CascadeBuilder
	CascadeClient               = contract.CascadeClient
	Schema                      = contract.Schema
	Table                       = contract.Table
	Field                       = contract.Field
	Resolver                    = contract.Resolver
	OnyxDocument                = contract.OnyxDocument
	Document                    = contract.Document
	OnyxDocumentsClient         = contract.OnyxDocumentsClient
	DocumentClient              = contract.DocumentClient
	OnyxSecret                  = contract.OnyxSecret
	Secret                      = contract.Secret
	OnyxSecretsClient           = contract.OnyxSecretsClient
	SecretClient                = contract.SecretClient
	Error                       = contract.Error
	AIChatMessage               = contract.AIChatMessage
	AIChatCompletionRequest     = contract.AIChatCompletionRequest
	AIChatCompletionResponse    = contract.AIChatCompletionResponse
	AIChatCompletionChoice      = contract.AIChatCompletionChoice
	AIChatCompletionUsage       = contract.AIChatCompletionUsage
	AIChatCompletionChunk       = contract.AIChatCompletionChunk
	AIChatCompletionChunkChoice = contract.AIChatCompletionChunkChoice
	AIChatCompletionChunkDelta  = contract.AIChatCompletionChunkDelta
	AIChatStream                = contract.AIChatStream
	AIModel                     = contract.AIModel
	AIModelsResponse            = contract.AIModelsResponse
	AITool                      = contract.AITool
	AIToolFunction              = contract.AIToolFunction
	AIToolCall                  = contract.AIToolCall
	AIToolCallFunction          = contract.AIToolCallFunction
	AIScriptApprovalRequest     = contract.AIScriptApprovalRequest
	AIScriptApprovalResponse    = contract.AIScriptApprovalResponse
)
