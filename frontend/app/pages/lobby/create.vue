<script setup lang="ts">
import type { FormSubmitEvent } from '@nuxt/ui'
import { Time, today, getLocalTimeZone, toCalendarDateTime } from '@internationalized/date'
import { lobbySchema, type LobbySchema } from '~/schemas/lobby'

const { data: countries } = await useCountries()

const minDate = today(getLocalTimeZone())

let nextId = 0
const state = reactive({
  title: '',
  description: '',
  date: minDate,
  time: new Time(12, 0, 0),
  countries: [{ id: nextId++, tag: '', players: 1 }],
  gameMode: ''
})

function addCountry() {
  state.countries.push({ id: nextId++, tag: '', players: 1 })
}

function removeCountry(id: number) {
  state.countries = state.countries.filter(c => c.id !== id)
}

const taken = computed(() => new Set(state.countries.map(c => c.tag).filter(Boolean)))

function itemsFor(tag: string) {
  return countries.value.filter(c => c.value === tag || !taken.value.has(c.value))
}

async function onSubmit(event: FormSubmitEvent<LobbySchema>) {
  const { date, time } = event.data

  await $fetch('/api/lobby', {
    baseURL: 'http://localhost:8000',
    method: 'POST',
    body: {
      LobbyName: event.data.title,
      Description: event.data.description,
      GameMode: event.data.gameMode,
      StartsAt: toCalendarDateTime(date, time)
        .toDate(getLocalTimeZone())
        .toISOString(),
      countries: event.data.countries.map(c => ({
        country_tag: c.tag,
        players: c.players,
      })),
    },
  })
}

const lobbyDate = useTemplateRef('lobbyDate')

const gameModes = ref(['Vanilla', 'modded', 'rp'])
</script>

<template>
  <UForm :schema="lobbySchema" :state="state" class="space-y-6" @submit="onSubmit">
    <h1 class="text-xl font-semibold">Create Lobby</h1>

    <div class="space-y-4">
      <UFormField label="Title" name="title" required>
        <UInput v-model="state.title" class="w-full" />
      </UFormField>

      <UFormField label="Description" name="description">
        <UTextarea v-model="state.description" class="w-full" />
      </UFormField>

      
      <div class="flex gap-4">
        <UFormField label="Gamemode" name="gamemode" required>
            <USelect v-model="state.gameMode" :items="gameModes" placeholder="Select a Gamemode" />
        </UFormField>

        <UFormField label="Date" name="date" required>
          <UInputDate ref="lobbyDate" v-model="state.date" :min-value="minDate">
            <template #trailing>
              <UPopover :reference="lobbyDate?.inputsRef[3]?.$el">
                <UButton
                  color="neutral"
                  variant="link"
                  size="sm"
                  icon="i-lucide-calendar"
                  aria-label="Select a date"
                  class="px-0"
                />
                <template #content>
                  <UCalendar v-model="state.date" :min-value="minDate" class="p-2" />
                </template>
              </UPopover>
            </template>
          </UInputDate>
        </UFormField>

        <UFormField label="Time" name="time" required>
          <UInputTime v-model="state.time" :hour-cycle="24" />
        </UFormField>
      </div>
    </div>

    <fieldset class="space-y-3">
      <legend class="text-sm font-medium">Countries</legend>

      <LobbyCountryRow
        v-for="(country, i) in state.countries"
        :key="country.id"
        v-model:tag="country.tag"
        v-model:players="country.players"
        :index="i"
        :items="itemsFor(country.tag)"
        :removable="state.countries.length > 1"
        @remove="removeCountry(country.id)"
      />

      <UButton
        type="button"
        icon="i-lucide-plus"
        variant="soft"
        :disabled="state.countries.length >= countries.length"
        @click="addCountry"
      >
        Add Country
      </UButton>
    </fieldset>

    <UButton type="submit">Create Lobby</UButton>
  </UForm>
</template>