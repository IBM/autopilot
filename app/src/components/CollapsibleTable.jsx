import * as React from 'react';
import { useState, useMemo } from 'react';
import styled from 'styled-components';
import PropTypes from 'prop-types';
import { Table, TableHead, TableRow, TableBody, TableCell, TableContainer, Button } from '@carbon/react';
import { ChevronDown, ChevronUp } from '@carbon/icons-react';
import ColumnFilter from './ColumnFilter';

const lightGreen = "#90EE90";
const lightRed = "#FAA0A0";

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
    background-color: ${(props) => (props.pass ? lightGreen : lightRed)};
`;

const ExpandableTableWrapper = styled.div`
    padding: 3px;
`;

const Row = ({ node }) => {
    const [open, setOpen] = useState(false);

    return (
        <>
            <StyledTableRow pass={node.gpuHealth === 'PASS'}>
                <TableCell style={{ padding: 0, height: '3rem' }}>
                    <Button
                        kind="ghost"
                        size="small"
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
                <TableCell align="left">{node.role}</TableCell>
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
                            <Table size="small" aria-label="resources">
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
                            <Table size="small" aria-label="resources">
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

function CollapsibleTable({ nodes }) {
    const [selectedGpuHealths, setSelectedGpuHealths] = useState([]);
    const [selectedStatuses, setSelectedStatuses] = useState([]);
    const [selectedRoles, setSelectedRoles] = useState([]);
    const [selectedVersions, setSelectedVersions] = useState([]);
    const [selectedArchitectures, setSelectedArchitectures] = useState([]);
    const [selectedGpuPresents, setSelectedGpuPresents] = useState([]);
    const [selectedGpuModels, setSelectedGpuModels] = useState([]);
    const [selectedGpuCounts, setSelectedGpuCounts] = useState([]);

    // Memoized filtered nodes
    const filteredNodes = useMemo(() => {
        return nodes.filter(node => {
            const gpuHealthMatch = selectedGpuHealths.length === 0 || selectedGpuHealths.includes(node.gpuHealth);
            const statusMatch = selectedStatuses.length === 0 || selectedStatuses.includes(node.status === 'True' ? 'Ready' : 'Not Ready');
            const roleMatch = selectedRoles.length === 0 || selectedRoles.includes(node.role);
            const versionMatch = selectedVersions.length === 0 || selectedVersions.includes(node.version);
            const architectureMatch = selectedArchitectures.length === 0 || selectedArchitectures.includes(node.architecture);
            const gpuPresentMatch = selectedGpuPresents.length === 0 || selectedGpuPresents.includes(node.gpuPresent);
            const gpuModelMatch = selectedGpuModels.length === 0 || selectedGpuModels.includes(node.gpuModel);
            const gpuCountMatch = selectedGpuCounts.length === 0 || selectedGpuCounts.includes(node.gpuCount.toString()); // Convert to string for comparison

            return gpuHealthMatch && statusMatch && roleMatch && versionMatch && architectureMatch && gpuPresentMatch && gpuModelMatch && gpuCountMatch;
        });
    }, [nodes, selectedGpuHealths, selectedStatuses, selectedRoles, selectedVersions, selectedArchitectures, selectedGpuPresents, selectedGpuModels, selectedGpuCounts]);

    const uniqueGpuHealths = useMemo(() => [...new Set(nodes.map(node => node.gpuHealth))], [nodes]);
    const uniqueStatuses = useMemo(() => [...new Set(nodes.map(node => (node.status === 'True' ? 'Ready' : 'Not Ready')))], [nodes]);
    const uniqueRoles = useMemo(() => [...new Set(nodes.map(node => node.role))], [nodes]);
    const uniqueVersions = useMemo(() => [...new Set(nodes.map(node => node.version))], [nodes]);
    const uniqueArchitectures = useMemo(() => [...new Set(nodes.map(node => node.architecture))], [nodes]);
    const uniqueGpuPresents = useMemo(() => [...new Set(nodes.map(node => node.gpuPresent))], [nodes]);
    const uniqueGpuModels = useMemo(() => [...new Set(nodes.map(node => node.gpuModel))], [nodes]);
    const uniqueGpuCounts = useMemo(() => [...new Set(nodes.map(node => node.gpuCount.toString()))], [nodes]); // Convert to string for unique values

    const handleGpuHealthFilterChange = (selectedItems) => setSelectedGpuHealths(selectedItems);
    const handleStatusFilterChange = (selectedItems) => setSelectedStatuses(selectedItems);
    const handleRoleFilterChange = (selectedItems) => setSelectedRoles(selectedItems);
    const handleVersionFilterChange = (selectedItems) => setSelectedVersions(selectedItems);
    const handleArchitectureFilterChange = (selectedItems) => setSelectedArchitectures(selectedItems);
    const handleGpuPresentFilterChange = (selectedItems) => setSelectedGpuPresents(selectedItems);
    const handleGpuModelFilterChange = (selectedItems) => setSelectedGpuModels(selectedItems);
    const handleGpuCountFilterChange = (selectedItems) => setSelectedGpuCounts(selectedItems);

    return (
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
                            Role
                            <ColumnFilter
                                label="Role"
                                items={uniqueRoles}
                                selectedFilters={selectedRoles}
                                onFilterChange={handleRoleFilterChange}
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
                <TableBody>
                    {filteredNodes.map((node) => (
                        <Row key={node.name} node={node} />
                    ))}
                </TableBody>
            </Table>
        </ResponsiveTableContainer>
    );
}



CollapsibleTable.propTypes = {
    nodes: PropTypes.arrayOf(
        PropTypes.shape({
            name: PropTypes.string.isRequired,
            status: PropTypes.string.isRequired,
            role: PropTypes.string.isRequired,
            version: PropTypes.string.isRequired,
            architecture: PropTypes.string.isRequired,
            gpuPresent: PropTypes.string.isRequired,
            gpuModel: PropTypes.string.isRequired,
            gpuCount: PropTypes.number.isRequired,
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
};

export default CollapsibleTable;
