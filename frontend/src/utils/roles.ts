import { Role } from '../api/types';

export const ROLE_LABEL: Record<Role, string> = {
  admin: 'Администратор',
  guard: 'Охрана',
  resident: 'Житель'
};

export const ROLE_OPTIONS: Array<{ value: Role; label: string }> = [
  { value: 'admin', label: ROLE_LABEL.admin },
  { value: 'guard', label: ROLE_LABEL.guard },
  { value: 'resident', label: ROLE_LABEL.resident }
];

export function roleLabel(role?: Role | null): string {
  if (!role) return '';
  return ROLE_LABEL[role] ?? role;
}
