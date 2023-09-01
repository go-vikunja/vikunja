<script setup lang="ts">
import ApiTokenService from '@/services/apiToken'
import {computed, onMounted, ref} from 'vue'
import {formatDateShort} from '@/helpers/time/formatDate'
import XButton from '@/components/input/button.vue'
import BaseButton from '@/components/base/BaseButton.vue'
import ApiTokenModel from '@/models/apiTokenModel'
import Fancycheckbox from '@/components/input/fancycheckbox.vue'
import {MILLISECONDS_A_DAY} from '@/constants/date'

const service = new ApiTokenService()
const tokens = ref([])
const apiDocsUrl = window.API_URL + '/docs'
const showCreateForm = ref(false)
const availableRoutes = ref(null)
const newToken = ref(new ApiTokenModel())
const newTokenExpiry = ref<string | number>(30)
const newTokenPermissions = ref({})

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

function deleteToken() {
}

async function createToken() {
	const expiry = Number(newTokenExpiry.value)
	if(!isNaN(expiry)) {
		// if it's a number, we assume it's the number of days in the future
		newToken.value.expiresAt = new Date((new Date()) + expiry * MILLISECONDS_A_DAY)
	}
	
	newToken.value.permissions = {}
	Object.entries(newTokenPermissions.value).forEach(([key, ps]) => {
		const all = Object.entries(ps)
			.filter(([_, v]) => v)
			.map(p => p[0])
		console.log({all})
		if (all.length > 0) {
			newToken.value.permissions[key] = all
		}
	})
	
	const token = await service.create(newToken.value)
	newToken.value = new ApiTokenModel()
	newTokenExpiry.value = 30
	resetPermissions()
	tokens.value.push(token)
	showCreateForm.value = false
}
</script>

<template>
	<card :title="$t('user.settings.apiTokens.title')">

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
				<td>
					<template v-for="(v, p) in tk.permissions">
						<strong>{{ p }}:</strong>
						{{ v.join(', ') }}
						<br/>
					</template>
				</td>
				<td>{{ formatDateShort(tk.expiresAt) }}</td>
				<td>{{ formatDateShort(tk.created) }}</td>
				<td class="has-text-right">
					<x-button variant="secondary" @click="deleteToken(tk)">
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
						type="text"
						v-focus
						v-model="newToken.title"/>
				</div>
			</div>

			<!-- Expiry -->
			<div class="field">
				<label class="label" for="apiTokenTitle">{{
						$t('user.settings.apiTokens.attributes.expiresAt')
					}}</label>
				<div class="control select">
					<select class="select" v-model="newTokenExpiry">
						<option value="30">{{ $t('user.settings.apiTokens.30d') }}</option>
						<option value="60">{{ $t('user.settings.apiTokens.60d') }}</option>
						<option value="90">{{ $t('user.settings.apiTokens.90d') }}</option>
						<option value="custom">{{ $t('misc.custom') }}</option>
					</select>
				</div>
			</div>

			<!-- Permissions -->
			<div class="field">
				<label class="label">{{ $t('user.settings.apiTokens.attributes.permissions') }}</label>
				<p>{{ $t('user.settings.apiTokens.permissionExplanation') }}</p>
				<div v-for="(routes, group) in availableRoutes" class="mb-2" :key="group">
					<strong>{{ group }}</strong><br/>
					<fancycheckbox 
						v-for="(paths, route) in routes"
						:key="group+'-'+route" 
						class="mr-2"
						v-model="newTokenPermissions[group][route]"
					>
						{{ route }}
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
	</card>
</template>
