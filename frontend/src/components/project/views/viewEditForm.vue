<script setup lang="ts">
import type {IProjectView} from '@/modelTypes/IProjectView'
import XButton from '@/components/input/button.vue'
import FilterInput from '@/components/project/partials/FilterInput.vue'
import {ref} from 'vue'

const model = defineModel<IProjectView>()
const titleValid = ref(true)
function validateTitle() {
	titleValid.value = model.value.title !== ''
}
</script>

<template>
	<form>
		<div class="field">
			<label
				class="label"
				for="title"
			>
				{{ $t('project.views.title') }}
			</label>
			<div class="control">
				<input
					id="title"
					v-model="model.title"
					v-focus
					class="input"
					:placeholder="$t('project.share.links.namePlaceholder')"
					@blur="validateTitle"
				>
			</div>
			<p 
				v-if="!titleValid"
				class="help is-danger"
			>
				{{ $t('project.views.titleRequired') }}
			</p>
		</div>

		<div class="field">
			<label
				class="label"
				for="kind"
			>
				{{ $t('project.views.kind') }}
			</label>
			<div class="control">
				<div class="select">
					<select
						id="kind"
						v-model="model.viewKind"
					>
						<option value="list">
							{{ $t('project.list.title') }}
						</option>
						<option value="gantt">
							{{ $t('project.gantt.title') }}
						</option>
						<option value="table">
							{{ $t('project.table.title') }}
						</option>
						<option value="kanban">
							{{ $t('project.kanban.title') }}
						</option>
					</select>
				</div>
			</div>
		</div>

		<FilterInput
			v-model="model.filter"
			:input-label="$t('project.views.filter')"
		/>

		<div
			v-if="model.viewKind === 'kanban'"
			class="field"
		>
			<label
				class="label"
				for="configMode"
			>
				{{ $t('project.views.bucketConfigMode') }}
			</label>
			<div class="control">
				<div class="select">
					<select
						id="configMode"
						v-model="model.bucketConfigurationMode"
					>
						<option value="manual">
							{{ $t('project.views.bucketConfigManual') }}
						</option>
						<option value="filter">
							{{ $t('project.views.filter') }}
						</option>
					</select>
				</div>
			</div>
		</div>

		<div
			v-if="model.viewKind === 'kanban' && model.bucketConfigurationMode === 'filter'"
			class="field"
		>
			<label class="label">
				{{ $t('project.views.bucketConfig') }}
			</label>
			<div class="control">
				<div
					v-for="(b, index) in model.bucketConfiguration"
					:key="'bucket_'+index"
					class="filter-bucket"
				>
					<button
						class="is-danger"
						@click.prevent="() => model.bucketConfiguration.splice(index, 1)"
					>
						<icon icon="trash-alt" />
					</button>
					<div class="filter-bucket-form">
						<div class="field">
							<label
								class="label"
								:for="'bucket_'+index+'_title'"
							>
								{{ $t('project.views.title') }}
							</label>
							<div class="control">
								<input
									:id="'bucket_'+index+'_title'"
									v-model="model.bucketConfiguration[index].title"
									class="input"
									:placeholder="$t('project.share.links.namePlaceholder')"
								>
							</div>
						</div>

						<FilterInput
							v-model="model.bucketConfiguration[index].filter"
							:input-label="$t('project.views.filter')"
						/>
					</div>
				</div>
				<div class="is-flex is-justify-content-end">
					<XButton
						variant="secondary"
						icon="plus"
						@click="() => model.bucketConfiguration.push({title: '', filter: ''})"
					>
						{{ $t('project.kanban.addBucket') }}
					</XButton>
				</div>
			</div>
		</div>
	</form>
</template>

<style scoped lang="scss">
.filter-bucket {
	display: flex;

	button {
		background: transparent;
		border: none;
		color: var(--danger);
		padding-right: .75rem;
		cursor: pointer;
	}

	&-form {
		margin-bottom: .5rem;
		padding: .5rem;
		border: 1px solid var(--grey-200);
		border-radius: $radius;
		width: 100%;
	}
}
</style>