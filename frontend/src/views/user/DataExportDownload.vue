<template>
	<div class="content">
		<h1>{{ $t('user.export.downloadTitle') }}</h1>
		<template v-if="isLocalUser">
			<p>{{ $t('user.export.descriptionPasswordRequired') }}</p>
			<FormField
				id="currentPasswordDataExport"
				ref="passwordInput"
				v-model="password"
				:label="$t('user.settings.currentPassword')"
				:class="{'is-danger': errPasswordRequired}"
				:placeholder="$t('user.settings.currentPasswordPlaceholder')"
				type="password"
				:error="errPasswordRequired ? $t('user.deletion.passwordRequired') : null"
				@keyup="() => errPasswordRequired = password === ''"
			/>
		</template>

		<XButton
			v-focus
			:loading="dataExportService.loading"
			class="mbs-4 mie-4"
			@click="download()"
		>
			{{ $t('misc.download') }}
		</XButton>
		<XButton
			class="mbs-4"
			:to="{name:'user.settings.data-export'}"
			variant="tertiary"
		>
			{{ $t('user.export.requestNew') }}
		</XButton>
	</div>
</template>

<script setup lang="ts">
import {ref, computed, reactive} from 'vue'
import DataExportService from '@/services/dataExport'
import FormField from '@/components/input/FormField.vue'
import {useAuthStore} from '@/stores/auth'

const dataExportService = reactive(new DataExportService())
const password = ref('')
const errPasswordRequired = ref(false)
const passwordInput = ref(null)

const authStore = useAuthStore()
const isLocalUser = computed(() => authStore.info?.isLocalUser)

function download() {
	if (password.value === '' && isLocalUser.value) {
		errPasswordRequired.value = true
		passwordInput.value.focus()
		return
	}

	dataExportService.download(password.value)
}
</script>
