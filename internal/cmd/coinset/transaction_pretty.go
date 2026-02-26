package cmd

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"os"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/chia-network/go-chia-libs/pkg/bech32m"
	"github.com/chia-network/go-chia-libs/pkg/types"
)

type transactionSummaryRecord struct {
	ID           string                   `json:"id"`
	Status       string                   `json:"status"`
	FeeMojos     string                   `json:"fee_mojos"`
	Cost         *int64                   `json:"cost"`
	FirstSeenMS  *int64                   `json:"first_seen_ms"`
	ConfirmedAt  *int64                   `json:"confirmed_at_ms"`
	LastUpdated  int64                    `json:"last_updated_ms"`
	Events       []map[string]interface{} `json:"events"`
	Source       string                   `json:"source"`
	SchemaNumber int                      `json:"schema_version"`
}

type transactionByIDEnvelope struct {
	Transaction *transactionSummaryRecord `json:"transaction"`
	Success     bool                      `json:"success"`
}

type transactionListEnvelope struct {
	Transactions []transactionSummaryRecord `json:"transactions"`
	Truncated    *bool                      `json:"truncated"`
	NextCursor   string                     `json:"next_cursor"`
	Success      bool                       `json:"success"`
}

type transactionsByCoinNameEnvelope struct {
	CoinName           string                    `json:"coin_name"`
	CreatedInTxID      *string                   `json:"created_in_tx_id"`
	SpentInTxID        *string                   `json:"spent_in_tx_id"`
	CreatedTransaction *transactionSummaryRecord `json:"created_transaction"`
	SpentTransaction   *transactionSummaryRecord `json:"spent_transaction"`
	Success            bool                      `json:"success"`
}

type ansiStyler struct {
	enabled bool
}

var (
	mojosPerXCH       = big.NewInt(1000000000000)
	mojosToSixDivisor = big.NewInt(1000000)
	sixDigitsMax      = big.NewInt(1000000)
	roundingOffset    = big.NewInt(500000)
)

const defaultIOLimit = 8

var ansiEscapeRegexp = regexp.MustCompile("\x1b\\[[0-9;]*m")

func makeTransactionRequest(path string, jsonData map[string]interface{}, pretty bool) {
	if !pretty {
		makeRequest(path, jsonData)
		return
	}

	if jq != "." || raw || describe {
		log.Fatal("--pretty cannot be combined with --query, --raw, or --describe")
	}

	jsonResponse, err := doRpc(path, jsonData)
	if err != nil {
		log.Fatal(err.Error())
	}

	if err := printPrettyTransactionResponse(jsonResponse); err != nil {
		log.Printf("pretty format unavailable (%v), falling back to JSON output", err)
		printJson(jsonResponse)
	}
}

func printPrettyTransactionResponse(jsonBytes []byte) error {
	var envelope map[string]json.RawMessage
	if err := json.Unmarshal(jsonBytes, &envelope); err != nil {
		return err
	}

	if !responseSucceeded(envelope) {
		printJson(jsonBytes)
		return nil
	}

	styler := newANSIStyler()

	if _, ok := envelope["transaction"]; ok {
		var payload transactionByIDEnvelope
		if err := json.Unmarshal(jsonBytes, &payload); err != nil {
			return err
		}
		if payload.Transaction == nil {
			fmt.Println("No transaction found.")
			return nil
		}
		fmt.Println(formatTransactionRecord(*payload.Transaction, styler))
		return nil
	}

	if _, ok := envelope["transactions"]; ok {
		var payload transactionListEnvelope
		if err := json.Unmarshal(jsonBytes, &payload); err != nil {
			return err
		}
		printTransactionList(payload, styler)
		return nil
	}

	if _, hasCreated := envelope["created_transaction"]; hasCreated {
		var payload transactionsByCoinNameEnvelope
		if err := json.Unmarshal(jsonBytes, &payload); err != nil {
			return err
		}
		printTransactionsByCoinName(payload, styler)
		return nil
	}

	return fmt.Errorf("unsupported transaction response shape")
}

func responseSucceeded(envelope map[string]json.RawMessage) bool {
	rawSuccess, ok := envelope["success"]
	if !ok {
		return true
	}
	var success bool
	if err := json.Unmarshal(rawSuccess, &success); err != nil {
		return true
	}
	return success
}

func printTransactionList(payload transactionListEnvelope, styler ansiStyler) {
	if len(payload.Transactions) == 0 {
		fmt.Println("No transactions.")
		return
	}

	for idx, tx := range payload.Transactions {
		if idx > 0 {
			fmt.Println()
		}
		fmt.Println(formatTransactionRecord(tx, styler))
	}

	if payload.Truncated != nil && *payload.Truncated {
		fmt.Printf("\nMore results available. Next cursor: %s\n", payload.NextCursor)
	}
}

func printTransactionsByCoinName(payload transactionsByCoinNameEnvelope, styler ansiStyler) {
	fmt.Printf("%s  %s\n", styler.bold("Coin"), shortenHexIdentifier(payload.CoinName, 10, 8))

	if payload.CreatedInTxID != nil && strings.TrimSpace(*payload.CreatedInTxID) != "" {
		fmt.Printf("%s  %s\n", styler.dim("created_in_tx"), shortenHexIdentifier(*payload.CreatedInTxID, 10, 8))
	}
	if payload.SpentInTxID != nil && strings.TrimSpace(*payload.SpentInTxID) != "" {
		fmt.Printf("%s  %s\n", styler.dim("spent_in_tx"), shortenHexIdentifier(*payload.SpentInTxID, 10, 8))
	}

	if payload.CreatedTransaction != nil {
		fmt.Printf("\n%s\n", styler.bold("Created Transaction"))
		fmt.Println(formatTransactionRecord(*payload.CreatedTransaction, styler))
	}
	if payload.SpentTransaction != nil {
		fmt.Printf("\n%s\n", styler.bold("Spent Transaction"))
		fmt.Println(formatTransactionRecord(*payload.SpentTransaction, styler))
	}
}

