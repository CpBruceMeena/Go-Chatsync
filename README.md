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