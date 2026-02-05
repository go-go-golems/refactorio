import { Topbar, type TopbarProps } from './Topbar'
import { Sidebar, type SidebarProps } from './Sidebar'

export interface AppShellProps {
  /** Main content area */
  children: React.ReactNode
  /** Sidebar props */
  sidebarProps?: Partial<SidebarProps>
  /** Topbar props */
  topbarProps?: Partial<TopbarProps>
  /** Whether sidebar is collapsed */
  sidebarCollapsed?: boolean
  /** Called when sidebar collapse toggles */
  onSidebarToggle?: () => void
  /** Hide sidebar entirely */
  hideSidebar?: boolean
  /** Hide topbar entirely */
  hideTopbar?: boolean
  /** Custom class name */
  className?: string
}

export function AppShell({
  children,
  sidebarProps,
  topbarProps,
  sidebarCollapsed = false,
  onSidebarToggle,
  hideSidebar = false,
  hideTopbar = false,
  className = '',
}: AppShellProps) {
  return (
    <div className={`app-shell d-flex flex-column vh-100 ${className}`}>
      {/* Topbar */}
      {!hideTopbar && (
        <Topbar {...topbarProps} />
      )}

      {/* Body: Sidebar + Main */}
      <div className="app-body d-flex flex-grow-1 overflow-hidden">
        {/* Sidebar */}
        {!hideSidebar && sidebarProps?.sections && (
          <Sidebar
            {...sidebarProps}
            sections={sidebarProps.sections}
            onNavigate={sidebarProps.onNavigate || (() => {})}
            collapsed={sidebarCollapsed}
          />
        )}

        {/* Main Content */}
        <main className="app-main flex-grow-1 overflow-auto">
          {children}
        </main>
      </div>
    </div>
  )
}
