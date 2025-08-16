import { defineStore } from 'pinia'
import type { User } from '@/types'

const API = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api'

export const useUsersStore = defineStore('users', {
	state: () => ({
		users: [] as User[],
		selected: null as User | null,
		loading: false,
		error: '' as string | ''
	}),
	actions: {
		async fetchUsers() {
			this.loading = true
			this.error = ''
			try {
				const res = await fetch(`${API}/users`)
				if (!res.ok) throw new Error(`Failed: ${res.status}`)
				this.users = await res.json()
			} catch (e: any) {
				this.error = e.message
			} finally {
				this.loading = false
			}
		},
		select(u: User | null) {
			this.selected = u ? { ...u } : null
		},
		async createUser(u: User) {
			const res = await fetch(`${API}/users`, {
				method: 'POST', headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(u)
			})
			if (!res.ok) throw new Error(await res.text())
			const created: User = await res.json()
			this.users.unshift(created)
		},
		async updateUser(u: User) {
			if (!u.id) throw new Error('Missing id')
			const res = await fetch(`${API}/users/${u.id}`, {
				method: 'PUT', headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ name: u.name, email: u.email })
			})
			if (!res.ok) throw new Error(await res.text())
			const updated: User = await res.json()
			const idx = this.users.findIndex(x => x.id === updated.id)
			if (idx !== -1) this.users[idx] = updated
		},
		async deleteUser(id: number) {
			const res = await fetch(`${API}/users/${id}`, { method: 'DELETE' })
			if (!res.ok && res.status !== 204) throw new Error(await res.text())
			this.users = this.users.filter(u => u.id !== id)
		}
	}
})