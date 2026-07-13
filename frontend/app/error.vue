<script setup lang="ts">
import type { NuxtError } from 'nuxt/app';


const props = defineProps<{ error: NuxtError }>()
</script>

<template>
    <h1>{{ error.status }}</h1>
    
    <NuxtLink to="/">Go back home</NuxtLink>
</template>
