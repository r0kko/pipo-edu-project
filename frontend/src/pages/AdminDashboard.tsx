import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  Box,
  Button,
  Card,
  CardContent,
  Divider,
  Grid,
  MenuItem,
  Stack,
  Switch,
  TextField,
  Typography
} from '@mui/material';
import { z } from 'zod';
import { zodResolver } from '@hookform/resolvers/zod';
import { useForm } from 'react-hook-form';
import { Layout } from '../components/Layout';
import api from '../api/client';
import { GuestRequest, Pass, User } from '../api/types';

const plateRegex = /^[ABEKMHOPCTYXАВЕКМНОРСТУХ]\d{3}[ABEKMHOPCTYXАВЕКМНОРСТУХ]{2}\d{2,3}$/i;

const userSchema = z
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

const passSchema = z.object({
  owner_user_id: z.string().min(1, 'Выберите жителя').uuid('Неверный UUID'),
  plate_number: z.string().regex(plateRegex, 'Неверный формат номера'),
  vehicle_brand: z.string().optional(),
  vehicle_color: z.string().optional()
});

const optionalTrimmed = <T extends z.ZodTypeAny>(schema: T) =>
  z.preprocess((val) => {
    if (typeof val !== 'string') {
      return val;
    }
    const trimmed = val.trim();
    return trimmed === '' ? undefined : trimmed;
  }, schema.optional());

const updateUserSchema = z.object({
  user_id: z.string().uuid('Выберите пользователя'),
  email: optionalTrimmed(z.string().email('Некорректный email')),
  password: optionalTrimmed(z.string().min(6, 'Минимум 6 символов')),
  role: optionalTrimmed(z.enum(['admin', 'guard', 'resident'])),
  full_name: optionalTrimmed(z.string().min(2, 'Минимум 2 символа')),
  plot_number: optionalTrimmed(z.string())
});

const guestSchema = z
  .object({
    resident_user_id: z.string().min(1, 'Выберите жителя').uuid('Неверный UUID'),
    guest_full_name: z.string().min(2, 'Минимум 2 символа'),
    plate_number: z.string().regex(plateRegex, 'Неверный формат номера'),
    valid_from: z.string().min(1, 'Укажите дату'),
    valid_to: z.string().min(1, 'Укажите дату')
  })
  .refine((data) => new Date(data.valid_to) >= new Date(data.valid_from), {
    path: ['valid_to'],
    message: 'Дата окончания раньше даты начала'
  });

