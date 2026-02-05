---
Title: UI Widget Catalog and Storybook Plan
Ticket: REF-007-INDEX-BROWSE-UI
Status: active
Topics:
    - ui
    - widgets
    - storybook
    - components
DocType: analysis
Intent: long-term
Owners: []
RelatedFiles:
    - Path: refactorio/ui/src/components
      Note: Component implementation directory
    - Path: refactorio/ui/.storybook
      Note: Storybook configuration
Summary: Comprehensive catalog of UI widgets needed for the Refactorio Workbench with Storybook story specifications.
LastUpdated: 2026-02-05T10:00:00-05:00
WhatFor: Guide incremental widget development with Storybook-driven workflow.
WhenToUse: Reference when implementing new components or adding stories.
---

# UI Widget Catalog and Storybook Plan

This document catalogs all UI widgets needed for the Refactorio Workbench MVP 1 (Investigation Workbench), organized by category with Storybook story specifications.

## Implementation Strategy

1. **Storybook-first development**: Each widget is developed in isolation with stories before integration
2. **Three variant levels**: Default (styled), Themed (customizable), Unstyled (headless)
3. **Mock data**: Stories use realistic mock data matching API response shapes
4. **Accessibility**: Each component includes a11y testing via Storybook addon

---

## 1. Foundation Components

### 1.1 CopyButton

**Purpose**: Copy text to clipboard with visual feedback

**Props**:
```typescript
interface CopyButtonProps {
  text: string           // Text to copy
  label?: string         // Button label (default: icon only)
  size?: 'sm' | 'md'     // Size variant
  variant?: 'icon' | 'text' | 'outline'
  onCopy?: () => void    // Callback after copy
}
```

**States**: idle, copying, copied (with timeout reset)

**Storybook Stories**:
- `Default` - Icon-only copy button
- `WithLabel` - "Copy" text label
- `AfterCopy` - Shows checkmark feedback
- `Sizes` - sm/md variants
- `Variants` - icon/text/outline styles

---

### 1.2 OpenInEditorButton

**Purpose**: Open file at location in external editor

**Props**:
```typescript
interface OpenInEditorButtonProps {
  filePath: string
  line?: number
  column?: number
  editor?: 'vscode' | 'idea' | 'custom'
  customScheme?: string
  repoRoot?: string
}
```

**States**: idle, opening, error (if scheme fails)

**Storybook Stories**:
- `Default` - VS Code link
- `WithLineNumber` - file:line format
- `JetBrains` - IntelliJ IDEA scheme
- `CustomScheme` - User-defined URL template
- `NoRepoRoot` - Disabled state with tooltip

---

### 1.3 StatusBadge

**Purpose**: Display status with color coding

**Props**:
```typescript
interface StatusBadgeProps {
  status: 'success' | 'running' | 'failed' | 'pending' | 'warning'
  label?: string
  size?: 'sm' | 'md'
  pulse?: boolean  // Animate for 'running'
}
```

**Storybook Stories**:
- `AllStatuses` - Grid of all status variants
- `WithLabels` - Custom label text
- `Running` - Pulse animation
- `Sizes` - Size comparison

---

### 1.4 EntityIcon

**Purpose**: Icon representing entity types

**Props**:
```typescript
interface EntityIconProps {
  type: 'symbol' | 'code_unit' | 'commit' | 'diff' | 'doc' | 'file' | 'folder' | 'run' | 'session'
  kind?: string  // Sub-type (e.g., 'func', 'type', 'method' for symbols)
  size?: 'sm' | 'md' | 'lg'
}
```

**Storybook Stories**:
- `AllTypes` - Grid of all entity icons
- `SymbolKinds` - func/type/method/const/var icons
- `Sizes` - Size comparison
- `WithLabel` - Icon + text combinations

---

### 1.5 Pagination

**Purpose**: Server-side pagination controls

**Props**:
```typescript
interface PaginationProps {
  total?: number
  limit: number
  offset: number
  onChange: (offset: number) => void
  showTotal?: boolean
  compact?: boolean
}
```

**Storybook Stories**:
- `Default` - Standard pagination
- `WithTotal` - Shows "1-50 of 1,234"
- `Compact` - Minimal prev/next only
- `FirstPage` - Prev disabled
- `LastPage` - Next disabled
- `Loading` - Skeleton state

---

## 2. Layout Components

### 2.1 AppShell

**Purpose**: Main application layout with topbar, sidebar, and content area

