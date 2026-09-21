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
			:loading="downloadMutation.isPending.value"
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
import {ref, computed} from 'vue'
import {useDownloadExportMutation} from '@/client/queries/dataExport'
import FormField from '@/components/input/FormField.vue'
import {useAuthStore} from '@/stores/auth'

const downloadMutation = useDownloadExportMutation()
const password = ref('')
const errPasswordRequired = ref(false)
const passwordInput = ref<InstanceType<typeof FormField>>()

const authStore = useAuthStore()
const isLocalUser = computed(() => authStore.info?.is_local_user)

async function download() {
	if (password.value === '' && isLocalUser.value) {
		errPasswordRequired.value = true
		passwordInput.value?.focus()
		return
	}

	try {
		await downloadMutation.mutateAsync(password.value)
		downloadMutation.reset()
	} catch { return }
}
</script>
