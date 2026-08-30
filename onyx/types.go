package onyx

import "github.com/OnyxDevTools/onyx-database-go/contract"

const (
	WireFormatJSON        = contract.WireFormatJSON
	WireFormatMessagePack = contract.WireFormatMessagePack

	HNSWQueryFormatVersion            = contract.HNSWQueryFormatVersion
	DefaultHNSWCandidates             = contract.DefaultHNSWCandidates
	DefaultHNSWEFSearch               = contract.DefaultHNSWEFSearch
	MaxHNSWCandidates                 = contract.MaxHNSWCandidates
	MaxHNSWEFSearch                   = contract.MaxHNSWEFSearch
	MaxHNSWVectorDimension            = contract.MaxHNSWVectorDimension
	DefaultApproximateIndexCandidates = contract.DefaultApproximateIndexCandidates
	MaxApproximateIndexCandidates     = contract.MaxApproximateIndexCandidates
	MaxApproximateIndexRouteValues    = contract.MaxApproximateIndexRouteValues
	DefaultVectorSearchCandidates     = contract.DefaultVectorSearchCandidates
	MaxVectorSearchCandidates         = contract.MaxVectorSearchCandidates
	IndexTypeDefault                  = contract.IndexTypeDefault
	IndexTypeVector                   = contract.IndexTypeVector
	TableTypeDefault                  = contract.TableTypeDefault
	TableTypeSearchable               = contract.TableTypeSearchable
	SearchSupportLexical              = contract.SearchSupportLexical
	SearchSupportSemantic             = contract.SearchSupportSemantic
	SearchSupportBoth                 = contract.SearchSupportBoth
	SearchModeLexical                 = contract.SearchModeLexical
	SearchModeSemantic                = contract.SearchModeSemantic
	SearchModeHybrid                  = contract.SearchModeHybrid
	SearchMatchAny                    = contract.SearchMatchAny
	SearchMatchAll                    = contract.SearchMatchAll
)

type (
	Client                         = contract.Client
	Config                         = contract.Config
	Query                          = contract.Query
	Condition                      = contract.Condition
	Sort                           = contract.Sort
	SearchMode                     = contract.SearchMode
	SearchMatch                    = contract.SearchMatch
	SearchOptions                  = contract.SearchOptions
	SemanticVectorSignature        = contract.SemanticVectorSignature
	VectorSearchQuery              = contract.VectorSearchQuery
	VectorSearchQueryInput         = contract.VectorSearchQueryInput
	HNSWSearchQuery                = contract.HNSWSearchQuery
	HNSWSearchQueryInput           = contract.HNSWSearchQueryInput
	ApproximateIndexCandidateQuery = contract.ApproximateIndexCandidateQuery
	QueryResults                   = contract.QueryResults
	PageResult                     = contract.PageResult
	Iterator                       = contract.Iterator
	CascadeSpec                    = contract.CascadeSpec
	CascadeBuilder                 = contract.CascadeBuilder
	CascadeClient                  = contract.CascadeClient
	Schema                         = contract.Schema
	Table                          = contract.Table
	TableType                      = contract.TableType
	SearchSupport                  = contract.SearchSupport
	Field                          = contract.Field
	Index                          = contract.Index
	IndexType                      = contract.IndexType
	Resolver                       = contract.Resolver
	OnyxDocument                   = contract.OnyxDocument
	Document                       = contract.Document
	OnyxDocumentsClient            = contract.OnyxDocumentsClient
	DocumentClient                 = contract.DocumentClient
	OnyxSecret                     = contract.OnyxSecret
	Secret                         = contract.Secret
	OnyxSecretsClient              = contract.OnyxSecretsClient
	SecretClient                   = contract.SecretClient
	Error                          = contract.Error
	AIChatMessage                  = contract.AIChatMessage
	AIChatCompletionRequest        = contract.AIChatCompletionRequest
	AIChatCompletionResponse       = contract.AIChatCompletionResponse
	AIChatCompletionChoice         = contract.AIChatCompletionChoice
	AIChatCompletionUsage          = contract.AIChatCompletionUsage
	AIChatCompletionChunk          = contract.AIChatCompletionChunk
	AIChatCompletionChunkChoice    = contract.AIChatCompletionChunkChoice
	AIChatCompletionChunkDelta     = contract.AIChatCompletionChunkDelta
	AIChatStream                   = contract.AIChatStream
	AIModel                        = contract.AIModel
	AIModelsResponse               = contract.AIModelsResponse
	AITool                         = contract.AITool
	AIToolFunction                 = contract.AIToolFunction
	AIToolCall                     = contract.AIToolCall
	AIToolCallFunction             = contract.AIToolCallFunction
	AIScriptApprovalRequest        = contract.AIScriptApprovalRequest
	AIScriptApprovalResponse       = contract.AIScriptApprovalResponse
)
