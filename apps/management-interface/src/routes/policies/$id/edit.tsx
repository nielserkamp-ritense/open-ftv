import {createFileRoute, useNavigate} from '@tanstack/react-router'
import {Description, Field, FieldGroup, Fieldset, Label} from "@/components/fieldset.tsx";
import {Input} from "@/components/input.tsx";
import {Select} from "@/components/select.tsx";
import {Textarea} from "@/components/textarea.tsx";
import {Heading} from "@/components/heading.tsx";
import {Button} from "@/components/button.tsx";
import {useEffect, useMemo, useState} from 'react';
import {PolicyResponse, usePolicy, useReplacePolicy} from '@/services/policies.ts';
import {useTags} from '@/services/tags.ts';
import {ScaleLoader} from "react-spinners";
import { TagsEditor } from "@/components/tags-editor.tsx";

export const Route = createFileRoute('/policies/$id/edit')({
    component: EditPolicyComponent,
})

function EditPolicyComponent() {
    const {id} = Route.useParams()
    const {status, data, error} = usePolicy(id)
    const [formData, setFormData] = useState<PolicyResponse>({
        id: '',
        language: '',
        data: '',
        audit: {
            created: '',
            createdBy: ''
        },
        metadata: {
            title: '',
            description: '',
            rvvaId: '',
            url: '',
            tags: [],
        }
    })
    const navigate = useNavigate();
    const replacePolicyMutation = useReplacePolicy();
    const [errorMessage, setErrorMessage] = useState<string | null>(null);

    const { data: allTagsData } = useTags();
    const allTagNames = useMemo(() => (allTagsData ?? []).map((t) => t.name ?? '').filter(Boolean), [allTagsData]);

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
                id: formData.id,
                policy: formData
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
                            <Label>Tags</Label>
                            <TagsEditor
                                tags={formData?.metadata?.tags ?? []}
                                allTagNames={allTagNames}
                                onChange={(newTags) =>
                                    setFormData((c) => ({
                                        ...c,
                                        metadata: {
                                            ...(c.metadata ?? {}),
                                            tags: newTags,
                                        },
                                    }))
                                }
                                className={"mt-3"}
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
