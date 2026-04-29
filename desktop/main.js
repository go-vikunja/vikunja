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
const fs = require('fs')
const express = require('express')
const portInUse = require('./portInUse.js')
const oauth = require('./oauth.js')

const frontendPath = 'frontend/'
const PROTOCOL = 'vikunja-desktop'
const SAFE_PROTOCOLS = new Set([
	'http:', 'https:', 'mailto:',
	'ftp:', 'git:', 'obsidian:', 'notion:', 'message:',
])

const QUICK_ENTRY_WIDTH = 680
const QUICK_ENTRY_COLLAPSED_HEIGHT = 56

const ZOOM_STEP = 0.5
const ZOOM_CONFIG_FILE = 'zoom.json'

const BASE_WEB_PREFERENCES = {
	nodeIntegration: false,
	contextIsolation: true,
	sandbox: true,
	webviewTag: false,
	navigateOnDragDrop: false,
}

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
let pendingDeepLinkUrl = null
let pendingApiUrl = null
let currentShortcut = null
let zoomLevel = 0
const DEFAULT_QUICK_ENTRY_SHORTCUT = 'CmdOrCtrl+Shift+A'
const launchedWithQuickEntry = process.argv.includes('--quick-entry')

// Ensure single instance so deep links reach the running app on Windows/Linux
const gotTheLock = app.requestSingleInstanceLock()
if (!gotTheLock) {
	app.quit()
	// Must exit the process immediately — app.quit() is async and the rest of this
	// file would still execute, potentially opening a blank window.
	process.exit(0)
}

// Register the custom protocol for deep links
if (process.defaultApp) {
	// During development, register with the path to the script
	if (process.argv.length >= 2) {
		app.setAsDefaultProtocolClient(PROTOCOL, process.execPath, [path.resolve(process.argv[1])])
	}
} else {
	app.setAsDefaultProtocolClient(PROTOCOL)
}

// Handle deep link on macOS (app already running or launched via URL)
app.on('open-url', (event, url) => {
	event.preventDefault()
	if (mainWindow) {
		handleDeepLink(url)
	} else {
		// Window not ready yet — buffer the URL for processing after createMainWindow()
		pendingDeepLinkUrl = url
	}
})

// Handle deep link on Windows/Linux when a second instance is launched
app.on('second-instance', (_event, argv) => {
	// Handle --quick-entry flag from second instance
	if (argv.includes('--quick-entry')) {
		if (serverPort) {
			toggleQuickEntry()
		}
		return
	}

	// Focus the main window
	if (mainWindow) {
		if (mainWindow.isMinimized()) mainWindow.restore()
		mainWindow.focus()
	}

	// Find the deep link URL in argv
	const deepLinkUrl = argv.find(arg => arg.startsWith(`${PROTOCOL}://`))
	if (deepLinkUrl) {
		handleDeepLink(deepLinkUrl)
	}
})

function handleDeepLink(url) {
	try {
		const parsed = new URL(url)
		if (parsed.hostname === 'callback') {
			const code = parsed.searchParams.get('code')
			if (code && mainWindow) {
				// Store the apiUrl that was used to start login so we can
				// exchange the code at the correct endpoint
				const apiUrl = pendingApiUrl
				if (!apiUrl) {
					mainWindow.webContents.send('oauth:error', 'No pending login session')
					return
				}

				oauth.exchangeCodeForTokens(apiUrl, code)
					.then(tokens => {
						mainWindow.webContents.send('oauth:tokens', tokens)
					})
					.catch(err => {
						mainWindow.webContents.send('oauth:error', err.message)
					})
			}
		}
	} catch {
		// Invalid URL, ignore
	}
}

// IPC: Start OAuth login flow
ipcMain.handle('oauth:start-login', async (_event, apiUrl) => {
	pendingApiUrl = apiUrl
	const authUrl = oauth.startLogin(apiUrl)
	await shell.openExternal(authUrl)
})

// IPC: Refresh access token
ipcMain.handle('oauth:refresh-token', async (_event, apiUrl, refreshToken) => {
	return oauth.refreshAccessToken(apiUrl, refreshToken)
})

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

