<template>
	<div
		:class="{ 'is-loading': loading}"
		class="loader-container"
	>
		<XButton
			:to="{name:'labels.create'}"
			class="is-pulled-right"
			icon="plus"
		>
			{{ $t('label.create.header') }}
		</XButton>

		<div class="content">
			<h1>{{ $t('label.manage') }}</h1>
			<p v-if="labelStore.labelsArray.length > 0">
				{{ $t('label.description') }}
			</p>
			<p
				v-else
				class="has-text-centered has-text-grey is-italic"
			>
				{{ $t('label.newCTA') }}
				<RouterLink :to="{name:'labels.create'}">
					{{ $t('label.create.title') }}.
				</RouterLink>
			</p>
		</div>

		<div class="columns">
			<div class="labels-list column">
				<span
					v-for="label in labelStore.labelsArray"
					:key="label.id"
					:class="{'disabled': userInfo.id !== label.createdBy.id}"
					:style="getLabelStyles(label)"
					class="tag"
				>
					<span
						v-if="userInfo.id !== label.createdBy.id"
						v-tooltip.bottom="$t('label.edit.forbidden')"
					>
						{{ label.title }}
					</span>
					<BaseButton
						v-else
						:style="{'color': label.textColor}"
						@click="editLabel(label)"
					>
						{{ label.title }}
					</BaseButton>
					<BaseButton
						v-if="userInfo.id === label.createdBy.id"
						class="delete is-small"
						@click="showDeleteDialoge(label)"
					/>
				</span>
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
						<div class="field">
							<label class="label">{{ $t('label.attributes.title') }}</label>
							<div class="control">
								<input
									v-model="labelEditLabel.title"
									class="input"
									:placeholder="$t('label.attributes.titlePlaceholder')"
									type="text"
								>
							</div>
						</div>
						<div class="field">
							<label class="label">{{ $t('label.attributes.description') }}</label>
							<div class="control">
								<Editor
									v-if="editorActive"
									v-model="labelEditLabel.description"
									:placeholder="$t('label.attributes.description')"
								/>
							</div>
						</div>
						<div class="field">
							<label class="label">{{ $t('label.attributes.color') }}</label>
							<div class="control">
								<ColorPicker v-model="labelEditLabel.hexColor" />
							</div>
						</div>
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
									class="is-danger"
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
	// Duplicating the label to make sure it does not look like changes take effect immediatly as the label 
	// object passed to this function here still has a reference to the store.
	labelEditLabel.value = new LabelModel({
		...label,
		// The model does not support passing dates into it directly so we need to convert them first				
		created: +label.created,
		updated: +label.updated,
	})
	isLabelEdit.value = true

	// This makes the editor trigger its mounted function again which makes it forget every input
	// it currently has in its textarea. This is a counter-hack to a hack inside of vue-easymde
	// which made it impossible to detect change from the outside. Therefore the component would
	// not update if new content from the outside was made available.
	// See https://github.com/NikulinIlya/vue-easymde/issues/3
	editorActive.value = false
	nextTick(() => editorActive.value = true)
}

function showDeleteDialoge(label: ILabel) {
	labelToDelete.value = label
	showDeleteModal.value = true
}
</script>
