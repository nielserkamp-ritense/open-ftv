import React from "react";
import clsx from "clsx";

export default function Card({className, header, children }: React.PropsWithChildren<{ header: React.ReactNode, className: string | undefined }>) {
    return (
        <div className={clsx(className, 'divide-y divide-gray-200 rounded-sm bg-white shadow-sm flex flex-col')}>
            <div className="px-4 py-5 sm:px-6">
                {header}
            </div>
            <div className="px-4 py-5 sm:p-6 flex-1">
                {children}
            </div>
        </div>
    )
}
