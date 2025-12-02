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
  const defaultStart = useMemo(() => "2025-11-20T07:20:50.52Z", []);
  const {data} = useAuthlogEntries({start: defaultStart});

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
            headerName: "Tijdstempel",
            field: "created",
            resizable: false,
            filter: 'agDateColumnFilter',
            cellStyle: {color: 'var(--color-content-secondary)'}
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
            valueGetter: () => "",
            resizable: true
        },
        {
            headerName: "Actie",
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
            valueGetter: params => params.data?.response?.reason,
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
