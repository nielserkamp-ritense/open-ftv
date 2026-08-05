import type { ComponentType } from 'react'
import {
  IconChecklist,
  IconDatabase,
  IconFileText,
  IconHelpCircleFilled,
  IconLayoutDashboard,
  IconLogs,
  IconNews, IconPencilShare,
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
  { href: '#', label: 'Tests', icon: IconChecklist, disabled: true, section: 'top' },
  { href: '/attributen', label: 'Context', icon: IconDatabase, section: 'top' },
  { href: '/regelingen', label: 'Regelingen', icon: IconSectionSign, disabled: true, section: 'top' },
  { href: '/publicaties', label: 'Publicaties', icon: IconPencilShare, section: 'top' },
  { href: '/logboek', label: 'Logboek', icon: IconLogs, disabled: false, section: 'top' },
  { href: '/instellingen', label: 'Instellingen', icon: IconSettings, section: 'top' },
  { href: '#', label: 'Ondersteuning', icon: IconHelpCircleFilled, section: 'bottom', disabled: true },
  { href: '#', label: 'Nieuws', icon: IconNews, section: 'bottom', disabled: true},
]
