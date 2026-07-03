import { app, BrowserWindow, screen, ipcMain } from 'electron/main'
import { fileURLToPath } from 'node:url'
import * as path from 'node:path'
import { spawn, ChildProcess } from 'node:child_process'
import * as http from 'node:http'

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)

let mainWindow: BrowserWindow | null = null
let goProcess: ChildProcess | null = null
let pythonProcess: ChildProcess | null = null

function createWindow() {
  const primaryDisplay = screen.getPrimaryDisplay()
  const { width, height } = primaryDisplay.workAreaSize

  mainWindow = new BrowserWindow({
    width: Math.min(1400, Math.floor(width * 0.85)),
    height: Math.min(900, Math.floor(height * 0.85)),
    minWidth: 1100,
    minHeight: 700,
    frame: true,
    backgroundColor: '#1a1d23',
    titleBarStyle: 'hiddenInset',
    show: false,
    webPreferences: {
      preload: path.join(__dirname, 'preload.js'),
      contextIsolation: true,
      nodeIntegration: false,
    },
  })

  mainWindow.once('ready-to-show', () => {
    mainWindow?.show()
  })

  if (process.env.VITE_DEV_MODE === 'true') {
    mainWindow.loadURL('http://localhost:5173')
    mainWindow.webContents.openDevTools()
  } else {
    mainWindow.loadFile(path.join(__dirname, '../dist/index.html'))
  }

  mainWindow.on('closed', () => {
    mainWindow = null
  })
}

ipcMain.on('window-minimize', () => {
  mainWindow?.minimize()
})

ipcMain.on('window-toggle-maximize', () => {
  if (!mainWindow) return
  if (mainWindow.isMaximized()) {
    mainWindow.unmaximize()
  } else {
    mainWindow.maximize()
  }
})

ipcMain.on('window-close', () => {
  mainWindow?.close()
})

function startGoBackend() {
  const isDev = process.env.VITE_DEV_MODE === 'true'
  const isWin = process.platform === 'win32'

  let target: string
  let args: string[]

  if (isDev) {
    target = isWin ? 'go.exe' : 'go'
    args = ['run', './main.go']
  } else {
    const binaryName = isWin ? 'a-share-backend.exe' : 'a-share-backend'
    target = path.join(process.resourcesPath, 'backend', binaryName)
    args = []
  }

  goProcess = spawn(target, args, {
    cwd: isDev ? path.join(__dirname, '../../backend') : undefined,
    stdio: ['ignore', 'pipe', 'pipe'],
  })

  if (goProcess.stdout) {
    goProcess.stdout.on('data', (data) => {
      console.log('[Go]', data.toString())
    })
  }
  if (goProcess.stderr) {
    goProcess.stderr.on('data', (data) => {
      console.error('[Go]', data.toString())
    })
  }

  goProcess.on('exit', (code) => {
    console.log('[Go] exited with code', code)
  })
}

function waitForHealth(maxAttempts = 30, delay = 1000): Promise<void> {
  return new Promise((resolve, reject) => {
    let attempts = 0

    const check = () => {
      attempts++
      const req = http.get('http://localhost:8080/health', (res) => {
        let data = ''
        res.on('data', (chunk) => data += chunk)
        res.on('end', () => {
          if (res.statusCode === 200) {
            console.log('[Backend] healthy')
            resolve()
          } else if (attempts < maxAttempts) {
            setTimeout(check, delay)
          } else {
            reject(new Error('Backend health check failed'))
          }
        })
      })
      req.on('error', () => {
        if (attempts < maxAttempts) {
          setTimeout(check, delay)
        } else {
          reject(new Error('Backend health check failed'))
        }
      })
      req.end()
    }

    check()
  })
}

function stopGoBackend() {
  if (!goProcess) return

  try {
    if (process.platform === 'win32') {
      goProcess.kill('SIGKILL')
    } else {
      goProcess.kill('SIGTERM')
      setTimeout(() => {
        if (!goProcess?.killed) {
          goProcess?.kill('SIGKILL')
        }
      }, 3000)
    }
  } catch {
    // ignore
  }
  goProcess = null
}

function startPythonService() {
  const isDev = process.env.VITE_DEV_MODE === 'true'
  const isWin = process.platform === 'win32'

  let target: string
  let args: string[]

  if (isDev) {
    // In dev mode, assume Python service is already running on :8081
    console.log('[Python] Dev mode: expecting service on :8081')
    return
  }

  const binaryName = isWin ? 'app.exe' : 'app'
  target = path.join(process.resourcesPath, 'python_service', binaryName)
  args = ['--host', '0.0.0.0', '--port', '8081']

  pythonProcess = spawn(target, args, {
    stdio: ['ignore', 'pipe', 'pipe'],
  })

  if (pythonProcess.stdout) {
    pythonProcess.stdout.on('data', (data) => {
      console.log('[Python]', data.toString())
    })
  }
  if (pythonProcess.stderr) {
    pythonProcess.stderr.on('data', (data) => {
      console.error('[Python]', data.toString())
    })
  }

  pythonProcess.on('exit', (code) => {
    console.log('[Python] exited with code', code)
  })
}

function waitForPythonHealth(maxAttempts = 20, delay = 1000): Promise<void> {
  return new Promise((resolve, reject) => {
    let attempts = 0

    const check = () => {
      attempts++
      const req = http.get('http://localhost:8081/health', (res) => {
        res.on('data', () => {})
        res.on('end', () => {
          if (res.statusCode === 200) {
            console.log('[Python] healthy')
            resolve()
          } else if (attempts < maxAttempts) {
            setTimeout(check, delay)
          } else {
            reject(new Error('Python service health check failed'))
          }
        })
      })
      req.on('error', () => {
        if (attempts < maxAttempts) {
          setTimeout(check, delay)
        } else {
          reject(new Error('Python service health check failed'))
        }
      })
      req.end()
    }

    check()
  })
}

function stopPythonService() {
  if (!pythonProcess) return
  try {
    if (process.platform === 'win32') {
      pythonProcess.kill('SIGKILL')
    } else {
      pythonProcess.kill('SIGTERM')
      setTimeout(() => {
        if (!pythonProcess?.killed) {
          pythonProcess?.kill('SIGKILL')
        }
      }, 3000)
    }
  } catch {
    // ignore
  }
  pythonProcess = null
}

app.whenReady().then(async () => {
  const isDev = process.env.VITE_DEV_MODE === 'true'

  if (!isDev) {
    startGoBackend()
    startPythonService()
  }

  try {
    await Promise.all([waitForHealth(), waitForPythonHealth()])
  } catch (err) {
    console.error('[Main] Health check failed:', err)
    if (isDev) {
      console.log('[Main] Dev mode: creating window anyway')
    } else {
      throw err
    }
  }
  createWindow()

  app.on('activate', () => {
    if (BrowserWindow.getAllWindows().length === 0) {
      createWindow()
    }
  })
})

app.on('window-all-closed', () => {
  stopGoBackend()
  stopPythonService()
  if (process.platform !== 'darwin') {
    app.quit()
  }
})

app.on('before-quit', () => {
  stopGoBackend()
  stopPythonService()
})
