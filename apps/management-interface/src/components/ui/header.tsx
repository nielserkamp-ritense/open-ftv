export function Header() {
    return <div className="w-full h-[108px] flex items-center justify-between py-rhc-space-300 px-rhc-space-500 bg-heading-primary border-b border-black">
        <div>
            <h1 className="font-light text-[22px] leading-[20px]">OpenFTV beheeromgeving</h1>
        </div>
        <div>
            <div className="w-[40px] h-[40px] bg-background-inverse-primary rounded-md flex items-center justify-center">
                <span className="text-white">TW</span>
            </div>
        </div>
    </div>;
}
