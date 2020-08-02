<template>
	<div class="card">
		<header class="card-header">
			<p class="card-header-title">
				Avatar
			</p>
		</header>
		<div class="card-content">
			<div class="control mb-4">
				<label class="radio">
					<input type="radio" name="avatarProvider" v-model="avatarProvider" value="default"/>
					Default
				</label>
				<label class="radio">
					<input type="radio" name="avatarProvider" v-model="avatarProvider" value="initials"/>
					Initials
				</label>
				<label class="radio">
					<input type="radio" name="avatarProvider" v-model="avatarProvider" value="gravatar"/>
					Gravatar
				</label>
				<label class="radio">
					<input type="radio" name="avatarProvider" v-model="avatarProvider" value="upload"/>
					Upload
				</label>
			</div>

			<template v-if="avatarProvider === 'upload'">
				<input
						type="file"
						ref="avatarUploadInput"
						@change="cropAvatar"
						class="is-hidden"
						accept="image/*"
				/>
				<a
						v-if="!isCropAvatar"
						class="button is-primary"
						@click="$refs.avatarUploadInput.click()"
						:class="{ 'is-loading': avatarService.loading || loading}">
					Upload Avatar
				</a>
				<template v-else>
					<cropper
							:src="avatarToCrop"
							class="mb-4"
							@ready="() => loading = false"
							:stencil-props="{aspectRatio: 1}"
							ref="cropper"/>
					<a
							class="button is-primary"
							@click="uploadAvatar"
							:class="{ 'is-loading': avatarService.loading || loading}">
						Upload Avatar
					</a>
				</template>
			</template>

			<div class="bigbuttons" v-if="avatarProvider !== 'upload'">
				<button @click="updateAvatarStatus()" class="button is-primary is-fullwidth"
						:class="{ 'is-loading': avatarService.loading || loading}">
					Save
				</button>
			</div>
		</div>
	</div>
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