export default function AdminDashboard() {
  const qc = useQueryClient();
  const [showDeleted, setShowDeleted] = useState(false);

  const usersQuery = useQuery({
    queryKey: ['users', showDeleted],
    queryFn: async () => (await api.get<User[]>('/users', { params: { includeDeleted: showDeleted } })).data
  });

  const passesQuery = useQuery({
    queryKey: ['passes', showDeleted],
    queryFn: async () => (await api.get<Pass[]>('/passes', { params: { includeDeleted: showDeleted } })).data
  });

  const guestsQuery = useQuery({
    queryKey: ['guest', showDeleted],
    queryFn: async () => (await api.get<GuestRequest[]>('/guest-requests', { params: { includeDeleted: showDeleted } })).data
  });

  const createUser = useMutation({
    mutationFn: (payload: { email: string; password: string; role: string; full_name: string; plot_number?: string }) =>
      api.post('/users', payload),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['users'] })
  });

  const createPass = useMutation({
    mutationFn: (payload: { owner_user_id: string; plate_number: string; vehicle_brand?: string; vehicle_color?: string }) =>
      api.post('/passes', payload),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['passes'] })
  });

  const createGuest = useMutation({
    mutationFn: (payload: { resident_user_id: string; guest_full_name: string; plate_number: string; valid_from: string; valid_to: string }) =>
      api.post('/guest-requests', payload),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['guest'] })
  });

  const userForm = useForm<z.infer<typeof userSchema>>({
    resolver: zodResolver(userSchema),
    defaultValues: { role: 'resident', plot_number: '' }
  });

  const updateUserForm = useForm<z.infer<typeof updateUserSchema>>({
    resolver: zodResolver(updateUserSchema),
    defaultValues: {
      user_id: '',
      email: '',
      password: '',
      role: undefined,
      full_name: '',
      plot_number: ''
    }
  });

  const updateUser = useMutation({
    mutationFn: (payload: { id: string; data: Record<string, unknown> }) =>
      api.patch(`/users/${payload.id}`, payload.data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['users'] })
  });

  const passForm = useForm<z.infer<typeof passSchema>>({
    resolver: zodResolver(passSchema),
    defaultValues: { owner_user_id: '' }
  });

  const guestForm = useForm<z.infer<typeof guestSchema>>({
    resolver: zodResolver(guestSchema),
    defaultValues: { resident_user_id: '' }
  });

  const residents = (usersQuery.data ?? [])
    .filter((user) => user.role === 'resident' && !user.deleted_at)
    .sort((a, b) => a.full_name.localeCompare(b.full_name));
  const activeUsers = (usersQuery.data ?? [])
    .filter((user) => !user.deleted_at)
    .sort((a, b) => a.full_name.localeCompare(b.full_name));
  const userById = new Map((usersQuery.data ?? []).map((user) => [user.id, user]));
  const selectedRole = userForm.watch('role');
  const selectedUpdateRole = updateUserForm.watch('role');
  const selectedUserId = updateUserForm.watch('user_id');
  const selectedUser = userById.get(selectedUserId);

  return (
    <Layout title="Администрирование">
      <Stack direction="row" alignItems="center" spacing={1} sx={{ mb: 2 }}>
        <Typography variant="body2">Показать удаленные</Typography>
        <Switch checked={showDeleted} onChange={(e) => setShowDeleted(e.target.checked)} />
      </Stack>
      <Grid container spacing={3}>
        <Grid item xs={12} md={4}>
          <Card sx={{ height: '100%' }}>
            <CardContent>
              <Typography variant="h6" sx={{ mb: 1, fontWeight: 700 }}>
                Пользователи
              </Typography>
              <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                Создание и контроль ролей
              </Typography>
              <Stack
                spacing={1.5}
                component="form"
                onSubmit={userForm.handleSubmit((data) => {
                  createUser.mutate({
                    email: data.email.trim(),
                    password: data.password,
                    role: data.role,
                    full_name: data.full_name.trim(),
                    plot_number: data.role === 'resident' ? data.plot_number?.trim() || '' : ''
                  });
                  userForm.reset({ role: 'resident', plot_number: '' });
                })}
              >
                <TextField
                  label="ФИО"
                  size="small"
                  {...userForm.register('full_name')}
                  error={!!userForm.formState.errors.full_name}
                  helperText={userForm.formState.errors.full_name?.message}
                />
                <TextField
                  label="Email"
                  size="small"
                  {...userForm.register('email')}
                  error={!!userForm.formState.errors.email}
                  helperText={userForm.formState.errors.email?.message}
                />
                <TextField
                  label="Пароль"
                  size="small"
                  type="password"
                  {...userForm.register('password')}
                  error={!!userForm.formState.errors.password}
                  helperText={userForm.formState.errors.password?.message}
                />
                <TextField
                  label="Роль"
                  size="small"
                  select
                  {...userForm.register('role')}
                  error={!!userForm.formState.errors.role}
                  helperText={userForm.formState.errors.role?.message}
                >
                  <MenuItem value="admin">Администратор</MenuItem>
                  <MenuItem value="guard">Охрана</MenuItem>
                  <MenuItem value="resident">Житель</MenuItem>
                </TextField>
                {selectedRole === 'resident' && (
                  <TextField
                    label="Участок"
                    size="small"
                    {...userForm.register('plot_number')}
                    error={!!userForm.formState.errors.plot_number}
                    helperText={userForm.formState.errors.plot_number?.message}
                  />
                )}
                <Button type="submit" variant="contained">Добавить пользователя</Button>
              </Stack>
              <Divider sx={{ my: 2 }} />
              <Box sx={{ maxHeight: 240, overflow: 'auto' }}>
                {usersQuery.data?.map((user) => (
                  <Box key={user.id} sx={{ mb: 1 }}>
                    <Typography variant="body2" sx={{ fontWeight: 600 }}>
                      {user.full_name} · {user.role}
                    </Typography>
                    <Typography variant="caption" color="text.secondary">
                      {user.email}
                      {user.plot_number ? ` · участок ${user.plot_number}` : ''}
                    </Typography>
                  </Box>
                ))}
              </Box>
              <Divider sx={{ my: 2 }} />
              <Typography variant="subtitle2" sx={{ mb: 1, fontWeight: 600 }}>
                Обновить пользователя
              </Typography>
              <Stack
                spacing={1.5}
                component="form"
                onSubmit={updateUserForm.handleSubmit((data) => {
                  if (!selectedUser) {
                    updateUserForm.setError('user_id', { message: 'Выберите пользователя' });
                    return;
                  }
                  const payload: Record<string, unknown> = {};
                  if (data.email) payload.email = data.email;
                  if (data.password) payload.password = data.password;
                  if (data.role) payload.role = data.role;
                  if (data.full_name) payload.full_name = data.full_name;
                  if (data.plot_number) payload.plot_number = data.plot_number;

                  if (Object.keys(payload).length === 0) {
                    updateUserForm.setError('email', { message: 'Укажите хотя бы одно изменение' });
                    return;
                  }

                  const nextRole = (data.role as User['role'] | undefined) ?? selectedUser.role;
                  const nextPlot = data.plot_number ?? selectedUser.plot_number ?? '';
                  if (nextRole === 'resident' && !nextPlot.trim()) {
                    updateUserForm.setError('plot_number', { message: 'Укажите участок' });
                    return;
                  }

                  updateUser.mutate({
                    id: selectedUser.id,
                    data: payload
                  });
                  updateUserForm.reset({
                    user_id: '',
                    email: '',
                    password: '',
                    role: undefined,
                    full_name: '',
                    plot_number: ''
                  });
                })}
              >
                <TextField
                  label="Пользователь"
                  size="small"
                  select
                  SelectProps={{ displayEmpty: true }}
                  {...updateUserForm.register('user_id')}
                  error={!!updateUserForm.formState.errors.user_id}
                  helperText={updateUserForm.formState.errors.user_id?.message}
                >
                  <MenuItem value="">
                    Выберите пользователя
                  </MenuItem>
                  {activeUsers.map((user) => (
                    <MenuItem key={user.id} value={user.id}>
                      {user.full_name} · {user.role}
                    </MenuItem>
                  ))}
                </TextField>
                <TextField
                  label="Email"
                  size="small"
                  {...updateUserForm.register('email')}
                  error={!!updateUserForm.formState.errors.email}
                  helperText={updateUserForm.formState.errors.email?.message || (selectedUser ? `Текущее: ${selectedUser.email}` : undefined)}
                />
                <TextField
                  label="Пароль"
                  size="small"
                  type="password"
                  {...updateUserForm.register('password')}
                  error={!!updateUserForm.formState.errors.password}
                  helperText={updateUserForm.formState.errors.password?.message || 'Оставьте пустым, если не нужно менять'}
                />
                <TextField
                  label="Роль"
                  size="small"
                  select
                  SelectProps={{ displayEmpty: true }}
                  {...updateUserForm.register('role')}
                  error={!!updateUserForm.formState.errors.role}
                  helperText={updateUserForm.formState.errors.role?.message || (selectedUser ? `Текущая: ${selectedUser.role}` : undefined)}
                >
                  <MenuItem value="">
                    Не менять
                  </MenuItem>
                  <MenuItem value="admin">Администратор</MenuItem>
                  <MenuItem value="guard">Охрана</MenuItem>
                  <MenuItem value="resident">Житель</MenuItem>
                </TextField>
                {(selectedUpdateRole === 'resident' || selectedUser?.role === 'resident') && (
                  <TextField
                    label="Участок"
                    size="small"
                    {...updateUserForm.register('plot_number')}
                    error={!!updateUserForm.formState.errors.plot_number}
                    helperText={updateUserForm.formState.errors.plot_number?.message || (selectedUser?.plot_number ? `Текущий: ${selectedUser.plot_number}` : undefined)}
                  />
                )}
                <Button type="submit" variant="outlined">Сохранить изменения</Button>
              </Stack>
            </CardContent>
          </Card>
        </Grid>
        <Grid item xs={12} md={4}>
          <Card sx={{ height: '100%' }}>
            <CardContent>
              <Typography variant="h6" sx={{ mb: 1, fontWeight: 700 }}>
                Пропуска
              </Typography>
              <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                Создание пропусков для жителей
              </Typography>
              <Stack
                spacing={1.5}
                component="form"
                onSubmit={passForm.handleSubmit((data) => {
                  createPass.mutate({
                    owner_user_id: data.owner_user_id.trim(),
                    plate_number: data.plate_number.trim().toUpperCase(),
                    vehicle_brand: data.vehicle_brand?.trim() || undefined,
                    vehicle_color: data.vehicle_color?.trim() || undefined
                  });
                  passForm.reset();
                })}
              >
                <TextField
                  label="Житель"
                  size="small"
                  select
                  SelectProps={{ displayEmpty: true }}
                  {...passForm.register('owner_user_id')}
                  error={!!passForm.formState.errors.owner_user_id}
                  helperText={residents.length === 0 ? 'Сначала создайте жителя' : passForm.formState.errors.owner_user_id?.message}
                  disabled={residents.length === 0}
                >
                  <MenuItem value="">
                    Выберите жителя
                  </MenuItem>
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
                  {...passForm.register('plate_number')}
                  error={!!passForm.formState.errors.plate_number}
                  helperText={passForm.formState.errors.plate_number?.message}
                />
                <TextField
                  label="Марка"
                  size="small"
                  {...passForm.register('vehicle_brand')}
                />
                <TextField
                  label="Цвет"
                  size="small"
                  {...passForm.register('vehicle_color')}
                />
                <Button type="submit" variant="contained">Добавить пропуск</Button>
              </Stack>
              <Divider sx={{ my: 2 }} />
              <Box sx={{ maxHeight: 240, overflow: 'auto' }}>
                {passesQuery.data?.map((pass) => (
                  <Box key={pass.id} sx={{ mb: 1 }}>
                    <Typography variant="body2" sx={{ fontWeight: 600 }}>
                      {pass.plate_number} · {pass.status}
                    </Typography>
                    <Typography variant="caption" color="text.secondary">
                      {(() => {
                        const owner = userById.get(pass.owner_user_id);
                        if (!owner) {
                          return `владелец: ${pass.owner_user_id}`;
                        }
                        return `владелец: ${owner.full_name}${owner.plot_number ? ` · участок ${owner.plot_number}` : ''}`;
                      })()}
                    </Typography>
                  </Box>
                ))}
              </Box>
            </CardContent>
          </Card>
        </Grid>
        <Grid item xs={12} md={4}>
          <Card sx={{ height: '100%' }}>
            <CardContent>
              <Typography variant="h6" sx={{ mb: 1, fontWeight: 700 }}>
                Гостевые заявки
              </Typography>
              <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                Управление гостевыми пропусками
              </Typography>
              <Stack
                spacing={1.5}
                component="form"
                onSubmit={guestForm.handleSubmit((data) => {
                  createGuest.mutate({
                    resident_user_id: data.resident_user_id.trim(),
                    guest_full_name: data.guest_full_name.trim(),
                    plate_number: data.plate_number.trim().toUpperCase(),
                    valid_from: new Date(data.valid_from).toISOString(),
                    valid_to: new Date(data.valid_to).toISOString()
                  });
                  guestForm.reset();
                })}
              >
                <TextField
                  label="Житель"
                  size="small"
                  select
                  SelectProps={{ displayEmpty: true }}
                  {...guestForm.register('resident_user_id')}
                  error={!!guestForm.formState.errors.resident_user_id}
                  helperText={residents.length === 0 ? 'Сначала создайте жителя' : guestForm.formState.errors.resident_user_id?.message}
                  disabled={residents.length === 0}
                >
                  <MenuItem value="">
                    Выберите жителя
                  </MenuItem>
                  {residents.map((resident) => (
                    <MenuItem key={resident.id} value={resident.id}>
                      {resident.full_name}
                      {resident.plot_number ? ` · участок ${resident.plot_number}` : ''}
                    </MenuItem>
                  ))}
                </TextField>
                <TextField
                  label="Гость"
                  size="small"
                  {...guestForm.register('guest_full_name')}
                  error={!!guestForm.formState.errors.guest_full_name}
                  helperText={guestForm.formState.errors.guest_full_name?.message}
                />
                <TextField
                  label="Номер авто"
                  size="small"
                  {...guestForm.register('plate_number')}
                  error={!!guestForm.formState.errors.plate_number}
                  helperText={guestForm.formState.errors.plate_number?.message}
                />
                <TextField
                  label="С"
                  size="small"
                  type="datetime-local"
                  InputLabelProps={{ shrink: true }}
                  {...guestForm.register('valid_from')}
                  error={!!guestForm.formState.errors.valid_from}
                  helperText={guestForm.formState.errors.valid_from?.message}
                />
                <TextField
                  label="По"
                  size="small"
                  type="datetime-local"
                  InputLabelProps={{ shrink: true }}
                  {...guestForm.register('valid_to')}
                  error={!!guestForm.formState.errors.valid_to}
                  helperText={guestForm.formState.errors.valid_to?.message}
                />
                <Button type="submit" variant="contained">Создать заявку</Button>
              </Stack>
              <Divider sx={{ my: 2 }} />
              <Box sx={{ maxHeight: 240, overflow: 'auto' }}>
                {guestsQuery.data?.map((guest) => (
                  <Box key={guest.id} sx={{ mb: 1 }}>
                    <Typography variant="body2" sx={{ fontWeight: 600 }}>
                      {guest.guest_full_name} · {guest.status}
                    </Typography>
                    <Typography variant="caption" color="text.secondary">
                      {guest.plate_number} · {guest.valid_from}
                      {(() => {
                        const resident = userById.get(guest.resident_user_id);
                        if (!resident) {
                          return ` · житель: ${guest.resident_user_id}`;
                        }
                        return ` · житель: ${resident.full_name}${resident.plot_number ? ` · участок ${resident.plot_number}` : ''}`;
                      })()}
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
