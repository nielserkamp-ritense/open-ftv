import {Heading} from '@/components/heading'
import {PolicyResponse, useDeletePolicy, usePolicies} from "@/services/policies.ts";
import {ScaleLoader} from "react-spinners";
import {createFileRoute, useNavigate} from "@tanstack/react-router";
import {Button} from "@/components/button.tsx";
import {useState} from "react";
import {Alert, AlertActions, AlertDescription, AlertTitle} from "@/components/alert.tsx";
import Card from "@/components/card.tsx";
import type {ColDef, RowClickedEvent} from "ag-grid-community";
import {IconBoltFilled, IconPlus} from "@tabler/icons-react";
import {Badge} from "@/components/badge";
import Grid from "@/components/grid.tsx";

function formatDateTime(value?: string) {
    if (!value) return '-';
    const d = new Date(value);
    if (isNaN(d.getTime())) return value;
    const dd = String(d.getDate()).padStart(2, '0');
    const mm = String(d.getMonth() + 1).padStart(2, '0');
    const yyyy = d.getFullYear();
    const hh = String(d.getHours()).padStart(2, '0');
    const min = String(d.getMinutes()).padStart(2, '0');
    return `${dd}-${mm}-${yyyy}, ${hh}:${min}`;
}

export const Route = createFileRoute('/policies/')({
    component: PoliciesComponent,
})

const SimpleComponent = () => (
    <Badge color="ftvgreen" className={"text-[12px] leading-none px-1.5 py-0.5"}>
        <IconBoltFilled className="w-4 h-4 shrink-0 align-middle" />
        <span className="leading-none">v 2.1</span>
    </Badge>
);

export default function PoliciesComponent() {
    const navigate = useNavigate({from: '/policies'})
    const {data, isLoading, error} = usePolicies();
    const deletePolicyMutation = useDeletePolicy();
    const [isOpen, setIsOpen] = useState(false)
    const [selectedId] = useState<string | null>(null)

    function handleRowClicked(e: RowClickedEvent<PolicyResponse>) {
        const data = e.data;

        if (!data) {
            console.error('No data found for row clicked');
            return
        }

        navigate({to: '/policies/$id', params: {id: data.id}})
            .catch(e => console.error('Navigation error:', e));
    }

    // Column Definitions: Defines & controls grid columns.
    const [colDefs] = useState<ColDef<PolicyResponse>[]>([
        {field: "metadata.title", headerName: "Titel", filter: true,},
        {headerName: "Doelbinding", valueGetter: () => "Laadpalen", cellStyle: {color: 'var(--color-content-secondary)'}},
        {headerName: "Actief", valueGetter: () => "v 2.14", cellRenderer: SimpleComponent, width: 100},
        {headerName: "Concept- en ingeplande versie", valueGetter: () => ""},
        {
            field: "audit.updated",
            headerName: "Laatst bewerkt", valueFormatter: (p) => {
                const value = p.value as string | undefined;
                return formatDateTime(value);
            },
            cellStyle: {color: 'var(--color-content-secondary)'}
        },
        {
            headerName: "Beslispunt",
            valueGetter: () => "Zaaksysteem",
            cellStyle: {color: 'var(--color-content-secondary)'}
        },
    ]);

    function deletePolicy() {
        const policy = data?.find(x => x.id == selectedId);

        if (!policy) {
            setIsOpen(false)
            return;
        }

        deletePolicyMutation.mutate({
            id: policy.id,
        })

        setIsOpen(false)
    }

    if (isLoading) {
        return (
            <>
                <Heading>Policies</Heading>
                <div className="mx-auto mt-4 flex w-[200px] items-center justify-center gap-y-2 flex-col">
                    <ScaleLoader height={16}/>
                    <div>Loading data...</div>
                </div>
            </>
        )
    }
    if (error) {
        return <div>Error: {error.message}</div>;
    }

    return (
        <>
            <div className="flex items-center justify-between my-4">
                <h1 className="text-rhc-lintblauw-500 text-[30px] leading-9">Beleidsregels</h1>
            </div>
            <Alert open={isOpen} onClose={setIsOpen}>
                <AlertTitle>Are you sure you want to delete policy {selectedId}?</AlertTitle>
                <AlertDescription>
                    This action cannot be undone. This will permanently delete the policy and all of its data.
                </AlertDescription>
                <AlertActions>
                    <Button plain onClick={() => setIsOpen(false)}>
                        Cancel
                    </Button>
                    <Button color={"red"} onClick={deletePolicy}>Delete</Button>
                </AlertActions>
            </Alert>
            <Card className="min-w-3xl flex-1 h-[836px] py-3" disablePadding={true}>
                <div className={"px-10 gap-8"}>
                    <div className={"pt-6 text-right"}>
                        <Button color={"primary"} href={"/policies/add"}>
                            <div className={"flex align-middle justify-center my-auto"}>
                                <IconPlus className={"text-content-inverse-secondary"} size={20}></IconPlus>
                            </div>
                            <span>Aanmaken</span>
                        </Button>
                    </div>
                    <div className={"w-full h-[600px] mt-2"}>
                        <Grid
                            onRowClicked={handleRowClicked}
                            rowData={data}
                            columnDefs={colDefs}
                        />
                    </div>
                </div>
            </Card>
        </>
    )
}
