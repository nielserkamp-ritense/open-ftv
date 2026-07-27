import { createFileRoute } from '@tanstack/react-router'
import { Description, Field, FieldGroup, Fieldset, Label } from '@/components/ui/fieldset.tsx'
import { Input } from '@/components/ui/input.tsx'
import { Heading, Subheading } from '@/components/ui/heading.tsx'
import { Button } from '@/components/ui/button.tsx'
import { useRef, useState } from 'react'
import { MANAGER_MAX_LOGO_SIZE } from '@/config/env'
import { DEFAULT_HEADER_COLOR, DEFAULT_HEADER_TITLE, DEFAULT_TITLE_COLOR, Settings, useSettings, useUpdateSettings } from '@/services/settings'

export const Route = createFileRoute('/instellingen/')({
    component: RouteComponent,
})

const ALLOWED_LOGO_TYPES = ['image/png', 'image/jpeg', 'image/svg+xml', 'image/webp']
const HEX_COLOR_PATTERN = /^#([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$/

function isValidHexColor(value: string): boolean {
    return HEX_COLOR_PATTERN.test(value)
}

// the native color input only accepts a full 6-digit hex value, so fall back
// to black while the user is still typing a partial/invalid code by hand.
function toColorInputValue(value: string): string {
    return /^#([0-9a-fA-F]{6})$/.test(value) ? value : '#000000'
}

function fileToBase64(file: File): Promise<string> {
    return new Promise((resolve, reject) => {
        const reader = new FileReader()
        reader.onload = () => {
            const result = reader.result as string
            // strip the "data:<mime>;base64," prefix
            resolve(result.substring(result.indexOf(',') + 1))
        }
        reader.onerror = () => reject(reader.error instanceof Error ? reader.error : new Error('failed to read file'))
        reader.readAsDataURL(file)
    })
}

function formatMaxLogoSize(bytes: number): string {
    if (bytes % 1024 === 0) {
        return `${bytes / 1024} KiB`
    }
    return `${bytes} bytes`
}

function RouteComponent() {
    const { data, isPending } = useSettings()

    if (isPending) {
        return <div className="p-4 text-sm text-content-secondary">Laden…</div>
    }

    return <SettingsForm initial={data ?? { headerTitle: DEFAULT_HEADER_TITLE, headerColor: DEFAULT_HEADER_COLOR, titleColor: DEFAULT_TITLE_COLOR }} />
}

