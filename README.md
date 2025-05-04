# ChatSync - Real-time Chat Application

ChatSync is a modern, real-time chat application built with Go and WebSocket technology. It provides a seamless messaging experience with support for both private and group chats, along with real-time user status updates.

## Features

- **Real-time Messaging**: Instant message delivery using WebSocket technology
- **Private Chat**: One-to-one conversations between users
- **Group Chat**: Create and manage group conversations
- **User Status**: Real-time online/offline status indicators
- **Message Persistence**: Messages are stored and retrieved for both private and group chats
- **Modern UI**: Clean and intuitive user interface
- **Responsive Design**: Works seamlessly across different devices

## Prerequisites

- Go 1.16 or higher
- Modern web browser with WebSocket support

## Quick Start

The easiest way to run ChatSync is using the provided run script:

```bash
# Clone the repository
git clone https://github.com/CpBruceMeena/Go-Chatsync.git
cd Go-Chatsync

# Run the application
./run.sh
```

The application will be available at `http://localhost:8080`

## Manual Setup

If you prefer to set up manually, follow these steps:

1. **Install Dependencies**
   ```bash
   go mod tidy
   ```

2. **Build the Project**
   ```bash
   go build -o chatsync
   ```

3. **Run the Application**
   ```bash
   ./chatsync
   ```

## Project Structure

```
ChatSync/
├── main.go           # Main application entry point
├── static/          # Static files
│   └── index.html   # Frontend application
├── go.mod           # Go module file
├── go.sum           # Go module checksum
└── run.sh          # Automated run script
```

## Features in Detail

### Private Chat
- Real-time one-to-one messaging
- Message history persistence
- Online/offline status indicators
- Message timestamps

### Group Chat
- Create new groups
- Add/remove members
- Group message history
- Member management
- Admin controls for group owners

### User Interface
- Clean and modern design
- Intuitive navigation
- Real-time status updates
- Responsive layout
- Easy group management

## Technical Details

### Backend
- Built with Go
- WebSocket for real-time communication
- Concurrent message handling
- Message persistence
- User session management

### Frontend
- Modern HTML5/CSS3
- JavaScript for real-time updates
- WebSocket client implementation
- Responsive design
- Font Awesome icons

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For support, please open an issue in the GitHub repository. 