import {createFileRoute, useNavigate} from '@tanstack/react-router'
import {Button} from "@/components/ui/button.tsx";
import Card from "@/components/ui/card.tsx";
import {IconCircleArrowUpFilled, IconCircleRectangleFilled, IconFileExport} from "@tabler/icons-react";
import {AuthlogEntry, useAuthlogEntries} from "@/services/authlog.ts";
import type {ColDef, RowClassRules, RowClassParams, RowClickedEvent} from "ag-grid-community";
import CondensedGrid from "@/components/ui/condensed-grid.tsx";
import {CustomCellRendererProps} from "ag-grid-react";
import {useMemo} from "react";
import {Breadcrumb} from "@/components/ui/breadcrumb.tsx";
import {formatDateTime} from "@/utilities/datetime.ts";

export const Route = createFileRoute('/logboek/')({
    component: RouteComponent,
})

const BeslissingRenderer = (params: CustomCellRendererProps) => {
    if (params.value) {
        return <div className={"flex items-center justify-center"}><IconCircleArrowUpFilled
            color={"var(--color-rhc-color-feedback-info-default)"}
            className={"mx-auto mt-1.5"}></IconCircleArrowUpFilled></div>
    } else {
        return <IconCircleRectangleFilled
            color={"var(--color-carrotnl-alert-icon-error-color)"}
            className={"mx-auto mt-1.5"}></IconCircleRectangleFilled>
    }
};

function RouteComponent() {
    const navigate = useNavigate();
    const { data } = useAuthlogEntries({recent: "168h"});

    const handleRowClick = (event: RowClickedEvent<AuthlogEntry>) => {
        if (event.data?.id) {
          navigate({
                to: '/logboek/$id',
                params: {id: String(event.data.id)},
                // @ts-expect-error - The type state is not defined in the type definition
                state: {entry: event.data}
            }).catch(e => console.error('Navigation error:', e));
        }
    };

    const colDefs: ColDef<AuthlogEntry>[] = [
        {
            headerName: "Trace ID",
            field: "traceId",
            resizable: true,
            cellStyle: {color: 'var(--color-content-secondary)'},
            filter: true,
        },
        {
            headerName: "Span ID",
            field: "spanId",
            resizable: true,
            cellStyle: {color: 'var(--color-content-secondary)'},
            filter: true,
        },
        {
            headerName: "Parent Span ID",
            field: "parentSpanId",
            resizable: true,
            cellStyle: {color: 'var(--color-content-secondary)'},
            filter: true,
        },
        {
            headerName: "Event",
            field: "eventName",
            resizable: true,
            filter: true,
        },
        {
            headerName: "Status",
            field: "status",
            resizable: false,
            filter: true,
        },
        {
            headerName: "Tijdstempel",
            field: "created",
            resizable: false,
            filter: 'agDateColumnFilter',
            cellStyle: {color: 'var(--color-content-secondary)'},
            // eslint-disable-next-line @typescript-eslint/no-explicit-any
            valueFormatter: (params: any) => params.value ? formatDateTime(params.value) : '',
        },
        {
            headerName: "Subject",
            // The OpenAPI types define `request` as Record<string, never>, so we need a safe cast here
            valueGetter: (params) => (params.data?.request as unknown as {
                subject?: { id?: string }
            } | undefined)?.subject?.id,
            resizable: true
        },
        {
            headerName: "Beslispunt",
            valueGetter: (params) => (params.data?.resource as unknown as { service?: string } | undefined)?.service,
            resizable: true
        },
        {
            headerName: "Actie",
            valueGetter: (params) => (params.data?.request as unknown as {
                action?: { name?: string }
            } | undefined)?.action?.name,
            resizable: true,
            cellStyle: {color: 'var(--color-content-secondary)'}
        },
        {
            headerName: "Beslissing",
            valueGetter: (params) => params.data?.response?.decision,
            cellRenderer: BeslissingRenderer,
            type: '',
            resizable: true
        },
        {
            headerName: "Reden",
            valueGetter: (params) => {
                const context = (params.data?.response as unknown as {
                    context?: { reason_user?: Record<string, string>, reason_admin?: Record<string, string> }
                } | undefined)?.context;
                const reasons = context?.reason_user ?? context?.reason_admin;
                return reasons ? Object.values(reasons)[0] : undefined;
            },
            resizable: false,
            cellStyle: {color: 'var(--color-content-secondary)'}
        }
    ]

    const rowClassRules: RowClassRules<AuthlogEntry> = useMemo(() => {
        return {
            'decision-denied': (params: RowClassParams<AuthlogEntry>) => params.data?.response?.decision === 'deny'
        }
    }, [])

    return (
        <>
            <Breadcrumb />
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
                            onRowClicked={handleRowClick}
                            rowSelection={undefined}
                        />
                    </div>
                </div>
            </Card>
        </>
    )
}