func formatTransactionRecord(tx transactionSummaryRecord, styler ansiStyler) string {
	var out strings.Builder

	txID := shortenHexIdentifier(tx.ID, 10, 8)
	status := normalizeTxStatus(tx.Status)
	cost := "—"
	if tx.Cost != nil {
		cost = formatSignedStringWithCommas(fmt.Sprintf("%d", *tx.Cost))
	}

	fee := fmt.Sprintf("%s XCH", formatMojosAsXCH(tx.FeeMojos))
	age := formatAge(txTimestampMs(tx))

	fmt.Fprintf(
		&out,
		"%s  %s  %s   %s %s   %s %s   %s %s\n",
		styler.bold("TX"),
		styler.bold(txID),
		styler.color(status, statusColorCode(status)),
		styler.dim("fee"),
		fee,
		styler.dim("cost"),
		cost,
		styler.dim("age"),
		age,
	)

	out.WriteString(styler.bold("Activity"))
	out.WriteByte('\n')
	if len(tx.Events) == 0 {
		out.WriteString("  (no events)")
		return out.String()
	}

	for i, event := range tx.Events {
		eventType := getStringFromMap(event, "type")
		headline := formatEventHeadline(event)
		if headline == "" {
			headline = titleCase(eventType)
		}

		chips := formatEventChips(event, styler)
		bar := styler.color("|", eventColorCode(eventType))

		out.WriteString(bar)
		out.WriteByte(' ')
		out.WriteString(headline)
		if chips != "" {
			out.WriteString("  ")
			out.WriteString(chips)
		}

		details := formatEventDetails(event, styler)
		if details != "" {
			out.WriteByte('\n')
			out.WriteString(details)
		}

		if i != len(tx.Events)-1 {
			out.WriteString("\n\n")
		}
	}

	return out.String()
}

func formatEventHeadline(event map[string]interface{}) string {
	eventType := strings.ToLower(getStringFromMap(event, "type"))

	switch eventType {
	case "transfer":
		return formatTransferHeadline(event)
	case "swap":
		return formatSwapHeadline(event)
	case "mint":
		return formatMintHeadline(event)
	case "melt":
		return formatMeltHeadline(event)
	case "pool":
		return formatPoolHeadline(event)
	case "revoke":
		return formatRevokeHeadline("Revoke", event)
	case "clawback":
		return formatRevokeHeadline("Clawback", event)
	case "combine":
		return formatReshapeHeadline("Combine", event)
	case "split":
		return formatReshapeHeadline("Split", event)
	default:
		return fallbackHeadline(titleCase(getStringFromMap(event, "type")), event)
	}
}

func formatTransferHeadline(event map[string]interface{}) string {
	participants := getSliceOfMaps(event, "participants")
	if len(participants) == 0 {
		return fallbackHeadline("Transfer", event)
	}

	var sender map[string]interface{}
	var receiver map[string]interface{}

	for _, participant := range participants {
		sent := hasAnyAmount(getMapFromMap(participant, "sent"))
		received := hasAnyAmount(getMapFromMap(participant, "received"))
		if sender == nil && sent && !received {
			sender = participant
		}
		if receiver == nil && received && !sent {
			receiver = participant
		}
	}
	if sender == nil {
		sender = participants[0]
	}
	if receiver == nil {
		for _, participant := range participants {
			if getStringFromMap(participant, "p2") != getStringFromMap(sender, "p2") {
				receiver = participant
				break
			}
		}
	}

	amount := formatAssetAmountsCompact(getMapFromMap(sender, "sent"))
	if amount == "" && receiver != nil {
		amount = formatAssetAmountsCompact(getMapFromMap(receiver, "received"))
	}

	if amount == "" {
		return fallbackHeadline("Transfer", event)
	}
	if receiver == nil {
		return fmt.Sprintf("Transfer · %s", amount)
	}

	return fmt.Sprintf(
		"Transfer · %s · %s -> %s",
		amount,
		shortP2(getStringFromMap(sender, "p2")),
		shortP2(getStringFromMap(receiver, "p2")),
	)
}

func formatSwapHeadline(event map[string]interface{}) string {
	legs := getSliceOfMaps(event, "legs")
	if len(legs) == 0 {
		return fallbackHeadline("Swap", event)
	}

	sent := formatAssetAmountsCompact(getMapFromMap(legs[0], "sent"))
	received := formatAssetAmountsCompact(getMapFromMap(legs[0], "received"))

	if sent == "" || received == "" {
		return fallbackHeadline("Swap", event)
	}
	return fmt.Sprintf("Swap · %s <-> %s", sent, received)
}

func formatMintHeadline(event map[string]interface{}) string {
	minted := getMapFromMap(event, "minted")
	label := formatMintedAsset(minted)
	if label == "" {
		return fallbackHeadline("Mint", event)
	}
	return fmt.Sprintf("Mint · %s", label)
}

