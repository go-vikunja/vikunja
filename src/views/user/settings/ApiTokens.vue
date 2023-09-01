<script setup lang="ts">
import ApiTokenService from '@/services/apiToken'
import {computed, onMounted, ref} from 'vue'
import { formatDateShort } from '@/helpers/time/formatDate'
import XButton from '@/components/input/button.vue'
import BaseButton from '@/components/base/BaseButton.vue'

const service = new ApiTokenService()
const tokens = ref([])

const apiDocsUrl = window.API_URL + '/docs'

onMounted(async () => {
	tokens.value = await service.getAll()
})

function deleteToken() {
}

function createToken() {
}
</script>

<template>
	<card :title="$t('user.settings.apiTokens.title')">
		
		<p>
			{{ $t('user.settings.apiTokens.general') }}
			<BaseButton :href="apiDocsUrl">{{ $t('user.settings.apiTokens.apiDocs') }}</BaseButton>.
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
		
		<x-button icon="plus" class="mb-4" @click="createToken" :loading="service.loading">
			{{ $t('user.settings.apiTokens.createToken') }}
		</x-button>
	</card>
</template>

<style scoped lang="scss">

</style>