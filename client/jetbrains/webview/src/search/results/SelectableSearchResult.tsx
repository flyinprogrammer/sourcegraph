import React, { useCallback } from 'react'

import { SearchMatch } from '@sourcegraph/shared/src/search/stream'

import { getResultId, LineMatchItem, SymbolMatchItem } from './utils'

import styles from './SelectableSearchResult.module.scss'

interface Props {
    children: (isActive: boolean) => React.ReactNode
    lineOrSymbolMatch?: LineMatchItem | SymbolMatchItem
    match: SearchMatch
    selectedResult: null | string
    selectResult: (id: string) => void
    openResult: (id: string) => void
}

export const SelectableSearchResult: React.FunctionComponent<Props> = ({
    children,
    lineOrSymbolMatch,
    match,
    selectedResult,
    selectResult,
    openResult,
}: Props) => {
    const resultId = getResultId(match, lineOrSymbolMatch)
    const onClick = useCallback((): void => selectResult(resultId), [selectResult, resultId])
    const onDoubleClick = useCallback((): void => openResult(resultId), [openResult, resultId])
    const isActive = resultId === selectedResult

    return (
        // The below element's accessibility is handled via a document level event listener.
        //
        // eslint-disable-next-line jsx-a11y/click-events-have-key-events,jsx-a11y/no-static-element-interactions
        <div
            id={`search-result-list-item-${resultId}`}
            className={styles.selectableSearchResult}
            onClick={onClick}
            onDoubleClick={onDoubleClick}
            key={resultId}
        >
            {children(isActive)}
        </div>
    )
}
