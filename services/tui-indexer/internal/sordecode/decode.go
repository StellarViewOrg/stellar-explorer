package sordecode

import (
	"fmt"
	"strings"

	"github.com/stellar/go-stellar-sdk/xdr"
)

// DecodeScVal renders one ScVal using an optional contract spec type definition.
func (r *SpecRegistry) DecodeScVal(value xdr.ScVal, typeDef *xdr.ScSpecTypeDef) (string, error) {
	if typeDef == nil {
		return ScValToString(value), nil
	}
	return r.decodeTypedScVal(value, *typeDef)
}

func (r *SpecRegistry) decodeTypedScVal(value xdr.ScVal, typeDef xdr.ScSpecTypeDef) (string, error) {
	switch typeDef.Type {
	case xdr.ScSpecTypeScSpecTypeVal:
		return ScValToString(value), nil
	case xdr.ScSpecTypeScSpecTypeOption:
		if value.Type == xdr.ScValTypeScvVoid {
			return "None", nil
		}
		inner, err := r.decodeTypedScVal(value, typeDef.MustOption().ValueType)
		if err != nil {
			return "", err
		}
		return "Some(" + inner + ")", nil
	case xdr.ScSpecTypeScSpecTypeResult:
		resultType := typeDef.MustResult()
		if value.Type == xdr.ScValTypeScvError {
			if value.Error != nil {
				return "Err(" + formatScError(*value.Error) + ")", nil
			}
			return "Err(error)", nil
		}
		inner, err := r.decodeTypedScVal(value, resultType.OkType)
		if err != nil {
			return "", err
		}
		return "Ok(" + inner + ")", nil
	case xdr.ScSpecTypeScSpecTypeVec:
		if value.Type != xdr.ScValTypeScvVec || value.Vec == nil || *value.Vec == nil {
			return ScValToString(value), nil
		}
		vec := **value.Vec
		elementType := typeDef.MustVec().ElementType
		parts := make([]string, 0, len(vec))
		for _, item := range vec {
			decoded, err := r.decodeTypedScVal(item, elementType)
			if err != nil {
				return "", err
			}
			parts = append(parts, decoded)
		}
		return "[" + strings.Join(parts, ",") + "]", nil
	case xdr.ScSpecTypeScSpecTypeMap:
		if value.Type != xdr.ScValTypeScvMap || value.Map == nil || *value.Map == nil {
			return ScValToString(value), nil
		}
		mapType := typeDef.MustMap()
		entries := **value.Map
		parts := make([]string, 0, len(entries))
		for _, entry := range entries {
			key, err := r.decodeTypedScVal(entry.Key, mapType.KeyType)
			if err != nil {
				return "", err
			}
			val, err := r.decodeTypedScVal(entry.Val, mapType.ValueType)
			if err != nil {
				return "", err
			}
			parts = append(parts, key+":"+val)
		}
		return "{" + strings.Join(parts, ",") + "}", nil
	case xdr.ScSpecTypeScSpecTypeTuple:
		if value.Type != xdr.ScValTypeScvVec || value.Vec == nil || *value.Vec == nil {
			return ScValToString(value), nil
		}
		vec := **value.Vec
		types := typeDef.MustTuple().ValueTypes
		parts := make([]string, 0, len(vec))
		for index, item := range vec {
			var itemType xdr.ScSpecTypeDef
			if index < len(types) {
				itemType = types[index]
			} else {
				itemType = xdr.ScSpecTypeDef{Type: xdr.ScSpecTypeScSpecTypeVal}
			}
			decoded, err := r.decodeTypedScVal(item, itemType)
			if err != nil {
				return "", err
			}
			parts = append(parts, decoded)
		}
		return "(" + strings.Join(parts, ",") + ")", nil
	case xdr.ScSpecTypeScSpecTypeUdt:
		return r.decodeUDT(value, typeDef.MustUdt().Name)
	default:
		if !primitiveTypeMatches(value, typeDef.Type) {
			return ScValToString(value), nil
		}
		return ScValToString(value), nil
	}
}

