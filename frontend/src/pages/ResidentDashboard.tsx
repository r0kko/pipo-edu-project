import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Box, Button, Card, CardContent, Divider, Grid, Stack, TextField, Typography } from '@mui/material';
import { Layout } from '../components/Layout';
import api from '../api/client';
import { GuestRequest, Pass } from '../api/types';
import { authStore } from '../store/auth';

export default function ResidentDashboard() {
  const qc = useQueryClient();
  const user = authStore.getUser();

  const passesQuery = useQuery({
    queryKey: ['passes'],
    queryFn: async () => (await api.get<Pass[]>('/passes')).data
  });

  const guestsQuery = useQuery({
    queryKey: ['guest'],
    queryFn: async () => (await api.get<GuestRequest[]>('/guest-requests')).data
  });

  const createPass = useMutation({
    mutationFn: (payload: { plate_number: string; vehicle_brand?: string; vehicle_color?: string }) =>
      api.post('/passes', payload),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['passes'] })
  });

  const createGuest = useMutation({
    mutationFn: (payload: { guest_full_name: string; plate_number: string; valid_from: string; valid_to: string }) =>
      api.post('/guest-requests', payload),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['guest'] })
  });

  return (
    <Layout title="Мои пропуска">
      {user?.plot_number && (
        <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
          Участок: {user.plot_number}
        </Typography>
      )}
      <Grid container spacing={3}>
        <Grid item xs={12} md={6}>
          <Card>
            <CardContent>
              <Typography variant="h6" sx={{ mb: 1, fontWeight: 700 }}>
                Личные пропуска
              </Typography>
              <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                Управляйте собственными авто
              </Typography>
              <Stack spacing={1.5} component="form" onSubmit={(e) => {
                e.preventDefault();
                const form = e.currentTarget as HTMLFormElement;
                const formData = new FormData(form);
                createPass.mutate({
                  plate_number: String(formData.get('plate_number')),
                  vehicle_brand: String(formData.get('vehicle_brand') || ''),
                  vehicle_color: String(formData.get('vehicle_color') || '')
                });
                form.reset();
              }}>
                <TextField name="plate_number" label="Номер авто" size="small" required />
                <TextField name="vehicle_brand" label="Марка" size="small" />
                <TextField name="vehicle_color" label="Цвет" size="small" />
                <Button variant="contained" type="submit">Добавить</Button>
              </Stack>
              <Divider sx={{ my: 2 }} />
              <Box sx={{ maxHeight: 240, overflow: 'auto' }}>
                {passesQuery.data?.map((pass) => (
                  <Box key={pass.id} sx={{ mb: 1 }}>
                    <Typography variant="body2" sx={{ fontWeight: 600 }}>
                      {pass.plate_number}
                    </Typography>
                    <Typography variant="caption" color="text.secondary">
                      {pass.status}
                    </Typography>
                  </Box>
                ))}
              </Box>
            </CardContent>
          </Card>
        </Grid>
        <Grid item xs={12} md={6}>
          <Card>
            <CardContent>
              <Typography variant="h6" sx={{ mb: 1, fontWeight: 700 }}>
                Гостевые заявки
              </Typography>
              <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                Заявки для гостей с ограничением по времени
              </Typography>
              <Stack spacing={1.5} component="form" onSubmit={(e) => {
                e.preventDefault();
                const form = e.currentTarget as HTMLFormElement;
                const formData = new FormData(form);
                const validFromRaw = String(formData.get('valid_from'));
                const validToRaw = String(formData.get('valid_to'));
                createGuest.mutate({
                  guest_full_name: String(formData.get('guest_full_name')),
                  plate_number: String(formData.get('plate_number')),
                  valid_from: new Date(validFromRaw).toISOString(),
                  valid_to: new Date(validToRaw).toISOString()
                });
                form.reset();
              }}>
                <TextField name="guest_full_name" label="ФИО гостя" size="small" required />
                <TextField name="plate_number" label="Номер авто" size="small" required />
                <TextField name="valid_from" label="С" size="small" type="datetime-local" InputLabelProps={{ shrink: true }} required />
                <TextField name="valid_to" label="По" size="small" type="datetime-local" InputLabelProps={{ shrink: true }} required />
                <Button variant="contained" type="submit">Создать</Button>
              </Stack>
              <Divider sx={{ my: 2 }} />
              <Box sx={{ maxHeight: 240, overflow: 'auto' }}>
                {guestsQuery.data?.map((guest) => (
                  <Box key={guest.id} sx={{ mb: 1 }}>
                    <Typography variant="body2" sx={{ fontWeight: 600 }}>
                      {guest.guest_full_name}
                    </Typography>
                    <Typography variant="caption" color="text.secondary">
                      {guest.status} · {guest.valid_from}
                    </Typography>
                  </Box>
                ))}
              </Box>
            </CardContent>
          </Card>
        </Grid>
      </Grid>
    </Layout>
  );
}
