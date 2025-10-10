import {createFileRoute} from '@tanstack/react-router'
import Card from "@/components/card.tsx";
import {Heading} from "@/components/heading.tsx";
import {Table, TableBody, TableCell, TableHead, TableHeader, TableRow} from "@/components/table.tsx";
import { Badge } from '@/components/badge';
import { Pagination, PaginationList, PaginationNext, PaginationPage, PaginationPrevious } from '@/components/pagination';
import { useAttributes } from '@/services/attributes';
import { ScaleLoader } from 'react-spinners';
import { Button } from '@/components/button.tsx';

export const Route = createFileRoute('/attributen/')({
    component: RouteComponent,
})

function RouteComponent() {
    const { status, data, error } = useAttributes();

    if (status === 'pending') {
        return (
            <>
                <div className="flex items-center justify-between">
                    <Heading className="lg:text-3xl ">Attributen</Heading>
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

    const attributes = data ?? [];

    return (<>
        <div className="flex items-center justify-between">
            <h1 className="text-rhc-color-cool-grey-900 font-bold text-[32px] leading-10">Attributen en bronnen</h1>
            <Button href="/attributen/toevoegen" color={"blue"}>Nieuw attribuut</Button>
        </div>
        <div className="flex flex-col 2xl:flex-row py-3 gap-6">
            <Card className="w-2/3 min-w-3xl flex-1 h-[836px]" disablePadding={true} header={
                <Heading className={"px-4 mt-4"}>Attributen menu</Heading>
            }>
                <div className="flex h-full flex-col">
                    <Table>
                        <TableHead className="bg-background-secondary h-12">
                            <TableRow className={""}>
                                <TableHeader><span className={"pl-5"}>Attribuut</span></TableHeader>
                                <TableHeader>Type</TableHeader>
                                <TableHeader>Bron</TableHeader>
                                <TableHeader>Status</TableHeader>
                                <TableHeader>Gebruik</TableHeader>
                                <TableHeader>Laatste Sync</TableHeader>
                                <TableHeader>Acties</TableHeader>
                            </TableRow>
                        </TableHead>
                        <TableBody>
                            {attributes.map((attribute) => {
                                const key = attribute.key;
                                const type = attribute.type ?? '-';
                                const source = attribute.metadata?.tags ?? '-';
                                const usageCount = attribute.usageData?.reduce((sum, x) => sum + (x.bundle?.length ?? 0), 0);
                                const lastSync = attribute.audit?.updated ?? attribute.audit?.created ?? '-';
                                return (
                                    <TableRow className={"h-20"} key={key} title={`Attribute ${key}`}>
                                        <TableCell><span className={"pl-5"}>{key}</span></TableCell>
                                        <TableCell>{type !== '-' ? <Badge color="cyan">{type}</Badge> : '-'}</TableCell>
                                        <TableCell>{source}</TableCell>
                                        <TableCell>-</TableCell>
                                        <TableCell>{usageCount} Beleidsregel(s)</TableCell>
                                        <TableCell>{lastSync}</TableCell>
                                        <TableCell>-</TableCell>
                                    </TableRow>
                                );
                            })}
                        </TableBody>
                    </Table>
                    <div className="mt-auto pt-4 pb-4 flex justify-center">
                        <Pagination>
                            <PaginationPrevious href={null} />
                            <PaginationList>
                                <PaginationPage href="/attributen" current>1</PaginationPage>
                            </PaginationList>
                            <PaginationNext href={null} />
                        </Pagination>
                    </div>
                </div>
            </Card>
        </div>
    </>)
}
