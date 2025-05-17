package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/CpBruceMeena/Go-Chatsync/static"
	"github.com/gorilla/websocket"
)

// Message types
const (
	// Frontend to Backend
	TypePrivateMessage    = "private_message"
	TypeGroupMessage      = "group_message"
	TypeCreateGroup       = "create_group"
	TypeAddGroupMember    = "add_group_member"
	TypeRemoveGroupMember = "remove_group_member"
	TypeLeaveGroup        = "leave_group"

	// Backend Storage
	TypePrivate = "private"
	TypeGroup   = "group"

	// Backend to Frontend
	TypeUserList  = "user_list"
	TypeGroupList = "group_list"
	TypeSystem    = "system"
)

// Message keys
const (
	KeyType      = "type"
	KeyFrom      = "from"
	KeyTo        = "to"
	KeyContent   = "content"
	KeyTimestamp = "timestamp"
	KeyUsers     = "users"
	KeyGroups    = "groups"
)

// Group keys
const (
	KeyName    = "name"
	KeyAdmin   = "admin"
	KeyMembers = "members"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins for development
		},
	}

	// WebSocket configuration
	maxMessageSize int64 = 512 * 1024 // 512KB
	pongWait             = 60 * time.Second
	writeWait            = 10 * time.Second
)

// Client represents a connected WebSocket client
type Client struct {
	Username string
	conn     *websocket.Conn
	send     chan []byte
}

// Message represents a chat message
type Message struct {
	Type      string `json:"type"`
	From      string `json:"from"`
	To        string `json:"to"`
	Content   string `json:"content"`
	Timestamp string `json:"timestamp"`
}

// Group represents a chat group
type Group struct {
	Name    string   `json:"name"`
	Admin   string   `json:"admin"`   // Group admin
	Members []string `json:"members"` // List of member usernames
}

var (
	// Global variables
	clients    = make(map[string]*Client)
	clientsMux sync.RWMutex
	groups     = make(map[string]*Group)
	groupsMux  sync.RWMutex
	messages   = make(map[string][]Message)
	msgMux     sync.RWMutex
)

// getConversationKey returns a consistent key for a conversation between two users
func getConversationKey(user1, user2 string) string {
	if user1 < user2 {
		return user1 + ":" + user2
	}
	return user2 + ":" + user1
}

// storeMessage stores a message in the message history
func storeMessage(msg Message) {
	msgMux.Lock()
	defer msgMux.Unlock()

	if msg.Type == TypePrivateMessage {
		key := getConversationKey(msg.From, msg.To)
		messages[key] = append(messages[key], msg)
		log.Printf("Stored private message: from=%s, to=%s, key=%s, total_messages=%d, message=%+v",
			msg.From, msg.To, key, len(messages[key]), msg)
	}

	if msg.Type == TypeGroupMessage {
		key := "group_" + msg.To
		messages[key] = append(messages[key], msg)
		log.Printf("Stored group message: group=%s, key=%s, total_messages=%d, message=%+v",
			msg.To, key, len(messages[key]), msg)
	}
}

// getConversationHistory returns the message history for a conversation
func getConversationHistory(user1, user2 string) []Message {
	key := getConversationKey(user1, user2)
	msgMux.RLock()
	store, exists := messages[key]
	msgMux.RUnlock()

	if !exists {
		log.Printf("No message history found for conversation %s between %s and %s", key, user1, user2)
		return []Message{}
	}

	log.Printf("Retrieved %d messages for conversation %s between %s and %s: %+v",
		len(store), key, user1, user2, store)
	return store
}

// getGroupHistory returns the message history for a group
func getGroupHistory(groupID string) []Message {
	key := "group_" + groupID
	msgMux.RLock()
	store, exists := messages[key]
	msgMux.RUnlock()

	if !exists {
		log.Printf("No message history found for group %s", key)
		return []Message{}
	}

	log.Printf("Retrieved %d messages for group %s", len(store), key)
	return store
}

