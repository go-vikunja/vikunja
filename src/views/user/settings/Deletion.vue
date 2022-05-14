<template>
	<card :title="$t('user.deletion.title')" v-if="userDeletionEnabled">
		<template v-if="deletionScheduledAt !== null">
			<form @submit.prevent="cancelDeletion()">
				<p>
					{{
						$t('user.deletion.scheduled', {
							date: formatDateShort(deletionScheduledAt),
							dateSince: formatDateSince(deletionScheduledAt),
						})
					}}
				</p>
				<p>
					{{ $t('user.deletion.scheduledCancelText') }}
				</p>
				<div class="field">
					<label class="label" for="currentPasswordAccountDelete">
						{{ $t('user.settings.currentPassword') }}
					</label>
					<div class="control">
						<input
							class="input"
							:class="{'is-danger': errPasswordRequired}"
							id="currentPasswordAccountDelete"
							:placeholder="$t('user.settings.currentPasswordPlaceholder')"
							type="password"
							v-model="password"
							@keyup="() => errPasswordRequired = password === ''"
							ref="passwordInput"
						/>
					</div>
					<p class="help is-danger" v-if="errPasswordRequired">
						{{ $t('user.deletion.passwordRequired') }}
					</p>
				</div>
			</form>

			<x-button
				:loading="accountDeleteService.loading"
				@click="cancelDeletion()"
				class="is-fullwidth mt-4">
				{{ $t('user.deletion.scheduledCancelConfirm') }}
			</x-button>
		</template>
		<template v-else>
			<form @submit.prevent="deleteAccount()">
				<p>
					{{ $t('user.deletion.text1') }}
				</p>
				<p>
					{{ $t('user.deletion.text2') }}
				</p>
				<div class="field">
					<label class="label" for="currentPasswordAccountDelete">
						{{ $t('user.settings.currentPassword') }}
					</label>
					<div class="control">
						<input
							class="input"
							:class="{'is-danger': errPasswordRequired}"
							id="currentPasswordAccountDelete"
							:placeholder="$t('user.settings.currentPasswordPlaceholder')"
							type="password"
							v-model="password"
							@keyup="() => errPasswordRequired = password === ''"
							ref="passwordInput"
						/>
					</div>
					<p class="help is-danger" v-if="errPasswordRequired">
						{{ $t('user.deletion.passwordRequired') }}
					</p>
				</div>
			</form>

			<x-button
				:loading="accountDeleteService.loading"
				@click="deleteAccount()"
				class="is-fullwidth mt-4 is-danger">
				{{ $t('user.deletion.confirm') }}
			</x-button>
		</template>
	</card>
</template>

<script lang="ts">
import {defineComponent} from 'vue'

export default defineComponent({
	name: 'user-settings-deletion',
})
</script>

<script setup lang="ts">
import {ref, shallowReactive, computed} from 'vue'
import {useStore} from 'vuex'
import {useI18n} from 'vue-i18n'

import AccountDeleteService from '@/services/accountDelete'
import {parseDateOrNull} from '@/helpers/parseDateOrNull'
import {useTitle} from '@/composables/useTitle'
import {success} from '@/message'

const {t} = useI18n()
useTitle(() => `${t('user.deletion.title')} - ${t('user.settings.title')}`)

const accountDeleteService = shallowReactive(new AccountDeleteService())
const password = ref('')
const errPasswordRequired = ref(false)

const store = useStore()
const userDeletionEnabled = computed(() => store.state.config.userDeletionEnabled)
const deletionScheduledAt = computed(() => parseDateOrNull(store.state.auth.info?.deletionScheduledAt))

const passwordInput = ref()
async function deleteAccount() {
	if (password.value === '') {
		errPasswordRequired.value = true
		passwordInput.value.focus()
		return
	}

	await accountDeleteService.request(password.value)
	success({message: t('user.deletion.requestSuccess')})
	password.value = ''
}

async function cancelDeletion() {
	if (password.value === '') {
		errPasswordRequired.value = true
		passwordInput.value.focus()
		return
	}

	await accountDeleteService.cancel(password.value)
	success({message: t('user.deletion.scheduledCancelSuccess')})
	store.dispatch('auth/refreshUserInfo')
	password.value = ''
}
</script>
