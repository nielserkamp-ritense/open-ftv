interface HeaderProps {
    onLogoutHandler: () => void;
}

export function Header({ onLogoutHandler }: HeaderProps) {
    return <div className="w-full h-[108px] flex items-center justify-between py-rhc-space-300 px-rhc-space-500 bg-heading-primary border-b border-black">
        <div>
            <h1 className="font-light text-[22px] leading-[20px]">OpenFTV beheeromgeving</h1>
        </div>
        <div className="relative group">
            <button
                type="button"
                aria-haspopup="menu"
                aria-label="Gebruikersmenu"
                className="w-[40px] h-[40px] bg-background-inverse-primary rounded-md flex items-center justify-center cursor-pointer"
            >
                <span className="text-white">TW</span>
            </button>

            <div className="absolute right-0 top-0 z-50 pt-12 hidden group-hover:block group-focus-within:block">
                <div className="min-w-40 rounded-xl bg-white p-1 shadow-lg ring-1 ring-zinc-950/10">
                    <button
                        type="button"
                        onClick={onLogoutHandler}
                        className="w-full flex items-center gap-2 rounded-lg px-3 py-1.5 text-left text-sm/6 text-zinc-950 cursor-pointer hover:bg-blue-500 hover:text-white focus:bg-blue-500 focus:text-white focus:outline-hidden"
                    >
                        Logout
                    </button>
                </div>
            </div>
        </div>
    </div>;
}