func main() {
	// Clear messages when server starts
	msgMux.Lock()
	messages = make(map[string][]Message)
	msgMux.Unlock()

	// Get the embedded filesystem
	buildFS, err := static.GetBuildFS()
	if err != nil {
		log.Fatal("Failed to get build filesystem:", err)
	}

	// Create a file server for the React app
	fileServer := http.FileServer(http.FS(buildFS))

	// Create a new mux
	mux := http.NewServeMux()

	// Handle WebSocket connections
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		username := r.URL.Query().Get("username")
		if username == "" {
			http.Error(w, "Username is required", http.StatusBadRequest)
			return
		}

		log.Printf("New WebSocket connection request from user: %s", username)

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("Error upgrading connection for user %s: %v", username, err)
			return
		}

		log.Printf("WebSocket connection established for user: %s", username)

		client := &Client{
			Username: username,
			conn:     conn,
			send:     make(chan []byte, 256),
		}

		clientsMux.Lock()
		if existingClient, exists := clients[username]; exists {
			log.Printf("Closing existing connection for user %s", username)
			close(existingClient.send)
		}
		clients[username] = client
		clientsMux.Unlock()

		log.Printf("Registering new client for user: %s", username)

		// Send initial user list and group list
		log.Printf("Sending initial data to user: %s", username)
		sendUserList()
		sendGroupList()

		// Broadcast system message about new user
		broadcastSystemMessage(fmt.Sprintf("%s joined the chat", username))

		go client.writePump()
		go client.readPump()
	})

	// Serve static files for the React app
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// If the request is for an API endpoint, return 404
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}

		// If the request is for a file that exists in the build directory, serve it
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}

		// Try to serve the requested file
		if _, err := buildFS.Open(path); err == nil {
			fileServer.ServeHTTP(w, r)
			return
		}

		// For all other requests, serve index.html (React router will handle the routing)
		indexFile, err := buildFS.Open("index.html")
		if err != nil {
			http.NotFound(w, r)
			return
		}
		defer indexFile.Close()

		http.ServeContent(w, r, "index.html", time.Now(), indexFile.(io.ReadSeeker))
	})

	// Start the server
	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal("Error starting server:", err)
	}
}

func (c *Client) readPump() {
	defer func() {
		clientsMux.Lock()
		delete(clients, c.Username)
		clientsMux.Unlock()
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Error reading message from client %s: %v", c.Username, err)
			}
			break
		}

		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Error unmarshaling message from client %s: %v", c.Username, err)
			continue
		}

		msg.From = c.Username
		msg.Timestamp = time.Now().Format(time.RFC3339)

		// Store message
		storeMessage(msg)

		// Handle different message types
		switch msg.Type {
		case TypePrivateMessage:
			msgBytes, _ := json.Marshal(msg)
			sendToUser(msg.To, msgBytes)
		case TypeGroupMessage:
			msgBytes, _ := json.Marshal(msg)
			sendToGroup(msg.To, msgBytes)
		case TypeCreateGroup:
			createGroup(msg)
		case TypeAddGroupMember:
			addGroupMember(msg)
		case TypeRemoveGroupMember:
			removeGroupMember(msg)
		case TypeLeaveGroup:
			leaveGroup(msg)
		default:
			log.Printf("Unknown message type from client %s: %s", c.Username, msg.Type)
		}
	}
}

func (c *Client) writePump() {
	defer func() {
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		}
	}
}

func sendUserList() {
	log.Printf("Starting sendUserList()")
	clientsMux.RLock()
	userList := make(map[string]string)
	for username := range clients {
		userList[username] = username
	}
	clientsMux.RUnlock()

	message := map[string]interface{}{
		KeyType:      TypeUserList,
		KeyUsers:     userList,
		KeyTimestamp: time.Now().Format(time.RFC3339),
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling user list: %v", err)
		return
	}

	// Get a copy of clients to minimize lock time
	clientsMux.RLock()
	clientsCopy := make(map[string]*Client, len(clients))
	for username, client := range clients {
		clientsCopy[username] = client
	}
	clientsMux.RUnlock()

	log.Printf("Broadcasting user list to %d clients: %v", len(clientsCopy), userList)
	for username, client := range clientsCopy {
		select {
		case client.send <- messageBytes:
			log.Printf("User list sent to client: %s", username)
		default:
			log.Printf("Failed to send user list to client: %s", username)
			// Remove client
			clientsMux.Lock()
			delete(clients, username)
			clientsMux.Unlock()
			close(client.send)
		}
	}
	log.Printf("Finished sending user list")
}