**Props**:
```typescript
interface AppShellProps {
  children: React.ReactNode
  sidebar?: React.ReactNode
  topbar?: React.ReactNode
  sidebarCollapsed?: boolean
  onSidebarToggle?: () => void
}
```

**Storybook Stories**:
- `Default` - Full app shell with placeholder content
- `CollapsedSidebar` - Narrow sidebar state
- `NoSidebar` - Content-only mode
- `Loading` - Skeleton placeholders

---

### 2.2 ThreePaneLayout

**Purpose**: List → Preview → Inspector pattern

**Props**:
```typescript
interface ThreePaneLayoutProps {
  left: React.ReactNode
  center: React.ReactNode
  right?: React.ReactNode
  leftWidth?: number
  rightWidth?: number
  showRight?: boolean
  onToggleRight?: () => void
}
```

**Storybook Stories**:
- `Default` - All three panes visible
- `TwoPanes` - Left + center only
- `CustomWidths` - Adjusted pane widths
- `CollapsibleRight` - Toggle inspector

---

### 2.3 Sidebar

**Purpose**: Navigation sidebar with collapsible sections

**Props**:
```typescript
interface SidebarProps {
  items: SidebarSection[]
  activeItem?: string
  onNavigate: (path: string) => void
  collapsed?: boolean
}

interface SidebarSection {
  label: string
  icon?: React.ReactNode
  items: SidebarItem[]
  collapsible?: boolean
  defaultOpen?: boolean
}

interface SidebarItem {
  id: string
  label: string
  icon?: React.ReactNode
  path: string
  badge?: string | number
}
```

**Storybook Stories**:
- `Default` - Full navigation tree
- `Collapsed` - Icons only
- `WithBadges` - Count badges on items
- `ActiveStates` - Highlighted active item
- `Sections` - Collapsible section groups

---

### 2.4 Topbar

**Purpose**: Top navigation bar with workspace/session selectors and search

**Props**:
```typescript
interface TopbarProps {
  workspaceName?: string
  sessionName?: string
  onWorkspaceClick?: () => void
  onSessionClick?: () => void
  onSearch?: (query: string) => void
  onCommandPalette?: () => void
}
```

**Storybook Stories**:
- `Default` - All elements present
- `NoSession` - Session selector disabled
- `Loading` - Skeleton state
- `SearchFocused` - Expanded search input

---

## 3. Selection Components

### 3.1 WorkspaceSelector

**Purpose**: Modal/dropdown to select workspace

**Props**:
```typescript
interface WorkspaceSelectorProps {
  workspaces: Workspace[]
  selected?: Workspace
  onSelect: (workspace: Workspace) => void
  onAdd?: () => void
  onEdit?: (workspace: Workspace) => void
  loading?: boolean
}
```

**Storybook Stories**:
- `Default` - List of workspaces
- `Empty` - No workspaces, add CTA
- `Loading` - Skeleton list
- `WithSelected` - Highlighted selection
- `Modal` - Full modal variant

---

### 3.2 SessionSelector

**Purpose**: Select index session (grouped runs)

**Props**:
```typescript
interface SessionSelectorProps {
  sessions: Session[]
  selected?: Session
  onSelect: (session: Session) => void
  loading?: boolean
}
```

**Storybook Stories**:
- `Default` - Session dropdown
- `WithAvailability` - Shows data availability badges
- `SingleSession` - Only one option
- `Loading` - Skeleton state

---

### 3.3 SessionCard

**Purpose**: Card showing session details and availability

**Props**:
```typescript
interface SessionCardProps {
  session: Session
  selected?: boolean
  onClick?: () => void
  onEdit?: () => void
}
```

**Storybook Stories**:
- `Default` - Full session card
- `Selected` - Highlighted state
- `AllAvailable` - All data types present
- `PartialAvailable` - Some missing/failed
- `Compact` - Minimal card variant

---

## 4. Data Display Components

### 4.1 EntityTable

**Purpose**: Server-paginated table for any entity type

**Props**:
```typescript
interface EntityTableProps<T> {
  columns: Column<T>[]
  data: T[]
  loading?: boolean
  pagination?: PaginationProps
  selectedId?: string
  onSelect?: (item: T) => void
  onSort?: (column: string, direction: 'asc' | 'desc') => void
  sortColumn?: string
  sortDirection?: 'asc' | 'desc'
  emptyMessage?: string
}

interface Column<T> {
  key: string
  header: string
  width?: string
  render?: (item: T) => React.ReactNode
  sortable?: boolean
}
```

