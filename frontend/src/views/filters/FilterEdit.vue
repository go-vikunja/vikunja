<template>
	<CreateEdit
		:loading="isBusy"
		:title="$t('filters.edit.title')"
		primary-icon=""
		:primary-label="$t('misc.save')"
		:primary-disabled="Boolean(loadError) || isBusy"
		:tertiary="$t('misc.delete')"
		@update:loading="isSubmitting = $event"
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
				:disabled="isBusy"
				:placeholder="$t('filters.attributes.titlePlaceholder')"
				type="text"
				:error="titleValid ? null : $t('filters.create.titleRequired')"
				@focusout="markTitleTouched"
			/>
			<FormField :label="$t('filters.attributes.description')">
				<Editor
					id="description"
					v-model="filter.description"
					:class="{ 'disabled': isBusy}"
					:disabled="isBusy"
					:placeholder="$t('filters.attributes.descriptionPlaceholder')"
				/>
			</FormField>
			<FormField :label="$t('filters.title')">
				<Filters
					v-model="filters"
					:class="{ 'disabled': isBusy}"
					:disabled="isBusy"
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
import {computed, onUnmounted, ref} from 'vue'
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

// useMounted() never resets on unmount.
const alive = ref(true)
onUnmounted(() => {
	alive.value = false
})

const {
	submit,
	filter,
	filters,
	isLoading,
	error: loadError,
	titleValid,
	markTitleTouched,
} = useSavedFilter(() => props.projectId)

// CreateEdit latches loading on click; the prop must toggle back on early return.
const isSubmitting = ref(false)
const isBusy = computed(() => isLoading.value || isSubmitting.value)

async function save() {
	const id = props.projectId
	isSubmitting.value = true
	let saved
	try {
		saved = await submit()
	} finally {
		isSubmitting.value = false
	}
	// The route param can change on this same instance, so a stale save must not navigate.
	if (!alive.value || props.projectId !== id) {
		return
	}
	if (saved) {
		success({message: t('filters.edit.success')})
		router.back()
	}
}
</script>
