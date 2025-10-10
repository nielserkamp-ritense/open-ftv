import {createFileRoute} from '@tanstack/react-router'
import {Button} from "@/components/button.tsx";
import Card from "@/components/card.tsx";
import {IconCircleArrowUpFilled, IconCircleRectangleFilled, IconFileExport} from "@tabler/icons-react";
import {AuthlogEntry, useAuthlogEntries} from "@/services/authlog.ts";
import type {ColDef, RowClassRules, RowClassParams} from "ag-grid-community";
import CondensedGrid from "@/components/condensed-grid.tsx";
import {CustomCellRendererProps} from "ag-grid-react";
import {useMemo} from "react";

export const Route = createFileRoute('/logboek/')({
    component: RouteComponent,
})

const BeslissingRenderer = (params: CustomCellRendererProps) => {
    if (params.value === 'permit') {
        return <div className={"flex items-center justify-center"}><IconCircleArrowUpFilled
            color={"var(--color-rhc-color-feedback-info-default)"}
            className={"mx-auto mt-1.5"}></IconCircleArrowUpFilled></div>
    } else {
        return <IconCircleRectangleFilled
            color={"var(--color-carrotnl-alert-icon-error-color)"} className={"mx-auto mt-1.5"}></IconCircleRectangleFilled>
    }
};

function RouteComponent() {
    const {data} = useAuthlogEntries();
    const colDefs: ColDef<AuthlogEntry>[] = [
        {
            headerName: "Trace ID",
            field: "traceId",
            resizable: false,
            cellStyle: {color: 'var(--color-content-secondary)'},
        },
        {
            headerName: "Tijdstempel",
            field: "created",
            resizable: false,
            cellStyle: {color: 'var(--color-content-secondary)'}
        },
        {
            headerName: "Subject",
            // The OpenAPI types define `request` as Record<string, never>, so we need a safe cast here
            valueGetter: (params) => (params.data?.request as unknown as { subject?: { id?: string } } | undefined)?.subject?.id,
            resizable: false
        },
        {
            headerName: "Beslispunt",
            valueGetter: () => "PDP 1 Vlierdam",
            resizable: false
        },
        {
            headerName: "Bundel",
            resizable: false
        },
        {
            headerName: "Regeling",
            resizable: false,
            cellStyle: {color: 'var(--color-content-secondary)'}
        },
        {
            headerName: "Actie",
            resizable: false,
            cellStyle: {color: 'var(--color-content-secondary)'}
        },
        {
            headerName: "Beslissing",
            valueGetter: (params) => params.data?.response?.decision,
            cellRenderer: BeslissingRenderer,
            type: '',
            resizable: false
        },
        {
            headerName: "Reden",
            valueGetter: params => params.data?.response?.reason,
            resizable: false,
            cellStyle: {color: 'var(--color-content-secondary)'}
        },
        {
            headerName: "",
            valueGetter: () => "...",
            sortable: false,
            resizable: false,
            type: "rightAligned",
            minWidth: 50
        }
    ]

    const rowClassRules: RowClassRules<AuthlogEntry> = useMemo(() => {
        return {
            'decision-denied': (params: RowClassParams<AuthlogEntry>) => params.data?.response?.decision === 'deny'
        }
    }, [])

    return (
        <>
            <div className="flex items-center justify-between my-4">
                <h1 className="text-rhc-lintblauw-500 text-[30px] leading-9">Logboek toegangsbeslissingen</h1>
            </div>
            <Card className="min-w-3xl flex-1 h-[836px] py-3 " disablePadding={true}>
                <div className={"px-10 gap-8"}>
                    <div className={"pt-6 text-right"}>
                        <Button color={"primary"} disabled>
                            <div className={"flex align-middle justify-center my-auto"}>
                                <IconFileExport className={"text-content-inverse-secondary"} size={20}></IconFileExport>
                            </div>
                            <span>Exporteer</span>
                        </Button>
                    </div>
                    <div className={"w-full h-[600px] mt-2"}>
                        <CondensedGrid
                            rowData={data}
                            columnDefs={colDefs}
                            rowClassRules={rowClassRules}
                        />
                    </div>
                </div>
            </Card>
        </>
    )
}
