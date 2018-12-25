<template>
	<div>
		<h2 class="title">Reset your password</h2>
		<div class="box">
			<form id="form" @submit.prevent="submit" v-if="!successMessage">
				<div class="field">
					<div class="control">
						<input v-focus type="password" class="input" name="password1" placeholder="Password" v-model="credentials.password" required>
					</div>
				</div>
				<div class="field">
					<div class="control">
						<input type="password" class="input" name="password2" placeholder="Retype password" v-model="credentials.password2" required>
					</div>
				</div>

				<div class="field is-grouped">
					<div class="control">
						<button type="submit" class="button is-primary" v-bind:class="{ 'is-loading': loading}">Reset your password</button>
					</div>
				</div>
				<div class="notification is-info" v-if="loading">
					Loading...
				</div>
				<div class="notification is-danger" v-if="error">
					{{ error }}
				</div>
			</form>
			<div v-if="successMessage" class="has-text-centered">
				<div class="notification is-success">
					{{ successMessage }}
				</div>
				<router-link :to="{ name: 'login' }" class="button is-primary">Login</router-link>
			</div>
		</div>
	</div>
</template>

<script>
    import {HTTP} from '../../http-common'
	import message from '../../message'

    export default {
        data() {
            return {
                credentials: {
                    password: '',
                    password2: '',
                },
                error: '',
				successMessage: '',
                loading: false
            }
        },
        methods: {
            submit() {
				const cancel = message.setLoading(this)
                this.error = ''

                if (this.credentials.password2 !== this.credentials.password) {
                    cancel()
                    this.error = 'Passwords don\'t match'
                    return
                }

				let resetPasswordPayload = {
                    token: localStorage.getItem('passwordResetToken'),
					new_password: this.credentials.password
				}

                HTTP.post(`user/password/reset`, resetPasswordPayload)
                    .then(response => {
						this.handleSuccess(response)
                        localStorage.removeItem('passwordResetToken')
						cancel()
                    })
                    .catch(e => {
                        this.error = e.response.data.message
						cancel()
                    })
            },
            handleError(e) {
                this.error = e.response.data.message
            },
            handleSuccess(e) {
                this.successMessage = e.data.message
            }
        }
    }
</script>

<style scoped>
	.button {
		margin: 0 0.4em 0 0;
	}
</style>
