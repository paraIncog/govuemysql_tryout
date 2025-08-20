import axios from "axios";
import type { User } from "../types";

const base =
  import.meta.env.VITE_API_BASE ??
  import.meta.env.VITE_API_BASE_URL ??
  "/api";


export const api = axios.create({ baseURL: base });

export async function fetchUsers(): Promise<User[]> {
  const { data } = await api.get<User[]>("/users");
  return Array.isArray(data) ? data : [];
}

export async function createUser(payload: Pick<User, "name" | "email">): Promise<User> {
  try {
    const { data } = await api.post<User>("/users", payload);
    return data;
  } catch (e: any) {
    throw new Error(e?.response?.data?.error ?? "Create failed");
  }
}

export async function updateUser(id: number, payload: Pick<User, "name" | "email">): Promise<User> {
  try {
    const { data } = await api.put<User>(`/users/${id}`, payload);
    return data;
  } catch (e: any) {
    throw new Error(e?.response?.data?.error ?? "Update failed");
  }
}

export async function deleteUser(id: number): Promise<void> {
  try {
    await api.delete(`/users/${id}`);
  } catch (e: any) {
    throw new Error(e?.response?.data?.error ?? "Delete failed");
  }
}
