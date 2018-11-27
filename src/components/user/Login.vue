<template>
	<div>
		<h2 class="title">Login</h2>
		<div class="box">
			<div v-if="confirmedEmailSuccess" class="notification is-success has-text-centered">
				You successfully confirmed your email! You can log in now.
			</div>
			<form id="loginform" @submit.prevent="submit">
				<div class="field">
					<div class="control">
						<input type="text" class="input" name="username" placeholder="Username" v-model="credentials.username" required>
					</div>
				</div>
				<div class="field">
					<div class="control">
						<input type="password" class="input" name="password" placeholder="Password" v-model="credentials.password" required>
					</div>
				</div>

				<div class="field is-grouped">
					<div class="control">
						<button type="submit" class="button is-primary" v-bind:class="{ 'is-loading': loading}">Login</button>
						<router-link :to="{ name: 'register' }" class="button">Register</router-link>
						<router-link :to="{ name: 'getPasswordReset' }" class="reset-password-link">Reset your password</router-link>
					</div>
				</div>
				<div class="notification is-danger" v-if="error">
					{{ error }}
				</div>
			</form>
		</div>
	</div>
</template>

<script>
    import auth from '../../auth'
    import router from '../../router'
    import {HTTP} from '../../http-common'
	import message from '../../message'

    export default {
        data() {
            return {
                credentials: {
                    username: '',
                    password: ''
                },
                error: '',
                confirmedEmailSuccess: false,
                loading: false
            }
        },
        beforeMount() {
            // Try to verify the email
			let emailVerifyToken = localStorage.getItem('emailConfirmToken')
			if (emailVerifyToken) {
				const cancel = message.setLoading(this)
				HTTP.post(`user/confirm`, {token: emailVerifyToken})
					.then(() => {
                        localStorage.removeItem('emailConfirmToken')
                        this.confirmedEmailSuccess = true
						cancel()
                    })
                    .catch(e => {
                        cancel()
                        this.error = e.response.data.message
                    })
			}

            // Check if the user is already logged in, if so, redirect him to the homepage
            if (auth.user.authenticated) {
                router.push({name: 'home'})
            }
        },
        methods: {
            submit() {
                this.loading = true
                this.error = ''
                let credentials = {
                    username: this.credentials.username,
                    password: this.credentials.password
                }

                auth.login(this, credentials, 'home')
            }
        }
    }
</script>

<style scoped>
	.button {
		margin: 0 0.4em 0 0;
	}

	.reset-password-link{
		display: inline-block;
		padding-top: 5px;
	}
</style>
