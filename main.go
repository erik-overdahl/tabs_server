package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	tabs "github.com/erik-overdahl/tabs_server/pkg/tabs"
)

func main() {
	cmd := "client"
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}
	switch cmd {
	case "gateway":
		gateway := tabs.MakeGateway()
		gateway.Start()
	case "client":
		client := tabs.MakeTabsClient()
		store := tabs.MakeTabStore()
		go func(s *tabs.TabStore) {
			log.Println("Listening for updates")
			for update := range client.Updates {
				log.Printf("Received: %#v", update)
				update.Apply(s)
			}
		}(store)

		client.ConnectBrowserGateway()

		tabList, err := client.GetList()
		if err != nil {
			log.Fatalf("Failed to get list of tabs: %v", err)
		}
		for _, tab := range tabList {
			store.Open[tab.ID] = tab
		}

		scanner := bufio.NewScanner(os.Stdin)
		fmt.Print("> ")
		for scanner.Scan() {
			input := strings.Split(cleanInput(scanner.Text()), " ")
			cmd := input[0]
			switch cmd {
			case "list":
				for _, tab := range store.Open {
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
				if tabId, err := client.Create(tabs.CreateProperties{Url: &url}); err != nil {
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
				if err := client.Move(tabId, tabs.MoveProperties{WindowId: windowId, Index: index}); err != nil {
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

func getIds(input []string) []int {
	ids := make([]int, len(input), len(input))
	for i, s := range input {
		id, _ := strconv.Atoi(s)
		ids[i] = id
	}
	return ids
}
