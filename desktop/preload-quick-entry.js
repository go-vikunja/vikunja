// desktop/preload-quick-entry.js
const { contextBridge, ipcRenderer } = require('electron')

contextBridge.exposeInMainWorld('quickEntry', {
	close: () => ipcRenderer.send('quick-entry:close'),
	resize: (width, height) => ipcRenderer.send('quick-entry:resize', width, height),
	showMainWindow: () => ipcRenderer.send('quick-entry:show-main-window'),
})
