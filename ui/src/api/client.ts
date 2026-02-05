// RTK Query API — barrel re-export
// Domain-specific endpoints are defined in separate slice files using injectEndpoints

export { api } from './baseApi'

// Workspaces
export {
  useGetWorkspacesQuery,
  useGetWorkspaceQuery,
  useGetDBInfoQuery,
  useCreateWorkspaceMutation,
  useUpdateWorkspaceMutation,
  useDeleteWorkspaceMutation,
} from './workspaces'

// Runs
export {
  useGetRunsQuery,
  useGetRunQuery,
  useGetRunSummaryQuery,
} from './runs'

// Sessions
export {
  useGetSessionsQuery,
  useGetSessionQuery,
} from './sessions'

// Symbols
export {
  useGetSymbolsQuery,
  useGetSymbolQuery,
  useGetSymbolRefsQuery,
} from './symbols'

// Code Units
export {
  useGetCodeUnitsQuery,
  useGetCodeUnitQuery,
  useGetCodeUnitHistoryQuery,
} from './codeUnits'

// Commits
export {
  useGetCommitsQuery,
  useGetCommitQuery,
  useGetCommitFilesQuery,
} from './commits'

// Diffs
export {
  useGetDiffRunsQuery,
  useGetDiffFilesQuery,
  useGetDiffFileQuery,
} from './diffs'

// Docs
export {
  useGetDocTermsQuery,
  useGetDocHitsQuery,
} from './docs'

// Files
export {
  useGetFilesQuery,
  useGetFileContentQuery,
  useGetFileHistoryQuery,
} from './files'

// Search
export {
  useSearchUnifiedQuery,
  useSearchSymbolsQuery,
  useSearchCodeUnitsQuery,
  useSearchCommitsQuery,
} from './search'
