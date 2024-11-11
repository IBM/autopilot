import * as React from 'react';
import { useState, useMemo, useEffect } from 'react';
import styled from 'styled-components';
import PropTypes from 'prop-types';
import { Table, TableHead, TableRow, TableBody, TableCell, TableContainer, Button, Pagination } from '@carbon/react';
import { ChevronDown, ChevronUp } from '@carbon/icons-react';
import ColumnFilter from './ColumnFilter';

const lightGreen = "#c2fdc2";
const lightRed = "#ffb4b4";

const ResponsiveTableContainer = styled(TableContainer)`
    width: 100%;
    overflow-x: auto;
    padding: 0;
    margin: 0;
    box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
`;

const StyledTableCell = styled(TableCell)`
    font-weight: bold !important; 
    font-size: 1.1rem !important;
    background-color: #f5f5f5 !important;
`;

const StyledTableRow = styled(TableRow)`
    background-color: ${(props) => (props.$pass ? lightGreen : lightRed)};
`;

const ExpandableTableWrapper = styled.div`
    padding: 3px;
`;

const Row = ({ node }) => {
    const [open, setOpen] = useState(false);

    return (
        <>
            <StyledTableRow $pass={node.gpuHealth === 'PASS'}>
                <TableCell style={{ padding: 0, height: '3rem' }}>
                    <Button
                        kind="ghost"
                        size="sm"
                        renderIcon={open ? ChevronUp : ChevronDown}
                        iconDescription={open ? 'Collapse' : 'Expand'}
                        onClick={() => setOpen(!open)}
                        className="cds--layout--size-small"
                        style={{
                            display: 'grid',
                            placeItems: 'center',
                            width: '100%',
                            height: '100%',
                            minHeight: '3rem',
                            margin: 0
                        }}
                    >
                        <span className="cds--assistive-text"></span>
                    </Button>
                </TableCell>

                <TableCell>{node.name}</TableCell>
                <TableCell align="left">{node.status === 'True' ? 'Ready' : 'Not Ready'}</TableCell>
                <TableCell align="left">{node.version}</TableCell>
                <TableCell align="left">{node.architecture}</TableCell>
                <TableCell align="left">{node.gpuPresent}</TableCell>
                <TableCell align="left">{node.gpuModel}</TableCell>
                <TableCell align="left">{node.gpuCount}</TableCell>
                <TableCell align="left">{node.gpuHealth}</TableCell>
            </StyledTableRow>

            {open && (
                <TableRow>
                    <TableCell colSpan={10}>
                        <ExpandableTableWrapper>
                            <h4><strong>Capacity / Allocatable Resources:</strong></h4>
                            <Table size="sm" aria-label="resources">
                                <TableHead>
                                    <TableRow>
                                        <StyledTableCell>Resource</StyledTableCell>
                                        <StyledTableCell>Capacity</StyledTableCell>
                                        <StyledTableCell>Allocatable</StyledTableCell>
                                    </TableRow>
                                </TableHead>
                                <TableBody>
                                    <TableRow>
                                        <TableCell>GPU</TableCell>
                                        <TableCell>{node.capacity.gpu}</TableCell>
                                        <TableCell>{node.allocatable.gpu}</TableCell>
                                    </TableRow>
                                    <TableRow>
                                        <TableCell>CPU</TableCell>
                                        <TableCell>{node.capacity.cpu}</TableCell>
                                        <TableCell>{node.allocatable.cpu}</TableCell>
                                    </TableRow>
                                    <TableRow>
                                        <TableCell>Memory</TableCell>
                                        <TableCell>{node.capacity.memory}</TableCell>
                                        <TableCell>{node.allocatable.memory}</TableCell>
                                    </TableRow>
                                </TableBody>
                            </Table>
                        </ExpandableTableWrapper>
                        <br />
                        <ExpandableTableWrapper>
                            <h4><strong>GPU DCGM Level 3 Diagnostics:</strong></h4>
                            <Table size="sm" aria-label="resources">
                                <TableHead>
                                    <TableRow>
                                        <StyledTableCell>DCGM Status</StyledTableCell>
                                        <StyledTableCell>Timestamp</StyledTableCell>
                                        <StyledTableCell>Details</StyledTableCell>
                                    </TableRow>
                                </TableHead>
                                <TableBody>
                                    <TableRow>
                                        <TableCell>{node.dcgmStatus}</TableCell>
                                        <TableCell>{node.dcgmTimestamp}</TableCell>
                                        <TableCell>
                                            {node.dcgmStatus === 'ERR' ? (
                                                <ul>
                                                    {node.dcgmDetails.map((detail, index) => (
                                                        <li key={index}>
                                                            {`Test: ${detail.testName}, GPU ID: ${detail.gpuID}`}
                                                        </li>
                                                    ))}
                                                </ul>
                                            ) : 'No Details Available'}
                                        </TableCell>
                                    </TableRow>
                                </TableBody>
                            </Table>
                        </ExpandableTableWrapper>
                    </TableCell>
                </TableRow>
            )}
        </>
    );
};