func primitiveTypeMatches(value xdr.ScVal, specType xdr.ScSpecType) bool {
	switch specType {
	case xdr.ScSpecTypeScSpecTypeBool:
		return value.Type == xdr.ScValTypeScvBool
	case xdr.ScSpecTypeScSpecTypeU32:
		return value.Type == xdr.ScValTypeScvU32
	case xdr.ScSpecTypeScSpecTypeI32:
		return value.Type == xdr.ScValTypeScvI32
	case xdr.ScSpecTypeScSpecTypeU64:
		return value.Type == xdr.ScValTypeScvU64
	case xdr.ScSpecTypeScSpecTypeI64:
		return value.Type == xdr.ScValTypeScvI64
	case xdr.ScSpecTypeScSpecTypeU128:
		return value.Type == xdr.ScValTypeScvU128
	case xdr.ScSpecTypeScSpecTypeI128:
		return value.Type == xdr.ScValTypeScvI128
	case xdr.ScSpecTypeScSpecTypeU256:
		return value.Type == xdr.ScValTypeScvU256
	case xdr.ScSpecTypeScSpecTypeI256:
		return value.Type == xdr.ScValTypeScvI256
	case xdr.ScSpecTypeScSpecTypeString:
		return value.Type == xdr.ScValTypeScvString
	case xdr.ScSpecTypeScSpecTypeSymbol:
		return value.Type == xdr.ScValTypeScvSymbol
	case xdr.ScSpecTypeScSpecTypeAddress, xdr.ScSpecTypeScSpecTypeMuxedAddress:
		return value.Type == xdr.ScValTypeScvAddress
	case xdr.ScSpecTypeScSpecTypeBytes, xdr.ScSpecTypeScSpecTypeBytesN:
		return value.Type == xdr.ScValTypeScvBytes
	case xdr.ScSpecTypeScSpecTypeVoid:
		return value.Type == xdr.ScValTypeScvVoid
	default:
		return true
	}
}

func (r *SpecRegistry) decodeUDT(value xdr.ScVal, name string) (string, error) {
	if s, ok := r.structs[name]; ok {
		return decodeStruct(s, value, r)
	}
	if u, ok := r.unions[name]; ok {
		return decodeUnion(u, value, r), nil
	}
	if e, ok := r.enums[name]; ok {
		return decodeEnum(e, value), nil
	}
	return ScValToString(value), nil
}

func decodeStruct(structType xdr.ScSpecUdtStructV0, value xdr.ScVal, registry *SpecRegistry) (string, error) {
	switch value.Type {
	case xdr.ScValTypeScvMap:
		if value.Map == nil || *value.Map == nil {
			return ScValToString(value), nil
		}
		entries := **value.Map
		lookup := make(map[string]xdr.ScVal, len(entries))
		for _, entry := range entries {
			lookup[ScValToString(entry.Key)] = entry.Val
		}
		parts := make([]string, 0, len(structType.Fields))
		for _, field := range structType.Fields {
			fieldValue, ok := lookup[string(field.Name)]
			if !ok {
				parts = append(parts, fmt.Sprintf("%s=<missing>", field.Name))
				continue
			}
			decoded, err := registry.decodeTypedScVal(fieldValue, field.Type)
			if err != nil {
				return "", err
			}
			parts = append(parts, fmt.Sprintf("%s=%s", field.Name, decoded))
		}
		return structType.Name + "{" + strings.Join(parts, ",") + "}", nil
	case xdr.ScValTypeScvVec:
		if value.Vec == nil || *value.Vec == nil {
			return ScValToString(value), nil
		}
		vec := **value.Vec
		parts := make([]string, 0, len(structType.Fields))
		for index, field := range structType.Fields {
			if index >= len(vec) {
				parts = append(parts, fmt.Sprintf("%s=<missing>", field.Name))
				continue
			}
			decoded, err := registry.decodeTypedScVal(vec[index], field.Type)
			if err != nil {
				return "", err
			}
			parts = append(parts, fmt.Sprintf("%s=%s", field.Name, decoded))
		}
		return structType.Name + "(" + strings.Join(parts, ",") + ")", nil
	default:
		return ScValToString(value), nil
	}
}

func decodeUnion(unionType xdr.ScSpecUdtUnionV0, value xdr.ScVal, registry *SpecRegistry) string {
	caseName := decodeUnionCase(unionType, value, registry)
	if caseName == "" {
		return ScValToString(value)
	}
	return unionType.Name + "::" + caseName
}

