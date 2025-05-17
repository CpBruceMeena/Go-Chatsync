# Go-Chatsync

A real-time chat application built with Go and React, featuring private messaging and group chat functionality.

## Features

- Real-time messaging using WebSocket
- Private messaging between users
- Group chat support
- User presence tracking
- Modern Material-UI interface
- Responsive design
- Message history (non-persistent)

## Tech Stack

### Backend
- Go 1.21
- Gorilla WebSocket
- In-memory message storage

### Frontend
- React
- Material-UI
- WebSocket client
- Context-based state management

## Prerequisites

- Go 1.21 or higher
- Node.js and npm
- Modern web browser

## Installation

1. Clone the repository:
```bash
git clone https://github.com/yourusername/Go-Chatsync.git
cd Go-Chatsync
```

2. Run the application:
```bash
./run.sh
```

The script will:
- Install Go dependencies
- Install frontend dependencies
- Build the React app
- Build the Go binary
- Start the server

## Usage

1. Open your browser and navigate to http://localhost:8080
2. Enter a username to start chatting
3. Features available:
   - Private messaging: Click on a user to start a private chat
   - Group chat: Create a group and add members
   - Real-time updates: See messages instantly
   - User presence: See who's online

## Development

### Project Structure
```
.
├── frontend/           # React application
│   ├── src/           # Source code
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

### Running in Development Mode

1. Start the backend:
```bash
cd backend
go run main.go
```

2. Start the frontend development server:
```bash
cd frontend
npm start
```

## Features in Detail

### Real-time Messaging
- Instant message delivery
- Message history for current session
- Support for private and group messages

### User Interface
- Clean, modern design with Material-UI
- Responsive layout
- User presence indicators
- Message timestamps
- Clear visual hierarchy

### Group Management
- Create new groups
- Add/remove members
- Group chat functionality

### Technical Features
- WebSocket communication
- Real-time updates
- Connection management
- Error handling
- State management

## Contributing

1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details. 