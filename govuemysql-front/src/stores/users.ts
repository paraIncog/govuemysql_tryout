import axios from "axios";
import type { User } from "../types";

const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE || "http://localhost:8080/api",
});

export async function fetchUsers(): Promise<User[]> {
  try {
    const { data } = await api.get<User[]>("/users");
    return data;
  } catch (e: any) {
    const msg = e?.response?.data?.error ?? "Failed to fetch users";
    throw new Error(msg);
  }
}

export async function createUser(payload: Pick<User, "name" | "email">): Promise<User> {
  try {
    const { data } = await api.post<User>("/users", payload);
    return data;
  } catch (e: any) {
    const msg = e?.response?.data?.error ?? "Create failed";
    throw new Error(msg);
  }
}

export async function updateUser(
  id: number,
  payload: Pick<User, "name" | "email">
): Promise<User> {
  try {
    const { data } = await api.put<User>(`/users/${id}`, payload);
    return data;
  } catch (e: any) {
    const msg = e?.response?.data?.error ?? "Update failed";
    throw new Error(msg);
  }
}

export async function deleteUser(id: number): Promise<void> {
  try {
    await api.delete(`/users/${id}`);
  } catch (e: any) {
    const msg = e?.response?.data?.error ?? "Delete failed";
    throw new Error(msg);
  }
}
