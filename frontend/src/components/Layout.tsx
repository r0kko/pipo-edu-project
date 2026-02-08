import { PropsWithChildren } from 'react';
import { AppBar, Box, Button, Container, Toolbar, Typography } from '@mui/material';
import { authStore } from '../store/auth';
import { useNavigate } from 'react-router-dom';
import { roleLabel } from '../utils/roles';

interface LayoutProps {
  title: string;
}

export function Layout({ title, children }: PropsWithChildren<LayoutProps>) {
  const navigate = useNavigate();
  const user = authStore.getUser();

  const handleLogout = () => {
    authStore.clear();
    navigate('/login');
  };

  return (
    <Box sx={{ minHeight: '100vh', background: 'linear-gradient(180deg, #F6F4EE 0%, #EDE7DB 100%)' }}>
      <AppBar position="static" sx={{ background: '#1A3D7C' }}>
        <Toolbar>
          <Typography variant="h6" sx={{ flexGrow: 1, fontWeight: 700 }}>
            PIPO · Пропуска
          </Typography>
          <Typography variant="body2" sx={{ mr: 2, opacity: 0.85 }}>
            {user?.full_name} ({roleLabel(user?.role)})
            {user?.plot_number ? ` · участок ${user.plot_number}` : ''}
          </Typography>
          <Button color="inherit" onClick={handleLogout}>
            Выйти
          </Button>
        </Toolbar>
      </AppBar>
      <Container sx={{ py: 4 }}>
        <Typography variant="h4" sx={{ fontWeight: 700, mb: 3 }}>
          {title}
        </Typography>
        {children}
      </Container>
    </Box>
  );
}
