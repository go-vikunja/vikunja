<template>
	<div>
	<h3>Hiiiii</h3>
		<span v-if="authenticated">Logged in</span>
	</div>
</template>

<script>
    import auth from '../auth'
    import router from '../router'

    export default {
        name: "Home",
		data() {
            return {
                authenticated: auth.user.authenticated
            }
		},
        beforeMount() {
            // Check if the user is already logged in, if so, redirect him to the homepage
            if (!auth.user.authenticated) {
                router.push({name: 'login'})
            }
        },
    }
</script>

<style scoped>

</style>