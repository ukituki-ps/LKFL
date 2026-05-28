import { useState, useEffect } from 'react'
import { TextInput } from '@mantine/core'
import { useDebouncedValue } from '@mantine/hooks'

interface SearchInputProps {
	value: string
	onChange: (value: string) => void
}

/**
 * Поле поиска с debounce (300ms).
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
			size="md"
			radius="md"
			mb="md"
		/>
	)
}
