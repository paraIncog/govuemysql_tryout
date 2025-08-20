<template>
  <main>
    <h1>Users</h1>
    <UserForm @saved="load" />

    <section class="card">
      <h2>All Users</h2>
      <div v-if="loading">Loading…</div>
      <div v-else-if="error">{{ error }}</div>
      <ul v-else>
        <li v-for="u in users" :key="u.id">{{ u.name }} – {{ u.email }}</li>
      </ul>

      <table class="users" v-if="users.length">
        <thead>
          <tr>
            <th>ID</th>
            <th>Name</th>
            <th>Email</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="u in users" :key="u.id">
            <td>{{ u.id }}</td>
            <td>
              <input v-if="editId === u.id" v-model="editForm.name" />
              <span v-else>{{ u.name }}</span>
            </td>
            <td>
              <input v-if="editId === u.id" v-model="editForm.email" />
              <span v-else>{{ u.email }}</span>
            </td>
            <td>
              <template v-if="editId === u.id">
                <button @click="saveEdit(u.id)">Save</button>
                <button @click="cancelEdit">Cancel</button>
              </template>
              <template v-else>
                <button @click="startEdit(u)">Edit</button>
                <button @click="remove(u.id)">Delete</button>
              </template>
            </td>
          </tr>
        </tbody>
      </table>
      <p v-else>No users yet.</p>
    </section>
  </main>
</template>

<script setup lang="ts">
import { ref, onMounted } from "vue";
import UserForm from "../components/UserForm.vue";
import { fetchUsers, updateUser, deleteUser } from "@/stores/users";
import type { User } from "../types";

const users = ref<User[]>([]);
const loading = ref(true);
const error = ref<string | null>(null);

const errList = ref("");
const editId = ref<number | null>(null);
const editForm = ref({ name: "", email: "" });

async function load() {
  try {
    errList.value = "";
    users.value = await fetchUsers();
  } catch (e: any) {
    errList.value = e.message;
  }
}

function startEdit(u: User) {
  editId.value = u.id;
  editForm.value = { name: u.name, email: u.email };
}
function cancelEdit() { editId.value = null; }

async function saveEdit(id: number) {
  try {
    await updateUser(id, editForm.value);
    editId.value = null;
    await load();
  } catch (e: any) { alert(e.message); }
}

async function remove(id: number) {
  try {
    await deleteUser(id);
    users.value = users.value.filter(u => u.id !== id);
  } catch (e: any) {
    error.value = e.message ?? "Delete failed";
  }
}

onMounted(async () => {
  try {
    users.value = await fetchUsers();
  } catch (e: any) {
    error.value = e.message ?? "Failed to load users";
    users.value = [];                   // <-- stay an array
  } finally {
    loading.value = false;
  }
});
</script>

<style scoped>
.error {
  color: red;
}

.users {
  width: 100%;
  border-collapse: collapse;
}

.users th,
.users td {
  border: 1px solid #ccc;
  padding: .5rem;
}
</style>
