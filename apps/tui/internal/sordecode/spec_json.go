package sordecode

import (
	"github.com/stellar/go-stellar-sdk/xdr"
)

// EntriesToJSON converts contract spec XDR entries into a JSON-serializable structure.
func EntriesToJSON(entries []xdr.ScSpecEntry) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(entries))
	for _, entry := range entries {
		result = append(result, entryToJSON(entry))
	}
	return result
}

func entryToJSON(entry xdr.ScSpecEntry) map[string]interface{} {
	m := map[string]interface{}{
		"kind": entry.Kind.String(),
	}
	switch entry.Kind {
	case xdr.ScSpecEntryKindScSpecEntryFunctionV0:
		fn := entry.MustFunctionV0()
		inputs := make([]map[string]interface{}, 0, len(fn.Inputs))
		for _, input := range fn.Inputs {
			inputs = append(inputs, map[string]interface{}{
				"name": string(input.Name),
				"doc":  string(input.Doc),
				"type": typeDefToJSON(input.Type),
			})
		}
		outputs := make([]map[string]interface{}, 0, len(fn.Outputs))
		for _, output := range fn.Outputs {
			outputs = append(outputs, typeDefToJSON(output))
		}
		m["name"] = string(fn.Name)
		m["doc"] = string(fn.Doc)
		m["inputs"] = inputs
		m["outputs"] = outputs
	case xdr.ScSpecEntryKindScSpecEntryUdtStructV0:
		s := entry.MustUdtStructV0()
		fields := make([]map[string]interface{}, 0, len(s.Fields))
		for _, field := range s.Fields {
			fields = append(fields, map[string]interface{}{
				"name": string(field.Name),
				"doc":  string(field.Doc),
				"type": typeDefToJSON(field.Type),
			})
		}
		m["name"] = s.Name
		m["doc"] = string(s.Doc)
		m["lib"] = string(s.Lib)
		m["fields"] = fields
	case xdr.ScSpecEntryKindScSpecEntryUdtUnionV0:
		u := entry.MustUdtUnionV0()
		cases := make([]map[string]interface{}, 0, len(u.Cases))
		for _, unionCase := range u.Cases {
			switch unionCase.Kind {
			case xdr.ScSpecUdtUnionCaseV0KindScSpecUdtUnionCaseVoidV0:
				voidCase := unionCase.MustVoidCase()
				cases = append(cases, map[string]interface{}{
					"kind": "void",
					"name": string(voidCase.Name),
					"doc":  string(voidCase.Doc),
				})
			case xdr.ScSpecUdtUnionCaseV0KindScSpecUdtUnionCaseTupleV0:
				tupleCase := unionCase.MustTupleCase()
				values := make([]map[string]interface{}, 0, len(tupleCase.Type))
				for _, valueType := range tupleCase.Type {
					values = append(values, typeDefToJSON(valueType))
				}
				cases = append(cases, map[string]interface{}{
					"kind":   "tuple",
					"name":   string(tupleCase.Name),
					"doc":    string(tupleCase.Doc),
					"values": values,
				})
			}
		}
		m["name"] = u.Name
		m["doc"] = string(u.Doc)
		m["lib"] = string(u.Lib)
		m["cases"] = cases
	case xdr.ScSpecEntryKindScSpecEntryUdtEnumV0:
		e := entry.MustUdtEnumV0()
		cases := make([]map[string]interface{}, 0, len(e.Cases))
		for _, enumCase := range e.Cases {
			cases = append(cases, map[string]interface{}{
				"name":  string(enumCase.Name),
				"doc":   string(enumCase.Doc),
				"value": enumCase.Value,
			})
		}
		m["name"] = e.Name
		m["doc"] = string(e.Doc)
		m["lib"] = string(e.Lib)
		m["cases"] = cases
	case xdr.ScSpecEntryKindScSpecEntryUdtErrorEnumV0:
		e := entry.MustUdtErrorEnumV0()
		cases := make([]map[string]interface{}, 0, len(e.Cases))
		for _, enumCase := range e.Cases {
			cases = append(cases, map[string]interface{}{
				"name":  string(enumCase.Name),
				"doc":   string(enumCase.Doc),
				"value": enumCase.Value,
			})
		}
		m["name"] = e.Name
		m["doc"] = string(e.Doc)
		m["lib"] = string(e.Lib)
		m["cases"] = cases
	case xdr.ScSpecEntryKindScSpecEntryEventV0:
		ev := entry.MustEventV0()
		prefixTopics := make([]string, 0, len(ev.PrefixTopics))
		for _, topic := range ev.PrefixTopics {
			prefixTopics = append(prefixTopics, string(topic))
		}
		params := make([]map[string]interface{}, 0, len(ev.Params))
		for _, param := range ev.Params {
			location := "data"
			if param.Location == xdr.ScSpecEventParamLocationV0ScSpecEventParamLocationTopicList {
				location = "topic"
			}
			params = append(params, map[string]interface{}{
				"name":     string(param.Name),
				"doc":      string(param.Doc),
				"type":     typeDefToJSON(param.Type),
				"location": location,
			})
		}
		m["name"] = string(ev.Name)
		m["doc"] = string(ev.Doc)
		m["lib"] = string(ev.Lib)
		m["prefix_topics"] = prefixTopics
		m["params"] = params
		m["data_format"] = ev.DataFormat.String()
	}
	return m
}

func typeDefToJSON(typeDef xdr.ScSpecTypeDef) map[string]interface{} {
	result := map[string]interface{}{
		"type": FormatTypeDef(typeDef),
	}
	switch typeDef.Type {
	case xdr.ScSpecTypeScSpecTypeOption:
		result["value_type"] = typeDefToJSON(typeDef.MustOption().ValueType)
	case xdr.ScSpecTypeScSpecTypeResult:
		resultType := typeDef.MustResult()
		result["ok_type"] = typeDefToJSON(resultType.OkType)
		result["error_type"] = typeDefToJSON(resultType.ErrorType)
	case xdr.ScSpecTypeScSpecTypeVec:
		result["element_type"] = typeDefToJSON(typeDef.MustVec().ElementType)
	case xdr.ScSpecTypeScSpecTypeMap:
		mapType := typeDef.MustMap()
		result["key_type"] = typeDefToJSON(mapType.KeyType)
		result["value_type"] = typeDefToJSON(mapType.ValueType)
	case xdr.ScSpecTypeScSpecTypeTuple:
		values := make([]map[string]interface{}, 0, len(typeDef.MustTuple().ValueTypes))
		for _, valueType := range typeDef.MustTuple().ValueTypes {
			values = append(values, typeDefToJSON(valueType))
		}
		result["value_types"] = values
	case xdr.ScSpecTypeScSpecTypeBytesN:
		result["n"] = typeDef.MustBytesN().N
	case xdr.ScSpecTypeScSpecTypeUdt:
		result["name"] = typeDef.MustUdt().Name
	}
	return result
}
