export interface Tab {
  id: string
  label: string
  badge?: string | number
  disabled?: boolean
}

export interface TabNavProps {
  /** Tab definitions */
  tabs: Tab[]
  /** Currently active tab ID */
  activeTab: string
  /** Called when a tab is selected */
  onChange: (tabId: string) => void
}

export function TabNav({ tabs, activeTab, onChange }: TabNavProps) {
  return (
    <ul className="nav nav-tabs" role="tablist">
      {tabs.map((tab) => (
        <li className="nav-item" key={tab.id} role="presentation">
          <button
            type="button"
            className={`nav-link ${tab.id === activeTab ? 'active' : ''} ${tab.disabled ? 'disabled' : ''}`}
            role="tab"
            aria-selected={tab.id === activeTab}
            aria-disabled={tab.disabled}
            tabIndex={tab.disabled ? -1 : 0}
            onClick={() => !tab.disabled && onChange(tab.id)}
          >
            {tab.label}
            {tab.badge != null && (
              <span className="badge bg-secondary rounded-pill ms-1" style={{ fontSize: '0.7em' }}>
                {tab.badge}
              </span>
            )}
          </button>
        </li>
      ))}
    </ul>
  )
}
