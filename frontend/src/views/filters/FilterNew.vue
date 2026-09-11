<template>
	<Modal
		variant="hint-modal"
		@close="$router.back()"
	>
		<Card
			class="has-no-shadow"
			:title="$t('filters.create.title')"
		>
			<p>
				{{ $t('filters.create.description') }}
			</p>
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
				@focusout="markTitleTouched"
			/>
			<FormField :label="$t('filters.attributes.description')">
				<Editor
					id="description"
					:key="filter.id"
					v-model="filter.description"
					:class="{ 'disabled': isLoading}"
					:disabled="isLoading"
					:placeholder="$t('filters.attributes.descriptionPlaceholder')"
				/>
			</FormField>
			<FormField :label="$t('filters.title')">
				<Filters
					v-model="filter.filters"
					:class="{ 'disabled': isLoading}"
					:disabled="isLoading"
					class="has-no-shadow has-no-border"
					:has-footer="false"
					:change-immediately="true"
				/>
			</FormField>

			<template #footer>
				<XButton
					:loading="isLoading"
					:disabled="isLoading || !titleValid"
					class="is-fullwidth"
					@click="create()"
				>
					{{ $t('filters.create.action') }}
				</XButton>
			</template>
		</Card>
	</Modal>
</template>

<script setup lang="ts">
import {onUnmounted, ref} from 'vue'
import {useRouter} from 'vue-router'

import Editor from '@/components/input/AsyncEditor'
import FormField from '@/components/input/FormField.vue'
import Filters from '@/components/project/partials/Filters.vue'

import {getProjectIdFromSavedFilterId} from '@/client/queries/projects'
import {useSavedFilter} from '@/composables/useSavedFilter'

const router = useRouter()

// useMounted() never resets on unmount.
const alive = ref(true)
onUnmounted(() => {
	alive.value = false
})

const {
	filter,
	submit,
	isLoading,
	titleValid,
	markTitleTouched,
} = useSavedFilter()

async function create() {
	const created = await submit()
	if (!alive.value) {
		return
	}
	if (created) {
		await router.push({
			name: 'project.index',
			params: {projectId: getProjectIdFromSavedFilterId(created.id)},
		})
	}
}
</script>
