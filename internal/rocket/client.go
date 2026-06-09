package rocket

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/RocketChat/Rocket.Chat.Go.SDK/models"
	"github.com/RocketChat/Rocket.Chat.Go.SDK/realtime"
	"github.com/RocketChat/Rocket.Chat.Go.SDK/rest"
)

type roomInfoResponse struct {
	Room struct {
		ID   string `json:"_id"`
		Type string `json:"t"`
	} `json:"room"`
	Success bool `json:"success"`
}

func (r *roomInfoResponse) OK() error {
	if !r.Success {
		return fmt.Errorf("request failed")
	}
	return nil
}

type Client struct {
	serverURL  *url.URL
	realtime   *realtime.Client
	rest       *rest.Client
	creds      *models.UserCredentials
	msgHandler MessageHandler
	msgChan    chan models.Message
	stopChan   chan struct{}
	mu         sync.Mutex
	connected  bool
	user       string
	password   string
	token      string
	userID     string
}

type MessageHandler func(msg IncomingMessage)

type IncomingMessage struct {
	Text      string
	UserID    string
	Username  string
	RoomID    string
	Timestamp time.Time
}

func New(serverURL string) (*Client, error) {
	u, err := url.Parse(serverURL)
	if err != nil {
		return nil, fmt.Errorf("invalid server URL: %w", err)
	}

	return &Client{
		serverURL: u,
		msgChan:   make(chan models.Message, 100),
		stopChan:  make(chan struct{}),
	}, nil
}

func (c *Client) Connect(user, password, token, userID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.user = user
	c.password = password
	c.token = token
	c.userID = userID

	log.Printf("Connecting to Rocket.Chat at %s", c.serverURL.String())

	if err := c.connect(); err != nil {
		if isAuthError(err) {
			return fmt.Errorf("authentication failed — check bot credentials: %w", err)
		}
		return err
	}

	go c.watchConnection()

	return nil
}

func (c *Client) connect() error {
	rt, err := realtime.NewClient(c.serverURL, false)
	if err != nil {
		return fmt.Errorf("realtime connect: %w", err)
	}

	var creds *models.UserCredentials
	if c.token != "" && c.userID != "" {
		creds = &models.UserCredentials{
			ID:    c.userID,
			Token: c.token,
		}
	} else {
		creds = &models.UserCredentials{
			Email:    c.user,
			Password: c.password,
		}
	}

	if _, err := rt.Login(creds); err != nil {
		rt.Close()
		return fmt.Errorf("realtime login: %w", err)
	}

	restClient := rest.NewClient(c.serverURL, false)
	if err := restClient.Login(creds); err != nil {
		rt.Close()
		return fmt.Errorf("rest login: %w", err)
	}

	if c.realtime != nil {
		c.realtime.Close()
	}

	c.realtime = rt
	c.rest = restClient
	c.creds = creds
	c.connected = true

	log.Printf("Connected as %s (user ID: %s)", c.user, creds.ID)
	return nil
}

func isAuthError(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	return strings.Contains(s, "401") ||
		strings.Contains(s, "User not found") ||
		strings.Contains(s, "error-login-blocked") ||
		strings.Contains(s, "Login has been temporarily blocked")
}

func (c *Client) Disconnect() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return
	}

	close(c.stopChan)
	c.realtime.Close()
	c.connected = false
	log.Println("Disconnected from Rocket.Chat")
}

func (c *Client) SubscribeToMessages(roomID string) error {
	channel := &models.Channel{ID: roomID}
	return c.realtime.SubscribeToMessageStream(channel, c.msgChan)
}

func (c *Client) SubscribeToMyMessages() error {
	return c.realtime.SubscribeToMyMessages(c.msgChan)
}

func (c *Client) SendMessage(roomID, text string) error {
	channel := &models.Channel{ID: roomID}
	msg := c.realtime.NewMessage(channel, text)
	_, err := c.realtime.SendMessage(msg)
	return err
}

func (c *Client) OnMessage(handler MessageHandler) {
	c.msgHandler = handler
}

func (c *Client) IsConnected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.connected
}

func (c *Client) CreateDM(username string) (string, error) {
	if c.rest == nil {
		return "", fmt.Errorf("not connected")
	}
	room, err := c.rest.CreateDirectMessage(username)
	if err != nil {
		return "", fmt.Errorf("create DM: %w", err)
	}
	rid := room.Rid
	if rid == "" {
		rid = room.ID
	}
	return rid, nil
}

func (c *Client) IsDMRoom(roomID string) (bool, error) {
	if c.rest == nil {
		return false, fmt.Errorf("not connected")
	}
	var resp roomInfoResponse
	if err := c.rest.Get("rooms.info", url.Values{"roomId": {roomID}}, &resp); err != nil {
		if e, ok := err.(*json.UnmarshalTypeError); !ok && e != nil {
			return false, err
		}
	}
	return resp.Room.Type == "d", nil
}

func (c *Client) RegisterSlashCommand(command string) error {
	if c.rest == nil {
		return fmt.Errorf("not connected")
	}
	body := fmt.Sprintf(`{"command": "%s", "clientOnly": true}`, command)
	var resp rest.Status
	if err := c.rest.Post("commands.register", bytes.NewBufferString(body), &resp); err != nil {
		return fmt.Errorf("register slash command: %w", err)
	}
	if !resp.Success {
		return fmt.Errorf("slash command registration failed: %s", resp.Error)
	}
	log.Printf("Registered /%s as a client-only slash command", command)
	return nil
}

func (c *Client) UserInfo(username string) (*models.User, error) {
	if c.rest == nil {
		return nil, fmt.Errorf("not connected")
	}
	return c.rest.UserInfo(username)
}

func (c *Client) listenForMessages() {
	for {
		select {
		case msg := <-c.msgChan:
			if c.msgHandler == nil {
				continue
			}
			incoming := IncomingMessage{
				Text:     msg.Msg,
				UserID:   msg.User.ID,
				Username: msg.User.UserName,
				RoomID:   msg.RoomID,
			}
			if msg.Timestamp != nil {
				incoming.Timestamp = *msg.Timestamp
			}
			c.msgHandler(incoming)

		case <-c.stopChan:
			return
		}
	}
}

func (c *Client) watchConnection() {
	const (
		initialDelay = 2 * time.Second
		maxDelay     = 30 * time.Second
	)

	c.realtime.AddStatusListener(func(status int) {
		if status == 0 {
			c.mu.Lock()
			c.connected = false
			c.mu.Unlock()
			log.Println("Connection lost. Attempting to reconnect...")

			delay := initialDelay
			for {
				select {
				case <-c.stopChan:
					return
				default:
				}

				time.Sleep(delay)
				log.Printf("Reconnecting in %v...", delay)

				c.mu.Lock()
				err := c.connect()
				c.mu.Unlock()

				if isAuthError(err) {
					log.Fatalf("Authentication failed — check bot credentials: %v", err)
				}

				if err == nil {
					log.Println("Reconnected successfully")

					c.mu.Lock()
					if c.creds != nil {
						_ = c.rest.Login(c.creds)
					}
					c.mu.Unlock()

					return
				}

				log.Printf("Reconnect failed: %v", err)
				delay *= 2
				if delay > maxDelay {
					delay = maxDelay
				}
			}
		}
	})

	go c.listenForMessages()
}
