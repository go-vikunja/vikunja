<template>
	<div class="content has-text-centered">
		<h2>Hi {{user.infos.username}}!</h2>
		<p>Click on a list or namespace on the left to get started.</p>
	</div>
</template>

<script>
    import auth from '../auth'
    import router from '../router'

    export default {
        name: "Home",
        data() {
            return {
                user: auth.user
            }
        },
        beforeMount() {
            // Check if the user is already logged in, if so, redirect him to the homepage
            if (!auth.user.authenticated) {
                router.push({name: 'login'})
            }
        },
        methods: {
            logout() {
                auth.logout()
            }
        },
    }
</script>

<style scoped>

</style>