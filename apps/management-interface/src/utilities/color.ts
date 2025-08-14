// Utility to deterministically map a string to a Badge color key.
// Keep this palette in sync with src/components/badge.tsx
export type BadgeColorKey =
    | 'red'
    | 'orange'
    | 'amber'
    | 'yellow'
    | 'lime'
    | 'green'
    | 'emerald'
    | 'teal'
    | 'cyan'
    | 'sky'
    | 'blue'
    | 'indigo'
    | 'violet'
    | 'purple'
    | 'fuchsia'
    | 'pink'
    | 'rose'
    | 'zinc'

const BADGE_COLOR_KEYS: BadgeColorKey[] = [
    'red',
    'orange',
    'amber',
    'yellow',
    'lime',
    'green',
    'emerald',
    'teal',
    'cyan',
    'sky',
    'blue',
    'indigo',
    'violet',
    'purple',
    'fuchsia',
    'pink',
    'rose',
    'zinc',
]

// djb2 hash for strings
function hashString(input: string): number {
    let hash = 5381
    for (let i = 0; i < input.length; i++) {
        hash = ((hash << 5) + hash) + input.charCodeAt(i) // hash * 33 + c
        hash |= 0 // force 32-bit
    }
    return Math.abs(hash)
}

export function badgeColorKeyFromString(input: string | undefined | null): BadgeColorKey {
    if (!input) {
        return 'zinc'
    }

    const idx = hashString(input.toLowerCase()) % BADGE_COLOR_KEYS.length
    return BADGE_COLOR_KEYS[idx]!
}
