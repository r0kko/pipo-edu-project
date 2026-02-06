import { createTheme } from '@mui/material/styles';

export const theme = createTheme({
  typography: {
    fontFamily: 'Manrope, system-ui, sans-serif'
  },
  palette: {
    mode: 'light',
    primary: {
      main: '#1A3D7C'
    },
    secondary: {
      main: '#00B894'
    },
    background: {
      default: '#F6F4EE',
      paper: '#FFFFFF'
    }
  },
  shape: {
    borderRadius: 14
  },
  components: {
    MuiButton: {
      styleOverrides: {
        root: {
          textTransform: 'none',
          fontWeight: 600
        }
      }
    }
  }
});
