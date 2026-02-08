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
import { Pass, User } from '../api/types';

const plateRegex = /^[ABEKMHOPCTYXАВЕКМНОРСТУХ]\d{3}[ABEKMHOPCTYXАВЕКМНОРСТУХ]{2}\d{2,3}$/i;
const passStatuses = ['active', 'inactive', 'revoked'] as const;

const createPassSchema = z.object({
  owner_user_id: z.string().uuid('Выберите жителя'),
  plate_number: z.string().regex(plateRegex, 'Неверный формат номера'),
  vehicle_brand: z.string().optional(),
  vehicle_color: z.string().optional(),
  status: z.enum(passStatuses)
});

const editPassSchema = z.object({
  plate_number: z.string().regex(plateRegex, 'Неверный формат номера'),
  vehicle_brand: z.string().optional(),
  vehicle_color: z.string().optional(),
  status: z.enum(passStatuses)
});

export default function AdminPassesPage() {
  const navigate = useNavigate();
  const qc = useQueryClient();
  const [showDeleted, setShowDeleted] = useState(false);
  const [search, setSearch] = useState('');
  const [createOpen, setCreateOpen] = useState(false);
  const [editingPass, setEditingPass] = useState<Pass | null>(null);

  const usersQuery = useQuery({
    queryKey: ['users', 'passes', 'all'],
    queryFn: async () => (await api.get<User[]>('/users', { params: { includeDeleted: true, limit: 500 } })).data
  });

  const passesQuery = useQuery({
    queryKey: ['passes', 'admin', 'all'],
    queryFn: async () => (await api.get<Pass[]>('/passes', { params: { includeDeleted: true, limit: 500 } })).data
  });

  const createPass = useMutation({
    mutationFn: (payload: { owner_user_id: string; plate_number: string; vehicle_brand?: string; vehicle_color?: string; status: string }) =>
      api.post('/passes', payload),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['passes'] })
  });

  const updatePass = useMutation({
    mutationFn: (payload: { id: string; data: Record<string, unknown> }) => api.patch(`/passes/${payload.id}`, payload.data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['passes'] })
  });

  const deletePass = useMutation({
    mutationFn: (id: string) => api.delete(`/passes/${id}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['passes'] })
  });

  const restorePass = useMutation({
    mutationFn: (id: string) => api.post(`/passes/${id}/restore`, {}),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['passes'] })
  });

  const createForm = useForm<z.infer<typeof createPassSchema>>({
    resolver: zodResolver(createPassSchema),
    defaultValues: { owner_user_id: '', status: 'active' }
  });

  const editForm = useForm<z.infer<typeof editPassSchema>>({
    resolver: zodResolver(editPassSchema),
    defaultValues: { status: 'active' }
  });

  const userByID = new Map((usersQuery.data ?? []).map((user) => [user.id, user]));
  const residents = (usersQuery.data ?? [])
    .filter((user) => user.role === 'resident' && !user.deleted_at && !user.blocked_at)
    .sort((a, b) => a.full_name.localeCompare(b.full_name));

  const filteredPasses = useMemo(() => {
    const q = search.trim().toLowerCase();
    return (passesQuery.data ?? [])
      .filter((pass) => (showDeleted ? true : !pass.deleted_at))
      .filter((pass) => {
        if (!q) return true;
        const owner = userByID.get(pass.owner_user_id);
        const ownerText = owner ? `${owner.full_name} ${owner.plot_number || ''}` : pass.owner_user_id;
        return `${pass.plate_number} ${pass.status} ${ownerText}`.toLowerCase().includes(q);
      })
      .sort((a, b) => a.plate_number.localeCompare(b.plate_number));
  }, [search, showDeleted, passesQuery.data, userByID]);

  const busy = createPass.isPending || updatePass.isPending || deletePass.isPending || restorePass.isPending;

  return (
    <Layout title="Пропуска">
      <Stack direction="row" spacing={1} sx={{ mb: 2, alignItems: 'center', flexWrap: 'wrap' }}>
        <Tooltip title="Назад в админ-панель">
          <IconButton aria-label="Назад в админ-панель" onClick={() => navigate('/admin')}>
            <ArrowBack />
          </IconButton>
        </Tooltip>
        <Tooltip title="Добавить пропуск">
          <IconButton aria-label="Добавить пропуск" color="primary" onClick={() => setCreateOpen(true)}>
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
        {filteredPasses.map((pass) => {
          const owner = userByID.get(pass.owner_user_id);
          return (
            <Card key={pass.id}>
              <CardContent sx={{ display: 'flex', gap: 2, alignItems: 'center', justifyContent: 'space-between', flexWrap: 'wrap' }}>
                <Box>
                  <Typography variant="h6" sx={{ fontWeight: 700 }}>
                    {pass.plate_number}
                  </Typography>
                  <Typography variant="body2" color="text.secondary">
                    {owner ? owner.full_name : pass.owner_user_id}
                    {owner?.plot_number ? ` · участок ${owner.plot_number}` : ''}
                    {` · статус ${pass.status}`}
                  </Typography>
                  <Stack direction="row" spacing={1} sx={{ mt: 1 }}>
                    {pass.deleted_at ? (
                      <Chip size="small" label="Удален" />
                    ) : (
                      <Chip size="small" color="success" label="Активен" />
                    )}
                  </Stack>
                </Box>
                <Stack direction="row" spacing={0.5}>
                  <Tooltip title="Изменить">
                    <span>
                      <IconButton
                        aria-label="Изменить пропуск"
                        onClick={() => {
                          setEditingPass(pass);
                          editForm.reset({
                            plate_number: pass.plate_number,
                            vehicle_brand: pass.vehicle_brand ?? '',
                            vehicle_color: pass.vehicle_color ?? '',
                            status: passStatuses.includes(pass.status as (typeof passStatuses)[number]) ? (pass.status as (typeof passStatuses)[number]) : 'active'
                          });
                        }}
                        disabled={!!pass.deleted_at || busy}
                      >
                        <Edit />
                      </IconButton>
                    </span>
                  </Tooltip>
                  {pass.deleted_at ? (
                    <Tooltip title="Восстановить">
                      <span>
                        <IconButton aria-label="Восстановить пропуск" color="success" onClick={() => restorePass.mutate(pass.id)} disabled={busy}>
                          <Restore />
                        </IconButton>
                      </span>
                    </Tooltip>
                  ) : (
                    <Tooltip title="Удалить">
                      <span>
                        <IconButton
                          aria-label="Удалить пропуск"
                          color="error"
                          onClick={() => {
                            if (window.confirm(`Удалить пропуск ${pass.plate_number}?`)) {
                              deletePass.mutate(pass.id);
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
        <DialogTitle>Добавление пропуска</DialogTitle>
        <DialogContent>
          <Stack spacing={1.5} sx={{ mt: 1 }}>
            <TextField
              label="Житель"
              size="small"
              select
              SelectProps={{ displayEmpty: true }}
              {...createForm.register('owner_user_id')}
              error={!!createForm.formState.errors.owner_user_id}
              helperText={createForm.formState.errors.owner_user_id?.message}
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
              label="Номер авто"
              size="small"
              {...createForm.register('plate_number')}
              error={!!createForm.formState.errors.plate_number}
              helperText={createForm.formState.errors.plate_number?.message}
            />
            <TextField label="Марка" size="small" {...createForm.register('vehicle_brand')} />
            <TextField label="Цвет" size="small" {...createForm.register('vehicle_color')} />
            <TextField label="Статус" size="small" select {...createForm.register('status')}>
              {passStatuses.map((status) => (
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
                aria-label="Сохранить пропуск"
                color="primary"
                disabled={busy}
                onClick={createForm.handleSubmit((data) => {
                  createPass.mutate(
                    {
                      owner_user_id: data.owner_user_id,
                      plate_number: data.plate_number.trim().toUpperCase(),
                      vehicle_brand: data.vehicle_brand?.trim() || undefined,
                      vehicle_color: data.vehicle_color?.trim() || undefined,
                      status: data.status
                    },
                    {
                      onSuccess: () => {
                        setCreateOpen(false);
                        createForm.reset({ owner_user_id: '', status: 'active' });
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

      <Dialog open={!!editingPass} onClose={() => setEditingPass(null)} fullWidth maxWidth="sm">
        <DialogTitle>Редактирование пропуска</DialogTitle>
        <DialogContent>
          <Stack spacing={1.5} sx={{ mt: 1 }}>
            <TextField
              label="Номер авто"
              size="small"
              {...editForm.register('plate_number')}
              error={!!editForm.formState.errors.plate_number}
              helperText={editForm.formState.errors.plate_number?.message}
            />
            <TextField label="Марка" size="small" {...editForm.register('vehicle_brand')} />
            <TextField label="Цвет" size="small" {...editForm.register('vehicle_color')} />
            <TextField label="Статус" size="small" select {...editForm.register('status')}>
              {passStatuses.map((status) => (
                <MenuItem key={status} value={status}>
                  {status}
                </MenuItem>
              ))}
            </TextField>
          </Stack>
        </DialogContent>
        <DialogActions>
          <Tooltip title="Отмена">
            <IconButton aria-label="Отмена" onClick={() => setEditingPass(null)}>
              <Close />
            </IconButton>
          </Tooltip>
          <Tooltip title="Сохранить">
            <span>
              <IconButton
                aria-label="Сохранить пропуск"
                color="primary"
                disabled={busy || !editingPass}
                onClick={editForm.handleSubmit((data) => {
                  if (!editingPass) return;
                  updatePass.mutate(
                    {
                      id: editingPass.id,
                      data: {
                        plate_number: data.plate_number.trim().toUpperCase(),
                        vehicle_brand: data.vehicle_brand?.trim() || null,
                        vehicle_color: data.vehicle_color?.trim() || null,
                        status: data.status
                      }
                    },
                    { onSuccess: () => setEditingPass(null) }
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
