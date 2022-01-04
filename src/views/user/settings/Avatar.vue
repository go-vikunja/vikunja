<template>
	<card :title="$t('user.settings.avatar.title')">
		<div class="control mb-4">
			<label class="radio">
				<input name="avatarProvider" type="radio" v-model="avatarProvider" value="default"/>
				{{ $t('misc.default') }}
			</label>
			<label class="radio">
				<input name="avatarProvider" type="radio" v-model="avatarProvider" value="initials"/>
				{{ $t('user.settings.avatar.initials') }}
			</label>
			<label class="radio">
				<input name="avatarProvider" type="radio" v-model="avatarProvider" value="gravatar"/>
				{{ $t('user.settings.avatar.gravatar') }}
			</label>
			<label class="radio">
				<input name="avatarProvider" type="radio" v-model="avatarProvider" value="marble"/>
				{{ $t('user.settings.avatar.marble') }}
			</label>
			<label class="radio">
				<input name="avatarProvider" type="radio" v-model="avatarProvider" value="upload"/>
				{{ $t('user.settings.avatar.upload') }}
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
				:loading="avatarService.loading || loading"
				@click="$refs.avatarUploadInput.click()"
				v-if="!isCropAvatar">
				{{ $t('user.settings.avatar.uploadAvatar') }}
			</x-button>
			<template v-else>
				<cropper
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

		<div class="mt-2" v-if="avatarProvider !== 'upload'">
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

<script>
import {Cropper} from 'vue-advanced-cropper'
import 'vue-advanced-cropper/dist/style.css'

import AvatarService from '@/services/avatar'
import AvatarModel from '@/models/avatar'

export default {
	name: 'user-settings-avatar',
	data() {
		return {
			avatarProvider: '',
			avatarService: new AvatarService(),
			isCropAvatar: false,
			avatarToCrop: null,
			loading: false, // Seperate variable because some things we're doing in browser take a bit
		}
	},
	created() {
		this.avatarStatus()
	},
	components: {
		Cropper,
	},
	mounted() {
		this.setTitle(`${this.$t('user.settings.avatar.title')} - ${this.$t('user.settings.title')}`)
	},
	methods: {
		async avatarStatus() {
			const { avatarProvider } = await this.avatarService.get({})
			this.avatarProvider = avatarProvider
		},

		async updateAvatarStatus() {
			const avatarStatus = new AvatarModel({avatarProvider: this.avatarProvider})
			await this.avatarService.update(avatarStatus)
			this.$message.success({message: this.$t('user.settings.avatar.statusUpdateSuccess')})
			this.$store.commit('auth/reloadAvatar')
		},

		async uploadAvatar() {
			this.loading = true
			const {canvas} = this.$refs.cropper.getResult()

			if (!canvas) {
				this.loading = false
				return
			}

			try {
				const blob = await new Promise(resolve => canvas.toBlob(blob => resolve(blob)))
				await this.avatarService.create(blob)
				this.$message.success({message: this.$t('user.settings.avatar.setSuccess')})
				this.$store.commit('auth/reloadAvatar')
			} finally {
				this.loading = false
				this.isCropAvatar = false
			}
		},

		cropAvatar() {
			const avatar = this.$refs.avatarUploadInput.files

			if (avatar.length === 0) {
				return
			}

			this.loading = true
			const reader = new FileReader()
			reader.onload = e => {
				this.avatarToCrop = e.target.result
				this.isCropAvatar = true
			}
			reader.onloadend = () => this.loading = false
			reader.readAsDataURL(avatar[0])
		},
	},
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
