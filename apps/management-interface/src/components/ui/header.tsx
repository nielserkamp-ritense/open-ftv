import { Link } from '@tanstack/react-router'
import { useProfile } from '@/auth/useProfile'
import { DEFAULT_HEADER_COLOR, DEFAULT_HEADER_TITLE, DEFAULT_TITLE_COLOR, useSettings } from '@/services/settings'

interface HeaderProps {
    onLogoutHandler: () => void;
}

const menuItemClass = 'w-full flex items-center gap-2 rounded-lg px-3 py-1.5 text-left text-sm/6 text-zinc-950 cursor-pointer hover:bg-blue-500 hover:text-white focus:bg-blue-500 focus:text-white focus:outline-hidden'

export function Header({ onLogoutHandler }: HeaderProps) {
    const { data: settings } = useSettings()
    const { displayName, initials } = useProfile()

    return <div
        className="w-full h-[108px] flex items-center justify-between pl-rhc-space-300 pr-rhc-space-500 border-b border-black"
        style={{ backgroundColor: settings?.headerColor ?? DEFAULT_HEADER_COLOR }}
    >
        <div className="flex items-center gap-4 min-w-0 h-full">
            {settings?.logo && (
                <div className="h-full flex items-center -ml-6 py-2 pl-2">
                    <img
                        src={`data:${settings.logoMediaType};base64,${settings.logo}`}
                        alt="Logo"
                        className="h-full w-auto object-contain"
                    />
                </div>
            )}
            <h1
                className="font-light text-[22px] leading-[28px] truncate"
                style={{ color: settings?.titleColor ?? DEFAULT_TITLE_COLOR }}
            >
                {settings?.headerTitle ?? DEFAULT_HEADER_TITLE}
            </h1>
        </div>
        <div className="relative group">
            <button
                type="button"
                aria-haspopup="menu"
                aria-label={`Gebruikersmenu — ${displayName}`}
                title={displayName}
                className="w-[40px] h-[40px] bg-background-inverse-primary rounded-md flex items-center justify-center cursor-pointer"
            >
                <span className="text-white">{initials}</span>
            </button>

            <div className="absolute right-0 top-0 z-50 pt-12 hidden group-hover:block group-focus-within:block">
                <div className="min-w-40 rounded-xl bg-white p-1 shadow-lg ring-1 ring-zinc-950/10">
                    <Link to="/profiel" className={menuItemClass}>
                        Profiel
                    </Link>
                    <button
                        type="button"
                        onClick={onLogoutHandler}
                        className={menuItemClass}
                    >
                        Uitloggen
                    </button>
                </div>
            </div>
        </div>
    </div>;
}
