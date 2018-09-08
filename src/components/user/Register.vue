<template>
	<div class="container has-text-centered">
		<div class="column is-4 is-offset-4">
			<h2 class="title">Register</h2>
			<div class="box">
				<form id="registerform" @submit.prevent="submit">
					<div class="field">
						<div class="control">
							<input type="text" class="input" name="username" placeholder="Username" v-model="credentials.username" required>
						</div>
					</div>
					<div class="field">
						<div class="control">
							<input type="text" class="input" name="email" placeholder="E-mail address" v-model="credentials.email" required>
						</div>
					</div>
					<div class="field">
						<div class="control">
							<input type="password" class="input" name="password1" placeholder="Password" v-model="credentials.password" required>
						</div>
					</div>
					<div class="field">
						<div class="control">
							<input type="password" class="input" name="password2" placeholder="Retype password" v-model="credentials.password2" required>
						</div>
					</div>

					<div class="field is-grouped">
						<div class="control">
							<button type="submit" class="button is-link">Register</button>
							<router-link :to="{ name: 'login' }" class="button">Login</router-link>
						</div>
					</div>
					<div class="notification is-info" v-if="loading">
						Loading...
					</div>
					<div class="notification is-danger" v-if="error">
						{{ error }}
					</div>
				</form>

			</div>
		</div>
	</div>
</template>

<script>
    import auth from '../../auth'
    import router from '../../router'

    export default {
        data() {
            return {
                credentials: {
                    username: '',
					email: '',
                    password: '',
                    password2: '',
                },
                error: '',
                loading: false
            }
        },
        beforeMount() {
            // Check if the user is already logged in, if so, redirect him to the homepage
            if (auth.user.authenticated) {
                router.push({name: 'home'})
            }
        },
        methods: {
            submit() {
                this.loading = true

                this.error = ''

				if (this.credentials.password2 !== this.credentials.password) {
                    this.loading = false
                    this.error = 'Passwords don\'t match'
					return
				}

                let credentials = {
                    username: this.credentials.username,
                    email: this.credentials.email,
                    password: this.credentials.password
                }

                auth.register(this, credentials, 'home')
            }
        }
    }
</script>

<style scoped>
	.button {
		margin: 0 0.4em 0 0;
	}
</style>
