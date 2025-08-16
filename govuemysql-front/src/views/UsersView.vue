<template>
  <main>
    <h1>Users</h1>

    <UserForm :key="formKey" :user="{ id: 0, name: '', email: '' }" @save="save" />

    <section class="card">
      <h2>All Users</h2>

      <p v-if="errList" class="error">{{ errList }}</p>

      <table class="users" v-if="users.length">
        <thead>
          <tr>
            <th style="width: 60px">ID</th>
            <th>Name</th>
            <th>Email</th>
            <th style="width: 220px">Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="u in users" :key="u.id">
            <td>{{ u.id }}</td>
            <td>
              <template v-if="editId === u.id">
                <input v-model.trim="editForm.name" />
              </template>
              <template v-else>{{ u.name }}</template>
            </td>
            <td>
              <template v-if="editId === u.id">
                <input v-model.trim="editForm.email" type="email" />
              </template>
              <template v-else>{{ u.email }}</template>
            </td>
            <td>
              <template v-if="editId === u.id">
                <button @click="saveEdit(u.id)" :disabled="busyEdit">Save</button>
                <button @click="cancelEdit" type="button">Cancel</button>
              </template>
              <template v-else>
                <button @click="startEdit(u)">Edit</button>
                <button @click="remove(u.id)" :disabled="busyDelete === u.id">Delete</button>
              </template>
            </td>
          </tr>
        </tbody>
      </table>

      <p v-else>No users yet—add one above.</p>
    </section>
  </main>
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue";
import UserForm from "../components/UserForm.vue";
import { fetchUsers, createUser, updateUser, deleteUser } from "@/stores/users"; // Adjust the import path as necessary
import type { User } from "@/types"; // Adjust the import path as necessary

const users = ref<User[]>([]);
const errList = ref("");
const busyEdit = ref(false);
const busyDelete = ref<number | null>(null);
const editId = ref<number | null>(null);
const editForm = ref({ name: "", email: "" });
const formKey = ref(1); // used to reset the child form

async function load() {
  try {
    errList.value = "";
    users.value = await fetchUsers();
  } catch (e: any) {
    errList.value = e.message || "Failed to fetch users";
  }
}

async function save(payload: { name: string; email: string }) {
  try {
    await createUser(payload);
    formKey.value++; // reset form
    await load();
  } catch (e: any) {
    // The child form displays its own errors via emitted failure; this is a fallback
    alert(e.message || "Create failed");
  }
}

function startEdit(u: User) {
  editId.value = u.id;
  editForm.value = { name: u.name, email: u.email };
}
function cancelEdit() {
  editId.value = null;
}

async function saveEdit(id: number) {
  try {
    busyEdit.value = true;
    await updateUser(id, { name: editForm.value.name, email: editForm.value.email });
    editId.value = null;
    await load();
  } catch (e: any) {
    alert(e.message || "Update failed");
  } finally {
    busyEdit.value = false;
  }
}

async function remove(id: number) {
  if (!confirm("Delete this user?")) return;
  try {
    busyDelete.value = id;
    await deleteUser(id);
    await load();
  } catch (e: any) {
    alert(e.message || "Delete failed");
  } finally {
    busyDelete.value = null;
  }
}

onMounted(load);
</script>

<style scoped>
h1 { margin-bottom: 1rem; }
.card {
  background: #fff;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  padding: 1rem;
  margin-top: 1.25rem;
}
.users { width: 100%; border-collapse: collapse; }
.users th, .users td { border-top: 1px solid #eee; padding: .6rem; text-align: left; }
button {
  padding: .45rem .75rem;
  border: 1px solid #d1d5db;
  background: #f9fafb;
  border-radius: 6px;
  cursor: pointer;
  margin-right: .5rem;
}
.error { color: #b91c1c; margin-bottom: .5rem; }
input {
  width: 100%;
  padding: .45rem .6rem;
  border: 1px solid #d1d5db;
  border-radius: 6px;
}
</style>
