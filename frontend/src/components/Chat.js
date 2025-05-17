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

const Chat = () => {
  const {
    username,
    users,
    groups,
    messages,
    selectedChat,
    setSelectedChat,
    sendMessage,
    createGroup
  } = useWebSocket();

  const [newMessage, setNewMessage] = useState('');
  const [isGroupDialogOpen, setIsGroupDialogOpen] = useState(false);
  const [newGroupName, setNewGroupName] = useState('');
  const [selectedMembers, setSelectedMembers] = useState([]);
  const messagesEndRef = useRef(null);

  const scrollToBottom = useCallback(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, []);

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

  const handleSendMessage = (e) => {
    e.preventDefault();
    if (newMessage.trim() && selectedChat) {
      const message = {
        type: selectedChat.type === 'private' ? 'private_message' : 'group_message',
        from: username,
        content: newMessage.trim(),
        to: selectedChat.id,
        timestamp: new Date().toISOString()
      };
      console.log('Chat: Sending message:', {
        message,
        selectedChat,
        username
      });
      sendMessage(message);
      setNewMessage('');
    }
  };

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

  // Memoize filtered messages
  const filteredMessages = useMemo(() => {
    const filtered = messages.filter(isMessageForCurrentChat);
    console.log('Chat: Filtered messages:', {
      allMessages: messages,
      filtered,
      selectedChat,
      username,
      messageCount: messages.length,
      filteredCount: filtered.length,
      filterDetails: messages.map(msg => ({
        message: msg,
        isRelevant: isMessageForCurrentChat(msg),
        reason: !selectedChat ? 'No chat selected' :
                selectedChat.type === 'private' ? 
                  (msg.type === 'private_message' ? 
                    ((msg.from === username && msg.to === selectedChat.id) ? 'Sent to selected user' :
                     (msg.from === selectedChat.id && msg.to === username) ? 'Received from selected user' :
                     'Not relevant to selected chat') :
                   'Not a private message') :
                  (msg.type === 'group_message' && msg.to === selectedChat.id) ? 'Group message' :
                  'Not a group message'
      }))
    });
    return filtered;
  }, [messages, isMessageForCurrentChat, selectedChat, username]);

  return (
    <Box sx={{ display: 'flex', height: '100vh', p: 2 }}>
      {/* Sidebar */}
      <Paper sx={{ width: 300, mr: 2, display: 'flex', flexDirection: 'column' }}>
        <Box sx={{ p: 2, display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <Typography variant="h6">Users ({users.length})</Typography>
          <IconButton onClick={() => setIsGroupDialogOpen(true)}>
            <AddIcon />
          </IconButton>
        </Box>
        <Divider />
        <List sx={{ flex: 1, overflow: 'auto' }}>
          {Array.isArray(users) && users.map((user) => (
            user !== username && (
              <ListItem key={user} disablePadding>
                <ListItemButton
                  selected={selectedChat?.type === 'private' && selectedChat?.id === user}
                  onClick={() => {
                    console.log('Selecting chat with user:', user);
                    setSelectedChat({ type: 'private', id: user, username: user });
                  }}
                >
                  <PersonIcon sx={{ mr: 1 }} />
                  <ListItemText primary={user} />
                </ListItemButton>
              </ListItem>
            )
          ))}
        </List>
        <Divider />
        <Box sx={{ p: 2 }}>
          <Typography variant="h6">Groups</Typography>
        </Box>
        <List sx={{ flex: 1, overflow: 'auto' }}>
          {Object.entries(groups)
            .filter(([_, group]) => group.members.includes(username))
            .map(([id, group]) => (
            <ListItem key={id} disablePadding>
              <ListItemButton
                selected={selectedChat?.type === 'group' && selectedChat?.id === id}
                onClick={() => setSelectedChat({ type: 'group', id, name: group.name })}
              >
                <GroupIcon sx={{ mr: 1 }} />
                <ListItemText 
                  primary={group.name}
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
            <Box sx={{ flexGrow: 1, overflow: 'auto', p: 2 }}>
              {filteredMessages.map((message) => (
                <Box
                  key={`${message.from}-${message.timestamp}`}
                  sx={{
                    display: 'flex',
                    justifyContent: message.from === username ? 'flex-end' : 'flex-start',
                    mb: 2
                  }}
                >
                  <Paper
                    sx={{
                      p: 2,
                      backgroundColor: message.from === username ? 'primary.light' : 'grey.100',
                      maxWidth: '70%'
                    }}
                  >
                    <Typography variant="body1">{message.content}</Typography>
                    <Typography variant="caption" color="textSecondary">
                      {new Date(message.timestamp).toLocaleTimeString()}
                    </Typography>
                  </Paper>
                </Box>
              ))}
              <div ref={messagesEndRef} />
            </Box>
            <Box
              component="form"
              onSubmit={handleSendMessage}
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
              <IconButton type="submit" color="primary">
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
      <Dialog open={isGroupDialogOpen} onClose={() => setIsGroupDialogOpen(false)}>
        <DialogTitle>Create New Group</DialogTitle>
        <DialogContent>
          <TextField
            autoFocus
            margin="dense"
            label="Group Name"
            fullWidth
            value={newGroupName}
            onChange={(e) => setNewGroupName(e.target.value)}
          />
          <FormControl fullWidth sx={{ mt: 2 }}>
            <InputLabel>Members</InputLabel>
            <Select
              multiple
              value={selectedMembers}
              onChange={(e) => setSelectedMembers(e.target.value)}
              renderValue={(selected) => (
                <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 0.5 }}>
                  {selected.map((value) => (
                    <Chip key={value} label={value} />
                  ))}
                </Box>
              )}
            >
              {users.map((user) => (
                user !== username && (
                  <MenuItem key={user} value={user}>
                    {user}
                  </MenuItem>
                )
              ))}
            </Select>
          </FormControl>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setIsGroupDialogOpen(false)}>Cancel</Button>
          <Button
            onClick={handleCreateGroup}
            disabled={!newGroupName.trim() || selectedMembers.length === 0}
          >
            Create
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
};

export default Chat; 