function SettingsForm({ initial }: { initial: Settings }) {
    const updateMutation = useUpdateSettings()
    const [errorMessage, setErrorMessage] = useState<string | null>(null)
    const [formData, setFormData] = useState<Settings>(initial)
    const [logoFileSize, setLogoFileSize] = useState<number | null>(null)
    const logoInputRef = useRef<HTMLInputElement>(null)

    const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        const { name, value } = e.target
        setFormData((c) => ({ ...c, [name]: value }))
    }

    const handleLogoChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
        const file = e.target.files?.[0]
        if (!file) {
            return
        }

        setErrorMessage(null)

        if (!ALLOWED_LOGO_TYPES.includes(file.type)) {
            setErrorMessage('Logo moet een PNG, JPEG, SVG of WebP-afbeelding zijn.')
            if (logoInputRef.current) {
                logoInputRef.current.value = ''
            }
            return
        }

        try {
            const base64 = await fileToBase64(file)
            setLogoFileSize(file.size)
            setFormData((c) => ({ ...c, logo: base64, logoMediaType: file.type as Settings['logoMediaType'] }))
        } catch {
            setErrorMessage('Logo kon niet worden gelezen.')
            setLogoFileSize(null)
        }
    }

    const handleRemoveLogo = () => {
        setFormData((c) => ({ ...c, logo: undefined, logoMediaType: undefined }))
        setLogoFileSize(null)
        if (logoInputRef.current) {
            logoInputRef.current.value = ''
        }
    }

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault()
        setErrorMessage(null)

        if (!formData.headerTitle) {
            setErrorMessage('Titel is verplicht')
            return
        }
        if (!formData.headerColor) {
            setErrorMessage('Kleur is verplicht')
            return
        }
        if (!isValidHexColor(formData.headerColor)) {
            setErrorMessage('Kleur van de header moet een geldige kleurcode zijn, bijv. #364f78')
            return
        }
        if (!formData.titleColor) {
            setErrorMessage('Kleur van de titel is verplicht')
            return
        }
        if (!isValidHexColor(formData.titleColor)) {
            setErrorMessage('Kleur van de titel moet een geldige kleurcode zijn, bijv. #fcfcfc')
            return
        }
        if (logoFileSize !== null && logoFileSize > MANAGER_MAX_LOGO_SIZE) {
            setErrorMessage(`Logo is te groot (maximaal ${formatMaxLogoSize(MANAGER_MAX_LOGO_SIZE)}).`)
            return
        }

        try {
            await updateMutation.mutateAsync(formData)
        } catch (err) {
            setErrorMessage(err instanceof Error ? err.message : 'Instellingen konden niet worden opgeslagen')
        }
    }

    return (
        <form onSubmit={(e) => { void handleSubmit(e) }}>
            <Heading>Instellingen</Heading>
            {errorMessage && (
                <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded mb-4">
                    {errorMessage}
                </div>
            )}
            <Fieldset>
                <Subheading>Huisstijl</Subheading>
                <FieldGroup>
                    <Field>
                        <Label>Titel</Label>
                        <Input name="headerTitle" value={formData.headerTitle} onChange={handleChange} required maxLength={200} className="max-w-[170ch]" />
                        <Description>Titel die in de header wordt getoond.</Description>
                    </Field>
                    <Field>
                        <Label>Kleur van de header</Label>
                        <div className="flex items-center gap-3">
                            <input
                                type="color"
                                name="headerColor"
                                value={toColorInputValue(formData.headerColor)}
                                onChange={handleChange}
                                className="h-10 w-14 cursor-pointer rounded-md border border-zinc-950/10 dark:border-white/10"
                            />
                            <Input
                                name="headerColor"
                                value={formData.headerColor}
                                onChange={handleChange}
                                placeholder="#364f78"
                                maxLength={7}
                                className="max-w-28"
                            />
                        </div>
                        <Description>Achtergrondkleur van de header. Kies een kleur of vul een kleurcode in, bijv. #364f78.</Description>
                    </Field>
                    <Field>
                        <Label>Kleur van de titel</Label>
                        <div className="flex items-center gap-3">
                            <input
                                type="color"
                                name="titleColor"
                                value={toColorInputValue(formData.titleColor)}
                                onChange={handleChange}
                                className="h-10 w-14 cursor-pointer rounded-md border border-zinc-950/10 dark:border-white/10"
                            />
                            <Input
                                name="titleColor"
                                value={formData.titleColor}
                                onChange={handleChange}
                                placeholder="#364f78"
                                maxLength={7}
                                className="max-w-28"
                            />
                        </div>
                        <Description>Tekstkleur van de titel in de header. Kies een kleur of vul een kleurcode in, bijv. #fcfcfc.</Description>
                    </Field>
                    <Field>
                        <Label>Logo</Label>
                        <div className="flex items-center gap-4">
                            {formData.logo && (
                                <img
                                    src={`data:${formData.logoMediaType};base64,${formData.logo}`}
                                    alt="Logo preview"
                                    className="h-16 w-auto max-w-[200px] object-contain rounded border border-zinc-950/10 dark:border-white/10 bg-white p-2"
                                />
                            )}
                            <input
                                ref={logoInputRef}
                                type="file"
                                accept={ALLOWED_LOGO_TYPES.join(',')}
                                onChange={(e) => { void handleLogoChange(e) }}
                                className="sr-only"
                            />
                            <Button type="button" color="zinc" onClick={() => logoInputRef.current?.click()}>
                                Logo kiezen
                            </Button>
                            {!formData.logo && (
                                <span className="text-sm text-content-secondary">Geen logo gekozen</span>
                            )}
                            {formData.logo && (
                                <Button type="button" color="zinc" onClick={handleRemoveLogo}>Verwijderen</Button>
                            )}
                        </div>
                        <Description>Optioneel. Wordt links in de header getoond, past automatisch binnen de hoogte van de header.</Description>
                    </Field>
                </FieldGroup>
                <FieldGroup>
                    <Fieldset className={'flex justify-between'}>
                        <Button type="button" href="/" color={'zinc'}>Annuleren</Button>
                        <Button type="submit" color={'emerald'} disabled={updateMutation.isPending}>
                            {updateMutation.isPending ? 'Opslaan...' : 'Opslaan'}
                        </Button>
                    </Fieldset>
                </FieldGroup>
            </Fieldset>
        </form>
    )
}
