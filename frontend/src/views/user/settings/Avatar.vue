<template>
	<Card :title="$t('user.settings.avatar.title')">
		<Message v-if="avatarProvider === 'ldap'">
			{{ $t('user.settings.avatar.ldap') }}
		</Message>

		<Message v-else-if="avatarProvider === 'openid'">
			{{ $t('user.settings.avatar.openid', {provider: authStore.info?.auth_provider}) }}
		</Message>

		<template v-else>
			<div class="control mbe-4">
				<label
					v-for="(label, providerId) in AVATAR_PROVIDERS"
					:key="providerId"
					class="radio"
				>
					<input
						v-model="avatarProvider"
						name="avatarProvider"
						type="radio"
						:value="providerId"
					>
					{{ label }}
				</label>
			</div>

			<template v-if="avatarProvider === 'upload'">
				<input
					ref="avatarUploadInput"
					accept="image/*"
					class="is-hidden"
					type="file"
					@change="cropAvatar"
				>

				<XButton
					v-if="!isCropAvatar"
					:loading="saving || loading"
					@click="avatarUploadInput.click()"
				>
					{{ $t('user.settings.avatar.uploadAvatar') }}
				</XButton>
				<template v-else>
					<Cropper
						ref="cropper"
						:src="avatarToCrop"
						:stencil-props="{aspectRatio: 1}"
						class="mbe-4 cropper"
						@ready="() => loading = false"
					/>
					<XButton
						v-cy="'uploadAvatar'"
						:loading="saving || loading"
						@click="uploadAvatar"
					>
						{{ $t('user.settings.avatar.uploadAvatar') }}
					</XButton>
				</template>
			</template>

			<div
				v-else
				class="mbs-2"
			>
				<XButton
					:loading="saving || loading"
					class="is-fullwidth"
					@click="updateAvatarStatus()"
				>
					{{ $t('misc.save') }}
				</XButton>
			</div>
		</template>
	</Card>
</template>


<script setup lang="ts">
import {computed, ref, watch} from 'vue'
import {useI18n} from 'vue-i18n'
import {Cropper} from 'vue-advanced-cropper'
import 'vue-advanced-cropper/dist/style.css'

import {useQuery} from '@tanstack/vue-query'
import {avatarProviderQuery, useUpdateAvatarProviderMutation, useUploadAvatarMutation} from '@/client/queries/avatars'
import {useTitle} from '@/composables/useTitle'
import {useAuthStore} from '@/stores/auth'
import Message from '@/components/misc/Message.vue'

defineOptions({name: 'UserSettingsAvatar'})

const {t} = useI18n({useScope: 'global'})
const authStore = useAuthStore()

const AVATAR_PROVIDERS = computed(() => ({
	default: t('misc.default'),
	initials: t('user.settings.avatar.initials'),
	gravatar: t('user.settings.avatar.gravatar'),
	marble: t('user.settings.avatar.marble'),
	upload: t('user.settings.avatar.upload'),
}))

useTitle(() => `${t('user.settings.avatar.title')} - ${t('user.settings.title')}`)

const providerQuery = useQuery(avatarProviderQuery())
const updateProvider = useUpdateAvatarProviderMutation()
const upload = useUploadAvatarMutation()
const saving = computed(() => updateProvider.isPending.value || upload.isPending.value)
const loading = ref(false)
const avatarProvider = ref<string>()
watch(providerQuery.data, data => {
	if (avatarProvider.value === undefined && data) avatarProvider.value = data.avatar_provider ?? 'default'
}, {immediate: true})

async function updateAvatarStatus() {
	if (!avatarProvider.value) return
	try {
		await updateProvider.mutateAsync({username: authStore.info?.username ?? '', provider: avatarProvider.value})
	} catch { return }
}

const cropper = ref()
const isCropAvatar = ref(false)

async function uploadAvatar() {
	loading.value = true
	const {canvas} = cropper.value.getResult()

	if (!canvas) {
		loading.value = false
		return
	}

	try {
		const blob = await new Promise<Blob | null>(resolve => canvas.toBlob(resolve))
		if (!blob) return
		await upload.mutateAsync({username: authStore.info?.username ?? '', blob})
	} catch {
		return
	} finally {
		loading.value = false
		isCropAvatar.value = false
	}
}

const avatarToCrop = ref()
const avatarUploadInput = ref()

function cropAvatar() {
	const avatar = avatarUploadInput.value.files

	if (avatar.length === 0) {
		return
	}

	loading.value = true
	const reader = new FileReader()
	reader.onload = e => {
		avatarToCrop.value = e.target.result
		isCropAvatar.value = true
		// Note: loading stays true until Cropper's @ready event fires
		// This ensures the canvas is ready before allowing upload
	}
	reader.onerror = () => loading.value = false
	reader.readAsDataURL(avatar[0])
}
</script>

<style lang="scss">
.cropper {
	block-size: 80vh;
	background: transparent;
}

.vue-advanced-cropper__background {
	background: var(--white);
}
</style>

<style lang="scss" scoped>
// Ported from bulma-css-variables/sass/form/checkbox-radio.sass
// (the %checkbox-radio placeholder), scoped to this component so we can
// drop the global Bulma import.
label.radio {
	cursor: pointer;
	display: inline-block;
	line-height: 1.25;
	position: relative;

	input {
		cursor: pointer;
	}

	&:hover {
		color: var(--input-hover-color);
	}

	&[disabled],
	input[disabled] {
		color: var(--input-disabled-color);
		cursor: not-allowed;
	}

	& + .radio {
		margin-inline-start: .5em;
	}
}
</style>
