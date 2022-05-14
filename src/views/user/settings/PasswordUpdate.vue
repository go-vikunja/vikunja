<template>
	<card v-if="isLocalUser" :title="$t('user.settings.newPasswordTitle')" :loading="passwordUpdateService.loading">
		<form @submit.prevent="updatePassword">
			<div class="field">
				<label class="label" for="newPassword">{{ $t('user.settings.newPassword') }}</label>
				<div class="control">
					<input
						autocomplete="new-password"
						@keyup.enter="updatePassword"
						class="input"
						id="newPassword"
						:placeholder="$t('user.auth.passwordPlaceholder')"
						type="password"
						v-model="passwordUpdate.newPassword"/>
				</div>
			</div>
			<div class="field">
				<label class="label" for="newPasswordConfirm">{{ $t('user.settings.newPasswordConfirm') }}</label>
				<div class="control">
					<input
						autocomplete="new-password"
						@keyup.enter="updatePassword"
						class="input"
						id="newPasswordConfirm"
						:placeholder="$t('user.auth.passwordPlaceholder')"
						type="password"
						v-model="passwordConfirm"/>
				</div>
			</div>
			<div class="field">
				<label class="label" for="currentPassword">{{ $t('user.settings.currentPassword') }}</label>
				<div class="control">
					<input
						autocomplete="current-password"
						@keyup.enter="updatePassword"
						class="input"
						id="currentPassword"
						:placeholder="$t('user.settings.currentPasswordPlaceholder')"
						type="password"
						v-model="passwordUpdate.oldPassword"/>
				</div>
			</div>
		</form>

		<x-button
			:loading="passwordUpdateService.loading"
			@click="updatePassword"
			class="is-fullwidth mt-4">
			{{ $t('misc.save') }}
		</x-button>
	</card>
</template>

<script lang="ts">
import {defineComponent} from 'vue'

export default defineComponent({
	name: 'user-settings-password-update',
})
</script>

<script setup lang="ts">
import {ref, reactive, shallowReactive, computed} from 'vue'
import { useI18n } from 'vue-i18n'
import { useStore } from 'vuex'

import PasswordUpdateService from '@/services/passwordUpdateService'
import PasswordUpdateModel from '@/models/passwordUpdate'

import {useTitle} from '@/composables/useTitle'
import {success, error} from '@/message'

const passwordUpdateService = shallowReactive(new PasswordUpdateService())
const passwordUpdate = reactive(new PasswordUpdateModel())
const passwordConfirm = ref('')

const {t} = useI18n()
const store = useStore()
useTitle(() => `${t('user.settings.newPasswordTitle')} - ${t('user.settings.title')}`)

const isLocalUser = computed(() => store.state.auth.info?.isLocalUser)

async function updatePassword() {
	if (passwordConfirm.value !== passwordUpdate.newPassword) {
		error({message: t('user.settings.passwordsDontMatch')})
		return
	}

	await passwordUpdateService.update(passwordUpdate)
	success({message: t('user.settings.passwordUpdateSuccess')})
}
</script>