function CollapsibleTable({ nodes, filters, onFilterChange }) {
    const [currentPage, setCurrentPage] = useState(1);
    const [itemsPerPage, setItemsPerPage] = useState(10);

    // Replace individual filter states with props
    const {
        gpuHealths: selectedGpuHealths,
        statuses: selectedStatuses,
        versions: selectedVersions,
        architectures: selectedArchitectures,
        gpuPresents: selectedGpuPresents,
        gpuModels: selectedGpuModels,
        gpuCounts: selectedGpuCounts
    } = filters;

    // Update filter handlers to use onFilterChange prop
    const handleGpuHealthFilterChange = (selectedItems) => onFilterChange('gpuHealths', selectedItems);
    const handleStatusFilterChange = (selectedItems) => onFilterChange('statuses', selectedItems);
    const handleVersionFilterChange = (selectedItems) => onFilterChange('versions', selectedItems);
    const handleArchitectureFilterChange = (selectedItems) => onFilterChange('architectures', selectedItems);
    const handleGpuPresentFilterChange = (selectedItems) => onFilterChange('gpuPresents', selectedItems);
    const handleGpuModelFilterChange = (selectedItems) => onFilterChange('gpuModels', selectedItems);
    const handleGpuCountFilterChange = (selectedItems) => onFilterChange('gpuCounts', selectedItems);

    // Memoized filtered nodes
    const filteredNodes = useMemo(() => {
        return nodes.filter(node => {
            const gpuHealthMatch = selectedGpuHealths.length === 0 || selectedGpuHealths.includes(node.gpuHealth);
            const statusMatch = selectedStatuses.length === 0 || selectedStatuses.includes(node.status === 'True' ? 'Ready' : 'Not Ready');
            const versionMatch = selectedVersions.length === 0 || selectedVersions.includes(node.version);
            const architectureMatch = selectedArchitectures.length === 0 || selectedArchitectures.includes(node.architecture);
            const gpuPresentMatch = selectedGpuPresents.length === 0 || selectedGpuPresents.includes(node.gpuPresent);
            const gpuModelMatch = selectedGpuModels.length === 0 || selectedGpuModels.includes(node.gpuModel);
            const gpuCountMatch = selectedGpuCounts.length === 0 || selectedGpuCounts.includes(node.gpuCount.toString()); // Convert to string for comparison

            return gpuHealthMatch && statusMatch && versionMatch && architectureMatch && gpuPresentMatch && gpuModelMatch && gpuCountMatch;
        });
    }, [nodes, selectedGpuHealths, selectedStatuses, selectedVersions, selectedArchitectures, selectedGpuPresents, selectedGpuModels, selectedGpuCounts]);

    // Calculate total pages and slice nodes
    const totalItems = filteredNodes.length;
    const nodesToDisplay = filteredNodes.slice(
        (currentPage - 1) * itemsPerPage,
        currentPage * itemsPerPage
    );

    // Handle page change
    const handlePageChange = ({ page, pageSize }) => {
        setCurrentPage(page);
        setItemsPerPage(pageSize);
    };

    // Reset current page when filters change
    useEffect(() => {
        setCurrentPage(1);
    }, [
        selectedGpuHealths,
        selectedStatuses,
        selectedVersions,
        selectedArchitectures,
        selectedGpuPresents,
        selectedGpuModels,
        selectedGpuCounts
    ]);

    const uniqueGpuHealths = useMemo(() => [...new Set(nodes.map(node => node.gpuHealth))], [nodes]);
    const uniqueStatuses = useMemo(() => [...new Set(nodes.map(node => (node.status === 'True' ? 'Ready' : 'Not Ready')))], [nodes]);
    const uniqueVersions = useMemo(() => [...new Set(nodes.map(node => node.version))], [nodes]);
    const uniqueArchitectures = useMemo(() => [...new Set(nodes.map(node => node.architecture))], [nodes]);
    const uniqueGpuPresents = useMemo(() => [...new Set(nodes.map(node => node.gpuPresent))], [nodes]);
    const uniqueGpuModels = useMemo(() => [...new Set(nodes.map(node => node.gpuModel))], [nodes]);
    const uniqueGpuCounts = useMemo(() => [...new Set(nodes.map(node => node.gpuCount.toString()))], [nodes]); // Convert to string for unique values

    return (
        <>
            <ResponsiveTableContainer>
                <Table>
                    <TableHead>
                        <TableRow>
                            <TableCell />
                            <StyledTableCell>Node Name</StyledTableCell>
                            <StyledTableCell>
                                Status
                                <ColumnFilter
                                    label="Status"
                                    items={uniqueStatuses}
                                    selectedFilters={selectedStatuses}
                                    onFilterChange={handleStatusFilterChange}
                                />
                            </StyledTableCell>
                            <StyledTableCell>
                                Version
                                <ColumnFilter
                                    label="Version"
                                    items={uniqueVersions}
                                    selectedFilters={selectedVersions}
                                    onFilterChange={handleVersionFilterChange}
                                />
                            </StyledTableCell>
                            <StyledTableCell>
                                Architecture
                                <ColumnFilter
                                    label="Architecture"
                                    items={uniqueArchitectures}
                                    selectedFilters={selectedArchitectures}
                                    onFilterChange={handleArchitectureFilterChange}
                                />
                            </StyledTableCell>
                            <StyledTableCell>
                                GPU Present
                                <ColumnFilter
                                    label="GPU Present"
                                    items={uniqueGpuPresents}
                                    selectedFilters={selectedGpuPresents}
                                    onFilterChange={handleGpuPresentFilterChange}
                                />
                            </StyledTableCell>
                            <StyledTableCell>
                                GPU Type
                                <ColumnFilter
                                    label="GPU Model"
                                    items={uniqueGpuModels}
                                    selectedFilters={selectedGpuModels}
                                    onFilterChange={handleGpuModelFilterChange}
                                />
                            </StyledTableCell>
                            <StyledTableCell>
                                GPU Count
                                <ColumnFilter
                                    label="GPU Count"
                                    items={uniqueGpuCounts}
                                    selectedFilters={selectedGpuCounts}
                                    onFilterChange={handleGpuCountFilterChange}
                                />
                            </StyledTableCell>
                            <StyledTableCell>
                                GPU Health
                                <ColumnFilter
                                    label="GPU Health"
                                    items={uniqueGpuHealths}
                                    selectedFilters={selectedGpuHealths}
                                    onFilterChange={handleGpuHealthFilterChange}
                                />
                            </StyledTableCell>
                        </TableRow>
                    </TableHead>
                    {filteredNodes.length > 0 ? (
                        <TableBody>
                            {nodesToDisplay.map((node) => (
                                <Row key={node.name} node={node} />
                            ))}
                        </TableBody>
                    ) : (
                        <TableBody>
                            <TableRow>
                                <TableCell colSpan="10" align="center">
                                    No matching nodes found.
                                </TableCell>
                            </TableRow>
                        </TableBody>
                    )}
                </Table>
            </ResponsiveTableContainer>
            <Pagination
                totalItems={totalItems}
                page={currentPage}
                pageSize={itemsPerPage}
                pageSizes={[5, 10, 20, 50]}
                onChange={handlePageChange}
            />
        </>
    );
}

