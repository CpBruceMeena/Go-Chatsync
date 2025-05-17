import React, { createContext, useContext, useEffect, useState, useCallback } from 'react';

const WebSocketContext = createContext(null);

export const useWebSocket = () => {
  const context = useContext(WebSocketContext);
  if (!context) {
    throw new Error('useWebSocket must be used within a WebSocketProvider');
  }
  return context;
};

export const WebSocketProvider = ({ children }) => {
  const [socket, setSocket] = useState(null);
  const [isConnected, setIsConnected] = useState(false);
  const [username, setUsername] = useState('');
  const [users, setUsers] = useState([]);
  const [groups, setGroups] = useState({});
  const [messages, setMessages] = useState([]);
  const [selectedChat, setSelectedChat] = useState(null);
  const wsRef = React.useRef(null);

  const connect = useCallback(() => {
    if (!username) {
      console.error('Username is required for WebSocket connection');
      return;
    }

    if (wsRef.current) {
      wsRef.current.close();
    }

    // Clear messages when connecting
    setMessages([]);

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/ws?username=${encodeURIComponent(username)}`;
    const ws = new WebSocket(wsUrl);
    
    ws.onopen = () => {
      setIsConnected(true);
    };

    ws.onclose = () => {
      setIsConnected(false);
      // Attempt to reconnect after 3 seconds
      setTimeout(() => {
        if (username) {
          connect();
        }
      }, 3000);
    };

    ws.onmessage = (event) => {
      const data = JSON.parse(event.data);
      handleMessage(data);
    };

    setSocket(ws);
    wsRef.current = ws;
  }, [username]);

  // Connect when username changes
  useEffect(() => {
    if (username) {
      connect();
    }
  }, [username, connect]);

  const handleMessage = (data) => {
    console.log('WebSocketContext: Received message:', {
      type: data.type,
      from: data.from,
      to: data.to,
      content: data.content,
      timestamp: data.timestamp,
      raw: data,
      selectedChat,
      username
    });
    switch (data.type) {
      case 'user_list': {
        console.log('WebSocketContext: Updating user list:', data.users);
        const userList = Object.keys(data.users || {});
        console.log('WebSocketContext: Processed user list:', userList);
        setUsers(userList);
        break;
      }
      case 'group_list':
        console.log('WebSocketContext: Updating group list:', data.groups);
        setGroups(data.groups || {});
        break;
      case 'private_message':
      case 'group_message':
      case 'system':
        console.log('WebSocketContext: Processing new message:', {
          type: data.type,
          from: data.from,
          to: data.to,
          content: data.content,
          timestamp: data.timestamp,
          selectedChat,
          username,
          isRelevant: selectedChat && (
            (data.type === 'private_message' && 
             ((data.from === username && data.to === selectedChat.id) ||
              (data.from === selectedChat.id && data.to === username))) ||
            (data.type === 'group_message' && data.to === selectedChat.id)
          )
        });
        setMessages(prev => {
          const newMessages = [...prev, data];
          console.log('WebSocketContext: Updated messages state:', {
            previousCount: prev.length,
            newCount: newMessages.length,
            newMessage: data,
            selectedChat,
            username,
            allMessages: newMessages
          });
          return newMessages;
        });
        break;
      default:
        console.log('WebSocketContext: Unknown message type:', data.type);
    }
  };

  const sendMessage = useCallback((message) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      // Ensure the from field is set
      const messageToSend = {
        ...message,
        from: username,  // Ensure from field is set
        timestamp: new Date().toISOString()  // Ensure timestamp is set
      };
      console.log('WebSocketContext: Sending message:', {
        message: messageToSend,
        readyState: wsRef.current.readyState,
        selectedChat,
        username
      });
      wsRef.current.send(JSON.stringify(messageToSend));
      // Add the message to local state immediately
      setMessages(prev => {
        const newMessages = [...prev, messageToSend];
        console.log('WebSocketContext: Added sent message to state:', {
          message: messageToSend,
          previousCount: prev.length,
          newCount: newMessages.length
        });
        return newMessages;
      });
      console.log('WebSocketContext: Message sent successfully');
    } else {
      console.error('WebSocketContext: WebSocket is not connected');
    }
  }, [username, selectedChat]);

  const createGroup = (groupName, members) => {
    if (!wsRef.current) {
      console.error('WebSocket is not connected');
      return;
    }
    const message = {
      type: 'create_group',
      to: groupName,
      content: members.join(','),
      timestamp: new Date().toISOString()
    };
    console.log('Sending group creation message:', message);
    wsRef.current.send(JSON.stringify(message));
  };

  const addGroupMember = (groupId, member) => {
    sendMessage({
      type: 'add_group_member',
      groupId,
      member
    });
  };

  const removeGroupMember = (groupId, member) => {
    sendMessage({
      type: 'remove_group_member',
      groupId,
      member
    });
  };

  useEffect(() => {
    if (!wsRef.current) return;

    wsRef.current.onmessage = (event) => {
      const message = JSON.parse(event.data);
      handleMessage(message);
    };

    return () => {
      if (socket) {
        socket.close();
      }
    };
  }, [socket]);

  const value = {
    isConnected,
    username,
    setUsername,
    users,
    groups,
    messages,
    selectedChat,
    setSelectedChat,
    connect,
    sendMessage,
    createGroup,
    addGroupMember,
    removeGroupMember
  };

  return (
    <WebSocketContext.Provider value={value}>
      {children}
    </WebSocketContext.Provider>
  );
}; 