func sendToUser(username string, message []byte) {
	clientsMux.RLock()
	client, exists := clients[username]
	clientsMux.RUnlock()

	if !exists {
		log.Printf("User %s not found", username)
		return
	}

	select {
	case client.send <- message:
		log.Printf("Message sent to user %s", username)
	default:
		log.Printf("Failed to send message to user %s", username)
		// Remove client
		clientsMux.Lock()
		delete(clients, username)
		clientsMux.Unlock()
		close(client.send)
	}
}

func sendToGroup(groupName string, message []byte) {
	groupsMux.RLock()
	group, exists := groups[groupName]
	groupsMux.RUnlock()

	if !exists {
		log.Printf("Group %s not found", groupName)
		return
	}

	// Get a copy of members to avoid holding the lock while sending
	members := make([]string, len(group.Members))
	copy(members, group.Members)

	for _, member := range members {
		sendToUser(member, message)
	}
}

func sendGroupList() {
	log.Printf("Starting sendGroupList()")
	groupsMux.RLock()
	groupsCopy := make(map[string]*Group, len(groups))
	for id, group := range groups {
		groupsCopy[id] = group
	}
	groupsMux.RUnlock()

	// Get a copy of clients to minimize lock time
	clientsMux.RLock()
	clientsCopy := make(map[string]*Client, len(clients))
	for username, client := range clients {
		clientsCopy[username] = client
	}
	clientsMux.RUnlock()

	// Send filtered group list to each user
	for username, client := range clientsCopy {
		// Filter groups for this user
		userGroups := make([]Group, 0)
		for _, group := range groupsCopy {
			if contains(group.Members, username) {
				userGroups = append(userGroups, *group)
			}
		}

		message := map[string]interface{}{
			KeyType:      TypeGroupList,
			KeyGroups:    userGroups,
			KeyTimestamp: time.Now().Format(time.RFC3339),
		}

		messageBytes, err := json.Marshal(message)
		if err != nil {
			log.Printf("Error marshaling group list for user %s: %v", username, err)
			continue
		}

		select {
		case client.send <- messageBytes:
			log.Printf("Group list sent to client: %s", username)
		default:
			log.Printf("Failed to send group list to client: %s", username)
			// Remove client
			clientsMux.Lock()
			delete(clients, username)
			clientsMux.Unlock()
			close(client.send)
		}
	}
	log.Printf("Finished sending group list")
}

// Helper function to check if a slice contains a string
func contains(slice []string, str string) bool {
	for _, v := range slice {
		if v == str {
			return true
		}
	}
	return false
}

func broadcastMessage(message []byte) {
	log.Printf("Starting broadcastMessage()")

	// Get a copy of clients to minimize lock time
	clientsMux.RLock()
	clientsCopy := make(map[string]*Client, len(clients))
	for username, client := range clients {
		clientsCopy[username] = client
	}
	clientsMux.RUnlock()

	log.Printf("Broadcasting message to %d clients", len(clientsCopy))
	for username, client := range clientsCopy {
		select {
		case client.send <- message:
			log.Printf("Message sent successfully to client %s", username)
		default:
			log.Printf("Failed to send message to client %s: channel full or closed", username)
			// Remove client
			clientsMux.Lock()
			delete(clients, username)
			clientsMux.Unlock()
			close(client.send)
		}
	}
	log.Printf("Finished broadcasting message")
}

func broadcastSystemMessage(content string) {
	message := Message{
		Type:      TypeSystem,
		Content:   content,
		Timestamp: time.Now().Format(time.RFC3339),
	}
	messageBytes, _ := json.Marshal(message)
	broadcastMessage(messageBytes)
}