**Storybook Stories**:
- `Symbols` - Symbol list with kind/package/file columns
- `Commits` - Commit list with hash/subject/author
- `Runs` - Run list with status/dates
- `Loading` - Skeleton rows
- `Empty` - Empty state message
- `Sortable` - Click to sort columns
- `Selected` - Row selection highlight

---

### 4.2 SearchResults

**Purpose**: Grouped search results display

**Props**:
```typescript
interface SearchResultsProps {
  results: SearchResult[]
  groupByType?: boolean
  selectedId?: string
  onSelect?: (result: SearchResult) => void
  loading?: boolean
}
```

**Storybook Stories**:
- `Default` - Mixed results
- `GroupedByType` - Collapsible type sections
- `SingleType` - Filtered to one type
- `WithSnippets` - Highlighted match snippets
- `Loading` - Skeleton results
- `NoResults` - Empty state

---

### 4.3 AvailabilityGrid

**Purpose**: Show data availability for a session

**Props**:
```typescript
interface AvailabilityGridProps {
  availability: SessionAvailability
  runs?: SessionRuns
  onCompute?: (type: string) => void
  onViewError?: (runId: number) => void
}
```

**Storybook Stories**:
- `AllAvailable` - All green checkmarks
- `PartialAvailable` - Mix of states
- `WithCounts` - Row counts displayed
- `WithActions` - Compute/retry buttons

---

## 5. Code Display Components

### 5.1 CodeViewer

**Purpose**: Display code with line numbers and highlights

**Props**:
```typescript
interface CodeViewerProps {
  content: string
  language?: string
  startLine?: number
  highlightLines?: number[]
  highlightRanges?: { start: number; end: number; className?: string }[]
  onLineClick?: (line: number) => void
  showLineNumbers?: boolean
  wrapLines?: boolean
}
```

**Storybook Stories**:
- `Default` - Plain code display
- `WithHighlights` - Highlighted lines
- `GoCode` - Go syntax example
- `TypeScript` - TypeScript syntax
- `LongFile` - Virtualized scrolling
- `Selection` - Line click handling
- `NoLineNumbers` - Clean view

---

### 5.2 DiffViewer

**Purpose**: Display diff hunks with add/remove styling

**Props**:
```typescript
interface DiffViewerProps {
  hunks: DiffHunk[]
  mode?: 'unified' | 'split'
  showLineNumbers?: boolean
  contextLines?: number
  onLineClick?: (line: DiffLine) => void
  highlightQuery?: string
}
```

**Storybook Stories**:
- `Unified` - Standard unified diff
- `Split` - Side-by-side view
- `AddOnly` - Only additions
- `RemoveOnly` - Only deletions
- `Mixed` - Complex changes
- `WithHighlight` - Search term highlighted
- `LargeHunk` - Many lines

---

### 5.3 SnippetPreview

**Purpose**: Small code snippet with context

**Props**:
```typescript
interface SnippetPreviewProps {
  content: string
  language?: string
  highlightRange?: { startLine: number; endLine: number; startCol?: number; endCol?: number }
  maxLines?: number
  showExpand?: boolean
  onExpand?: () => void
}
```

**Storybook Stories**:
- `Default` - Simple snippet
- `WithHighlight` - Highlighted match
- `Truncated` - "Show more" link
- `FullContext` - Expanded view

---

## 6. Detail Components

### 6.1 SymbolDetail

**Purpose**: Full symbol information panel

**Props**:
```typescript
interface SymbolDetailProps {
  symbol: Symbol
  refs?: SymbolRef[]
  refsLoading?: boolean
  refsAvailable?: boolean
  onComputeRefs?: () => void
  onAddToPlan?: () => void
  onOpenInEditor?: () => void
}
```

**Storybook Stories**:
- `Default` - Full symbol details
- `WithRefs` - References list
- `NoRefs` - "Compute refs" CTA
- `Loading` - Skeleton state
- `Exported` - Exported symbol styling
- `Private` - Unexported symbol

---

### 6.2 CodeUnitDetail

**Purpose**: Code unit (function/type) details

**Props**:
```typescript
interface CodeUnitDetailProps {
  codeUnit: CodeUnitDetail
  history?: CodeUnit[]
  onDiff?: (hash1: string, hash2: string) => void
  onAddToPlan?: () => void
}
```

**Storybook Stories**:
- `Function` - Function code unit
- `Type` - Type definition
- `Method` - Method with receiver
- `WithHistory` - Version timeline
- `WithDoc` - Doc comment displayed

