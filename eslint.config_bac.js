// import originalEslintPluginVue from 'eslint-plugin-vue'
import vueTsRecommended from '@vue/eslint-config-typescript/recommended.js'
// import vueParser from 'vue-eslint-parser'
import tsParser from "@typescript-eslint/parser"

const vue3Recommended = vue.configs['vue3-recommended']

import {default as originalVuePlugin} from "eslint-plugin-vue";

// see https://github.com/eslint/eslint/issues/16875#issuecomment-1426594123
const eslintPluginVue = {
	...originalVuePlugin,
	parsers: {
  	'parser': {
			parseForESLint: originalVuePlugin.parseForESLint
		}
	}
}

// export default [{
//   files:   ["**/*.json", "**/*.jsonc", "**/*.json5"],
//   plugins: {
//     vue: { ...vue, parsers}
//     /* same as
//     jsonc: {
//       parsers: {
//         'jsonc-eslint-parser': {
//           parseForESLint
//         }
//       }
//     } */
//   },
//   languageOptions: {
//      parser: 'vue/vue-eslint-parser'
//   },
//   rules: {...}
// }];

export default [
	// 'eslint:recommended',
	{
		files:   ["**/*.vue"],
		plugins: {
			vue: eslintPluginVue,
		},
		languageOptions: {
			parser: 'vue/parser'
	 },
	},
	// {
	// 	plugins: {
	// 		// vue: vue3Recommended,
	// 		// '@typescript-eslint': vueTsRecommended,
	// 	},
	// 	languageOptions: {
	// 		// parser: eslintPluginVue,
	// 		// parser: 'vue/vue-eslint-parser',
	// 		parserOptions: {
	// 			parser: '@typescript-eslint/parser',
	// 			// 'ecmaVersion': 2022,
	// 			// 'sourceType': 'module',
	// 		},
	// 	}
	// }
	// {
	// 	files: ["./src/**/*.vue"],
	// 	// files: ["./src/**/*.js"],
	// 	// ignores: ["**/*.config.js"],
	// 	rules: {
	// 			semi: "error"
	// 	},
	// 	plugins: {
	// 		vue: vue3Recommended,
	// 		// '@typescript-eslint': vueTsRecommended,
	// 	},
	// },
	// {
	// 	files: ["src/**/*.vue"],
	// 	// files: [
	// 	// 	'src/**/*.vue',
	// 	// 	'src/**/*.js',
	// 	// 	'src/**/*.ts',
	// 	// 	// 'src/**/*.+(vue|js|ts)',
	// 	// ],
	// 	ignores: [
	// 		'*.test.*',
	// 		'cypress/*',
	// 	],
	// 	plugins: {
	// 		vue: vue3Recommended,
	// 		'@typescript-eslint': vueTsRecommended,
	// 	},
	// 	rules: {
	// 		'vue/html-quotes': ['error', 'double'],
	// 		'quotes': ['error', 'single'],
	// 		'comma-dangle': ['error', 'always-multiline'],
	// 		'semi': ['error', 'never'],
	// 		'vue/multi-word-component-names': 0,
	// 		// disabled until we have support for reactivityTransform
	// 		// See https://github.com/vuejs/eslint-plugin-vue/issues/1948
	// 		// see also setting in `vite.config`
	// 		'vue/no-setup-props-destructure': 0,
	// 	},
	// 	// overwrite the following with correct values
	// 	// eslint-plugin-vue/lib/configs/base.js

	// 	// parser: 
	// 	parserOptions: {
	// 		ecmaVersion: 2022,


	// 		'parser': '@typescript-eslint/parser',
	// 		'sourceType': 'module',
	// 	},
	// 	globals: {
	// 		'browser': true,
	// 		'es2022': true,
	// 		'node': true,
	// 		'vue/setup-compiler-macros': true,
	// 	}
	// },
]