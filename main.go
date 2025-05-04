package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all connections in development
	},
}

// Message represents the structure of our chat messages
type Message struct {
	Type      string   `json:"type"`       // message types: private_message, group_message, system
	From      string   `json:"from"`       // sender's username
	To        string   `json:"to"`         // recipient's username
	Content   string   `json:"content"`    // message content
	Timestamp string   `json:"timestamp"`  // message timestamp
	GroupID   string   `json:"group_id"`   // group identifier
	GroupName string   `json:"group_name"` // group name for creation
	Member    string   `json:"member"`     // member username for group operations
	Members   []string `json:"members"`    // multiple members for group operations
}

// MessageStore represents a conversation between two users or a group
type MessageStore struct {
	Messages []Message
	mu       sync.RWMutex
}

type Client struct {
	conn     *websocket.Conn
	send     chan []byte
	ID       string
	Username string
	LastSeen time.Time
	Groups   map[string]bool
}

type Group struct {
	ID      string
	Name    string
	Members map[string]*Client
	Admin   string
}

var (
	// Store clients by username instead of ID
	clients         = make(map[string]*Client)
	groups          = make(map[string]*Group)
	messageStore    = make(map[string]*MessageStore)
	clientsMux      sync.Mutex
	messageStoreMux sync.RWMutex
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
	if msg.Type != "private_message" && msg.Type != "group_message" {
		return
	}

	if msg.Type == "private_message" {
		key := getConversationKey(msg.From, msg.To)
		messageStoreMux.Lock()
		if _, exists := messageStore[key]; !exists {
			messageStore[key] = &MessageStore{
				Messages: make([]Message, 0),
			}
		}
		messageStore[key].mu.Lock()
		messageStore[key].Messages = append(messageStore[key].Messages, msg)
		messageStore[key].mu.Unlock()
		messageStoreMux.Unlock()
		log.Printf("Stored private message in conversation %s: %+v", key, msg)
	} else if msg.Type == "group_message" {
		key := "group:" + msg.GroupID
		messageStoreMux.Lock()
		if _, exists := messageStore[key]; !exists {
			messageStore[key] = &MessageStore{
				Messages: make([]Message, 0),
			}
		}
		messageStore[key].mu.Lock()
		messageStore[key].Messages = append(messageStore[key].Messages, msg)
		messageStore[key].mu.Unlock()
		messageStoreMux.Unlock()
		log.Printf("Stored group message in group %s: %+v", key, msg)
	}
}

// getConversationHistory returns the message history for a conversation
func getConversationHistory(user1, user2 string) []Message {
	key := getConversationKey(user1, user2)
	messageStoreMux.RLock()
	store, exists := messageStore[key]
	messageStoreMux.RUnlock()

	if !exists {
		log.Printf("No message history found for conversation %s", key)
		return []Message{}
	}

	store.mu.RLock()
	messages := make([]Message, len(store.Messages))
	copy(messages, store.Messages)
	store.mu.RUnlock()

	log.Printf("Retrieved %d messages for conversation %s", len(messages), key)
	return messages
}

// getGroupHistory returns the message history for a group
func getGroupHistory(groupID string) []Message {
	key := "group:" + groupID
	messageStoreMux.RLock()
	store, exists := messageStore[key]
	messageStoreMux.RUnlock()

	if !exists {
		log.Printf("No message history found for group %s", key)
		return []Message{}
	}

	store.mu.RLock()
	messages := make([]Message, len(store.Messages))
	copy(messages, store.Messages)
	store.mu.RUnlock()

	log.Printf("Retrieved %d messages for group %s", len(messages), key)
	return messages
}

