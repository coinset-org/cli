package cmd

import "fmt"

func applyPaginationOptions(jsonData map[string]interface{}, cursor string, limit int, order string) error {
	if cursor != "" {
		jsonData["cursor"] = cursor
	}

	if limit < 0 {
		return fmt.Errorf("limit must be >= 1 when provided")
	}
	if limit > 0 {
		jsonData["limit"] = limit
	}

	if order != "" {
		if !isAscOrDesc(order) {
			return fmt.Errorf(`invalid order %q: must be "asc" or "desc"`, order)
		}
		jsonData["order"] = order
	}

	return nil
}

func normalizeHexArg(value, fieldName string) (string, error) {
	if !isHex(value) {
		return "", fmt.Errorf("invalid hex value specified for %s: %s", fieldName, value)
	}
	return formatHex(value), nil
}

func normalizeP2Arg(value string) (string, error) {
	return convertAddressOrPuzzleHash(value)
}

func normalizeMaybeHex(value string) string {
	if isHex(value) {
		return formatHex(value)
	}
	return value
}

func isValidOfferStatus(status string) bool {
	switch status {
	case "open", "pending", "confirmed", "cancel_pending", "cancelled", "expired":
		return true
	default:
		return false
	}
}

func isValidOfferFilter(filter string) bool {
	switch filter {
	case "all", "requested", "offered":
		return true
	default:
		return false
	}
}
