export type Role = 'admin' | 'guard' | 'resident';

export interface User {
  id: string;
  email: string;
  role: Role;
  full_name: string;
  plot_number?: string;
  created_at: string;
  updated_at: string;
  created_by?: string;
  updated_by?: string;
  deleted_at?: string;
}

export interface Pass {
  id: string;
  owner_user_id: string;
  owner_full_name?: string;
  owner_plot_number?: string;
  plate_number: string;
  vehicle_brand?: string;
  vehicle_color?: string;
  status: string;
  created_at: string;
  updated_at: string;
  created_by?: string;
  updated_by?: string;
  deleted_at?: string;
}

export interface GuestRequest {
  id: string;
  resident_user_id: string;
  guest_full_name: string;
  plate_number: string;
  valid_from: string;
  valid_to: string;
  status: string;
  created_at: string;
  updated_at: string;
  created_by?: string;
  updated_by?: string;
  deleted_at?: string;
}

export interface EntryLog {
  id: string;
  pass_id: string;
  guard_user_id: string;
  action: string;
  action_at: string;
  comment?: string;
}

export interface TokenResponse {
  access_token: string;
  refresh_token: string;
  user: User;
}
