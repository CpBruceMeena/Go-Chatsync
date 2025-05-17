# Go-Chatsync

A real-time chat application built with Go and React, featuring private messaging and group chat functionality.

## Features

- Real-time messaging using WebSocket
- Private messaging between users
- Group chat functionality
  - Create new groups
  - Add/remove members
  - Group-specific message history
- Modern UI with Material-UI components
- Responsive design
- Message history persistence
- User presence tracking

## WebSocket Implementation

### Overview
The application uses WebSocket for real-time bidirectional communication between the server and clients. This implementation ensures instant message delivery, efficient connection management, and robust error handling.

### Connection Management
- **Connection Establishment**: When a user connects to the application, their HTTP connection is upgraded to a WebSocket connection
- **Client Registration**: Each connected client is registered with a unique username and maintained in an active clients map
- **Connection Monitoring**: The server tracks all active connections and handles disconnections gracefully
- **User Presence**: Real-time tracking of online/offline status for all users

### Message Flow
1. **Message Types**:
   - Private Messages: One-to-one communication between users
   - Group Messages: Communication within group channels
   - System Messages: Server notifications and status updates
   - History Requests: Retrieving message history
   - Group Management: Creating and managing groups

2. **Message Processing**:
   - Messages are sent as JSON objects with type, content, and metadata
   - Server validates and processes messages based on their type
   - Messages are stored in memory for the current session
   - Recipients receive messages in real-time if online

3. **Message Delivery**:
   - Private messages are delivered directly to the intended recipient
   - Group messages are broadcast to all online group members
   - System messages are sent to relevant users
   - Message acknowledgments are sent back to the sender

### Real-time Features
1. **Instant Messaging**:
   - Messages are delivered immediately to online users
   - No polling or refresh required
   - Efficient use of server resources

2. **User Status Updates**:
   - Real-time online/offline status
   - Automatic status updates when users connect/disconnect
   - Presence indicators in the UI

3. **Group Management**:
   - Real-time group creation and updates
   - Instant member addition/removal notifications
   - Live group member list updates

### Error Handling
- **Connection Errors**: Automatic handling of connection drops
- **Message Errors**: Validation and error reporting for malformed messages
- **Recovery**: Automatic reconnection attempts from the client
- **Cleanup**: Proper resource cleanup on disconnection

### Performance Considerations
- **Efficient Communication**: WebSocket maintains a single persistent connection
- **Resource Management**: Proper handling of connection resources
- **Scalability**: Support for multiple concurrent connections
- **Message Buffering**: Efficient message queuing and delivery

### Security Features
- **Connection Validation**: Secure WebSocket upgrade process
- **Message Validation**: Input validation for all messages
- **Error Isolation**: Errors in one connection don't affect others
- **Resource Protection**: Prevention of resource exhaustion

## Project Structure

```
.
├── frontend/           # React application
│   ├── src/           # Source code
│   │   ├── components/    # React components
│   │   ├── contexts/      # React contexts
│   │   └── theme.js       # Material-UI theme
│   ├── public/        # Static files
│   ├── package.json   # Frontend dependencies
│   └── README.md      # Frontend documentation
├── backend/           # Go application
│   ├── main.go       # Backend entry point
│   ├── static/       # Static assets
│   ├── scripts/      # Build scripts
│   ├── go.mod        # Go module file
│   └── go.sum        # Go module checksum
└── run.sh            # Build and run script
```

## Getting Started

### Prerequisites

- Go 1.16 or later
- Node.js 14 or later
- npm or yarn

### Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/CpBruceMeena/Go-Chatsync.git
   cd Go-Chatsync
   ```

2. Install frontend dependencies:
   ```bash
   cd frontend
   npm install
   ```

3. Install backend dependencies:
   ```bash
   cd ../backend
   go mod download
   ```

### Running the Application

1. Start the application using the provided script:
   ```bash
   ./run.sh
   ```

   This will:
   - Build the frontend
   - Start the backend server
   - Serve the application at http://localhost:8080

## Features in Detail

### Private Messaging
- One-to-one chat between users
- Real-time message delivery
- Message history persistence
- Online/offline status indicators

### Group Chat
- Create new groups with custom names
- Add or remove members from groups
- Group-specific message history
- Member count display
- Group visibility limited to members only

### User Interface
- Clean and modern Material-UI design
- Responsive layout
- Intuitive navigation
- Real-time updates
- Message timestamps
- User presence indicators

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details. 