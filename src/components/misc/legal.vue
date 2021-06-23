<template>
	<div class="legal-links">
		<a :href="imprintUrl" target="_blank" v-if="imprintUrl">{{ $t('navigation.imprint') }}</a>
		<span v-if="imprintUrl && privacyPolicyUrl"> | </span>
		<a :href="privacyPolicyUrl" target="_blank" v-if="privacyPolicyUrl">{{ $t('navigation.privacy') }}</a>
	</div>
</template>

<script>
import {mapState} from 'vuex'

export default {
	name: 'legal',
	computed: mapState({
		imprintUrl: state => state.config.legal.imprintUrl,
		privacyPolicyUrl: state => state.config.legal.privacyPolicyUrl,
	}),
}
</script>
