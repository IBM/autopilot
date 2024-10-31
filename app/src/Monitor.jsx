import React, { useState, useEffect } from 'react';
import styled from "styled-components";
// import { useNodesWithStatus } from "./api/getNodesWithStatus.js";
import watchNodesWithStatus from "./api/watchNodesWithStatus.js";
import CollapsibleTable from "./components/CollapsibleTable.jsx";
import SearchInput from './components/SearchInput';
import { Helmet } from 'react-helmet';

// Displaying live node labels and status + current health checks

const MonitorWrapper = styled.div`
    padding-left: 0;
    margin: 0;
    width: 100%;
    
    @media (max-width: 768px){
        margin: 0;
        padding-top: 70px;
        
    }

    h1 {
        text-align: center;
        padding-bottom: 20px;
    }
`;

function Monitor() {
    // const { nodes, error } = useNodesWithStatus();
    const[nodes, setNodes] = useState([]);
    const [searchQuery, setSearchQuery] = useState(''); // State for search query

    /*
    if (error) {
        return <div>Error loading node status</div>;
    }

    if (!nodes.length) {
        return <div>Loading...</div>;
    }
    */

    // Log the first node for debugging purposes
    // console.log(nodes[0]);

    useEffect(() => {
        const handleNodeChange = (node, isDeleted = false) => {
            setNodes(prevNodes => {
                if (isDeleted) { // Removing deleted node
                    return prevNodes.filter(n => n.name !== node.name);
                } else {
                    // Update the existing node with new details
                    const existingNodeIndex = prevNodes.findIndex(n => n.name === node.name);
                    if (existingNodeIndex >= 0) {
                        return prevNodes.map((n, i) => (i === existingNodeIndex ? node : n));
                    }
                    // Add new node
                    return [...prevNodes, node];
                }
            });
        };

        watchNodesWithStatus(handleNodeChange)
            .then(() => console.log('Started watching nodes'))
            .catch((err) => {
                console.error('Error fetching nodes:', err);
            });
    }, []);

    // Filter nodes based on search query
    const filteredNodes = nodes.filter(node => {
        const searchQueryLower = searchQuery.toLowerCase();
        return Object.values(node).some(value =>
            value.toString().toLowerCase().includes(searchQueryLower)
        );
    });

    // // Filter nodes by all fields except memory, including the readiness condition for status
    // const filteredNodes = nodes.filter(node => {
    //     const {
    //         name, role, status, version, architecture, gpuPresent,
    //         gpuHealth, gpuCount, gpuModel, dcgmStatus, dcgmTimestamp, capacity, allocatable
    //     } = node;
    //
    //     const readinessStatus = status === 'True' ? 'Ready' : 'Not Ready'; // Convert status to Ready/Not Ready
    //     const searchQueryLower = searchQuery.toLowerCase();
    //
    //     return (
    //         name.toLowerCase().includes(searchQueryLower) ||
    //         role.toLowerCase().includes(searchQueryLower) ||
    //         readinessStatus.toLowerCase().includes(searchQueryLower) || // Add readiness condition to filtering
    //         version.toLowerCase().includes(searchQueryLower) ||
    //         architecture.toLowerCase().includes(searchQueryLower) ||
    //         gpuPresent.toLowerCase().includes(searchQueryLower) ||
    //         gpuHealth.toLowerCase().includes(searchQueryLower) ||
    //         gpuCount.toLowerCase().includes(searchQueryLower) ||
    //         gpuModel.toLowerCase().includes(searchQueryLower) ||
    //         dcgmStatus.toLowerCase().includes(searchQueryLower) ||
    //         dcgmTimestamp.toLowerCase().includes(searchQueryLower) ||
    //         capacity.gpu.toLowerCase().includes(searchQueryLower) ||
    //         capacity.cpu.toLowerCase().includes(searchQueryLower) ||
    //         allocatable.gpu.toLowerCase().includes(searchQueryLower) ||
    //         allocatable.cpu.toLowerCase().includes(searchQueryLower)
    //     );
    // });

    return (
        <MonitorWrapper>
            <Helmet>
                <title>Monitor Cluster</title> {/* Set the page title here */}
            </Helmet>
            <h1>Monitor Cluster</h1>
            <SearchInput
                searchQuery={searchQuery}
                setSearchQuery={setSearchQuery}
                label="Search Features"
            />
            <CollapsibleTable nodes={filteredNodes} />
            {/*<CollapsibleTable nodes={filteredNodes.map(node => ({*/}
            {/*    ...node,*/}
            {/*    readiness: node.status === 'True' ? 'Ready' : 'Not Ready' // Add readiness status to the node object*/}
            {/*}))} /> /!* Display filtered nodes with readiness status *!/*/}
        </MonitorWrapper>
    );
}


export default Monitor;
