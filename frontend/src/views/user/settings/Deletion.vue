<template>
	<Card
		v-if="userDeletionEnabled"
		:title="$t('user.deletion.title')"
	>
		<template v-if="deletionScheduledAt !== null">
			<form @submit.prevent="cancelDeletion()">
				<p>
					{{
						$t('user.deletion.scheduled', {
							date: formatDisplayDate(deletionScheduledAt),
							dateSince: formatDateSince(deletionScheduledAt),
						})
					}}
				</p>
				<template v-if="isLocalUser">
					<p>
						{{ $t('user.deletion.scheduledCancelText') }}
					</p>
					<div class="field">
						<label
							class="label"
							for="currentPasswordAccountDelete"
						>
							{{ $t('user.settings.currentPassword') }}
						</label>
						<div class="control">
							<input
								id="currentPasswordAccountDelete"
								ref="passwordInput"
								v-model="password"
								class="input"
								:class="{'is-danger': errPasswordRequired}"
								:placeholder="$t('user.settings.currentPasswordPlaceholder')"
								type="password"
								@keyup="() => errPasswordRequired = password === ''"
							>
						</div>
						<p
							v-if="errPasswordRequired"
							class="help is-danger"
						>
							{{ $t('user.deletion.passwordRequired') }}
						</p>
					</div>
				</template>
				<p v-else>
					{{ $t('user.deletion.scheduledCancelButton') }}
				</p>
			</form>

			<XButton
				:loading="accountDeleteService.loading"
				class="is-fullwidth mbs-4"
				@click="cancelDeletion()"
			>
				{{ $t('user.deletion.scheduledCancelConfirm') }}
			</XButton>
		</template>
		<template v-else>
			<p>
				{{ $t('user.deletion.text1') }}
			</p>
			<form
				v-if="isLocalUser"
				@submit.prevent="deleteAccount()"
			>
				<p>
					{{ $t('user.deletion.text2') }}
				</p>
				<div class="field">
					<label
						class="label"
						for="currentPasswordAccountDelete"
					>
						{{ $t('user.settings.currentPassword') }}
					</label>
					<div class="control">
						<input
							id="currentPasswordAccountDelete"
							ref="passwordInput"
							v-model="password"
							class="input"
							:class="{'is-danger': errPasswordRequired}"
							:placeholder="$t('user.settings.currentPasswordPlaceholder')"
							type="password"
							@keyup="() => errPasswordRequired = password === ''"
						>
					</div>
					<p
						v-if="errPasswordRequired"
						class="help is-danger"
					>
						{{ $t('user.deletion.passwordRequired') }}
					</p>
				</div>
			</form>
			<p v-else>
				{{ $t('user.deletion.text3') }}
			</p>

			<XButton
				:loading="accountDeleteService.loading"
				class="is-fullwidth mbs-4 is-danger"
				@click="deleteAccount()"
			>
				{{ $t('user.deletion.confirm') }}
			</XButton>
		</template>
	</Card>
</template>

<script setup lang="ts">
import {ref, shallowReactive, computed} from 'vue'
import {useI18n} from 'vue-i18n'

import AccountDeleteService from '@/services/accountDelete'
import {parseDateOrNull} from '@/helpers/parseDateOrNull'
import {formatDateSince, formatDisplayDate} from '@/helpers/time/formatDate'
import {useTitle} from '@/composables/useTitle'
import {success} from '@/message'
import {useAuthStore} from '@/stores/auth'
import {useConfigStore} from '@/stores/config'

defineOptions({name: 'UserSettingsDeletion'})

const {t} = useI18n({useScope: 'global'})
useTitle(() => `${t('user.deletion.title')} - ${t('user.settings.title')}`)

const accountDeleteService = shallowReactive(new AccountDeleteService())
const password = ref('')
const errPasswordRequired = ref(false)

const authStore = useAuthStore()
const configStore = useConfigStore()

const userDeletionEnabled = computed(() => configStore.userDeletionEnabled)
const deletionScheduledAt = computed(() => parseDateOrNull(authStore.info?.deletionScheduledAt))

const isLocalUser = computed(() => authStore.info?.isLocalUser)

const passwordInput = ref()

async function deleteAccount() {
	if (isLocalUser.value && password.value === '') {
		errPasswordRequired.value = true
		passwordInput.value.focus()
		return
	}

	await accountDeleteService.request(password.value)
	success({message: t('user.deletion.requestSuccess')})
	password.value = ''
}

async function cancelDeletion() {
	if (isLocalUser.value && password.value === '') {
		errPasswordRequired.value = true
		passwordInput.value.focus()
		return
	}

	await accountDeleteService.cancel(password.value)
	success({message: t('user.deletion.scheduledCancelSuccess')})
	authStore.refreshUserInfo()
	password.value = ''
}
</script>