func main() {
	http.HandleFunc("/", serveHome)
	http.HandleFunc("/ws", handleWebSocket)

	fmt.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/index.html")
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading connection: %v", err)
		return
	}

	client := &Client{
		conn:     conn,
		send:     make(chan []byte, 256),
		ID:       uuid.New().String(),
		Username: username,
		LastSeen: time.Now(),
		Groups:   make(map[string]bool),
	}

	clientsMux.Lock()
	// Check if username already exists
	if existingClient, exists := clients[username]; exists {
		// Close existing connection
		existingClient.conn.Close()
		log.Printf("Closed existing connection for user %s", username)
	}
	clients[username] = client
	clientsMux.Unlock()

	// Send client ID to the client
	clientIdMessage := map[string]interface{}{
		"type":      "client_id",
		"client_id": client.ID,
	}
	clientIdBytes, _ := json.Marshal(clientIdMessage)
	client.send <- clientIdBytes

	// Send initial user list and group list
	sendUserList()
	sendGroupList()

	// Broadcast user joined message
	broadcastSystemMessage(fmt.Sprintf("%s joined the chat", username))

	go client.writePump()
	go client.readPump()
}

func (c *Client) readPump() {
	defer func() {
		c.conn.Close()
		clientsMux.Lock()
		delete(clients, c.Username)
		clientsMux.Unlock()
		broadcastSystemMessage(fmt.Sprintf("%s left the chat", c.Username))
		sendUserList()
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			continue
		}

		// Set the sender's username
		msg.From = c.Username
		msg.Timestamp = time.Now().Format(time.RFC3339)

		log.Printf("Received message: %+v", msg)

		switch msg.Type {
		case "private_message":
			if msg.To != "" {
				// Verify the recipient exists
				clientsMux.Lock()
				_, exists := clients[msg.To]
				clientsMux.Unlock()

				if !exists {
					// Send error message to sender
					errorMsg := Message{
						Type:      "system",
						Content:   fmt.Sprintf("User %s is not online", msg.To),
						Timestamp: time.Now().Format(time.RFC3339),
					}
					sendToUser(errorMsg)
					continue
				}

				// Store the message
				storeMessage(msg)

				// Send to recipient
				sendToUser(msg)

				// Send confirmation to sender
				confirmation := Message{
					Type:      "system",
					Content:   "Message sent",
					Timestamp: time.Now().Format(time.RFC3339),
				}
				sendToUser(confirmation)
			}
		case "get_history":
			// Handle history request
			if msg.To != "" {
				history := getConversationHistory(c.Username, msg.To)
				historyBytes, _ := json.Marshal(map[string]interface{}{
					"type":    "history",
					"from":    c.Username,
					"to":      msg.To,
					"history": history,
				})
				log.Printf("Sending private history response: %+v", history)
				c.send <- historyBytes
			}
		case "get_group_history":
			// Handle group history request
			if msg.GroupID != "" {
				history := getGroupHistory(msg.GroupID)
				historyBytes, _ := json.Marshal(map[string]interface{}{
					"type":     "group_history",
					"group_id": msg.GroupID,
					"history":  history,
				})
				log.Printf("Sending group history response: %+v", history)
				c.send <- historyBytes
			}
		case "group_message":
			if msg.GroupID != "" {
				// Verify the group exists
				clientsMux.Lock()
				_, exists := groups[msg.GroupID]
				clientsMux.Unlock()

				if !exists {
					// Send error message to sender
					errorMsg := Message{
						Type:      "system",
						Content:   fmt.Sprintf("Group %s does not exist", msg.GroupID),
						Timestamp: time.Now().Format(time.RFC3339),
					}
					sendToUser(errorMsg)
					continue
				}

				// Store the message
				storeMessage(msg)

				// Send to group
				sendToGroup(msg)
			}
		case "group_create":
			createGroup(msg)
		case "add_group_member":
			addGroupMember(msg)
		case "remove_group_member":
			removeGroupMember(msg)
		default:
			log.Printf("Unknown message type: %s", msg.Type)
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
	clientsMux.Lock()
	userList := make(map[string]string)
	for username := range clients {
		userList[username] = username
	}
	clientsMux.Unlock()

	messageBytes, _ := json.Marshal(map[string]interface{}{
		"type":  "user_list",
		"users": userList,
	})

	log.Printf("Broadcasting user list: %+v", userList)
	broadcastMessage(messageBytes)
}

func sendToUser(message Message) {
	clientsMux.Lock()
	defer clientsMux.Unlock()

	messageBytes, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	// Find recipient by username
	var recipientFound bool
	if recipient, exists := clients[message.To]; exists {
		select {
		case recipient.send <- messageBytes:
			recipientFound = true
			log.Printf("Message sent to %s: %s", recipient.Username, message.Content)
		default:
			log.Printf("Failed to send message to %s: channel full or closed", recipient.Username)
		}
	}

	// Send to sender (for confirmation)
	if sender, exists := clients[message.From]; exists {
		select {
		case sender.send <- messageBytes:
			log.Printf("Message confirmation sent to sender %s: %s", sender.Username, message.Content)
		default:
			log.Printf("Failed to send confirmation to sender %s: channel full or closed", sender.Username)
		}
	}

	if !recipientFound {
		log.Printf("Recipient %s not found for message from %s", message.To, message.From)
	}
}

func sendToGroup(message Message) {
	messageBytes, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	if group, exists := groups[message.GroupID]; exists {
		// Send to all group members
		for _, member := range group.Members {
			select {
			case member.send <- messageBytes:
				// Message sent successfully
			default:
				// Channel is full or closed, remove the client
				delete(group.Members, member.Username)
				delete(member.Groups, message.GroupID)
			}
		}

		// If group is empty after sending, delete it
		if len(group.Members) == 0 {
			delete(groups, message.GroupID)
			sendGroupList()
		}
	}
}

func sendGroupList() {
	log.Printf("Starting sendGroupList()")

	groupList := make(map[string]map[string]interface{})
	for id, group := range groups {
		members := make([]string, 0, len(group.Members))
		for username := range group.Members {
			members = append(members, username)
		}
		groupList[id] = map[string]interface{}{
			"name":    group.Name,
			"members": members,
			"admin":   group.Admin,
		}
		log.Printf("Added group to list - ID: %s, Name: %s, Members: %v", id, group.Name, members)
	}

	messageBytes, err := json.Marshal(map[string]interface{}{
		"type":   "group_list",
		"groups": groupList,
	})
	if err != nil {
		log.Printf("Error marshaling group list: %v", err)
		return
	}

	log.Printf("Broadcasting group list: %+v", groupList)
	broadcastMessage(messageBytes)
	log.Printf("Finished sendGroupList()")
}

func broadcastMessage(message []byte) {
	log.Printf("Starting broadcastMessage()")
	clientsMux.Lock()
	defer clientsMux.Unlock()

	log.Printf("Broadcasting message to %d clients", len(clients))
	for username, client := range clients {
		select {
		case client.send <- message:
			log.Printf("Message sent successfully to client %s", username)
		default:
			log.Printf("Failed to send message to client %s: channel full or closed", username)
		}
	}
	log.Printf("Finished broadcasting message")
}

func broadcastSystemMessage(content string) {
	message := Message{
		Type:      "system",
		Content:   content,
		Timestamp: time.Now().Format(time.RFC3339),
	}
	messageBytes, _ := json.Marshal(message)
	broadcastMessage(messageBytes)
}

func createGroup(message Message) {
	log.Printf("Creating group with message: %+v", message)

	// Use group name as the group ID
	groupID := message.GroupName
	group := &Group{
		ID:      groupID,
		Name:    message.GroupName,
		Members: make(map[string]*Client),
		Admin:   message.From,
	}

	// Add creator as member
	if creator, exists := clients[message.From]; exists {
		log.Printf("Adding creator %s to group %s", creator.Username, groupID)
		group.Members[creator.Username] = creator
		creator.Groups[groupID] = true

		// Send confirmation to creator
		messageBytes, _ := json.Marshal(map[string]interface{}{
			"type":     "group_created",
			"group_id": groupID,
			"name":     group.Name,
		})
		creator.send <- messageBytes
		log.Printf("Group creation confirmation sent to creator")
	} else {
		log.Printf("Creator %s not found in clients", message.From)
		return
	}

	// Add initial members if provided
	if len(message.Members) > 0 {
		for _, memberUsername := range message.Members {
			if member, exists := clients[memberUsername]; exists {
				// Add member to group
				group.Members[memberUsername] = member
				member.Groups[groupID] = true

				// Notify the new member
				notification := Message{
					Type:      "system",
					Content:   fmt.Sprintf("You have been added to group %s", group.Name),
					Timestamp: time.Now().Format(time.RFC3339),
					GroupID:   groupID,
					GroupName: group.Name,
				}
				sendToUser(notification)
			} else {
				log.Printf("User %s not found", memberUsername)
			}
		}
	}

	groups[groupID] = group
	log.Printf("Group %s created successfully with %d members", groupID, len(group.Members))

	// Update group list for all clients
	log.Printf("Calling sendGroupList() after group creation")
	sendGroupList()

	// Broadcast system message
	if creator, exists := clients[message.From]; exists {
		memberCount := len(group.Members)
		if memberCount > 1 {
			broadcastSystemMessage(fmt.Sprintf("Group '%s' created by %s with %d members", group.Name, creator.Username, memberCount))
		} else {
			broadcastSystemMessage(fmt.Sprintf("Group '%s' created by %s", group.Name, creator.Username))
		}
	}
}

func addGroupMember(message Message) {
	clientsMux.Lock()
	defer clientsMux.Unlock()

	group, exists := groups[message.GroupID]
	if !exists {
		log.Printf("Group %s does not exist", message.GroupID)
		return
	}

	// Check if the sender is the admin
	if group.Admin != message.From {
		log.Printf("User %s is not authorized to add members to group %s", message.From, message.GroupID)
		return
	}

	// Handle multiple members if provided
	if len(message.Members) > 0 {
		for _, memberUsername := range message.Members {
			if member, exists := clients[memberUsername]; exists {
				// Add member to group
				group.Members[memberUsername] = member
				member.Groups[message.GroupID] = true

				// Notify the new member
				notification := Message{
					Type:      "system",
					Content:   fmt.Sprintf("You have been added to group %s", group.Name),
					Timestamp: time.Now().Format(time.RFC3339),
					GroupID:   message.GroupID,
					GroupName: group.Name,
				}
				sendToUser(notification)

				// Notify group members
				groupNotification := Message{
					Type:      "system",
					Content:   fmt.Sprintf("%s has been added to the group", memberUsername),
					Timestamp: time.Now().Format(time.RFC3339),
					GroupID:   message.GroupID,
					GroupName: group.Name,
				}
				sendToGroup(groupNotification)
			} else {
				log.Printf("User %s not found", memberUsername)
			}
		}
	} else if message.Member != "" {
		// Handle single member (backward compatibility)
		if member, exists := clients[message.Member]; exists {
			// Add member to group
			group.Members[message.Member] = member
			member.Groups[message.GroupID] = true

			// Notify the new member
			notification := Message{
				Type:      "system",
				Content:   fmt.Sprintf("You have been added to group %s", group.Name),
				Timestamp: time.Now().Format(time.RFC3339),
				GroupID:   message.GroupID,
				GroupName: group.Name,
			}
			sendToUser(notification)

			// Notify group members
			groupNotification := Message{
				Type:      "system",
				Content:   fmt.Sprintf("%s has been added to the group", message.Member),
				Timestamp: time.Now().Format(time.RFC3339),
				GroupID:   message.GroupID,
				GroupName: group.Name,
			}
			sendToGroup(groupNotification)
		} else {
			log.Printf("User %s not found", message.Member)
		}
	}

	// Update group list for all users
	sendGroupList()
}

func removeGroupMember(message Message) {
	if group, exists := groups[message.GroupID]; exists {
		// Check if the sender is the group admin
		if group.Admin != message.From {
			log.Printf("User %s is not authorized to remove members from group %s", message.From, message.GroupID)
			return
		}

		// Find the client to remove
		if member, exists := clients[message.Member]; exists {
			// Remove member from group
			delete(group.Members, member.Username)
			delete(member.Groups, message.GroupID)

			// Notify all group members
			broadcastSystemMessage(fmt.Sprintf("%s removed %s from group '%s'", message.From, message.Member, group.Name))
		}

		// If group is empty after removal, delete it
		if len(group.Members) == 0 {
			delete(groups, message.GroupID)
		}

		// Update group list for all clients
		sendGroupList()
	}
}
