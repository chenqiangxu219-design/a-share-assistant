export async function resolve(specifier, parentURL, next) {
  if (specifier === 'electron') {
    // For CJS require('electron'), we want to resolve to electron/main
    // But we need to return a file URL that the ESM loader can handle
    return next(specifier);
  }
  return next(specifier);
}