func formatMeltHeadline(event map[string]interface{}) string {
	melted := getMapFromMap(event, "melted")
	label := formatMintedAsset(melted)
	if label == "" {
		return fallbackHeadline("Melt", event)
	}
	return fmt.Sprintf("Melt · %s", label)
}

func formatPoolHeadline(event map[string]interface{}) string {
	action := getMapFromMap(event, "action")
	actionName := getStringFromMap(action, "action")
	if actionName == "" {
		return fallbackHeadline("Pool", event)
	}
	return fmt.Sprintf("Pool · %s", titleCase(actionName))
}

func formatRevokeHeadline(kind string, event map[string]interface{}) string {
	asset := getMapFromMap(event, "asset")
	assetType := strings.ToLower(getStringFromMap(asset, "asset_type"))
	assetID := getStringFromMap(asset, "asset_id")
	amount := getStringFromMap(event, "amount")
	label := formatAmountByAssetType(assetType, assetID, amount)
	if label == "" {
		return fallbackHeadline(kind, event)
	}
	return fmt.Sprintf("%s · %s", kind, label)
}

func formatReshapeHeadline(kind string, event map[string]interface{}) string {
	inCount := intFromAny(event["input_coin_count"])
	outCount := intFromAny(event["output_coin_count"])
	base := fmt.Sprintf("%s · %d -> %d", kind, inCount, outCount)

	asset := getMapFromMap(event, "asset")
	assetType := strings.ToLower(getStringFromMap(asset, "asset_type"))
	assetID := getStringFromMap(asset, "asset_id")
	amount := getStringFromMap(event, "output_amount")
	if amount == "" {
		amount = getStringFromMap(event, "input_amount")
	}
	label := formatAmountByAssetType(assetType, assetID, amount)
	if label == "" {
		return base
	}
	return fmt.Sprintf("%s · %s", base, label)
}

func fallbackHeadline(prefix string, event map[string]interface{}) string {
	inputs := len(getSliceOfMaps(event, "inputs"))
	outputs := len(getSliceOfMaps(event, "outputs"))
	if inputs == 0 && outputs == 0 {
		return prefix
	}
	return fmt.Sprintf("%s · %d -> %d", prefix, inputs, outputs)
}

func formatMintedAsset(asset map[string]interface{}) string {
	assetType := strings.ToLower(getStringFromMap(asset, "asset_type"))
	assetID := getStringFromMap(asset, "asset_id")
	amount := getStringFromMap(asset, "amount")
	switch assetType {
	case "nft":
		return "1 NFT"
	case "cat":
		return formatAmountByAssetType("cat", assetID, amount)
	default:
		return ""
	}
}

func formatAssetAmountsCompact(assetAmounts map[string]interface{}) string {
	xchAmount := getStringFromMap(assetAmounts, "xch")
	cats := getSliceOfMaps(assetAmounts, "cats")
	nfts := getSlice(assetAmounts, "nfts")

	if isNonZeroAmount(xchAmount) {
		primary := fmt.Sprintf("%s XCH", formatMojosAsXCH(xchAmount))
		extras := []string{}
		if len(cats) > 0 {
			extras = append(extras, formatCatExtrasSummary(cats))
		}
		if len(nfts) > 0 {
			if len(nfts) == 1 {
				extras = append(extras, "1 NFT")
			} else {
				extras = append(extras, fmt.Sprintf("%d NFTs", len(nfts)))
			}
		}
		if len(extras) == 0 {
			return primary
		}
		return fmt.Sprintf("%s + %s", primary, strings.Join(extras, " + "))
	}

	if len(cats) > 0 {
		primary := formatCatAmountCompact(cats[0])
		remaining := len(cats) - 1
		if remaining > 0 {
			primary = fmt.Sprintf("%s +%d assets", primary, remaining)
		}
		if len(nfts) > 0 {
			primary = fmt.Sprintf("%s + %d NFT", primary, len(nfts))
			if len(nfts) > 1 {
				primary = fmt.Sprintf("%ss", primary)
			}
		}
		return primary
	}

	if len(nfts) > 0 {
		if len(nfts) == 1 {
			return "1 NFT"
		}
		return fmt.Sprintf("%d NFTs", len(nfts))
	}

	return ""
}

func formatAmountByAssetType(assetType string, assetID string, amount string) string {
	switch strings.ToLower(assetType) {
	case "xch":
		if amount == "" {
			return ""
		}
		return fmt.Sprintf("%s XCH", formatMojosAsXCH(amount))
	case "cat":
		return formatCatAmount(amount, assetID)
	case "nft":
		if amount == "" || amount == "1" {
			return "1 NFT"
		}
		return fmt.Sprintf("%s NFTs", formatSignedStringWithCommas(amount))
	default:
		if amount == "" {
			return ""
		}
		return formatSignedStringWithCommas(amount)
	}
}

func formatCatExtrasSummary(cats []map[string]interface{}) string {
	if len(cats) == 0 {
		return ""
	}

	primary := formatCatAmountCompact(cats[0])
	if len(cats) == 1 {
		return primary
	}

	return fmt.Sprintf("%s +%d assets", primary, len(cats)-1)
}

func formatCatAmountCompact(cat map[string]interface{}) string {
	return formatCatAmount(getStringFromMap(cat, "amount"), getStringFromMap(cat, "asset_id"))
}

func formatCatAmount(amount string, assetID string) string {
	label := catDisplayLabel(assetID)
	if amount == "" {
		return label
	}
	return fmt.Sprintf("%s %s", formatSignedStringWithCommas(amount), label)
}

