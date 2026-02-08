import { useMemo, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  Box,
  Card,
  CardContent,
  Chip,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  IconButton,
  InputAdornment,
  MenuItem,
  Stack,
  Switch,
  TextField,
  Tooltip,
  Typography
} from '@mui/material';
import ArrowBack from '@mui/icons-material/ArrowBack';
import Add from '@mui/icons-material/Add';
import Edit from '@mui/icons-material/Edit';
import Delete from '@mui/icons-material/Delete';
import Restore from '@mui/icons-material/Restore';
import Check from '@mui/icons-material/Check';
import Close from '@mui/icons-material/Close';
import Search from '@mui/icons-material/Search';
import { useNavigate } from 'react-router-dom';
import { useForm } from 'react-hook-form';
import { z } from 'zod';
import { zodResolver } from '@hookform/resolvers/zod';
import { Layout } from '../components/Layout';
import api from '../api/client';
import { GuestRequest, User } from '../api/types';

const plateRegex = /^[ABEKMHOPCTYXАВЕКМНОРСТУХ]\d{3}[ABEKMHOPCTYXАВЕКМНОРСТУХ]{2}\d{2,3}$/i;
const guestStatuses = ['pending', 'approved', 'rejected'] as const;

const createGuestSchema = z
  .object({
    resident_user_id: z.string().uuid('Выберите жителя'),
    guest_full_name: z.string().min(2, 'Минимум 2 символа'),
    plate_number: z.string().regex(plateRegex, 'Неверный формат номера'),
    valid_from: z.string().min(1, 'Укажите дату'),
    valid_to: z.string().min(1, 'Укажите дату'),
    status: z.enum(guestStatuses)
  })
  .refine((data) => new Date(data.valid_to) >= new Date(data.valid_from), {
    path: ['valid_to'],
    message: 'Дата окончания раньше даты начала'
  });

const editGuestSchema = z
  .object({
    guest_full_name: z.string().min(2, 'Минимум 2 символа'),
    plate_number: z.string().regex(plateRegex, 'Неверный формат номера'),
    valid_from: z.string().min(1, 'Укажите дату'),
    valid_to: z.string().min(1, 'Укажите дату'),
    status: z.enum(guestStatuses)
  })
  .refine((data) => new Date(data.valid_to) >= new Date(data.valid_from), {
    path: ['valid_to'],
    message: 'Дата окончания раньше даты начала'
  });

