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
	TypeRequestHistory    = "request_history"
	TypeUpdateLastSeen    = "update_last_seen" // New type for updating last seen timestamp

	// Backend Storage
	TypePrivate = "private"
	TypeGroup   = "group"

	// Backend to Frontend
	TypeUserList    = "user_list"
	TypeGroupList   = "group_list"
	TypeSystem      = "system"
	TypeHistory     = "history"
	TypeUnreadCount = "unread_count" // New type for sending unread message counts
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

	// Message storage
	privateMessages = make(map[string][]Message) // key: "user1:user2"
	groupMessages   = make(map[string][]Message) // key: "group_name"
	msgMux          sync.RWMutex

	// Last seen tracking
	lastSeenTimestamps = make(map[string]map[string]string) // key: username -> map[chatID]timestamp
	lastSeenMux        sync.RWMutex
)

// getConversationKey returns a consistent key for a conversation between two users
func getConversationKey(user1, user2 string) string {
	if user1 < user2 {
		return user1 + ":" + user2
	}
	return user2 + ":" + user1
}

// storeMessage stores a message in the appropriate message history
func storeMessage(msg Message) {
	msgMux.Lock()
	defer msgMux.Unlock()

	if msg.Type == TypePrivateMessage {
		key := getConversationKey(msg.From, msg.To)
		privateMessages[key] = append(privateMessages[key], msg)
		log.Printf("Stored private message: from=%s, to=%s, key=%s, total_messages=%d",
			msg.From, msg.To, key, len(privateMessages[key]))
	}

	if msg.Type == TypeGroupMessage {
		key := msg.To // Use group name as key
		groupMessages[key] = append(groupMessages[key], msg)
		log.Printf("Stored group message: group=%s, key=%s, total_messages=%d",
			msg.To, key, len(groupMessages[key]))
	}
}

// getConversationHistory returns the message history for a conversation
func getConversationHistory(user1, user2 string) []Message {
	key := getConversationKey(user1, user2)
	msgMux.RLock()
	store, exists := privateMessages[key]
	msgMux.RUnlock()

	if !exists {
		log.Printf("No message history found for conversation %s between %s and %s", key, user1, user2)
		return []Message{}
	}

	log.Printf("Retrieved %d messages for conversation %s between %s and %s",
		len(store), key, user1, user2)
	return store
}

// getGroupHistory returns the message history for a group
func getGroupHistory(groupID string) []Message {
	msgMux.RLock()
	store, exists := groupMessages[groupID]
	msgMux.RUnlock()

	if !exists {
		log.Printf("No message history found for group %s", groupID)
		return []Message{}
	}

	log.Printf("Retrieved %d messages for group %s", len(store), groupID)
	return store
}