func catDisplayLabel(assetID string) string {
	_, code, ok := LookupCatNameCodeByID(assetID)
	if ok && strings.TrimSpace(code) != "" {
		return colorCatLabel(code, true)
	}

	normalized := normalizeCatAssetID(assetID)
	if normalized == "" {
		return colorCatLabel("CAT", false)
	}
	if len(normalized) >= 8 {
		return colorCatLabel(fmt.Sprintf("CAT(%s...)", normalized[:8]), false)
	}
	return colorCatLabel(fmt.Sprintf("CAT(%s)", normalized), false)
}

func colorFlowAmount(value string, isSent bool, styler ansiStyler) string {
	if value == "" || value == "—" {
		return value
	}
	if isSent {
		return styler.color(value, "31")
	}
	return styler.color(value, "32")
}

func styleIOHeader(value string, isInput bool, styler ansiStyler) string {
	_ = isInput
	return styler.bold(value)
}

func colorCatLabel(value string, known bool) string {
	styler := newANSIStyler()
	if known {
		return styler.color(value, "36")
	}
	return styler.color(value, "33")
}

func formatEventChips(event map[string]interface{}, styler ansiStyler) string {
	chips := []string{}
	seen := map[string]bool{}
	addChip := func(chip string) {
		chip = strings.TrimSpace(chip)
		if chip == "" || seen[chip] || len(chips) >= 3 {
			return
		}
		seen[chip] = true
		chips = append(chips, chip)
	}

	eventType := strings.ToLower(getStringFromMap(event, "type"))
	amm := getMapFromMap(event, "amm")
	if protocol := getStringFromMap(amm, "protocol"); protocol != "" {
		addChip(titleCase(protocol))
	} else if eventType == "swap" {
		addChip("Offer")
	}

	if len(getSliceOfMaps(event, "royalties")) > 0 {
		addChip("Royalties")
	}
	if eventHasClawback(event) {
		addChip("Clawback")
	}
	if len(getSlice(event, "memos")) > 0 {
		addChip("Memos")
	}

	if len(chips) == 0 {
		return ""
	}

	parts := make([]string, 0, len(chips))
	for _, chip := range chips {
		parts = append(parts, styler.dim("["+chip+"]"))
	}
	return strings.Join(parts, " ")
}

func formatEventDetails(event map[string]interface{}, styler ansiStyler) string {
	eventType := strings.ToLower(getStringFromMap(event, "type"))
	sections := []string{}

	switch eventType {
	case "transfer":
		if section := formatParticipantsSection("Participants", getSliceOfMaps(event, "participants"), "p2", styler); section != "" {
			sections = append(sections, section)
		}
	case "swap":
		if section := formatParticipantsSection("Legs", getSliceOfMaps(event, "legs"), "p2", styler); section != "" {
			sections = append(sections, section)
		}
		if section := formatRoyaltiesSection(getSliceOfMaps(event, "royalties"), styler); section != "" {
			sections = append(sections, section)
		}
		if section := formatAmmSection(getMapFromMap(event, "amm"), styler); section != "" {
			sections = append(sections, section)
		}
	case "mint":
		if section := formatMintSection(event, styler); section != "" {
			sections = append(sections, section)
		}
	case "melt":
		if section := formatMeltSection(event, styler); section != "" {
			sections = append(sections, section)
		}
	case "revoke", "clawback":
		if section := formatRevokeSection(event, styler); section != "" {
			sections = append(sections, section)
		}
	case "combine", "split":
		if section := formatReshapeSection(event, styler); section != "" {
			sections = append(sections, section)
		}
	case "pool":
		if section := formatPoolSection(event, styler); section != "" {
			sections = append(sections, section)
		}
	}

	if section := formatIOTableSection(event, styler); section != "" {
		sections = append(sections, section)
	}
	if section := formatMemosSection(event, styler); section != "" {
		sections = append(sections, section)
	}
	if section := formatConstraintsSection(event, styler); section != "" {
		sections = append(sections, section)
	}

	return strings.Join(sections, "\n")
}

func formatParticipantsSection(title string, entries []map[string]interface{}, addrKey string, styler ansiStyler) string {
	if len(entries) == 0 {
		return ""
	}

	type participantRow struct {
		addr     string
		sent     string
		received string
	}

	rows := make([]participantRow, 0, len(entries))
	maxPrefixWidth := 0

	for _, entry := range entries {
		addr := shortP2(getStringFromMap(entry, addrKey))
		sent := formatAssetAmountsCompact(getMapFromMap(entry, "sent"))
		received := formatAssetAmountsCompact(getMapFromMap(entry, "received"))
		if sent == "" {
			sent = "—"
		}
		if received == "" {
			received = "—"
		}
		sent = colorFlowAmount(sent, true, styler)
		received = colorFlowAmount(received, false, styler)

		prefix := fmt.Sprintf("%s  sent: %s", addr, sent)
		if width := displayWidth(prefix); width > maxPrefixWidth {
			maxPrefixWidth = width
		}

		rows = append(rows, participantRow{
			addr:     addr,
			sent:     sent,
			received: received,
		})
	}

	lines := make([]string, 0, len(rows))
	for _, row := range rows {
		prefix := fmt.Sprintf("%s  sent: %s", row.addr, row.sent)
		pad := maxPrefixWidth - displayWidth(prefix)
		if pad < 0 {
			pad = 0
		}
		lines = append(lines, fmt.Sprintf("%s%s   received: %s", prefix, strings.Repeat(" ", pad), row.received))
	}

	return formatSection(title, lines, styler)
}

