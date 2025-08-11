import {createFileRoute} from '@tanstack/react-router'
import Card from "@/components/card.tsx";
import {Heading} from "@/components/heading.tsx";
import {Table, TableBody, TableCell, TableHead, TableHeader, TableRow} from "@/components/table.tsx";
import { Badge } from '@/components/badge';

const data = [{
    id: '1',
    name: 'organisatie.id',
    description: 'Organisatie ID',
    type: 'String',
    source: 'HR Systeem',
    status: 'Actief',
    usage: 19,
    lastSync: '2022-01-01',
},
    {
        id: '2',
        name: 'feestdagen',
        description: 'Lijst van feestdagen',
        type: 'Array',
        source: 'HR Systeem',
        status: 'Actief',
        usage: 2,
        lastSync: '2022-01-01',
    },
    {
        id: '3',
        name: 'classificatieniveau',
        description: 'Classificatieniveau',
        type: 'string',
        source: 'HR Systeem',
        status: 'Actief',
        usage: 15,
        lastSync: '2022-01-01',
    }]

export const Route = createFileRoute('/attributen/')({
    component: RouteComponent,
})

function RouteComponent() {
    return (<>
        <div className="flex items-center justify-between">
            <h1 className="text-rhc-color-cool-grey-900 font-normal text-[32px] leading-10">Attributen</h1>
        </div>
        <div className="flex flex-col 2xl:flex-row py-3 gap-6">
            <Card className="w-2/3 min-w-3xl flex-1 h-[836px]" disablePadding={true} header={
                <Heading className={"px-4 mt-4"}>Attributen menu</Heading>
            }>
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
                        {data?.map((attribute) => (
                            <TableRow className={"h-20"} key={attribute.id} href={"/attribute/" + attribute.id}
                                      title={`Attribute #${attribute.id}`}>
                                <TableCell><span className={"pl-5"}>{attribute.name}</span></TableCell>
                                <TableCell><Badge color="cyan">{attribute.type}</Badge></TableCell>
                                <TableCell>{attribute.source}</TableCell>
                                <TableCell><Badge color="green">{attribute.status}</Badge></TableCell>
                                <TableCell>{attribute.usage} Beleidsregel(s)</TableCell>
                                <TableCell>{attribute.lastSync}</TableCell>
                                <TableCell>{attribute.usage}</TableCell>
                            </TableRow>
                        ))}
                    </TableBody>
                </Table>
            </Card>
        </div>
    </>)
}
