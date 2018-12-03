import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import * as React from 'react'
import { Observable, of, Subject, Subscription } from 'rxjs'
import { catchError, distinctUntilChanged, endWith, map, startWith, switchMap } from 'rxjs/operators'
import { Location } from '../../api/protocol/plainTypes'
import { FetchFileCtx } from '../../components/CodeExcerpt'
import { RepositoryIcon } from '../../components/icons' // TODO: Switch to mdi icon
import { RepoLink } from '../../components/RepoLink'
import { Resizable } from '../../components/Resizable'
import { ErrorLike, isErrorLike } from '../../util/errors'
import { asError } from '../../util/errors'
import { parseRepoURI } from '../../util/url'
import { FileLocations, FileLocationsError, FileLocationsNotFound } from './FileLocations'

interface Props {
    /**
     * The observable that emits the locations.
     */
    locations: Observable<Location[] | null>

    /** The icon to use for each location. */
    icon: React.ComponentType<{ className?: string }>

    /** Called when an item in the tree is selected. */
    onSelectTree?: () => void

    /** Called when a location is selected. */
    onSelectLocation?: () => void

    className?: string

    isLightTheme: boolean

    fetchHighlightedFileLines: (ctx: FetchFileCtx, force?: boolean) => Observable<string[]>
}

const LOADING: 'loading' = 'loading'

interface State {
    /**
     * Locations (inside files identified by LSP-style git:// URIs) to display, loading, or an error if they failed
     * to load.
     */
    locationsOrError: typeof LOADING | Location[] | ErrorLike

    locationsComplete: boolean

    selectedGroups?: string[]
}

/**
 * Displays a multi-column view to drill down (by repository, file, etc.) to a list of locations in files.
 */
export class HierarchicalLocationsView extends React.PureComponent<Props, State> {
    public state: State = { locationsOrError: LOADING, locationsComplete: false }

    private componentUpdates = new Subject<Props>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        const locationsChanges = this.componentUpdates.pipe(
            map(({ locations }) => locations),
            distinctUntilChanged()
        )

        this.subscriptions.add(
            locationsChanges
                .pipe(
                    switchMap(locations =>
                        locations.pipe(
                            catchError(error => [asError(error) as ErrorLike]),
                            map(result => ({ locationsOrError: result || [], locationsComplete: false })),
                            startWith<Pick<State, 'locationsOrError' | 'locationsComplete'>>({
                                locationsOrError: LOADING,
                                locationsComplete: false,
                            }),
                            endWith({ locationsComplete: true })
                        )
                    )
                )
                .subscribe(stateUpdate => this.setState(stateUpdate), error => console.error(error))
        )

