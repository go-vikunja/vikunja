<template>
	<card title="Avatar">
		<div class="control mb-4">
			<label class="radio">
				<input name="avatarProvider" type="radio" v-model="avatarProvider" value="default"/>
				Default
			</label>
			<label class="radio">
				<input name="avatarProvider" type="radio" v-model="avatarProvider" value="initials"/>
				Initials
			</label>
			<label class="radio">
				<input name="avatarProvider" type="radio" v-model="avatarProvider" value="gravatar"/>
				Gravatar
			</label>
			<label class="radio">
				<input name="avatarProvider" type="radio" v-model="avatarProvider" value="upload"/>
				Upload
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
				Upload Avatar
			</x-button>
			<template v-else>
				<cropper
					:src="avatarToCrop"
					:stencil-props="{aspectRatio: 1}"
					@ready="() => loading = false"
					class="mb-4"
					ref="cropper"/>
				<x-button
					:loading="avatarService.loading || loading"
					@click="uploadAvatar"
				>
					Upload Avatar
				</x-button>
			</template>
		</template>

		<div class="mt-2" v-if="avatarProvider !== 'upload'">
			<x-button
				:loading="avatarService.loading || loading"
				@click="updateAvatarStatus()"
				class="is-fullwidth"
			>
				Save
			</x-button>
		</div>
	</card>
</template>

<script>
import {Cropper} from 'vue-advanced-cropper'

import AvatarService from '../../services/avatar'
import AvatarModel from '../../models/avatar'

export default {
	name: 'avatar-settings',
	data() {
		return {
			avatarProvider: '',
			avatarService: AvatarService,
			isCropAvatar: false,
			avatarToCrop: null,
			loading: false, // Seperate variable because some things we're doing in browser take a bit
		}
	},
	created() {
		this.avatarService = new AvatarService()
		this.avatarStatus()
	},
	components: {
		Cropper,
	},
	methods: {
		avatarStatus() {
			this.avatarService.get({})
				.then(r => {
					this.avatarProvider = r.avatarProvider
				})
				.catch(e => this.error(e, this))
		},
		updateAvatarStatus() {
			const avatarStatus = new AvatarModel({avatarProvider: this.avatarProvider})
			this.avatarService.update(avatarStatus)
				.then(() => {
					this.success({message: 'Avatar status was updated successfully!'}, this)
					this.$store.commit('auth/reloadAvatar')
				})
				.catch(e => this.error(e, this))
		},
		uploadAvatar() {
			this.loading = true
			const {canvas} = this.$refs.cropper.getResult()

			if (canvas) {
				canvas.toBlob(blob => {
					this.avatarService.create(blob)
						.then(() => {
							this.success({message: 'The avatar has been set successfully!'}, this)
							this.$store.commit('auth/reloadAvatar')
						})
						.catch(e => this.error(e, this))
						.finally(() => {
							this.loading = false
							this.isCropAvatar = false
						})
				})
			} else {
				this.loading = false
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
