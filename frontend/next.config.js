/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  async rewrites() {
    return [
      { source: '/api/alerts', destination: 'http://localhost:8081/api/alerts' },
      { source: '/api/stats', destination: 'http://localhost:8081/api/stats' },
    ]
  }
}
module.exports = nextConfig
