<template>
	<section class="grid gap-4">
		<div class="flex items-center justify-between gap-2">
			<input v-model="q" placeholder="Search by name/email" class="border rounded p-2 flex-1" />
			<button class="px-4 py-2 border rounded" @click="openCreate">+ New</button>
		</div>

		<div v-if="store.error" class="text-red-600">{{ store.error }}</div>

		<table class="w-full border-collapse">
			<thead>
				<tr>
					<th class="border p-2 text-left">ID</th>
					<th class="border p-2 text-left">Name</th>
					<th class="border p-2 text-left">Email</th>
					<th class="border p-2 text-left">Created</th>
					<th class="border p-2">Actions</th>
				</tr>
			</thead>
			<tbody>
				<tr v-for="u in filtered" :key="u.id">
					<td class="border p-2">{{ u.id }}</td>
					<td class="border p-2">{{ u.name }}</td>
					<td class="border p-2">{{ u.email }}</td>
					<td class="border p-2">{{ new Date(u.created_at!).toLocaleString() }}</td>
					<td class="border p-2 text-center">
						<button class="px-2 py-1 border rounded mr-2" @click="edit(u)">Edit</button>
						<button class="px-2 py-1 border rounded" @click="remove(u)">Delete</button>
					</td>
				</tr>
			</tbody>
		</table>

		<UserForm v-if="showForm" :user="store.selected" @save="save" @cancel="closeForm" />
	</section>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useUsersStore } from '@/stores/users'
import type { User } from '@/types'
import UserForm from '@/components/UserForm.vue'

const store = useUsersStore()
const q = ref('')
const showForm = ref(false)

onMounted(() => store.fetchUsers())

const filtered = computed(() =>
	store.users.filter(u =>
		[u.name, u.email].some(v => v?.toLowerCase().includes(q.value.toLowerCase()))
	)
)

function openCreate() { store.select(null); showForm.value = true }
function edit(u: User) { store.select(u); showForm.value = true }
function closeForm() { showForm.value = false }

async function save(u: User) {
	if (u.id) await store.updateUser(u)
	else await store.createUser(u)
	closeForm()
}

async function remove(u: User) {
	if (u.id && confirm(`Delete user #${u.id}?`)) await store.deleteUser(u.id)
}
</script>