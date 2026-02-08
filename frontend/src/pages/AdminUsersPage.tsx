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
import Block from '@mui/icons-material/Block';
import LockOpen from '@mui/icons-material/LockOpen';
import Check from '@mui/icons-material/Check';
import Close from '@mui/icons-material/Close';
import Search from '@mui/icons-material/Search';
import { z } from 'zod';
import { zodResolver } from '@hookform/resolvers/zod';
import { useForm } from 'react-hook-form';
import { useNavigate } from 'react-router-dom';
import { Layout } from '../components/Layout';
import api from '../api/client';
import { Role, User } from '../api/types';
import { ROLE_OPTIONS, roleLabel } from '../utils/roles';

const createSchema = z
  .object({
    full_name: z.string().min(2, 'Минимум 2 символа'),
    email: z.string().email('Некорректный email'),
    password: z.string().min(6, 'Минимум 6 символов'),
    role: z.enum(['admin', 'guard', 'resident']),
    plot_number: z.string().optional()
  })
  .refine((data) => (data.role !== 'resident' ? true : !!data.plot_number?.trim()), {
    path: ['plot_number'],
    message: 'Укажите участок'
  });

const editSchema = z
  .object({
    full_name: z.string().min(2, 'Минимум 2 символа'),
    email: z.string().email('Некорректный email'),
    password: z.string().optional(),
    role: z.enum(['admin', 'guard', 'resident']),
    plot_number: z.string().optional()
  })
  .refine((data) => (data.role !== 'resident' ? true : !!data.plot_number?.trim()), {
    path: ['plot_number'],
    message: 'Укажите участок'
  })
  .refine((data) => (!data.password ? true : data.password.length >= 6), {
    path: ['password'],
    message: 'Минимум 6 символов'
  });

