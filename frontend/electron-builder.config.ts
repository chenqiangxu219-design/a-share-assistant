import { Configuration } from 'electron-builder'

const config: Configuration = {
  appId: 'com.astock.assistant',
  productName: 'A股智能助手',
  directories: {
    output: 'release',
  },
  files: [
    'dist/**/*',
    'electron/dist/main.js',
    'electron/dist/preload.js',
    'electron/dist/package.json',
  ],
  mac: {
    category: 'public.app-category.finance',
    target: [
      {
        target: 'dmg',
        arch: ['arm64', 'x64'],
      },
    ],
    artifactName: '${productName}-${version}-${arch}.${ext}',
    extraResources: [
      { from: '../builds/a-share-backend', to: 'backend/a-share-backend' },
      { from: '../builds/a-share-python-service', to: 'python_service/app' },
    ],
  },
  win: {
    target: [
      {
        target: 'nsis',
        arch: ['x64'],
      },
    ],
    artifactName: '${productName}-${version}.${ext}',
    extraResources: [
      { from: '../builds/a-share-backend.exe', to: 'backend/a-share-backend.exe' },
    ],
  },
  linux: {
    target: ['AppImage'],
    category: 'Finance',
    extraResources: [
      { from: '../builds/a-share-backend-linux', to: 'backend/a-share-backend' },
      // Python service is macOS-only; not included in Linux builds
    ],
  },
  npmRebuild: false,
}

export default config
