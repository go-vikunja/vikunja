<template>
	<div class="loader-container is-loading"></div>
</template>

<style scoped lang="scss">
.loader-container {
	height: 100%;
	min-height: 200px;
	width: 100%;
	min-width: 600px;
	max-width: 100vw;

	&.is-loading-small {
		min-height: 50px;
		min-width: 100px;
	}
}
</style>