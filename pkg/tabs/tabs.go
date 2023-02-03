package tabs

import (
	"fmt"
)

// The structs exactly as they come over the wire

type MutedInfoReason string

type MutedInfo struct {
	Muted       bool            `json:"muted"`
	Reason      MutedInfoReason `json:"reason"`
	ExtensionId string          `json:"extensionId"`
}

type SharingState struct {
	Screen     string `json:"screen"`
	Camera     bool   `json:"camera"`
	Microphone bool   `json:"microphone"`
}

type Tab struct {
	ID             int           `json:"id"`
	Index          int           `json:"index"`
	WindowId       int           `json:"windowId"`
	OpenerTabId    int           `json:"openerTabId"`
	Highlighted    bool          `json:"highlighted"`
	Active         bool          `json:"active"`
	Pinned         bool          `json:"pinned"`
	LastAccessed   int           `json:"lastAccessed"`
	Audible        bool          `json:"audible"`
	MutedInfo      *MutedInfo    `json:"mutedInfo"`
	Url            string        `json:"url"`
	Title          string        `json:"title"`
	FavIconUrl     string        `json:"favIconUrl"`
	Status         string        `json:"status"`
	Discarded      bool          `json:"discarded"`
	Incognito      bool          `json:"incognito"`
	Width          int           `json:"width"`
	Height         int           `json:"height"`
	Hidden         bool          `json:"hidden"`
	SessionId      string        `json:"sessionId"`
	CookieStoreId  string        `json:"cookieStoreId"`
	IsArticle      bool          `json:"isArticle"`
	IsInReaderMode bool          `json:"isInReaderMode"`
	SharingState   *SharingState `json:"sharingState"`
	Attention      bool          `json:"attention"`
	SuccessorTabId int           `json:"successorTabId"`
}

type TabDelta struct {
	Attention    *bool         `json:"attention"`
	Audible      *bool         `json:"audible"`
	Discarded    *bool         `json:"discarded"`
	FavIconUrl   *string       `json:"favIconUrl"`
	Hidden       *bool         `json:"hidden"`
	IsArticle    *bool         `json:"isArticle"`
	MutedInfo    *MutedInfo    `json:"mutedInfo"`
	Pinned       *bool         `json:"pinned"`
	SharingState *SharingState `json:"sharingState"`
	Status       *string       `json:"status"`
	Title        *string       `json:"title"`
	Url          *string       `json:"url"`
}

// Events

type Event interface {
	Name() string
	Apply(*TabStore) error
}

type CreatedMsg Tab

func (_ *CreatedMsg) Name() string {
	return "created"
}

func (msg *CreatedMsg) Apply(store *TabStore) error {
	tab := Tab(*msg)
	_, err := store.Get(tab.ID)
	if err == nil {
		return fmt.Errorf("ERROR: Create: Tab with id %d already exists", tab.ID)
	}
	store.Open[tab.ID] = &tab
	return nil
}

type ActivatedMsg struct {
	TabId    int `json:"tabId"`
	Previous int `json:"previous"`
	WindowId int `json:"windowId"`
}

func (_ *ActivatedMsg) Name() string {
	return "activated"
}

func (msg *ActivatedMsg) Apply(store *TabStore) error {
	// do we care if the previously active tab isn't found?
	if previous, exists := store.Open[msg.Previous]; exists {
		previous.Active = false
	}
	tab, err := store.Get(msg.TabId)
	if err != nil {
		return fmt.Errorf("ERROR: Activate: %v", err)
	}
	tab.Active = true
	return nil
}

type UpdatedMsg struct {
	TabId int       `json:"tabId"`
	Delta TabDelta  `json:"delta"`
}

func (_ *UpdatedMsg) Name() string {
	return "updated"
}

func (msg *UpdatedMsg) Apply(store *TabStore) error {
	tab, err := store.Get(msg.TabId)
	if err != nil {
		return fmt.Errorf("ERROR: Update: %v", err)
	}
	d := msg.Delta
	if d.Attention != nil {
		tab.Attention = *d.Attention
	}
	if d.Audible != nil {
		tab.Audible = *d.Audible
	}
	if d.Pinned != nil {
		tab.Pinned = *d.Pinned
	}
	if d.FavIconUrl != nil {
		tab.FavIconUrl = *d.FavIconUrl
	}
	if d.Hidden != nil {
		tab.Hidden = *d.Hidden
	}
	if d.IsArticle != nil {
		tab.IsArticle = *d.IsArticle
	}
	if d.MutedInfo != nil {
		tab.MutedInfo = d.MutedInfo
	}
	if d.Pinned != nil {
		tab.Pinned = *d.Pinned
	}
	if d.SharingState != nil {
		tab.SharingState = d.SharingState
	}
	if d.Status != nil {
		tab.Status = *d.Status
	}
	if d.Title != nil {
		tab.Title = *d.Title
	}
	if d.Url != nil {
		tab.Url = *d.Url
	}
	return nil
}

type MovedMsg struct {
	TabId     int `json:"tabId"`
	WindowId  int `json:"windowId"`
	FromIndex int `json:"fromIndex"`
	ToIndex   int `json:"toIndex"`
}

func (_ *MovedMsg) Name() string {
	return "updated"
}
// do reshuffled tabs get moved?
func (msg *MovedMsg) Apply(store *TabStore) error {
	tab, err := store.Get(msg.TabId)
	if err != nil {
		return fmt.Errorf("ERROR: Move: %v", err)
	}
	tab.Index = msg.ToIndex
	return nil
}

type RemovedMsg struct {
	TabId           int  `json:"tabId"`
	WindowId        int  `json:"windowId"`
	IsWindowClosing bool `json:"isWindowClosing"`
}

func (_ *RemovedMsg) Name() string {
	return "removed"
}

func (msg *RemovedMsg) Apply(store *TabStore) error {
	tab, err := store.Get(msg.TabId)
	if err != nil {
		return fmt.Errorf("ERROR: Remove: %v", err)
	}
	store.Closed = append(store.Closed, tab)
	delete(store.Open, tab.ID)
	return nil
}

type AttachedMsg struct {
	TabId    int `json:"tabId"`
	WindowId int `json:"windowId"`
	Position int `json:"position"`
}

func (_ *AttachedMsg) Name() string {
	return "attached"
}

func (msg *AttachedMsg) Apply(store *TabStore) error {
	tab, err := store.Get(msg.TabId)
	if err != nil {
		return fmt.Errorf("ERROR: WindowChange: %v", err)
	}
	tab.WindowId = msg.WindowId
	tab.Index = msg.Position
	return nil
}

type TabStore struct {
	Open   map[int]*Tab
	Closed []*Tab
}

func MakeTabStore() *TabStore {
	return &TabStore{Open: make(map[int]*Tab), Closed: []*Tab{}}
}

func (s *TabStore) Get(id int) (*Tab, error) {
	if tab, exists := s.Open[id]; exists {
		return tab, nil
	}
	return nil, fmt.Errorf("Tab with id %d not found", id)
}
