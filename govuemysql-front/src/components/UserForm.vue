<template>
	<form @submit.prevent="onSubmit" class="grid gap-3 p-4 border rounded">
		<div class="grid gap-1">
			<label>Name</label>
			<input v-model="local.name" required placeholder="Jane Doe" class="border rounded p-2" />
		</div>
		<div class="grid gap-1">
			<label>Email</label>
			<input v-model="local.email" required type="email" placeholder="jane@example.com"
				class="border rounded p-2" />
		</div>
		<div class="flex gap-2 mt-2">
			<button type="submit" class="px-4 py-2 border rounded">{{ local.id ? 'Update' : 'Create' }}</button>
			<button type="button" class="px-4 py-2 border rounded" @click="$emit('cancel')">Cancel</button>
		</div>
	</form>
</template>

<script setup lang="ts">
import { reactive, watch } from 'vue'
import type { User } from '@/types'

const props = defineProps<{ user: User | null }>()
const emit = defineEmits<{ (e: 'save', user: User): void, (e: 'cancel'): void }>()

const local = reactive<User>({ name: '', email: '' })

watch(() => props.user, (u) => {
	if (u) Object.assign(local, u)
	else Object.assign(local, { id: undefined, name: '', email: '' })
}, { immediate: true })

async function onSubmit() {
	emit('save', { ...local })
}
</script>