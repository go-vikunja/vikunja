<template>
	<create-edit
		:title="title"
		primary-icon=""
		:primary-label="$t('misc.save')"
		@primary="save"
		:tertiary="$t('misc.delete')"
		@tertiary="$router.push({ name: 'namespace.settings.delete', params: { id: $route.params.id } })"
	>
		<form @submit.prevent="save()">
			<div class="field">
				<label class="label" for="namespacetext">{{ $t('namespace.attributes.title') }}</label>
				<div class="control">
					<input
						:class="{ 'disabled': namespaceService.loading}"
						:disabled="namespaceService.loading || undefined"
						class="input"
						id="namespacetext"
						:placeholder="$t('namespace.attributes.titlePlaceholder')"
						type="text"
						v-focus
						v-model="namespace.title"/>
				</div>
			</div>
			<div class="field">
				<label class="label" for="namespacedescription">{{ $t('namespace.attributes.description') }}</label>
				<div class="control">
					<AsyncEditor
						:class="{ 'disabled': namespaceService.loading}"
						:preview-is-default="false"
						id="namespacedescription"
						:placeholder="$t('namespace.attributes.descriptionPlaceholder')"
						v-if="editorActive"
						v-model="namespace.description"
					/>
				</div>
			</div>
			<div class="field">
				<label class="label" for="isArchivedCheck">{{ $t('namespace.attributes.archived') }}</label>
				<div class="control">
					<fancycheckbox
						v-model="namespace.isArchived"
						v-tooltip="$t('namespace.archive.description')">
						{{ $t('namespace.attributes.isArchived') }}
					</fancycheckbox>
				</div>
			</div>
			<div class="field">
				<label class="label">{{ $t('namespace.attributes.color') }}</label>
				<div class="control">
					<color-picker v-model="namespace.hexColor"/>
				</div>
			</div>
		</form>
	</create-edit>
</template>

<script lang="ts" setup>
import {nextTick, ref, watch} from 'vue'
import {success} from '@/message'
import router from '@/router'

import AsyncEditor from '@/components/input/AsyncEditor'
import Fancycheckbox from '@/components/input/fancycheckbox.vue'
import ColorPicker from '@/components/input/colorPicker.vue'
import CreateEdit from '@/components/misc/create-edit.vue'

import NamespaceService from '@/services/namespace'
import NamespaceModel from '@/models/namespace'
import {useI18n} from 'vue-i18n'
import {useTitle} from '@/composables/useTitle'
import {useNamespaceStore} from '@/stores/namespaces'

const {t} = useI18n({useScope: 'global'})
const namespaceStore = useNamespaceStore()

const namespaceService = ref(new NamespaceService())
const namespace = ref(new NamespaceModel())
const editorActive = ref(false)
const title = ref('')
useTitle(() => title.value)

const props = defineProps({
	namespaceId: {
		type: Number,
		required: true,
	},
})

watch(
	() => props.namespaceId,
	loadNamespace,
	{
		immediate: true,
	},
)

async function loadNamespace() {
	// HACK: This makes the editor trigger its mounted function again which makes it forget every input
	// it currently has in its textarea. This is a counter-hack to a hack inside of vue-easymde
	// which made it impossible to detect change from the outside. Therefore the component would
	// not update if new content from the outside was made available.
	// See https://github.com/NikulinIlya/vue-easymde/issues/3
	editorActive.value = false
	nextTick(() => editorActive.value = true)

	namespace.value = await namespaceService.value.get({id: props.namespaceId})
	title.value = t('namespace.edit.title', {namespace: namespace.value.title})
}

async function save() {
	const updatedNamespace = await namespaceService.value.update(namespace.value)
	// Update the namespace in the parent
	namespaceStore.setNamespaceById(updatedNamespace)
	success({message: t('namespace.edit.success')})
	router.back()
}
</script>