---

### 6.3 CommitDetail

**Purpose**: Commit information and files

**Props**:
```typescript
interface CommitDetailProps {
  commit: Commit
  files?: CommitFile[]
  onFileClick?: (file: CommitFile) => void
  onViewDiff?: () => void
}
```

**Storybook Stories**:
- `Default` - Commit metadata
- `WithFiles` - Changed files list
- `LongMessage` - Multi-paragraph body
- `MergeCommit` - Multiple parents

---

### 6.4 InspectorPanel

**Purpose**: Right-side contextual panel

**Props**:
```typescript
interface InspectorPanelProps {
  title: string
  subtitle?: string
  actions?: React.ReactNode
  children: React.ReactNode
  onClose?: () => void
}
```

**Storybook Stories**:
- `Symbol` - Symbol inspector
- `CodeUnit` - Code unit inspector
- `File` - File inspector
- `Empty` - No selection
- `Loading` - Skeleton content

---

## 7. Navigation Components

### 7.1 FileTree

**Purpose**: Hierarchical file browser

**Props**:
```typescript
interface FileTreeProps {
  entries: FileEntry[]
  selectedPath?: string
  expandedPaths?: string[]
  onSelect?: (entry: FileEntry) => void
  onExpand?: (path: string) => void
  onCollapse?: (path: string) => void
  loading?: boolean
  badges?: Record<string, string | number>
}
```

**Storybook Stories**:
- `Default` - Basic tree
- `WithBadges` - File badges (diff count, etc.)
- `DeepNesting` - Many levels
- `Selected` - Highlighted file
- `Loading` - Lazy load children
- `Empty` - No files

---

### 7.2 Breadcrumb

**Purpose**: Navigation path display

**Props**:
```typescript
interface BreadcrumbProps {
  items: { label: string; path?: string }[]
  onNavigate?: (path: string) => void
}
```

**Storybook Stories**:
- `Default` - Multi-level path
- `Clickable` - Interactive segments
- `Truncated` - Long paths with ellipsis

---

### 7.3 TabNav

**Purpose**: Tab navigation within views

**Props**:
```typescript
interface TabNavProps {
  tabs: { id: string; label: string; badge?: string | number; disabled?: boolean }[]
  activeTab: string
  onChange: (tabId: string) => void
}
```

**Storybook Stories**:
- `Default` - Multiple tabs
- `WithBadges` - Count badges
- `Disabled` - Some tabs disabled
- `Overflow` - Scrollable tabs

---

## 8. Search Components

### 8.1 GlobalSearchBar

**Purpose**: Main search input with suggestions

**Props**:
```typescript
interface GlobalSearchBarProps {
  value?: string
  onChange?: (value: string) => void
  onSubmit?: (value: string) => void
  placeholder?: string
  suggestions?: string[]
  loading?: boolean
  autoFocus?: boolean
}
```

**Storybook Stories**:
- `Default` - Empty search bar
- `WithValue` - Pre-filled query
- `WithSuggestions` - Autocomplete dropdown
- `Loading` - Search in progress
- `Focused` - Expanded state

---

### 8.2 FilterPanel

**Purpose**: Search filters sidebar

**Props**:
```typescript
interface FilterPanelProps {
  filters: FilterConfig[]
  values: Record<string, unknown>
  onChange: (key: string, value: unknown) => void
  onReset?: () => void
}

interface FilterConfig {
  key: string
  label: string
  type: 'checkbox-group' | 'select' | 'text' | 'date-range'
  options?: { value: string; label: string }[]
}
```

**Storybook Stories**:
- `TypeFilters` - Entity type checkboxes
- `KindFilters` - Symbol kind selection
- `PathFilter` - File path input
- `DateRange` - Date range picker
- `Combined` - All filter types
- `Applied` - Active filters shown

---

### 8.3 CommandPalette

**Purpose**: Keyboard-driven command/search modal

**Props**:
```typescript
interface CommandPaletteProps {
  open: boolean
  onClose: () => void
  commands: Command[]
  recentSearches?: string[]
  onCommand: (command: Command) => void
  onSearch: (query: string) => void
}

interface Command {
  id: string
  label: string
  shortcut?: string
  icon?: React.ReactNode
  category?: string
}
```

**Storybook Stories**:
- `Default` - Open palette
- `WithCommands` - Command list
- `WithRecent` - Recent searches
- `Filtered` - Typed query filtering
- `Categories` - Grouped commands

