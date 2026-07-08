package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
)

var (
	eventsFilter      string
	eventsTxIDs       []string
	eventsTxStatus    string
	eventsIncoming    bool
	eventsOutgoing    bool
	eventsOfferIDs    []string
	eventsOfferStatus string
	eventsP2          []string
)

var eventsCmd = &cobra.Command{
	Use:   "events [type]",
	Short: "Connect to WebSocket and display events",
	Long: `Connect to the Coinset WebSocket and display events.
Optionally filter by event type. Valid types are:
  - peak
  - transaction
  - offer
  - balance
  - dashboard
If no type is specified, all events will be displayed.`,
	ValidArgs: []string{"peak", "transaction", "offer", "balance", "dashboard"},
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.MaximumNArgs(1)(cmd, args); err != nil {
			return err
		}

		if len(args) > 0 && eventsFilter != "" {
			return fmt.Errorf("use either positional type or --events, not both")
		}

		if len(args) > 0 && !isValidEventType(args[0]) {
			return fmt.Errorf("invalid event type: %s", args[0])
		}

		if err := validateEventFilter(eventsFilter); err != nil {
			return err
		}
		if err := validateTxStatusFilter(eventsTxStatus); err != nil {
			return err
		}
		if err := validateOfferStatusFilter(eventsOfferStatus); err != nil {
			return err
		}
		for _, p2 := range eventsP2 {
			trimmed := strings.TrimSpace(p2)
			if trimmed == "" {
				continue
			}
			if _, err := convertAddressOrPuzzleHash(trimmed); err != nil {
				return err
			}
		}

		return nil
	},
	Run: runEvents,
}

func init() {
	eventsCmd.Flags().StringVar(&eventsFilter, "events", "", "Comma-separated event types to subscribe to")
	eventsCmd.Flags().StringArrayVar(&eventsTxIDs, "tx-id", []string{}, "Transaction id filter (repeatable)")
	eventsCmd.Flags().StringVar(&eventsTxStatus, "tx-status", "", "Comma-separated tx statuses: pending,confirmed,removed")
	eventsCmd.Flags().BoolVar(&eventsIncoming, "incoming", false, "Filter transaction events to incoming flow (optionally combined with --p2)")
	eventsCmd.Flags().BoolVar(&eventsOutgoing, "outgoing", false, "Filter transaction events to outgoing flow (optionally combined with --p2)")
	eventsCmd.Flags().StringArrayVar(&eventsOfferIDs, "offer-id", []string{}, "Offer id filter (repeatable)")
	eventsCmd.Flags().StringVar(&eventsOfferStatus, "offer-status", "", "Comma-separated offer lifecycle statuses")
	eventsCmd.Flags().StringArrayVar(&eventsP2, "p2", []string{}, "Puzzle hash or address filter for transaction/offer participants (repeatable)")
	rootCmd.AddCommand(eventsCmd)
}

type WsMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type WsEnvelope struct {
	Network    string    `json:"network"`
	Region     string    `json:"region"`
	InstanceID string    `json:"instance_id"`
	Seq        int64     `json:"seq"`
	Message    WsMessage `json:"message"`
}

func isValidEventType(eventType string) bool {
	switch eventType {
	case "peak", "transaction", "offer", "balance", "dashboard":
		return true
	default:
		return false
	}
}

func validateEventFilter(value string) error {
	if value == "" {
		return nil
	}

	for _, item := range strings.Split(value, ",") {
		eventType := strings.TrimSpace(item)
		if eventType == "" {
			continue
		}
		if !isValidEventType(eventType) {
			return fmt.Errorf("invalid event type in --events: %s", eventType)
		}
	}

	return nil
}

func validateTxStatusFilter(value string) error {
	if value == "" {
		return nil
	}

	for _, item := range strings.Split(value, ",") {
		status := strings.TrimSpace(item)
		if status == "" {
			continue
		}
		if status != "pending" && status != "confirmed" && status != "removed" {
			return fmt.Errorf("invalid tx status in --tx-status: %s", status)
		}
	}

	return nil
}

