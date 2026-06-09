import {createFileRoute} from '@tanstack/react-router'
import {FTVHeading, Heading} from "@/components/ui/heading.tsx";
import Card from "@/components/ui/card.tsx";
import {PolicyResponse, usePolicy, useReplacePolicy} from "@/services/policies.ts";
import {ScaleLoader} from "react-spinners";
import {DescriptionDetails, DescriptionList, DescriptionTerm} from "@/components/ui/description-list.tsx";
import { Badge } from "@/components/ui/badge.tsx";
import { badgeColorKeyFromString } from "@/utilities/color.ts";
import { Button } from "@/components/ui/button.tsx";
import { Breadcrumb } from "@/components/ui/breadcrumb.tsx";
import { useCapabilities } from "@/auth/useCapabilities.ts";
import { MonacoEditor } from "@/components/ui/monaco-editor.tsx";
import { useState, useEffect, useMemo } from 'react';
import { Field, FieldGroup, Fieldset, Label, Description } from "@/components/ui/fieldset.tsx";
import { Input } from "@/components/ui/input.tsx";
import { Select } from "@/components/ui/select.tsx";
import { TagsEditor } from "@/components/ui/tags-editor.tsx";
import { useTags } from '@/services/tags.ts';

export const Route = createFileRoute('/policies/$id/')({
    component: RouteComponent,
})

