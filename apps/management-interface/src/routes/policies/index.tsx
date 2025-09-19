import { Heading } from '@/components/heading'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/table'
import {useDeletePolicy, usePolicies} from "@/services/policies.ts";
import {ScaleLoader} from "react-spinners";
import {createFileRoute} from "@tanstack/react-router";
import {Button} from "@/components/button.tsx";
import {useState} from "react";
import {Alert, AlertActions, AlertDescription, AlertTitle} from "@/components/alert.tsx";
import Card from "@/components/card.tsx";
import {Badge} from "@/components/badge.tsx";
import {IconBoltFilled} from "@tabler/icons-react";

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

export default function PoliciesComponent() {
    const { data, isLoading, error } = usePolicies();
    const deletePolicyMutation = useDeletePolicy();
    const [isOpen, setIsOpen] = useState(false)
    const [selectedId] = useState<string | null>(null)

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
                      <Button color={"dark"} href={"/policies/add"}>Toevoegen</Button>
                  </div>
                  <Table className="[--gutter:--spacing(6)] lg:[--gutter:--spacing(10)] pt-6 pb-8">
                      <TableHead>
                          <TableRow>
                              <TableHeader className="text-content-secondary font-normal">Titel</TableHeader>
                              <TableHeader className="text-content-secondary font-normal">Doelbinding</TableHeader>
                              <TableHeader className="text-content-secondary font-normal">Actief</TableHeader>
                              <TableHeader className="text-content-secondary text-right font-normal">Concept- en ingeplande versie</TableHeader>
                              <TableHeader className="text-content-secondary text-right font-normal">Laatst bewerkt</TableHeader>
                              <TableHeader className="text-content-secondary text-right font-normal">Beslispunt</TableHeader>
                          </TableRow>
                      </TableHead>
                      <TableBody className={"text-[16px]"}>
                          {data?.map((policy) => (
                              <TableRow key={policy.id} href={"/policies/"+policy.id} title={`Policy #${policy.id}`}>
                                  <TableCell>{policy.metadata.title}</TableCell>
                                  <TableCell className="text-zinc-500">Laadpalen</TableCell>
                                  <TableCell className="py-0 gap-x-0">
                                      <Badge color="ftvgreen" className={"text-[12px] px-0"}><IconBoltFilled className={"h-[16px]"}></IconBoltFilled>v 2.1</Badge>
                                  </TableCell>
                                  <TableCell className="text-right">{policy.metadata?.rvvaId}</TableCell>
                                  <TableCell className="text-right text-zinc-500">{formatDateTime(policy.audit?.updated)}</TableCell>
                                  <TableCell className="text-right text-zinc-500">Zaaksysteem</TableCell>
                              </TableRow>
                          ))}
                      </TableBody>
                  </Table>
              </div>
          </Card>
      </>
  )
}