func formatRoyaltiesSection(royalties []map[string]interface{}, styler ansiStyler) string {
	if len(royalties) == 0 {
		return ""
	}

	lines := make([]string, 0, len(royalties))
	for _, royalty := range royalties {
		addr := shortP2(getStringFromMap(royalty, "recipient_p2"))
		amount := formatAssetAmountsCompact(getMapFromMap(royalty, "amount"))
		if amount == "" {
			amount = "—"
		}
		lines = append(lines, fmt.Sprintf("%s  %s", addr, amount))
	}

	return formatSection("Royalties", lines, styler)
}

func formatAmmSection(amm map[string]interface{}, styler ansiStyler) string {
	if len(amm) == 0 {
		return ""
	}

	lines := []string{}
	if protocol := getStringFromMap(amm, "protocol"); protocol != "" {
		lines = append(lines, "protocol: "+titleCase(protocol))
	}
	if pair := getStringFromMap(amm, "pair"); pair != "" {
		lines = append(lines, "pair: "+shortenIdentifier(pair, 14, 10))
	}

	return formatSection("AMM", lines, styler)
}

func formatMintSection(event map[string]interface{}, styler ansiStyler) string {
	lines := []string{}

	if recipient := getStringFromMap(event, "recipient_p2"); recipient != "" {
		lines = append(lines, "recipient: "+shortP2(recipient))
	}
	if minter := getStringFromMap(event, "minter_p2"); minter != "" {
		lines = append(lines, "minter: "+shortP2(minter))
	}

	minted := formatMintedAsset(getMapFromMap(event, "minted"))
	if minted != "" {
		lines = append(lines, "asset: "+minted)
	}

	return formatSection("Mint", lines, styler)
}

func formatMeltSection(event map[string]interface{}, styler ansiStyler) string {
	lines := []string{}

	if owner := getStringFromMap(event, "owner_p2"); owner != "" {
		lines = append(lines, "owner: "+shortP2(owner))
	}

	melted := formatMintedAsset(getMapFromMap(event, "melted"))
	if melted != "" {
		lines = append(lines, "asset: "+melted)
	}

	return formatSection("Melt", lines, styler)
}

func formatRevokeSection(event map[string]interface{}, styler ansiStyler) string {
	lines := []string{}

	if by := getStringFromMap(event, "by_p2"); by != "" {
		lines = append(lines, "by: "+shortP2(by))
	}
	if from := getStringFromMap(event, "from_p2"); from != "" {
		lines = append(lines, "from: "+shortP2(from))
	}
	if to := getStringFromMap(event, "to_p2"); to != "" {
		lines = append(lines, "to: "+shortP2(to))
	}

	asset := getMapFromMap(event, "asset")
	assetType := getStringFromMap(asset, "asset_type")
	assetID := getStringFromMap(asset, "asset_id")
	amount := formatAmountByAssetType(assetType, assetID, getStringFromMap(event, "amount"))
	if amount != "" {
		lines = append(lines, "amount: "+amount)
	}

	if mechanism := getStringFromMap(event, "mechanism"); mechanism != "" {
		lines = append(lines, "mechanism: "+titleCase(mechanism))
	}
	if confidence := getStringFromMap(event, "confidence"); confidence != "" {
		lines = append(lines, "confidence: "+titleCase(confidence))
	}

	kind := titleCase(getStringFromMap(event, "type"))
	if kind == "" {
		kind = "Revoke"
	}
	return formatSection(kind, lines, styler)
}

func formatReshapeSection(event map[string]interface{}, styler ansiStyler) string {
	lines := []string{}

	if p2 := getStringFromMap(event, "p2"); p2 != "" {
		lines = append(lines, "p2: "+shortP2(p2))
	}

	inCount := intFromAny(event["input_coin_count"])
	outCount := intFromAny(event["output_coin_count"])
	lines = append(lines, fmt.Sprintf("coins: %d -> %d", inCount, outCount))

	asset := getMapFromMap(event, "asset")
	assetType := getStringFromMap(asset, "asset_type")
	assetID := getStringFromMap(asset, "asset_id")
	amount := getStringFromMap(event, "output_amount")
	if amount == "" {
		amount = getStringFromMap(event, "input_amount")
	}
	if label := formatAmountByAssetType(assetType, assetID, amount); label != "" {
		lines = append(lines, "amount: "+label)
	}

	kind := titleCase(getStringFromMap(event, "type"))
	if kind == "" {
		kind = "Reshape"
	}
	return formatSection(kind, lines, styler)
}

func formatPoolSection(event map[string]interface{}, styler ansiStyler) string {
	lines := []string{}

	if launcher := getStringFromMap(event, "launcher_id"); launcher != "" {
		lines = append(lines, "launcher: "+shortenHexIdentifier(launcher, 10, 8))
	}
	if owner := getStringFromMap(event, "owner_p2"); owner != "" {
		lines = append(lines, "owner: "+shortP2(owner))
	}

	action := getMapFromMap(event, "action")
	if actionName := getStringFromMap(action, "action"); actionName != "" {
		lines = append(lines, "action: "+titleCase(actionName))
	}

	return formatSection("Pool", lines, styler)
}

