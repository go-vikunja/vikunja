const {
	app,
	BrowserWindow,
	globalShortcut,
	ipcMain,
	Menu,
	nativeImage,
	shell,
	Tray,
	screen,
} = require('electron')
const path = require('path')
const express = require('express')
const portInUse = require('./portInUse.js')

const frontendPath = 'frontend/'
const SAFE_PROTOCOLS = new Set([
	'http:', 'https:', 'mailto:',
	'ftp:', 'git:', 'obsidian:', 'notion:', 'message:',
])

function safeOpenExternal(url) {
	try {
		const parsed = new URL(url)
		if (SAFE_PROTOCOLS.has(parsed.protocol)) {
			shell.openExternal(url)
		}
	} catch {
		// Ignore malformed URLs
	}
}

// Module-scope state
let mainWindow = null
let quickEntryWindow = null
let tray = null
let serverPort = null
let isQuitting = false

// ─── Express server ──────────────────────────────────────────────────
function startServer(callback) {
	const eApp = express()
	let port = 45735

	portInUse(port, (used) => {
		if (used) {
			console.log(`Port ${port} already used, switching to a random one`)
			port = 0
		}

		eApp.use(express.static(path.join(__dirname, frontendPath)))
		eApp.use((request, response) => {
			response.sendFile(path.join(__dirname, frontendPath, 'index.html'))
		})

		const server = eApp.listen(port, '127.0.0.1', () => {
			serverPort = server.address().port
			console.log(`Server started on port ${serverPort}`)
			callback(serverPort)
		})
	})
}

// ─── Main window ─────────────────────────────────────────────────────
function createMainWindow() {
	mainWindow = new BrowserWindow({
		width: 1680,
		height: 960,
		webPreferences: {
			nodeIntegration: false,
			contextIsolation: true,
			sandbox: true,
			webviewTag: false,
			navigateOnDragDrop: false,
		},
	})

	mainWindow.webContents.setWindowOpenHandler(({url}) => {
		safeOpenExternal(url)
		return {action: 'deny'}
	})

	// Prevent same-window navigation to external origins.
	// Only allow navigation to the local express server.
	mainWindow.webContents.on('will-navigate', (event, navigationUrl) => {
		const parsedUrl = new URL(navigationUrl)
		// Allow navigations to the local express server
		if (parsedUrl.hostname === '127.0.0.1' || parsedUrl.hostname === 'localhost') {
			return
		}
		event.preventDefault()
	})

	mainWindow.setMenuBarVisibility(false)

	mainWindow.on('close', (e) => {
		if (!isQuitting && tray) {
			e.preventDefault()
			mainWindow.hide()
		}
	})

	mainWindow.on('closed', () => {
		mainWindow = null
	})

	mainWindow.loadURL(`http://127.0.0.1:${serverPort}`)
}

// ─── Quick Entry window ──────────────────────────────────────────────
function createQuickEntryWindow() {
	const display = screen.getPrimaryDisplay()
	const {width: screenWidth, height: screenHeight} = display.workAreaSize
	const winWidth = 680
	const winHeight = 500

	quickEntryWindow = new BrowserWindow({
		width: winWidth,
		height: winHeight,
		x: Math.round((screenWidth - winWidth) / 2),
		y: Math.round(screenHeight / 3 - winHeight / 2),
		frame: false,
		transparent: true,
		alwaysOnTop: true,
		skipTaskbar: true,
		resizable: false,
		show: false,
		webPreferences: {
			nodeIntegration: false,
			contextIsolation: true,
			preload: path.join(__dirname, 'preload-quick-entry.js'),
		},
	})

	quickEntryWindow.webContents.setWindowOpenHandler(({url}) => {
		safeOpenExternal(url)
		return {action: 'deny'}
	})

	quickEntryWindow.loadURL(`http://127.0.0.1:${serverPort}/?mode=quick-add`)

	// Hide on blur (user clicked outside)
	let blurTimeout = null
	quickEntryWindow.on('blur', () => {
		// Debounce to avoid hiding during DevTools focus changes
		blurTimeout = setTimeout(() => hideQuickEntry(), 100)
	})
	quickEntryWindow.on('focus', () => {
		if (blurTimeout) {
			clearTimeout(blurTimeout)
			blurTimeout = null
		}
	})

	quickEntryWindow.on('closed', () => {
		quickEntryWindow = null
	})
}

function showQuickEntry() {
	if (!quickEntryWindow) {
		createQuickEntryWindow()
		quickEntryWindow.once('ready-to-show', () => {
			quickEntryWindow.show()
			quickEntryWindow.focus()
		})
		return
	}

	// Reload to reset Vue state (clear previous input)
	quickEntryWindow.loadURL(`http://127.0.0.1:${serverPort}/?mode=quick-add`)
	quickEntryWindow.show()
	quickEntryWindow.focus()
}

function hideQuickEntry() {
	if (quickEntryWindow && quickEntryWindow.isVisible()) {
		quickEntryWindow.hide()
	}
}

function toggleQuickEntry() {
	if (quickEntryWindow && quickEntryWindow.isVisible()) {
		hideQuickEntry()
	} else {
		showQuickEntry()
	}
}

// ─── System tray ─────────────────────────────────────────────────────
function setupTray() {
	const iconPath = path.join(__dirname, 'build', 'icon.png')
	const icon = nativeImage.createFromPath(iconPath).resize({width: 16, height: 16})
	tray = new Tray(icon)

	const contextMenu = Menu.buildFromTemplate([
		{
			label: 'Show Vikunja',
			click: () => {
				if (mainWindow) {
					mainWindow.show()
					mainWindow.focus()
				} else {
					createMainWindow()
				}
			},
		},
		{
			label: 'Quick Add Task',
			accelerator: 'CmdOrCtrl+Shift+A',
			click: () => showQuickEntry(),
		},
		{type: 'separator'},
		{
			label: 'Quit',
			click: () => {
				isQuitting = true
				app.quit()
			},
		},
	])

	tray.setToolTip('Vikunja')
	tray.setContextMenu(contextMenu)

	tray.on('click', () => {
		if (mainWindow) {
			mainWindow.show()
			mainWindow.focus()
		} else {
			createMainWindow()
		}
	})
}

// ─── IPC handlers ────────────────────────────────────────────────────
ipcMain.on('quick-entry:close', () => {
	hideQuickEntry()
})

// ─── App lifecycle ───────────────────────────────────────────────────
app.whenReady().then(() => {
	startServer((port) => {
		createMainWindow()
		createQuickEntryWindow()
		setupTray()

		const registered = globalShortcut.register('CmdOrCtrl+Shift+A', toggleQuickEntry)
		if (!registered) {
			console.warn('Failed to register global shortcut CmdOrCtrl+Shift+A — it may be in use by another application')
		}
	})

	app.on('activate', () => {
		if (BrowserWindow.getAllWindows().length === 0) {
			if (serverPort) {
				createMainWindow()
			}
		} else if (mainWindow) {
			mainWindow.show()
			mainWindow.focus()
		}
	})
})

app.on('before-quit', () => {
	isQuitting = true
})

app.on('will-quit', () => {
	globalShortcut.unregisterAll()
})

app.on('window-all-closed', () => {
	// Don't quit if tray exists (user can still use global shortcut)
	if (process.platform !== 'darwin' && !tray) {
		app.quit()
	}
})
