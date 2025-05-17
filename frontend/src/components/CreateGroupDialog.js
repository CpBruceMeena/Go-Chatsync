import React, { useState } from 'react';
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  Button,
  List,
  ListItem,
  ListItemText,
  ListItemButton,
  Checkbox,
  Box,
  Typography,
  Paper
} from '@mui/material';

const CreateGroupDialog = ({ open, onClose, onCreateGroup, users }) => {
  const [groupName, setGroupName] = useState('');
  const [selectedMembers, setSelectedMembers] = useState([]);

  const handleSubmit = (e) => {
    e.preventDefault();
    if (groupName.trim() && selectedMembers.length > 0) {
      onCreateGroup(groupName.trim(), selectedMembers);
      setGroupName('');
      setSelectedMembers([]);
      onClose();
    }
  };

  const handleMemberToggle = (username) => {
    setSelectedMembers(prev => {
      if (prev.includes(username)) {
        return prev.filter(member => member !== username);
      }
      return [...prev, username];
    });
  };

  return (
    <Dialog 
      open={open} 
      onClose={onClose}
      maxWidth="sm"
      fullWidth
      PaperProps={{
        sx: {
          minHeight: '60vh',
          maxHeight: '80vh',
        }
      }}
    >
      <DialogTitle sx={{ 
        backgroundColor: 'primary.main',
        color: 'white',
        '& .MuiTypography-root': {
          fontWeight: 'bold'
        }
      }}>
        Create New Group
      </DialogTitle>
      <DialogContent sx={{ p: 3 }}>
        <Box component="form" onSubmit={handleSubmit}>
          <TextField
            fullWidth
            label="Group Name"
            value={groupName}
            onChange={(e) => setGroupName(e.target.value)}
            margin="normal"
            required
            sx={{ mb: 2 }}
          />
          <Typography variant="subtitle1" sx={{ mb: 1, color: 'text.secondary' }}>
            Select Members
          </Typography>
          <Paper 
            elevation={0} 
            sx={{ 
              maxHeight: '40vh',
              overflow: 'auto',
              border: '1px solid',
              borderColor: 'divider',
              borderRadius: 1
            }}
          >
            <List>
              {users.map((user) => (
                <ListItem key={user} disablePadding>
                  <ListItemButton
                    onClick={() => handleMemberToggle(user)}
                    dense
                    sx={{
                      '&:hover': {
                        backgroundColor: 'primary.light',
                        '& .MuiTypography-root': {
                          color: 'white'
                        }
                      }
                    }}
                  >
                    <Checkbox
                      edge="start"
                      checked={selectedMembers.includes(user)}
                      tabIndex={-1}
                      disableRipple
                      sx={{
                        color: 'primary.main',
                        '&.Mui-checked': {
                          color: 'primary.main',
                        },
                      }}
                    />
                    <ListItemText 
                      primary={user}
                      sx={{
                        '& .MuiTypography-root': {
                          color: 'text.primary'
                        }
                      }}
                    />
                  </ListItemButton>
                </ListItem>
              ))}
            </List>
          </Paper>
        </Box>
      </DialogContent>
      <DialogActions sx={{ p: 2, backgroundColor: 'background.paper' }}>
        <Button 
          onClick={onClose}
          sx={{ 
            color: 'text.secondary',
            '&:hover': {
              backgroundColor: 'primary.light',
              color: 'white'
            }
          }}
        >
          Cancel
        </Button>
        <Button
          onClick={handleSubmit}
          variant="contained"
          disabled={!groupName.trim() || selectedMembers.length === 0}
          sx={{
            backgroundColor: 'primary.main',
            '&:hover': {
              backgroundColor: 'primary.dark',
            },
            '&.Mui-disabled': {
              backgroundColor: 'grey.300',
              color: 'grey.500',
            },
          }}
        >
          Create Group
        </Button>
      </DialogActions>
    </Dialog>
  );
};

export default CreateGroupDialog; 