func isValidWSOfferStatus(status string) bool {
	if isValidOfferStatus(status) {
		return true
	}

	// WebSocket streams expose parse failures even though indexed offer queries do not.
	return status == "parse_failed"
}

func validateOfferStatusFilter(value string) error {
	if value == "" {
		return nil
	}

	for _, item := range strings.Split(value, ",") {
		status := strings.TrimSpace(item)
		if status == "" {
			continue
		}
		if !isValidWSOfferStatus(status) {
			return fmt.Errorf("invalid offer status in --offer-status: %s", status)
		}
	}

	return nil
}

func websocketEndpoint(singleEventType string) (string, error) {
	baseUrl, err := url.Parse(apiRoot())
	if err != nil {
		return "", fmt.Errorf("invalid API URL: %v", err)
	}

	scheme := "wss"
	switch baseUrl.Scheme {
	case "http":
		scheme = "ws"
	case "ws", "wss":
		scheme = baseUrl.Scheme
	}

	basePath := strings.TrimRight(baseUrl.Path, "/")
	wsPath := "/ws"
	if basePath != "" {
		wsPath = basePath + "/ws"
	}

	wsURL := &url.URL{
		Scheme: scheme,
		Host:   baseUrl.Host,
		Path:   wsPath,
	}

	queryValues := wsURL.Query()

	if eventsFilter != "" {
		queryValues.Set("events", eventsFilter)
	} else if singleEventType != "" {
		queryValues.Set("events", singleEventType)
	} else {
		queryValues.Set("events", "peak,transaction,offer,balance,dashboard")
	}

	for _, txID := range eventsTxIDs {
		trimmed := strings.TrimSpace(txID)
		if trimmed != "" {
			queryValues.Add("tx_id", trimmed)
		}
	}

	if eventsTxStatus != "" {
		queryValues.Set("tx_status", eventsTxStatus)
	}

	if eventsIncoming && !eventsOutgoing {
		queryValues.Set("tx_direction", "incoming")
	} else if eventsOutgoing && !eventsIncoming {
		queryValues.Set("tx_direction", "outgoing")
	}

	for _, offerID := range eventsOfferIDs {
		trimmed := strings.TrimSpace(offerID)
		if trimmed != "" {
			queryValues.Add("offer_id", trimmed)
		}
	}

	if eventsOfferStatus != "" {
		queryValues.Set("offer_status", eventsOfferStatus)
	}

	for _, p2 := range eventsP2 {
		trimmed := strings.TrimSpace(p2)
		if trimmed == "" {
			continue
		}
		normalized, err := convertAddressOrPuzzleHash(trimmed)
		if err != nil {
			return "", err
		}
		queryValues.Add("p2", normalized)
	}

	wsURL.RawQuery = queryValues.Encode()
	return wsURL.String(), nil
}

func runEvents(cmd *cobra.Command, args []string) {
	eventType := ""
	if len(args) > 0 {
		eventType = args[0]
	}

	endpoint, err := websocketEndpoint(eventType)
	if err != nil {
		log.Fatal(err.Error())
	}

	c, _, err := websocket.DefaultDialer.Dial(endpoint, nil)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer c.Close()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
					return
				}
				fmt.Println(err)
				return
			}

			var envelope WsEnvelope
			if err := json.Unmarshal(message, &envelope); err != nil {
				fmt.Println(err)
				continue
			}

			messageType := envelope.Message.Type
			if messageType == "" {
				// Backward compatibility: older stream format sent top-level type/data.
				var legacyMessage WsMessage
				if err := json.Unmarshal(message, &legacyMessage); err == nil {
					messageType = legacyMessage.Type
				}
			}

			if eventType == "" || messageType == eventType {
				printJson([]byte(message))
			}
		}
	}()

	for {
		select {
		case <-done:
			return
		case <-interrupt:
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				fmt.Println(err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}