func formatIOTableSection(event map[string]interface{}, styler ansiStyler) string {
	inputs := getSliceOfMaps(event, "inputs")
	outputs := getSliceOfMaps(event, "outputs")
	if len(inputs) == 0 && len(outputs) == 0 {
		return ""
	}

	missingRow := "- —"

	showInputs := minInt(len(inputs), defaultIOLimit)
	showOutputs := minInt(len(outputs), defaultIOLimit)

	leftHeader := fmt.Sprintf("Inputs (%d/%d)", showInputs, len(inputs))
	rightHeader := fmt.Sprintf("Outputs (%d/%d)", showOutputs, len(outputs))

	leftRows := make([]string, 0, showInputs)
	for i := 0; i < showInputs; i++ {
		leftRows = append(leftRows, formatCoinRow(inputs[i]))
	}

	rightRows := make([]string, 0, showOutputs)
	for i := 0; i < showOutputs; i++ {
		rightRows = append(rightRows, formatCoinRow(outputs[i]))
	}

	leftWidth := displayWidth(leftHeader)
	for _, row := range leftRows {
		if rowWidth := displayWidth(row); rowWidth > leftWidth {
			leftWidth = rowWidth
		}
	}
	if rowWidth := displayWidth(missingRow); rowWidth > leftWidth {
		leftWidth = rowWidth
	}

	var out strings.Builder
	out.WriteString("  ")
	out.WriteString(styleIOHeader(leftHeader, true, styler))
	if pad := leftWidth - displayWidth(leftHeader); pad > 0 {
		out.WriteString(strings.Repeat(" ", pad))
	}
	out.WriteString("    ")
	out.WriteString(styleIOHeader(rightHeader, false, styler))

	rowCount := maxInt(showInputs, showOutputs)
	for i := 0; i < rowCount; i++ {
		left := missingRow
		right := missingRow
		if i < len(leftRows) {
			left = leftRows[i]
		}
		if i < len(rightRows) {
			right = rightRows[i]
		}

		out.WriteByte('\n')
		out.WriteString("  ")
		out.WriteString(left)
		if pad := leftWidth - displayWidth(left); pad > 0 {
			out.WriteString(strings.Repeat(" ", pad))
		}
		out.WriteString("    ")
		out.WriteString(right)
	}

	moreInputs := len(inputs) - showInputs
	moreOutputs := len(outputs) - showOutputs
	if moreInputs > 0 || moreOutputs > 0 {
		hints := []string{}
		if moreInputs > 0 {
			hints = append(hints, fmt.Sprintf("+%d more inputs", moreInputs))
		}
		if moreOutputs > 0 {
			hints = append(hints, fmt.Sprintf("+%d more outputs", moreOutputs))
		}

		out.WriteByte('\n')
		out.WriteString("  ... ")
		out.WriteString(strings.Join(hints, ", "))
	}

	return out.String()
}

func formatCoinRow(coin map[string]interface{}) string {
	coinID := shortenHexIdentifier(getStringFromMap(coin, "coin_id"), 10, 8)
	amount := formatCoinAmountLabel(coin)
	if amount == "" {
		amount = "—"
	}
	return fmt.Sprintf("- %s  %s", coinID, amount)
}

func formatCoinAmountLabel(coin map[string]interface{}) string {
	coinType := getMapFromMap(coin, "coin_type")

	custodyType := strings.ToLower(getStringFromMap(coinType, "custody_puzzle_type"))
	if custodyType == "vault" {
		return "1 Vault"
	}

	outerType := strings.ToLower(getStringFromMap(coinType, "outer_puzzle_type"))
	assetID := getStringFromMap(coinType, "asset_id")

	switch outerType {
	case "cat":
		return formatCatAmount(getStringFromMap(coin, "amount"), assetID)
	case "nft", "singleton":
		amount := getStringFromMap(coin, "amount")
		if amount == "" || amount == "1" {
			return "1 NFT"
		}
		return fmt.Sprintf("%s NFTs", formatSignedStringWithCommas(amount))
	}

	amount := getStringFromMap(coin, "amount")
	if amount == "" {
		return ""
	}
	return fmt.Sprintf("%s XCH", formatMojosAsXCH(amount))
}

func formatMemosSection(event map[string]interface{}, styler ansiStyler) string {
	memos := getSlice(event, "memos")
	if len(memos) == 0 {
		return ""
	}

	lines := make([]string, 0, len(memos))
	for _, memo := range memos {
		lines = append(lines, fmt.Sprintf("%v", memo))
	}

	return formatSection(fmt.Sprintf("Memos (%d)", len(memos)), lines, styler)
}

func formatConstraintsSection(event map[string]interface{}, styler ansiStyler) string {
	constraints := getMapFromMap(event, "constraints")
	if len(constraints) == 0 {
		return ""
	}

	pretty, err := json.MarshalIndent(constraints, "", "  ")
	if err != nil {
		return formatSection("Constraints", []string{fmt.Sprintf("%v", constraints)}, styler)
	}

	var out strings.Builder
	out.WriteString("  ")
	out.WriteString(styler.bold("Constraints"))
	for _, line := range strings.Split(string(pretty), "\n") {
		out.WriteByte('\n')
		out.WriteString("    ")
		out.WriteString(line)
	}
	return out.String()
}

func formatSection(title string, lines []string, styler ansiStyler) string {
	if len(lines) == 0 {
		return ""
	}

	var out strings.Builder
	out.WriteString("  ")
	out.WriteString(styler.bold(title))
	for _, line := range lines {
		out.WriteByte('\n')
		out.WriteString("  - ")
		out.WriteString(line)
	}
	return out.String()
}

