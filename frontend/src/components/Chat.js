import React, { useState, useRef, useEffect, useCallback, useMemo } from 'react';
import { useWebSocket } from '../contexts/WebSocketContext';
import {
  Box,
  Paper,
  Typography,
  TextField,
  List,
  ListItem,
  ListItemText,
  ListItemButton,
  Divider,
  IconButton,
  Dialog,
  DialogContent,
  DialogActions,
  Badge
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
    setMessages,
    ws
  } = useWebSocket();

  const [newMessage, setNewMessage] = useState('');
  const [isGroupDialogOpen, setIsGroupDialogOpen] = useState(false);
  const [newGroupName, setNewGroupName] = useState('');
  const [selectedMembers, setSelectedMembers] = useState([]);
  const [selectedGroup, setSelectedGroup] = useState(null);
  const [unreadMessages, setUnreadMessages] = useState({});
  const [lastSeenTimestamps, setLastSeenTimestamps] = useState({});
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

  // Function to update last seen timestamp when chat is selected
  const updateLastSeen = useCallback((chatId) => {
    const currentTime = new Date().toISOString();
    setLastSeenTimestamps(prev => ({
      ...prev,
      [chatId]: currentTime
    }));
  }, []);

  // Update unread messages when receiving unread count message
  useEffect(() => {
    const handleUnreadCount = (event) => {
      console.log('Chat: Received unread counts event:', event.detail);
      setUnreadMessages(prev => {
        const newCounts = { ...prev, ...event.detail };
        console.log('Chat: Updated unread counts:', newCounts);
        return newCounts;
      });
    };

    console.log('Chat: Adding unread counts event listener');
    window.addEventListener('unreadCounts', handleUnreadCount);

    return () => {
      console.log('Chat: Removing unread counts event listener');
      window.removeEventListener('unreadCounts', handleUnreadCount);
    };
  }, []);

  const handleChatSelect = (type, id) => {
    console.log('Chat: Selecting chat:', { type, id });
    setSelectedChat({ type, id });
    
    // Clear unread count for selected chat
    setUnreadMessages(prev => {
      const newCounts = { ...prev, [id]: 0 };
      console.log('Chat: Cleared unread count for chat:', id, 'New counts:', newCounts);
      return newCounts;
    });

    // Send last seen update to backend with current timestamp
    const currentTime = new Date().toISOString();
    const message = {
      type: 'update_last_seen',
      to: id,
      content: currentTime,
      timestamp: currentTime
    };
    console.log('Chat: Sending last seen update:', message);
    sendMessage(message);

    // Update local last seen timestamp
    setLastSeenTimestamps(prev => ({
      ...prev,
      [id]: currentTime
    }));
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
    <Box sx={{ display: 'flex', height: '100vh', p: 2, bgcolor: 'background.default' }}>
      {/* Sidebar */}
      <Paper 
        sx={{ 
          width: 300, 
          mr: 2, 
          display: 'flex', 
          flexDirection: 'column',
          bgcolor: 'background.paper',
          boxShadow: 3
        }}
      >
        <Box sx={{ p: 2, display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <Typography variant="h6">Users ({users.length})</Typography>
        </Box>
        <Divider />
        <List sx={{ flex: 1, overflow: 'auto' }}>
          {users.map((user) => (
            user !== username && (
              <ListItem key={user} disablePadding>
                <ListItemButton
                  selected={selectedChat?.type === 'private' && selectedChat?.id === user}
                  onClick={() => handleChatSelect('private', user)}
                  sx={{
                    '&.Mui-selected': {
                      bgcolor: 'primary.light',
                      color: 'white',
                      '&:hover': {
                        bgcolor: 'primary.main',
                      }
                    }
                  }}
                >
                  <Badge
                    badgeContent={unreadMessages[user] || 0}
                    color="error"
                    max={999999}
                    sx={{
                      '& .MuiBadge-badge': {
                        right: -3,
                        top: 3,
                      }
                    }}
                  >
                    <PersonIcon sx={{ mr: 1, color: 'primary.main' }} />
                  </Badge>
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
                selected={selectedChat?.type === 'group' && selectedChat?.id === groupName}
                onClick={() => handleChatSelect('group', groupName)}
                sx={{
                  '&.Mui-selected': {
                    bgcolor: 'primary.light',
                    color: 'white',
                    '&:hover': {
                      bgcolor: 'primary.main',
                    }
                  }
                }}
              >
                <Badge
                  badgeContent={unreadMessages[groupName] || 0}
                  color="error"
                  max={999999}
                  sx={{
                    '& .MuiBadge-badge': {
                      right: -3,
                      top: 3,
                    }
                  }}
                >
                  <GroupIcon sx={{ mr: 1, color: 'primary.main' }} />
                </Badge>
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
      <Paper 
        sx={{ 
          flex: 1, 
          display: 'flex', 
          flexDirection: 'column',
          bgcolor: 'background.paper',
          boxShadow: 3
        }}
      >
        <Box 
          sx={{ 
            p: 2, 
            borderBottom: 1, 
            borderColor: 'divider', 
            display: 'flex', 
            justifyContent: 'space-between', 
            alignItems: 'center',
            bgcolor: 'primary.main',
            color: 'white'
          }}
        >
          <Typography variant="h6">
            {selectedChat ? (
              selectedChat.type === 'private' ? (
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                  <PersonIcon sx={{ fontSize: 20 }} />
                  {selectedChat.id}
                </Box>
              ) : (
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                  <GroupIcon sx={{ fontSize: 20 }} />
                  {groups[selectedChat.id]?.name}
                </Box>
              )
            ) : (
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                <GroupIcon sx={{ fontSize: 20 }} />
                Welcome to ChatSync
              </Box>
            )}
          </Typography>
          <Box 
            sx={{ 
              display: 'flex', 
              alignItems: 'center', 
              gap: 1,
              bgcolor: 'primary.dark',
              px: 2,
              py: 0.5,
              borderRadius: 2
            }}
          >
            <PersonIcon sx={{ color: 'white', fontSize: 20 }} />
            <Typography variant="subtitle1" sx={{ color: 'white' }}>
              {username}
            </Typography>
          </Box>
        </Box>
        {selectedChat ? (
          <>
            <Box 
              sx={{ 
                flex: 1, 
                overflow: 'auto', 
                p: 2, 
                display: 'flex', 
                flexDirection: 'column', 
                gap: 1,
                bgcolor: 'background.default'
              }}
            >
              {filteredMessages.map((msg) => (
                <Box
                  key={`${msg.from}-${msg.timestamp}-${msg.content}`}
                  sx={{
                    display: 'flex',
                    justifyContent: msg.from === username ? 'flex-end' : 'flex-start',
                    mb: 1
                  }}
                >
                  <Box
                    sx={{
                      maxWidth: '50%',
                      p: 1.5,
                      borderRadius: 2,
                      bgcolor: msg.from === username ? 'primary.main' : 'secondary.main',
                      color: msg.from === username ? 'white' : 'text.primary',
                      boxShadow: 1,
                      position: 'relative'
                    }}
                  >
                    {msg.from !== username && selectedChat?.type === 'group' && (
                      <Typography 
                        variant="caption" 
                        sx={{ 
                          display: 'block', 
                          mb: 0.5,
                          color: msg.from === username ? 'white' : 'text.secondary',
                          fontWeight: 'bold',
                          fontSize: '0.7rem'
                        }}
                      >
                        {msg.from}
                      </Typography>
                    )}
                    <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-end', gap: 0.5 }}>
                      <Typography 
                        variant="body2" 
                        sx={{ 
                          fontSize: '0.9rem',
                          lineHeight: 1.4,
                          pr: 0.5
                        }}
                      >
                        {msg.content}
                      </Typography>
                      <Typography 
                        variant="caption" 
                        sx={{ 
                          color: msg.from === username ? 'white' : 'text.secondary',
                          opacity: 0.7,
                          fontSize: '0.65rem',
                          whiteSpace: 'nowrap',
                          alignSelf: 'flex-end'
                        }}
                      >
                        {new Date(msg.timestamp).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                      </Typography>
                    </Box>
                  </Box>
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
                  setNewMessage('');
                }
              }}
              sx={{ 
                p: 2, 
                borderTop: 1, 
                borderColor: 'divider', 
                display: 'flex', 
                gap: 1,
                bgcolor: 'background.paper'
              }}
            >
              <TextField
                fullWidth
                value={newMessage}
                onChange={(e) => setNewMessage(e.target.value)}
                placeholder="Type a message..."
                variant="outlined"
                size="small"
                sx={{
                  '& .MuiOutlinedInput-root': {
                    '&:hover fieldset': {
                      borderColor: 'primary.main',
                    },
                    '&.Mui-focused fieldset': {
                      borderColor: 'primary.main',
                    }
                  },
                }}
              />
              <IconButton 
                type="submit" 
                color="primary"
                disabled={!newMessage.trim()}
                sx={{
                  bgcolor: 'primary.main',
                  color: 'white',
                  '&:hover': {
                    bgcolor: 'primary.dark',
                  },
                  '&.Mui-disabled': {
                    bgcolor: 'grey.300',
                    color: 'grey.500'
                  }
                }}
              >
                <SendIcon />
              </IconButton>
            </Box>
          </>
        ) : (
          <Box 
            sx={{ 
              display: 'flex', 
              flexDirection: 'column',
              alignItems: 'center', 
              justifyContent: 'center', 
              height: '100%',
              gap: 3,
              bgcolor: 'background.default',
              p: 4
            }}
          >
            <GroupIcon sx={{ fontSize: 80, color: 'primary.main', opacity: 0.5 }} />
            <Typography 
              variant="h5" 
              color="text.secondary"
              sx={{ 
                textAlign: 'center',
                maxWidth: '80%',
                lineHeight: 1.5,
                fontWeight: 500
              }}
            >
              Start a conversation by selecting a user or group from the sidebar
            </Typography>
            <Typography 
              variant="body1" 
              color="text.secondary"
              sx={{ 
                textAlign: 'center',
                maxWidth: '60%',
                opacity: 0.7
              }}
            >
              You can create new groups and invite users to collaborate
            </Typography>
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