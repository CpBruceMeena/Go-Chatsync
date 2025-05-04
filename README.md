# ChatSync - Real-time WebSocket Chat Application

ChatSync is a modern, real-time chat application built with Go and WebSocket technology. It supports both private messaging and group chats with a clean, intuitive interface.

## Features

### Core Features
- Real-time messaging using WebSocket
- Private one-to-one chat
- Group chat functionality
- Message history for private chats
- User presence tracking
- System notifications

### Group Management
- Create groups with custom names
- Add/remove group members
- Group admin controls
- Group member list display

### User Interface
- Clean, modern design
- Separate panels for users and groups
- Real-time message updates
- Message timestamps
- System notifications
- User online status

## Technical Architecture

### Backend (Go)
- WebSocket server implementation
- Concurrent message handling
- Group and user management
- Message history storage
- Real-time broadcasting

### Frontend (HTML/CSS/JavaScript)
- Responsive design
- Real-time UI updates
- Dynamic group/user lists
- Message history display
- Error handling and notifications

## Getting Started

### Prerequisites
- Go 1.16 or higher
- Modern web browser

### Installation
1. Clone the repository:
```bash
git clone https://github.com/yourusername/chatsync.git
cd chatsync
```

2. Run the server:
```bash
go run main.go
```

3. Access the application:
Open your browser and navigate to `http://localhost:8080`

### Usage
1. Enter your username to join the chat
2. Select a user from the list to start a private chat
3. Create or join groups using the group panel
4. Send messages in private or group chats

## Message Types
- Private Messages: One-to-one communication
- Group Messages: Broadcast to all group members
- System Messages: Notifications and status updates

## Security Features
- Username-based authentication
- Group admin controls
- Message validation
- Connection state management

## Future Improvements
- Message persistence
- File sharing
- User authentication
- Message encryption
- Mobile responsiveness
- Message search
- User profiles
- Message reactions

## Contributing
Contributions are welcome! Please feel free to submit a Pull Request.

## License
This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments
- Built with Go and WebSocket
- Inspired by modern chat applications
- Designed for simplicity and reliability 