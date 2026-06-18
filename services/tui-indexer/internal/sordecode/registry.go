package sordecode

import (
	"strings"

	"github.com/stellar/go-stellar-sdk/xdr"
)

// NamedParam is one decoded contract function argument with spec metadata.
type NamedParam struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Value string `json:"value"`
}

// NamedField is one decoded event or storage field with spec metadata.
type NamedField struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Location string `json:"location,omitempty"`
	Value    string `json:"value"`
}

// DecodedInvocation is a spec-aware view of one contract function call.
type DecodedInvocation struct {
	Function string       `json:"function"`
	Params   []NamedParam `json:"params"`
}

// DecodedEvent is a spec-aware view of one contract event payload.
type DecodedEvent struct {
	Name   string       `json:"name,omitempty"`
	Fields []NamedField `json:"fields"`
}

// DecodedStorage is a spec-aware view of one contract storage entry.
type DecodedStorage struct {
	Key      string       `json:"key"`
	Value    string       `json:"value"`
	KeyUDT   string       `json:"key_udt,omitempty"`
	ValueUDT string       `json:"value_udt,omitempty"`
	Fields   []NamedField `json:"fields,omitempty"`
}

// SpecRegistry indexes contract spec XDR entries for spec-aware decoding.
type SpecRegistry struct {
	functions  map[string]xdr.ScSpecFunctionV0
	events     []xdr.ScSpecEventV0
	structs    map[string]xdr.ScSpecUdtStructV0
	unions     map[string]xdr.ScSpecUdtUnionV0
	enums      map[string]xdr.ScSpecUdtEnumV0
	errorEnums map[string]xdr.ScSpecUdtErrorEnumV0
}

// NewRegistry builds a lookup table from decoded contract spec entries.
func NewRegistry(entries []xdr.ScSpecEntry) *SpecRegistry {
	registry := &SpecRegistry{
		functions:  make(map[string]xdr.ScSpecFunctionV0),
		structs:    make(map[string]xdr.ScSpecUdtStructV0),
		unions:     make(map[string]xdr.ScSpecUdtUnionV0),
		enums:      make(map[string]xdr.ScSpecUdtEnumV0),
		errorEnums: make(map[string]xdr.ScSpecUdtErrorEnumV0),
	}
	for _, entry := range entries {
		switch entry.Kind {
		case xdr.ScSpecEntryKindScSpecEntryFunctionV0:
			fn := entry.MustFunctionV0()
			registry.functions[strings.ToLower(string(fn.Name))] = fn
		case xdr.ScSpecEntryKindScSpecEntryUdtStructV0:
			s := entry.MustUdtStructV0()
			registry.structs[s.Name] = s
		case xdr.ScSpecEntryKindScSpecEntryUdtUnionV0:
			u := entry.MustUdtUnionV0()
			registry.unions[u.Name] = u
		case xdr.ScSpecEntryKindScSpecEntryUdtEnumV0:
			e := entry.MustUdtEnumV0()
			registry.enums[e.Name] = e
		case xdr.ScSpecEntryKindScSpecEntryUdtErrorEnumV0:
			e := entry.MustUdtErrorEnumV0()
			registry.errorEnums[e.Name] = e
		case xdr.ScSpecEntryKindScSpecEntryEventV0:
			registry.events = append(registry.events, entry.MustEventV0())
		}
	}
	return registry
}

func (r *SpecRegistry) Function(name string) (xdr.ScSpecFunctionV0, bool) {
	if r == nil {
		return xdr.ScSpecFunctionV0{}, false
	}
	fn, ok := r.functions[strings.ToLower(strings.TrimSpace(name))]
	return fn, ok
}

func (r *SpecRegistry) EventCount() int {
	if r == nil {
		return 0
	}
	return len(r.events)
}

func (r *SpecRegistry) FindEvent(topics []xdr.ScVal) (*xdr.ScSpecEventV0, bool) {
	if r == nil {
		return nil, false
	}
	for index := range r.events {
		event := r.events[index]
		if len(topics) < len(event.PrefixTopics) {
			continue
		}
		match := true
		for i, prefix := range event.PrefixTopics {
			if ScValToString(topics[i]) != string(prefix) {
				match = false
				break
			}
		}
		if match {
			return &event, true
		}
	}
	return nil, false
}

func (r *SpecRegistry) MatchUnionKey(key xdr.ScVal) (xdr.ScSpecUdtUnionV0, bool) {
	if r == nil {
		return xdr.ScSpecUdtUnionV0{}, false
	}
	for _, union := range r.unions {
		if decodeUnionCase(union, key, r) != "" {
			return union, true
		}
	}
	return xdr.ScSpecUdtUnionV0{}, false
}
