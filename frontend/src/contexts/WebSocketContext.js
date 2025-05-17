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

  const handleMessage = useCallback((message) => {
    console.log('WebSocketContext: Received message:', message);
    switch (message.type) {
      case 'user_list':
        setUsers(Object.keys(message.users));
        break;
      case 'group_list': {
        const groupsMap = message.groups.reduce((acc, group) => {
          acc[group.name] = group;
          return acc;
        }, {});
        console.log('WebSocketContext: Updating groups:', groupsMap);
        setGroups(groupsMap);
        break;
      }
      case 'private_message':
      case 'group_message':
        if (message.from !== username) {
          setMessages(prev => [...prev, message]);
        }
        break;
      case 'history':
        console.log('WebSocketContext: Received message history:', message.content);
        setMessages(message.content);
        break;
      case 'system':
        console.log('System message:', message.content);
        break;
      case 'unread_count':
        console.log('WebSocketContext: Received unread counts:', message.content);
        try {
          const counts = JSON.parse(message.content);
          console.log('WebSocketContext: Parsed unread counts:', counts);
          // Dispatch a custom event with the unread counts
          const event = new CustomEvent('unreadCounts', { detail: counts });
          window.dispatchEvent(event);
        } catch (error) {
          console.error('WebSocketContext: Error parsing unread counts:', error);
        }
        break;
      default:
        console.log('Unknown message type:', message.type);
    }
  }, [username]);

  const sendMessage = useCallback((message) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      const messageToSend = {
        ...message,
        from: username,
        timestamp: new Date().toISOString()
      };
      console.log('WebSocketContext: Sending message:', {
        message: messageToSend,
        readyState: wsRef.current.readyState,
        selectedChat,
        username
      });
      
      setMessages(prev => [...prev, messageToSend]);
      
      wsRef.current.send(JSON.stringify(messageToSend));
      console.log('WebSocketContext: Message sent successfully');
    } else {
      console.error('WebSocketContext: WebSocket is not connected');
    }
  }, [username, selectedChat]);

  // Function to request message history
  const requestMessageHistory = useCallback((chatType, chatId) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      const message = {
        type: 'request_history',
        to: chatType,
        content: chatId,
        timestamp: new Date().toISOString()
      };
      console.log('WebSocketContext: Requesting message history:', message);
      wsRef.current.send(JSON.stringify(message));
    }
  }, []);

  // Effect to request message history when chat is selected
  useEffect(() => {
    if (selectedChat) {
      console.log('Chat selected, requesting history:', selectedChat);
      requestMessageHistory(selectedChat.type, selectedChat.id);
    }
  }, [selectedChat, requestMessageHistory]);

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
  }, [socket, handleMessage]);

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
    removeGroupMember,
    ws: wsRef.current
  };

  return (
    <WebSocketContext.Provider value={value}>
      {children}
    </WebSocketContext.Provider>
  );
}; 