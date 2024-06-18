/* eslint-env node */
require('@rushstack/eslint-patch/modern-module-resolution')

module.exports = {
	'root': true,
	'env': {
		'browser': true,
		'es2022': true,
		'node': true,
	},
	'extends': [
		'eslint:recommended',
		'plugin:vue/vue3-recommended',
		'@vue/eslint-config-typescript/recommended',
	],
	'rules': {
		'quotes': ['error', 'single'],
		'comma-dangle': ['error', 'always-multiline'],
		'semi': ['error', 'never'],

		'vue/v-on-event-hyphenation': ['warn', 'never', { 'autofix': true }],
		'vue/multi-word-component-names': 'off',

		// uncategorized rules:
		'vue/component-api-style': ['error', ['script-setup']],
		'vue/component-name-in-template-casing': ['error', 'PascalCase', {
			'globals': ['RouterView', 'RouterLink', 'Icon', 'Notifications', 'Modal', 'Card'],
		}],
		'vue/custom-event-name-casing': ['error', 'camelCase'],
		'vue/define-macros-order': 'error',
		'vue/match-component-file-name': ['error', {
			'extensions': ['.js', '.jsx', '.ts', '.tsx', '.vue'],
			'shouldMatchCase': true,
		}],
		'vue/no-boolean-default': ['warn', 'default-false'],
		'vue/match-component-import-name': 'error',
		'vue/prefer-separate-static-class': 'warn',

		'vue/padding-line-between-blocks': 'error',
		'vue/next-tick-style': ['error', 'promise'],
		'vue/block-lang': [
			'error',
			{ 'script': { 'lang': 'ts' } },
		],
		'vue/no-required-prop-with-default': ['error', { 'autofix': true }],
		'vue/no-duplicate-attr-inheritance': 'error',
		'vue/no-empty-component-block': 'error',
		'vue/html-indent': ['error', 'tab'],

		// vue3
		'vue/no-ref-object-reactivity-loss': 'error',
		'vue/no-setup-props-reactivity-loss': 'warn', // TODO: switch to error after vite `propsDestructure` is removed
	},
	'parser': 'vue-eslint-parser',
	'parserOptions': {
		'parser': '@typescript-eslint/parser',
		'ecmaVersion': 'latest',
		'tsconfigRootDir': __dirname,
	},
	'ignorePatterns': [
		'*.test.*',
		'cypress/*',
	],
}