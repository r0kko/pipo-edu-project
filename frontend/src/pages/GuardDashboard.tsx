import { useMutation, useQuery } from '@tanstack/react-query';
import { Box, Button, Card, CardContent, Stack, TextField, Typography } from '@mui/material';
import { Layout } from '../components/Layout';
import api from '../api/client';
import { Pass } from '../api/types';
import { useState } from 'react';

export default function GuardDashboard() {
  const [plate, setPlate] = useState('');
  const [search, setSearch] = useState('');

  const passesQuery = useQuery({
    queryKey: ['passes-search', search],
    queryFn: async () => (await api.get<Pass[]>('/passes/search', { params: { plate: search } })).data,
    enabled: !!search
  });

  const entryMutation = useMutation({
    mutationFn: (id: string) => api.post(`/passes/${id}/entry`, {})
  });

  const exitMutation = useMutation({
    mutationFn: (id: string) => api.post(`/passes/${id}/exit`, {})
  });

  return (
    <Layout title="Пульт охраны">
      <Card>
        <CardContent>
          <Typography variant="h6" sx={{ mb: 1, fontWeight: 700 }}>
            Поиск пропуска
          </Typography>
          <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
            Введите номер авто, чтобы найти активный пропуск
          </Typography>
          <Stack direction={{ xs: 'column', sm: 'row' }} spacing={2}>
            <TextField
              label="Номер авто"
              value={plate}
              onChange={(e) => setPlate(e.target.value)}
              fullWidth
            />
            <Button variant="contained" onClick={() => setSearch(plate)}>Найти</Button>
          </Stack>
        </CardContent>
      </Card>
      <Box sx={{ mt: 3 }}>
        {passesQuery.data?.map((pass) => (
          <Card key={pass.id} sx={{ mb: 2 }}>
            <CardContent sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', flexWrap: 'wrap', gap: 2 }}>
              <Box>
                <Typography variant="h6" sx={{ fontWeight: 700 }}>
                  {pass.plate_number}
                </Typography>
                <Typography variant="body2" color="text.secondary">
                  Статус: {pass.status} · Владелец: {pass.owner_full_name || pass.owner_user_id}
                  {pass.owner_plot_number ? ` · участок ${pass.owner_plot_number}` : ''}
                </Typography>
              </Box>
              <Stack direction="row" spacing={1}>
                <Button variant="outlined" onClick={() => entryMutation.mutate(pass.id)}>
                  Въезд
                </Button>
                <Button variant="contained" onClick={() => exitMutation.mutate(pass.id)}>
                  Выезд
                </Button>
              </Stack>
            </CardContent>
          </Card>
        ))}
      </Box>
    </Layout>
  );
}
