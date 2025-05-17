import React, { useState, useRef, useEffect, useCallback, useMemo } from 'react';
import { useWebSocket } from '../contexts/WebSocketContext';
import {
  Box,
  Paper,
  Typography,
  TextField,
  Button,
  List,
  ListItem,
  ListItemText,
  ListItemButton,
  Divider,
  IconButton,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Chip
} from '@mui/material';
import {
  Add as AddIcon,
  Send as SendIcon,
  Group as GroupIcon,
  Person as PersonIcon
} from '@mui/icons-material';
import CreateGroupDialog from './CreateGroupDialog';

const Chat = () => {
  const {
    username,
    users,
    groups,
    messages,
    selectedChat,
    setSelectedChat,
    sendMessage,
    createGroup,
    addGroupMember,
    removeGroupMember,
    setMessages
  } = useWebSocket();

  const [newMessage, setNewMessage] = useState('');
  const [isGroupDialogOpen, setIsGroupDialogOpen] = useState(false);
  const [newGroupName, setNewGroupName] = useState('');
  const [selectedMembers, setSelectedMembers] = useState([]);
  const [selectedGroup, setSelectedGroup] = useState(null);
  const messagesEndRef = useRef(null);

  const scrollToBottom = useCallback(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, []);

  // Scroll to bottom when messages change
  useEffect(() => {
    scrollToBottom();
  }, [scrollToBottom]);

  useEffect(() => {
    console.log('Users updated in Chat component:', users);
    console.log('Current selected chat:', selectedChat);
  }, [users, selectedChat]);

  // Add effect to load message history when chat is selected
  useEffect(() => {
    if (selectedChat && selectedChat.type === 'private') {
      console.log('Chat: Loading message history for private chat:', {
        selectedChat,
        username
      });
      // The backend will send the message history when the chat is selected
      // We just need to ensure our filtering logic is correct
    }
  }, [selectedChat, username]);

  useEffect(() => {
    console.log('Groups updated in Chat component:', groups);
    console.log('Current groups state:', groups);
  }, [groups]);

  const handleChatSelect = (type, id) => {
    console.log('Chat: Selecting chat:', { type, id });
    setSelectedChat({ type, id });
    // Clear messages when switching chats - they will be loaded from history
    setMessages([]);
  };

  const handleSendMessage = (content) => {
    if (!content.trim() || !selectedChat) return;

    const message = {
      type: selectedChat.type === 'private' ? 'private_message' : 'group_message',
      to: selectedChat.id,
      content: content.trim()
    };

    sendMessage(message);
    setNewMessage(''); // Clear the input field after sending
  };

  // Filter messages for the selected chat
  const filteredMessages = useMemo(() => {
    if (!selectedChat) return [];
    
    if (selectedChat.type === 'private') {
      return messages.filter(msg => 
        msg.type === 'private_message' && 
        ((msg.from === selectedChat.id && msg.to === username) || 
         (msg.from === username && msg.to === selectedChat.id))
      );
    }
    
    return messages.filter(msg => 
      msg.type === 'group_message' && msg.to === selectedChat.id
    );
  }, [messages, selectedChat, username]);

  const handleCreateGroup = () => {
    if (newGroupName.trim() && selectedMembers.length > 0) {
      console.log('Creating group:', {
        name: newGroupName.trim(),
        members: selectedMembers
      });
      createGroup(newGroupName.trim(), selectedMembers);
      setNewGroupName('');
      setSelectedMembers([]);
      setIsGroupDialogOpen(false);
    }
  };

  // Memoize the message filter function
  const isMessageForCurrentChat = useCallback((message) => {
    if (!selectedChat) return false;
    
    if (selectedChat.type === 'private') {
      // For private messages, show if:
      // 1. Message is from current user to selected user
      // 2. Message is from selected user to current user
      const isRelevant = (message.type === 'private_message' && 
                         ((message.from === username && message.to === selectedChat.id) ||
                          (message.from === selectedChat.id && message.to === username)));
      console.log('Chat: Checking message relevance:', {
        message,
        selectedChat,
        username,
        isRelevant,
        condition1: message.from === username && message.to === selectedChat.id,
        condition2: message.from === selectedChat.id && message.to === username,
        messageType: message.type
      });
      return isRelevant;
    }
    
    // For group messages, show if message is for the selected group
    return selectedChat.type === 'group' && message.type === 'group_message' && message.to === selectedChat.id;
  }, [selectedChat, username]);

  // Add a function to get message statistics
  const getMessageStats = useCallback(() => {
    const stats = {
      sent: {},
      received: {}
    };

    for (const message of messages) {
      if (message.type === 'private_message') {
        // Count sent messages
        if (message.from === username) {
          stats.sent[message.to] = (stats.sent[message.to] || 0) + 1;
        }
        // Count received messages
        if (message.to === username) {
          stats.received[message.from] = (stats.received[message.from] || 0) + 1;
        }
      }
    }

    return stats;
  }, [messages, username]);

  return (
    <Box sx={{ display: 'flex', height: '100vh', p: 2 }}>
      {/* Sidebar */}
      <Paper sx={{ width: 300, mr: 2, display: 'flex', flexDirection: 'column' }}>
        <Box sx={{ p: 2, display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <Typography variant="h6">Users ({users.length})</Typography>
        </Box>
        <Divider />
        <List sx={{ flex: 1, overflow: 'auto' }}>
          {users.map((user) => (
            user !== username && (
              <ListItem key={user} disablePadding>
                <ListItemButton
                  selected={selectedChat?.type === 'private' && selectedChat?.name === user}
                  onClick={() => handleChatSelect('private', user)}
                >
                  <ListItemText primary={user} />
                </ListItemButton>
              </ListItem>
            )
          ))}
        </List>
        <Divider />
        <Box sx={{ p: 2, display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <Typography variant="h6">Groups ({Object.keys(groups).length})</Typography>
          <IconButton 
            onClick={() => setIsGroupDialogOpen(true)}
            sx={{
              color: 'primary.main',
              '&:hover': {
                backgroundColor: 'primary.light',
                color: 'white'
              }
            }}
          >
            <AddIcon />
          </IconButton>
        </Box>
        <Divider />
        <List sx={{ flex: 1, overflow: 'auto' }}>
          {Object.entries(groups).map(([groupName, group]) => (
            <ListItem key={groupName} disablePadding>
              <ListItemButton
                selected={selectedChat?.type === 'group' && selectedChat?.name === groupName}
                onClick={() => handleChatSelect('group', groupName)}
              >
                <ListItemText 
                  primary={groupName}
                  secondary={`${group.members.length} members`}
                />
              </ListItemButton>
            </ListItem>
          ))}
        </List>
      </Paper>

      {/* Chat Area */}
      <Paper sx={{ flex: 1, display: 'flex', flexDirection: 'column' }}>
        <Box sx={{ p: 2, borderBottom: 1, borderColor: 'divider', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <Typography variant="h6">
            {selectedChat ? (selectedChat.type === 'private' ? selectedChat.id : groups[selectedChat.id]?.name) : 'Select a chat'}
          </Typography>
          <Typography variant="subtitle1" color="text.secondary">
            Logged in as: {username}
          </Typography>
        </Box>
        {selectedChat ? (
          <>
            <Box sx={{ flex: 1, overflow: 'auto', p: 2, display: 'flex', flexDirection: 'column', gap: 1 }}>
              {filteredMessages.map((msg) => (
                <Box
                  key={`${msg.from}-${msg.timestamp}-${msg.content}`}
                  sx={{
                    display: 'flex',
                    justifyContent: msg.from === username ? 'flex-end' : 'flex-start',
                    mb: 1
                  }}
                >
                  <Paper
                    elevation={1}
                    sx={{
                      p: 1,
                      maxWidth: '70%',
                      backgroundColor: msg.from === username ? 'primary.light' : 'grey.100',
                      color: msg.from === username ? 'white' : 'text.primary'
                    }}
                  >
                    <Typography variant="body1">{msg.content}</Typography>
                    <Typography variant="caption" sx={{ display: 'block', mt: 0.5, opacity: 0.7 }}>
                      {new Date(msg.timestamp).toLocaleTimeString()}
                    </Typography>
                  </Paper>
                </Box>
              ))}
              <div ref={messagesEndRef} />
            </Box>
            <Box
              component="form"
              onSubmit={(e) => {
                e.preventDefault();
                if (newMessage.trim() && selectedChat) {
                  const message = {
                    type: selectedChat.type === 'private' ? 'private_message' : 'group_message',
                    to: selectedChat.id,
                    content: newMessage.trim()
                  };
                  sendMessage(message);
                  setNewMessage(''); // Clear the input field
                }
              }}
              sx={{ p: 2, borderTop: 1, borderColor: 'divider', display: 'flex', gap: 1 }}
            >
              <TextField
                fullWidth
                value={newMessage}
                onChange={(e) => setNewMessage(e.target.value)}
                placeholder="Type a message..."
                variant="outlined"
                size="small"
              />
              <IconButton 
                type="submit" 
                color="primary"
                disabled={!newMessage.trim()}
              >
                <SendIcon />
              </IconButton>
            </Box>
          </>
        ) : (
          <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: '100%' }}>
            <Typography color="text.secondary">Select a chat to start messaging</Typography>
          </Box>
        )}
      </Paper>

      {/* Create Group Dialog */}
      <CreateGroupDialog
        open={isGroupDialogOpen}
        onClose={() => setIsGroupDialogOpen(false)}
        onCreateGroup={createGroup}
        users={users.filter(user => user !== username)}
      />
    </Box>
  );
};

export default Chat; 