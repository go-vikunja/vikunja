<template>
	<CreateEdit
		:loading="isLoading"
		:title="$t('filters.edit.title')"
		primary-icon=""
		:primary-label="$t('misc.save')"
		:primary-disabled="Boolean(loadError) || isLoading"
		:tertiary="$t('misc.delete')"
		@primary="save"
		@tertiary="$router.push({ name: 'filter.settings.delete', params: { id: projectId } })"
	>
		<ErrorMessage v-if="loadError" />
		<form
			v-else
			@submit.prevent="save()"
		>
			<FormField
				id="Title"
				v-model="filter.title"
				v-focus
				:label="$t('filters.attributes.title')"
				:class="{ 'is-danger': !titleValid }"
				:disabled="isLoading"
				:placeholder="$t('filters.attributes.titlePlaceholder')"
				type="text"
				:error="titleValid ? null : $t('filters.create.titleRequired')"
				@focusout="validateTitleField"
			/>
			<FormField :label="$t('filters.attributes.description')">
				<Editor
					id="description"
					v-model="filter.description"
					:class="{ 'disabled': isLoading}"
					:disabled="isLoading"
					:placeholder="$t('filters.attributes.descriptionPlaceholder')"
				/>
			</FormField>
			<FormField :label="$t('filters.title')">
				<Filters
					v-model="filters"
					:class="{ 'disabled': isLoading}"
					:disabled="isLoading"
					class="has-no-shadow has-no-border"
					:has-footer="false"
					:change-immediately="true"
				/>
			</FormField>
			<button
				type="submit"
				class="is-hidden"
				tabindex="-1"
			/>
		</form>
	</CreateEdit>
</template>

<script setup lang="ts">
import {useMounted} from '@vueuse/core'
import {useI18n} from 'vue-i18n'
import {useRouter} from 'vue-router'

import Editor from '@/components/input/AsyncEditor'
import CreateEdit from '@/components/misc/CreateEdit.vue'
import FormField from '@/components/input/FormField.vue'
import Filters from '@/components/project/partials/Filters.vue'
import ErrorMessage from '@/components/misc/Error.vue'

import {useSavedFilter} from '@/composables/useSavedFilter'
import {success} from '@/message'

const props = defineProps<{
	projectId: number,
}>()

const {t} = useI18n({useScope: 'global'})
const router = useRouter()
const isMounted = useMounted()

const {
	saveFilter,
	filter,
	filters,
	isLoading,
	error: loadError,
	titleValid,
	validateTitleField,
} = useSavedFilter(() => props.projectId)

async function save() {
	const saved = await saveFilter()
	if (!isMounted.value) {
		return
	}
	if (saved) {
		success({message: t('filters.edit.success')})
		router.back()
	}
}
</script>