func eventHasClawback(event map[string]interface{}) bool {
	if strings.EqualFold(getStringFromMap(event, "type"), "Clawback") {
		return true
	}

	inputs := getSliceOfMaps(event, "inputs")
	for _, input := range inputs {
		if len(getMapFromMap(input, "clawback")) > 0 {
			return true
		}
	}

	outputs := getSliceOfMaps(event, "outputs")
	for _, output := range outputs {
		if len(getMapFromMap(output, "clawback")) > 0 {
			return true
		}
	}

	return false
}

func txTimestampMs(tx transactionSummaryRecord) int64 {
	if tx.ConfirmedAt != nil && *tx.ConfirmedAt > 0 {
		return *tx.ConfirmedAt
	}
	if tx.FirstSeenMS != nil && *tx.FirstSeenMS > 0 {
		return *tx.FirstSeenMS
	}
	if tx.LastUpdated > 0 {
		return tx.LastUpdated
	}
	return 0
}

func formatAge(timestampMs int64) string {
	if timestampMs <= 0 {
		return "—"
	}

	timestamp := time.UnixMilli(timestampMs).Local()
	relative := formatRelativeDuration(time.Since(timestamp))
	absolute := timestamp.Format("2006-01-02 15:04:05")
	return fmt.Sprintf("%s (%s)", relative, absolute)
}

func formatRelativeDuration(duration time.Duration) string {
	future := false
	if duration < 0 {
		future = true
		duration = -duration
	}
	if duration < time.Second {
		if future {
			return "in 0s"
		}
		return "0s"
	}

	seconds := int(duration.Seconds())
	days := seconds / 86400
	hours := (seconds % 86400) / 3600
	minutes := (seconds % 3600) / 60
	secs := seconds % 60

	var value string
	switch {
	case days > 0:
		value = fmt.Sprintf("%dd%dh", days, hours)
	case hours > 0:
		value = fmt.Sprintf("%dh%dm", hours, minutes)
	case minutes > 0:
		value = fmt.Sprintf("%dm%ds", minutes, secs)
	default:
		value = fmt.Sprintf("%ds", secs)
	}

	if future {
		return "in " + value
	}
	return value
}

func normalizeTxStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "", "confirmed":
		return "CONFIRMED"
	case "pending", "mempool":
		return "PENDING"
	default:
		return strings.ToUpper(status)
	}
}

func formatMojosAsXCH(mojos string) string {
	mojoInt, ok := parseAmount(mojos)
	if !ok {
		return "0"
	}

	sign := ""
	if mojoInt.Sign() < 0 {
		sign = "-"
	}

	absMojos := new(big.Int).Abs(mojoInt)
	whole := new(big.Int)
	remainder := new(big.Int)
	whole.QuoRem(absMojos, mojosPerXCH, remainder)

	if remainder.Sign() == 0 {
		return sign + whole.String()
	}

	roundedSix := new(big.Int).Add(remainder, roundingOffset)
	roundedSix.Quo(roundedSix, mojosToSixDivisor)
	roundedWhole := new(big.Int).Set(whole)

	if roundedSix.Cmp(sixDigitsMax) >= 0 {
		roundedWhole.Add(roundedWhole, big.NewInt(1))
		roundedSix.SetInt64(0)
	}

	if roundedSix.Sign() == 0 {
		if whole.Sign() == 0 && remainder.Sign() > 0 {
			return sign + "0." + leftPad(remainder.String(), 12)
		}
		return sign + roundedWhole.String()
	}

	return sign + roundedWhole.String() + "." + leftPad(roundedSix.String(), 6)
}

func parseAmount(value string) (*big.Int, bool) {
	clean := strings.TrimSpace(value)
	if clean == "" {
		return nil, false
	}
	v := new(big.Int)
	if _, ok := v.SetString(clean, 10); !ok {
		return nil, false
	}
	return v, true
}

func isNonZeroAmount(value string) bool {
	parsed, ok := parseAmount(value)
	return ok && parsed.Sign() != 0
}

func formatSignedStringWithCommas(value string) string {
	parsed, ok := parseAmount(value)
	if !ok {
		return value
	}
	sign := ""
	if parsed.Sign() < 0 {
		sign = "-"
	}
	digits := new(big.Int).Abs(parsed).String()
	return sign + addCommas(digits)
}

func addCommas(digits string) string {
	if len(digits) <= 3 {
		return digits
	}

	prefix := len(digits) % 3
	var b strings.Builder
	if prefix > 0 {
		b.WriteString(digits[:prefix])
		if len(digits) > prefix {
			b.WriteByte(',')
		}
	}

	for i := prefix; i < len(digits); i += 3 {
		b.WriteString(digits[i : i+3])
		if i+3 < len(digits) {
			b.WriteByte(',')
		}
	}

	return b.String()
}

func leftPad(value string, width int) string {
	if len(value) >= width {
		return value
	}
	return strings.Repeat("0", width-len(value)) + value
}

func displayWidth(value string) int {
	clean := ansiEscapeRegexp.ReplaceAllString(value, "")
	return utf8.RuneCountInString(clean)
}

