import { useState, useEffect } from 'react'
import { TextInput } from '@mantine/core'
import { useDebouncedValue } from '@mantine/hooks'
import { AprilIconSearch } from '@ukituki-ps/april-ui'

interface SearchInputProps {
	value: string
	onChange: (value: string) => void
}

/**
 * Поле поиска с debounce (300ms).
 *
 * Стили по прототипу:
 * - border: 1.5px solid var(--brand-border)
 * - border-radius: 8px
 * - padding: 8px 14px
 * - Иконка search слева (AprilIconSearch)
 * - font-size: 13px
 *
 * Используется в каталоге для поиска льгот по названию и описанию.
 */
export function SearchInput({ value, onChange }: SearchInputProps) {
	const [inputValue, setInputValue] = useState(value)
	const [debouncedValue] = useDebouncedValue(inputValue, 300)

	// Когда debounced значение изменилось — вызываем onChange
	useEffect(() => {
		onChange(debouncedValue)
	}, [debouncedValue, onChange])

	// Синхронизация с внешним value (если изменён извне, напр. сброс фильтров)
	useEffect(() => {
		setInputValue(value)
	}, [value])

	return (
		<TextInput
			value={inputValue}
			onChange={(e) => setInputValue(e.currentTarget.value)}
			placeholder="Поиск льгот..."
			leftSectionWidth={42}
			leftSection={
				<AprilIconSearch
					size={16}
					style={{ color: 'var(--brand-text-muted)', flexShrink: 0 }}
				/>
			}
			style={{
				fontSize: '13px',
			}}
			styles={{
				input: {
					border: '1.5px solid var(--brand-border)',
					borderRadius: '8px',
					padding: '8px 14px',
				},
			}}
			mb="md"
		/>
	)
}