// PropTypes for Row and CollapsibleTable
Row.propTypes = {
    node: PropTypes.shape({
        name: PropTypes.string.isRequired,
        status: PropTypes.string.isRequired,
        version: PropTypes.string.isRequired,
        architecture: PropTypes.string.isRequired,
        gpuPresent: PropTypes.string.isRequired,
        gpuModel: PropTypes.string.isRequired,
        gpuCount: PropTypes.string.isRequired,
        gpuHealth: PropTypes.string.isRequired,
        capacity: PropTypes.shape({
            gpu: PropTypes.string.isRequired,
            cpu: PropTypes.string.isRequired,
            memory: PropTypes.string.isRequired,
        }).isRequired,
        allocatable: PropTypes.shape({
            gpu: PropTypes.string.isRequired,
            cpu: PropTypes.string.isRequired,
            memory: PropTypes.string.isRequired,
        }).isRequired,
        dcgmStatus: PropTypes.string.isRequired,
        dcgmTimestamp: PropTypes.string.isRequired,
        dcgmDetails: PropTypes.arrayOf(
            PropTypes.shape({
                testName: PropTypes.string.isRequired,
                gpuID: PropTypes.string.isRequired,
            })
        ).isRequired,
    }).isRequired,
};


CollapsibleTable.propTypes = {
    nodes: PropTypes.arrayOf(
        PropTypes.shape({
            name: PropTypes.string.isRequired,
            status: PropTypes.string.isRequired,
            version: PropTypes.string.isRequired,
            architecture: PropTypes.string.isRequired,
            gpuPresent: PropTypes.string.isRequired,
            gpuModel: PropTypes.string.isRequired,
            gpuCount: PropTypes.string.isRequired,
            gpuHealth: PropTypes.string.isRequired,
            capacity: PropTypes.object.isRequired,
            allocatable: PropTypes.object.isRequired,
            dcgmStatus: PropTypes.string.isRequired,
            dcgmTimestamp: PropTypes.string.isRequired,
            dcgmDetails: PropTypes.arrayOf(
                PropTypes.shape({
                    testName: PropTypes.string.isRequired,
                    gpuID: PropTypes.string.isRequired,
                })
            ).isRequired,
        })
    ).isRequired,
    filters: PropTypes.shape({
        gpuHealths: PropTypes.arrayOf(PropTypes.string).isRequired,
        statuses: PropTypes.arrayOf(PropTypes.string).isRequired,
        versions: PropTypes.arrayOf(PropTypes.string).isRequired,
        architectures: PropTypes.arrayOf(PropTypes.string).isRequired,
        gpuPresents: PropTypes.arrayOf(PropTypes.string).isRequired,
        gpuModels: PropTypes.arrayOf(PropTypes.string).isRequired,
        gpuCounts: PropTypes.arrayOf(PropTypes.string).isRequired
    }).isRequired,
    onFilterChange: PropTypes.func.isRequired
};

export default CollapsibleTable;