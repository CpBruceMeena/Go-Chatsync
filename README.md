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
- **Group Management**: Create groups with multiple members, add/remove members, view group members, and group-specific chat history

## UI Features

- **Clean and modern interface**: Provides a user-friendly and visually appealing experience
- **Responsive design**: Works well on various devices, including desktops, tablets, and smartphones
- **Real-time status updates**: Users can see the online/offline status of other users in real-time
- **Intuitive group creation dialog**: Allows users to easily create new groups
- **Member management interface**: Provides tools to manage group members and their permissions
- **Message history with timestamps**: Allows users to view the history of messages in both private and group chats
- **Online/offline indicators**: Users can see if other users are online or offline
- **System notifications**: Users receive notifications about new messages or group events

## Group Management

- **Create new groups with custom names**: Users can create new groups with specific names
- **Add multiple members at once**: Users can add multiple members to a group in one operation
- **Remove members from groups**: Users can remove members from a group if needed
- **View group members**: Users can view the list of members in a group
- **Group-specific chat history**: Users can view the history of messages in a specific group
- **Admin controls for group management**: Group owners can manage the group, including adding or removing members

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