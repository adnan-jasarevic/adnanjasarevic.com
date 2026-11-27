# adnanjasarevic.com

Source for my personal portfolio site, adnanjasarevic.com. The site hasn't been deployed yet.

- **backend/**: Go API using only the standard library `net/http`, PostgreSQL via pgx, migrations with goose. Serves CV and project data, handles the contact form and an admin area with GitHub sign-in.
- **frontend/**: React, TypeScript and Vite. Loads its content from the API at runtime.

The API runs on Render with Postgres on Neon and Resend for contact form emails. The frontend will be hosted on Cloudflare Pages.

## Running locally

Requires Go, Node and Docker.

```sh
docker compose up -d                  # Postgres on localhost:5433
cd backend && go run ./cmd/api        # API on localhost:8080, reads backend/.env
cd frontend && npm install && npm run dev   # site on localhost:5173
```

Run the last two in separate terminals from the repo root.

The backend needs at least `DATABASE_URL`. Email notifications and admin login are turned off when their environment variables (`RESEND_API_KEY`, `GITHUB_CLIENT_ID`, ...) are missing; see `backend/internal/config`.
