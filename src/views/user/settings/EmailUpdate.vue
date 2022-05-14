<template>
	<card v-if="isLocalUser" :title="$t('user.settings.updateEmailTitle')">
		<form @submit.prevent="updateEmail">
			<div class="field">
				<label class="label" for="newEmail">{{ $t('user.settings.updateEmailNew') }}</label>
				<div class="control">
					<input
						@keyup.enter="updateEmail"
						class="input"
						id="newEmail"
						:placeholder="$t('user.auth.emailPlaceholder')"
						type="email"
						v-model="emailUpdate.newEmail"/>
				</div>
			</div>
			<div class="field">
				<label class="label" for="currentPasswordEmail">{{ $t('user.settings.currentPassword') }}</label>
				<div class="control">
					<input
						@keyup.enter="updateEmail"
						class="input"
						id="currentPasswordEmail"
						:placeholder="$t('user.settings.currentPasswordPlaceholder')"
						type="password"
						v-model="emailUpdate.password"/>
				</div>
			</div>
		</form>

		<x-button
			:loading="emailUpdateService.loading"
			@click="updateEmail"
			class="is-fullwidth mt-4">
			{{ $t('misc.save') }}
		</x-button>
	</card>
</template>

<script lang="ts">
import {defineComponent} from 'vue'
export default defineComponent({
	name: 'user-settings-update-email',
})
</script>

<script setup lang="ts">
import {reactive, computed, shallowReactive} from 'vue'
import {useI18n} from 'vue-i18n'
import {useStore} from 'vuex'

import EmailUpdateService from '@/services/emailUpdate'
import EmailUpdateModel from '@/models/emailUpdate'
import {success} from '@/message'
import {useTitle} from '@/composables/useTitle'

const {t} = useI18n()
useTitle(() => `${t('user.settings.updateEmailTitle')} - ${t('user.settings.title')}`)

const store = useStore()
const isLocalUser = computed(() => store.state.auth.info?.isLocalUser)

const emailUpdate = reactive(new EmailUpdateModel())
const emailUpdateService = shallowReactive(new EmailUpdateService())
async function updateEmail() {
	await emailUpdateService.update(emailUpdate)
	success({message: t('user.settings.updateEmailSuccess')})
}
</script>
