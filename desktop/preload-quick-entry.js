// desktop/preload-quick-entry.js
const { contextBridge, ipcRenderer } = require('electron')

contextBridge.exposeInMainWorld('quickEntry', {
	close: () => ipcRenderer.send('quick-entry:close'),
})
