import * as React from 'react';
import { useState, useMemo, useRef, useEffect } from 'react';
import styled from 'styled-components';
import PropTypes from 'prop-types';
import { Table, TableHead, TableRow, TableBody, TableCell, TableContainer, Button, Dropdown, MultiSelect } from '@carbon/react';
import { ChevronDown, ChevronUp, Filter } from '@carbon/icons-react'; 
import ReactDOM from 'react-dom'; 

const lightGreen = "#90EE90";
const lightRed = "#FAA0A0";

const ResponsiveTableContainer = styled(TableContainer)`
    width: 100%;
    overflow-x: auto;
    padding: 0;
    margin: 0;
    box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);

    @media (max-width: 768px) {
        width: 100%;
        padding: 0;
        margin: 0;
    }
`;

const StyledTableCell = styled(TableCell)`
  font-weight: bold !important; 
  font-size: 1.1rem !important;
  background-color: #f5f5f5 !important;
`;

const StyledTableRow = styled(TableRow).withConfig({
    shouldForwardProp: (prop) => prop !== 'pass',
})`
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

                {/* Main table */}
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
                        <br/>
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
                                            ) : `No Details Available`}
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

import ColumnFilter from './ColumnFilter'; 

function CollapsibleTable({ nodes }) {
    const [selectedGpuHealths, setSelectedGpuHealths] = useState([]);

    // Memoized filtered nodes
    const filteredNodes = useMemo(() => {
        if (selectedGpuHealths.length === 0) return nodes;
        return nodes.filter(node => selectedGpuHealths.includes(node.gpuHealth));
    }, [nodes, selectedGpuHealths]);

    const uniqueGpuHealths = useMemo(() => {
        return [...new Set(nodes.map(node => node.gpuHealth))];
    }, [nodes]);

    const handleGpuHealthFilterChange = (selectedItems) => {
        setSelectedGpuHealths(selectedItems);
    };

    

    return (
        <ResponsiveTableContainer>
            <Table>
                <TableHead>
                    <TableRow>
                        <TableCell />
                        <StyledTableCell>Node Name</StyledTableCell>
                        <StyledTableCell>Status</StyledTableCell>
                        <StyledTableCell>Role</StyledTableCell>
                        <StyledTableCell>Version</StyledTableCell>
                        <StyledTableCell>Architecture</StyledTableCell>
                        <StyledTableCell>GPU Present</StyledTableCell>
                        <StyledTableCell>GPU Type</StyledTableCell>
                        <StyledTableCell>GPU Count</StyledTableCell>
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
