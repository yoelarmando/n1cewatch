# nextjs-developer Agent — Dashboard

**Skills:** `nextjs-developer`, `frontend-design`, `typescript-pro`
**Scope:** `frontend/**`, `lib/store` HTTP API

## Responsibilities
- Next.js 14 dashboard: `/` alerts table (filter by level), process tree, stats cards
- SWR polling `http://localhost:8081/api/alerts` (2s) — SSR not needed on server agent
- Proxy `next.config.js` rewrites to `:8081`

## Design Tokens
- BG `#0f172a`, card `#1e293b`, accent `#38bdf8`, critical `#ef4444`

## Verification
- `cd frontend && npm run build`
- Manual: trigger `curl|bash` anomaly, see alert in 2s
