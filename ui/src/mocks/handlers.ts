import { http, HttpResponse, delay } from 'msw'
import {
  mockWorkspaces,
  mockDBInfo,
  mockRuns,
  mockRunSummary,
  mockSessions,
  mockSymbols,
  mockSymbolRefs,
  mockCodeUnits,
  mockCommits,
  mockCommitFiles,
  mockDiffFiles,
  mockDiffHunks,
  mockDocTerms,
  mockDocHits,
  mockFileTree,
  mockSearchResults,
} from './data'

const API_BASE = '/api'

export const handlers = [
  // Health check
  http.get(`${API_BASE}/health`, async () => {
    await delay(100)
    return HttpResponse.json({ status: 'ok' })
  }),

  // Workspaces
  http.get(`${API_BASE}/workspaces`, async () => {
    await delay(150)
    return HttpResponse.json({ items: mockWorkspaces })
  }),

  http.get(`${API_BASE}/workspaces/:id`, async ({ params }) => {
    await delay(100)
    const workspace = mockWorkspaces.find((w) => w.id === params.id)
    if (!workspace) {
      return HttpResponse.json({ error: 'Workspace not found' }, { status: 404 })
    }
    return HttpResponse.json(workspace)
  }),

  http.post(`${API_BASE}/workspaces`, async ({ request }) => {
    await delay(200)
    const body = (await request.json()) as Record<string, unknown>
    const id = typeof body.id === 'string' && body.id.trim() ? body.id : `ws-${Date.now()}`
    return HttpResponse.json({
      id,
      name: body.name || id || 'New Workspace',
      db_path: body.db_path || '/tmp/new.db',
      repo_root: body.repo_root || '',
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    }, { status: 201 })
  }),

  http.patch(`${API_BASE}/workspaces/:id`, async ({ params, request }) => {
    await delay(200)
    const body = (await request.json()) as Record<string, unknown>
    const workspace = mockWorkspaces.find((w) => w.id === params.id)
    if (!workspace) {
      return HttpResponse.json({ error: 'Not found' }, { status: 404 })
    }
    return HttpResponse.json({ ...workspace, ...body, updated_at: new Date().toISOString() })
  }),

  // DB Info
  http.get(`${API_BASE}/db/info`, async () => {
    await delay(200)
    return HttpResponse.json(mockDBInfo)
  }),

  // Runs
  http.get(`${API_BASE}/runs`, async ({ request }) => {
    await delay(150)
    const url = new URL(request.url)
    const limit = parseInt(url.searchParams.get('limit') || '50')
    const offset = parseInt(url.searchParams.get('offset') || '0')
    const items = mockRuns.slice(offset, offset + limit)
    return HttpResponse.json({ items, total: mockRuns.length, limit, offset })
  }),

  http.get(`${API_BASE}/runs/:id`, async ({ params }) => {
    await delay(100)
    const run = mockRuns.find((r) => r.id === parseInt(params.id as string))
    if (!run) {
      return HttpResponse.json({ error: 'Run not found' }, { status: 404 })
    }
    return HttpResponse.json(run)
  }),

  http.get(`${API_BASE}/runs/:id/summary`, async () => {
    await delay(150)
    return HttpResponse.json(mockRunSummary)
  }),

  // Sessions
  http.get(`${API_BASE}/sessions`, async () => {
    await delay(150)
    return HttpResponse.json({ items: mockSessions })
  }),

  http.get(`${API_BASE}/sessions/:id`, async ({ params }) => {
    await delay(100)
    const session = mockSessions.find((s) => s.id === params.id)
    if (!session) {
      return HttpResponse.json({ error: 'Session not found' }, { status: 404 })
    }
    return HttpResponse.json(session)
  }),

  // Symbols
  http.get(`${API_BASE}/symbols`, async ({ request }) => {
    await delay(200)
    const url = new URL(request.url)
    const limit = parseInt(url.searchParams.get('limit') || '50')
    const offset = parseInt(url.searchParams.get('offset') || '0')
    const query = url.searchParams.get('name')
    const kind = url.searchParams.get('kind')

    let items = mockSymbols
    if (query) {
      items = items.filter(
        (s) =>
          s.name.toLowerCase().includes(query.toLowerCase()) ||
          s.pkg.toLowerCase().includes(query.toLowerCase())
      )
    }
    if (kind) {
      items = items.filter((s) => s.kind === kind)
    }
    const paginated = items.slice(offset, offset + limit)
    return HttpResponse.json({ items: paginated, total: items.length, limit, offset })
  }),

  http.get(`${API_BASE}/symbols/:hash`, async ({ params }) => {
    await delay(100)
    const symbol = mockSymbols.find((s) => s.symbol_hash === params.hash)
    if (!symbol) {
      return HttpResponse.json({ error: 'Symbol not found' }, { status: 404 })
    }
    return HttpResponse.json(symbol)
  }),

  http.get(`${API_BASE}/symbols/:hash/refs`, async () => {
    await delay(200)
    return HttpResponse.json({ items: mockSymbolRefs })
  }),

  // Code Units
  http.get(`${API_BASE}/code-units`, async ({ request }) => {
    await delay(200)
    const url = new URL(request.url)
    const limit = parseInt(url.searchParams.get('limit') || '50')
    const offset = parseInt(url.searchParams.get('offset') || '0')
    const kind = url.searchParams.get('kind')
    const query = url.searchParams.get('q')

    let items = mockCodeUnits
    if (kind) {
      items = items.filter((u) => u.kind === kind)
    }
    if (query) {
      items = items.filter((u) => u.name.toLowerCase().includes(query.toLowerCase()))
    }
    const paginated = items.slice(offset, offset + limit)
    return HttpResponse.json({ items: paginated, total: items.length, limit, offset })
  }),

  http.get(`${API_BASE}/code-units/:hash`, async ({ params }) => {
    await delay(100)
    const unit = mockCodeUnits.find((u) => u.unit_hash === params.hash)
    if (!unit) {
      return HttpResponse.json({ error: 'Code unit not found' }, { status: 404 })
    }
    return HttpResponse.json({
      ...unit,
      body_text: `func NewCommandProcessor(opts ...Option) CommandProcessor {
  impl := &commandProcessorImpl{
    middleware: make([]Middleware, 0),
    validators: make([]Validator, 0),
  }
  for _, opt := range opts {
    opt(impl)
  }
  return impl
}`,
      doc_text: '// NewCommandProcessor creates a new CommandProcessor with the given options.',
    })
  }),

  // Commits
  http.get(`${API_BASE}/commits`, async ({ request }) => {
    await delay(200)
    const url = new URL(request.url)
    const limit = parseInt(url.searchParams.get('limit') || '50')
    const offset = parseInt(url.searchParams.get('offset') || '0')
    const query = url.searchParams.get('q')

    let items = mockCommits
    if (query) {
      const q = query.toLowerCase()
      items = items.filter(
        (c) =>
          (c.subject ?? '').toLowerCase().includes(q) ||
          (c.author_name ?? '').toLowerCase().includes(q)
      )
    }
    const paginated = items.slice(offset, offset + limit)
    return HttpResponse.json({ items: paginated, total: items.length, limit, offset })
  }),

  http.get(`${API_BASE}/commits/:hash`, async ({ params }) => {
    await delay(100)
    const commit = mockCommits.find((c) => c.hash === params.hash)
    if (!commit) {
      return HttpResponse.json({ error: 'Commit not found' }, { status: 404 })
    }
    return HttpResponse.json(commit)
  }),

  http.get(`${API_BASE}/commits/:hash/files`, async () => {
    await delay(150)
    return HttpResponse.json({ items: mockCommitFiles })
  }),

  // Diff
  http.get(`${API_BASE}/diff-runs`, async () => {
    await delay(150)
    return HttpResponse.json({
      items: [
        {
          id: 43,
          root_path: '/Users/dev/src/glazed',
          git_from: 'HEAD~20',
          git_to: 'HEAD',
        },
      ],
    })
  }),

  http.get(`${API_BASE}/diff/:runId/files`, async () => {
    await delay(150)
    return HttpResponse.json({ items: mockDiffFiles })
  }),

  http.get(`${API_BASE}/diff/:runId/file`, async () => {
    await delay(100)
    return HttpResponse.json({ hunks: mockDiffHunks })
  }),

  // Docs
  http.get(`${API_BASE}/docs/terms`, async () => {
    await delay(150)
    return HttpResponse.json({ items: mockDocTerms })
  }),

  http.get(`${API_BASE}/docs/hits`, async ({ request }) => {
    await delay(150)
    const url = new URL(request.url)
    const term = url.searchParams.get('term')
    let items = mockDocHits
    if (term) {
      items = items.filter((h) => h.term.toLowerCase().includes(term.toLowerCase()))
    }
    return HttpResponse.json({ items })
  }),

  // Files
  http.get(`${API_BASE}/files`, async () => {
    await delay(150)
    return HttpResponse.json({ items: mockFileTree })
  }),

  http.get(`${API_BASE}/file`, async ({ request }) => {
    await delay(200)
    const url = new URL(request.url)
    const path = url.searchParams.get('path')
    return HttpResponse.json({
      path,
      content: `package handlers

import (
	"context"
)

// CommandProcessor handles command execution
type CommandProcessor interface {
	Process(ctx context.Context, cmd Command) (Result, error)
	Validate(cmd Command) error
}

// NewCommandProcessor creates a new CommandProcessor with the given options.
func NewCommandProcessor(opts ...Option) CommandProcessor {
	impl := &commandProcessorImpl{
		middleware: make([]Middleware, 0),
		validators: make([]Validator, 0),
	}
	for _, opt := range opts {
		opt(impl)
	}
	return impl
}
`,
    })
  }),

  // Search
  http.post(`${API_BASE}/search`, async ({ request }) => {
    await delay(300)
    const body = (await request.json()) as { query?: string; types?: string[] }
    let results = mockSearchResults

    if (body.query) {
      const q = body.query.toLowerCase()
      results = results.filter(
        (r) => r.primary.toLowerCase().includes(q) || r.snippet?.toLowerCase().includes(q)
      )
    }

    if (body.types && body.types.length > 0) {
      results = results.filter((r) => body.types!.includes(r.type))
    }

    return HttpResponse.json({ items: results })
  }),
]
