<template>
	<create-edit
		:title="$t('label.create.title')"
		@create="newLabel()"
		:primary-disabled="label.title === ''"
	>
		<div class="field">
			<label class="label" for="labelTitle">{{ $t('label.attributes.title') }}</label>
			<div
				class="control is-expanded"
				:class="{ 'is-loading': loading }"
			>
				<input
					:class="{ disabled: loading }"
					class="input"
					:placeholder="$t('label.attributes.titlePlaceholder')"
					type="text"
					id="labelTitle"
					v-focus
					v-model="label.title"
					@keyup.enter="newLabel()"
				/>
			</div>
		</div>
		<p class="help is-danger" v-if="showError && label.title === ''">
			{{ $t('label.create.titleRequired') }}
		</p>
		<div class="field">
			<label class="label">{{ $t('label.attributes.color') }}</label>
			<div class="control">
				<color-picker v-model="label.hexColor"/>
			</div>
		</div>
	</create-edit>
</template>

<script setup lang="ts">
import {computed, ref} from 'vue'
import {useI18n} from 'vue-i18n'
import {useRouter} from 'vue-router'

import CreateEdit from '@/components/misc/create-edit.vue'
import ColorPicker from '@/components/input/ColorPicker.vue'

import LabelModel from '@/models/label'
import {useLabelStore} from '@/stores/labels'
import {useTitle} from '@/composables/useTitle'
import {success} from '@/message'

const router = useRouter()

const {t} = useI18n({useScope: 'global'})
useTitle(() => t('label.create.title'))

const labelStore = useLabelStore()
const label = ref(new LabelModel())

const showError = ref(false)
const loading = computed(() => labelStore.isLoading)

async function newLabel() {
	if (label.value.title === '') {
		showError.value = true
		return
	}
	showError.value = false

	const newLabel = await labelStore.createLabel(label.value)
	router.push({
		name: 'labels.index',
		params: {id: newLabel.id},
	})
	success({message: t('label.create.success')})
}
</script>
