<template>
	<div class="legal-links">
		<a :href="imprintUrl" v-if="imprintUrl" target="_blank">Imprint</a>
		<span v-if="imprintUrl && privacyPolicyUrl"> | </span>
		<a :href="privacyPolicyUrl" v-if="privacyPolicyUrl" target="_blank">Privacy policy</a>
	</div>
</template>

<script>
	import {mapState} from 'vuex'

	export default {
		name: 'legal',
		computed: mapState({
			imprintUrl: state => state.config.legal.imprintUrl,
			privacyPolicyUrl: state => state.config.legal.privacyPolicyUrl,
		})
	}
</script>
