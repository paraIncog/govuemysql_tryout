<template>
  <form class="card" @submit.prevent="onSubmit">
    <h2>Create User</h2>

    <label>
      Name
      <input v-model.trim="form.name" required placeholder="Alice Johnson" />
    </label>

    <label>
      Email
      <input v-model.trim="form.email" type="email" required placeholder="alice@example.com" />
    </label>

    <button :disabled="busy">Create</button>
    <p v-if="err" class="error">{{ err }}</p>
  </form>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { createUser } from "@/stores/users";

const emit = defineEmits<{ (e: "save", payload: { name: string; email: string }): void }>();

const form = ref({ name: "", email: "" });
const err = ref("");
const busy = ref(false);

async function onSubmit() {
  try {
    busy.value = true;
    err.value = "";
    await createUser({ name: form.value.name, email: form.value.email });
    emit("save", { name: form.value.name, email: form.value.email });
    form.value = { name: "", email: "" };
  } catch (e: any) {
    err.value = e.message || "Create failed";
  } finally {
    busy.value = false;
  }
}
</script>

<style scoped>
.card {
  background: #fff;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  padding: 1rem;
}
label { display: block; margin-bottom: .8rem; }
input {
  width: 100%;
  padding: .5rem .6rem;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  margin-top: .25rem;
}
button {
  padding: .5rem .8rem;
  border: 1px solid #d1d5db;
  background: #f9fafb;
  border-radius: 6px;
  cursor: pointer;
}
.error { color: #b91c1c; margin-top: .5rem; }
</style>