---

## 9. Form Components

### 9.1 WorkspaceForm

**Purpose**: Add/edit workspace

**Props**:
```typescript
interface WorkspaceFormProps {
  workspace?: Workspace
  onSubmit: (data: WorkspaceFormData) => void
  onCancel?: () => void
  loading?: boolean
  error?: string
}

interface WorkspaceFormData {
  name: string
  db_path: string
  repo_root?: string
}
```

**Storybook Stories**:
- `Add` - Empty form
- `Edit` - Pre-filled form
- `Validation` - Error states
- `Loading` - Submit in progress

---

## Storybook Configuration

### Directory Structure

```
refactorio/ui/
├── .storybook/
│   ├── main.ts
│   ├── preview.ts
│   └── preview-head.html
├── src/
│   ├── components/
│   │   ├── foundation/
│   │   │   ├── CopyButton.tsx
│   │   │   ├── CopyButton.stories.tsx
│   │   │   └── ...
│   │   ├── layout/
│   │   ├── selection/
│   │   ├── data-display/
│   │   ├── code-display/
│   │   ├── detail/
│   │   ├── navigation/
│   │   ├── search/
│   │   └── form/
│   └── ...
```

### Story Template

```typescript
// Example: CopyButton.stories.tsx
import type { Meta, StoryObj } from '@storybook/react'
import { CopyButton } from './CopyButton'

const meta: Meta<typeof CopyButton> = {
  title: 'Foundation/CopyButton',
  component: CopyButton,
  tags: ['autodocs'],
  argTypes: {
    size: { control: 'radio', options: ['sm', 'md'] },
    variant: { control: 'radio', options: ['icon', 'text', 'outline'] },
  },
}

export default meta
type Story = StoryObj<typeof CopyButton>

export const Default: Story = {
  args: {
    text: 'pkg/handlers/command.go:45',
  },
}

export const WithLabel: Story = {
  args: {
    text: 'pkg/handlers/command.go:45',
    label: 'Copy path',
  },
}

export const Sizes: Story = {
  render: () => (
    <div style={{ display: 'flex', gap: '1rem', alignItems: 'center' }}>
      <CopyButton text="small" size="sm" />
      <CopyButton text="medium" size="md" />
    </div>
  ),
}
```

### Mock Data

Create `src/mocks/` directory with realistic data:

```typescript
// src/mocks/symbols.ts
export const mockSymbols: Symbol[] = [
  {
    symbol_hash: 'a7b3c9f2',
    name: 'CommandProcessor',
    kind: 'type',
    package_path: 'github.com/example/pkg/handlers',
    signature: 'type CommandProcessor interface { ... }',
    exported: true,
    file_path: 'pkg/handlers/command.go',
    start_line: 45,
    start_col: 1,
    end_line: 52,
    end_col: 2,
    run_id: 44,
  },
  // ... more mock data
]
```

---

## Implementation Order

Based on dependencies and building blocks:

### Phase 1: Foundation (Week 1)
1. CopyButton
2. StatusBadge
3. EntityIcon
4. Pagination

### Phase 2: Layout (Week 1)
5. AppShell
6. Sidebar
7. Topbar
8. ThreePaneLayout

### Phase 3: Data Display (Week 2)
9. EntityTable
10. SearchResults
11. AvailabilityGrid

### Phase 4: Code Display (Week 2)
12. CodeViewer
13. DiffViewer
14. SnippetPreview

### Phase 5: Navigation (Week 3)
15. FileTree
16. Breadcrumb
17. TabNav
18. GlobalSearchBar
19. FilterPanel

### Phase 6: Selection (Week 3)
20. WorkspaceSelector
21. SessionSelector
22. SessionCard

### Phase 7: Detail Panels (Week 4)
23. InspectorPanel
24. SymbolDetail
25. CodeUnitDetail
26. CommitDetail

### Phase 8: Advanced (Week 4)
27. CommandPalette
28. WorkspaceForm
29. OpenInEditorButton

---

## Testing Strategy

Each component should have:

1. **Visual regression tests** via Storybook chromatic (optional)
2. **Interaction tests** via Storybook play functions
3. **Accessibility tests** via @storybook/addon-a11y

Example interaction test:

```typescript
export const CopyInteraction: Story = {
  args: { text: 'test' },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)
    const button = canvas.getByRole('button')
    await userEvent.click(button)
    // Assert clipboard (mocked) was called
  },
}
```