func decodeUnionCase(unionType xdr.ScSpecUdtUnionV0, value xdr.ScVal, registry *SpecRegistry) string {
	switch value.Type {
	case xdr.ScValTypeScvSymbol:
		if value.Sym == nil {
			return ""
		}
		caseName := string(*value.Sym)
		for _, unionCase := range unionType.Cases {
			if unionCase.Kind == xdr.ScSpecUdtUnionCaseV0KindScSpecUdtUnionCaseVoidV0 {
				if string(unionCase.MustVoidCase().Name) == caseName {
					return caseName
				}
			}
		}
		return caseName
	case xdr.ScValTypeScvVec:
		if value.Vec == nil || *value.Vec == nil || len(**value.Vec) == 0 {
			return ""
		}
		vec := **value.Vec
		if vec[0].Type != xdr.ScValTypeScvSymbol || vec[0].Sym == nil {
			return ""
		}
		caseName := string(*vec[0].Sym)
		for _, unionCase := range unionType.Cases {
			if unionCase.Kind != xdr.ScSpecUdtUnionCaseV0KindScSpecUdtUnionCaseTupleV0 {
				continue
			}
			tupleCase := unionCase.MustTupleCase()
			if string(tupleCase.Name) != caseName {
				continue
			}
			parts := make([]string, 0, len(tupleCase.Type))
			for index, fieldType := range tupleCase.Type {
				valueIndex := index + 1
				if valueIndex >= len(vec) {
					parts = append(parts, "<missing>")
					continue
				}
				decoded, err := registry.decodeTypedScVal(vec[valueIndex], fieldType)
				if err != nil {
					parts = append(parts, ScValToString(vec[valueIndex]))
					continue
				}
				parts = append(parts, decoded)
			}
			if len(parts) == 0 {
				return caseName
			}
			return caseName + "(" + strings.Join(parts, ",") + ")"
		}
		return caseName
	default:
		return ""
	}
}

func decodeEnum(enumType xdr.ScSpecUdtEnumV0, value xdr.ScVal) string {
	switch value.Type {
	case xdr.ScValTypeScvSymbol:
		if value.Sym != nil {
			return enumType.Name + "::" + string(*value.Sym)
		}
	case xdr.ScValTypeScvU32:
		if value.U32 != nil {
			for _, enumCase := range enumType.Cases {
				if enumCase.Value == *value.U32 {
					return enumType.Name + "::" + string(enumCase.Name)
				}
			}
			return fmt.Sprintf("%s(%d)", enumType.Name, *value.U32)
		}
	}
	return ScValToString(value)
}

// DecodeInvocation decodes invoke_host_function arguments using the contract spec.
func (r *SpecRegistry) DecodeInvocation(functionName string, args []xdr.ScVal) (DecodedInvocation, error) {
	result := DecodedInvocation{Function: strings.TrimSpace(functionName)}
	fn, ok := r.Function(functionName)
	if !ok {
		for index, arg := range args {
			result.Params = append(result.Params, NamedParam{
				Name:  fmt.Sprintf("arg%d", index+1),
				Type:  "Val",
				Value: ScValToString(arg),
			})
		}
		return result, nil
	}

	for index, input := range fn.Inputs {
		param := NamedParam{
			Name: string(input.Name),
			Type: FormatTypeDef(input.Type),
		}
		if index < len(args) {
			decoded, err := r.decodeTypedScVal(args[index], input.Type)
			if err != nil {
				return result, err
			}
			param.Value = decoded
		} else {
			param.Value = "<missing>"
		}
		result.Params = append(result.Params, param)
	}
	return result, nil
}

// DecodeInvocationResult decodes the return value for one function using the contract spec.
func (r *SpecRegistry) DecodeInvocationResult(functionName string, value xdr.ScVal) (string, error) {
	fn, ok := r.Function(functionName)
	if !ok || len(fn.Outputs) == 0 {
		return ScValToString(value), nil
	}
	return r.decodeTypedScVal(value, fn.Outputs[0])
}