function RouteComponent() {
    const {id} = Route.useParams()
    const {status, data, error} = usePolicy(id)
    const {canWrite} = useCapabilities();
    const [isEditMode, setIsEditMode] = useState(false)
    const [formData, setFormData] = useState<PolicyResponse>({
        id: '',
        language: '',
        data: '',
        status: '',
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
    const [errorMessage, setErrorMessage] = useState<string | null>(null)
    const replacePolicyMutation = useReplacePolicy()
    const { data: allTagsData } = useTags()
    const allTagNames = useMemo(() => (allTagsData ?? []).map((t) => t.name ?? '').filter(Boolean), [allTagsData])

    useEffect(() => {
        if (data) {
            setFormData(data as PolicyResponse)
        }
    }, [data])

    const handleEdit = () => {
        setIsEditMode(true)
        setErrorMessage(null)
    }

    const handleCancel = () => {
        setIsEditMode(false)
        setFormData(data!)
        setErrorMessage(null)
    }

    const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>) => {
        const {name, value} = e.target
        const metadataFields = new Set(["title", "description", "rvvaId", "url"])
        if (metadataFields.has(name)) {
            setFormData((c) => ({
                ...c,
                metadata: {
                    ...(c.metadata ?? {}),
                    [name]: value,
                },
            }))
        } else {
            setFormData((c) => ({
                ...c,
                [name]: value,
            }))
        }
    }

    const handleSave = async () => {
        if (!formData?.metadata?.description || !formData?.language || !formData.data) {
            setErrorMessage('Vul alle verplichte velden in')
            return
        }

        try {
            await replacePolicyMutation.mutateAsync({
                id: formData.id,
                policy: formData
            })
            setIsEditMode(false)
            setErrorMessage(null)
        } catch (err) {
            setErrorMessage(err instanceof Error ? err.message : 'Opslaan mislukt')
        }
    }

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

    const tags: string[] = data?.metadata?.tags ?? []

    return <>
        <Breadcrumb items={[
            { label: 'Home', href: '/' },
            { label: 'Beleidsregels', href: '/policies' },
            { label: data?.metadata?.title || 'Details' }
        ]} />
        <div className="flex items-center justify-between">
            <h1 className="text-rhc-color-cool-grey-900 font-normal text-[32px] leading-10">Beleidsregel</h1>
        </div>
        {errorMessage && (
            <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded mb-4">
                {errorMessage}
            </div>
        )}
        <div className="flex items-center justify-end gap-2 py-2">
            {!isEditMode ? (
                canWrite ? <Button color={"primary"} onClick={handleEdit}>Bewerken</Button> : null
            ) : (
                <>
                    <Button color={"zinc"} onClick={handleCancel}>Annuleren</Button>
                    <Button
                        color={"emerald"}
                        onClick={() => { void handleSave(); }}
                        disabled={replacePolicyMutation.isPending}
                    >
                        {replacePolicyMutation.isPending ? 'Bezig met opslaan...' : 'Opslaan'}
                    </Button>
                </>
            )}
        </div>
        <div className="flex flex-col 2xl:flex-row py-3 gap-6">
            <Card className="w-2/3 min-w-3xl flex-1 h-[836px]" disablePadding={true}>
                <div className="flex flex-col h-full px-4 py-5 sm:p-6">
                    <FTVHeading level={2}>{data?.metadata?.title}</FTVHeading>
                    <p className={"text-rhc-color-cool-grey-900 py-2"}>De broncode van de beleidsregel.</p>
                    <div className="flex-1 mt-3 h-[836px] ">
                        <MonacoEditor
                            value={isEditMode ? formData?.data : data?.data}
                            language={isEditMode ? formData?.language : data?.language}
                            readOnly={!isEditMode}
                            height="600px"
                            className={"border-1 border-rhc-color-cool-grey-200 rounded-md overflow-hidden"}
                            onChange={(value) => {
                                if (isEditMode) {
                                    setFormData((c) => ({
                                        ...c,
                                        data: value ?? '',
                                    }))
                                }
                            }}
                        />
                    </div>
                </div>
            </Card>
            <div className="w-1/3 min-w-3xl 2xl:min-w-lg flex flex-col gap-6">
                <Card className="flex-1">
                    {!isEditMode ? (
                        <DescriptionList>
                            <DescriptionTerm>Titel</DescriptionTerm>
                            <DescriptionDetails>{data?.metadata?.title}</DescriptionDetails>
                            <DescriptionTerm>Omschrijving</DescriptionTerm>
                            <DescriptionDetails>{data?.metadata?.description}</DescriptionDetails>
                            <DescriptionTerm>Regeltaal</DescriptionTerm>
                            <DescriptionDetails>{data?.language?.charAt(0).toUpperCase() + data?.language?.slice(1).toLowerCase()}</DescriptionDetails>
                            <DescriptionTerm>Register van verwerking ID</DescriptionTerm>
                            <DescriptionDetails>{data?.metadata?.rvvaId}</DescriptionDetails>
                            <DescriptionTerm>URL</DescriptionTerm>
                            <DescriptionDetails>{data?.metadata?.url}</DescriptionDetails>
                            <DescriptionTerm>Beslispunten</DescriptionTerm>
                            <DescriptionDetails>
                                <div className="flex flex-wrap gap-2">
                                    {tags.length > 0 ? (
                                        tags.map((tag) => (
                                            <Badge key={tag} color={badgeColorKeyFromString(tag)}>{tag}</Badge>
                                        ))
                                    ) : (
                                        <span>Geen tags</span>
                                    )}
                                </div>
                            </DescriptionDetails>
                            <DescriptionTerm>Gemaakt door</DescriptionTerm>
                            <DescriptionDetails>{data?.audit?.createdBy}</DescriptionDetails>
                            <DescriptionTerm>Gemaakt op</DescriptionTerm>
                            <DescriptionDetails>{data?.audit?.created}</DescriptionDetails>
                            <DescriptionTerm>Laatst bijgewerkt op</DescriptionTerm>
                            <DescriptionDetails>{data?.audit?.updated}</DescriptionDetails>
                        </DescriptionList>
                    ) : (
                        <Fieldset>
                            <FieldGroup>
                                <Field>
                                    <Label>Titel</Label>
                                    <Input
                                        name="title"
                                        value={formData?.metadata?.title ?? ''}
                                        onChange={handleChange}
                                    />
                                </Field>
                                <Field>
                                    <Label>Omschrijving</Label>
                                    <Input
                                        name="description"
                                        value={formData?.metadata?.description ?? ''}
                                        onChange={handleChange}
                                        required
                                    />
                                </Field>
                                <Field>
                                    <Label>Regeltaal</Label>
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
                                    <Description>De regeltaal van de beleidsregel.</Description>
                                </Field>
                                <Field>
                                    <Label>Register van verwerkingsactiviteiten ID</Label>
                                    <Input
                                        name="rvvaId"
                                        value={formData?.metadata?.rvvaId ?? ''}
                                        onChange={handleChange}
                                    />
                                </Field>
                                <Field>
                                    <Label>URL</Label>
                                    <Input
                                        name="url"
                                        value={formData?.metadata?.url ?? ''}
                                        onChange={handleChange}
                                    />
                                </Field>
                                <Field>
                                    <Label>Beslispunten</Label>
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
                            </FieldGroup>
                            <FieldGroup>
                                <DescriptionList>
                                    <DescriptionTerm>Gemaakt door</DescriptionTerm>
                                    <DescriptionDetails>{data?.audit?.createdBy}</DescriptionDetails>
                                    <DescriptionTerm>Gemaakt op</DescriptionTerm>
                                    <DescriptionDetails>{data?.audit?.created}</DescriptionDetails>
                                    <DescriptionTerm>Laatst bijgewerkt op</DescriptionTerm>
                                    <DescriptionDetails>{data?.audit?.updated}</DescriptionDetails>
                                </DescriptionList>
                            </FieldGroup>
                        </Fieldset>
                    )}
                </Card>
                <Card className="flex-1" header={
                    <Heading>Versies</Heading>
                }>
                </Card>
            </div>
        </div>
    </>
}
