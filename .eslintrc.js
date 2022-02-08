module.exports = {
	'root': true,
	'env': {
		'browser': true,
		'es2021': true,
		'node': true,
		'vue/setup-compiler-macros': true,
	},
	'extends': [
		'eslint:recommended',
		'plugin:vue/vue3-essential',
		'@vue/typescript',
	],
	'rules': {
		'vue/html-quotes': [
			'error',
			'double',
		],
		'quotes': [
			'error',
			'single',
		],
		'comma-dangle': [
			'error',
			'always-multiline',
		],
		'semi': [
			'error',
			'never',
		],
		'vue/script-setup-uses-vars': 'error',

		// see https://segmentfault.com/q/1010000040813116/a-1020000041134455 (original in chinese)
		'no-unused-vars': 'off',
		'@typescript-eslint/no-unused-vars': ['error', { vars: 'all', args: 'after-used', ignoreRestSiblings: true }],

		'vue/multi-word-component-names': 0,
	},
	'parser': 'vue-eslint-parser',
	'parserOptions': {
		'parser': '@typescript-eslint/parser',
		'ecmaVersion': 2022,
	},
	'ignorePatterns': [
		'*.test.*',
		'cypress/*',
	],
	'globals': {
		'defineProps': 'readonly',
	},
}