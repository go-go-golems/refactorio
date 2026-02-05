import { useState } from 'react'
import { EntityIcon } from '../foundation'

export interface SidebarItem {
  id: string
  label: string
  icon?: React.ReactNode
  path: string
  badge?: string | number
}

export interface SidebarSection {
  id: string
  label: string
  icon?: React.ReactNode
  items: SidebarItem[]
  collapsible?: boolean
  defaultOpen?: boolean
}

export interface SidebarProps {
  /** Sidebar sections with items */
  sections: SidebarSection[]
  /** Currently active item ID */
  activeItem?: string
  /** Called when an item is clicked */
  onNavigate: (path: string) => void
  /** Collapsed mode (icons only) */
  collapsed?: boolean
  /** Custom class name */
  className?: string
}

export function Sidebar({
  sections,
  activeItem,
  onNavigate,
  collapsed = false,
  className = '',
}: SidebarProps) {
  const [openSections, setOpenSections] = useState<Set<string>>(() => {
    const open = new Set<string>()
    sections.forEach((section) => {
      if (section.defaultOpen !== false) {
        open.add(section.id)
      }
    })
    return open
  })

  const toggleSection = (sectionId: string) => {
    setOpenSections((prev) => {
      const next = new Set(prev)
      if (next.has(sectionId)) {
        next.delete(sectionId)
      } else {
        next.add(sectionId)
      }
      return next
    })
  }

  return (
    <nav
      className={`sidebar bg-body-tertiary border-end h-100 ${collapsed ? 'sidebar-collapsed' : ''} ${className}`}
      style={{ width: collapsed ? '56px' : '240px', transition: 'width 0.2s' }}
    >
      <div className="sidebar-content p-2">
        {sections.map((section) => {
          const isOpen = openSections.has(section.id)
          const hasItems = section.items.length > 0

          return (
            <div key={section.id} className="sidebar-section mb-2">
              {section.collapsible && hasItems ? (
                <button
                  type="button"
                  className="btn btn-link text-decoration-none w-100 d-flex align-items-center justify-content-between p-2 text-body-secondary sidebar-section-header"
                  onClick={() => toggleSection(section.id)}
                  aria-expanded={isOpen}
                >
                  <span className="d-flex align-items-center gap-2">
                    {section.icon}
                    {!collapsed && <span className="fw-semibold small text-uppercase">{section.label}</span>}
                  </span>
                  {!collapsed && (
                    <span className="sidebar-chevron" style={{ transform: isOpen ? 'rotate(90deg)' : 'rotate(0deg)', transition: 'transform 0.2s' }}>
                      ›
                    </span>
                  )}
                </button>
              ) : (
                <div className="p-2 text-body-secondary">
                  <span className="d-flex align-items-center gap-2">
                    {section.icon}
                    {!collapsed && <span className="fw-semibold small text-uppercase">{section.label}</span>}
                  </span>
                </div>
              )}

              {(isOpen || !section.collapsible) && hasItems && (
                <ul className="nav flex-column">
                  {section.items.map((item) => (
                    <li key={item.id} className="nav-item">
                      <button
                        type="button"
                        className={`nav-link d-flex align-items-center gap-2 w-100 text-start ${activeItem === item.id ? 'active bg-primary-subtle text-primary' : 'text-body'}`}
                        onClick={() => onNavigate(item.path)}
                        title={collapsed ? item.label : undefined}
                      >
                        {item.icon}
                        {!collapsed && (
                          <>
                            <span className="flex-grow-1">{item.label}</span>
                            {item.badge !== undefined && (
                              <span className="badge bg-secondary-subtle text-secondary">{item.badge}</span>
                            )}
                          </>
                        )}
                      </button>
                    </li>
                  ))}
                </ul>
              )}
            </div>
          )
        })}
      </div>

      <style>{`
        .sidebar-section-header:hover {
          background-color: var(--bs-tertiary-bg);
        }
        .nav-link {
          border-radius: 0.375rem;
          padding: 0.5rem;
        }
        .nav-link:hover {
          background-color: var(--bs-tertiary-bg);
        }
        .sidebar-collapsed .nav-link {
          justify-content: center;
        }
      `}</style>
    </nav>
  )
}

// Helper to create navigation items with EntityIcons
export function createNavItems(): SidebarSection[] {
  return [
    {
      id: 'main',
      label: '',
      items: [
        { id: 'dashboard', label: 'Dashboard', icon: <EntityIcon type="session" size="sm" />, path: '/' },
        { id: 'search', label: 'Search', icon: <span>🔍</span>, path: '/search' },
      ],
    },
    {
      id: 'explore',
      label: 'Explore',
      collapsible: true,
      defaultOpen: true,
      items: [
        { id: 'files', label: 'Files', icon: <EntityIcon type="folder" size="sm" />, path: '/explore/files' },
        { id: 'symbols', label: 'Symbols', icon: <EntityIcon type="symbol" size="sm" />, path: '/explore/symbols' },
        { id: 'code-units', label: 'Code Units', icon: <EntityIcon type="code_unit" size="sm" />, path: '/explore/code-units' },
        { id: 'commits', label: 'Commits', icon: <EntityIcon type="commit" size="sm" />, path: '/explore/commits' },
        { id: 'diffs', label: 'Diffs', icon: <EntityIcon type="diff" size="sm" />, path: '/explore/diffs' },
        { id: 'docs', label: 'Docs/Terms', icon: <EntityIcon type="doc" size="sm" />, path: '/explore/docs' },
      ],
    },
    {
      id: 'data',
      label: 'Data/Admin',
      collapsible: true,
      defaultOpen: false,
      items: [
        { id: 'runs', label: 'All Runs', icon: <EntityIcon type="run" size="sm" />, path: '/data/runs' },
        { id: 'raw-outputs', label: 'Raw Outputs', icon: <EntityIcon type="file" size="sm" />, path: '/data/raw-outputs' },
        { id: 'schema', label: 'Schema Info', icon: <span>📋</span>, path: '/data/schema' },
      ],
    },
  ]
}
