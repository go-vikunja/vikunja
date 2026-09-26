<template>
	<CreateEdit
		v-model:loading="loadingModel"
		:title="title"
		:primary-disabled="team.name === ''"
		@create="createTeam()"
	>
		<FormField
			id="teamName"
			v-model="team.name"
			v-focus
			:label="$t('team.attributes.name')"
			:disabled="createTeamMutation.isPending.value"
			:loading="createTeamMutation.isPending.value"
			:placeholder="$t('team.attributes.namePlaceholder')"
			type="text"
			:error="showError && team.name === '' ? $t('team.attributes.nameRequired') : null"
			@keyup.enter="createTeam"
		/>
		<FormField
			v-if="configStore.public_teams_enabled"
			:label="$t('team.attributes.isPublic')"
		>
			<FancyCheckbox
				v-model="team.is_public"
				:class="{ 'disabled': createTeamMutation.isPending.value }"
				:disabled="createTeamMutation.isPending.value"
			>
				{{ $t('team.attributes.isPublicDescription') }}
			</FancyCheckbox>
		</FormField>
	</CreateEdit>
</template>

<script setup lang="ts">
import {computed, reactive, ref} from 'vue'
import {useI18n} from 'vue-i18n'

import {createTeamDraft, useCreateTeamMutation} from '@/client/queries/teams'

import CreateEdit from '@/components/misc/CreateEdit.vue'
import FancyCheckbox from '@/components/input/FancyCheckbox.vue'
import FormField from '@/components/input/FormField.vue'

import {useTitle} from '@/composables/useTitle'
import {useRouter} from 'vue-router'

import {useConfigStore} from '@/stores/config'

defineOptions({name: 'NewTeam'})

const {t} = useI18n()
const title = computed(() => t('team.create.title'))
useTitle(title)
const router = useRouter()

const createTeamMutation = useCreateTeamMutation()
const team = reactive(createTeamDraft())
const showError = ref(false)
const isSubmitting = ref(false)

const loadingModel = computed({
	get: () => isSubmitting.value || createTeamMutation.isPending.value,
	set(value: boolean) {
		isSubmitting.value = value
	},
})

const configStore = useConfigStore()

async function createTeam() {
	if (team.name === '') {
		showError.value = true
		return
	}
	showError.value = false

	if (isSubmitting.value) {
		return
	}

	isSubmitting.value = true

	let response
	try {
		response = await createTeamMutation.mutateAsync(team)
	} catch {
		return
	} finally {
		isSubmitting.value = false
	}
	router.push({
		name: 'teams.edit',
		params: { id: response.id },
	})
}
</script>
