import {createFileRoute, useNavigate} from '@tanstack/react-router'
import Card from "@/components/ui/card.tsx";
import {Heading} from "@/components/ui/heading.tsx";
import { Badge } from '@/components/ui/badge.tsx';
import { useAttributes, type Attribute } from '@/services/attributes';
import { ScaleLoader } from 'react-spinners';
import { Button } from '@/components/ui/button.tsx';
import {IconPlus} from "@tabler/icons-react";
import Grid from "@/components/ui/grid.tsx";
import type {ColDef, RowClickedEvent} from "ag-grid-community";
import {useState} from "react";
import {Breadcrumb} from "@/components/ui/breadcrumb.tsx";
import {useCapabilities} from "@/auth/useCapabilities.ts";

export const Route = createFileRoute('/attributen/')({
    component: RouteComponent,
})

// Component to render Badge for type
const TypeBadgeComponent = (props: { value: string }) => {
    const type = props.value;
    return type && type !== '-' ? <Badge color="cyan">{type}</Badge> : <span>-</span>;
};

function RouteComponent() {
    const navigate = useNavigate({from: '/attributen/'})
    const { status, data, error } = useAttributes();
    const {canWrite} = useCapabilities();

    function handleRowClicked(e: RowClickedEvent<Attribute>) {
        const rowData = e.data;

        if (!rowData) {
            console.error('No data found for row clicked');
            return;
        }

        navigate({to: '/attributen/$key', params: {key: rowData.key}})
            .catch(e => console.error('Navigation error:', e));
    }

    // Column Definitions for AG Grid
    const [colDefs] = useState<ColDef<Attribute>[]>([
        {
            field: "key",
            headerName: "ID",
            filter: true,
        },
        {
            field: "metadata.title",
            headerName: "Omschrijving",
            filter: true,
        },
        {
            field: "type",
            headerName: "Type",
            cellRenderer: TypeBadgeComponent,
            width: 150,
        },
        {
            headerName: "Gepubliceerd",
            valueGetter: () => "-",
        },
        {
          headerName: "Concept",
          valueGetter: () => "18-11-2025 13:26",
        },
        {
            headerName: "Gebruik",
            valueGetter: (params) => {
                const usageCount = params.data?.usageData?.reduce((sum, x) => sum + (x.bundle?.length ?? 0), 0) ?? 0;
                return usageCount > 0 ? `${usageCount} Beleidsregel(s)` : '0';
            },
        },
    ]);

    if (status === 'pending') {
        return (
            <>
                <div className="flex items-center justify-between">
                    <Heading className="lg:text-3xl ">Context</Heading>
                </div>
                <div className="flex flex-col 2xl:flex-row py-3 gap-6">
                    <div className="mx-auto mt-4 flex w-[200px] items-center justify-center gap-y-2 flex-col">
                        <ScaleLoader height={16} />
                        <div>Loading data...</div>
                    </div>
                </div>
            </>
        );
    }

    if (status === 'error') {
        return <span>Error: {error.message}</span>;
    }

    return (<>
        <Breadcrumb />
        <div className="flex items-center justify-between my-4">
            <h1 className="text-rhc-lintblauw-500 text-[30px] leading-9">Context</h1>
            {canWrite && (
            <Button href="/attributen/toevoegen" color={"primary"}>
                <div className={"flex align-middle justify-center my-auto"}>
                    <IconPlus className={"text-content-inverse-secondary"} size={20}></IconPlus>
                </div>
                <span>Aanmaken</span>
            </Button>
            )}
        </div>
        <Card className="min-w-3xl flex-1 h-[836px] py-3" disablePadding={true}>
            <div className="pt-6 pb-8 px-10 gap-8 flex-1">
                <div className="flex justify-between items-center">
                    <div>
                        <span className={"text-rhc-lintblauw-500 text-[20px] font-normal"}>Statische context</span>
                    </div>
                </div>
                <div className={"w-full h-[700px] mt-6"}>
                    <Grid
                        onRowClicked={handleRowClicked}
                        rowData={data}
                        columnDefs={colDefs}
                    />
                </div>
            </div>
        </Card>
    </>)
}
