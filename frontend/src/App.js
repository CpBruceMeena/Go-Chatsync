import React from 'react';
import { ThemeProvider, createTheme, CssBaseline } from '@mui/material';
import { WebSocketProvider, useWebSocket } from './contexts/WebSocketContext';
import Login from './components/Login';
import Chat from './components/Chat';

const theme = createTheme({
  palette: {
    primary: {
      main: '#1a73e8',
    },
    background: {
      default: '#f0f2f5',
    },
  },
});

const AppContent = () => {
  const { isConnected } = useWebSocket();
  return isConnected ? <Chat /> : <Login />;
};

function App() {
  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <WebSocketProvider>
        <AppContent />
      </WebSocketProvider>
    </ThemeProvider>
  );
}

export default App;
