import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Button, Card, CardContent, Stack, TextField, Typography } from '@mui/material';
import { useForm } from 'react-hook-form';
import api from '../api/client';
import { authStore } from '../store/auth';
import { TokenResponse } from '../api/types';

interface FormValues {
  email: string;
  password: string;
}

export default function LoginPage() {
  const { register, handleSubmit } = useForm<FormValues>();
  const [error, setError] = useState<string | null>(null);
  const navigate = useNavigate();

  const onSubmit = async (values: FormValues) => {
    setError(null);
    try {
      const { data } = await api.post<TokenResponse>('/auth/login', values);
      authStore.setSession(data.access_token, data.refresh_token, data.user);
      if (data.user.role === 'admin') {
        navigate('/admin');
      } else if (data.user.role === 'guard') {
        navigate('/guard');
      } else {
        navigate('/resident');
      }
    } catch (err) {
      setError('Неверный логин или пароль');
    }
  };

  return (
    <Stack
      direction="column"
      alignItems="center"
      justifyContent="center"
      sx={{ minHeight: '100vh', background: 'linear-gradient(140deg, #1A3D7C 0%, #284F9C 50%, #2F6D7E 100%)', p: 2 }}
    >
      <Card sx={{ maxWidth: 420, width: '100%', borderRadius: 4, boxShadow: '0 24px 60px rgba(0,0,0,0.25)' }}>
        <CardContent sx={{ p: 4 }}>
          <Typography variant="h5" sx={{ fontWeight: 700, mb: 1 }}>
            Система пропусков
          </Typography>
          <Typography variant="body2" sx={{ mb: 3, color: 'text.secondary' }}>
            Войдите, чтобы управлять пропусками и заявками
          </Typography>
          <Stack spacing={2} component="form" onSubmit={handleSubmit(onSubmit)}>
            <TextField label="Email" type="email" required {...register('email')} />
            <TextField label="Пароль" type="password" required {...register('password')} />
            {error && (
              <Typography variant="body2" color="error">
                {error}
              </Typography>
            )}
            <Button variant="contained" size="large" type="submit">
              Войти
            </Button>
          </Stack>
        </CardContent>
      </Card>
    </Stack>
  );
}