        this.componentUpdates.next(this.props)
    }

    public componentWillReceiveProps(nextProps: Props): void {
        this.componentUpdates.next(nextProps)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        if (isErrorLike(this.state.locationsOrError)) {
            return <FileLocationsError error={this.state.locationsOrError} />
        }
        if (
            this.state.locationsOrError === LOADING ||
            (!this.state.locationsComplete && this.state.locationsOrError.length === 0)
        ) {
            return <LoadingSpinner className="icon-inline m-1" />
        }
        if (this.state.locationsOrError.length === 0) {
            return <FileLocationsNotFound />
        }

        const GROUPS: {
            name: string
            defaultSize: number
            icon?: React.ComponentType<{ className?: string }>
            key: (uri: string) => string
        }[] = [
            {
                name: 'repo',
                defaultSize: 100,
                // icon: RepositoryIcon,
                key: (uri: string): string => parseRepoURI(uri).repoPath,
            },
        ]
        if (localStorage.getItem('locationsByFile') !== null) {
            GROUPS.push({
                name: 'file',
                defaultSize: 200,
                // icon: FileIcon,
                key: (uri: string): string => parseRepoURI(uri).filePath || '.',
            })
        }

        // Grouped locations.
        const locationsByGroup: Map<string, Location[]>[] = []

        // Groups with >0 locations, in order (preserving order reduces jitter as more results stream in).
        const orderedGroups: string[][] = []

        // The selected groups, or the first group if none is selected.
        const { locationsOrError } = this.state
        const selectedGroups: string[] | undefined = GROUPS.map(
            ({ key }, i) =>
                this.state.selectedGroups && this.state.selectedGroups[i]
                    ? this.state.selectedGroups[i]
                    : key(locationsOrError[0].uri)
        )

        if (selectedGroups) {
            for (const loc of locationsOrError) {
                const groups = GROUPS.map(({ key }) => key(loc.uri))
                for (const [i, group] of groups.entries()) {
                    if (!locationsByGroup[i]) {
                        locationsByGroup[i] = new Map<string, Location[]>()
                    }
                    let locs = locationsByGroup[i].get(group)
                    if (!locs) {
                        locs = []
                        if (!orderedGroups[i]) {
                            orderedGroups[i] = []
                        }
                        orderedGroups[i].push(group)
                    }
                    locs.push(loc)
                    locationsByGroup[i].set(group, locs)

                    if (selectedGroups[i] !== group) {
                        break
                    }
                }
            }
        }

        return (
            <div className={`hierarchical-locations-view ${this.props.className || ''}`}>
                {selectedGroups &&
                    GROUPS.map(
                        (group, i) =>
                            orderedGroups[i].length > 1 && (
                                <Resizable
                                    key={i}
                                    className="hierarchical-locations-view__resizable"
                                    handlePosition="right"
                                    storageKey={`hierarchical-locations-view-resizable:${group.name}`}
                                    defaultSize={group.defaultSize}
                                    element={
                                        <div className="list-group list-group-flush hierarchical-locations-view__list">
                                            {orderedGroups[i].map((groupName, j) => (
                                                <span
                                                    key={j}
                                                    className={`list-group-item hierarchical-locations-view__item ${
                                                        selectedGroups[i] === groupName ? 'active' : ''
                                                    }`}
                                                    // tslint:disable-next-line:jsx-no-lambda
                                                    onClick={e => this.onSelectTree(e, i, groupName)}
                                                >
                                                    <span
                                                        className="hierarchical-locations-view__item-name"
                                                        title={groupName}
                                                    >
                                                        {group.icon && (
                                                            <group.icon className="icon-inline hierarchical-locations-view__item-icon" />
                                                        )}
                                                        <span className="hierarchical-locations-view__item-name-text">
                                                            <RepoLink to={null} repoPath={groupName} />
                                                        </span>
                                                    </span>
                                                    <span className="badge badge-secondary badge-pill hierarchical-locations-view__item-badge">
                                                        {locationsByGroup[i].get(groupName)!.length}
                                                    </span>
                                                </span>
                                            ))}
                                            {!this.state.locationsComplete && (
                                                <LoadingSpinner className="icon-inline m-2" />
                                            )}
                                        </div>
                                    }
                                />
                            )
                    )}
                <FileLocations
                    className="hierarchical-locations-view__content"
                    locations={of(
                        locationsByGroup[locationsByGroup.length - 1].get(selectedGroups[selectedGroups.length - 1]) ||
                            null
                    )}
                    onSelect={this.props.onSelectLocation}
                    icon={RepositoryIcon}
                    isLightTheme={this.props.isLightTheme}
                    fetchHighlightedFileLines={this.props.fetchHighlightedFileLines}
                />
            </div>
        )
    }

    private onSelectTree = (e: React.MouseEvent<HTMLElement>, i: number, group: string): void => {
        e.preventDefault()

        const selectedGroups = [...(this.state.selectedGroups || [])]
        selectedGroups[i] = group
        this.setState({ selectedGroups })

        if (this.props.onSelectTree) {
            this.props.onSelectTree()
        }
    }
}