export default function AdminUsersPage() {
  const navigate = useNavigate();
  const qc = useQueryClient();
  const [search, setSearch] = useState('');
  const [showDeleted, setShowDeleted] = useState(false);
  const [showBlocked, setShowBlocked] = useState(true);
  const [createOpen, setCreateOpen] = useState(false);
  const [editingUser, setEditingUser] = useState<User | null>(null);

  const usersQuery = useQuery({
    queryKey: ['users', 'admin', 'all'],
    queryFn: async () => (await api.get<User[]>('/users', { params: { includeDeleted: true, limit: 500 } })).data
  });

  const createUser = useMutation({
    mutationFn: (payload: { email: string; password: string; role: Role; full_name: string; plot_number?: string }) => api.post('/users', payload),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['users'] })
  });

  const updateUser = useMutation({
    mutationFn: (payload: { id: string; data: Record<string, unknown> }) => api.patch(`/users/${payload.id}`, payload.data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['users'] })
  });

  const deleteUser = useMutation({
    mutationFn: (id: string) => api.delete(`/users/${id}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['users'] })
  });

  const restoreUser = useMutation({
    mutationFn: (id: string) => api.post(`/users/${id}/restore`, {}),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['users'] })
  });

  const blockUser = useMutation({
    mutationFn: (id: string) => api.post(`/users/${id}/block`, {}),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['users'] })
  });

  const unblockUser = useMutation({
    mutationFn: (id: string) => api.post(`/users/${id}/unblock`, {}),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['users'] })
  });

  const createForm = useForm<z.infer<typeof createSchema>>({
    resolver: zodResolver(createSchema),
    defaultValues: { role: 'resident', plot_number: '' }
  });

  const editForm = useForm<z.infer<typeof editSchema>>({
    resolver: zodResolver(editSchema),
    defaultValues: { role: 'resident', plot_number: '', password: '' }
  });

  const createRole = createForm.watch('role');
  const editRole = editForm.watch('role');

  const filteredUsers = useMemo(() => {
    const query = search.trim().toLowerCase();
    return (usersQuery.data ?? [])
      .filter((user) => (showDeleted ? true : !user.deleted_at))
      .filter((user) => (showBlocked ? true : !user.blocked_at))
      .filter((user) => {
        if (!query) return true;
        const parts = [user.full_name, user.email, user.plot_number || '', roleLabel(user.role)].join(' ').toLowerCase();
        return parts.includes(query);
      })
      .sort((a, b) => a.full_name.localeCompare(b.full_name));
  }, [search, showDeleted, showBlocked, usersQuery.data]);

  const busy =
    createUser.isPending ||
    updateUser.isPending ||
    deleteUser.isPending ||
    restoreUser.isPending ||
    blockUser.isPending ||
    unblockUser.isPending;

  const onOpenEdit = (user: User) => {
    setEditingUser(user);
    editForm.reset({
      full_name: user.full_name,
      email: user.email,
      role: user.role,
      password: '',
      plot_number: user.plot_number ?? ''
    });
  };

  return (
    <Layout title="Пользователи">
      <Stack direction="row" spacing={1} sx={{ mb: 2, alignItems: 'center', flexWrap: 'wrap' }}>
        <Tooltip title="Назад в админ-панель">
          <IconButton aria-label="Назад в админ-панель" onClick={() => navigate('/admin')}>
            <ArrowBack />
          </IconButton>
        </Tooltip>
        <Tooltip title="Добавить пользователя">
          <IconButton aria-label="Добавить пользователя" color="primary" onClick={() => setCreateOpen(true)}>
            <Add />
          </IconButton>
        </Tooltip>
        <TextField
          size="small"
          label="Поиск"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          sx={{ minWidth: { xs: '100%', md: 360 } }}
          InputProps={{
            startAdornment: (
              <InputAdornment position="start">
                <Search fontSize="small" />
              </InputAdornment>
            )
          }}
        />
      </Stack>

      <Stack direction="row" spacing={2} alignItems="center" sx={{ mb: 2, flexWrap: 'wrap' }}>
        <Stack direction="row" spacing={1} alignItems="center">
          <Typography variant="body2">Показывать удаленных</Typography>
          <Switch checked={showDeleted} onChange={(e) => setShowDeleted(e.target.checked)} />
        </Stack>
        <Stack direction="row" spacing={1} alignItems="center">
          <Typography variant="body2">Показывать заблокированных</Typography>
          <Switch checked={showBlocked} onChange={(e) => setShowBlocked(e.target.checked)} />
        </Stack>
      </Stack>

      <Stack spacing={2}>
        {filteredUsers.map((user) => (
          <Card key={user.id}>
            <CardContent sx={{ display: 'flex', gap: 2, alignItems: 'center', justifyContent: 'space-between', flexWrap: 'wrap' }}>
              <Box>
                <Typography variant="h6" sx={{ fontWeight: 700 }}>
                  {user.full_name}
                </Typography>
                <Typography variant="body2" color="text.secondary">
                  {user.email} · {roleLabel(user.role)}
                  {user.plot_number ? ` · участок ${user.plot_number}` : ''}
                </Typography>
                <Stack direction="row" spacing={1} sx={{ mt: 1, flexWrap: 'wrap' }}>
                  {!user.deleted_at && !user.blocked_at && <Chip size="small" color="success" label="Активен" />}
                  {user.blocked_at && <Chip size="small" color="warning" label="Заблокирован" />}
                  {user.deleted_at && <Chip size="small" color="default" label="Удален" />}
                </Stack>
              </Box>
              <Stack direction="row" spacing={0.5}>
                <Tooltip title="Изменить">
                  <span>
                    <IconButton aria-label="Изменить пользователя" onClick={() => onOpenEdit(user)} disabled={!!user.deleted_at || busy}>
                      <Edit />
                    </IconButton>
                  </span>
                </Tooltip>
                {user.deleted_at ? (
                  <Tooltip title="Восстановить">
                    <span>
                      <IconButton aria-label="Восстановить пользователя" color="success" disabled={busy} onClick={() => restoreUser.mutate(user.id)}>
                        <Restore />
                      </IconButton>
                    </span>
                  </Tooltip>
                ) : (
                  <Tooltip title="Удалить">
                    <span>
                      <IconButton
                        aria-label="Удалить пользователя"
                        color="error"
                        disabled={busy}
                        onClick={() => {
                          if (window.confirm(`Удалить пользователя ${user.full_name}?`)) {
                            deleteUser.mutate(user.id);
                          }
                        }}
                      >
                        <Delete />
                      </IconButton>
                    </span>
                  </Tooltip>
                )}
                {!user.deleted_at && (
                  user.blocked_at ? (
                    <Tooltip title="Разблокировать">
                      <span>
                        <IconButton aria-label="Разблокировать пользователя" color="success" disabled={busy} onClick={() => unblockUser.mutate(user.id)}>
                          <LockOpen />
                        </IconButton>
                      </span>
                    </Tooltip>
                  ) : (
                    <Tooltip title="Заблокировать">
                      <span>
                        <IconButton aria-label="Заблокировать пользователя" color="warning" disabled={busy} onClick={() => blockUser.mutate(user.id)}>
                          <Block />
                        </IconButton>
                      </span>
                    </Tooltip>
                  )
                )}
              </Stack>
            </CardContent>
          </Card>
        ))}
      </Stack>

      <Dialog open={createOpen} onClose={() => setCreateOpen(false)} fullWidth maxWidth="sm">
        <DialogTitle>Добавление пользователя</DialogTitle>
        <DialogContent>
          <Stack spacing={1.5} sx={{ mt: 1 }}>
            <TextField
              label="ФИО"
              size="small"
              {...createForm.register('full_name')}
              error={!!createForm.formState.errors.full_name}
              helperText={createForm.formState.errors.full_name?.message}
            />
            <TextField
              label="Email"
              size="small"
              {...createForm.register('email')}
              error={!!createForm.formState.errors.email}
              helperText={createForm.formState.errors.email?.message}
            />
            <TextField
              label="Пароль"
              size="small"
              type="password"
              {...createForm.register('password')}
              error={!!createForm.formState.errors.password}
              helperText={createForm.formState.errors.password?.message}
            />
            <TextField
              label="Роль"
              size="small"
              select
              {...createForm.register('role')}
              error={!!createForm.formState.errors.role}
              helperText={createForm.formState.errors.role?.message}
            >
              {ROLE_OPTIONS.map((option) => (
                <MenuItem key={option.value} value={option.value}>
                  {option.label}
                </MenuItem>
              ))}
            </TextField>
            {createRole === 'resident' && (
              <TextField
                label="Участок"
                size="small"
                {...createForm.register('plot_number')}
                error={!!createForm.formState.errors.plot_number}
                helperText={createForm.formState.errors.plot_number?.message}
              />
            )}
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
                aria-label="Сохранить пользователя"
                color="primary"
                disabled={busy}
                onClick={createForm.handleSubmit((data) => {
                  createUser.mutate(
                    {
                      email: data.email.trim(),
                      password: data.password,
                      role: data.role,
                      full_name: data.full_name.trim(),
                      plot_number: data.role === 'resident' ? data.plot_number?.trim() || '' : ''
                    },
                    {
                      onSuccess: () => {
                        setCreateOpen(false);
                        createForm.reset({ role: 'resident', plot_number: '' });
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

      <Dialog open={!!editingUser} onClose={() => setEditingUser(null)} fullWidth maxWidth="sm">
        <DialogTitle>Редактирование пользователя</DialogTitle>
        <DialogContent>
          <Stack spacing={1.5} sx={{ mt: 1 }}>
            <TextField
              label="ФИО"
              size="small"
              {...editForm.register('full_name')}
              error={!!editForm.formState.errors.full_name}
              helperText={editForm.formState.errors.full_name?.message}
            />
            <TextField
              label="Email"
              size="small"
              {...editForm.register('email')}
              error={!!editForm.formState.errors.email}
              helperText={editForm.formState.errors.email?.message}
            />
            <TextField
              label="Новый пароль (опционально)"
              size="small"
              type="password"
              {...editForm.register('password')}
              error={!!editForm.formState.errors.password}
              helperText={editForm.formState.errors.password?.message}
            />
            <TextField
              label="Роль"
              size="small"
              select
              {...editForm.register('role')}
              error={!!editForm.formState.errors.role}
              helperText={editForm.formState.errors.role?.message}
            >
              {ROLE_OPTIONS.map((option) => (
                <MenuItem key={option.value} value={option.value}>
                  {option.label}
                </MenuItem>
              ))}
            </TextField>
            {editRole === 'resident' && (
              <TextField
                label="Участок"
                size="small"
                {...editForm.register('plot_number')}
                error={!!editForm.formState.errors.plot_number}
                helperText={editForm.formState.errors.plot_number?.message}
              />
            )}
          </Stack>
        </DialogContent>
        <DialogActions>
          <Tooltip title="Отмена">
            <IconButton aria-label="Отмена" onClick={() => setEditingUser(null)}>
              <Close />
            </IconButton>
          </Tooltip>
          <Tooltip title="Сохранить">
            <span>
              <IconButton
                aria-label="Сохранить пользователя"
                color="primary"
                disabled={busy || !editingUser}
                onClick={editForm.handleSubmit((data) => {
                  if (!editingUser) return;
                  const payload: Record<string, unknown> = {
                    full_name: data.full_name.trim(),
                    email: data.email.trim(),
                    role: data.role,
                    plot_number: data.role === 'resident' ? data.plot_number?.trim() || '' : ''
                  };
                  if (data.password?.trim()) {
                    payload.password = data.password.trim();
                  }
                  updateUser.mutate(
                    { id: editingUser.id, data: payload },
                    { onSuccess: () => setEditingUser(null) }
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
