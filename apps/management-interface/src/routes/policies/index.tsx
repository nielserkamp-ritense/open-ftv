import { Heading } from '@/components/heading'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/table'
import {useDeletePolicy, usePolicies} from "@/services/policies.ts";
import {ScaleLoader} from "react-spinners";
import {createFileRoute} from "@tanstack/react-router";
import {Button} from "@/components/button.tsx";
import {Dropdown, DropdownButton, DropdownItem, DropdownMenu} from '@/components/dropdown';
import {EllipsisHorizontalIcon} from "@heroicons/react/16/solid";
import {useState} from "react";
import {Alert, AlertActions, AlertDescription, AlertTitle} from "@/components/alert.tsx";

export const Route = createFileRoute('/policies/')({
  component: PoliciesComponent,
})

export default function PoliciesComponent() {
    const { data, isLoading, error } = usePolicies();
    const deletePolicyMutation = useDeletePolicy();
    const [isOpen, setIsOpen] = useState(false)
    const [selectedId, setSelectedId] = useState<string | null>(null)

    function showDeleteModal(id: string) {
        setSelectedId(id)
        setIsOpen(true)
    }

    function deletePolicy() {
        const policy = data?.find(x => x.id == selectedId);

        if (!policy) {
            setIsOpen(false)
            return;
        }

        deletePolicyMutation.mutate({
            language: policy.language,
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
          <div className="flex items-center justify-between">
              <h1 className="text-rhc-color-cool-grey-900 font-normal text-[32px] leading-10">Beleidsoverzicht</h1>
              <Button color={"emerald"} href={"/policies/add"}>Toevoegen</Button>
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
          <Table className="mt-4 [--gutter:--spacing(6)] lg:[--gutter:--spacing(10)]">
              <TableHead>
                  <TableRow>
                      <TableHeader>ID</TableHeader>
                      <TableHeader>Taal</TableHeader>
                      <TableHeader>URL</TableHeader>
                      <TableHeader className="text-right">rvvaId</TableHeader>
                      <TableHeader className="text-right">Acties</TableHeader>
                  </TableRow>
              </TableHead>
              <TableBody>
                  {data?.map((policy) => (
                      <TableRow key={policy.id} href={"/policies/"+policy.language+"/"+policy.id} title={`Policy #${policy.id}`}>
                          <TableCell>{policy.id}</TableCell>
                          <TableCell className="text-zinc-500">{policy.language}</TableCell>
                          <TableCell>{policy.metadata?.url}</TableCell>
                          <TableCell className="text-right">{policy.metadata?.rvvaId}</TableCell>
                          <TableCell>
                              <div className="pr-2 -mx-3 -my-1.5 sm:-mx-2.5 text-right">
                                  <Dropdown>
                                      <DropdownButton plain aria-label="More options">
                                          <EllipsisHorizontalIcon />
                                      </DropdownButton>
                                      <DropdownMenu anchor="bottom end">
                                          <DropdownItem href={`/policies/${policy.language}/${policy.id}`}>View</DropdownItem>
                                          <DropdownItem href={`/policies/${policy.language}/${policy.id}/edit`}>Edit</DropdownItem>
                                          <DropdownItem onClick={() => showDeleteModal(policy.id)}>Delete</DropdownItem>
                                      </DropdownMenu>
                                  </Dropdown>
                              </div>
                          </TableCell>
                      </TableRow>
                  ))}
              </TableBody>
          </Table>
      </>
  )
}