function toLocalDateTimeInput(iso: string): string {
  const d = new Date(iso);
  const pad = (n: number) => String(n).padStart(2, '0');
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`;
}

export default function AdminGuestsPage() {
  const navigate = useNavigate();
  const qc = useQueryClient();
  const [showDeleted, setShowDeleted] = useState(false);
  const [search, setSearch] = useState('');
  const [createOpen, setCreateOpen] = useState(false);
  const [editingGuest, setEditingGuest] = useState<GuestRequest | null>(null);

  const usersQuery = useQuery({
    queryKey: ['users', 'guests', 'all'],
    queryFn: async () => (await api.get<User[]>('/users', { params: { includeDeleted: true, limit: 500 } })).data
  });

  const guestsQuery = useQuery({
    queryKey: ['guest', 'admin', 'all'],
    queryFn: async () => (await api.get<GuestRequest[]>('/guest-requests', { params: { includeDeleted: true, limit: 500 } })).data
  });

  const createGuest = useMutation({
    mutationFn: (payload: { resident_user_id: string; guest_full_name: string; plate_number: string; valid_from: string; valid_to: string; status: string }) =>
      api.post('/guest-requests', payload),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['guest'] })
  });

  const updateGuest = useMutation({
    mutationFn: (payload: { id: string; data: Record<string, unknown> }) => api.patch(`/guest-requests/${payload.id}`, payload.data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['guest'] })
  });

  const deleteGuest = useMutation({
    mutationFn: (id: string) => api.delete(`/guest-requests/${id}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['guest'] })
  });

  const restoreGuest = useMutation({
    mutationFn: (id: string) => api.post(`/guest-requests/${id}/restore`, {}),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['guest'] })
  });

  const createForm = useForm<z.infer<typeof createGuestSchema>>({
    resolver: zodResolver(createGuestSchema),
    defaultValues: { resident_user_id: '', status: 'pending' }
  });

  const editForm = useForm<z.infer<typeof editGuestSchema>>({
    resolver: zodResolver(editGuestSchema),
    defaultValues: { status: 'pending' }
  });

  const residents = (usersQuery.data ?? [])
    .filter((user) => user.role === 'resident' && !user.deleted_at && !user.blocked_at)
    .sort((a, b) => a.full_name.localeCompare(b.full_name));
  const userByID = new Map((usersQuery.data ?? []).map((user) => [user.id, user]));

  const filteredGuests = useMemo(() => {
    const q = search.trim().toLowerCase();
    return (guestsQuery.data ?? [])
      .filter((guest) => (showDeleted ? true : !guest.deleted_at))
      .filter((guest) => {
        if (!q) return true;
        const resident = userByID.get(guest.resident_user_id);
        const residentText = resident ? `${resident.full_name} ${resident.plot_number || ''}` : guest.resident_user_id;
        return `${guest.guest_full_name} ${guest.plate_number} ${guest.status} ${residentText}`.toLowerCase().includes(q);
      })
      .sort((a, b) => new Date(b.valid_from).getTime() - new Date(a.valid_from).getTime());
  }, [search, showDeleted, guestsQuery.data, userByID]);

  const busy = createGuest.isPending || updateGuest.isPending || deleteGuest.isPending || restoreGuest.isPending;

  return (
    <Layout title="Гостевые заявки">
      <Stack direction="row" spacing={1} sx={{ mb: 2, alignItems: 'center', flexWrap: 'wrap' }}>
        <Tooltip title="Назад в админ-панель">
          <IconButton aria-label="Назад в админ-панель" onClick={() => navigate('/admin')}>
            <ArrowBack />
          </IconButton>
        </Tooltip>
        <Tooltip title="Добавить заявку">
          <IconButton aria-label="Добавить заявку" color="primary" onClick={() => setCreateOpen(true)}>
            <Add />
          </IconButton>
        </Tooltip>
        <TextField
          size="small"
          label="Поиск"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          sx={{ minWidth: { xs: '100%', md: 320 } }}
          InputProps={{
            startAdornment: (
              <InputAdornment position="start">
                <Search fontSize="small" />
              </InputAdornment>
            )
          }}
        />
        <Stack direction="row" spacing={1} alignItems="center">
          <Typography variant="body2">Удаленные</Typography>
          <Switch checked={showDeleted} onChange={(e) => setShowDeleted(e.target.checked)} />
        </Stack>
      </Stack>

      <Stack spacing={2}>
        {filteredGuests.map((guest) => {
          const resident = userByID.get(guest.resident_user_id);
          return (
            <Card key={guest.id}>
              <CardContent sx={{ display: 'flex', gap: 2, alignItems: 'center', justifyContent: 'space-between', flexWrap: 'wrap' }}>
                <Box>
                  <Typography variant="h6" sx={{ fontWeight: 700 }}>
                    {guest.guest_full_name}
                  </Typography>
                  <Typography variant="body2" color="text.secondary">
                    {guest.plate_number} · статус {guest.status}
                  </Typography>
                  <Typography variant="body2" color="text.secondary">
                    {resident ? resident.full_name : guest.resident_user_id}
                    {resident?.plot_number ? ` · участок ${resident.plot_number}` : ''}
                  </Typography>
                  <Typography variant="caption" color="text.secondary">
                    {new Date(guest.valid_from).toLocaleString('ru-RU')} - {new Date(guest.valid_to).toLocaleString('ru-RU')}
                  </Typography>
                  <Stack direction="row" spacing={1} sx={{ mt: 1 }}>
                    {guest.deleted_at ? (
                      <Chip size="small" label="Удалена" />
                    ) : (
                      <Chip size="small" color="success" label="Активна" />
                    )}
                  </Stack>
                </Box>
                <Stack direction="row" spacing={0.5}>
                  <Tooltip title="Изменить">
                    <span>
                      <IconButton
                        aria-label="Изменить заявку"
                        onClick={() => {
                          setEditingGuest(guest);
                          editForm.reset({
                            guest_full_name: guest.guest_full_name,
                            plate_number: guest.plate_number,
                            valid_from: toLocalDateTimeInput(guest.valid_from),
                            valid_to: toLocalDateTimeInput(guest.valid_to),
                            status: guestStatuses.includes(guest.status as (typeof guestStatuses)[number]) ? (guest.status as (typeof guestStatuses)[number]) : 'pending'
                          });
                        }}
                        disabled={!!guest.deleted_at || busy}
                      >
                        <Edit />
                      </IconButton>
                    </span>
                  </Tooltip>
                  {guest.deleted_at ? (
                    <Tooltip title="Восстановить">
                      <span>
                        <IconButton aria-label="Восстановить заявку" color="success" onClick={() => restoreGuest.mutate(guest.id)} disabled={busy}>
                          <Restore />
                        </IconButton>
                      </span>
                    </Tooltip>
                  ) : (
                    <Tooltip title="Удалить">
                      <span>
                        <IconButton
                          aria-label="Удалить заявку"
                          color="error"
                          onClick={() => {
                            if (window.confirm(`Удалить заявку ${guest.guest_full_name}?`)) {
                              deleteGuest.mutate(guest.id);
                            }
                          }}
                          disabled={busy}
                        >
                          <Delete />
                        </IconButton>
                      </span>
                    </Tooltip>
                  )}
                </Stack>
              </CardContent>
            </Card>
          );
        })}
      </Stack>

      <Dialog open={createOpen} onClose={() => setCreateOpen(false)} fullWidth maxWidth="sm">
        <DialogTitle>Добавление гостевой заявки</DialogTitle>
        <DialogContent>
          <Stack spacing={1.5} sx={{ mt: 1 }}>
            <TextField
              label="Житель"
              size="small"
              select
              SelectProps={{ displayEmpty: true }}
              {...createForm.register('resident_user_id')}
              error={!!createForm.formState.errors.resident_user_id}
              helperText={createForm.formState.errors.resident_user_id?.message}
              disabled={residents.length === 0}
            >
              <MenuItem value="">Выберите жителя</MenuItem>
              {residents.map((resident) => (
                <MenuItem key={resident.id} value={resident.id}>
                  {resident.full_name}
                  {resident.plot_number ? ` · участок ${resident.plot_number}` : ''}
                </MenuItem>
              ))}
            </TextField>
            <TextField
              label="ФИО гостя"
              size="small"
              {...createForm.register('guest_full_name')}
              error={!!createForm.formState.errors.guest_full_name}
              helperText={createForm.formState.errors.guest_full_name?.message}
            />
            <TextField
              label="Номер авто"
              size="small"
              {...createForm.register('plate_number')}
              error={!!createForm.formState.errors.plate_number}
              helperText={createForm.formState.errors.plate_number?.message}
            />
            <TextField
              label="С"
              size="small"
              type="datetime-local"
              InputLabelProps={{ shrink: true }}
              {...createForm.register('valid_from')}
              error={!!createForm.formState.errors.valid_from}
              helperText={createForm.formState.errors.valid_from?.message}
            />
            <TextField
              label="По"
              size="small"
              type="datetime-local"
              InputLabelProps={{ shrink: true }}
              {...createForm.register('valid_to')}
              error={!!createForm.formState.errors.valid_to}
              helperText={createForm.formState.errors.valid_to?.message}
            />
            <TextField label="Статус" size="small" select {...createForm.register('status')}>
              {guestStatuses.map((status) => (
                <MenuItem key={status} value={status}>
                  {status}
                </MenuItem>
              ))}
            </TextField>
          </Stack>
        </DialogContent>
        <DialogActions>
          <Tooltip title="Отмена">
            <IconButton aria-label="Отмена" onClick={() => setCreateOpen(false)}>
              <Close />
            </IconButton>
          </Tooltip>
          <Tooltip title="Сохранить">
            <span>
              <IconButton
                aria-label="Сохранить заявку"
                color="primary"
                disabled={busy}
                onClick={createForm.handleSubmit((data) => {
                  createGuest.mutate(
                    {
                      resident_user_id: data.resident_user_id,
                      guest_full_name: data.guest_full_name.trim(),
                      plate_number: data.plate_number.trim().toUpperCase(),
                      valid_from: new Date(data.valid_from).toISOString(),
                      valid_to: new Date(data.valid_to).toISOString(),
                      status: data.status
                    },
                    {
                      onSuccess: () => {
                        setCreateOpen(false);
                        createForm.reset({ resident_user_id: '', status: 'pending' });
                      }
                    }
                  );
                })}
              >
                <Check />
              </IconButton>
            </span>
          </Tooltip>
        </DialogActions>
      </Dialog>

      <Dialog open={!!editingGuest} onClose={() => setEditingGuest(null)} fullWidth maxWidth="sm">
        <DialogTitle>Редактирование заявки</DialogTitle>
        <DialogContent>
          <Stack spacing={1.5} sx={{ mt: 1 }}>
            <TextField
              label="ФИО гостя"
              size="small"
              {...editForm.register('guest_full_name')}
              error={!!editForm.formState.errors.guest_full_name}
              helperText={editForm.formState.errors.guest_full_name?.message}
            />
            <TextField
              label="Номер авто"
              size="small"
              {...editForm.register('plate_number')}
              error={!!editForm.formState.errors.plate_number}
              helperText={editForm.formState.errors.plate_number?.message}
            />
            <TextField
              label="С"
              size="small"
              type="datetime-local"
              InputLabelProps={{ shrink: true }}
              {...editForm.register('valid_from')}
              error={!!editForm.formState.errors.valid_from}
              helperText={editForm.formState.errors.valid_from?.message}
            />
            <TextField
              label="По"
              size="small"
              type="datetime-local"
              InputLabelProps={{ shrink: true }}
              {...editForm.register('valid_to')}
              error={!!editForm.formState.errors.valid_to}
              helperText={editForm.formState.errors.valid_to?.message}
            />
            <TextField label="Статус" size="small" select {...editForm.register('status')}>
              {guestStatuses.map((status) => (
                <MenuItem key={status} value={status}>
                  {status}
                </MenuItem>
              ))}
            </TextField>
          </Stack>
        </DialogContent>
        <DialogActions>
          <Tooltip title="Отмена">
            <IconButton aria-label="Отмена" onClick={() => setEditingGuest(null)}>
              <Close />
            </IconButton>
          </Tooltip>
          <Tooltip title="Сохранить">
            <span>
              <IconButton
                aria-label="Сохранить заявку"
                color="primary"
                disabled={busy || !editingGuest}
                onClick={editForm.handleSubmit((data) => {
                  if (!editingGuest) return;
                  updateGuest.mutate(
                    {
                      id: editingGuest.id,
                      data: {
                        guest_full_name: data.guest_full_name.trim(),
                        plate_number: data.plate_number.trim().toUpperCase(),
                        valid_from: new Date(data.valid_from).toISOString(),
                        valid_to: new Date(data.valid_to).toISOString(),
                        status: data.status
                      }
                    },
                    { onSuccess: () => setEditingGuest(null) }
                  );
                })}
              >
                <Check />
              </IconButton>
            </span>
          </Tooltip>
        </DialogActions>
      </Dialog>
    </Layout>
  );
}
