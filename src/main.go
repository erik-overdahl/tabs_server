package main

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

func main() {
	cmd := "serve"
	if len(os.Args) > 0 {
		cmd = os.Args[1]
	}
	switch cmd {
	case "gateway":
		gateway := MakeGateway()
		gateway.Start()
	case "client":
		client := MakeTabsClient()
		client.ConnectBrowserGateway()
		response, err := client.Request(Message{ID: uuid.New(), Action: "list"})
		if err != nil {
			log.Fatalf("Failed to get list of tabs: %v", err)
		}

		tabs := MakeTabStore()
		tabs.Apply(response)
		go func(store *TabStore) {
			log.Println("Listening for updates")
			for update := range client.Updates {
				tabs.Apply(update)
			}
		}(tabs)

		scanner := bufio.NewScanner(os.Stdin)
		fmt.Print("> ")
		for scanner.Scan() {
			input := strings.Split(cleanInput(scanner.Text()), " ")
			cmd := input[0]
			switch cmd {
			case "list":
				for _, tab := range tabs.Open {
					marker := " "
					if tab.Active {
						marker = "*"
					}
					fmt.Printf("%s %d.%d\t%s\t%s\n",
						marker, tab.WindowId, tab.ID, tab.Title, tab.Url)
				}
			case "switch_to":
				tabId, _ := strconv.Atoi(input[1])
				if err := client.Activate(tabId); err != nil {
					fmt.Println("ERROR:", err)
				} else {
					fmt.Println("SUCCESS")
				}
			case "create":
				url := input[1]
				if tabId, err := client.Create(CreateProperties{Url: &url}); err != nil {
					fmt.Println("ERROR:", err)
				} else {
					fmt.Printf("Created tab (id %d)\n", tabId)
				}
			case "duplicate":
				tabId, _ := strconv.Atoi(input[1])
				if tabId, err := client.Duplicate(tabId, nil); err != nil {
					fmt.Println("ERROR:", err)
				} else {
					fmt.Printf("Created tab (id %d)\n", tabId)
				}
			case "close":
				ids := getIds(input[1:])
				if err := client.Close(ids[0], ids[1:]...); err != nil {
					fmt.Println("ERROR:", err)
				} else {
					fmt.Println("SUCCESS")
				}
			case "reload":
				tabId, _ := strconv.Atoi(input[1])
				if err := client.Reload(tabId, nil); err != nil {
					fmt.Println("ERROR:", err)
				} else {
					fmt.Println("SUCCESS")
				}
			case "move":
				tabId, _ := strconv.Atoi(input[1])
				moveTo := strings.Split(input[2], ".")
				windowId, _ := strconv.Atoi(moveTo[0])
				index, _ := strconv.Atoi(moveTo[1])
				if err := client.Move(tabId, MoveProperties{WindowId: windowId, Index: index}); err != nil {
					fmt.Println("ERROR:", err)
				} else {
					fmt.Println("SUCCESS")
				}
			case "discard":
				ids := getIds(input[1:])
				if err := client.Discard(ids[0], ids[1:]...); err != nil {
					fmt.Println("ERROR:", err)
				} else {
					fmt.Println("SUCCESS")
				}
			case "hide":
				ids := getIds(input[1:])
				if err := client.Hide(ids[0], ids[1:]...); err != nil {
					fmt.Println("ERROR:", err)
				} else {
					fmt.Println("SUCCESS")
				}
			case "show":
				ids := getIds(input[1:])
				if err := client.Show(ids[0], ids[1:]...); err != nil {
					fmt.Println("ERROR:", err)
				} else {
					fmt.Println("SUCCESS")
				}
			case "toggle_reader_mode":
				tabId, _ := strconv.Atoi(input[1])
				if err := client.ToggleReaderMode(tabId); err != nil {
					fmt.Println("ERROR:", err)
				} else {
					fmt.Println("SUCCESS")
				}
			case "go_back":
				tabId, _ := strconv.Atoi(input[1])
				if err := client.GoBack(tabId); err != nil {
					fmt.Println("ERROR:", err)
				} else {
					fmt.Println("SUCCESS")
				}
			case "go_forward":
				tabId, _ := strconv.Atoi(input[1])
				if err := client.GoForward(tabId); err != nil {
					fmt.Println("ERROR:", err)
				} else {
					fmt.Println("SUCCESS")
				}

			case "exit":
				fmt.Println("Goodbye")
				os.Exit(0)
			default:
				fmt.Println("ERROR: Unknown command:", cmd)
			}
			fmt.Print("\n> ")
		}
	default:
		log.Printf("ERROR: passed unknown command '%s'", cmd)
		os.Exit(1)
	}
}

func cleanInput(text string) string {
	output := strings.TrimSpace(text)
	output = strings.ToLower(output)
	return output
}

type Message struct {
	ID      uuid.UUID       `json:"id"`
	Action  string          `json:"action"`
	Content json.RawMessage `json:"content"`
}

type EmptyObj map[string]any

func MakeMessage(action string, content any) *Message {
	contentBuf, _ := json.Marshal(content)
	return &Message{
		ID:      uuid.New(),
		Action:  action,
		Content: contentBuf,
	}
}

func ReadMsg(r io.Reader) (*Message, error) {
	buf := make([]byte, 4)
	_, err := io.ReadFull(r, buf)
	if err == io.EOF {
		return nil, err
	} else if err != nil {
		return nil, fmt.Errorf("ERROR: reading msg size: %v", err)
	}
	msgSize := int(binary.LittleEndian.Uint32(buf))
	buf = append(buf, make([]byte, msgSize)...)
	_, err = io.ReadFull(r, buf[4:])
	if err == io.EOF {
		return nil, err
	} else if err != nil {
		return nil, fmt.Errorf("ERROR: reading msg: %v", err)
	}
	msg := &Message{}
	if err := json.Unmarshal(buf[4:], msg); err != nil {
		return nil, err
	}
	return msg, nil
}

func SendMsg(w io.Writer, msg Message) error {
	// add size if not already there
	buf, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	size := make([]byte, 4)
	binary.LittleEndian.PutUint32(size, uint32(len(buf)))
	buf = append(size, buf...)
	_, err = w.Write(buf)
	return err
}

func CastAndCall[T any](content json.RawMessage, f func(*T) error) error {
	var msg T
	if err := json.Unmarshal(content, &msg); err != nil {
		return fmt.Errorf("ERROR: Failed to parse message content: %v", err)
	}
	log.Printf("RECEIVED: %#v", msg)
	return f(&msg)
}

func handleBrowserError(err *string) error {
	log.Printf("Browser returned error: %s", *err)
	return nil
}

func getIds(input []string) []int {
	ids := make([]int, len(input), len(input))
	for i, s := range input {
		id, _ := strconv.Atoi(s)
		ids[i] = id
	}
	return ids
}