// ─── Zoom ────────────────────────────────────────────────────────────
function zoomConfigPath() {
	return path.join(app.getPath('userData'), ZOOM_CONFIG_FILE)
}

function loadZoomLevel() {
	try {
		const raw = fs.readFileSync(zoomConfigPath(), 'utf8')
		const parsed = JSON.parse(raw)
		if (typeof parsed.zoomLevel === 'number' && Number.isFinite(parsed.zoomLevel)) {
			return parsed.zoomLevel
		}
	} catch {
		// First run or unreadable file, fall back to default
	}
	return 0
}

function saveZoomLevel(level) {
	try {
		fs.writeFileSync(zoomConfigPath(), JSON.stringify({zoomLevel: level}))
	} catch (err) {
		console.warn('Failed to persist zoom level:', err.message)
	}
}

function applyZoom(webContents, level) {
	zoomLevel = level
	webContents.setZoomLevel(level)
	saveZoomLevel(level)
}

function wireZoomHandlers(win) {
	win.webContents.on('before-input-event', (event, input) => {
		if (input.type !== 'keyDown' || !input.control || input.alt || input.meta) return
		const key = input.key
		if (key === '=' || key === '+') {
			applyZoom(win.webContents, zoomLevel + ZOOM_STEP)
			event.preventDefault()
		} else if (key === '-') {
			applyZoom(win.webContents, zoomLevel - ZOOM_STEP)
			event.preventDefault()
		} else if (key === '0') {
			applyZoom(win.webContents, 0)
			event.preventDefault()
		}
	})

	win.webContents.on('zoom-changed', (_event, direction) => {
		const delta = direction === 'in' ? ZOOM_STEP : -ZOOM_STEP
		applyZoom(win.webContents, zoomLevel + delta)
	})
}