func main() {
	// Clear messages when server starts
	msgMux.Lock()
	privateMessages = make(map[string][]Message)
	groupMessages = make(map[string][]Message)
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

		switch msg.Type {
		case TypePrivateMessage:
			// Store message
			storeMessage(msg)
			// Send to recipient
			msgBytes, _ := json.Marshal(msg)
			sendToUser(msg.To, msgBytes)
			// Send unread counts to recipient
			sendUnreadCounts(msg.To)
		case TypeGroupMessage:
			// Store message
			storeMessage(msg)
			// Send to group members
			msgBytes, _ := json.Marshal(msg)
			sendToGroup(msg.To, msgBytes)
			// Send unread counts to all group members
			groupsMux.RLock()
			if group, ok := groups[msg.To]; ok {
				for _, member := range group.Members {
					if member != msg.From {
						sendUnreadCounts(member)
					}
				}
			}
			groupsMux.RUnlock()
		case TypeUpdateLastSeen:
			// Update last seen timestamp
			updateLastSeen(c.Username, msg.To, msg.Timestamp)
			// Send updated unread counts
			sendUnreadCounts(c.Username)
		case TypeRequestHistory:
			// Send message history
			sendMessageHistory(c, msg.To, msg.Content)
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

// sendMessageHistory sends the message history to a client
func sendMessageHistory(client *Client, chatType, chatID string) {
	var history []Message

	if chatType == "private" {
		history = getConversationHistory(client.Username, chatID)
	} else if chatType == "group" {
		history = getGroupHistory(chatID)
	}

	message := map[string]interface{}{
		KeyType:      TypeHistory,
		KeyContent:   history,
		KeyTimestamp: time.Now().Format(time.RFC3339),
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling history for client %s: %v", client.Username, err)
		return
	}

	select {
	case client.send <- messageBytes:
		log.Printf("Message history sent to client: %s", client.Username)
	default:
		log.Printf("Failed to send message history to client: %s", client.Username)
	}
}

// updateLastSeen updates the last seen timestamp for a user's chat
func updateLastSeen(username, chatID string, timestamp string) {
	lastSeenMux.Lock()
	defer lastSeenMux.Unlock()

	if _, exists := lastSeenTimestamps[username]; !exists {
		lastSeenTimestamps[username] = make(map[string]string)
	}

	// Use the provided timestamp if it's valid, otherwise use current time
	if timestamp != "" {
		// Validate the timestamp format
		if _, err := time.Parse(time.RFC3339, timestamp); err == nil {
			lastSeenTimestamps[username][chatID] = timestamp
			log.Printf("Updated last seen for %s in chat %s to %s", username, chatID, timestamp)
			return
		}
	}

	// Fallback to current time if timestamp is invalid
	currentTime := time.Now().Format(time.RFC3339)
	lastSeenTimestamps[username][chatID] = currentTime
	log.Printf("Updated last seen for %s in chat %s to current time %s", username, chatID, currentTime)
}

// getUnreadCount returns the number of unread messages for a user in a chat
func getUnreadCount(username, chatID string) int {
	lastSeenMux.RLock()
	lastSeen, exists := lastSeenTimestamps[username][chatID]
	lastSeenMux.RUnlock()

	if !exists {
		// If no last seen timestamp, count all messages
		msgMux.RLock()
		defer msgMux.RUnlock()

		count := 0
		// Check private messages
		for key, messages := range privateMessages {
			parts := strings.Split(key, ":")
			if len(parts) != 2 {
				continue
			}

			otherUser := parts[0]
			if otherUser == username {
				otherUser = parts[1]
			}

			if otherUser == chatID {
				for _, msg := range messages {
					if msg.From != username {
						count++
					}
				}
			}
		}

		// Check group messages
		if messages, exists := groupMessages[chatID]; exists {
			for _, msg := range messages {
				if msg.From != username {
					count++
				}
			}
		}
		return count
	}

	lastSeenTime, err := time.Parse(time.RFC3339, lastSeen)
	if err != nil {
		log.Printf("Error parsing last seen timestamp: %v", err)
		return 0
	}

	msgMux.RLock()
	defer msgMux.RUnlock()

	count := 0

	// Check private messages
	for key, messages := range privateMessages {
		parts := strings.Split(key, ":")
		if len(parts) != 2 {
			continue
		}

		otherUser := parts[0]
		if otherUser == username {
			otherUser = parts[1]
		}

		if otherUser == chatID {
			for _, msg := range messages {
				if msg.From != username {
					msgTime, err := time.Parse(time.RFC3339, msg.Timestamp)
					if err != nil {
						continue
					}
					if msgTime.After(lastSeenTime) {
						count++
					}
				}
			}
		}
	}

	// Check group messages
	if messages, exists := groupMessages[chatID]; exists {
		for _, msg := range messages {
			if msg.From != username {
				msgTime, err := time.Parse(time.RFC3339, msg.Timestamp)
				if err != nil {
					continue
				}
				if msgTime.After(lastSeenTime) {
					count++
				}
			}
		}
	}

	log.Printf("Unread count for %s in %s: %d", username, chatID, count)
	return count
}

// sendUnreadCounts sends unread message counts to a user
func sendUnreadCounts(username string) {
	unreadCounts := make(map[string]int)

	// Get unread counts for private chats
	clientsMux.RLock()
	for otherUser := range clients {
		if otherUser != username {
			count := getUnreadCount(username, otherUser)
			if count > 0 {
				unreadCounts[otherUser] = count
			}
		}
	}
	clientsMux.RUnlock()

	// Get unread counts for groups
	groupsMux.RLock()
	for groupName, group := range groups {
		if contains(group.Members, username) {
			count := getUnreadCount(username, groupName)
			if count > 0 {
				unreadCounts[groupName] = count
			}
		}
	}
	groupsMux.RUnlock()

	// Convert unreadCounts to JSON string
	countsJSON, err := json.Marshal(unreadCounts)
	if err != nil {
		log.Printf("Error marshaling unread counts: %v", err)
		return
	}

	// Send unread counts to user
	message := Message{
		Type:      TypeUnreadCount,
		From:      "system",
		To:        username,
		Content:   string(countsJSON),
		Timestamp: time.Now().Format(time.RFC3339),
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling unread counts message: %v", err)
		return
	}

	log.Printf("Sending unread counts to %s: %s", username, string(countsJSON))
	sendToUser(username, messageBytes)
}
