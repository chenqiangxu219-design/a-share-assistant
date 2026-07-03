const BACKEND_URL = 'http://localhost:8080'

export function apiPath(path: string): string {
  return BACKEND_URL + path
}
