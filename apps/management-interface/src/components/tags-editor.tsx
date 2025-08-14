import {useId, useMemo, useState} from 'react'
import {Input} from '@/components/input'
import {Button} from '@/components/button'
import {Badge} from '@/components/badge'
import {IconSquareRoundedX} from '@tabler/icons-react'
import {badgeColorKeyFromString} from '@/utilities/color'

export interface TagsEditorProps {
    tags: string[]
    onChange: (tags: string[]) => void
    allTagNames?: string[]
    placeholder?: string
    addButtonLabel?: string
    className?: string
}

/**
 * Reusable tags editor allowing adding/removing string tags with optional suggestions list.
 */
export function TagsEditor({
                               tags,
                               onChange,
                               allTagNames = [],
                               placeholder = 'Type or select a tag',
                               addButtonLabel = 'Add',
                               className,
                           }: TagsEditorProps) {
    const [newTag, setNewTag] = useState('')
    const datalistId = useId()

    const normalizedAllTagNames = useMemo(
        () => (allTagNames ?? []).map((t) => t ?? '').filter(Boolean),
        [allTagNames]
    )

    const handleAddTag = () => {
        const tag = (newTag || '').trim()
        if (!tag) return
        const exists = (tags ?? []).some((t) => t.toLowerCase() === tag.toLowerCase())
        if (exists) {
            setNewTag('')
            return
        }
        onChange([...(tags ?? []), tag])
        setNewTag('')
    }

    const handleRemoveTag = (tagToRemove: string) => {
        onChange((tags ?? []).filter((t) => t !== tagToRemove))
    }

    return (
        <div className={className}>
            <div className="flex gap-2 items-start">
                <div className="basis-1/4">
                    <Input
                        name="newTag"
                        value={newTag}
                        onChange={(e) => setNewTag(e.currentTarget.value)}
                        onKeyDown={(e) => {
                            if (e.key === 'Enter') {
                                e.preventDefault()
                                handleAddTag()
                            }
                        }}
                        list={datalistId}
                        placeholder={placeholder}
                    />
                    <datalist id={datalistId}>
                        {normalizedAllTagNames.map((t) => (
                            <option key={t} value={t}/>
                        ))}
                    </datalist>
                </div>
                <div className={"basis-1/4"}>
                    <Button type="button" color="blue" onClick={handleAddTag}>
                        {addButtonLabel}
                    </Button>
                </div>
                <div className="basis-1/2 flex flex-wrap gap-2">
                    {(tags ?? []).length > 0 ? (
                        (tags ?? []).map((tag) => (
                            <Badge key={tag} color={badgeColorKeyFromString(tag)}>
                                <span className={"text-lg"}>{tag}</span>
                                <button
                                    type="button"
                                    aria-label={`Remove tag ${tag}`}
                                    onClick={(e) => {
                                        e.stopPropagation()
                                        handleRemoveTag(tag)
                                    }}
                                    className="inline-flex items-center rounded hover:opacity-80 focus:outline-hidden focus:ring-2 focus:ring-blue-500"
                                >
                                    <IconSquareRoundedX size={18} />
                                </button>
                            </Badge>
                        ))
                    ) : (
                        <span className="text-sm text-zinc-500">Geen tags</span>
                    )}
                </div>
            </div>
            <div className="mt-2 space-y-2">
            </div>
        </div>
    )
}
