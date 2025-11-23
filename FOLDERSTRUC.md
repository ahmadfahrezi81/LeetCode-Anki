leetcode-anki/
├── backend/                    # Go API
│   ├── cmd/
│   │   └── api/
│   │       └── main.go         # Entry point
│   ├── internal/
│   │   ├── handlers/           # HTTP handlers (Gin routes)
│   │   │   ├── health.go
│   │   │   ├── review.go       # Submit answer, get next card
│   │   │   └── dashboard.go        # User stats/dashboard
│   │   ├── services/
│   │   │   ├── srs.go          # SM-2 algorithm logic
│   │   │   ├── llm.go          # OpenAI scoring
│   │   │   └── queue.go        # Card queue logic (new vs review)
│   │   ├── models/
│   │   │   ├── question.go
│   │   │   ├── review.go
│   │   │   └── user.go
│   │   ├── database/
│   │   │   ├── db.go           # Supabase connection
│   │   │   └── queries.go      # SQL queries
│   │   └── middleware/
│   │       └── auth.go         # Supabase JWT validation
│   ├── config/
│   │   └── config.go           # Environment variables
│   ├── go.mod
│   ├── go.sum
│   └── Dockerfile
│
├── worker/                     # Python Crawler
│   ├── main.py                 # Entry point (cron job)
│   ├── scraper.py              # Crawl4AI logic
│   ├── requirements.txt
│   └── Dockerfile
│
├── frontend/                   # Next.js 14 App Router
│   ├── app/
│   │   ├── layout.tsx
│   │   ├── page.tsx            # Dashboard (stats + study button)
│   │   ├── study/
│   │   │   └── page.tsx        # Study session (card display)
│   │   └── login/
│   │       └── page.tsx        # Supabase Auth
│   ├── components/
│   │   ├── Card.tsx            # Problem card UI
│   │   ├── AnswerInput.tsx     # Text area for explanation
│   │   └── StatsPanel.tsx      # Review stats
│   ├── lib/
│   │   ├── supabase.ts         # Supabase client
│   │   └── api.ts              # API calls to Go backend
│   ├── tailwind.config.ts
│   ├── package.json
│   └── Dockerfile
│
├── docker-compose.yml          # Orchestrates all services
└── README.md