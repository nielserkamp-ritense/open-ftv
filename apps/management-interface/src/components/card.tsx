import React from "react";
import clsx from "clsx";

export default function Card({
    className, 
    header, 
    children, 
    disablePadding = false 
}: React.PropsWithChildren<{ 
    header?: React.ReactNode, 
    className: string | undefined,
    disablePadding?: boolean
}>) {
    return (
        <div className={clsx(className, 'divide-y divide-gray-200 rounded-sm bg-white shadow-sm flex flex-col')}>
            {header && (
                <div className={clsx(!disablePadding && "px-4 py-5 sm:px-6", "min-h-[73px]")}>
                    {header}
                </div>
            )}
            <div className={clsx(!disablePadding && "px-4 py-4 sm:p-6", "flex-1")}>
                {children}
            </div>
        </div>
    )
}
