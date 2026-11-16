// Keep API requests and refresh cookies on the frontend origin.
const apiUrl = (process.env.VITE_API_URL || '').trim().replace(/\/+$/, '')
if (!apiUrl) {
  throw new Error('Configure VITE_API_URL na Vercel com a raiz HTTPS pública da API Docker.')
}
const api = new URL(apiUrl)
if (api.protocol !== 'https:' || api.username || api.password || api.search || api.hash || api.pathname !== '/' || ['localhost', '127.0.0.1', '[::1]'].includes(api.hostname)) {
  throw new Error('VITE_API_URL deve ser uma origem pública HTTPS, sem /api/v2, caminhos, credenciais ou parâmetros.')
}

export const config = {
  framework: null,
  installCommand: 'npm install --global pnpm@12.16.1 && cd frontend && PUPPETEER_SKIP_DOWNLOAD=true pnpm install --frozen-lockfile',
  buildCommand: 'cd frontend && pnpm run build',
  outputDirectory: 'frontend/dist',
  rewrites: [
    {source: '/api/:path*', destination: `${api.origin}/api/:path*`},
    {source: '/(.*)', destination: '/index.html'},
  ],
  headers: [
    {source: '/api/:path*', headers: [{key: 'Cache-Control', value: 'private, no-store'}]},
  ],
}
