export interface ElectronAPI {
  isElectron: boolean
  platform: string
  minimize: () => void
  toggleMaximize: () => void
  close: () => void
}

export {}
declare global {
  interface Window {
    electronAPI: ElectronAPI
  }
}
