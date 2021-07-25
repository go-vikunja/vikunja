import { nodeResolve } from '@rollup/plugin-node-resolve'
import babel from '@rollup/plugin-babel'
import commonjs from '@rollup/plugin-commonjs'
import { terser } from 'rollup-plugin-terser'
import replace from '@rollup/plugin-replace'

// Building the sw stuff ourselves to have more controle over it - see https://github.com/antfu/vite-plugin-pwa/issues/35#issuecomment-797942573
export default {
	input: './dist/sw.js',
	output: {
		dir: 'dist',
		format: 'esm',
	},
	plugins: [
		replace({
			'process.env.NODE_ENV': JSON.stringify('production'),
			'preventAssignment': true,
		}),
		nodeResolve({
			browser: true,
		}),
		commonjs(),
		babel({
			exclude: '**/node_modules/**',
			extensions: ['js'],
			babelHelpers: 'runtime',
			presets: [
				[
					'@babel/preset-env',
					{
						corejs: 3,
						useBuiltIns: 'entry',
						targets: {
							'esmodules': true,
						},
						modules: false,
					},
				],
			],
		}),
		terser(),
	],
}
