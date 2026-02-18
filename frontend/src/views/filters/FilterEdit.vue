<template>
	<CreateEdit
		v-model:loading="loadingModel"
		:title="$t('filters.edit.title')"
		primary-icon=""
		:primary-label="$t('misc.save')"
		:tertiary="$t('misc.delete')"
		@primary="handleSave"
		@tertiary="$router.push({ name: 'filter.settings.delete', params: { id: projectId } })"
	>
		<form @submit.prevent="handleSave()">
			<FormField
				id="Title"
				v-model="filter.title"
				v-focus
				:label="$t('filters.attributes.title')"
				:class="{ 'is-danger': !titleValid }"
				:disabled="filterService.loading"
				:placeholder="$t('filters.attributes.titlePlaceholder')"
				type="text"
				:error="titleValid ? null : $t('filters.create.titleRequired')"
				@focusout="validateTitleField"
			/>
			<FormField :label="$t('filters.attributes.description')">
				<Editor
					id="description"
					v-model="filter.description"
					:class="{ 'disabled': filterService.loading}"
					:disabled="filterService.loading"
					:placeholder="$t('filters.attributes.descriptionPlaceholder')"
				/>
			</FormField>
			<FormField :label="$t('filters.title')">
				<Filters
					v-model="filters"
					:class="{ 'disabled': filterService.loading}"
					:disabled="filterService.loading"
					class="has-no-shadow has-no-border"
					:has-footer="false"
					:change-immediately="true"
				/>
			</FormField>
		</form>
	</CreateEdit>
</template>

<script setup lang="ts">
import {computed, ref} from 'vue'

import Editor from '@/components/input/AsyncEditor'
import CreateEdit from '@/components/misc/CreateEdit.vue'
import FormField from '@/components/input/FormField.vue'
import Filters from '@/components/project/partials/Filters.vue'

import {useSavedFilter} from '@/services/savedFilter'

import type {IProject} from '@/modelTypes/IProject'

const props = defineProps<{
	projectId: IProject['id'],
}>()

const {
	saveFilterWithValidation,
	filter,
	filters,
	filterService,
	titleValid,
	validateTitleField,
} = useSavedFilter(() => props.projectId)

const isSubmitting = ref(false)

const loadingModel = computed({
	get: () => isSubmitting.value || filterService.loading,
	set(value: boolean) {
		isSubmitting.value = value
	},
})

async function handleSave() {
	if (isSubmitting.value) {
		return
	}

	isSubmitting.value = true

	try {
		await saveFilterWithValidation()
	} finally {
		isSubmitting.value = false
	}
}
</script>
