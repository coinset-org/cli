package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
)

var eventsCmd = &cobra.Command{
	Use:   "events [type]",
	Short: "Connect to WebSocket and display events",
	Long: `Connect to the Coinset WebSocket and display events.
Optionally filter by event type. Valid types are:
  - peak
  - transaction
  - offer
If no type is specified, all events will be displayed.`,
	ValidArgs: []string{"peak", "transaction", "offer"},
	Args:      cobra.MaximumNArgs(1),
	Run:       runEvents,
}

func init() {
	rootCmd.AddCommand(eventsCmd)
}

type Event struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

func runEvents(cmd *cobra.Command, args []string) {
	eventType := ""
	if len(args) > 0 {
		eventType = args[0]
	}

	c, _, err := websocket.DefaultDialer.Dial("wss://api.coinset.org/ws", nil)
	if err != nil {
		fmt.Println(err)
		return
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

			var event Event
			if err := json.Unmarshal(message, &event); err != nil {
				fmt.Println(err)
				continue
			}

			if eventType == "" || event.Type == eventType {
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