// ─── Main window ─────────────────────────────────────────────────────
function createMainWindow() {
	mainWindow = new BrowserWindow({
		width: 1680,
		height: 960,
		webPreferences: {
			...BASE_WEB_PREFERENCES,
			preload: path.join(__dirname, 'preload.js'),
		},
	})

	mainWindow.webContents.setWindowOpenHandler(({url}) => {
		safeOpenExternal(url)
		return {action: 'deny'}
	})

	// Prevent same-window navigation to external origins.
	// Only allow navigation to the local express server on the exact port.
	mainWindow.webContents.on('will-navigate', (event, navigationUrl) => {
		const parsedUrl = new URL(navigationUrl)
		if (parsedUrl.origin === `http://127.0.0.1:${serverPort}`) {
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

	wireZoomHandlers(mainWindow)
	mainWindow.webContents.on('did-finish-load', () => {
		mainWindow.webContents.setZoomLevel(zoomLevel)
	})

	// Process any deep link that arrived before the page was ready,
	// either buffered from open-url or passed via process.argv on first launch
	mainWindow.webContents.once('did-finish-load', () => {
		if (!pendingDeepLinkUrl) {
			pendingDeepLinkUrl = process.argv.find(arg => arg.startsWith(`${PROTOCOL}://`)) || null
		}
		if (pendingDeepLinkUrl) {
			handleDeepLink(pendingDeepLinkUrl)
			pendingDeepLinkUrl = null
		}
	})
}

// ─── Quick Entry window ──────────────────────────────────────────────
function getQuickEntryPosition() {
	const cursorPoint = screen.getCursorScreenPoint()
	const display = screen.getDisplayNearestPoint(cursorPoint)
	const {x: areaX, y: areaY, width: areaWidth, height: areaHeight} = display.workArea
	return {
		x: Math.round(areaX + (areaWidth - QUICK_ENTRY_WIDTH) / 2),
		y: Math.round(areaY + areaHeight / 3 - QUICK_ENTRY_COLLAPSED_HEIGHT / 2),
	}
}

function createQuickEntryWindow() {
	const {x, y} = getQuickEntryPosition()

	quickEntryWindow = new BrowserWindow({
		width: QUICK_ENTRY_WIDTH,
		height: QUICK_ENTRY_COLLAPSED_HEIGHT,
		x,
		y,
		frame: false,
		transparent: true,
		alwaysOnTop: true,
		skipTaskbar: true,
		resizable: false,
		show: false,
		webPreferences: {
			...BASE_WEB_PREFERENCES,
			preload: path.join(__dirname, 'preload-quick-entry.js'),
		},
	})

	quickEntryWindow.webContents.setWindowOpenHandler(({url}) => {
		safeOpenExternal(url)
		return {action: 'deny'}
	})

	quickEntryWindow.webContents.on('will-navigate', (event, navigationUrl) => {
		const parsedUrl = new URL(navigationUrl)
		if (parsedUrl.origin === `http://127.0.0.1:${serverPort}`) {
			return
		}
		event.preventDefault()
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
			quickEntryWindow.webContents.focus()
		})
		return
	}

	// Reset size and move to the active display
	quickEntryWindow.setSize(QUICK_ENTRY_WIDTH, QUICK_ENTRY_COLLAPSED_HEIGHT)
	const {x, y} = getQuickEntryPosition()
	quickEntryWindow.setPosition(x, y)

	// Reload to reset Vue state (clear previous input)
	quickEntryWindow.loadURL(`http://127.0.0.1:${serverPort}/?mode=quick-add`)
	// Wait for page to finish loading before showing, so the input gets focused
	quickEntryWindow.webContents.once('did-finish-load', () => {
		quickEntryWindow.show()
		quickEntryWindow.focus()
		quickEntryWindow.webContents.focus()
	})
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
	if (!tray) {
		const iconPath = path.join(__dirname, 'build', 'icon.png')
		const icon = nativeImage.createFromPath(iconPath).resize({width: 16, height: 16})
		tray = new Tray(icon)
		tray.setToolTip('Vikunja')
		tray.on('click', () => {
			if (mainWindow) {
				mainWindow.show()
				mainWindow.focus()
			} else {
				createMainWindow()
			}
		})
	}

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
			accelerator: currentShortcut || undefined,
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

	tray.setContextMenu(contextMenu)
}

// ─── IPC handlers ────────────────────────────────────────────────────
ipcMain.on('quick-entry:close', () => {
	hideQuickEntry()
})

ipcMain.on('quick-entry:resize', (_event, width, height) => {
	if (!quickEntryWindow) return
	if (!Number.isFinite(width) || !Number.isFinite(height)) return

	const display = screen.getDisplayNearestPoint(screen.getCursorScreenPoint())
	const maxWidth = display.workAreaSize.width
	const maxHeight = display.workAreaSize.height

	const w = Math.max(100, Math.min(Math.round(width), maxWidth))
	const h = Math.max(40, Math.min(Math.round(height), maxHeight))
	quickEntryWindow.setSize(w, h)
})

ipcMain.on('quick-entry:show-main-window', () => {
	if (mainWindow) {
		mainWindow.show()
		mainWindow.focus()
	} else {
		createMainWindow()
	}
})

// ─── Shortcut management ────────────────────────────────────────────
function registerQuickEntryShortcut(shortcut) {
	if (currentShortcut) {
		globalShortcut.unregister(currentShortcut)
	}

	if (!shortcut) {
		currentShortcut = null
		return
	}

	const registered = globalShortcut.register(shortcut, toggleQuickEntry)
	if (registered) {
		currentShortcut = shortcut
	} else {
		console.warn(`Failed to register global shortcut ${shortcut} — it may be in use by another application`)
		currentShortcut = null
	}
}

ipcMain.on('desktop:update-quick-entry-shortcut', (_event, shortcut) => {
	registerQuickEntryShortcut(shortcut)
	// Rebuild tray menu to reflect the new accelerator
	if (tray) {
		setupTray()
	}
})

// ─── App lifecycle ───────────────────────────────────────────────────
app.whenReady().then(() => {
	zoomLevel = loadZoomLevel()

	startServer(() => {
		createMainWindow()
		createQuickEntryWindow()
		setupTray()

		registerQuickEntryShortcut(DEFAULT_QUICK_ENTRY_SHORTCUT)

		// If launched with --quick-entry, show the quick entry window immediately
		if (launchedWithQuickEntry) {
			showQuickEntry()
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
