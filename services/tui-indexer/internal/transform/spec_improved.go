package transform

import (
	"strings"
)

// OperationDetailsSpecImproved reports whether re-enrichment produced a better spec decode payload.
func OperationDetailsSpecImproved(before, after string) bool {
	if strings.TrimSpace(before) == strings.TrimSpace(after) {
		return false
	}
	beforeDetails, beforeErr := parseOperationDetailsMap(before)
	afterDetails, afterErr := parseOperationDetailsMap(after)
	if beforeErr != nil || afterErr != nil {
		return strings.TrimSpace(before) != strings.TrimSpace(after)
	}
	beforeStatus := stringFromDetails(beforeDetails, "spec_decode_status")
	afterStatus := stringFromDetails(afterDetails, "spec_decode_status")
	if afterStatus == specDecodeStatusDecoded && beforeStatus != specDecodeStatusDecoded {
		return true
	}
	if afterStatus == specDecodeStatusPartial &&
		(beforeStatus == specDecodeStatusNoSpec || beforeStatus == "") {
		return true
	}
	_, hadDecodedArgs := beforeDetails["arguments_decoded"]
	_, hasDecodedArgs := afterDetails["arguments_decoded"]
	if hasDecodedArgs && !hadDecodedArgs {
		return true
	}
	_, hadDecodedResult := beforeDetails["result_decoded"]
	_, hasDecodedResult := afterDetails["result_decoded"]
	return hasDecodedResult && !hadDecodedResult
}

// EventValueSpecImproved reports whether re-enrichment produced a better event decode payload.
func EventValueSpecImproved(before, after *string) bool {
	beforeRaw := stringOrEmptyPtr(before)
	afterRaw := stringOrEmptyPtr(after)
	if beforeRaw == afterRaw {
		return false
	}
	if !strings.HasPrefix(afterRaw, "{") {
		return false
	}
	if !strings.HasPrefix(beforeRaw, "{") {
		return true
	}
	beforeStatus := eventPayloadSpecStatus(beforeRaw)
	afterStatus := eventPayloadSpecStatus(afterRaw)
	if afterStatus == specDecodeStatusDecoded && beforeStatus != specDecodeStatusDecoded {
		return true
	}
	return afterStatus == specDecodeStatusPartial &&
		(beforeStatus == specDecodeStatusNoSpec || beforeStatus == "")
}

func eventPayloadSpecStatus(raw string) string {
	details, err := parseOperationDetailsMap(raw)
	if err != nil {
		return ""
	}
	return stringFromDetails(details, "spec_decode_status")
}

func stringOrEmptyPtr(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}
