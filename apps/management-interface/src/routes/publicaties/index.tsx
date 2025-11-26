import {Button} from "@/components/ui/button.tsx";
import Grid from "@/components/ui/grid.tsx";
import {useDeployments, useStatuses, useStartDeployment, type Deployment} from "@/services/bundles.ts";
import {useState} from "react";
import {createFileRoute} from "@tanstack/react-router";
import Card from "@/components/ui/card.tsx";
import type {ColDef} from "ag-grid-community";
import {ScaleLoader} from "react-spinners";
import {Breadcrumb} from "@/components/ui/breadcrumb.tsx";

export const Route = createFileRoute('/publicaties/')({
    component: PublicatiesComponent,
})

// Helper function to format RFC3339 timestamp to "DD-MM-YYYY HH:mm"
function formatDateTime(isoString: string): string {
    const date = new Date(isoString);
    const day = String(date.getDate()).padStart(2, '0');
    const month = String(date.getMonth() + 1).padStart(2, '0');
    const year = date.getFullYear();
    const hours = String(date.getHours()).padStart(2, '0');
    const minutes = String(date.getMinutes()).padStart(2, '0');
    return `${day}-${month}-${year} ${hours}:${minutes}`;
}

export default function PublicatiesComponent() {
    const {data, isLoading, isError} = useDeployments();
    const {data: statusesData} = useStatuses();
    const { mutate: startDeployment, isPending, error: startError } = useStartDeployment();

    // Create a lookup map for status codes to labels
    const statusLookup = statusesData?.reduce((acc, status) => {
        acc[status.code] = status.name;
        return acc;
    }, {} as Record<number, string>) || {};

    // Column definitions for deployments table
    const [colDefs] = useState<ColDef<Deployment>[]>([

        {field: "title", headerName: "Title"},
        {field: "description", headerName: "Description"},
        {
            field: "audit.created",
            headerName: "Created",
            // eslint-disable-next-line @typescript-eslint/no-explicit-any
            valueFormatter: (params: any) => params.value ? formatDateTime(params.value) : ''
        },
        {field: "audit.createdBy", headerName: "Created By", flex: 1},
        {
            field: "status",
            headerName: "Status",
            // eslint-disable-next-line @typescript-eslint/no-explicit-any
            valueFormatter: (params: any) => statusLookup[params.value] || `Status ${params.value}`
        },
        {field: "version", headerName: "Versie", width: 25},
        {field: "message", headerName: "Message", cellStyle: {color: 'var(--color-content-secondary)'} },
    ]);

    if (isLoading) {
        return (
            <>
                <div className="flex items-center justify-between my-4">
                    <h1 className="text-rhc-lintblauw-500 text-[30px] leading-9">Publicaties</h1>
                </div>
                <div className="mx-auto mt-4 flex w-[200px] items-center justify-center gap-y-2 flex-col">
                    <ScaleLoader height={16}/>
                    <div>Laden...</div>
                </div>
            </>
        )
    }

    if (isError) {
        return (
            <>
                <div className="flex items-center justify-between my-4">
                    <h1 className="text-rhc-lintblauw-500 text-[30px] leading-9">Publicaties</h1>
                </div>
                <div className="text-danger-600">Er is een fout opgetreden bij het laden van bundels.</div>
            </>
        );
    }

    return (
        <>
            <Breadcrumb />
            <div className="flex items-center justify-between my-4">
                <h1 className="text-rhc-lintblauw-500 text-[30px] leading-9">Publicaties</h1>
            </div>
            <Card className="min-w-3xl flex-1 h-[836px] py-3 flex" disablePadding={true}>
                <div className={"pt-6 pb-8 px-10 gap-8 flex-1"}>
                    <div className="flex justify-between items-center">
                        <div>
                            <span className={"text-rhc-lintblauw-500 text-[20px] font-normal"}>Publicatie overzicht</span>
                        </div>
                        <div className="flex items-center gap-3">
                            {startError && (
                                <span className="text-danger-600 text-sm">Starten van publicatie mislukt</span>
                            )}
                            <Button
                                color={"primary"}
                                disabled={isPending}
                                onClick={() => {
                                    // Minimal body for starting a deployment
                                    startDeployment({
                                        title: "Publicatie",
                                        description: undefined,
                                    });
                                }}
                            >
                                <span>{isPending ? 'Publiceren…' : 'Publiceren'}</span>
                            </Button>
                        </div>
                    </div>
                    <div className={"w-full h-[600px] mt-2"}>
                        <Grid
                            rowData={data}
                            columnDefs={colDefs}
                        />
                    </div>
                </div>
            </Card>
        </>
    )
}

