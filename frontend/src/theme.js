import { createTheme } from '@mui/material/styles';

const theme = createTheme({
  palette: {
    primary: {
      main: '#219EBC', // blue-green
      light: '#8ECAE6', // sky-blue
      dark: '#023047', // prussian-blue
    },
    secondary: {
      main: '#FFB703', // selective-yellow
      dark: '#FB8500', // ut-orange
    },
    background: {
      default: '#8ECAE6', // sky-blue
      paper: '#FFFFFF',
    },
    text: {
      primary: '#023047', // prussian-blue
      secondary: '#219EBC', // blue-green
    },
  },
  components: {
    MuiButton: {
      styleOverrides: {
        root: {
          borderRadius: 8,
          textTransform: 'none',
        },
        contained: {
          '&:hover': {
            backgroundColor: '#023047', // prussian-blue
          },
        },
      },
    },
    MuiPaper: {
      styleOverrides: {
        root: {
          borderRadius: 8,
        },
      },
    },
    MuiTextField: {
      styleOverrides: {
        root: {
          '& .MuiOutlinedInput-root': {
            borderRadius: 8,
            '&:hover fieldset': {
              borderColor: '#219EBC', // blue-green
            },
          },
        },
      },
    },
  },
});

export default theme; 