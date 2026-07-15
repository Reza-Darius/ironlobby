<script setup lang="ts">
const tag = defineModel<string>('tag', { required: true })
const players = defineModel<number>('players', { required: true })

interface Props {
  index: number
  items: { label: string; value: string }[]
  removable: boolean
}

defineProps<Props>()

defineEmits<{ remove: [] }>()
</script>

<template>
  <div class="flex items-start gap-2">
    <UFormField :label="`Country ${index + 1}`" :name="`countries.${index}.tag`" class="flex-1">
      <USelect v-model="tag" :items="items" placeholder="Select a Country" class="w-full" />
    </UFormField>

    <UFormField label="Players" :name="`countries.${index}.players`">
      <UInputNumber v-model="players" :min="1" orientation="vertical" />
    </UFormField>

    <UButton
      type="button"
      icon="i-lucide-x"
      color="neutral"
      variant="ghost"
      class="mt-6"
      :disabled="!removable"
      :aria-label="`Remove country ${index + 1}`"
      @click="$emit('remove')"
    />
  </div>
</template>
