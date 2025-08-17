<template>
  <form @submit.prevent="onSubmit" class="card">
    <label>
      Name
      <input v-model="form.name" required />
    </label>
    <label>
      Email
      <input v-model="form.email" type="email" required />
    </label>
    <button :disabled="busy">Create</button>
    <p v-if="err" class="error">{{ err }}</p>
  </form>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { createUser } from "@/stores/users";

const emit = defineEmits<{ (e: "saved"): void }>();

const form = ref({ name: "", email: "" });
const busy = ref(false);
const err = ref("");

async function onSubmit() {
  try {
    busy.value = true;
    err.value = "";
    await createUser({ ...form.value });
    form.value = { name: "", email: "" };
    emit("saved");
  } catch (e: any) {
    err.value = e.message;
  } finally {
    busy.value = false;
  }
}
</script>

<style scoped>
.error { color:red; margin-top:.5rem; }
</style>
