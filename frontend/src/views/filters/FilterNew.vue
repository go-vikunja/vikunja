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
			<div class="field">
				<label
					class="label"
					for="title"
				>{{ $t('filters.attributes.title') }}</label>
				<div class="control">
					<input
						id="Title"
						v-model="filter.title"
						v-focus
						:class="{ 'disabled': filterService.loading, 'is-danger': !titleValid }"
						:disabled="filterService.loading || undefined"
						class="input"
						:placeholder="$t('filters.attributes.titlePlaceholder')"
						type="text"
						@focusout="validateTitleField"
					>
				</div>
				<p
					v-if="!titleValid"
					class="help is-danger"
				>
					{{ $t('filters.create.titleRequired') }}
				</p>
			</div>
			<div class="field">
				<label
					class="label"
					for="description"
				>{{ $t('filters.attributes.description') }}</label>
				<div class="control">
					<Editor
						id="description"
						:key="filter.id"
						v-model="filter.description"
						:class="{ 'disabled': filterService.loading}"
						:disabled="filterService.loading"
						:placeholder="$t('filters.attributes.descriptionPlaceholder')"
					/>
				</div>
			</div>
			<div class="field">
				<label
					class="label"
					for="filters"
				>{{ $t('filters.title') }}</label>
				<div class="control">
					<Filters
						v-model="filters"
						:class="{ 'disabled': filterService.loading}"
						:disabled="filterService.loading"
						class="has-no-shadow has-no-border"
						:has-footer="false"
						:change-immediately="true"
					/>
				</div>
			</div>

			<template #footer>
				<XButton
					:loading="filterService.loading"
					:disabled="filterService.loading || !titleValid"
					class="is-fullwidth"
					@click="createFilterWithValidation()"
				>
					{{ $t('filters.create.action') }}
				</XButton>
			</template>
		</Card>
	</Modal>
</template>

<script setup lang="ts">
import Editor from '@/components/input/AsyncEditor'
import Filters from '@/components/project/partials/Filters.vue'

import {useSavedFilter} from '@/services/savedFilter'

const {
	filter,
	filters,
	createFilterWithValidation,
	filterService,
	titleValid,
	validateTitleField,
} = useSavedFilter()
</script>
