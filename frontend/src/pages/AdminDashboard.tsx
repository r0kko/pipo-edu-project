import { Card, CardActionArea, CardContent, Grid, Stack, Typography } from '@mui/material';
import { useNavigate } from 'react-router-dom';
import type { ReactNode } from 'react';
import People from '@mui/icons-material/People';
import DirectionsCar from '@mui/icons-material/DirectionsCar';
import Badge from '@mui/icons-material/Badge';
import { Layout } from '../components/Layout';

interface AdminSectionCardProps {
  title: string;
  description: string;
  onClick: () => void;
  icon: ReactNode;
}

function AdminSectionCard({ title, description, onClick, icon }: AdminSectionCardProps) {
  return (
    <Card sx={{ height: '100%' }}>
      <CardActionArea sx={{ height: '100%' }} onClick={onClick}>
        <CardContent sx={{ height: '100%', display: 'flex', flexDirection: 'column', gap: 1.5 }}>
          <Stack direction="row" alignItems="center" spacing={1.5}>
            {icon}
            <Typography variant="h6" sx={{ fontWeight: 700 }}>
              {title}
            </Typography>
          </Stack>
          <Typography variant="body2" color="text.secondary">
            {description}
          </Typography>
        </CardContent>
      </CardActionArea>
    </Card>
  );
}

export default function AdminDashboard() {
  const navigate = useNavigate();

  return (
    <Layout title="Администрирование">
      <Typography variant="body1" sx={{ mb: 3 }}>
        Выберите раздел управления:
      </Typography>

      <Grid container spacing={3}>
        <Grid item xs={12} md={4}>
          <AdminSectionCard
            title="Пользователи"
            description="Добавление, редактирование, блокировка и удаление пользователей."
            onClick={() => navigate('/admin/users')}
            icon={<People color="primary" />}
          />
        </Grid>
        <Grid item xs={12} md={4}>
          <AdminSectionCard
            title="Пропуска"
            description="Управление пропусками жителей в отдельном рабочем разделе."
            onClick={() => navigate('/admin/passes')}
            icon={<DirectionsCar color="primary" />}
          />
        </Grid>
        <Grid item xs={12} md={4}>
          <AdminSectionCard
            title="Гостевые заявки"
            description="Создание, изменение и контроль гостевых заявок."
            onClick={() => navigate('/admin/guests')}
            icon={<Badge color="primary" />}
          />
        </Grid>
      </Grid>
    </Layout>
  );
}
