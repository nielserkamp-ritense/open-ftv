import {createFileRoute, useNavigate} from '@tanstack/react-router'
import {Description, Field, FieldGroup, Fieldset, Label} from "@/components/fieldset.tsx";
import {Input} from "@/components/input.tsx";
import {Select} from "@/components/select.tsx";
import {Textarea} from "@/components/textarea.tsx";
import {Heading} from "@/components/heading.tsx";
import {Button} from "@/components/button.tsx";
import {useEffect, useState} from 'react';
import {PolicyResponse, usePolicy, useReplacePolicy} from '@/services/policies';
import {ScaleLoader} from "react-spinners";

export const Route = createFileRoute('/policies/$language/$policyId/edit')({
    component: EditPolicyComponent,
})

function EditPolicyComponent() {
    const {language, policyId} = Route.useParams()
    const {status, data, error} = usePolicy(language, policyId)
    const [formData, setFormData] = useState<PolicyResponse>({
        id: '',
        language: '',
        data: '',
        metadata: {
            title: '',
            description: '',
            rvvaId: '',
            url: '',
        }
    })
    const navigate = useNavigate();
    const replacePolicyMutation = useReplacePolicy();
    const [errorMessage, setErrorMessage] = useState<string | null>(null);

    useEffect(() => {
        // @ts-expect-error error
        setFormData(data)
    }, [data]);

    if (status == "pending") {
        return <>
            <div className="flex items-center justify-between">
                <Heading className="lg:text-3xl ">Policy</Heading>
            </div>
            <div className="flex flex-col 2xl:flex-row py-3 gap-6">
                <div className="mx-auto mt-4 flex w-[200px] items-center justify-center gap-y-2 flex-col">
                    <ScaleLoader height={16}/>
                    <div>Loading data...</div>
                </div>
            </div>
        </>
    }

    if (status == "error") {
        return <>
            <span>Error: {error.message}</span>
        </>
    }

    const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>) => {
        const {name, value} = e.target;
        const metadataFields = new Set(["title", "description", "rvvaId", "url"]);
        if (metadataFields.has(name)) {
            setFormData((c) => ({
                ...c,
                metadata: {
                    ...(c.metadata ?? {}),
                    [name]: value,
                },
            }));
        } else {
            setFormData((c) => ({
                ...c,
                [name]: value,
            }));
        }
    };

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();

        if (!formData?.metadata?.description || !formData?.language || !formData.data) {
            return;
        }

        try {
            await replacePolicyMutation.mutateAsync({
                language: formData.language.toLowerCase(),
                id: formData.id,
                policy: {
                    id: formData.id,
                    language: formData.language.toLowerCase(),
                    data: formData.data,
                    metadata: formData.metadata,
                }
            });

            // Redirect to policies list on success
            await navigate({to: '/policies'});
        } catch (err) {
            setErrorMessage(err instanceof Error ? err.message : 'Failed to add policy');
        }
    };

    return (
        <>
            <form onSubmit={(e) => { void handleSubmit(e); }}>
                <Heading>Edit policy</Heading>
                {error && (
                    <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded mb-4">
                        {errorMessage}
                    </div>
                )}
                <Fieldset>
                    <FieldGroup>
                        <Field>
                            <Label>Title</Label>
                            <Input
                                name="title"
                                value={formData?.metadata?.title ?? ''}
                                onChange={handleChange}
                            />
                        </Field>
                        <Field>
                            <Label>Policy description</Label>
                            <Input
                                name="description"
                                value={formData?.metadata?.description ?? ''}
                                onChange={handleChange}
                                required
                            />
                        </Field>
                        <Field>
                            <Label>Language</Label>
                            <Select
                                name="language"
                                value={formData?.language}
                                onChange={handleChange}
                            >
                                <option value="cedar">Cedar</option>
                                <option value="rego">Rego</option>
                                <option value="cerbos">Cerbos</option>
                                <option value="openfga">OpenFGA</option>
                            </Select>
                            <Description>The language of the policy that is added.</Description>
                        </Field>
                        <Field>
                            <Label>rvva ID</Label>
                            <Input
                                name="rvvaId"
                                value={formData?.metadata?.rvvaId ?? ''}
                                onChange={handleChange}
                            />
                        </Field>
                        <Field>
                            <Label>Policy code</Label>
                            <Textarea
                                name="data"
                                rows={10}
                                value={formData?.data}
                                onChange={handleChange}
                                required
                            />
                        </Field>
                    </FieldGroup>
                    <FieldGroup>
                        <Fieldset className={"flex justify-between"}>
                            <Button type="button" href="/policies/" color={"zinc"}>Discard</Button>
                            <Button
                                type="submit"
                                color={"emerald"}
                                disabled={replacePolicyMutation.isPending}
                            >
                                {replacePolicyMutation.isPending ? 'Saving...' : 'Save'}
                            </Button>
                        </Fieldset>
                    </FieldGroup>
                </Fieldset>
            </form>
        </>
    )
}
