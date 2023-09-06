<script setup lang="ts">
import ApiTokenService from '@/services/apiToken'
import {computed, onMounted, ref} from 'vue'
import {formatDateShort, formatDateSince} from '@/helpers/time/formatDate'
import XButton from '@/components/input/button.vue'
import BaseButton from '@/components/base/BaseButton.vue'
import ApiTokenModel from '@/models/apiTokenModel'
import Fancycheckbox from '@/components/input/fancycheckbox.vue'
import {MILLISECONDS_A_DAY} from '@/constants/date'
import flatPickr from 'vue-flatpickr-component'
import 'flatpickr/dist/flatpickr.css'
import {useI18n} from 'vue-i18n'
import {useAuthStore} from '@/stores/auth'
import Message from '@/components/misc/message.vue'

const service = new ApiTokenService()
const tokens = ref([])
const apiDocsUrl = window.API_URL + '/docs'
const showCreateForm = ref(false)
const availableRoutes = ref(null)
const newToken = ref(new ApiTokenModel())
const newTokenExpiry = ref<string | number>(30)
const newTokenExpiryCustom = ref(new Date())
const newTokenPermissions = ref({})
const newTokenTitleValid = ref(true)
const apiTokenTitle = ref()
const tokenCreatedSuccessMessage = ref('')

const showDeleteModal = ref(false)
const tokenToDelete = ref(null)

const {t} = useI18n()
const authStore = useAuthStore()

const now = new Date()
const flatPickerConfig = computed(() => ({
	altFormat: t('date.altFormatLong'),
	altInput: true,
	dateFormat: 'Y-m-d H:i',
	enableTime: true,
	time_24hr: true,
	locale: {
		firstDayOfWeek: authStore.settings.weekStart,
	},
	minDate: now,
}))

onMounted(async () => {
	tokens.value = await service.getAll()
	availableRoutes.value = await service.getAvailableRoutes()
	resetPermissions()
})

function resetPermissions() {
	newTokenPermissions.value = {}
	Object.entries(availableRoutes.value).forEach(entry => {
		const [group, routes] = entry
		newTokenPermissions.value[group] = {}
		Object.keys(routes).forEach(r => {
			newTokenPermissions.value[group][r] = false
		})
	})
}

async function deleteToken() {
	await service.delete(tokenToDelete.value)
	showDeleteModal.value = false
	tokenToDelete.value = null
	const index = tokens.value.findIndex(el => el.id === tokenToDelete.value.id)
	if (index === -1) {
		return
	}
	tokens.value.splice(index, 1)
}

async function createToken() {
	if (!newTokenTitleValid.value) {
		apiTokenTitle.value.focus()
		return
	}

	const expiry = Number(newTokenExpiry.value)
	if (!isNaN(expiry)) {
		// if it's a number, we assume it's the number of days in the future
		newToken.value.expiresAt = new Date((+new Date()) + expiry * MILLISECONDS_A_DAY)
	} else {
		newToken.value.expiresAt = new Date(newTokenExpiryCustom.value)
	}

	newToken.value.permissions = {}
	Object.entries(newTokenPermissions.value).forEach(([key, ps]) => {
		const all = Object.entries(ps)
			// eslint-disable-next-line @typescript-eslint/no-unused-vars
			.filter(([_, v]) => v)
			.map(p => p[0])
		if (all.length > 0) {
			newToken.value.permissions[key] = all
		}
	})

	const token = await service.create(newToken.value)
	tokenCreatedSuccessMessage.value = t('user.settings.apiTokens.tokenCreatedSuccess', {token: token.token})
	newToken.value = new ApiTokenModel()
	newTokenExpiry.value = 30
	newTokenExpiryCustom.value = new Date()
	resetPermissions()
	tokens.value.push(token)
	showCreateForm.value = false
}

function formatPermissionTitle(title: string): string {
	return title.replaceAll('_', ' ')
}
</script>