// createGroup creates a new group
func createGroup(msg Message) {
	log.Printf("Creating group: %s by user: %s", msg.To, msg.From)

	// Parse members from content
	members := strings.Split(msg.Content, ",")
	if len(members) == 0 {
		log.Printf("No members provided for group creation")
		return
	}

	// Create new group
	group := &Group{
		Name:    msg.To,
		Admin:   msg.From,
		Members: append([]string{msg.From}, members...), // Add creator as first member
	}

	// Store group
	groupsMux.Lock()
	groups[msg.To] = group
	groupsMux.Unlock()

	// Notify group members
	notification := Message{
		Type:      TypeSystem,
		Content:   fmt.Sprintf("Group '%s' created by %s", msg.To, msg.From),
		Timestamp: time.Now().Format(time.RFC3339),
	}
	msgBytes, _ := json.Marshal(notification)
	sendToGroup(msg.To, msgBytes)

	// Update group list for all users
	sendGroupList()
}

// addGroupMember adds a member to a group
func addGroupMember(msg Message) {
	log.Printf("Adding member %s to group %s", msg.Content, msg.To)

	groupsMux.Lock()
	group, exists := groups[msg.To]
	if !exists {
		groupsMux.Unlock()
		log.Printf("Group %s not found", msg.To)
		return
	}

	// Check if user is admin
	if group.Admin != msg.From {
		groupsMux.Unlock()
		log.Printf("User %s is not authorized to add members", msg.From)
		return
	}

	// Add new member
	group.Members = append(group.Members, msg.Content)
	groupsMux.Unlock()

	// Notify group members
	notification := Message{
		Type:      TypeSystem,
		Content:   fmt.Sprintf("%s added %s to the group", msg.From, msg.Content),
		Timestamp: time.Now().Format(time.RFC3339),
	}
	msgBytes, _ := json.Marshal(notification)
	sendToGroup(msg.To, msgBytes)

	// Update group list
	sendGroupList()
}

// removeGroupMember removes a member from a group
func removeGroupMember(msg Message) {
	log.Printf("Removing member %s from group %s", msg.Content, msg.To)

	groupsMux.Lock()
	group, exists := groups[msg.To]
	if !exists {
		groupsMux.Unlock()
		log.Printf("Group %s not found", msg.To)
		return
	}

	// Check if user is admin
	if group.Admin != msg.From {
		groupsMux.Unlock()
		log.Printf("User %s is not authorized to remove members", msg.From)
		return
	}

	// Remove member
	newMembers := make([]string, 0, len(group.Members))
	for _, member := range group.Members {
		if member != msg.Content {
			newMembers = append(newMembers, member)
		}
	}
	group.Members = newMembers
	groupsMux.Unlock()

	// Notify group members
	notification := Message{
		Type:      TypeSystem,
		Content:   fmt.Sprintf("%s removed %s from the group", msg.From, msg.Content),
		Timestamp: time.Now().Format(time.RFC3339),
	}
	msgBytes, _ := json.Marshal(notification)
	sendToGroup(msg.To, msgBytes)

	// Update group list
	sendGroupList()
}

// leaveGroup allows a user to leave a group
func leaveGroup(msg Message) {
	log.Printf("User %s leaving group %s", msg.From, msg.To)

	groupsMux.Lock()
	group, exists := groups[msg.To]
	if !exists {
		groupsMux.Unlock()
		log.Printf("Group %s not found", msg.To)
		return
	}

	// Remove member
	newMembers := make([]string, 0, len(group.Members))
	for _, member := range group.Members {
		if member != msg.From {
			newMembers = append(newMembers, member)
		}
	}
	group.Members = newMembers

	// If group is empty, delete it
	if len(group.Members) == 0 {
		delete(groups, msg.To)
		groupsMux.Unlock()
		log.Printf("Group %s deleted as it's empty", msg.To)
	} else {
		// If admin left, assign new admin
		if group.Admin == msg.From {
			group.Admin = group.Members[0]
			log.Printf("New admin for group %s: %s", msg.To, group.Admin)
		}
		groupsMux.Unlock()
	}

	// Notify group members
	notification := Message{
		Type:      TypeSystem,
		Content:   fmt.Sprintf("%s left the group", msg.From),
		Timestamp: time.Now().Format(time.RFC3339),
	}
	msgBytes, _ := json.Marshal(notification)
	sendToGroup(msg.To, msgBytes)

	// Update group list
	sendGroupList()
}
