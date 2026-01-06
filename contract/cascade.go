package contract

import (
	"context"
	"strings"
)

// CascadeClient executes cascading save/delete operations using a cascade specification.
type CascadeClient interface {
	Save(ctx context.Context, table string, entity any) error
	Delete(ctx context.Context, table, id string) error
}

// CascadeSpec describes a cascade graph.
type CascadeSpec interface {
	String() string
}

// CascadeBuilder builds CascadeSpec values programmatically.
type CascadeBuilder interface {
	Graph(name string) CascadeBuilder
	GraphType(table string) CascadeBuilder
	SourceField(field string) CascadeBuilder
	TargetField(field string) CascadeBuilder
	Build() CascadeSpec
}

type cascadeSpec string

func (c cascadeSpec) String() string { return string(c) }

// Cascade creates a CascadeSpec from a string representation.
func Cascade(spec string) CascadeSpec { return cascadeSpec(spec) }

type cascadeBuilder struct {
	graphName   string
	graphType   string
	sourceField string
	targetField string
}

// NewCascadeBuilder returns a CascadeBuilder instance.
func NewCascadeBuilder() CascadeBuilder {
	return &cascadeBuilder{}
}

func (c *cascadeBuilder) Graph(name string) CascadeBuilder {
	c.graphName = name
	return c
}

func (c *cascadeBuilder) GraphType(table string) CascadeBuilder {
	c.graphType = table
	return c
}

func (c *cascadeBuilder) SourceField(field string) CascadeBuilder {
	c.sourceField = field
	return c
}

func (c *cascadeBuilder) TargetField(field string) CascadeBuilder {
	c.targetField = field
	return c
}

func (c *cascadeBuilder) Build() CascadeSpec {
	var base string
	if c.graphName != "" {
		base = c.graphName + ":" + c.graphType
	} else {
		base = c.graphType
	}

	var fields []string
	if c.sourceField != "" {
		fields = append(fields, c.sourceField)
	}
	if c.targetField != "" {
		fields = append(fields, c.targetField)
	}

	if len(fields) > 0 {
		base += "(" + strings.Join(fields, ",") + ")"
	}

	return cascadeSpec(base)
}
