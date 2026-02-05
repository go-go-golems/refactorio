export interface EntityIconProps {
  /** Entity type */
  type: 'symbol' | 'code_unit' | 'commit' | 'diff' | 'doc' | 'file' | 'folder' | 'run' | 'session'
  /** Sub-type for more specific icons (e.g., 'func', 'type' for symbols) */
  kind?: string
  /** Size variant */
  size?: 'sm' | 'md' | 'lg'
  /** Custom class name */
  className?: string
}

// Icon components using simple SVG paths
const icons = {
  symbol: {
    default: (
      <path d="M8 0a8 8 0 100 16A8 8 0 008 0zm0 14.5a6.5 6.5 0 110-13 6.5 6.5 0 010 13zM7 4h2v4.5H7V4zm0 6h2v2H7v-2z" />
    ),
    func: <path d="M2 4h2v8H2V4zm4-2h6v2H6V2zm0 4h4v2H6V6zm0 4h6v2H6v-2zm6-6h2v8h-2V4z" />,
    type: <path d="M0 2h16v3H9v9H7V5H0V2z" />,
    method: (
      <path d="M2 4h2v8H2V4zm4-2h6v2H6V2zm0 4h4v2H6V6zm0 4h6v2H6v-2zm6-6h2v8h-2V4zM0 0h1v16H0V0z" />
    ),
    const: <path d="M3 2h10v2H3V2zm0 4h10v2H3V6zm0 4h10v2H3v-2zm0 4h10v2H3v-2z" />,
    var: <path d="M2 2l6 12 6-12h-3L8 9 5 2H2z" />,
  },
  code_unit: <path d="M4 2a2 2 0 00-2 2v8a2 2 0 002 2h8a2 2 0 002-2V4a2 2 0 00-2-2H4zm1 3h6v2H5V5zm0 4h4v2H5V9z" />,
  commit: (
    <path d="M8 0a8 8 0 100 16A8 8 0 008 0zm0 11a3 3 0 110-6 3 3 0 010 6zm0-4a1 1 0 100 2 1 1 0 000-2z" />
  ),
  diff: <path d="M2 0h12v16H2V0zm2 2v12h8V2H4zm1 2h2v2H5V4zm0 4h6v2H5V8zm0 4h4v2H5v-2zm4-8h2v2H9V4z" />,
  doc: <path d="M2 0h8l4 4v12H2V0zm8 1v3h3L10 1zM4 6h8v2H4V6zm0 4h8v2H4v-2z" />,
  file: <path d="M2 0h8l4 4v12H2V0zm8 1v3h3L10 1z" />,
  folder: <path d="M0 2h6l2 2h8v10H0V2zm2 4v6h12V6H2z" />,
  run: (
    <path d="M8 0a8 8 0 100 16A8 8 0 008 0zm-.5 4l5 4-5 4V4z" />
  ),
  session: (
    <path d="M0 3h16v2H0V3zm0 4h16v2H0V7zm0 4h16v2H0v-2z" />
  ),
}

const sizeMap = {
  sm: 12,
  md: 16,
  lg: 20,
}

const colorMap: Record<string, string> = {
  symbol: '#6f42c1',
  code_unit: '#0d6efd',
  commit: '#198754',
  diff: '#fd7e14',
  doc: '#6c757d',
  file: '#495057',
  folder: '#ffc107',
  run: '#20c997',
  session: '#0dcaf0',
  func: '#0d6efd',
  type: '#6f42c1',
  method: '#d63384',
  const: '#198754',
  var: '#fd7e14',
}

export function EntityIcon({ type, kind, size = 'md', className = '' }: EntityIconProps) {
  const iconSize = sizeMap[size]
  const color = colorMap[kind || type] || colorMap[type]

  let iconPath: React.ReactNode
  if (type === 'symbol' && kind && icons.symbol[kind as keyof typeof icons.symbol]) {
    iconPath = icons.symbol[kind as keyof typeof icons.symbol]
  } else if (type === 'symbol') {
    iconPath = icons.symbol.default
  } else {
    iconPath = icons[type]
  }

  return (
    <svg
      width={iconSize}
      height={iconSize}
      viewBox="0 0 16 16"
      fill={color}
      className={className}
      aria-hidden="true"
    >
      {iconPath}
    </svg>
  )
}
