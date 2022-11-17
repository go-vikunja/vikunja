<template>
	<create-edit
		:title="$t('namespace.create.title')"
		@create="newNamespace()"
		:primary-disabled="namespace.title === ''"
	>
		<div class="field">
			<label class="label" for="namespaceTitle">{{ $t('namespace.attributes.title') }}</label>
			<div
				class="control is-expanded"
				:class="{ 'is-loading': namespaceService.loading }"
			>
				<!-- The user should be able to close the modal by pressing escape - that already works with the default modal.
					But with the input modal here since it autofocuses the input that input field catches the focus instead.
					Hence we place the listener on the input field directly. -->
				<input
					@keyup.enter="newNamespace()"
					@keyup.esc="$router.back()"
					class="input"
					:placeholder="$t('namespace.attributes.titlePlaceholder')"
					type="text"
					:class="{ disabled: namespaceService.loading }"
					v-focus
					v-model="namespace.title"
				/>
			</div>
		</div>
		<p class="help is-danger" v-if="showError && namespace.title === ''">
			{{ $t('namespace.create.titleRequired') }}
		</p>
		<div class="field">
			<label class="label">{{ $t('namespace.attributes.color') }}</label>
			<div class="control">
				<color-picker v-model="namespace.hexColor"/>
			</div>
		</div>

		<message class="mt-4">
			<h4 class="title">{{ $t('namespace.create.tooltip') }}</h4>

			{{ $t('namespace.create.explanation') }}
		</message>
	</create-edit>
</template>

<script setup lang="ts">
import {ref, shallowReactive} from 'vue'
import {useI18n} from 'vue-i18n'
import {useRouter} from 'vue-router'

import Message from '@/components/misc/message.vue'
import CreateEdit from '@/components/misc/create-edit.vue'
import ColorPicker from '@/components/input/ColorPicker.vue'

import NamespaceModel from '@/models/namespace'
import NamespaceService from '@/services/namespace'
import {useNamespaceStore} from '@/stores/namespaces'
import type {INamespace} from '@/modelTypes/INamespace'

import {useTitle} from '@/composables/useTitle'
import {success} from '@/message'

const showError = ref(false)
const namespace = ref<INamespace>(new NamespaceModel())
const namespaceService = shallowReactive(new NamespaceService())

const {t} = useI18n({useScope: 'global'})
const router = useRouter()

useTitle(() => t('namespace.create.title'))

async function newNamespace() {
	if (namespace.value.title === '') {
		showError.value = true
		return
	}
	showError.value = false

	const newNamespace = await namespaceService.create(namespace.value)
	useNamespaceStore().addNamespace(newNamespace)
	success({message: t('namespace.create.success')})
	router.back()
}
</script>
