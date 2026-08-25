import {createFileRoute, useNavigate} from '@tanstack/react-router'
import {Description, Field, FieldGroup, Fieldset, Label} from "@/components/ui/fieldset.tsx";
import {Input} from "@/components/ui/input.tsx";
import {Select} from "@/components/ui/select.tsx";
import {Textarea} from "@/components/ui/textarea.tsx";
import {Heading} from "@/components/ui/heading.tsx";
import {Button} from "@/components/ui/button.tsx";
import {useState, useMemo} from 'react';
import {useAddPolicy} from '@/services/policies';
import {v7 as uuidv7} from 'uuid';
import { TagsEditor } from "@/components/ui/tags-editor.tsx";
import { useTags } from '@/services/tags.ts';
import { lookupErrorMessage, SOURCE_REQUIRED_MESSAGE } from '@/utilities/errorMessages';

export const Route = createFileRoute('/policies/add')({
    component: AddPolicyComponent,
})

function AddPolicyComponent() {
    const navigate = useNavigate();
    const [formData, setFormData] = useState<{
        policy_name: string;
        title: string;
        language: string;
        data: string;
        rvvaId: string;
        description: string;
        tags: string[];
    }>({
        policy_name: '',
        title: '',
        language: 'cedar',
        data: '',
        rvvaId: '',
        description: '',
        tags: [],
    });
    const [error, setError] = useState<string | null>(null);

    const addPolicyMutation = useAddPolicy();

    const { data: allTagsData } = useTags();
    const allTagNames = useMemo(() => (allTagsData ?? []).map((t) => t.name ?? '').filter(Boolean), [allTagsData]);

    const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>) => {
        const {name, value} = e.target;
        setFormData(prev => ({
            ...prev,
            [name]: value
        }));
    };

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError(null);

        if (!formData.data?.trim()) {
            setError(SOURCE_REQUIRED_MESSAGE);
            return;
        }

        const id = uuidv7()
        try {
            await addPolicyMutation.mutateAsync({
                id: id,
                policy: {
                    id: id,
                    status: "concept",
                    language: formData.language.toLowerCase(),
                    data: formData.data,
                    audit: {
                        created: '',
                        createdBy: {id: ''}
                    },
                    metadata: {
                        rvvaId: formData.rvvaId,
                        description: formData.description,
                        title: formData.title,
                        tags: formData.tags,
                    },
                }
            });

            // Redirect to policies list on success
            await navigate({to: '/policies'});
        } catch (err) {
            setError(lookupErrorMessage(err));
        }
    };

    return (
        <>
            <form onSubmit={e => { void handleSubmit(e); }}>
                <Heading>Toevoegen beleidsregel</Heading>
                {error && (
                    <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded mb-4">
                        {error}
                    </div>
                )}
                <Fieldset>
                    <FieldGroup>
                        <Field>
                            <Label>Titel</Label>
                            <Input
                                name="title"
                                value={formData.title}
                                onChange={handleChange}
                            />
                        </Field>
                        <Field>
                            <Label>Omschrijving</Label>
                            <Input
                                name="description"
                                value={formData.description}
                                onChange={handleChange}
                                required
                            />
                        </Field>
                        <Field>
                            <Label>Regeltaal</Label>
                            <Select
                                name="language"
                                value={formData.language}
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
                                value={formData.rvvaId}
                                onChange={handleChange}
                            />
                        </Field>
                        <Field>
                            <Label>Beslispunten</Label>
                            <Description>Beslispunten waar de beleidsregel gepubliceerd wordt.</Description>
                            <TagsEditor
                                tags={formData.tags ?? []}
                                allTagNames={allTagNames}
                                onChange={(newTags) => setFormData(prev => ({ ...prev, tags: newTags }))}
                                className={"mt-3"}
                            />
                        </Field>
                        <Field>
                            <Label>Broncode</Label>
                            <Textarea
                                name="data"
                                rows={10}
                                value={formData.data}
                                onChange={handleChange}
                                required
                            />
                        </Field>
                    </FieldGroup>
                    <FieldGroup>
                        <Fieldset className={"flex justify-between"}>
                            <Button type="button" href="/policies/" color={"zinc"}>Annuleren</Button>
                            <Button
                                type="submit"
                                color={"emerald"}
                                disabled={addPolicyMutation.isPending}
                            >
                                {addPolicyMutation.isPending ? 'Toevoegen...' : 'Beleidsregel toevoegen'}
                            </Button>
                        </Fieldset>
                    </FieldGroup>
                </Fieldset>
            </form>
        </>
    )
}
