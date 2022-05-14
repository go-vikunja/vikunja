<template>
	<card :title="$t('user.settings.avatar.title')">
		<div class="control mb-4">
			<label
				v-for="(label, providerId) in AVATAR_PROVIDERS"
				:key="providerId"
				class="radio"
			>
				<input name="avatarProvider" type="radio" v-model="avatarProvider" :value="providerId"/>
				{{ label }}
			</label>
		</div>

		<template v-if="avatarProvider === 'upload'">
			<input
				@change="cropAvatar"
				accept="image/*"
				class="is-hidden"
				ref="avatarUploadInput"
				type="file"
			/>

			<x-button
				v-if="!isCropAvatar"
				:loading="avatarService.loading || loading"
				@click="$refs.avatarUploadInput.click()"
			>
				{{ $t('user.settings.avatar.uploadAvatar') }}
			</x-button>
			<template v-else>
				<Cropper
					:src="avatarToCrop"
					:stencil-props="{aspectRatio: 1}"
					@ready="() => loading = false"
					class="mb-4 cropper"
					ref="cropper"
				/>
				<x-button
					:loading="avatarService.loading || loading"
					@click="uploadAvatar"
					v-cy="'uploadAvatar'"
				>
					{{ $t('user.settings.avatar.uploadAvatar') }}
				</x-button>
			</template>
		</template>

		<div class="mt-2" v-else>
			<x-button
				:loading="avatarService.loading || loading"
				@click="updateAvatarStatus()"
				class="is-fullwidth"
			>
				{{ $t('misc.save') }}
			</x-button>
		</div>
	</card>
</template>

<script lang="ts">
import {defineComponent} from 'vue'

export default defineComponent({
	name: 'user-settings-avatar',
})
</script>

<script setup lang="ts">
import {ref, shallowReactive} from 'vue'
import {useI18n} from 'vue-i18n'
import {useStore} from 'vuex'
import {Cropper} from 'vue-advanced-cropper'
import 'vue-advanced-cropper/dist/style.css'

import AvatarService from '@/services/avatar'
import AvatarModel from '@/models/avatar'
import { useTitle } from '@/composables/useTitle'
import { success } from '@/message'

const {t} = useI18n()
const store = useStore()

const AVATAR_PROVIDERS = {
	default: t('misc.default'),
	initials: t('user.settings.avatar.initials'),
	gravatar: t('user.settings.avatar.gravatar'),
	marble: t('user.settings.avatar.marble'),
	upload: t('user.settings.avatar.upload'),
}

useTitle(() => `${t('user.settings.avatar.title')} - ${t('user.settings.title')}`)

const avatarService = shallowReactive(new AvatarService())
// Seperate variable because some things we're doing in browser take a bit
const loading = ref(false)


const avatarProvider = ref('')
async function avatarStatus() {
	const { avatarProvider: currentProvider } = await avatarService.get({})
	avatarProvider.value = currentProvider
}
avatarStatus()


async function updateAvatarStatus() {
	await avatarService.update(new AvatarModel({avatarProvider: avatarProvider.value}))
	success({message: t('user.settings.avatar.statusUpdateSuccess')})
	store.commit('auth/reloadAvatar')
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
		const blob = await new Promise(resolve => canvas.toBlob(blob => resolve(blob)))
		await avatarService.create(blob)
		success({message: t('user.settings.avatar.setSuccess')})
		store.commit('auth/reloadAvatar')
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
	}
	reader.onloadend = () => loading.value = false
	reader.readAsDataURL(avatar[0])
}
</script>

<style lang="scss">
.cropper {
	height: 80vh;
	background: transparent;
}

.vue-advanced-cropper__background {
	background: var(--white);
}
</style>
