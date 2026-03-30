const {contextBridge, ipcRenderer} = require('electron')

contextBridge.exposeInMainWorld('vikunjaDesktop', {
	startOAuthLogin: (apiUrl) => ipcRenderer.invoke('oauth:start-login', apiUrl),
	onOAuthTokens: (callback) => {
		ipcRenderer.removeAllListeners('oauth:tokens')
		ipcRenderer.on('oauth:tokens', (_event, tokens) => callback(tokens))
	},
	onOAuthError: (callback) => {
		ipcRenderer.removeAllListeners('oauth:error')
		ipcRenderer.on('oauth:error', (_event, error) => callback(error))
	},
	refreshToken: (apiUrl, refreshToken) => ipcRenderer.invoke('oauth:refresh-token', apiUrl, refreshToken),
	isDesktop: true,
})