// DecodeEvent decodes contract event topics and data using the contract spec.
func (r *SpecRegistry) DecodeEvent(topics []xdr.ScVal, data xdr.ScVal) (DecodedEvent, error) {
	result := DecodedEvent{}
	event, ok := r.FindEvent(topics)
	if !ok {
		for index, topic := range topics {
			result.Fields = append(result.Fields, NamedField{
				Name:     fmt.Sprintf("topic_%d", index+1),
				Type:     "Symbol",
				Location: "topic",
				Value:    ScValToString(topic),
			})
		}
		result.Fields = append(result.Fields, NamedField{
			Name:     "data",
			Type:     "Val",
			Location: "data",
			Value:    ScValToString(data),
		})
		return result, nil
	}

	result.Name = string(event.Name)
	topicIndex := len(event.PrefixTopics)
	var dataValues []xdr.ScVal
	switch event.DataFormat {
	case xdr.ScSpecEventDataFormatScSpecEventDataFormatVec:
		if data.Type == xdr.ScValTypeScvVec && data.Vec != nil && *data.Vec != nil {
			dataValues = **data.Vec
		}
	case xdr.ScSpecEventDataFormatScSpecEventDataFormatMap:
		if data.Type == xdr.ScValTypeScvMap && data.Map != nil && *data.Map != nil {
			for _, entry := range **data.Map {
				dataValues = append(dataValues, entry.Val)
			}
		}
	default:
		dataValues = []xdr.ScVal{data}
	}
	dataIndex := 0

	for _, param := range event.Params {
		field := NamedField{
			Name: string(param.Name),
			Type: FormatTypeDef(param.Type),
		}
		switch param.Location {
		case xdr.ScSpecEventParamLocationV0ScSpecEventParamLocationTopicList:
			field.Location = "topic"
			if topicIndex < len(topics) {
				decoded, err := r.decodeTypedScVal(topics[topicIndex], param.Type)
				if err != nil {
					return result, err
				}
				field.Value = decoded
				topicIndex++
			} else {
				field.Value = "<missing>"
			}
		default:
			field.Location = "data"
			if dataIndex < len(dataValues) {
				decoded, err := r.decodeTypedScVal(dataValues[dataIndex], param.Type)
				if err != nil {
					return result, err
				}
				field.Value = decoded
				dataIndex++
			} else if event.DataFormat == xdr.ScSpecEventDataFormatScSpecEventDataFormatMap && data.Type == xdr.ScValTypeScvMap && data.Map != nil && *data.Map != nil {
				field.Value = lookupMapValue(**data.Map, string(param.Name), param.Type, r)
			} else {
				field.Value = "<missing>"
			}
		}
		result.Fields = append(result.Fields, field)
	}
	return result, nil
}

func lookupMapValue(entries []xdr.ScMapEntry, key string, typeDef xdr.ScSpecTypeDef, registry *SpecRegistry) string {
	for _, entry := range entries {
		if ScValToString(entry.Key) != key {
			continue
		}
		decoded, err := registry.decodeTypedScVal(entry.Val, typeDef)
		if err != nil {
			return ScValToString(entry.Val)
		}
		return decoded
	}
	return "<missing>"
}

// DecodeStorage decodes one contract storage key/value using spec UDT hints when possible.
func (r *SpecRegistry) DecodeStorage(key xdr.ScVal, value xdr.ScVal) DecodedStorage {
	result := DecodedStorage{
		Key:   ScValToString(key),
		Value: ScValToString(value),
	}
	if union, ok := r.MatchUnionKey(key); ok {
		result.KeyUDT = union.Name
		result.Key = decodeUnion(union, key, r)
	}
	for _, structType := range r.structs {
		if key.Type == xdr.ScValTypeScvMap {
			if decoded, err := decodeStruct(structType, key, r); err == nil && strings.HasPrefix(decoded, structType.Name) {
				result.KeyUDT = structType.Name
				result.Key = decoded
				break
			}
		}
	}
	return result
}

// FormatNamedParams renders decoded invocation params as a compact string.
func FormatNamedParams(params []NamedParam) string {
	parts := make([]string, 0, len(params))
	for _, param := range params {
		name := param.Name
		if name == "" {
			name = "arg"
		}
		parts = append(parts, fmt.Sprintf("%s=%s", name, param.Value))
	}
	return strings.Join(parts, ", ")
}
