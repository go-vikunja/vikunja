<template>
	<div class="content">
		<h1>{{ $t('user.export.downloadTitle') }}</h1>
		<template v-if="isLocalUser">
			<p>{{ $t('user.export.descriptionPasswordRequired') }}</p>
			<div class="field">
				<label
					class="label"
					for="currentPasswordDataExport"
				>
					{{ $t('user.settings.currentPassword') }}
				</label>
				<div class="control">
					<input
						id="currentPasswordDataExport"
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

		<XButton
			v-focus
			:loading="dataExportService.loading"
			class="mbs-4 mie-4"
			@click="download()"
		>
			{{ $t('misc.download') }}
		</XButton>
		<XButton
			class="button mbs-4"
			:to="{name:'user.settings.data-export'}"
			variant="tertary"
		>
			{{ $t('user.export.requestNew') }}
		</XButton>
	</div>
</template>

<script setup lang="ts">
import {ref, computed, reactive} from 'vue'
import DataExportService from '@/services/dataExport'
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
