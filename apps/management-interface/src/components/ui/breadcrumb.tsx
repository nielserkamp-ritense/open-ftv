import { Link } from '@tanstack/react-router'
import { useMatches } from '@tanstack/react-router'
import { IconChevronRight } from '@tabler/icons-react'
import clsx from 'clsx'

export interface BreadcrumbItem {
  label: string
  href?: string
}

interface BreadcrumbProps {
  items?: BreadcrumbItem[]
  className?: string
}

/**
 * Breadcrumb component that displays navigation hierarchy
 *
 * Usage with auto-generation from routes:
 * <Breadcrumb />
 *
 * Usage with custom items:
 * <Breadcrumb items={[
 *   { label: 'Home', href: '/' },
 *   { label: 'Policies', href: '/policies' },
 *   { label: 'Details' }
 * ]} />
 */
export function Breadcrumb({ items, className }: BreadcrumbProps) {
  const matches = useMatches()

  // If custom items provided, use them; otherwise generate from route matches
  const breadcrumbItems: BreadcrumbItem[] = items ?? generateBreadcrumbsFromMatches(matches)

  if (breadcrumbItems.length === 0) {
    return null
  }

  if (breadcrumbItems.length === 1 && breadcrumbItems[0]?.label !== 'Home') {
    breadcrumbItems.unshift({label: 'Home', href: '/'})
  }

  return (
    <nav aria-label="Breadcrumb" className={clsx('flex items-center gap-2 text-sm mb-4', className)}>
      <ol className="flex items-center gap-2">
        {breadcrumbItems.map((item, index) => {
          const isLast = index === breadcrumbItems.length - 1

          return (
            <li key={index} className="flex items-center gap-2">
              {index > 0 && (
                <IconChevronRight
                  className="h-4 w-4 text-content-secondary"
                  aria-hidden="true"
                />
              )}
              {item.href && !isLast ? (
                <Link
                  to={item.href}
                  className="text-carrotnl-breadcrumb-link-color underline text-carrotnl-breadcrumb-fontsize transition-colors"
                >
                  {item.label.toUpperCase()}
                </Link>
              ) : (
                <span
                  className={clsx(
                    isLast ? 'text-content-primary font-medium' : 'text-content-secondary', 'text-carrotnl-breadcrumb-fontsize'
                  )}
                  aria-current={isLast ? 'page' : undefined}
                >
                  {item.label.toUpperCase()}
                </span>
              )}
            </li>
          )
        })}
      </ol>
    </nav>
  )
}

/**
 * Generates breadcrumb items from TanStack Router matches
 * Override route breadcrumb labels by adding a `breadcrumb` property to route options
 */
function generateBreadcrumbsFromMatches(matches: any[]): BreadcrumbItem[] {
  const items: BreadcrumbItem[] = []

  // Add home breadcrumb
  if (matches.length > 0 && matches[0].pathname !== '/') {
    items.push({ label: 'Home', href: '/' })
  }

  matches.forEach((match) => {
    // Skip root route
    if (match.pathname === '/') return

    // Get breadcrumb label from route context or generate from path
    const label = match.context?.breadcrumb ||
                  match.routeContext?.breadcrumb ||
                  generateLabelFromPath(match.pathname)

    if (label) {
      items.push({
        label,
        href: match.pathname
      })
    }
  })

  return items
}

/**
 * Generates a readable label from a route path
 * Example: '/policies/add' -> 'Add'
 */
function generateLabelFromPath(path: string): string {
  const segments = path.split('/').filter(Boolean)
  const lastSegment = segments[segments.length - 1]

  if (!lastSegment) return ''

  // Handle dynamic route params (e.g., $id)
  if (lastSegment.startsWith('$')) {
    return lastSegment.slice(1).charAt(0).toUpperCase() + lastSegment.slice(2)
  }

  // Capitalize first letter
  return lastSegment.charAt(0).toUpperCase() + lastSegment.slice(1)
}
