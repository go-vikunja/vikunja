// in order to use postcss-preset-env correctly we need some client side plugins
import focusWithinInit from 'postcss-focus-within/browser'
import cssHasPseudo from 'css-has-pseudo/browser'

focusWithinInit()
cssHasPseudo(document)