func minInt(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

func shortP2(p2 string) string {
	if p2 == "" {
		return "—"
	}
	if strings.HasPrefix(p2, "xch") || strings.HasPrefix(p2, "txch") {
		return shortenIdentifier(p2, 8, 6)
	}

	if !isHex(p2) {
		return shortenIdentifier(p2, 8, 6)
	}

	normalized := formatHex(p2)
	address, err := p2ToAddress(normalized)
	if err == nil {
		return shortenIdentifier(address, 8, 6)
	}
	return shortenHexIdentifier(normalized, 8, 6)
}

func p2ToAddress(p2 string) (string, error) {
	hexBytes, err := hex.DecodeString(strings.TrimPrefix(p2, "0x"))
	if err != nil {
		return "", err
	}
	hexBytes32, err := types.BytesToBytes32(hexBytes)
	if err != nil {
		return "", err
	}

	prefix := "xch"
	if testnet {
		prefix = "txch"
	}

	return bech32m.EncodePuzzleHash(hexBytes32, prefix)
}

func shortenHexIdentifier(value string, left int, right int) string {
	if strings.TrimSpace(value) == "" {
		return "—"
	}

	lower := strings.ToLower(strings.TrimSpace(value))
	if strings.HasPrefix(lower, "0x") {
		lower = lower[2:]
	}

	if !isHexDigits(lower) {
		return shortenIdentifier(value, left, right)
	}

	if len(lower) <= left+right {
		return "0x" + lower
	}
	return "0x" + lower[:left] + "..." + lower[len(lower)-right:]
}

func shortenIdentifier(value string, left int, right int) string {
	clean := strings.TrimSpace(value)
	if clean == "" {
		return "—"
	}
	if len(clean) <= left+right {
		return clean
	}
	return clean[:left] + "..." + clean[len(clean)-right:]
}

func isHexDigits(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		isDigit := r >= '0' && r <= '9'
		isLower := r >= 'a' && r <= 'f'
		if !isDigit && !isLower {
			return false
		}
	}
	return true
}

func hasAnyAmount(amounts map[string]interface{}) bool {
	if isNonZeroAmount(getStringFromMap(amounts, "xch")) {
		return true
	}
	cats := getSliceOfMaps(amounts, "cats")
	for _, cat := range cats {
		if isNonZeroAmount(getStringFromMap(cat, "amount")) {
			return true
		}
	}
	return len(getSlice(amounts, "nfts")) > 0
}

func getStringFromMap(input map[string]interface{}, key string) string {
	if input == nil {
		return ""
	}
	raw, ok := input[key]
	if !ok || raw == nil {
		return ""
	}
	switch v := raw.(type) {
	case string:
		return v
	default:
		return fmt.Sprintf("%v", v)
	}
}

func getMapFromMap(input map[string]interface{}, key string) map[string]interface{} {
	if input == nil {
		return nil
	}
	raw, ok := input[key]
	if !ok || raw == nil {
		return nil
	}
	output, ok := raw.(map[string]interface{})
	if !ok {
		return nil
	}
	return output
}

func getSliceOfMaps(input map[string]interface{}, key string) []map[string]interface{} {
	items := getSlice(input, key)
	if len(items) == 0 {
		return nil
	}
	result := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		entry, ok := item.(map[string]interface{})
		if ok {
			result = append(result, entry)
		}
	}
	return result
}

func getSlice(input map[string]interface{}, key string) []interface{} {
	if input == nil {
		return nil
	}
	raw, ok := input[key]
	if !ok || raw == nil {
		return nil
	}
	items, ok := raw.([]interface{})
	if !ok {
		return nil
	}
	return items
}

func intFromAny(value interface{}) int {
	switch v := value.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	case json.Number:
		n, err := v.Int64()
		if err == nil {
			return int(n)
		}
	case string:
		parsed, ok := parseAmount(v)
		if ok {
			return int(parsed.Int64())
		}
	}
	return 0
}

func newANSIStyler() ansiStyler {
	return ansiStyler{
		enabled: stdoutIsTerminal() && os.Getenv("NO_COLOR") == "",
	}
}

func stdoutIsTerminal() bool {
	info, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) != 0
}

func (s ansiStyler) dim(value string) string {
	return s.wrap("2", value)
}

func (s ansiStyler) bold(value string) string {
	return s.wrap("1", value)
}

func (s ansiStyler) color(value string, code string) string {
	if code == "" {
		return value
	}
	return s.wrap(code, value)
}

func (s ansiStyler) wrap(code string, value string) string {
	if !s.enabled {
		return value
	}
	return "\033[" + code + "m" + value + "\033[0m"
}

func statusColorCode(status string) string {
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case "CONFIRMED":
		return "32"
	case "PENDING":
		return "33"
	case "REMOVED":
		return "31"
	default:
		return "90"
	}
}

func eventColorCode(eventType string) string {
	switch strings.ToLower(strings.TrimSpace(eventType)) {
	case "transfer":
		return "36"
	case "swap":
		return "32"
	case "mint":
		return "35"
	case "melt":
		return "33"
	case "pool":
		return "33"
	case "revoke", "clawback":
		return "31"
	case "combine", "split":
		return "32"
	default:
		return "90"
	}
}

func titleCase(value string) string {
	clean := strings.TrimSpace(value)
	if clean == "" {
		return ""
	}
	clean = strings.ReplaceAll(clean, "_", " ")
	clean = strings.ReplaceAll(clean, "-", " ")

	parts := strings.Fields(strings.ToLower(clean))
	for i, part := range parts {
		if len(part) == 0 {
			continue
		}
		parts[i] = strings.ToUpper(part[:1]) + part[1:]
	}
	return strings.Join(parts, " ")
}
