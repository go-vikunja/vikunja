<template>
	<div class="loader-container" v-bind:class="{ 'is-loading': passwordUpdateService.loading}">
		<div class="card">
			<header class="card-header">
				<p class="card-header-title">
					Update Your Password
				</p>
			</header>
			<div class="card-content">
				<div class="content">
					<form @submit.prevent="updatePassword()">
						<div class="field">
							<label class="label" for="newPassword">New Password</label>
							<div class="control">
								<input class="input" type="password" id="newPassword" placeholder="The new password..."
									v-model="passwordUpdate.newPassword" @keyup.enter="updatePassword"/>
							</div>
						</div>
						<div class="field">
							<label class="label" for="newPasswordConfirm">New Password Confirmation</label>
							<div class="control">
								<input class="input" type="password" id="newPasswordConfirm" placeholder="Confirm your new password..."
									v-model="passwordConfirm" @keyup.enter="updatePassword"/>
							</div>
						</div>
						<div class="field">
							<label class="label" for="currentPassword">Current Password</label>
							<div class="control">
								<input class="input" type="password" id="currentPassword" placeholder="Your current password"
									v-model="passwordUpdate.oldPassword" @keyup.enter="updatePassword"/>
							</div>
						</div>
					</form>

					<div class="bigbuttons">
						<button @click="updatePassword()" class="button is-primary is-fullwidth"
								:class="{ 'is-loading': passwordUpdateService.loading}">
							Save
						</button>
					</div>
				</div>
			</div>
		</div>
		<div class="card">
			<header class="card-header">
				<p class="card-header-title">
					Update Your E-Mail Address
				</p>
			</header>
			<div class="card-content">
				<div class="content">
					<form @submit.prevent="updateEmail()">
						<div class="field">
							<label class="label" for="newEmail">New Email Address</label>
							<div class="control">
								<input class="input" type="email" id="newEmail" placeholder="The new email address..."
									v-model="emailUpdate.newEmail" @keyup.enter="updateEmail"/>
							</div>
						</div>
						<div class="field">
							<label class="label" for="currentPassword">Current Password</label>
							<div class="control">
								<input class="input" type="password" id="currentPassword" placeholder="Your current password"
									v-model="emailUpdate.password" @keyup.enter="updateEmail"/>
							</div>
						</div>
					</form>

					<div class="bigbuttons">
						<button @click="updateEmail()" class="button is-primary is-fullwidth"
								:class="{ 'is-loading': emailUpdateService.loading}">
							Save
						</button>
					</div>
				</div>
			</div>
		</div>
	</div>
</template>

<script>
	import PasswordUpdateModel from '../../models/passwordUpdate'
	import PasswordUpdateService from '../../services/passwordUpdateService'
	import EmailUpdateService from '../../services/emailUpdate'
	import EmailUpdateModel from '../../models/emailUpdate'

	export default {
		name: 'Settings',
		data() {
			return {
				passwordUpdateService: PasswordUpdateService,
				passwordUpdate: PasswordUpdateModel,
				passwordConfirm: '',

				emailUpdateService: EmailUpdateService,
				emailUpdate: EmailUpdateModel,
			}
		},
		created() {
			this.passwordUpdate = new PasswordUpdateModel()
			this.passwordUpdateService = new PasswordUpdateService()

			this.emailUpdate = new EmailUpdateModel()
			this.emailUpdateService = new EmailUpdateService()
		},
		methods: {
			updatePassword() {
				if (this.passwordConfirm !== this.passwordUpdate.newPassword) {
					this.error({message: 'The new password and its confirmation don\'t match.'}, this)
					return
				}

				this.passwordUpdateService.update(this.passwordUpdate)
					.then(() => {
						this.success({message: 'The password was successfully updated.'}, this)
					})
					.catch(e => this.error(e, this))
			},
			updateEmail() {
				this.emailUpdateService.update(this.emailUpdate)
					.then(() => {
						this.success({message: 'Your email address was successfully updated. We\'ve sent you a link to confirm it.'}, this)
					})
					.catch(e => this.error(e, this))
			},
		},
	}
</script>
