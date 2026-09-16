<template>
	<CreateEdit
		v-model:loading="loadingModel"
		:title="$t('filters.edit.title')"
		primary-icon=""
		:primary-label="$t('misc.save')"
		:primary-disabled="Boolean(loadError) || loadingModel"
		:tertiary="$t('misc.delete')"
		@primary="save"
		@tertiary="$router.push({ name: 'filter.settings.delete', params: { projectId } })"
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
				:disabled="loadingModel"
				:placeholder="$t('filters.attributes.titlePlaceholder')"
				type="text"
				:error="titleValid ? null : $t('filters.create.titleRequired')"
				@focusout="markTitleTouched"
			/>
			<FormField :label="$t('filters.attributes.description')">
				<Editor
					id="description"
					v-model="filter.description"
					:class="{ 'disabled': loadingModel}"
					:disabled="loadingModel"
					:placeholder="$t('filters.attributes.descriptionPlaceholder')"
				/>
			</FormField>
			<FormField :label="$t('filters.title')">
				<Filters
					v-model="filter.filters"
					:class="{ 'disabled': loadingModel}"
					:disabled="loadingModel"
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
import {computed, nextTick, ref} from 'vue'
import {useRouter} from 'vue-router'

import Editor from '@/components/input/AsyncEditor'
import CreateEdit from '@/components/misc/CreateEdit.vue'
import FormField from '@/components/input/FormField.vue'
import Filters from '@/components/project/partials/Filters.vue'
import ErrorMessage from '@/components/misc/Error.vue'

import {useIsAlive} from '@/composables/useIsAlive'
import {useSavedFilter} from '@/composables/useSavedFilter'
import {getSavedFilterIdFromProjectId} from '@/client/queries/projects'
import {useUpdateSavedFilterMutation} from '@/client/queries/savedFilters'

const props = defineProps<{
	projectId: number,
}>()

const router = useRouter()

const alive = useIsAlive()

const {
	validate,
	filter,
	isLoading,
	error: loadError,
	titleValid,
	markTitleTouched,
} = useSavedFilter(() => props.projectId)
const updateMutation = useUpdateSavedFilterMutation(({id}) =>
	alive.value && getSavedFilterIdFromProjectId(props.projectId) === id,
)

// CreateEdit latches loading on click; the prop must toggle back on early return.
const isSubmitting = ref(false)
const loadingModel = computed({
	get: () => isLoading.value || isSubmitting.value,
	set(value: boolean) {
		isSubmitting.value = value
	},
})

async function save() {
	// The hidden submit button bypasses CreateEdit's own click latch.
	if (isSubmitting.value) {
		return
	}

	const id = props.projectId
	isSubmitting.value = true
	let saved
	try {
		const payload = validate()
		if (!payload) {
			await nextTick()
			return
		}
		saved = await updateMutation.mutateAsync({id: filter.value.id, ...payload}).catch(() => undefined)
	} finally {
		isSubmitting.value = false
	}
	// The route param can change on this same instance, so a stale save must not navigate.
	if (!alive.value || props.projectId !== id) {
		return
	}
	if (saved) {
		router.back()
	}
}
</script>
