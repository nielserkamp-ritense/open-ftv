import type { ComponentType } from 'react'
import {
  IconChecklist,
  IconDatabase,
  IconFileText,
  IconHelpCircleFilled,
  IconLayoutDashboard,
  IconLogs,
  IconNews,
  IconSectionSign,
  IconSettings,
} from '@tabler/icons-react'

export type MenuSection = 'top' | 'bottom'

export interface MenuItem {
  href: string
  label: string
  icon: ComponentType
  disabled?: boolean
  section: MenuSection
}

export const menuItems: MenuItem[] = [
  { href: '/', label: 'Overzicht', icon: IconLayoutDashboard, section: 'top' },
  { href: '/policies', label: 'Beleidsregels', icon: IconFileText, section: 'top' },
  { href: '#', label: 'Test cases', icon: IconChecklist, disabled: true, section: 'top' },
  { href: '/attributen', label: 'Bronnen', icon: IconDatabase, section: 'top' },
  { href: '/regelingen', label: 'Regelingen', icon: IconSectionSign, disabled: true, section: 'top' },
  { href: '/logboek', label: 'Logboek', icon: IconLogs, disabled: false, section: 'top' },
  { href: '#', label: 'Systeem', icon: IconSettings, disabled: true, section: 'top' },
  { href: '#', label: 'Ondersteuning', icon: IconHelpCircleFilled, section: 'bottom' },
  { href: '#', label: 'Nieuws', icon: IconNews, section: 'bottom' },
]