<template>
	<card :title="$t('user.settings.apiTokens.title')">

		<message v-if="tokenCreatedSuccessMessage !== ''" class="has-text-centered mb-4">
			{{ tokenCreatedSuccessMessage }}<br/>
			{{ $t('user.settings.apiTokens.tokenCreatedNotSeeAgain') }}
		</message>

		<p>
			{{ $t('user.settings.apiTokens.general') }}
			<BaseButton :href="apiDocsUrl">{{ $t('user.settings.apiTokens.apiDocs') }}</BaseButton>
			.
		</p>

		<table class="table" v-if="tokens.length > 0">
			<tr>
				<th>{{ $t('misc.id') }}</th>
				<th>{{ $t('user.settings.apiTokens.attributes.title') }}</th>
				<th>{{ $t('user.settings.apiTokens.attributes.permissions') }}</th>
				<th>{{ $t('user.settings.apiTokens.attributes.expiresAt') }}</th>
				<th>{{ $t('misc.created') }}</th>
				<th class="has-text-right">{{ $t('misc.actions') }}</th>
			</tr>
			<tr v-for="tk in tokens" :key="tk.id">
				<td>{{ tk.id }}</td>
				<td>{{ tk.title }}</td>
				<td class="is-capitalized">
					<template v-for="(v, p) in tk.permissions" :key="'permission-' + p">
						<strong>{{ formatPermissionTitle(p) }}:</strong>
						{{ v.map(formatPermissionTitle).join(', ') }}
						<br/>
					</template>
				</td>
				<td>
					{{ formatDateShort(tk.expiresAt) }}
					<p v-if="tk.expiresAt < new Date()" class="has-text-danger">
						{{ $t('user.settings.apiTokens.expired', {ago: formatDateSince(tk.expiresAt)}) }}
					</p>
				</td>
				<td>{{ formatDateShort(tk.created) }}</td>
				<td class="has-text-right">
					<x-button variant="secondary" @click="() => {tokenToDelete = tk; showDeleteModal = true}">
						{{ $t('misc.delete') }}
					</x-button>
				</td>
			</tr>
		</table>

		<form
			v-if="showCreateForm"
			@submit.prevent="createToken"
		>
			<!-- Title -->
			<div class="field">
				<label class="label" for="apiTokenTitle">{{ $t('user.settings.apiTokens.attributes.title') }}</label>
				<div class="control">
					<input
						class="input"
						id="apiTokenTitle"
						ref="apiTokenTitle"
						type="text"
						v-focus
						:placeholder="$t('user.settings.apiTokens.attributes.titlePlaceholder')"
						v-model="newToken.title"
						@keyup="() => newTokenTitleValid = newToken.title !== ''"
						@focusout="() => newTokenTitleValid = newToken.title !== ''"
					/>
				</div>
				<p class="help is-danger" v-if="!newTokenTitleValid">
					{{ $t('user.settings.apiTokens.titleRequired') }}
				</p>
			</div>

			<!-- Expiry -->
			<div class="field">
				<label class="label" for="apiTokenExpiry">
					{{ $t('user.settings.apiTokens.attributes.expiresAt') }}
				</label>
				<div class="is-flex">
					<div class="control select">
						<select class="select" v-model="newTokenExpiry" id="apiTokenExpiry">
							<option value="30">{{ $t('user.settings.apiTokens.30d') }}</option>
							<option value="60">{{ $t('user.settings.apiTokens.60d') }}</option>
							<option value="90">{{ $t('user.settings.apiTokens.90d') }}</option>
							<option value="custom">{{ $t('misc.custom') }}</option>
						</select>
					</div>
					<flat-pickr
						v-if="newTokenExpiry === 'custom'"
						class="ml-2"
						:config="flatPickerConfig"
						v-model="newTokenExpiryCustom"
					/>
				</div>
			</div>

			<!-- Permissions -->
			<div class="field">
				<label class="label">{{ $t('user.settings.apiTokens.attributes.permissions') }}</label>
				<p>{{ $t('user.settings.apiTokens.permissionExplanation') }}</p>
				<div v-for="(routes, group) in availableRoutes" class="mb-2" :key="group">
					<strong class="is-capitalized">{{ formatPermissionTitle(group) }}</strong><br/>
					<fancycheckbox
						v-for="(paths, route) in routes"
						:key="group+'-'+route"
						class="mr-2 is-capitalized"
						v-model="newTokenPermissions[group][route]"
					>
						{{ formatPermissionTitle(route) }}
					</fancycheckbox>
					<br/>
				</div>
			</div>

			<x-button :loading="service.loading" @click="createToken">
				{{ $t('user.settings.apiTokens.createToken') }}
			</x-button>
		</form>

		<x-button
			v-else
			icon="plus"
			class="mb-4"
			@click="() => showCreateForm = true"
			:loading="service.loading"
		>
			{{ $t('user.settings.apiTokens.createAToken') }}
		</x-button>

		<modal
			:enabled="showDeleteModal"
			@close="showDeleteModal = false"
			@submit="deleteToken()"
		>
			<template #header>
				{{ $t('user.settings.apiTokens.delete.header') }}
			</template>

			<template #text>
				<p>
					{{ $t('user.settings.apiTokens.delete.text1', {token: tokenToDelete.title}) }}<br/>
					{{ $t('user.settings.apiTokens.delete.text2') }}
				</p>
			</template>
		</modal>
	</card>
</template>
