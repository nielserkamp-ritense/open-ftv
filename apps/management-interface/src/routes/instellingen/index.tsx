import { createFileRoute } from '@tanstack/react-router'
import { Description, Field, FieldGroup, Label } from '@/components/ui/fieldset.tsx'
import { Input } from '@/components/ui/input.tsx'
import { Divider } from '@/components/ui/divider.tsx'
import { Breadcrumb } from '@/components/ui/breadcrumb.tsx'
import { Button } from '@/components/ui/button.tsx'
import { useRef, useState } from 'react'
import { MANAGER_MAX_LOGO_SIZE } from '@/config/env'
import { DEFAULT_HEADER_COLOR, DEFAULT_HEADER_TITLE, DEFAULT_TITLE_COLOR, Settings, useSettings, useUpdateSettings } from '@/services/settings'
import { useCapabilities } from '@/auth/useCapabilities'
import { OIDC_AUTHORITY } from '@/config/env'
import { extractPdpUrls, useBundleConfigurations } from '@/services/bundles'
import { lookupErrorMessage } from '@/utilities/errorMessages'

export const Route = createFileRoute('/instellingen/')({
    component: RouteComponent,
})

const NOT_CONFIGURED = 'Niet geconfigureerd'
const ALLOWED_LOGO_TYPES = ['image/png', 'image/jpeg', 'image/svg+xml', 'image/webp']

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

function SettingsSection({ title, children }: { title: string; children: React.ReactNode }) {
    return (
        <div>
            <h2 className="text-lg font-semibold text-zinc-950 dark:text-white">{title}</h2>
            <Divider className="mt-2 mb-6" soft />
            <FieldGroup>{children}</FieldGroup>
        </div>
    )
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
    const { isAdmin } = useCapabilities()
    const updateMutation = useUpdateSettings()
    const { data: bundleConfigs } = useBundleConfigurations()
    const [errorMessage, setErrorMessage] = useState<string | null>(null)
    const [formData, setFormData] = useState<Settings>(initial)
    const [logoFileSize, setLogoFileSize] = useState<number | null>(null)
    const logoInputRef = useRef<HTMLInputElement>(null)

    const pdpUrls = extractPdpUrls(bundleConfigs)

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

        if (!isAdmin) {
            return
        }

        setErrorMessage(null)

        // Logo size can't be validated by the backend before the whole (base64-encoded) body
        // has already been uploaded, so this check must stay client-side. Everything else is
        // validated server-side, which returns a Code the Dutch message below is looked up by.
        if (logoFileSize !== null && logoFileSize > MANAGER_MAX_LOGO_SIZE) {
            setErrorMessage(`Logo is te groot (maximaal ${formatMaxLogoSize(MANAGER_MAX_LOGO_SIZE)}).`)
            return
        }

        try {
            await updateMutation.mutateAsync(formData)
        } catch (err) {
            setErrorMessage(lookupErrorMessage(err))
        }
    }

    return (
        <form onSubmit={(e) => { void handleSubmit(e) }}>
            <Breadcrumb />
            <div className="flex items-center justify-between my-4">
                <h1 className="text-rhc-lintblauw-500 text-[30px] leading-9">Instellingen</h1>
            </div>
            {errorMessage && (
                <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded mb-4">
                    {errorMessage}
                </div>
            )}
            <div className="space-y-12">
                <SettingsSection title="Huisstijl">
                    <Field>
                        <Label>Titel</Label>
                        <Input name="headerTitle" value={formData.headerTitle} onChange={handleChange} required maxLength={200} className="max-w-[170ch]" disabled={!isAdmin} />
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
                                disabled={!isAdmin}
                                className="h-10 w-14 cursor-pointer rounded-md border border-zinc-950/10 dark:border-white/10 disabled:cursor-not-allowed"
                            />
                            <Input
                                name="headerColor"
                                value={formData.headerColor}
                                onChange={handleChange}
                                maxLength={7}
                                disabled={!isAdmin}
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
                                disabled={!isAdmin}
                                className="h-10 w-14 cursor-pointer rounded-md border border-zinc-950/10 dark:border-white/10 disabled:cursor-not-allowed"
                            />
                            <Input
                                name="titleColor"
                                value={formData.titleColor}
                                onChange={handleChange}
                                maxLength={7}
                                disabled={!isAdmin}
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
                            {isAdmin && (
                                <Button type="button" color="zinc" onClick={() => logoInputRef.current?.click()}>
                                    Logo kiezen
                                </Button>
                            )}
                            {!formData.logo && (
                                <span className="text-sm text-content-secondary">Geen logo gekozen</span>
                            )}
                            {formData.logo && isAdmin && (
                                <Button type="button" color="zinc" onClick={handleRemoveLogo}>Verwijderen</Button>
                            )}
                        </div>
                        <Description>Optioneel. Wordt links in de header getoond, past automatisch binnen de hoogte van de header.</Description>
                    </Field>
                </SettingsSection>

                <SettingsSection title="Toegangsbeheer Infrastructuur">
                    <Field>
                        <Label>IdP URL</Label>
                        <Input value={OIDC_AUTHORITY || NOT_CONFIGURED} readOnly disabled className="max-w-[170ch]" />
                        <Description>Identiteitsprovider waar gebruikers inloggen en hun identiteit wordt geverifieerd.</Description>
                    </Field>
                    <Field>
                        <Label>PDP URL(S)</Label>
                        <div data-slot="control" className="space-y-2">
                            {(pdpUrls.length > 0 ? pdpUrls : [NOT_CONFIGURED]).map((url) => (
                                <Input key={url} value={url} readOnly disabled className="max-w-[170ch]" />
                            ))}
                        </div>
                        <Description>Beleidsbeslissingspunten waar beleidsbundels naar worden gepubliceerd.</Description>
                    </Field>
                </SettingsSection>
            </div>

            <div className="mt-10 flex justify-between">
                <Button type="button" href="/" color={'zinc'}>{isAdmin ? 'Annuleren' : 'Terug'}</Button>
                {isAdmin && (
                    <Button type="submit" color={'emerald'} disabled={updateMutation.isPending}>
                        {updateMutation.isPending ? 'Opslaan...' : 'Opslaan'}
                    </Button>
                )}
            </div>
        </form>
    )
}
