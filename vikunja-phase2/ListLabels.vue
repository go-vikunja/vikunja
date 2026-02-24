<template>
	<div
		:class="{ 'is-loading': loading}"
		class="content-widescreen loader-container"
	>
		<h2>{{ $t('label.manage') }}</h2>
		<p class="has-text-grey">
			{{ $t('label.description') }}
		</p>

		<hr class="page-separator">

		<div class="tab-actions">
			<XButton
				:to="{name:'labels.create'}"
				icon="plus"
				:shadow="false"
			>
				{{ $t('label.create.header') }}
			</XButton>
		</div>

		<p
			v-if="labelStore.labelsArray.length === 0 && !loading"
			class="has-text-centered has-text-grey p-4"
		>
			{{ $t('label.newCTA') }}
			<RouterLink :to="{name:'labels.create'}">
				{{ $t('label.create.title') }}.
			</RouterLink>
		</p>

		<div
			v-else
			class="columns"
		>
			<div class="labels-list column">
				<RouterLink
					v-for="label in labelStore.labelsArray"
					:key="label.id"
					:to="{name: 'home', query: {labels: label.id.toString()}}"
					:style="getLabelStyles(label)"
					class="tag"
				>
					<span>{{ label.title }}</span>
					<BaseButton
						v-if="userInfo.id === label.createdBy.id"
						class="label-edit-button is-small"
						@click.stop.prevent="editLabel(label)"
					>
						<Icon
							icon="pen"
							class="icon"
						/>
					</BaseButton>
				</RouterLink>
			</div>
			<div
				v-if="isLabelEdit"
				class="column is-4"
			>
				<Card
					:title="$t('label.edit.header')"
					:show-close="true"
					@close="() => isLabelEdit = false"
				>
					<form @submit.prevent="editLabelSubmit()">
						<FormField
							v-model="labelEditLabel.title"
							:label="$t('label.attributes.title')"
							:placeholder="$t('label.attributes.titlePlaceholder')"
							type="text"
						/>
						<FormField :label="$t('label.attributes.description')">
							<Editor
								v-if="editorActive"
								v-model="labelEditLabel.description"
								:placeholder="$t('label.attributes.description')"
							/>
						</FormField>
						<FormField :label="$t('label.attributes.color')">
							<ColorPicker v-model="labelEditLabel.hexColor" />
						</FormField>
						<div class="field has-addons">
							<div class="control is-expanded">
								<XButton
									:loading="loading"
									class="is-fullwidth"
									@click="editLabelSubmit()"
								>
									{{ $t('misc.save') }}
								</XButton>
							</div>
							<div class="control">
								<XButton
									icon="trash-alt"
									danger
									@click="showDeleteDialoge(labelEditLabel)"
								/>
							</div>
						</div>
					</form>
				</Card>
			</div>

			<Modal
				:enabled="showDeleteModal"
				@close="showDeleteModal = false"
				@submit="deleteLabel(labelToDelete)"
			>
				<template #header>
					<span>{{ $t('task.label.delete.header') }}</span>
				</template>

				<template #text>
					<p>
						{{ $t('task.label.delete.text1') }}<br>
						{{ $t('task.label.delete.text2') }}
					</p>
				</template>
			</Modal>
		</div>
	</div>
</template>

<script setup lang="ts">
import {computed, nextTick, ref} from 'vue'
import {useI18n} from 'vue-i18n'

import BaseButton from '@/components/base/BaseButton.vue'
import Editor from '@/components/input/AsyncEditor'
import ColorPicker from '@/components/input/ColorPicker.vue'
import FormField from '@/components/input/FormField.vue'

import LabelModel from '@/models/label'
import type {ILabel} from '@/modelTypes/ILabel'
import {useAuthStore} from '@/stores/auth'
import {useLabelStore} from '@/stores/labels'

import { useTitle } from '@/composables/useTitle'
import {useLabelStyles} from '@/composables/useLabelStyles'

const {t} = useI18n({useScope: 'global'})

const labelEditLabel = ref<ILabel>(new LabelModel())
const isLabelEdit = ref(false)
const editorActive = ref(false)
const showDeleteModal = ref(false)
const labelToDelete = ref<ILabel | undefined>(undefined)

useTitle(() => t('label.title'))

const authStore = useAuthStore()
const userInfo = computed(() => authStore.info)

const labelStore = useLabelStore()
labelStore.loadAllLabels()

const loading = computed(() => labelStore.isLoading)
const {getLabelStyles} = useLabelStyles()

function deleteLabel(label?: ILabel) {
	if (!label) {
		return
	}

	showDeleteModal.value = false
	isLabelEdit.value = false
	return labelStore.deleteLabel(label)
}

function editLabelSubmit() {
	return labelStore.updateLabel(labelEditLabel.value)
}

function editLabel(label: ILabel) {
	if (label.createdBy.id !== userInfo.value.id) {
		return
	}
	labelEditLabel.value = new LabelModel({
		...label,
		created: +label.created,
		updated: +label.updated,
	})
	isLabelEdit.value = true

	editorActive.value = false
	nextTick(() => editorActive.value = true)
}

function showDeleteDialoge(label: ILabel) {
	labelToDelete.value = label
	showDeleteModal.value = true
}
</script>

<style lang="scss" scoped>
.content-widescreen {
	max-inline-size: 900px;
	margin: 0 auto;
	padding: 1.5rem 1rem;
}

.page-separator {
	border: none;
	border-block-start: 2px solid var(--grey-200);
	margin-block: 1rem 1.5rem;
}

.tab-actions {
	margin-block-end: 1.5rem;
}

.label-edit-button {
	border-radius: 100%;
	background-color: rgba(0,0,0,0.2);
	inline-size: 1rem;
	block-size: 1rem;
	display: flex;
	align-items: center;
	justify-content: center;
	color: #ffffff; // always white
	margin-inline-start: .25rem;

	.icon {
		block-size: .5rem;
	}
}
</style>
