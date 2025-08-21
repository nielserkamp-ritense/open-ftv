import { createFileRoute, useNavigate } from '@tanstack/react-router'
import { Description, Field, FieldGroup, Fieldset, Label } from '@/components/fieldset.tsx'
import { Input } from '@/components/input.tsx'
import { Select } from '@/components/select.tsx'
import { Textarea } from '@/components/textarea.tsx'
import { Heading } from '@/components/heading.tsx'
import { Button } from '@/components/button.tsx'
import { useState } from 'react'
import { Attribute, useAddAttribute } from '@/services/attributes'

export const Route = createFileRoute('/attributen/toevoegen')({
    component: NewAttributeComponent,
})

function NewAttributeComponent() {
    const navigate = useNavigate()
    const addMutation = useAddAttribute()
    const [errorMessage, setErrorMessage] = useState<string | null>(null)
    const [formData, setFormData] = useState<Attribute>({
        key: '',
        value: '',
        type: 'string',
        audit: {
            created: '',
            createdBy: ''
        },
        metadata: {
            title: '',
            description: '',
        },
    })

    const handleChange = (
        e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>,
    ) => {
        const { name, value } = e.target

        const metadataFields = new Set(['title', 'description'])
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

    const coerceValue = (val: string, type?: string): unknown => {
        switch (type) {
            case 'integer':
                return Number.isNaN(parseInt(val, 10)) ? val : parseInt(val, 10)
            case 'float':
                return Number.isNaN(parseFloat(val)) ? val : parseFloat(val)
            case 'boolean':
                return /^(true|1)$/i.test(val)
            // For date/time/timestamp we send as string as backend validates RFC3339
            case 'date':
            case 'time':
            case 'timestamp':
            case 'string':
            default:
                return val
        }
    }

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault()

        setErrorMessage(null)

        if (!formData.key) {
            setErrorMessage('Key is required')
            return
        }

        try {
            const payload: Attribute = {
                ...formData,
                value: coerceValue(String(formData.value ?? ''), formData.type),
            }

            await addMutation.mutateAsync({
                key: formData.key,
                attribute: payload,
            })

            await navigate({ to: '/attributen' })
        } catch (err) {
            setErrorMessage(err instanceof Error ? err.message : 'Failed to add attribute')
        }
    }

    return (
        <>
            <form onSubmit={(e) => { void handleSubmit(e) }}>
                <Heading>Nieuw attribuut</Heading>
                {errorMessage && (
                    <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded mb-4">
                        {errorMessage}
                    </div>
                )}
                <Fieldset>
                    <FieldGroup>
                        <Field>
                            <Label>Key</Label>
                            <Input name="key" value={formData.key} onChange={handleChange} required />
                            <Description>Unieke sleutel van het attribuut.</Description>
                        </Field>
                        <Field>
                            <Label>Type</Label>
                            <Select name="type" value={formData.type ?? ''} onChange={handleChange}>
                                <option value="string">string</option>
                                <option value="integer">integer</option>
                                <option value="float">float</option>
                                <option value="boolean">boolean</option>
                                <option value="date">date</option>
                                <option value="time">time</option>
                                <option value="timestamp">timestamp</option>
                            </Select>
                            <Description>Het type bepaalt hoe de waarde geïnterpreteerd wordt.</Description>
                        </Field>
                        <Field>
                            <Label>Value</Label>
                            <Input name="value" value={String(formData.value ?? '')} onChange={handleChange} />
                        </Field>
                        <Field>
                            <Label>Titel</Label>
                            <Input name="title" value={formData.metadata?.title ?? ''} onChange={handleChange} />
                        </Field>
                        <Field>
                            <Label>Beschrijving</Label>
                            <Textarea name="description" rows={4} value={formData.metadata?.description ?? ''} onChange={handleChange} />
                        </Field>
                    </FieldGroup>
                    <FieldGroup>
                        <Fieldset className={"flex justify-between"}>
                            <Button type="button" href="/attributen" color={"zinc"}>Discard</Button>
                            <Button type="submit" color={"emerald"} disabled={addMutation.isPending}>
                                {addMutation.isPending ? 'Saving...' : 'Save'}
                            </Button>
                        </Fieldset>
                    </FieldGroup>
                </Fieldset>
            </form>
        </>
    )
}
