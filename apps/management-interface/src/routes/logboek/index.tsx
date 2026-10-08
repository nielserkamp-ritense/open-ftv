import {createFileRoute, useNavigate} from '@tanstack/react-router'
import {Button} from "@/components/ui/button.tsx";
import Card from "@/components/ui/card.tsx";
import {IconAlertTriangleFilled, IconCircleArrowUpFilled, IconCircleRectangleFilled, IconFileExport} from "@tabler/icons-react";
import {AuthlogEntry, useAuthlogEntries} from "@/services/authlog.ts";
import type {ColDef, FilterChangedEvent, GridApi, GridReadyEvent, RowClassRules, RowClassParams, RowClickedEvent, TextFilterModel} from "ag-grid-community";
import CondensedGrid from "@/components/ui/condensed-grid.tsx";
import {CustomCellRendererProps} from "ag-grid-react";
import {useEffect, useMemo, useState} from "react";
import {Breadcrumb} from "@/components/ui/breadcrumb.tsx";
import {formatDateTime} from "@/utilities/datetime.ts";
import {Decision, entryDecision, entryReason} from "@/services/decision.ts";

interface LogboekSearch {
    traceId?: string
}

export const Route = createFileRoute('/logboek/')({
    validateSearch: (search: Record<string, unknown>): LogboekSearch => ({
        traceId: typeof search.traceId === 'string' && search.traceId !== '' ? search.traceId : undefined,
    }),
    component: RouteComponent,
})

const BeslissingRenderer = (params: CustomCellRendererProps<AuthlogEntry, Decision>) => {
    switch (params.value) {
        case 'permitted':
            return <>
                <IconCircleArrowUpFilled
                    color={"var(--color-rhc-color-feedback-info-default)"}
                    className={"mx-auto mt-1.5"} aria-hidden="true"></IconCircleArrowUpFilled>
                <span className={"sr-only"}>Toegestaan</span>
            </>
        case 'denied':
            return <>
                <IconCircleRectangleFilled
                    color={"var(--color-carrotnl-alert-icon-error-color)"}
                    className={"mx-auto mt-1.5"} aria-hidden="true"></IconCircleRectangleFilled>
                <span className={"sr-only"}>Geweigerd</span>
            </>
        case 'error':
            return <>
                <IconAlertTriangleFilled
                    color={"var(--color-rhc-color-feedback-warning-default)"}
                    className={"mx-auto mt-1.5"} aria-hidden="true"></IconAlertTriangleFilled>
                <span className={"sr-only"}>Fout</span>
            </>
        default:
            return null
    }
};

const traceIdFilterModel = (traceId?: string): TextFilterModel | null =>
    traceId ? {filterType: 'text', type: 'equals', filter: traceId} : null;

const filteredTraceId = (api: GridApi<AuthlogEntry>): string | undefined => {
    const value = api.getColumnFilterModel<TextFilterModel>('traceId')?.filter?.trim().toLowerCase();
    return value === '' ? undefined : value;
};

function RouteComponent() {
    const navigate = useNavigate();
    const {traceId} = Route.useSearch();
    // A trace ID searches the whole log, so its decision is found however old the decision is.
    const { data } = useAuthlogEntries(traceId ? {traceId} : {recent: "168h"});

    const [gridApi, setGridApi] = useState<GridApi<AuthlogEntry>>();

    const handleGridReady = (event: GridReadyEvent<AuthlogEntry>) => setGridApi(event.api);

    // The Trace ID filter of the grid searches the whole log, through the URL, instead of the rows it has.
    const handleFilterChanged = (event: FilterChangedEvent<AuthlogEntry>) => {
        const value = filteredTraceId(event.api);
        if (value !== traceId) {
            navigate({to: '/logboek', search: {traceId: value}})
                .catch(e => console.error('Navigation error:', e));
        }
    };

    // A link or the back button changes the URL with the grid already open, so its filter follows the URL.
    useEffect(() => {
        if (!gridApi || filteredTraceId(gridApi) === traceId) return;

        gridApi.setColumnFilterModel('traceId', traceIdFilterModel(traceId))
            .then(() => gridApi.onFilterChanged())
            .catch(e => console.error('Filter error:', e));
    }, [gridApi, traceId]);

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
            filter: 'agTextColumnFilter',
            // The decision log only finds a whole trace ID, and Enter applies it, not every key typed.
            filterParams: {filterOptions: ['equals'], maxNumConditions: 1, buttons: ['apply', 'reset'], closeOnApply: true},
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
            valueGetter: (params) => {
                const resource = params.data?.resource as unknown as { organization?: string, instanceId?: string } | undefined;
                if (!resource?.organization && !resource?.instanceId) return undefined;
                return [resource.organization, resource.instanceId].filter(Boolean).join(' / ');
            },
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
            valueGetter: (params) => entryDecision(params.data),
            cellRenderer: BeslissingRenderer,
            type: '',
            resizable: true
        },
        {
            headerName: "Reden",
            valueGetter: (params) => entryReason(params.data),
            resizable: false,
            cellStyle: {color: 'var(--color-content-secondary)'}
        }
    ]

    const rowClassRules: RowClassRules<AuthlogEntry> = useMemo(() => {
        return {
            'decision-denied': (params: RowClassParams<AuthlogEntry>) => entryDecision(params.data) === 'denied',
            'decision-error': (params: RowClassParams<AuthlogEntry>) => entryDecision(params.data) === 'error',
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
                            onGridReady={handleGridReady}
                            onFilterChanged={handleFilterChanged}
                            initialState={{filter: {filterModel: traceId ? {traceId: traceIdFilterModel(traceId)} : {}}}}
                            rowSelection={undefined}
                        />
                    </div>
                </div>
            </Card>
        </>
    )
}
