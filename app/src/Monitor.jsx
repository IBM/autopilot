import React, { useState } from 'react';
import styled from "styled-components";
import {useNodesWithStatus} from "./api/getNodesWithStatus.js";
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
`;

function Monitor() {
    const { nodes, error } = useNodesWithStatus();
    const [searchQuery, setSearchQuery] = useState(''); // State for search query

    if (error) {
        return <div>Error loading node status</div>;
    }

    if (!nodes.length) {
        return <div>Loading...</div>;
    }
    
    // Filter nodes 
    const filteredNodes = nodes.filter(node =>
        node.name.toLowerCase().includes(searchQuery.toLowerCase())
    );

    return (
        <MonitorWrapper>
            <Helmet>
                <title>Monitor Cluster</title> {/* Set the page title here */}
            </Helmet>
            <h1>Monitor Cluster</h1>
            <SearchInput searchQuery={searchQuery} setSearchQuery={setSearchQuery} /> {/* Add Search Input */}
            <CollapsibleTable nodes={filteredNodes} /> {/* Display filtered nodes */}
        </MonitorWrapper>
    );
}

export default Monitor;
